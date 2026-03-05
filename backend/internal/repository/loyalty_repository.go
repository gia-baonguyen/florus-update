package repository

import (
	"github.com/florus/backend/internal/models"
	"gorm.io/gorm"
)

type LoyaltyRepository interface {
	CreateTransaction(tx *gorm.DB, t *models.LoyaltyTransaction) error
	GetUserPoints(tx *gorm.DB, userID uint) (int64, error)
	UpdateUserPointsAndTier(tx *gorm.DB, userID uint, pointsDelta int64, newTier string) error
	GetDB() *gorm.DB
}

type loyaltyRepository struct {
	db *gorm.DB
}

func NewLoyaltyRepository(db *gorm.DB) LoyaltyRepository {
	return &loyaltyRepository{db: db}
}

func (r *loyaltyRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *loyaltyRepository) CreateTransaction(tx *gorm.DB, t *models.LoyaltyTransaction) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(t).Error
}

func (r *loyaltyRepository) GetUserPoints(tx *gorm.DB, userID uint) (int64, error) {
	if tx == nil {
		tx = r.db
	}
	var points int64
	err := tx.Model(&models.User{}).
		Select("loyalty_points").
		Where("id = ?", userID).
		Scan(&points).Error
	return points, err
}

func (r *loyaltyRepository) UpdateUserPointsAndTier(tx *gorm.DB, userID uint, pointsDelta int64, newTier string) error {
	if tx == nil {
		tx = r.db
	}

	updates := map[string]interface{}{
		"loyalty_points": gorm.Expr("loyalty_points + ?", pointsDelta),
	}
	if newTier != "" {
		updates["loyalty_tier"] = newTier
	}

	return tx.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates).Error
}

