package service

import (
	"context"
	"log"

	"github.com/florus/backend/internal/dto"
	"github.com/redis/go-redis/v9"
)

// cachedRecommendationService wraps RecommendationService with caching
type cachedRecommendationService struct {
	base  RecommendationService
	cache CacheService
}

// NewCachedRecommendationService creates a new cached recommendation service
func NewCachedRecommendationService(base RecommendationService, cache CacheService) RecommendationService {
	return &cachedRecommendationService{
		base:  base,
		cache: cache,
	}
}

func (s *cachedRecommendationService) GetColdStartRecommendations(limit int) ([]dto.ProductResponse, error) {
	ctx := context.Background()
	key := s.cache.GetColdStartKey(limit)

	// Try cache first
	var cached []dto.ProductResponse
	if err := s.cache.GetJSON(ctx, key, &cached); err == nil {
		log.Printf("[CACHE HIT] cold-start recommendations (limit=%d)", limit)
		return cached, nil
	} else if err != redis.Nil {
		log.Printf("[CACHE ERROR] cold-start: %v", err)
	}

	// Cache miss - get from base service
	result, err := s.base.GetColdStartRecommendations(limit)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := s.cache.SetJSON(ctx, key, result, CacheTTLColdStart); err != nil {
		log.Printf("[CACHE ERROR] failed to cache cold-start: %v", err)
	} else {
		log.Printf("[CACHE SET] cold-start recommendations (limit=%d)", limit)
	}

	return result, nil
}

func (s *cachedRecommendationService) GetWarmStartRecommendations(userID uint, limit int) ([]dto.ProductResponse, error) {
	ctx := context.Background()
	key := s.cache.GetWarmStartKey(userID, limit)

	// Try cache first
	var cached []dto.ProductResponse
	if err := s.cache.GetJSON(ctx, key, &cached); err == nil {
		log.Printf("[CACHE HIT] warm-start recommendations (user=%d, limit=%d)", userID, limit)
		return cached, nil
	} else if err != redis.Nil {
		log.Printf("[CACHE ERROR] warm-start: %v", err)
	}

	// Cache miss - get from base service
	result, err := s.base.GetWarmStartRecommendations(userID, limit)
	if err != nil {
		return nil, err
	}

	// Cache the result (shorter TTL for personalized content)
	if err := s.cache.SetJSON(ctx, key, result, CacheTTLWarmStart); err != nil {
		log.Printf("[CACHE ERROR] failed to cache warm-start: %v", err)
	} else {
		log.Printf("[CACHE SET] warm-start recommendations (user=%d, limit=%d)", userID, limit)
	}

	return result, nil
}

func (s *cachedRecommendationService) GetSimilarProducts(productID uint, limit int) ([]dto.ProductResponse, error) {
	ctx := context.Background()
	key := s.cache.GetSimilarKey(productID, limit)

	// Try cache first
	var cached []dto.ProductResponse
	if err := s.cache.GetJSON(ctx, key, &cached); err == nil {
		log.Printf("[CACHE HIT] similar products (product=%d, limit=%d)", productID, limit)
		return cached, nil
	} else if err != redis.Nil {
		log.Printf("[CACHE ERROR] similar: %v", err)
	}

	// Cache miss - get from base service
	result, err := s.base.GetSimilarProducts(productID, limit)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := s.cache.SetJSON(ctx, key, result, CacheTTLSimilar); err != nil {
		log.Printf("[CACHE ERROR] failed to cache similar: %v", err)
	} else {
		log.Printf("[CACHE SET] similar products (product=%d, limit=%d)", productID, limit)
	}

	return result, nil
}

