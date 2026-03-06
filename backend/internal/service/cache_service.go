package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache TTL constants
const (
	CacheTTLColdStart  = 30 * time.Minute
	CacheTTLTrending   = 10 * time.Minute
	CacheTTLSimilar    = 1 * time.Hour
	CacheTTLWarmStart  = 5 * time.Minute
	CacheTTLProduct    = 1 * time.Hour
	CacheTTLCategory   = 30 * time.Minute
	CacheTTLUserProfile = 3 * time.Minute
)

// Cache key prefixes
const (
	KeyPrefixColdStart   = "rec:cold"
	KeyPrefixTrending    = "rec:trending"
	KeyPrefixSimilar     = "rec:similar"
	KeyPrefixWarmStart   = "rec:warm"
	KeyPrefixCoViewed    = "rec:coviewed"
	KeyPrefixCrossSell   = "rec:crosssell"
	KeyPrefixPopularCat  = "rec:popular_cat"
	KeyPrefixProduct     = "product"
	KeyPrefixUserProfile = "user:profile"
)

// CacheService interface for caching operations
type CacheService interface {
	// Basic operations
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// JSON operations
	GetJSON(ctx context.Context, key string, dest interface{}) error
	SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Recommendation cache helpers
	GetColdStartKey(limit int) string
	GetTrendingKey(limit int) string
	GetSimilarKey(productID uint, limit int) string
	GetWarmStartKey(userID uint, limit int) string
	GetCoViewedKey(productID uint, limit int) string
	GetCrossSellKey(productID uint, limit int) string
	GetPopularCategoryKey(categoryID uint, limit int) string

	// Invalidation
	InvalidatePattern(ctx context.Context, pattern string) error

	// Health check
	IsConnected() bool
}

// redisCacheService implements CacheService using Redis
type redisCacheService struct {
	client *redis.Client
}

// NewCacheService creates a new Redis-based cache service
func NewCacheService(client *redis.Client) CacheService {
	return &redisCacheService{client: client}
}

func (s *redisCacheService) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

func (s *redisCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

func (s *redisCacheService) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

func (s *redisCacheService) Exists(ctx context.Context, key string) (bool, error) {
	result, err := s.client.Exists(ctx, key).Result()
	return result > 0, err
}

func (s *redisCacheService) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (s *redisCacheService) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, key, data, ttl).Err()
}

func (s *redisCacheService) GetColdStartKey(limit int) string {
	return fmt.Sprintf("%s:%d", KeyPrefixColdStart, limit)
}

func (s *redisCacheService) GetTrendingKey(limit int) string {
	return fmt.Sprintf("%s:%d", KeyPrefixTrending, limit)
}

func (s *redisCacheService) GetSimilarKey(productID uint, limit int) string {
	return fmt.Sprintf("%s:%d:%d", KeyPrefixSimilar, productID, limit)
}

func (s *redisCacheService) GetWarmStartKey(userID uint, limit int) string {
	return fmt.Sprintf("%s:%d:%d", KeyPrefixWarmStart, userID, limit)
}

func (s *redisCacheService) GetCoViewedKey(productID uint, limit int) string {
	return fmt.Sprintf("%s:%d:%d", KeyPrefixCoViewed, productID, limit)
}

func (s *redisCacheService) GetCrossSellKey(productID uint, limit int) string {
	return fmt.Sprintf("%s:%d:%d", KeyPrefixCrossSell, productID, limit)
}

func (s *redisCacheService) GetPopularCategoryKey(categoryID uint, limit int) string {
	return fmt.Sprintf("%s:%d:%d", KeyPrefixPopularCat, categoryID, limit)
}

func (s *redisCacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	iter := s.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := s.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (s *redisCacheService) IsConnected() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return s.client.Ping(ctx).Err() == nil
}

// noOpCacheService is a no-op implementation when Redis is disabled
type noOpCacheService struct{}

// NewNoOpCacheService creates a cache service that does nothing (for when Redis is disabled)
func NewNoOpCacheService() CacheService {
	return &noOpCacheService{}
}

func (s *noOpCacheService) Get(ctx context.Context, key string) (string, error) {
	return "", redis.Nil
}

func (s *noOpCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

func (s *noOpCacheService) Delete(ctx context.Context, key string) error {
	return nil
}

func (s *noOpCacheService) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (s *noOpCacheService) GetJSON(ctx context.Context, key string, dest interface{}) error {
	return redis.Nil
}

func (s *noOpCacheService) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

func (s *noOpCacheService) GetColdStartKey(limit int) string {
	return ""
}

func (s *noOpCacheService) GetTrendingKey(limit int) string {
	return ""
}

func (s *noOpCacheService) GetSimilarKey(productID uint, limit int) string {
	return ""
}

func (s *noOpCacheService) GetWarmStartKey(userID uint, limit int) string {
	return ""
}

func (s *noOpCacheService) GetCoViewedKey(productID uint, limit int) string {
	return ""
}

func (s *noOpCacheService) GetCrossSellKey(productID uint, limit int) string {
	return ""
}

func (s *noOpCacheService) GetPopularCategoryKey(categoryID uint, limit int) string {
	return ""
}

func (s *noOpCacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	return nil
}

func (s *noOpCacheService) IsConnected() bool {
	return false
}
