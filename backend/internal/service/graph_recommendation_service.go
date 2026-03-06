package service

import (
	"context"
	"log"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
	"github.com/florus/backend/pkg/neo4j"
)

// GraphRecommendationService provides graph-based recommendations using Neo4j
type GraphRecommendationService interface {
	// Check if Neo4j is available
	IsAvailable() bool

	// Get similar products based on graph relationships
	GetGraphSimilarProducts(ctx context.Context, productID uint, limit int) ([]dto.ProductResponse, error)

	// Get frequently bought together products
	GetFrequentlyBoughtTogether(ctx context.Context, productID uint, limit int) ([]dto.ProductResponse, error)

	// Get user-based collaborative filtering recommendations
	GetUserBasedRecommendations(ctx context.Context, userID uint, limit int) ([]dto.ProductResponse, error)

	// Get category-based recommendations
	GetCategoryRecommendations(ctx context.Context, categoryID uint, excludeProductID uint, limit int) ([]dto.ProductResponse, error)

	// Get serendipity recommendations (discovery of new categories)
	GetSerendipityRecommendations(ctx context.Context, userID uint, limit int) ([]dto.ProductResponse, error)

	// Record user view event in graph
	RecordView(ctx context.Context, userID, productID uint) error

	// Record user purchase event in graph
	RecordPurchase(ctx context.Context, userID, productID uint) error

	// Update co-viewed relationships for products in session
	UpdateCoViewedRelationships(ctx context.Context, productIDs []uint) error
}

type graphRecommendationService struct {
	neo4jClient *neo4j.Client
	productRepo repository.ProductRepository
	fallback    RecommendationService
}

// NewGraphRecommendationService creates a new graph-based recommendation service
func NewGraphRecommendationService(
	neo4jClient *neo4j.Client,
	productRepo repository.ProductRepository,
	fallback RecommendationService,
) GraphRecommendationService {
	return &graphRecommendationService{
		neo4jClient: neo4jClient,
		productRepo: productRepo,
		fallback:    fallback,
	}
}

// IsAvailable checks if Neo4j is connected and available
func (s *graphRecommendationService) IsAvailable() bool {
	return s.neo4jClient != nil
}

// GetGraphSimilarProducts returns products similar to the given product based on graph relationships
func (s *graphRecommendationService) GetGraphSimilarProducts(ctx context.Context, productID uint, limit int) ([]dto.ProductResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	// Check if Neo4j is available
	if s.neo4jClient == nil {
		log.Printf("[GRAPH] Neo4j not available, falling back to content-based similarity")
		return s.fallback.GetSimilarProducts(productID, limit)
	}

	// Get recommendations from Neo4j
	recs, err := s.neo4jClient.GetSimilarProducts(ctx, uint64(productID), limit)
	if err != nil {
		log.Printf("[GRAPH] Error getting similar products: %v, falling back", err)
		return s.fallback.GetSimilarProducts(productID, limit)
	}

	if len(recs) == 0 {
		log.Printf("[GRAPH] No similar products found for %d, falling back", productID)
		return s.fallback.GetSimilarProducts(productID, limit)
	}

	// Get product IDs from recommendations
	productIDs := make([]uint, len(recs))
	for i, rec := range recs {
		productIDs[i] = uint(rec.Product.ID)
	}

	// Fetch full product details from database
	products, err := s.productRepo.FindByIDs(productIDs)
	if err != nil {
		log.Printf("[GRAPH] Error fetching products: %v, falling back", err)
		return s.fallback.GetSimilarProducts(productID, limit)
	}

	return toGraphProductResponses(products), nil
}

// GetFrequentlyBoughtTogether returns products frequently bought with the given product
func (s *graphRecommendationService) GetFrequentlyBoughtTogether(ctx context.Context, productID uint, limit int) ([]dto.ProductResponse, error) {
	if limit <= 0 {
		limit = 6
	}

	// Check if Neo4j is available
	if s.neo4jClient == nil {
		log.Printf("[GRAPH] Neo4j not available, falling back to cross-sell")
		return s.fallback.GetCrossSellProducts(productID, limit)
	}

	// Get recommendations from Neo4j
	recs, err := s.neo4jClient.GetFrequentlyBoughtTogether(ctx, uint64(productID), limit)
	if err != nil {
		log.Printf("[GRAPH] Error getting frequently bought together: %v, falling back", err)
		return s.fallback.GetCrossSellProducts(productID, limit)
	}

	if len(recs) == 0 {
		log.Printf("[GRAPH] No frequently bought together found for %d, falling back", productID)
		return s.fallback.GetCrossSellProducts(productID, limit)
	}

	// Get product IDs from recommendations
	productIDs := make([]uint, len(recs))
	for i, rec := range recs {
		productIDs[i] = uint(rec.Product.ID)
	}

	// Fetch full product details from database
	products, err := s.productRepo.FindByIDs(productIDs)
	if err != nil {
		log.Printf("[GRAPH] Error fetching products: %v, falling back", err)
		return s.fallback.GetCrossSellProducts(productID, limit)
	}

	return toGraphProductResponses(products), nil
}

