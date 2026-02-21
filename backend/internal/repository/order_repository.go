package repository

import (
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/pkg/utils"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(tx *gorm.DB, order *models.Order) error
	CreateItems(tx *gorm.DB, items []models.OrderItem) error
	FindByID(id uint) (*models.Order, error)
	FindByUserID(userID uint, pagination utils.Pagination) ([]models.Order, int64, error)
	FindByOrderCode(code string) (*models.Order, error)
	UpdateStatus(id uint, status models.OrderStatus) error
	GetDB() *gorm.DB
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(tx *gorm.DB, order *models.Order) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(order).Error
}

func (r *orderRepository) CreateItems(tx *gorm.DB, items []models.OrderItem) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(&items).Error
}

func (r *orderRepository) FindByID(id uint) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("OrderItems").Preload("OrderItems.Product").
		Preload("OrderItems.Product.Category").
		First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByUserID(userID uint, pagination utils.Pagination) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{}).Where("user_id = ?", userID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("OrderItems").Preload("OrderItems.Product").
		Offset(pagination.GetOffset()).
		Limit(pagination.Limit).
		Order("created_at DESC").
		Find(&orders).Error

	return orders, total, err
}

func (r *orderRepository) FindByOrderCode(code string) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("OrderItems").Preload("OrderItems.Product").
		Where("order_code = ?", code).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) UpdateStatus(id uint, status models.OrderStatus) error {
	return r.db.Model(&models.Order{}).Where("id = ?", id).Update("status", status).Error
}

func (r *orderRepository) GetDB() *gorm.DB {
	return r.db
}
