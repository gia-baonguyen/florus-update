package service

import (
	"math"

	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
	"gorm.io/gorm"
)

type LoyaltyService interface {
	// Earn points based on order subtotal (after discounts, before shipping)
	EarnPoints(order *models.Order, userID uint) error
	// Redeem points for an order; returns discount amount actually applied
	RedeemPoints(userID uint, requestedPoints int64, subtotal float64) (appliedPoints int64, discountAmount float64, err error)
	// Recalculate and update tier for a user based on spend (simplified)
	RecalculateTier(userID uint, totalSpent float64) (string, error)
	// Get basic loyalty summary for a user
	GetUserLoyalty(userID uint) (tier string, points int64, err error)
}

type loyaltyService struct {
	repo repository.LoyaltyRepository
}

func NewLoyaltyService(repo repository.LoyaltyRepository) LoyaltyService {
	return &loyaltyService{repo: repo}
}

// earn 1 point per 10,000 VND subtotal (rounded down)
const (
	loyaltyEarnRate      = 10000.0 // 1 point per 10k
	loyaltyRedeemValue   = 1000.0  // 1 point = 1k discount
	tierSilverThreshold  = 0.0
	tierGoldThreshold    = 5000000.0
	tierPlatinumThreshold = 15000000.0
)

func (s *loyaltyService) EarnPoints(order *models.Order, userID uint) error {
	points := int64(math.Floor(order.Subtotal / loyaltyEarnRate))
	if points <= 0 {
		return nil
	}

	db := s.repo.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		// transaction record
		trx := &models.LoyaltyTransaction{
			UserID:  userID,
			OrderID: &order.ID,
			Type:    models.LoyaltyTransactionTypeEarn,
			Points:  points,
		}
		if err := s.repo.CreateTransaction(tx, trx); err != nil {
			return err
		}
		// update balance; tier recomputation can be done separately
		return s.repo.UpdateUserPointsAndTier(tx, userID, points, "")
	})
}

func (s *loyaltyService) RedeemPoints(userID uint, requestedPoints int64, subtotal float64) (int64, float64, error) {
	if requestedPoints <= 0 {
		return 0, 0, nil
	}
	db := s.repo.GetDB()

	var appliedPoints int64
	var discountAmount float64

	err := db.Transaction(func(tx *gorm.DB) error {
		currentPoints, err := s.repo.GetUserPoints(tx, userID)
		if err != nil {
			return err
		}
		if currentPoints <= 0 {
			return nil
		}

		if requestedPoints > currentPoints {
			requestedPoints = currentPoints
		}

		// cannot discount more than subtotal
		maxPointsBySubtotal := int64(math.Floor(subtotal / loyaltyRedeemValue))
		if requestedPoints > maxPointsBySubtotal {
			requestedPoints = maxPointsBySubtotal
		}

		if requestedPoints <= 0 {
			return nil
		}

		appliedPoints = requestedPoints
		discountAmount = float64(appliedPoints) * loyaltyRedeemValue

		// create transaction
		trx := &models.LoyaltyTransaction{
			UserID: userID,
			Type:   models.LoyaltyTransactionTypeRedeem,
			Points: -appliedPoints,
		}
		if err := s.repo.CreateTransaction(tx, trx); err != nil {
			return err
		}

		// update balance
		return s.repo.UpdateUserPointsAndTier(tx, userID, -appliedPoints, "")
	})

	return appliedPoints, discountAmount, err
}

func (s *loyaltyService) RecalculateTier(userID uint, totalSpent float64) (string, error) {
	var tier string
	switch {
	case totalSpent >= tierPlatinumThreshold:
		tier = "Platinum"
	case totalSpent >= tierGoldThreshold:
		tier = "Gold"
	default:
		tier = "Silver"
	}
	err := s.repo.UpdateUserPointsAndTier(nil, userID, 0, tier)
	return tier, err
}

func (s *loyaltyService) GetUserLoyalty(userID uint) (string, int64, error) {
	db := s.repo.GetDB()
	var user models.User
	if err := db.Select("loyalty_tier, loyalty_points").First(&user, userID).Error; err != nil {
		return "", 0, err
	}
	return user.LoyaltyTier, user.LoyaltyPoints, nil
}

