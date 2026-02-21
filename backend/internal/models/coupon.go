package models

import "time"

type DiscountType string

const (
	DiscountTypePercent DiscountType = "percent"
	DiscountTypeFixed   DiscountType = "fixed"
)

type Coupon struct {
	ID               uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	Code             string       `gorm:"uniqueIndex;size:50;not null" json:"code"`
	DiscountType     DiscountType `gorm:"type:varchar(20);not null" json:"discount_type"`
	DiscountValue    float64      `gorm:"not null" json:"discount_value"`
	MinOrderAmount   float64      `gorm:"default:0" json:"min_order_amount"`
	MaxDiscountAmount float64     `gorm:"default:0" json:"max_discount_amount"` // 0 = no limit
	UsageLimit       int          `gorm:"default:0" json:"usage_limit"`         // 0 = unlimited
	UsedCount        int          `gorm:"default:0" json:"used_count"`
	StartDate        time.Time    `gorm:"not null" json:"start_date"`
	EndDate          time.Time    `gorm:"not null" json:"end_date"`
	IsActive         bool         `gorm:"default:true" json:"is_active"`
	Description      string       `gorm:"type:text" json:"description"`
	CreatedAt        time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Coupon) TableName() string {
	return "coupons"
}

// IsValid checks if the coupon is valid for use
func (c *Coupon) IsValid() bool {
	now := time.Now()
	if !c.IsActive {
		return false
	}
	if now.Before(c.StartDate) || now.After(c.EndDate) {
		return false
	}
	if c.UsageLimit > 0 && c.UsedCount >= c.UsageLimit {
		return false
	}
	return true
}

// CalculateDiscount calculates the discount amount for a given order total
func (c *Coupon) CalculateDiscount(orderTotal float64) float64 {
	if orderTotal < c.MinOrderAmount {
		return 0
	}

	var discount float64
	if c.DiscountType == DiscountTypePercent {
		discount = orderTotal * (c.DiscountValue / 100)
	} else {
		discount = c.DiscountValue
	}

	// Apply max discount limit if set
	if c.MaxDiscountAmount > 0 && discount > c.MaxDiscountAmount {
		discount = c.MaxDiscountAmount
	}

	// Discount cannot exceed order total
	if discount > orderTotal {
		discount = orderTotal
	}

	return discount
}
