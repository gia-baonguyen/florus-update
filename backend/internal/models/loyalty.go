package models

import "time"

// LoyaltyTransactionType represents the type of loyalty transaction.
type LoyaltyTransactionType string

const (
	LoyaltyTransactionTypeEarn   LoyaltyTransactionType = "earn"
	LoyaltyTransactionTypeRedeem LoyaltyTransactionType = "redeem"
	LoyaltyTransactionTypeAdjust LoyaltyTransactionType = "adjust"
)

type LoyaltyTransaction struct {
	ID          uint                    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint                    `gorm:"index;not null" json:"user_id"`
	OrderID     *uint                   `gorm:"index" json:"order_id,omitempty"`
	Type        LoyaltyTransactionType  `gorm:"type:varchar(20);not null" json:"type"`
	Points      int64                   `gorm:"type:number(10);not null" json:"points"`
	Description string                  `gorm:"type:varchar(255)" json:"description,omitempty"`
	CreatedAt   time.Time               `gorm:"autoCreateTime" json:"created_at"`
}

func (LoyaltyTransaction) TableName() string {
	return "loyalty_transactions"
}

