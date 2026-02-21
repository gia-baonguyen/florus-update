package service

import (
	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
)

// Static recommendation mappings (from frontend data)
var (
	// Cold-start: Popular products for new users
	coldStartProductIDs = []uint{1, 2, 7, 8, 9, 3}

	// Warm-start: Personalized products for active users
	warmStartProductIDs = []uint{2, 5, 10, 13}

	// Similar products (content-based)
	similarProducts = map[uint][]uint{
		1:  {13, 9, 11},  // Serum -> Eyeshadow, Perfume, Mask
		2:  {4, 10, 6},   // Moisturizer -> Eye Cream, Sunscreen, Cleanser
		3:  {8, 7, 9},    // Lipstick -> Blush, Mascara, Perfume
		4:  {2, 11, 6},   // Eye Cream -> Moisturizer, Mask, Cleanser
		5:  {8, 3, 7},    // Foundation -> Blush, Lipstick, Mascara
		6:  {11, 2, 10},  // Cleanser -> Mask, Moisturizer, Sunscreen
		7:  {5, 3, 13},   // Mascara -> Foundation, Lipstick, Eyeshadow
		8:  {3, 5, 7},    // Blush -> Lipstick, Foundation, Mascara
		9:  {3, 8, 12},   // Perfume -> Lipstick, Blush, Nail Polish
		10: {2, 5, 6},    // Sunscreen -> Moisturizer, Foundation, Cleanser
		11: {6, 2, 1},    // Mask -> Cleanser, Moisturizer, Serum
		12: {3, 9, 8},    // Nail Polish -> Lipstick, Perfume, Blush
		13: {7, 3, 8},    // Eyeshadow -> Mascara, Lipstick, Blush
	}

	// Co-viewed products (collaborative filtering)
	coViewedProducts = map[uint][]uint{
		1:  {2, 6, 11},
		2:  {1, 6, 4},
		3:  {8, 7, 5},
		4:  {2, 1, 6},
		5:  {3, 8, 7},
		6:  {2, 1, 11},
		7:  {3, 5, 13},
		8:  {3, 5, 7},
		9:  {3, 12, 8},
		10: {2, 5, 6},
		11: {6, 1, 2},
		12: {3, 9, 8},
		13: {7, 3, 5},
	}

	// Cross-sell products (frequently bought together)
	crossSellProducts = map[uint][]uint{
		1:  {6, 2},
		2:  {1, 6, 10},
		3:  {8, 12},
		5:  {10, 8},
		6:  {2, 11},
		7:  {5, 13},
		11: {6, 1},
		13: {7, 5},
	}
)

type RecommendationService interface {
	GetColdStartRecommendations(limit int) ([]dto.ProductResponse, error)
	GetWarmStartRecommendations(limit int) ([]dto.ProductResponse, error)
	GetSimilarProducts(productID uint, limit int) ([]dto.ProductResponse, error)
	GetCoViewedProducts(productID uint, limit int) ([]dto.ProductResponse, error)
	GetCrossSellProducts(productID uint, limit int) ([]dto.ProductResponse, error)
}

type recommendationService struct {
	productRepo repository.ProductRepository
}

func NewRecommendationService(productRepo repository.ProductRepository) RecommendationService {
	return &recommendationService{
		productRepo: productRepo,
	}
}

func (s *recommendationService) GetColdStartRecommendations(limit int) ([]dto.ProductResponse, error) {
	ids := coldStartProductIDs
	if limit > 0 && limit < len(ids) {
		ids = ids[:limit]
	}

	products, err := s.productRepo.FindByIDs(ids)
	if err != nil {
		return nil, err
	}

	return s.toProductResponses(products), nil
}

func (s *recommendationService) GetWarmStartRecommendations(limit int) ([]dto.ProductResponse, error) {
	ids := warmStartProductIDs
	if limit > 0 && limit < len(ids) {
		ids = ids[:limit]
	}

	products, err := s.productRepo.FindByIDs(ids)
	if err != nil {
		return nil, err
	}

	return s.toProductResponses(products), nil
}

func (s *recommendationService) GetSimilarProducts(productID uint, limit int) ([]dto.ProductResponse, error) {
	ids, ok := similarProducts[productID]
	if !ok {
		return []dto.ProductResponse{}, nil
	}

	if limit > 0 && limit < len(ids) {
		ids = ids[:limit]
	}

	products, err := s.productRepo.FindByIDs(ids)
	if err != nil {
		return nil, err
	}

	return s.toProductResponses(products), nil
}

func (s *recommendationService) GetCoViewedProducts(productID uint, limit int) ([]dto.ProductResponse, error) {
	ids, ok := coViewedProducts[productID]
	if !ok {
		return []dto.ProductResponse{}, nil
	}

	if limit > 0 && limit < len(ids) {
		ids = ids[:limit]
	}

	products, err := s.productRepo.FindByIDs(ids)
	if err != nil {
		return nil, err
	}

	return s.toProductResponses(products), nil
}

func (s *recommendationService) GetCrossSellProducts(productID uint, limit int) ([]dto.ProductResponse, error) {
	ids, ok := crossSellProducts[productID]
	if !ok {
		return []dto.ProductResponse{}, nil
	}

	if limit > 0 && limit < len(ids) {
		ids = ids[:limit]
	}

	products, err := s.productRepo.FindByIDs(ids)
	if err != nil {
		return nil, err
	}

	return s.toProductResponses(products), nil
}

func (s *recommendationService) toProductResponses(products []models.Product) []dto.ProductResponse {
	var responses []dto.ProductResponse
	for _, p := range products {
		responses = append(responses, dto.ToProductResponse(&p))
	}
	return responses
}
