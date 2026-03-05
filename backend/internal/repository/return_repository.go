package repository

import (
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/pkg/utils"
	"gorm.io/gorm"
)

type ReturnRepository interface {
	Create(tx *gorm.DB, r *models.Return, items []models.ReturnItem) error
	FindByIDForUser(id, userID uint) (*models.Return, error)
	FindByUserID(userID uint, pagination utils.Pagination) ([]models.Return, int64, error)
	UpdateStatus(tx *gorm.DB, id uint, status models.ReturnStatus) error
	GetDB() *gorm.DB
}

type returnRepository struct {
	db *gorm.DB
}

func NewReturnRepository(db *gorm.DB) ReturnRepository {
	return &returnRepository{db: db}
}

func (r *returnRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *returnRepository) Create(tx *gorm.DB, ret *models.Return, items []models.ReturnItem) error {
	if tx == nil {
		tx = r.db
	}

	return tx.Transaction(func(txx *gorm.DB) error {
		if err := txx.Create(ret).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].ReturnID = ret.ID
		}
		if len(items) > 0 {
			if err := txx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *returnRepository) FindByIDForUser(id, userID uint) (*models.Return, error) {
	var ret models.Return
	err := r.db.Preload("Items").Preload("Items.OrderItem").Preload("Items.OrderItem.Product").
		Where("id = ? AND user_id = ?", id, userID).
		First(&ret).Error
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (r *returnRepository) FindByUserID(userID uint, pagination utils.Pagination) ([]models.Return, int64, error) {
	var returns []models.Return
	var total int64

	query := r.db.Model(&models.Return{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("Items").
		Offset(pagination.GetOffset()).
		Limit(pagination.Limit).
		Order("created_at DESC").
		Find(&returns).Error; err != nil {
		return nil, 0, err
	}

	return returns, total, nil
}

func (r *returnRepository) UpdateStatus(tx *gorm.DB, id uint, status models.ReturnStatus) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Model(&models.Return{}).
		Where("id = ?", id).
		Update("status", status).Error
}