func (s *cachedRecommendationService) GetCoViewedProducts(productID uint, limit int) ([]dto.ProductResponse, error) {
	ctx := context.Background()
	key := s.cache.GetCoViewedKey(productID, limit)

	// Try cache first
	var cached []dto.ProductResponse
	if err := s.cache.GetJSON(ctx, key, &cached); err == nil {
		log.Printf("[CACHE HIT] co-viewed products (product=%d, limit=%d)", productID, limit)
		return cached, nil
	} else if err != redis.Nil {
		log.Printf("[CACHE ERROR] co-viewed: %v", err)
	}

	// Cache miss - get from base service
	result, err := s.base.GetCoViewedProducts(productID, limit)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := s.cache.SetJSON(ctx, key, result, CacheTTLSimilar); err != nil {
		log.Printf("[CACHE ERROR] failed to cache co-viewed: %v", err)
	} else {
		log.Printf("[CACHE SET] co-viewed products (product=%d, limit=%d)", productID, limit)
	}

	return result, nil
}

func (s *cachedRecommendationService) GetCrossSellProducts(productID uint, limit int) ([]dto.ProductResponse, error) {
	ctx := context.Background()
	key := s.cache.GetCrossSellKey(productID, limit)

	// Try cache first
	var cached []dto.ProductResponse
	if err := s.cache.GetJSON(ctx, key, &cached); err == nil {
		log.Printf("[CACHE HIT] cross-sell products (product=%d, limit=%d)", productID, limit)
		return cached, nil
	} else if err != redis.Nil {
		log.Printf("[CACHE ERROR] cross-sell: %v", err)
	}

	// Cache miss - get from base service
	result, err := s.base.GetCrossSellProducts(productID, limit)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := s.cache.SetJSON(ctx, key, result, CacheTTLSimilar); err != nil {
		log.Printf("[CACHE ERROR] failed to cache cross-sell: %v", err)
	} else {
		log.Printf("[CACHE SET] cross-sell products (product=%d, limit=%d)", productID, limit)
	}

	return result, nil
}

func (s *cachedRecommendationService) GetTrendingProducts(limit int) ([]dto.ProductResponse, error) {
	ctx := context.Background()
	key := s.cache.GetTrendingKey(limit)

	// Try cache first
	var cached []dto.ProductResponse
	if err := s.cache.GetJSON(ctx, key, &cached); err == nil {
		log.Printf("[CACHE HIT] trending products (limit=%d)", limit)
		return cached, nil
	} else if err != redis.Nil {
		log.Printf("[CACHE ERROR] trending: %v", err)
	}

	// Cache miss - get from base service
	result, err := s.base.GetTrendingProducts(limit)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := s.cache.SetJSON(ctx, key, result, CacheTTLTrending); err != nil {
		log.Printf("[CACHE ERROR] failed to cache trending: %v", err)
	} else {
		log.Printf("[CACHE SET] trending products (limit=%d)", limit)
	}

	return result, nil
}

func (s *cachedRecommendationService) GetPopularByCategory(categoryID uint, limit int) ([]dto.ProductResponse, error) {
	ctx := context.Background()
	key := s.cache.GetPopularCategoryKey(categoryID, limit)

	// Try cache first
	var cached []dto.ProductResponse
	if err := s.cache.GetJSON(ctx, key, &cached); err == nil {
		log.Printf("[CACHE HIT] popular by category (cat=%d, limit=%d)", categoryID, limit)
		return cached, nil
	} else if err != redis.Nil {
		log.Printf("[CACHE ERROR] popular-cat: %v", err)
	}

	// Cache miss - get from base service
	result, err := s.base.GetPopularByCategory(categoryID, limit)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := s.cache.SetJSON(ctx, key, result, CacheTTLCategory); err != nil {
		log.Printf("[CACHE ERROR] failed to cache popular-cat: %v", err)
	} else {
		log.Printf("[CACHE SET] popular by category (cat=%d, limit=%d)", categoryID, limit)
	}

	return result, nil
}

func (s *cachedRecommendationService) GetPersonalizedRecommendations(userID uint, limit int) ([]dto.ProductResponse, error) {
	// Personalized uses the same as warm-start
	return s.GetWarmStartRecommendations(userID, limit)
}
