#!/usr/bin/env python3
"""
Neo4j Data Sync Script
Syncs product data and relationships from Oracle database to Neo4j graph database.

Usage:
    python sync_data.py --mode full      # Full sync (initial load)
    python sync_data.py --mode products  # Sync only products
    python sync_data.py --mode relations # Sync only relationships
    python sync_data.py --mode users     # Sync only user interactions
"""

import os
import sys
import argparse
import logging
from datetime import datetime, timedelta
from typing import List, Dict, Any

# Neo4j driver
try:
    from neo4j import GraphDatabase
except ImportError:
    print("Please install neo4j driver: pip install neo4j")
    sys.exit(1)

# Oracle driver (cx_Oracle or oracledb)
try:
    import oracledb as cx_Oracle
except ImportError:
    try:
        import cx_Oracle
    except ImportError:
        print("Please install Oracle driver: pip install oracledb")
        sys.exit(1)

# Logging setup
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Configuration from environment
NEO4J_URI = os.getenv('NEO4J_URI', 'bolt://localhost:7687')
NEO4J_USER = os.getenv('NEO4J_USER', 'neo4j')
NEO4J_PASSWORD = os.getenv('NEO4J_PASSWORD', 'florusneo4j123')

ORACLE_HOST = os.getenv('ORACLE_HOST', 'localhost')
ORACLE_PORT = os.getenv('ORACLE_PORT', '1522')
ORACLE_SERVICE = os.getenv('ORACLE_SERVICE', 'XEPDB1')
ORACLE_USER = os.getenv('ORACLE_USER', 'florus')
ORACLE_PASSWORD = os.getenv('ORACLE_PASSWORD', 'florus123')