// GetUserBasedRecommendations returns recommendations based on similar user behavior
func (s *graphRecommendationService) GetUserBasedRecommendations(ctx context.Context, userID uint, limit int) ([]dto.ProductResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	// Check if Neo4j is available
	if s.neo4jClient == nil {
		log.Printf("[GRAPH] Neo4j not available, falling back to warm-start")
		return s.fallback.GetWarmStartRecommendations(userID, limit)
	}

	// Get recommendations from Neo4j
	recs, err := s.neo4jClient.GetUserBasedRecommendations(ctx, uint64(userID), limit)
	if err != nil {
		log.Printf("[GRAPH] Error getting user-based recommendations: %v, falling back", err)
		return s.fallback.GetWarmStartRecommendations(userID, limit)
	}

	if len(recs) == 0 {
		log.Printf("[GRAPH] No user-based recommendations for user %d, falling back", userID)
		return s.fallback.GetWarmStartRecommendations(userID, limit)
	}

	// Get product IDs from recommendations
	productIDs := make([]uint, len(recs))
	for i, rec := range recs {
		productIDs[i] = uint(rec.Product.ID)
	}

	// Fetch full product details from database
	products, err := s.productRepo.FindByIDs(productIDs)
	if err != nil {
		log.Printf("[GRAPH] Error fetching products: %v, falling back", err)
		return s.fallback.GetWarmStartRecommendations(userID, limit)
	}

	return toGraphProductResponses(products), nil
}

// GetCategoryRecommendations returns popular products in the same category
func (s *graphRecommendationService) GetCategoryRecommendations(ctx context.Context, categoryID uint, excludeProductID uint, limit int) ([]dto.ProductResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	// Check if Neo4j is available
	if s.neo4jClient == nil {
		log.Printf("[GRAPH] Neo4j not available, falling back to category-based")
		return s.fallback.GetPopularByCategory(categoryID, limit)
	}

	// Get recommendations from Neo4j
	recs, err := s.neo4jClient.GetCategoryRecommendations(ctx, uint64(categoryID), uint64(excludeProductID), limit)
	if err != nil {
		log.Printf("[GRAPH] Error getting category recommendations: %v, falling back", err)
		return s.fallback.GetPopularByCategory(categoryID, limit)
	}

	if len(recs) == 0 {
		log.Printf("[GRAPH] No category recommendations for category %d, falling back", categoryID)
		return s.fallback.GetPopularByCategory(categoryID, limit)
	}

	// Get product IDs from recommendations
	productIDs := make([]uint, len(recs))
	for i, rec := range recs {
		productIDs[i] = uint(rec.Product.ID)
	}

	// Fetch full product details from database
	products, err := s.productRepo.FindByIDs(productIDs)
	if err != nil {
		log.Printf("[GRAPH] Error fetching products: %v, falling back", err)
		return s.fallback.GetPopularByCategory(categoryID, limit)
	}

	return toGraphProductResponses(products), nil
}

// GetSerendipityRecommendations returns diverse recommendations for discovery
func (s *graphRecommendationService) GetSerendipityRecommendations(ctx context.Context, userID uint, limit int) ([]dto.ProductResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	// Check if Neo4j is available
	if s.neo4jClient == nil {
		log.Printf("[GRAPH] Neo4j not available, falling back to trending")
		return s.fallback.GetTrendingProducts(limit)
	}

	// Get recommendations from Neo4j
	recs, err := s.neo4jClient.GetSerendipityRecommendations(ctx, uint64(userID), limit)
	if err != nil {
		log.Printf("[GRAPH] Error getting serendipity recommendations: %v, falling back", err)
		return s.fallback.GetTrendingProducts(limit)
	}

	if len(recs) == 0 {
		log.Printf("[GRAPH] No serendipity recommendations for user %d, falling back", userID)
		return s.fallback.GetTrendingProducts(limit)
	}

	// Get product IDs from recommendations
	productIDs := make([]uint, len(recs))
	for i, rec := range recs {
		productIDs[i] = uint(rec.Product.ID)
	}

	// Fetch full product details from database
	products, err := s.productRepo.FindByIDs(productIDs)
	if err != nil {
		log.Printf("[GRAPH] Error fetching products: %v, falling back", err)
		return s.fallback.GetTrendingProducts(limit)
	}

	return toGraphProductResponses(products), nil
}

// RecordView records a user view event in the graph database
func (s *graphRecommendationService) RecordView(ctx context.Context, userID, productID uint) error {
	if s.neo4jClient == nil {
		return nil // Silent skip if Neo4j not available
	}

	if err := s.neo4jClient.RecordUserView(ctx, uint64(userID), uint64(productID)); err != nil {
		log.Printf("[GRAPH] Error recording view: %v", err)
		return err
	}

	return nil
}

// RecordPurchase records a user purchase event in the graph database
func (s *graphRecommendationService) RecordPurchase(ctx context.Context, userID, productID uint) error {
	if s.neo4jClient == nil {
		return nil // Silent skip if Neo4j not available
	}

	if err := s.neo4jClient.RecordUserPurchase(ctx, uint64(userID), uint64(productID)); err != nil {
		log.Printf("[GRAPH] Error recording purchase: %v", err)
		return err
	}

	return nil
}

// UpdateCoViewedRelationships updates co-viewed relationships for products in a session
func (s *graphRecommendationService) UpdateCoViewedRelationships(ctx context.Context, productIDs []uint) error {
	if s.neo4jClient == nil {
		return nil // Silent skip if Neo4j not available
	}

	// Convert to uint64
	ids := make([]uint64, len(productIDs))
	for i, id := range productIDs {
		ids[i] = uint64(id)
	}

	if err := s.neo4jClient.UpdateCoViewedRelationships(ctx, ids); err != nil {
		log.Printf("[GRAPH] Error updating co-viewed relationships: %v", err)
		return err
	}

	return nil
}

// Helper function
func toGraphProductResponses(products []models.Product) []dto.ProductResponse {
	responses := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		responses[i] = dto.ToProductResponse(&p)
	}
	return responses
}
