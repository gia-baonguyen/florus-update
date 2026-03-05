package repository

import (
	"time"

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
	// Image methods
	AddImages(reviewID uint, urls []string) error
	DeleteImages(reviewID uint) error
	ReplaceImages(reviewID uint, urls []string) error
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
	err := r.db.Preload("User").Preload("Images").First(&review, id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) FindByProductID(productID uint) ([]models.Review, error) {
	var reviews []models.Review
	err := r.db.Preload("User").Preload("Images").
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
	return r.db.Model(review).Omit("User", "Product").Updates(models.Review{
		Rating:    review.Rating,
		Comment:   review.Comment,
		UpdatedAt: time.Now(),
	}).Error
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

func (r *reviewRepository) AddImages(reviewID uint, urls []string) error {
	if len(urls) == 0 {
		return nil
	}
	images := make([]models.ReviewImage, len(urls))
	for i, url := range urls {
		images[i] = models.ReviewImage{
			ReviewID:  reviewID,
			URL:       url,
			SortOrder: i,
		}
	}
	return r.db.Create(&images).Error
}

func (r *reviewRepository) DeleteImages(reviewID uint) error {
	return r.db.Where("review_id = ?", reviewID).Delete(&models.ReviewImage{}).Error
}

func (r *reviewRepository) ReplaceImages(reviewID uint, urls []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing images first using Unscoped for hard delete
		if err := tx.Unscoped().Where("review_id = ?", reviewID).Delete(&models.ReviewImage{}).Error; err != nil {
			return err
		}

		// Add new images if any
		if len(urls) > 0 {
			images := make([]models.ReviewImage, len(urls))
			for i, url := range urls {
				images[i] = models.ReviewImage{
					ReviewID:  reviewID,
					URL:       url,
					SortOrder: i,
				}
			}
			if err := tx.Create(&images).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
