package neo4j

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Client struct {
	driver neo4j.DriverWithContext
	uri    string
}

type Config struct {
	URI      string
	User     string
	Password string
}

func NewClient(cfg Config) (*Client, error) {
	driver, err := neo4j.NewDriverWithContext(
		cfg.URI,
		neo4j.BasicAuth(cfg.User, cfg.Password, ""),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	// Verify connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("failed to verify Neo4j connectivity: %w", err)
	}

	log.Printf("Connected to Neo4j at %s", cfg.URI)

	return &Client{
		driver: driver,
		uri:    cfg.URI,
	}, nil
}

func (c *Client) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.driver.Close(ctx)
}

func (c *Client) Driver() neo4j.DriverWithContext {
	return c.driver
}

func (c *Client) ExecuteRead(ctx context.Context, work func(tx neo4j.ManagedTransaction) (any, error)) (any, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	return session.ExecuteRead(ctx, work)
}

func (c *Client) ExecuteWrite(ctx context.Context, work func(tx neo4j.ManagedTransaction) (any, error)) (any, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	return session.ExecuteWrite(ctx, work)
}

// Product represents a product node in the graph
type Product struct {
	ID            uint64  `json:"id"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	CategoryID    uint64  `json:"category_id"`
	Brand         string  `json:"brand"`
	AverageRating float64 `json:"average_rating"`
	ViewCount     int64   `json:"view_count"`
	PurchaseCount int64   `json:"purchase_count"`
}

// Recommendation represents a recommendation result
type Recommendation struct {
	Product *Product `json:"product"`
	Score   float64  `json:"score"`
	Reason  string   `json:"reason"`
}

// GetSimilarProducts returns similar products based on graph relationships
func (c *Client) GetSimilarProducts(ctx context.Context, productID uint64, limit int) ([]Recommendation, error) {
	result, err := c.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (p:Product {id: $productId})-[r:SIMILAR_TO]->(similar:Product)
			RETURN similar.id AS id, similar.name AS name, similar.price AS price,
				   similar.category_id AS categoryId, similar.brand AS brand,
				   similar.average_rating AS avgRating, similar.view_count AS views,
				   similar.purchase_count AS purchases, r.score AS score
			ORDER BY r.score DESC
			LIMIT $limit
		`
		result, err := tx.Run(ctx, query, map[string]any{
			"productId": productID,
			"limit":     limit,
		})
		if err != nil {
			return nil, err
		}

		var recommendations []Recommendation
		for result.Next(ctx) {
			record := result.Record()
			rec := Recommendation{
				Product: &Product{
					ID:            uint64(record.Values[0].(int64)),
					Name:          record.Values[1].(string),
					Price:         record.Values[2].(float64),
					CategoryID:    uint64(record.Values[3].(int64)),
					Brand:         record.Values[4].(string),
					AverageRating: record.Values[5].(float64),
					ViewCount:     record.Values[6].(int64),
					PurchaseCount: record.Values[7].(int64),
				},
				Score:  record.Values[8].(float64),
				Reason: "similar_product",
			}
			recommendations = append(recommendations, rec)
		}
		return recommendations, nil
	})

	if err != nil {
		return nil, err
	}
	return result.([]Recommendation), nil
}

