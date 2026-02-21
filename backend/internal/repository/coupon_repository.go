package repository

import (
	"github.com/florus/backend/internal/models"
	"gorm.io/gorm"
)

type CouponRepository interface {
	Create(coupon *models.Coupon) error
	GetByID(id uint) (*models.Coupon, error)
	GetByCode(code string) (*models.Coupon, error)
	GetAll(page, limit int) ([]models.Coupon, int64, error)
	Update(coupon *models.Coupon) error
	Delete(id uint) error
	IncrementUsedCount(id uint) error
}

type couponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) CouponRepository {
	return &couponRepository{db: db}
}

func (r *couponRepository) Create(coupon *models.Coupon) error {
	return r.db.Create(coupon).Error
}

func (r *couponRepository) GetByID(id uint) (*models.Coupon, error) {
	var coupon models.Coupon
	err := r.db.First(&coupon, id).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) GetByCode(code string) (*models.Coupon, error) {
	var coupon models.Coupon
	err := r.db.Where("code = ?", code).First(&coupon).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) GetAll(page, limit int) ([]models.Coupon, int64, error) {
	var coupons []models.Coupon
	var total int64

	offset := (page - 1) * limit

	err := r.db.Model(&models.Coupon{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&coupons).Error
	if err != nil {
		return nil, 0, err
	}

	return coupons, total, nil
}

func (r *couponRepository) Update(coupon *models.Coupon) error {
	return r.db.Save(coupon).Error
}

func (r *couponRepository) Delete(id uint) error {
	return r.db.Delete(&models.Coupon{}, id).Error
}

func (r *couponRepository) IncrementUsedCount(id uint) error {
	return r.db.Model(&models.Coupon{}).Where("id = ?", id).
		UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error
}