class DataSyncer:
    def __init__(self):
        self.neo4j_driver = None
        self.oracle_conn = None

    def connect_neo4j(self):
        """Connect to Neo4j database"""
        logger.info(f"Connecting to Neo4j at {NEO4J_URI}")
        self.neo4j_driver = GraphDatabase.driver(
            NEO4J_URI,
            auth=(NEO4J_USER, NEO4J_PASSWORD)
        )
        # Verify connection
        with self.neo4j_driver.session() as session:
            result = session.run("RETURN 1 AS test")
            result.single()
        logger.info("Neo4j connection successful")

    def connect_oracle(self):
        """Connect to Oracle database"""
        dsn = f"{ORACLE_HOST}:{ORACLE_PORT}/{ORACLE_SERVICE}"
        logger.info(f"Connecting to Oracle at {dsn}")
        self.oracle_conn = cx_Oracle.connect(
            user=ORACLE_USER,
            password=ORACLE_PASSWORD,
            dsn=dsn
        )
        logger.info("Oracle connection successful")

    def close(self):
        """Close all connections"""
        if self.neo4j_driver:
            self.neo4j_driver.close()
        if self.oracle_conn:
            self.oracle_conn.close()

    def init_schema(self):
        """Initialize Neo4j schema (constraints and indexes)"""
        logger.info("Initializing Neo4j schema...")
        with self.neo4j_driver.session() as session:
            # Constraints
            constraints = [
                "CREATE CONSTRAINT product_id IF NOT EXISTS FOR (p:Product) REQUIRE p.id IS UNIQUE",
                "CREATE CONSTRAINT user_id IF NOT EXISTS FOR (u:User) REQUIRE u.id IS UNIQUE",
                "CREATE CONSTRAINT category_id IF NOT EXISTS FOR (c:Category) REQUIRE c.id IS UNIQUE",
                "CREATE CONSTRAINT brand_id IF NOT EXISTS FOR (b:Brand) REQUIRE b.id IS UNIQUE",
            ]
            for c in constraints:
                try:
                    session.run(c)
                except Exception as e:
                    logger.warning(f"Constraint may already exist: {e}")

            # Indexes
            indexes = [
                "CREATE INDEX product_category IF NOT EXISTS FOR (p:Product) ON (p.category_id)",
                "CREATE INDEX product_rating IF NOT EXISTS FOR (p:Product) ON (p.average_rating)",
                "CREATE INDEX product_price IF NOT EXISTS FOR (p:Product) ON (p.price)",
            ]
            for idx in indexes:
                try:
                    session.run(idx)
                except Exception as e:
                    logger.warning(f"Index may already exist: {e}")
        logger.info("Schema initialization complete")

    def sync_categories(self):
        """Sync categories from Oracle to Neo4j"""
        logger.info("Syncing categories...")
        cursor = self.oracle_conn.cursor()
        cursor.execute("SELECT id, name, parent_id FROM categories")
        categories = cursor.fetchall()
        cursor.close()

        with self.neo4j_driver.session() as session:
            for cat_id, name, parent_id in categories:
                session.run("""
                    MERGE (c:Category {id: $id})
                    SET c.name = $name, c.parent_id = $parent_id
                """, id=cat_id, name=name, parent_id=parent_id)
        logger.info(f"Synced {len(categories)} categories")

    def sync_products(self, batch_size: int = 1000):
        """Sync products from Oracle to Neo4j"""
        logger.info("Syncing products...")
        cursor = self.oracle_conn.cursor()
        cursor.execute("""
            SELECT
                p.id, p.name, p.price, p.category_id, p.brand,
                p.average_rating, p.view_count, p.purchase_count,
                p.created_at
            FROM products p
            WHERE p.deleted_at IS NULL
        """)

        count = 0
        batch = []
        for row in cursor:
            batch.append({
                'id': row[0],
                'name': row[1],
                'price': float(row[2]) if row[2] else 0,
                'category_id': row[3],
                'brand': row[4] or 'Unknown',
                'average_rating': float(row[5]) if row[5] else 0,
                'view_count': row[6] or 0,
                'purchase_count': row[7] or 0,
                'created_at': row[8].isoformat() if row[8] else None
            })

            if len(batch) >= batch_size:
                self._insert_products_batch(batch)
                count += len(batch)
                batch = []
                logger.info(f"Synced {count} products...")

        if batch:
            self._insert_products_batch(batch)
            count += len(batch)

        cursor.close()
        logger.info(f"Total products synced: {count}")

    def _insert_products_batch(self, products: List[Dict]):
        """Insert a batch of products into Neo4j"""
        with self.neo4j_driver.session() as session:
            session.run("""
                UNWIND $products AS p
                MERGE (prod:Product {id: p.id})
                SET prod.name = p.name,
                    prod.price = p.price,
                    prod.category_id = p.category_id,
                    prod.brand = p.brand,
                    prod.average_rating = p.average_rating,
                    prod.view_count = p.view_count,
                    prod.purchase_count = p.purchase_count,
                    prod.created_at = p.created_at
            """, products=products)

            # Create category relationships
            session.run("""
                UNWIND $products AS p
                MATCH (prod:Product {id: p.id})
                MATCH (cat:Category {id: p.category_id})
                MERGE (prod)-[:BELONGS_TO]->(cat)
            """, products=products)

    def sync_product_relationships(self):
        """Sync product relationships from Oracle to Neo4j"""
        logger.info("Syncing product relationships...")
        cursor = self.oracle_conn.cursor()
        cursor.execute("""
            SELECT
                source_product_id, target_product_id,
                relationship_type, score
            FROM product_relationships
        """)

        relationships = {
            'similar': [],
            'co_viewed': [],
            'frequently_bought_with': []
        }

        for row in cursor:
            rel = {
                'source': row[0],
                'target': row[1],
                'score': float(row[3]) if row[3] else 0.5
            }
            rel_type = row[2] or 'similar'
            if rel_type in relationships:
                relationships[rel_type].append(rel)

        cursor.close()

        with self.neo4j_driver.session() as session:
            # Similar products
            if relationships['similar']:
                session.run("""
                    UNWIND $rels AS r
                    MATCH (p1:Product {id: r.source})
                    MATCH (p2:Product {id: r.target})
                    MERGE (p1)-[rel:SIMILAR_TO]->(p2)
                    SET rel.score = r.score
                """, rels=relationships['similar'])
                logger.info(f"Synced {len(relationships['similar'])} SIMILAR_TO relationships")

            # Co-viewed products
            if relationships['co_viewed']:
                session.run("""
                    UNWIND $rels AS r
                    MATCH (p1:Product {id: r.source})
                    MATCH (p2:Product {id: r.target})
                    MERGE (p1)-[rel:CO_VIEWED]->(p2)
                    SET rel.count = toInteger(r.score * 100)
                """, rels=relationships['co_viewed'])
                logger.info(f"Synced {len(relationships['co_viewed'])} CO_VIEWED relationships")

            # Frequently bought together
            if relationships['frequently_bought_with']:
                session.run("""
                    UNWIND $rels AS r
                    MATCH (p1:Product {id: r.source})
                    MATCH (p2:Product {id: r.target})
                    MERGE (p1)-[rel:FREQUENTLY_BOUGHT_WITH]->(p2)
                    SET rel.count = toInteger(r.score * 100)
                """, rels=relationships['frequently_bought_with'])
                logger.info(f"Synced {len(relationships['frequently_bought_with'])} FREQUENTLY_BOUGHT_WITH relationships")

    def sync_user_interactions(self, days: int = 30):
        """Sync recent user interactions from Oracle to Neo4j"""
        logger.info(f"Syncing user interactions from last {days} days...")
        since_date = datetime.now() - timedelta(days=days)

        cursor = self.oracle_conn.cursor()

        # Sync purchases
        cursor.execute("""
            SELECT
                user_id, product_id, COUNT(*) as purchase_count,
                MAX(created_at) as last_purchase
            FROM user_events
            WHERE event_type = 'purchase'
              AND created_at > :since_date
            GROUP BY user_id, product_id
        """, since_date=since_date)

        purchases = [
            {'user_id': r[0], 'product_id': r[1], 'count': r[2], 'timestamp': r[3].isoformat() if r[3] else None}
            for r in cursor
        ]

        if purchases:
            with self.neo4j_driver.session() as session:
                session.run("""
                    UNWIND $purchases AS p
                    MERGE (u:User {id: p.user_id})
                    WITH u, p
                    MATCH (prod:Product {id: p.product_id})
                    MERGE (u)-[rel:PURCHASED]->(prod)
                    SET rel.count = p.count, rel.last_purchase = p.timestamp
                """, purchases=purchases)
            logger.info(f"Synced {len(purchases)} purchase interactions")

        # Sync views
        cursor.execute("""
            SELECT
                user_id, product_id, COUNT(*) as view_count,
                MAX(created_at) as last_viewed
            FROM user_events
            WHERE event_type = 'product_view'
              AND created_at > :since_date
            GROUP BY user_id, product_id
        """, since_date=since_date)

        views = [
            {'user_id': r[0], 'product_id': r[1], 'count': r[2], 'timestamp': r[3].isoformat() if r[3] else None}
            for r in cursor
        ]

        if views:
            with self.neo4j_driver.session() as session:
                session.run("""
                    UNWIND $views AS v
                    MERGE (u:User {id: v.user_id})
                    WITH u, v
                    MATCH (prod:Product {id: v.product_id})
                    MERGE (u)-[rel:VIEWED]->(prod)
                    SET rel.count = v.count, rel.last_viewed = v.timestamp
                """, views=views)
            logger.info(f"Synced {len(views)} view interactions")

        cursor.close()

    def create_similarity_relationships(self):
        """Create similarity relationships based on category and price"""
        logger.info("Creating category-based similarity relationships...")
        with self.neo4j_driver.session() as session:
            # Products in same category with similar price (within 20% range)
            result = session.run("""
                MATCH (p1:Product), (p2:Product)
                WHERE p1.id < p2.id
                  AND p1.category_id = p2.category_id
                  AND abs(p1.price - p2.price) / ((p1.price + p2.price) / 2) < 0.2
                  AND NOT EXISTS((p1)-[:SIMILAR_TO]->(p2))
                WITH p1, p2,
                     0.7 + (0.3 * (1 - abs(p1.price - p2.price) / ((p1.price + p2.price) / 2))) AS score
                LIMIT 10000
                CREATE (p1)-[:SIMILAR_TO {score: score, source: 'category_price'}]->(p2)
                RETURN count(*) AS created
            """)
            count = result.single()['created']
            logger.info(f"Created {count} category-based similarity relationships")


def main():
    parser = argparse.ArgumentParser(description='Sync data from Oracle to Neo4j')
    parser.add_argument('--mode', choices=['full', 'products', 'relations', 'users', 'similarity'],
                        default='full', help='Sync mode')
    parser.add_argument('--days', type=int, default=30,
                        help='Number of days for user interaction sync')
    args = parser.parse_args()

    syncer = DataSyncer()

    try:
        syncer.connect_neo4j()
        syncer.connect_oracle()
        syncer.init_schema()

        if args.mode == 'full':
            syncer.sync_categories()
            syncer.sync_products()
            syncer.sync_product_relationships()
            syncer.sync_user_interactions(args.days)
            syncer.create_similarity_relationships()
        elif args.mode == 'products':
            syncer.sync_categories()
            syncer.sync_products()
        elif args.mode == 'relations':
            syncer.sync_product_relationships()
            syncer.create_similarity_relationships()
        elif args.mode == 'users':
            syncer.sync_user_interactions(args.days)
        elif args.mode == 'similarity':
            syncer.create_similarity_relationships()

        logger.info("Sync completed successfully!")

    except Exception as e:
        logger.error(f"Sync failed: {e}")
        raise
    finally:
        syncer.close()


if __name__ == '__main__':
    main()