// GetFrequentlyBoughtTogether returns products frequently bought with the given product
func (c *Client) GetFrequentlyBoughtTogether(ctx context.Context, productID uint64, limit int) ([]Recommendation, error) {
	result, err := c.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (p:Product {id: $productId})-[r:FREQUENTLY_BOUGHT_WITH]->(bought:Product)
			RETURN bought.id AS id, bought.name AS name, bought.price AS price,
				   bought.category_id AS categoryId, bought.brand AS brand,
				   bought.average_rating AS avgRating, bought.view_count AS views,
				   bought.purchase_count AS purchases, r.count AS count
			ORDER BY r.count DESC
			LIMIT $limit
		`
		result, err := tx.Run(ctx, query, map[string]any{
			"productId": productID,
			"limit":     limit,
		})
		if err != nil {
			return nil, err
		}

		var recommendations []Recommendation
		for result.Next(ctx) {
			record := result.Record()
			count := record.Values[8].(int64)
			rec := Recommendation{
				Product: &Product{
					ID:            uint64(record.Values[0].(int64)),
					Name:          record.Values[1].(string),
					Price:         record.Values[2].(float64),
					CategoryID:    uint64(record.Values[3].(int64)),
					Brand:         record.Values[4].(string),
					AverageRating: record.Values[5].(float64),
					ViewCount:     record.Values[6].(int64),
					PurchaseCount: record.Values[7].(int64),
				},
				Score:  float64(count) / 100.0,
				Reason: "frequently_bought_together",
			}
			recommendations = append(recommendations, rec)
		}
		return recommendations, nil
	})

	if err != nil {
		return nil, err
	}
	return result.([]Recommendation), nil
}

// GetUserBasedRecommendations returns recommendations based on similar user behavior
func (c *Client) GetUserBasedRecommendations(ctx context.Context, userID uint64, limit int) ([]Recommendation, error) {
	result, err := c.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			// Find products viewed/purchased by similar users (collaborative filtering)
			MATCH (u:User {id: $userId})-[:PURCHASED|VIEWED]->(p:Product)<-[:PURCHASED|VIEWED]-(other:User)
			MATCH (other)-[:PURCHASED]->(rec:Product)
			WHERE NOT (u)-[:PURCHASED]->(rec) AND NOT (u)-[:VIEWED]->(rec)
			WITH rec, COUNT(DISTINCT other) AS commonUsers, AVG(rec.average_rating) AS avgRating
			RETURN rec.id AS id, rec.name AS name, rec.price AS price,
				   rec.category_id AS categoryId, rec.brand AS brand,
				   avgRating, rec.view_count AS views, rec.purchase_count AS purchases,
				   commonUsers
			ORDER BY commonUsers DESC, avgRating DESC
			LIMIT $limit
		`
		result, err := tx.Run(ctx, query, map[string]any{
			"userId": userID,
			"limit":  limit,
		})
		if err != nil {
			return nil, err
		}

		var recommendations []Recommendation
		for result.Next(ctx) {
			record := result.Record()
			commonUsers := record.Values[8].(int64)
			rec := Recommendation{
				Product: &Product{
					ID:            uint64(record.Values[0].(int64)),
					Name:          record.Values[1].(string),
					Price:         record.Values[2].(float64),
					CategoryID:    uint64(record.Values[3].(int64)),
					Brand:         record.Values[4].(string),
					AverageRating: record.Values[5].(float64),
					ViewCount:     record.Values[6].(int64),
					PurchaseCount: record.Values[7].(int64),
				},
				Score:  float64(commonUsers),
				Reason: "user_collaborative_filtering",
			}
			recommendations = append(recommendations, rec)
		}
		return recommendations, nil
	})

	if err != nil {
		return nil, err
	}
	return result.([]Recommendation), nil
}

