package service

import (
	"log"
	"time"

	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
	"gorm.io/gorm"
)

// AnalyticsService handles product relationship computation and analytics
type AnalyticsService interface {
	// Compute and update product relationships
	ComputeProductRelationships() error
	ComputeCoViewedRelationships(since time.Time) error
	ComputeFrequentlyBoughtTogether(since time.Time) error
	ComputeSimilarProducts() error

	// Analytics queries
	GetProductAnalytics(productID uint) (*ProductAnalytics, error)
	GetTopPerformingProducts(limit int) ([]ProductAnalytics, error)
	GetConversionFunnel(since time.Time) (*ConversionFunnel, error)
}

type ProductAnalytics struct {
	ProductID      uint    `json:"product_id"`
	ProductName    string  `json:"product_name"`
	Views          int64   `json:"views"`
	AddToCarts     int64   `json:"add_to_carts"`
	Purchases      int64   `json:"purchases"`
	Revenue        float64 `json:"revenue"`
	ConversionRate float64 `json:"conversion_rate"`
	AvgRating      float64 `json:"avg_rating"`
}

type ConversionFunnel struct {
	Views           int64   `json:"views"`
	AddToCarts      int64   `json:"add_to_carts"`
	Checkouts       int64   `json:"checkouts"`
	Purchases       int64   `json:"purchases"`
	ViewToCart      float64 `json:"view_to_cart_rate"`
	CartToPurchase  float64 `json:"cart_to_purchase_rate"`
	OverallConv     float64 `json:"overall_conversion_rate"`
}

type analyticsService struct {
	db          *gorm.DB
	eventRepo   repository.EventRepository
	productRepo repository.ProductRepository
}

func NewAnalyticsService(db *gorm.DB, eventRepo repository.EventRepository, productRepo repository.ProductRepository) AnalyticsService {
	return &analyticsService{
		db:          db,
		eventRepo:   eventRepo,
		productRepo: productRepo,
	}
}

// ComputeProductRelationships runs all relationship computations
func (s *analyticsService) ComputeProductRelationships() error {
	since := time.Now().AddDate(0, -3, 0) // Last 3 months

	log.Println("[Analytics] Starting product relationship computation...")

	// Compute co-viewed relationships
	if err := s.ComputeCoViewedRelationships(since); err != nil {
		log.Printf("[Analytics] Error computing co-viewed: %v", err)
	}

	// Compute frequently bought together
	if err := s.ComputeFrequentlyBoughtTogether(since); err != nil {
		log.Printf("[Analytics] Error computing frequently bought together: %v", err)
	}

	// Compute similar products (content-based)
	if err := s.ComputeSimilarProducts(); err != nil {
		log.Printf("[Analytics] Error computing similar products: %v", err)
	}

	log.Println("[Analytics] Product relationship computation complete")
	return nil
}

// ComputeCoViewedRelationships computes "users who viewed X also viewed Y"
func (s *analyticsService) ComputeCoViewedRelationships(since time.Time) error {
	log.Println("[Analytics] Computing co-viewed relationships...")

	// Get all products with views
	var productIDs []uint
	err := s.db.Model(&models.UserEvent{}).
		Select("DISTINCT product_id").
		Where("event_type = ? AND product_id IS NOT NULL AND created_at >= ?",
			models.EventProductView, since).
		Pluck("product_id", &productIDs).Error
	if err != nil {
		return err
	}

	log.Printf("[Analytics] Processing %d products for co-viewed relationships", len(productIDs))

	for _, productID := range productIDs {
		coViewedIDs, err := s.eventRepo.GetCoViewedProducts(productID, 10)
		if err != nil || len(coViewedIDs) == 0 {
			continue
		}

		// Save relationships
		for i, targetID := range coViewedIDs {
			score := float64(10-i) / 10.0 // Score decreases with rank

			relationship := models.ProductRelationship{
				SourceProductID:  productID,
				TargetProductID:  targetID,
				RelationshipType: models.RelationCoViewed,
				Score:            &score,
			}

			// Upsert relationship
			s.db.Where("source_product_id = ? AND target_product_id = ? AND relationship_type = ?",
				productID, targetID, models.RelationCoViewed).
				Assign(relationship).
				FirstOrCreate(&relationship)
		}
	}

	log.Println("[Analytics] Co-viewed relationships computed")
	return nil
}

// ComputeFrequentlyBoughtTogether computes "frequently bought together"
func (s *analyticsService) ComputeFrequentlyBoughtTogether(since time.Time) error {
	log.Println("[Analytics] Computing frequently bought together...")

	// Get all products with purchases
	var productIDs []uint
	err := s.db.Model(&models.UserEvent{}).
		Select("DISTINCT product_id").
		Where("event_type = ? AND product_id IS NOT NULL AND created_at >= ?",
			models.EventPurchase, since).
		Pluck("product_id", &productIDs).Error
	if err != nil {
		return err
	}

	log.Printf("[Analytics] Processing %d products for frequently bought together", len(productIDs))

	for _, productID := range productIDs {
		fbtIDs, err := s.eventRepo.GetFrequentlyBoughtTogether(productID, 5)
		if err != nil || len(fbtIDs) == 0 {
			continue
		}

		// Save relationships
		for i, targetID := range fbtIDs {
			score := float64(5-i) / 5.0

			relationship := models.ProductRelationship{
				SourceProductID:  productID,
				TargetProductID:  targetID,
				RelationshipType: models.RelationFrequentlyBought,
				Score:            &score,
			}

			s.db.Where("source_product_id = ? AND target_product_id = ? AND relationship_type = ?",
				productID, targetID, models.RelationFrequentlyBought).
				Assign(relationship).
				FirstOrCreate(&relationship)
		}
	}

	log.Println("[Analytics] Frequently bought together computed")
	return nil
}

