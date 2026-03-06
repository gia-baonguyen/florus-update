// Florus Neo4j Graph Schema Initialization
// Run this script after Neo4j starts to create constraints and indexes

// ============================================
// CONSTRAINTS (Unique Node IDs)
// ============================================

// Product nodes must have unique IDs
CREATE CONSTRAINT product_id IF NOT EXISTS FOR (p:Product) REQUIRE p.id IS UNIQUE;

// User nodes must have unique IDs
CREATE CONSTRAINT user_id IF NOT EXISTS FOR (u:User) REQUIRE u.id IS UNIQUE;

// Category nodes must have unique IDs
CREATE CONSTRAINT category_id IF NOT EXISTS FOR (c:Category) REQUIRE c.id IS UNIQUE;

// Brand nodes must have unique IDs
CREATE CONSTRAINT brand_id IF NOT EXISTS FOR (b:Brand) REQUIRE b.id IS UNIQUE;

// ============================================
// INDEXES (Performance Optimization)
// ============================================

// Index for product lookups by category
CREATE INDEX product_category IF NOT EXISTS FOR (p:Product) ON (p.category_id);

// Index for product lookups by rating (for quality filtering)
CREATE INDEX product_rating IF NOT EXISTS FOR (p:Product) ON (p.average_rating);

// Index for product lookups by price range
CREATE INDEX product_price IF NOT EXISTS FOR (p:Product) ON (p.price);

// Index for product lookups by brand
CREATE INDEX product_brand IF NOT EXISTS FOR (p:Product) ON (p.brand);

// Index for user activity tracking
CREATE INDEX user_active IF NOT EXISTS FOR (u:User) ON (u.last_active);

// Composite index for relationship score lookups
CREATE INDEX rel_score IF NOT EXISTS FOR ()-[r:SIMILAR_TO]-() ON (r.score);

// ============================================
// NODE LABELS SUMMARY
// ============================================
// :Product - Products in the catalog
//   Properties: id, name, price, category_id, brand, average_rating, view_count, purchase_count, created_at
//
// :User - User accounts
//   Properties: id, name, email, preferences, last_active
//
// :Category - Product categories
//   Properties: id, name, parent_id
//
// :Brand - Product brands
//   Properties: id, name

// ============================================
// RELATIONSHIP TYPES SUMMARY
// ============================================
// (Product)-[:SIMILAR_TO {score: float}]->(Product)
//   - Content-based similarity (category, price, brand)
//   - ML-based similarity from Spark ALS
//
// (Product)-[:FREQUENTLY_BOUGHT_WITH {count: int, last_updated: datetime}]->(Product)
//   - Co-purchase associations from order history
//
// (Product)-[:CO_VIEWED {count: int, session_count: int}]->(Product)
//   - Products viewed together in same session
//
// (Product)-[:BELONGS_TO]->(Category)
//   - Product to category relationship
//
// (Product)-[:MADE_BY]->(Brand)
//   - Product to brand relationship
//
// (User)-[:PURCHASED {timestamp: datetime, order_id: int, quantity: int}]->(Product)
//   - User purchase history
//
// (User)-[:VIEWED {count: int, last_viewed: datetime}]->(Product)
//   - User product view history
//
// (User)-[:RATED {score: float, timestamp: datetime}]->(Product)
//   - User product ratings
//
// (User)-[:ADDED_TO_CART {count: int, last_added: datetime}]->(Product)
//   - User cart activity

RETURN "Schema initialization completed successfully!" AS message;