// GetCategoryRecommendations returns popular products in the same category
func (c *Client) GetCategoryRecommendations(ctx context.Context, categoryID uint64, excludeProductID uint64, limit int) ([]Recommendation, error) {
	result, err := c.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (p:Product)-[:BELONGS_TO]->(c:Category {id: $categoryId})
			WHERE p.id <> $excludeId
			RETURN p.id AS id, p.name AS name, p.price AS price,
				   p.category_id AS categoryId, p.brand AS brand,
				   p.average_rating AS avgRating, p.view_count AS views,
				   p.purchase_count AS purchases
			ORDER BY p.purchase_count DESC, p.average_rating DESC
			LIMIT $limit
		`
		result, err := tx.Run(ctx, query, map[string]any{
			"categoryId": categoryID,
			"excludeId":  excludeProductID,
			"limit":      limit,
		})
		if err != nil {
			return nil, err
		}

		var recommendations []Recommendation
		for result.Next(ctx) {
			record := result.Record()
			rec := Recommendation{
				Product: &Product{
					ID:            uint64(record.Values[0].(int64)),
					Name:          record.Values[1].(string),
					Price:         record.Values[2].(float64),
					CategoryID:    uint64(record.Values[3].(int64)),
					Brand:         record.Values[4].(string),
					AverageRating: record.Values[5].(float64),
					ViewCount:     record.Values[6].(int64),
					PurchaseCount: record.Values[7].(int64),
				},
				Score:  float64(record.Values[7].(int64)),
				Reason: "popular_in_category",
			}
			recommendations = append(recommendations, rec)
		}
		return recommendations, nil
	})

	if err != nil {
		return nil, err
	}
	return result.([]Recommendation), nil
}

// GetSerendipityRecommendations returns diverse recommendations for discovery
func (c *Client) GetSerendipityRecommendations(ctx context.Context, userID uint64, limit int) ([]Recommendation, error) {
	result, err := c.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			// Get categories user hasn't explored much
			MATCH (u:User {id: $userId})-[:PURCHASED|VIEWED]->(p:Product)-[:BELONGS_TO]->(userCat:Category)
			WITH u, COLLECT(DISTINCT userCat.id) AS userCategories

			// Find popular products in different categories
			MATCH (rec:Product)-[:BELONGS_TO]->(cat:Category)
			WHERE NOT cat.id IN userCategories
			  AND rec.average_rating >= 4.0
			  AND rec.purchase_count > 10
			WITH rec, rand() AS randomOrder
			RETURN rec.id AS id, rec.name AS name, rec.price AS price,
				   rec.category_id AS categoryId, rec.brand AS brand,
				   rec.average_rating AS avgRating, rec.view_count AS views,
				   rec.purchase_count AS purchases
			ORDER BY randomOrder
			LIMIT $limit
		`
		result, err := tx.Run(ctx, query, map[string]any{
			"userId": userID,
			"limit":  limit,
		})
		if err != nil {
			return nil, err
		}

		var recommendations []Recommendation
		for result.Next(ctx) {
			record := result.Record()
			rec := Recommendation{
				Product: &Product{
					ID:            uint64(record.Values[0].(int64)),
					Name:          record.Values[1].(string),
					Price:         record.Values[2].(float64),
					CategoryID:    uint64(record.Values[3].(int64)),
					Brand:         record.Values[4].(string),
					AverageRating: record.Values[5].(float64),
					ViewCount:     record.Values[6].(int64),
					PurchaseCount: record.Values[7].(int64),
				},
				Score:  record.Values[5].(float64),
				Reason: "serendipity_discovery",
			}
			recommendations = append(recommendations, rec)
		}
		return recommendations, nil
	})

	if err != nil {
		return nil, err
	}
	return result.([]Recommendation), nil
}

// RecordUserView records a user view event in the graph
func (c *Client) RecordUserView(ctx context.Context, userID, productID uint64) error {
	_, err := c.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MERGE (u:User {id: $userId})
			MERGE (p:Product {id: $productId})
			MERGE (u)-[r:VIEWED]->(p)
			SET r.count = COALESCE(r.count, 0) + 1,
				r.last_viewed = datetime()
		`
		_, err := tx.Run(ctx, query, map[string]any{
			"userId":    userID,
			"productId": productID,
		})
		return nil, err
	})
	return err
}

// RecordUserPurchase records a user purchase event in the graph
func (c *Client) RecordUserPurchase(ctx context.Context, userID, productID uint64) error {
	_, err := c.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MERGE (u:User {id: $userId})
			MERGE (p:Product {id: $productId})
			MERGE (u)-[r:PURCHASED]->(p)
			SET r.count = COALESCE(r.count, 0) + 1,
				r.last_purchase = datetime()
		`
		_, err := tx.Run(ctx, query, map[string]any{
			"userId":    userID,
			"productId": productID,
		})
		return nil, err
	})
	return err
}

// UpdateCoViewedRelationships updates co-viewed relationships between products
func (c *Client) UpdateCoViewedRelationships(ctx context.Context, productIDs []uint64) error {
	if len(productIDs) < 2 {
		return nil
	}

	_, err := c.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			UNWIND $productIds AS pid1
			UNWIND $productIds AS pid2
			WITH pid1, pid2 WHERE pid1 < pid2
			MATCH (p1:Product {id: pid1}), (p2:Product {id: pid2})
			MERGE (p1)-[r:CO_VIEWED]->(p2)
			SET r.count = COALESCE(r.count, 0) + 1
		`
		_, err := tx.Run(ctx, query, map[string]any{
			"productIds": productIDs,
		})
		return nil, err
	})
	return err
}
