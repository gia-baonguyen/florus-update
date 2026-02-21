package repository

import (
	"github.com/florus/backend/internal/models"
	"gorm.io/gorm"
)

type WishlistRepository interface {
	Create(item *models.WishlistItem) error
	FindByUserID(userID uint) ([]models.WishlistItem, error)
	FindByUserAndProduct(userID, productID uint) (*models.WishlistItem, error)
	Delete(id uint) error
	DeleteByUserAndProduct(userID, productID uint) error
	Exists(userID, productID uint) (bool, error)
	GetProductIDs(userID uint) ([]uint, error)
}

type wishlistRepository struct {
	db *gorm.DB
}

func NewWishlistRepository(db *gorm.DB) WishlistRepository {
	return &wishlistRepository{db: db}
}

func (r *wishlistRepository) Create(item *models.WishlistItem) error {
	return r.db.Create(item).Error
}

func (r *wishlistRepository) FindByUserID(userID uint) ([]models.WishlistItem, error) {
	var items []models.WishlistItem
	err := r.db.Preload("Product").
		Preload("Product.Category").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}

func (r *wishlistRepository) FindByUserAndProduct(userID, productID uint) (*models.WishlistItem, error) {
	var item models.WishlistItem
	err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *wishlistRepository) Delete(id uint) error {
	return r.db.Delete(&models.WishlistItem{}, id).Error
}

func (r *wishlistRepository) DeleteByUserAndProduct(userID, productID uint) error {
	return r.db.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&models.WishlistItem{}).Error
}

func (r *wishlistRepository) Exists(userID, productID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.WishlistItem{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Count(&count).Error
	return count > 0, err
}

func (r *wishlistRepository) GetProductIDs(userID uint) ([]uint, error) {
	var productIDs []uint
	err := r.db.Model(&models.WishlistItem{}).
		Where("user_id = ?", userID).
		Pluck("product_id", &productIDs).Error
	return productIDs, err
}
