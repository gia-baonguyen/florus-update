package service

import (
	"time"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
)

type RecommendationService interface {
	// Cold-start recommendations (for new users)
	GetColdStartRecommendations(limit int) ([]dto.ProductResponse, error)

	// Warm-start recommendations (for users with history)
	GetWarmStartRecommendations(userID uint, limit int) ([]dto.ProductResponse, error)

	// Product-based recommendations
	GetSimilarProducts(productID uint, limit int) ([]dto.ProductResponse, error)
	GetCoViewedProducts(productID uint, limit int) ([]dto.ProductResponse, error)
	GetCrossSellProducts(productID uint, limit int) ([]dto.ProductResponse, error)

	// Trending and popular
	GetTrendingProducts(limit int) ([]dto.ProductResponse, error)
	GetPopularByCategory(categoryID uint, limit int) ([]dto.ProductResponse, error)

	// Personalized
	GetPersonalizedRecommendations(userID uint, limit int) ([]dto.ProductResponse, error)
}

type recommendationService struct {
	productRepo repository.ProductRepository
	eventRepo   repository.EventRepository
}

func NewRecommendationService(productRepo repository.ProductRepository, eventRepo repository.EventRepository) RecommendationService {
	return &recommendationService{
		productRepo: productRepo,
		eventRepo:   eventRepo,
	}
}

// GetColdStartRecommendations returns popular products for new users
// Strategy: Most viewed/purchased products in last 30 days
func (s *recommendationService) GetColdStartRecommendations(limit int) ([]dto.ProductResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	// Get popular products from last 30 days
	since := time.Now().AddDate(0, 0, -30)
	productIDs, err := s.eventRepo.GetPopularProducts(limit*2, since)

	// If not enough event data, fall back to top-rated products
	if err != nil || len(productIDs) < limit {
		return s.getTopRatedProducts(limit)
	}

	// Limit results
	if len(productIDs) > limit {
		productIDs = productIDs[:limit]
	}

	return s.getProductsByIDs(productIDs)
}

// GetWarmStartRecommendations returns personalized recommendations for active users
// Strategy: Collaborative filtering + Content-based hybrid
func (s *recommendationService) GetWarmStartRecommendations(userID uint, limit int) ([]dto.ProductResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	// Get user's behavior profile
	profile, err := s.eventRepo.GetUserBehaviorProfile(userID)
	if err != nil || profile.TotalEvents < 3 {
		// Not enough data, fall back to cold-start
		return s.GetColdStartRecommendations(limit)
	}

	// Strategy 1: Get products similar to what user has viewed/purchased
	var recommendedIDs []uint

	// Get products from user's preferred categories
	for _, categoryID := range profile.PreferredCategories {
		catProducts, err := s.getPopularInCategory(categoryID, 5)
		if err == nil {
			recommendedIDs = append(recommendedIDs, catProducts...)
		}
	}

	// Get co-viewed products from user's recent views
	for _, productID := range profile.RecentProducts[:min(5, len(profile.RecentProducts))] {
		coViewed, err := s.eventRepo.GetCoViewedProducts(productID, 3)
		if err == nil {
			recommendedIDs = append(recommendedIDs, coViewed...)
		}
	}

	// Remove duplicates and products user has already viewed
	recommendedIDs = s.removeDuplicatesAndViewed(recommendedIDs, profile.RecentProducts)

	// Limit results
	if len(recommendedIDs) > limit {
		recommendedIDs = recommendedIDs[:limit]
	}

	// If not enough recommendations, supplement with popular items
	if len(recommendedIDs) < limit {
		popular, _ := s.GetColdStartRecommendations(limit - len(recommendedIDs))
		products, _ := s.getProductsByIDs(recommendedIDs)
		return append(products, popular...), nil
	}

	return s.getProductsByIDs(recommendedIDs)
}

// GetSimilarProducts returns products similar to the given product
// Strategy: Same category + similar price range + same brand (content-based)
func (s *recommendationService) GetSimilarProducts(productID uint, limit int) ([]dto.ProductResponse, error) {
	if limit <= 0 {
		limit = 6
	}

	// Get the source product
	sourceProduct, err := s.productRepo.FindByID(productID)
	if err != nil {
		return []dto.ProductResponse{}, nil
	}

	// Find similar products based on content
	similarProducts, err := s.productRepo.FindSimilar(productID, sourceProduct.CategoryID, sourceProduct.Brand, sourceProduct.Price, limit)
	if err != nil {
		return []dto.ProductResponse{}, nil
	}

	return s.toProductResponses(similarProducts), nil
}

