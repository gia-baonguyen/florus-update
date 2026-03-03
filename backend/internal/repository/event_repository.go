package repository

import (
	"time"

	"github.com/florus/backend/internal/models"
	"gorm.io/gorm"
)

type EventRepository interface {
	Create(event *models.UserEvent) error
	CreateBatch(events []models.UserEvent) error
	GetByID(id uint) (*models.UserEvent, error)
	GetByUserID(userID uint, limit int) ([]models.UserEvent, error)
	GetByProductID(productID uint, limit int) ([]models.UserEvent, error)
	GetByEventType(eventType models.EventType, limit int) ([]models.UserEvent, error)
	GetUserProductViews(userID uint, limit int) ([]uint, error)
	GetProductViewCount(productID uint) (int64, error)
	GetProductPurchaseCount(productID uint) (int64, error)
	GetUserBehaviorProfile(userID uint) (*models.UserBehaviorProfile, error)
	GetProductEventStats(productID uint) (*models.ProductEventStats, error)
	GetPopularProducts(limit int, since time.Time) ([]uint, error)
	GetFrequentlyBoughtTogether(productID uint, limit int) ([]uint, error)
	GetCoViewedProducts(productID uint, limit int) ([]uint, error)
	GetEventStats(since time.Time) ([]models.UserEventStats, error)
	DeleteOldEvents(before time.Time) (int64, error)
}

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Create(event *models.UserEvent) error {
	return r.db.Create(event).Error
}

func (r *eventRepository) CreateBatch(events []models.UserEvent) error {
	return r.db.CreateInBatches(events, 100).Error
}

func (r *eventRepository) GetByID(id uint) (*models.UserEvent, error) {
	var event models.UserEvent
	err := r.db.First(&event, id).Error
	return &event, err
}

func (r *eventRepository) GetByUserID(userID uint, limit int) ([]models.UserEvent, error) {
	var events []models.UserEvent
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *eventRepository) GetByProductID(productID uint, limit int) ([]models.UserEvent, error) {
	var events []models.UserEvent
	err := r.db.Where("product_id = ?", productID).
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *eventRepository) GetByEventType(eventType models.EventType, limit int) ([]models.UserEvent, error) {
	var events []models.UserEvent
	err := r.db.Where("event_type = ?", eventType).
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

// GetUserProductViews returns product IDs that user has viewed
func (r *eventRepository) GetUserProductViews(userID uint, limit int) ([]uint, error) {
	var productIDs []uint
	err := r.db.Model(&models.UserEvent{}).
		Select("DISTINCT product_id").
		Where("user_id = ? AND event_type = ? AND product_id IS NOT NULL", userID, models.EventProductView).
		Order("created_at DESC").
		Limit(limit).
		Pluck("product_id", &productIDs).Error
	return productIDs, err
}

func (r *eventRepository) GetProductViewCount(productID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.UserEvent{}).
		Where("product_id = ? AND event_type = ?", productID, models.EventProductView).
		Count(&count).Error
	return count, err
}

func (r *eventRepository) GetProductPurchaseCount(productID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.UserEvent{}).
		Where("product_id = ? AND event_type = ?", productID, models.EventPurchase).
		Count(&count).Error
	return count, err
}

// GetUserBehaviorProfile returns aggregated user behavior for recommendations
func (r *eventRepository) GetUserBehaviorProfile(userID uint) (*models.UserBehaviorProfile, error) {
	profile := &models.UserBehaviorProfile{UserID: userID}

	// Total events
	r.db.Model(&models.UserEvent{}).Where("user_id = ?", userID).Count(&profile.TotalEvents)

	// Product views
	r.db.Model(&models.UserEvent{}).
		Where("user_id = ? AND event_type = ?", userID, models.EventProductView).
		Count(&profile.ProductViews)

	// Purchases
	r.db.Model(&models.UserEvent{}).
		Where("user_id = ? AND event_type = ?", userID, models.EventPurchase).
		Count(&profile.Purchases)

	// Average rating
	r.db.Model(&models.UserEvent{}).
		Select("AVG(rating)").
		Where("user_id = ? AND event_type = ? AND rating IS NOT NULL", userID, models.EventRating).
		Scan(&profile.AverageRating)

	// Preferred categories (top 5)
	r.db.Model(&models.UserEvent{}).
		Select("category_id").
		Where("user_id = ? AND category_id IS NOT NULL", userID).
		Group("category_id").
		Order("COUNT(*) DESC").
		Limit(5).
		Pluck("category_id", &profile.PreferredCategories)

	// Recent products (last 20)
	r.db.Model(&models.UserEvent{}).
		Select("DISTINCT product_id").
		Where("user_id = ? AND product_id IS NOT NULL", userID).
		Order("created_at DESC").
		Limit(20).
		Pluck("product_id", &profile.RecentProducts)

	// Last activity
	r.db.Model(&models.UserEvent{}).
		Select("MAX(created_at)").
		Where("user_id = ?", userID).
		Scan(&profile.LastActivity)

	return profile, nil
}