// ComputeSimilarProducts computes content-based similar products
func (s *analyticsService) ComputeSimilarProducts() error {
	log.Println("[Analytics] Computing similar products...")

	// Get all active products
	var products []models.Product
	err := s.db.Where("is_active = ?", true).Find(&products).Error
	if err != nil {
		return err
	}

	log.Printf("[Analytics] Processing %d products for similarity", len(products))

	for _, product := range products {
		// Find similar products based on content
		similarProducts, err := s.productRepo.FindSimilar(product.ID, product.CategoryID, product.Brand, product.Price, 6)
		if err != nil || len(similarProducts) == 0 {
			continue
		}

		// Save relationships
		for i, similar := range similarProducts {
			score := float64(6-i) / 6.0

			relationship := models.ProductRelationship{
				SourceProductID:  product.ID,
				TargetProductID:  similar.ID,
				RelationshipType: models.RelationSimilar,
				Score:            &score,
			}

			s.db.Where("source_product_id = ? AND target_product_id = ? AND relationship_type = ?",
				product.ID, similar.ID, models.RelationSimilar).
				Assign(relationship).
				FirstOrCreate(&relationship)
		}
	}

	log.Println("[Analytics] Similar products computed")
	return nil
}

// GetProductAnalytics returns analytics for a specific product
func (s *analyticsService) GetProductAnalytics(productID uint) (*ProductAnalytics, error) {
	stats, err := s.eventRepo.GetProductEventStats(productID)
	if err != nil {
		return nil, err
	}

	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, err
	}

	// Calculate revenue from purchase events
	var revenue float64
	s.db.Model(&models.UserEvent{}).
		Select("COALESCE(SUM(price * quantity), 0)").
		Where("product_id = ? AND event_type = ?", productID, models.EventPurchase).
		Scan(&revenue)

	return &ProductAnalytics{
		ProductID:      productID,
		ProductName:    product.Name,
		Views:          stats.Views,
		AddToCarts:     stats.AddToCarts,
		Purchases:      stats.Purchases,
		Revenue:        revenue,
		ConversionRate: stats.ConversionRate,
		AvgRating:      product.Rating,
	}, nil
}

// GetTopPerformingProducts returns top performing products by revenue or conversions
func (s *analyticsService) GetTopPerformingProducts(limit int) ([]ProductAnalytics, error) {
	var results []ProductAnalytics

	// Get top products by purchase count
	type ProductStat struct {
		ProductID uint
		Purchases int64
		Revenue   float64
	}

	var stats []ProductStat
	err := s.db.Model(&models.UserEvent{}).
		Select("product_id, COUNT(*) as purchases, COALESCE(SUM(price * quantity), 0) as revenue").
		Where("event_type = ? AND product_id IS NOT NULL", models.EventPurchase).
		Group("product_id").
		Order("purchases DESC").
		Limit(limit).
		Find(&stats).Error
	if err != nil {
		return nil, err
	}

	for _, stat := range stats {
		analytics, err := s.GetProductAnalytics(stat.ProductID)
		if err != nil {
			continue
		}
		results = append(results, *analytics)
	}

	return results, nil
}

// GetConversionFunnel returns conversion funnel metrics
func (s *analyticsService) GetConversionFunnel(since time.Time) (*ConversionFunnel, error) {
	funnel := &ConversionFunnel{}

	// Count views
	s.db.Model(&models.UserEvent{}).
		Where("event_type = ? AND created_at >= ?", models.EventProductView, since).
		Count(&funnel.Views)

	// Count add to carts
	s.db.Model(&models.UserEvent{}).
		Where("event_type = ? AND created_at >= ?", models.EventAddToCart, since).
		Count(&funnel.AddToCarts)

	// Count purchases
	s.db.Model(&models.UserEvent{}).
		Where("event_type = ? AND created_at >= ?", models.EventPurchase, since).
		Count(&funnel.Purchases)

	// Calculate rates
	if funnel.Views > 0 {
		funnel.ViewToCart = float64(funnel.AddToCarts) / float64(funnel.Views) * 100
		funnel.OverallConv = float64(funnel.Purchases) / float64(funnel.Views) * 100
	}
	if funnel.AddToCarts > 0 {
		funnel.CartToPurchase = float64(funnel.Purchases) / float64(funnel.AddToCarts) * 100
	}

	return funnel, nil
}
