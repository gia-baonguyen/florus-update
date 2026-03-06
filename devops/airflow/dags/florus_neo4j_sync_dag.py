"""
Florus Neo4j Data Sync DAG
Synchronizes product and user data from Oracle to Neo4j graph database.
Runs multiple times daily to keep the graph updated for recommendations.
"""

import os
from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.python import PythonOperator, BranchPythonOperator
from airflow.operators.bash import BashOperator
from airflow.operators.empty import EmptyOperator
from airflow.utils.trigger_rule import TriggerRule
import redis
import json

# Environment variables
NEO4J_HOST = os.getenv("NEO4J_HOST", "florus-neo4j")
NEO4J_BOLT_PORT = int(os.getenv("NEO4J_BOLT_PORT", "7687"))
NEO4J_USER = os.getenv("NEO4J_USER", "neo4j")
NEO4J_PASSWORD = os.getenv("NEO4J_PASSWORD", "florusneo4j123")

ORACLE_HOST = os.getenv("ORACLE_HOST", "florus-oracle")
ORACLE_PORT = int(os.getenv("ORACLE_PORT", "1521"))
ORACLE_SERVICE = os.getenv("ORACLE_SERVICE", "XEPDB1")
ORACLE_USER = os.getenv("ORACLE_USER", "florus")
ORACLE_PASSWORD = os.getenv("ORACLE_PASSWORD", "florus123")

REDIS_HOST = os.getenv("REDIS_HOST", "florus-redis")
REDIS_PORT = int(os.getenv("REDIS_PORT", "6379"))

default_args = {
    "owner": "florus",
    "depends_on_past": False,
    "email_on_failure": False,
    "email_on_retry": False,
    "retries": 2,
    "retry_delay": timedelta(minutes=5),
}


def check_neo4j_health(**context):
    """Check if Neo4j is healthy and accessible."""
    try:
        from neo4j import GraphDatabase

        uri = f"bolt://{NEO4J_HOST}:{NEO4J_BOLT_PORT}"
        driver = GraphDatabase.driver(uri, auth=(NEO4J_USER, NEO4J_PASSWORD))

        with driver.session() as session:
            result = session.run("RETURN 1 AS test")
            record = result.single()
            if record and record["test"] == 1:
                print("Neo4j health check passed")
                context["ti"].xcom_push(key="neo4j_healthy", value=True)
                driver.close()
                return "sync_products"

        driver.close()
        context["ti"].xcom_push(key="neo4j_healthy", value=False)
        return "skip_sync"

    except Exception as e:
        print(f"Neo4j health check failed: {e}")
        context["ti"].xcom_push(key="neo4j_healthy", value=False)
        return "skip_sync"