// GetCoViewedProducts returns products often viewed together with the given product
// Strategy: Collaborative filtering - "Users who viewed this also viewed"
func (s *recommendationService) GetCoViewedProducts(productID uint, limit int) ([]dto.ProductResponse, error) {
	if limit <= 0 {
		limit = 6
	}

	// Get co-viewed products from user events
	productIDs, err := s.eventRepo.GetCoViewedProducts(productID, limit)
	if err != nil || len(productIDs) == 0 {
		// Fall back to similar products
		return s.GetSimilarProducts(productID, limit)
	}

	return s.getProductsByIDs(productIDs)
}

// GetCrossSellProducts returns products frequently bought together
// Strategy: "Frequently bought together" from purchase history
func (s *recommendationService) GetCrossSellProducts(productID uint, limit int) ([]dto.ProductResponse, error) {
	if limit <= 0 {
		limit = 4
	}

	// Get frequently bought together from user events
	productIDs, err := s.eventRepo.GetFrequentlyBoughtTogether(productID, limit)
	if err != nil || len(productIDs) == 0 {
		// Fall back to co-viewed
		return s.GetCoViewedProducts(productID, limit)
	}

	return s.getProductsByIDs(productIDs)
}

// GetTrendingProducts returns trending products (high activity recently)
// Strategy: Most interactions in last 7 days
func (s *recommendationService) GetTrendingProducts(limit int) ([]dto.ProductResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	// Get popular products from last 7 days
	since := time.Now().AddDate(0, 0, -7)
	productIDs, err := s.eventRepo.GetPopularProducts(limit, since)
	if err != nil || len(productIDs) == 0 {
		return s.GetColdStartRecommendations(limit)
	}

	return s.getProductsByIDs(productIDs)
}

// GetPopularByCategory returns popular products in a category
func (s *recommendationService) GetPopularByCategory(categoryID uint, limit int) ([]dto.ProductResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	productIDs, err := s.getPopularInCategory(categoryID, limit)
	if err != nil || len(productIDs) == 0 {
		// Fall back to top-rated in category
		return s.getTopRatedInCategory(categoryID, limit)
	}

	return s.getProductsByIDs(productIDs)
}

// GetPersonalizedRecommendations returns highly personalized recommendations
// Strategy: Hybrid of collaborative + content-based + trending
func (s *recommendationService) GetPersonalizedRecommendations(userID uint, limit int) ([]dto.ProductResponse, error) {
	return s.GetWarmStartRecommendations(userID, limit)
}

// Helper functions

func (s *recommendationService) getProductsByIDs(ids []uint) ([]dto.ProductResponse, error) {
	if len(ids) == 0 {
		return []dto.ProductResponse{}, nil
	}

	products, err := s.productRepo.FindByIDs(ids)
	if err != nil {
		return nil, err
	}

	return s.toProductResponses(products), nil
}

func (s *recommendationService) getTopRatedProducts(limit int) ([]dto.ProductResponse, error) {
	products, err := s.productRepo.FindTopRated(limit)
	if err != nil {
		return nil, err
	}
	return s.toProductResponses(products), nil
}

func (s *recommendationService) getTopRatedInCategory(categoryID uint, limit int) ([]dto.ProductResponse, error) {
	products, err := s.productRepo.FindTopRatedInCategory(categoryID, limit)
	if err != nil {
		return nil, err
	}
	return s.toProductResponses(products), nil
}

func (s *recommendationService) getPopularInCategory(categoryID uint, limit int) ([]uint, error) {
	// This would need a custom query in event repository
	// For now, return empty and let caller handle fallback
	return []uint{}, nil
}

func (s *recommendationService) removeDuplicatesAndViewed(ids []uint, viewedIDs []uint) []uint {
	seen := make(map[uint]bool)

	// Mark viewed products
	for _, id := range viewedIDs {
		seen[id] = true
	}

	var result []uint
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}

	return result
}

func (s *recommendationService) toProductResponses(products []models.Product) []dto.ProductResponse {
	var responses []dto.ProductResponse
	for _, p := range products {
		responses = append(responses, dto.ToProductResponse(&p))
	}
	return responses
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
