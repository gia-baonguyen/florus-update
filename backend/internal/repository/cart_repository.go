package repository

import (
	"github.com/florus/backend/internal/models"
	"gorm.io/gorm"
)

type CartRepository interface {
	AddItem(item *models.CartItem) error
	FindByUserID(userID uint) ([]models.CartItem, error)
	FindItem(userID, productID uint) (*models.CartItem, error)
	FindItemByID(id uint) (*models.CartItem, error)
	UpdateQuantity(id uint, quantity int) error
	DeleteItem(id uint) error
	ClearCart(userID uint) error
	GetCartTotal(userID uint) (float64, error)
}

type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) AddItem(item *models.CartItem) error {
	return r.db.Create(item).Error
}

func (r *cartRepository) FindByUserID(userID uint) ([]models.CartItem, error) {
	var items []models.CartItem
	err := r.db.Preload("Product").Preload("Product.Category").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}

func (r *cartRepository) FindItem(userID, productID uint) (*models.CartItem, error) {
	var item models.CartItem
	err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *cartRepository) FindItemByID(id uint) (*models.CartItem, error) {
	var item models.CartItem
	err := r.db.Preload("Product").First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *cartRepository) UpdateQuantity(id uint, quantity int) error {
	return r.db.Model(&models.CartItem{}).Where("id = ?", id).Update("quantity", quantity).Error
}

func (r *cartRepository) DeleteItem(id uint) error {
	return r.db.Delete(&models.CartItem{}, id).Error
}

func (r *cartRepository) ClearCart(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.CartItem{}).Error
}

func (r *cartRepository) GetCartTotal(userID uint) (float64, error) {
	var total float64
	err := r.db.Model(&models.CartItem{}).
		Select("COALESCE(SUM(cart_items.quantity * products.price), 0)").
		Joins("JOIN products ON products.id = cart_items.product_id").
		Where("cart_items.user_id = ?", userID).
		Scan(&total).Error
	return total, err
}
