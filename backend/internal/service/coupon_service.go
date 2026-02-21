package service

import (
	"errors"
	"strings"

	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
)

var (
	ErrCouponNotFound     = errors.New("coupon not found")
	ErrCouponInvalid      = errors.New("coupon is invalid or expired")
	ErrCouponMinAmount    = errors.New("order total does not meet minimum amount")
	ErrCouponUsageLimit   = errors.New("coupon usage limit reached")
	ErrCouponCodeExists   = errors.New("coupon code already exists")
)

type CouponService interface {
	Create(coupon *models.Coupon) error
	GetByID(id uint) (*models.Coupon, error)
	GetAll(page, limit int) ([]models.Coupon, int64, error)
	Update(coupon *models.Coupon) error
	Delete(id uint) error
	ValidateCoupon(code string, orderTotal float64) (*models.Coupon, float64, error)
	ApplyCoupon(id uint) error
}

type couponService struct {
	couponRepo repository.CouponRepository
}

func NewCouponService(couponRepo repository.CouponRepository) CouponService {
	return &couponService{
		couponRepo: couponRepo,
	}
}

func (s *couponService) Create(coupon *models.Coupon) error {
	// Normalize code to uppercase
	coupon.Code = strings.ToUpper(strings.TrimSpace(coupon.Code))

	// Check if code already exists
	existing, _ := s.couponRepo.GetByCode(coupon.Code)
	if existing != nil {
		return ErrCouponCodeExists
	}

	return s.couponRepo.Create(coupon)
}

func (s *couponService) GetByID(id uint) (*models.Coupon, error) {
	return s.couponRepo.GetByID(id)
}

func (s *couponService) GetAll(page, limit int) ([]models.Coupon, int64, error) {
	return s.couponRepo.GetAll(page, limit)
}

func (s *couponService) Update(coupon *models.Coupon) error {
	existing, err := s.couponRepo.GetByID(coupon.ID)
	if err != nil {
		return ErrCouponNotFound
	}

	// If code is changed, check for duplicates
	newCode := strings.ToUpper(strings.TrimSpace(coupon.Code))
	if newCode != existing.Code {
		existingWithCode, _ := s.couponRepo.GetByCode(newCode)
		if existingWithCode != nil {
			return ErrCouponCodeExists
		}
	}
	coupon.Code = newCode

	return s.couponRepo.Update(coupon)
}

func (s *couponService) Delete(id uint) error {
	_, err := s.couponRepo.GetByID(id)
	if err != nil {
		return ErrCouponNotFound
	}
	return s.couponRepo.Delete(id)
}

func (s *couponService) ValidateCoupon(code string, orderTotal float64) (*models.Coupon, float64, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	coupon, err := s.couponRepo.GetByCode(code)
	if err != nil {
		return nil, 0, ErrCouponNotFound
	}

	if !coupon.IsValid() {
		return nil, 0, ErrCouponInvalid
	}

	if orderTotal < coupon.MinOrderAmount {
		return nil, 0, ErrCouponMinAmount
	}

	discount := coupon.CalculateDiscount(orderTotal)

	return coupon, discount, nil
}

func (s *couponService) ApplyCoupon(id uint) error {
	return s.couponRepo.IncrementUsedCount(id)
}
