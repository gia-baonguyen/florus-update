package repository

import (
	"github.com/florus/backend/internal/models"
	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(review *models.Review) error
	FindByID(id uint) (*models.Review, error)
	FindByProductID(productID uint) ([]models.Review, error)
	FindByUserAndProduct(userID, productID uint) (*models.Review, error)
	Update(review *models.Review) error
	Delete(id uint) error
	GetAverageRating(productID uint) (float64, int, error)
}

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(review *models.Review) error {
	return r.db.Create(review).Error
}

func (r *reviewRepository) FindByID(id uint) (*models.Review, error) {
	var review models.Review
	err := r.db.Preload("User").First(&review, id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) FindByProductID(productID uint) ([]models.Review, error) {
	var reviews []models.Review
	err := r.db.Preload("User").
		Where("product_id = ?", productID).
		Order("created_at DESC").
		Find(&reviews).Error
	return reviews, err
}

func (r *reviewRepository) FindByUserAndProduct(userID, productID uint) (*models.Review, error) {
	var review models.Review
	err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) Update(review *models.Review) error {
	return r.db.Save(review).Error
}

func (r *reviewRepository) Delete(id uint) error {
	return r.db.Delete(&models.Review{}, id).Error
}

func (r *reviewRepository) GetAverageRating(productID uint) (float64, int, error) {
	var result struct {
		Average float64
		Count   int
	}
	err := r.db.Model(&models.Review{}).
		Select("COALESCE(AVG(rating), 0) as average, COUNT(*) as count").
		Where("product_id = ?", productID).
		Scan(&result).Error
	return result.Average, result.Count, err
}
