package models

import (
	"database/sql/driver"
	"time"

	"gorm.io/gorm"
)

type PaymentMethod string

const (
	PaymentMethodCOD     PaymentMethod = "cod"
	PaymentMethodZaloPay PaymentMethod = "zalopay"
	PaymentMethodMoMo    PaymentMethod = "momo"
	PaymentMethodVNPay   PaymentMethod = "vnpay"
)

// Value implements driver.Valuer for Oracle compatibility
func (m PaymentMethod) Value() (driver.Value, error) {
	return string(m), nil
}

// Scan implements sql.Scanner for Oracle compatibility
func (m *PaymentMethod) Scan(value interface{}) error {
	if value == nil {
		*m = PaymentMethodCOD
		return nil
	}
	switch v := value.(type) {
	case string:
		*m = PaymentMethod(v)
	case []byte:
		*m = PaymentMethod(string(v))
	}
	return nil
}

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSuccess   PaymentStatus = "success"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// Value implements driver.Valuer for Oracle compatibility
func (s PaymentStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan implements sql.Scanner for Oracle compatibility
func (s *PaymentStatus) Scan(value interface{}) error {
	if value == nil {
		*s = PaymentStatusPending
		return nil
	}
	switch v := value.(type) {
	case string:
		*s = PaymentStatus(v)
	case []byte:
		*s = PaymentStatus(string(v))
	}
	return nil
}

type Payment struct {
	ID              uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID         uint          `gorm:"index;not null" json:"order_id"`
	TransactionID   string        `gorm:"type:varchar(100);uniqueIndex" json:"transaction_id"`
	Method          PaymentMethod `gorm:"type:varchar(20);not null" json:"method"`
	Amount          float64       `gorm:"type:decimal(15,2);not null" json:"amount"`
	Status          PaymentStatus `gorm:"type:varchar(20);default:pending" json:"status"`
	ProviderTransID string        `gorm:"type:varchar(100)" json:"provider_trans_id,omitempty"`
	PaymentURL      string        `gorm:"type:varchar(1000)" json:"payment_url,omitempty"`
	CallbackData    string        `gorm:"type:text" json:"-"`
	ErrorMessage    string        `gorm:"type:varchar(500)" json:"error_message,omitempty"`
	PaidAt          *time.Time    `json:"paid_at,omitempty"`
	CreatedAt       time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time     `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Order Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

func (Payment) TableName() string {
	return "payments"
}

func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.Status == "" {
		p.Status = PaymentStatusPending
	}
	return nil
}

func (p *Payment) MarkAsSuccess(providerTransID string) {
	p.Status = PaymentStatusSuccess
	p.ProviderTransID = providerTransID
	now := time.Now()
	p.PaidAt = &now
}

func (p *Payment) MarkAsFailed(errorMsg string) {
	p.Status = PaymentStatusFailed
	p.ErrorMessage = errorMsg
}

func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusPending
}

func (p *Payment) IsSuccess() bool {
	return p.Status == PaymentStatusSuccess
}
