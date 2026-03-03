package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipping   OrderStatus = "shipping"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

// Value implements driver.Valuer for Oracle compatibility
func (s OrderStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan implements sql.Scanner for Oracle compatibility
func (s *OrderStatus) Scan(value interface{}) error {
	if value == nil {
		*s = OrderStatusPending
		return nil
	}
	switch v := value.(type) {
	case string:
		*s = OrderStatus(v)
	case []byte:
		*s = OrderStatus(string(v))
	}
	return nil
}

const (
	FreeShippingThreshold = 5000000 // 5M VND
	ShippingFee           = 50000   // 50K VND
)

type Order struct {
	ID              uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint          `gorm:"index;not null" json:"user_id"`
	OrderCode       string        `gorm:"type:varchar(50);uniqueIndex;not null" json:"order_code"`
	Subtotal        float64       `gorm:"type:decimal(15,2);not null" json:"subtotal"`
	ShippingFee     float64       `gorm:"type:decimal(15,2);default:0" json:"shipping_fee"`
	Discount        float64       `gorm:"type:decimal(15,2);default:0" json:"discount"`
	Total           float64       `gorm:"type:decimal(15,2);not null" json:"total"`
	Status          OrderStatus   `gorm:"type:varchar(50);default:pending" json:"status"`
	PaymentMethod   PaymentMethod `gorm:"type:varchar(20);default:cod" json:"payment_method"`
	PaymentStatus   PaymentStatus `gorm:"type:varchar(20);default:pending" json:"payment_status"`
	ShippingAddress string        `gorm:"type:varchar(500)" json:"shipping_address,omitempty"`
	Note            string        `gorm:"type:varchar(500)" json:"note,omitempty"`
	CreatedAt       time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time     `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	User       User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	OrderItems []OrderItem `gorm:"foreignKey:OrderID" json:"order_items,omitempty"`
	Payment    *Payment    `gorm:"foreignKey:OrderID" json:"payment,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.OrderCode == "" {
		o.OrderCode = generateOrderCode()
	}
	if o.Status == "" {
		o.Status = OrderStatusPending
	}
	if o.PaymentMethod == "" {
		o.PaymentMethod = PaymentMethodCOD
	}
	if o.PaymentStatus == "" {
		o.PaymentStatus = PaymentStatusPending
	}
	return nil
}

func generateOrderCode() string {
	return fmt.Sprintf("ORD-%s", uuid.New().String()[:8])
}

// CalculateShippingFee returns shipping fee based on subtotal
func CalculateShippingFee(subtotal float64) float64 {
	if subtotal >= FreeShippingThreshold {
		return 0
	}
	return ShippingFee
}

// CanCancel checks if order can be cancelled
func (o *Order) CanCancel() bool {
	return o.Status == OrderStatusPending || o.Status == OrderStatusConfirmed
}
