package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
	"github.com/redis/go-redis/v9"
)

// SparkRecommendation represents a recommendation from Spark
type SparkRecommendation struct {
	ProductID uint    `json:"product_id"`
	Score     float64 `json:"score"`
}

// SparkRecommendationService provides ML-based recommendations from Spark
type SparkRecommendationService interface {
	// Get Spark-based recommendations for a user
	GetSparkUserRecommendations(userID uint, limit int) ([]dto.ProductResponse, error)

	// Get Spark-based similar products
	GetSparkSimilarProducts(productID uint, limit int) ([]dto.ProductResponse, error)

	// Get last update time
	GetLastUpdateTime() (time.Time, error)

	// Check if Spark recommendations are available
	IsAvailable() bool
}

type sparkRecommendationService struct {
	redisClient *redis.Client
	productRepo repository.ProductRepository
	fallback    RecommendationService
}

// NewSparkRecommendationService creates a new Spark recommendation service
func NewSparkRecommendationService(
	redisClient *redis.Client,
	productRepo repository.ProductRepository,
	fallback RecommendationService,
) SparkRecommendationService {
	return &sparkRecommendationService{
		redisClient: redisClient,
		productRepo: productRepo,
		fallback:    fallback,
	}
}

// IsAvailable checks if Spark recommendations are available
func (s *sparkRecommendationService) IsAvailable() bool {
	if s.redisClient == nil {
		return false
	}

	ctx := context.Background()
	lastUpdate, err := s.redisClient.Get(ctx, "spark:recommendations:last_updated").Result()
	if err != nil || lastUpdate == "" {
		return false
	}

	// Check if recommendations are not too old (7 days)
	updateTime, err := time.Parse(time.RFC3339, lastUpdate)
	if err != nil {
		return false
	}

	return time.Since(updateTime) < 7*24*time.Hour
}

// GetLastUpdateTime returns when Spark recommendations were last updated
func (s *sparkRecommendationService) GetLastUpdateTime() (time.Time, error) {
	if s.redisClient == nil {
		return time.Time{}, fmt.Errorf("redis not available")
	}

	ctx := context.Background()
	lastUpdate, err := s.redisClient.Get(ctx, "spark:recommendations:last_updated").Result()
	if err != nil {
		return time.Time{}, err
	}

	return time.Parse(time.RFC3339, lastUpdate)
}

// GetSparkUserRecommendations returns ML-based recommendations for a user
func (s *sparkRecommendationService) GetSparkUserRecommendations(userID uint, limit int) ([]dto.ProductResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	// Check if Redis is available
	if s.redisClient == nil {
		log.Printf("[SPARK] Redis not available, falling back to rule-based recommendations")
		return s.fallback.GetWarmStartRecommendations(userID, limit)
	}

	ctx := context.Background()
	key := fmt.Sprintf("spark:recommendations:user:%d", userID)

	// Get recommendations from Redis
	data, err := s.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Printf("[SPARK] No recommendations for user %d, falling back", userID)
		return s.fallback.GetWarmStartRecommendations(userID, limit)
	}
	if err != nil {
		log.Printf("[SPARK] Redis error: %v, falling back", err)
		return s.fallback.GetWarmStartRecommendations(userID, limit)
	}

	// Parse recommendations
	var recs []SparkRecommendation
	if err := json.Unmarshal([]byte(data), &recs); err != nil {
		log.Printf("[SPARK] Parse error: %v, falling back", err)
		return s.fallback.GetWarmStartRecommendations(userID, limit)
	}

	// Limit results
	if len(recs) > limit {
		recs = recs[:limit]
	}

	// Get product IDs
	productIDs := make([]uint, len(recs))
	for i, rec := range recs {
		productIDs[i] = rec.ProductID
	}

	// Fetch product details
	products, err := s.productRepo.FindByIDs(productIDs)
	if err != nil {
		return s.fallback.GetWarmStartRecommendations(userID, limit)
	}

	// Convert to response
	return toProductResponses(products), nil
}

// GetSparkSimilarProducts returns ML-based similar products
func (s *sparkRecommendationService) GetSparkSimilarProducts(productID uint, limit int) ([]dto.ProductResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	// Check if Redis is available
	if s.redisClient == nil {
		log.Printf("[SPARK] Redis not available, falling back to content-based similarity")
		return s.fallback.GetSimilarProducts(productID, limit)
	}

	ctx := context.Background()
	key := fmt.Sprintf("spark:similar:product:%d", productID)

	// Get similar products from Redis
	data, err := s.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Printf("[SPARK] No similar products for %d, falling back", productID)
		return s.fallback.GetSimilarProducts(productID, limit)
	}
	if err != nil {
		log.Printf("[SPARK] Redis error: %v, falling back", err)
		return s.fallback.GetSimilarProducts(productID, limit)
	}

	// Parse similarities
	var sims []SparkRecommendation
	if err := json.Unmarshal([]byte(data), &sims); err != nil {
		log.Printf("[SPARK] Parse error: %v, falling back", err)
		return s.fallback.GetSimilarProducts(productID, limit)
	}

	// Limit results
	if len(sims) > limit {
		sims = sims[:limit]
	}

	// Get product IDs
	productIDs := make([]uint, len(sims))
	for i, sim := range sims {
		productIDs[i] = sim.ProductID
	}

	// Fetch product details
	products, err := s.productRepo.FindByIDs(productIDs)
	if err != nil {
		return s.fallback.GetSimilarProducts(productID, limit)
	}

	return toProductResponses(products), nil
}

// Helper function
func toProductResponses(products []models.Product) []dto.ProductResponse {
	responses := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		responses[i] = dto.ToProductResponse(&p)
	}
	return responses
}