// GetProductEventStats returns event statistics for a product
func (r *eventRepository) GetProductEventStats(productID uint) (*models.ProductEventStats, error) {
	stats := &models.ProductEventStats{ProductID: productID}

	r.db.Model(&models.UserEvent{}).
		Where("product_id = ? AND event_type = ?", productID, models.EventProductView).
		Count(&stats.Views)

	r.db.Model(&models.UserEvent{}).
		Where("product_id = ? AND event_type = ?", productID, models.EventProductClick).
		Count(&stats.Clicks)

	r.db.Model(&models.UserEvent{}).
		Where("product_id = ? AND event_type = ?", productID, models.EventAddToCart).
		Count(&stats.AddToCarts)

	r.db.Model(&models.UserEvent{}).
		Where("product_id = ? AND event_type = ?", productID, models.EventPurchase).
		Count(&stats.Purchases)

	if stats.Views > 0 {
		stats.ConversionRate = float64(stats.Purchases) / float64(stats.Views) * 100
	}

	return stats, nil
}

// GetPopularProducts returns most viewed/purchased products
func (r *eventRepository) GetPopularProducts(limit int, since time.Time) ([]uint, error) {
	var productIDs []uint
	err := r.db.Model(&models.UserEvent{}).
		Select("product_id").
		Where("product_id IS NOT NULL AND created_at >= ? AND event_type IN (?, ?)",
			since, models.EventProductView, models.EventPurchase).
		Group("product_id").
		Order("COUNT(*) DESC").
		Limit(limit).
		Pluck("product_id", &productIDs).Error
	return productIDs, err
}

// GetFrequentlyBoughtTogether returns products often purchased with the given product
func (r *eventRepository) GetFrequentlyBoughtTogether(productID uint, limit int) ([]uint, error) {
	var productIDs []uint

	// Find orders containing this product, then find other products in those orders
	subQuery := r.db.Model(&models.UserEvent{}).
		Select("order_id").
		Where("product_id = ? AND event_type = ? AND order_id IS NOT NULL", productID, models.EventPurchase)

	err := r.db.Model(&models.UserEvent{}).
		Select("product_id").
		Where("order_id IN (?) AND product_id != ? AND event_type = ?", subQuery, productID, models.EventPurchase).
		Group("product_id").
		Order("COUNT(*) DESC").
		Limit(limit).
		Pluck("product_id", &productIDs).Error
	return productIDs, err
}

// GetCoViewedProducts returns products often viewed by users who viewed the given product
func (r *eventRepository) GetCoViewedProducts(productID uint, limit int) ([]uint, error) {
	var productIDs []uint

	// Find users who viewed this product
	subQuery := r.db.Model(&models.UserEvent{}).
		Select("user_id").
		Where("product_id = ? AND event_type = ? AND user_id IS NOT NULL", productID, models.EventProductView)

	// Find other products viewed by those users
	err := r.db.Model(&models.UserEvent{}).
		Select("product_id").
		Where("user_id IN (?) AND product_id != ? AND event_type = ? AND product_id IS NOT NULL",
			subQuery, productID, models.EventProductView).
		Group("product_id").
		Order("COUNT(*) DESC").
		Limit(limit).
		Pluck("product_id", &productIDs).Error
	return productIDs, err
}

// GetEventStats returns aggregated event statistics
func (r *eventRepository) GetEventStats(since time.Time) ([]models.UserEventStats, error) {
	var stats []models.UserEventStats
	err := r.db.Model(&models.UserEvent{}).
		Select("event_type, COUNT(*) as count, COUNT(DISTINCT user_id) as unique_users, MAX(created_at) as last_occurred").
		Where("created_at >= ?", since).
		Group("event_type").
		Find(&stats).Error
	return stats, err
}

// DeleteOldEvents removes events older than the specified time
func (r *eventRepository) DeleteOldEvents(before time.Time) (int64, error) {
	result := r.db.Where("created_at < ?", before).Delete(&models.UserEvent{})
	return result.RowsAffected, result.Error
}