def sync_products_task(**context):
    """Sync products from Oracle to Neo4j."""
    try:
        from neo4j import GraphDatabase
        import oracledb as cx_Oracle

        # Connect to Oracle
        oracle_dsn = f"{ORACLE_HOST}:{ORACLE_PORT}/{ORACLE_SERVICE}"
        oracle_conn = cx_Oracle.connect(
            user=ORACLE_USER,
            password=ORACLE_PASSWORD,
            dsn=oracle_dsn
        )

        # Connect to Neo4j
        neo4j_uri = f"bolt://{NEO4J_HOST}:{NEO4J_BOLT_PORT}"
        neo4j_driver = GraphDatabase.driver(neo4j_uri, auth=(NEO4J_USER, NEO4J_PASSWORD))

        # Fetch products from Oracle
        cursor = oracle_conn.cursor()
        cursor.execute("""
            SELECT id, name, price, category_id, brand,
                   average_rating, view_count, purchase_count
            FROM products
            WHERE deleted_at IS NULL
        """)

        products = []
        for row in cursor:
            products.append({
                'id': row[0],
                'name': row[1],
                'price': float(row[2]) if row[2] else 0,
                'category_id': row[3],
                'brand': row[4] or 'Unknown',
                'average_rating': float(row[5]) if row[5] else 0,
                'view_count': row[6] or 0,
                'purchase_count': row[7] or 0
            })

        cursor.close()

        # Insert into Neo4j in batches
        batch_size = 500
        with neo4j_driver.session() as session:
            for i in range(0, len(products), batch_size):
                batch = products[i:i+batch_size]
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
                        prod.synced_at = datetime()
                """, products=batch)

        print(f"Synced {len(products)} products to Neo4j")
        context["ti"].xcom_push(key="products_synced", value=len(products))

        oracle_conn.close()
        neo4j_driver.close()

    except Exception as e:
        print(f"Error syncing products: {e}")
        context["ti"].xcom_push(key="products_synced", value=0)
        raise


def sync_categories_task(**context):
    """Sync categories from Oracle to Neo4j."""
    try:
        from neo4j import GraphDatabase
        import oracledb as cx_Oracle

        oracle_dsn = f"{ORACLE_HOST}:{ORACLE_PORT}/{ORACLE_SERVICE}"
        oracle_conn = cx_Oracle.connect(
            user=ORACLE_USER,
            password=ORACLE_PASSWORD,
            dsn=oracle_dsn
        )

        neo4j_uri = f"bolt://{NEO4J_HOST}:{NEO4J_BOLT_PORT}"
        neo4j_driver = GraphDatabase.driver(neo4j_uri, auth=(NEO4J_USER, NEO4J_PASSWORD))

        cursor = oracle_conn.cursor()
        cursor.execute("SELECT id, name, parent_id FROM categories")
        categories = [(row[0], row[1], row[2]) for row in cursor]
        cursor.close()

        with neo4j_driver.session() as session:
            for cat_id, name, parent_id in categories:
                session.run("""
                    MERGE (c:Category {id: $id})
                    SET c.name = $name, c.parent_id = $parent_id
                """, id=cat_id, name=name, parent_id=parent_id)

            # Create BELONGS_TO relationships for products
            session.run("""
                MATCH (p:Product), (c:Category)
                WHERE p.category_id = c.id AND NOT (p)-[:BELONGS_TO]->(c)
                CREATE (p)-[:BELONGS_TO]->(c)
            """)

        print(f"Synced {len(categories)} categories to Neo4j")
        context["ti"].xcom_push(key="categories_synced", value=len(categories))

        oracle_conn.close()
        neo4j_driver.close()

    except Exception as e:
        print(f"Error syncing categories: {e}")
        context["ti"].xcom_push(key="categories_synced", value=0)
        raise


def sync_user_interactions_task(**context):
    """Sync recent user interactions from Oracle to Neo4j."""
    try:
        from neo4j import GraphDatabase
        import oracledb as cx_Oracle

        oracle_dsn = f"{ORACLE_HOST}:{ORACLE_PORT}/{ORACLE_SERVICE}"
        oracle_conn = cx_Oracle.connect(
            user=ORACLE_USER,
            password=ORACLE_PASSWORD,
            dsn=oracle_dsn
        )

        neo4j_uri = f"bolt://{NEO4J_HOST}:{NEO4J_BOLT_PORT}"
        neo4j_driver = GraphDatabase.driver(neo4j_uri, auth=(NEO4J_USER, NEO4J_PASSWORD))

        cursor = oracle_conn.cursor()

        # Sync purchases from last 30 days
        cursor.execute("""
            SELECT user_id, product_id, COUNT(*) as cnt, MAX(created_at) as last_date
            FROM user_events
            WHERE event_type = 'purchase'
              AND created_at > SYSDATE - 30
            GROUP BY user_id, product_id
        """)

        purchases = []
        for row in cursor:
            purchases.append({
                'user_id': row[0],
                'product_id': row[1],
                'count': row[2],
                'last_date': row[3].isoformat() if row[3] else None
            })

        # Sync views from last 7 days
        cursor.execute("""
            SELECT user_id, product_id, COUNT(*) as cnt, MAX(created_at) as last_date
            FROM user_events
            WHERE event_type = 'product_view'
              AND created_at > SYSDATE - 7
            GROUP BY user_id, product_id
        """)

        views = []
        for row in cursor:
            views.append({
                'user_id': row[0],
                'product_id': row[1],
                'count': row[2],
                'last_date': row[3].isoformat() if row[3] else None
            })

        cursor.close()

        with neo4j_driver.session() as session:
            # Insert purchases
            if purchases:
                session.run("""
                    UNWIND $purchases AS p
                    MERGE (u:User {id: p.user_id})
                    WITH u, p
                    MATCH (prod:Product {id: p.product_id})
                    MERGE (u)-[r:PURCHASED]->(prod)
                    SET r.count = p.count, r.last_purchase = p.last_date
                """, purchases=purchases)

            # Insert views
            if views:
                session.run("""
                    UNWIND $views AS v
                    MERGE (u:User {id: v.user_id})
                    WITH u, v
                    MATCH (prod:Product {id: v.product_id})
                    MERGE (u)-[r:VIEWED]->(prod)
                    SET r.count = v.count, r.last_viewed = v.last_date
                """, views=views)

        print(f"Synced {len(purchases)} purchases and {len(views)} views to Neo4j")
        context["ti"].xcom_push(key="interactions_synced", value=len(purchases) + len(views))

        oracle_conn.close()
        neo4j_driver.close()

    except Exception as e:
        print(f"Error syncing user interactions: {e}")
        context["ti"].xcom_push(key="interactions_synced", value=0)
        raise


def create_similarity_relationships(**context):
    """Create or update similarity relationships between products."""
    try:
        from neo4j import GraphDatabase

        neo4j_uri = f"bolt://{NEO4J_HOST}:{NEO4J_BOLT_PORT}"
        neo4j_driver = GraphDatabase.driver(neo4j_uri, auth=(NEO4J_USER, NEO4J_PASSWORD))

        with neo4j_driver.session() as session:
            # Create similarity based on category and price
            result = session.run("""
                MATCH (p1:Product), (p2:Product)
                WHERE p1.id < p2.id
                  AND p1.category_id = p2.category_id
                  AND abs(p1.price - p2.price) / ((p1.price + p2.price) / 2) < 0.2
                  AND NOT EXISTS((p1)-[:SIMILAR_TO]->(p2))
                WITH p1, p2,
                     0.7 + (0.3 * (1 - abs(p1.price - p2.price) / ((p1.price + p2.price) / 2))) AS score
                LIMIT 5000
                CREATE (p1)-[:SIMILAR_TO {score: score, source: 'category_price'}]->(p2)
                RETURN count(*) AS created
            """)
            created = result.single()["created"]
            print(f"Created {created} similarity relationships")

            # Create frequently bought together from co-purchases
            result = session.run("""
                MATCH (u:User)-[:PURCHASED]->(p1:Product)
                MATCH (u)-[:PURCHASED]->(p2:Product)
                WHERE p1.id < p2.id
                WITH p1, p2, COUNT(DISTINCT u) AS common_buyers
                WHERE common_buyers >= 3
                MERGE (p1)-[r:FREQUENTLY_BOUGHT_WITH]->(p2)
                SET r.count = common_buyers
                RETURN count(*) AS updated
            """)
            fbt_updated = result.single()["updated"]
            print(f"Updated {fbt_updated} frequently bought together relationships")

        context["ti"].xcom_push(key="relationships_created", value=created + fbt_updated)
        neo4j_driver.close()

    except Exception as e:
        print(f"Error creating similarity relationships: {e}")
        context["ti"].xcom_push(key="relationships_created", value=0)
        raise


def update_sync_metrics(**context):
    """Update sync metrics in Redis."""
    try:
        r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)

        products = context["ti"].xcom_pull(key="products_synced", task_ids="sync_products") or 0
        categories = context["ti"].xcom_pull(key="categories_synced", task_ids="sync_categories") or 0
        interactions = context["ti"].xcom_pull(key="interactions_synced", task_ids="sync_user_interactions") or 0
        relationships = context["ti"].xcom_pull(key="relationships_created", task_ids="create_similarity_relationships") or 0

        metrics = {
            "last_sync": datetime.now().isoformat(),
            "products_synced": str(products),
            "categories_synced": str(categories),
            "interactions_synced": str(interactions),
            "relationships_created": str(relationships),
            "status": "success"
        }

        for key, value in metrics.items():
            r.set(f"airflow:neo4j_sync:{key}", value)

        print(f"Updated Neo4j sync metrics: {metrics}")

    except Exception as e:
        print(f"Warning: Could not update Redis metrics: {e}")


with DAG(
    dag_id="florus_neo4j_sync",
    default_args=default_args,
    description="Sync product and user data from Oracle to Neo4j",
    schedule_interval="0 */4 * * *",  # Every 4 hours
    start_date=datetime(2024, 1, 1),
    catchup=False,
    tags=["florus", "neo4j", "sync", "graph"],
) as dag:

    # Check Neo4j health
    check_health = BranchPythonOperator(
        task_id="check_neo4j_health",
        python_callable=check_neo4j_health,
        provide_context=True,
    )

    # Skip sync if Neo4j is unhealthy
    skip_sync = EmptyOperator(
        task_id="skip_sync",
    )

    # Sync categories first (products depend on them)
    sync_categories = PythonOperator(
        task_id="sync_categories",
        python_callable=sync_categories_task,
        provide_context=True,
    )

    # Sync products
    sync_products = PythonOperator(
        task_id="sync_products",
        python_callable=sync_products_task,
        provide_context=True,
    )

    # Sync user interactions
    sync_interactions = PythonOperator(
        task_id="sync_user_interactions",
        python_callable=sync_user_interactions_task,
        provide_context=True,
    )

    # Create similarity relationships
    create_relationships = PythonOperator(
        task_id="create_similarity_relationships",
        python_callable=create_similarity_relationships,
        provide_context=True,
    )

    # Update metrics
    update_metrics = PythonOperator(
        task_id="update_sync_metrics",
        python_callable=update_sync_metrics,
        provide_context=True,
        trigger_rule=TriggerRule.ALL_DONE,
    )

    # Define task dependencies
    check_health >> [skip_sync, sync_categories]
    sync_categories >> sync_products >> sync_interactions >> create_relationships
    [skip_sync, create_relationships] >> update_metrics
