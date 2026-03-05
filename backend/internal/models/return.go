package models

import "time"

// ReturnStatus represents status of a return / refund request.
type ReturnStatus string

const (
	ReturnStatusRequested ReturnStatus = "requested"
	ReturnStatusApproved  ReturnStatus = "approved"
	ReturnStatusRejected  ReturnStatus = "rejected"
	ReturnStatusReceived  ReturnStatus = "received"
	ReturnStatusRefunded  ReturnStatus = "refunded"
)

type Return struct {
	ID        uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID   uint         `gorm:"index;not null" json:"order_id"`
	UserID    uint         `gorm:"index;not null" json:"user_id"`
	Status    ReturnStatus `gorm:"type:varchar(20);default:'requested'" json:"status"`
	Reason    string       `gorm:"type:varchar(500)" json:"reason,omitempty"`
	Note      string       `gorm:"type:varchar(500)" json:"note,omitempty"`
	CreatedAt time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time    `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Order  Order         `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Items  []ReturnItem  `gorm:"foreignKey:ReturnID" json:"items,omitempty"`
}

func (Return) TableName() string {
	return "returns"
}

type ReturnItem struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ReturnID     uint      `gorm:"index;not null" json:"return_id"`
	OrderItemID  uint      `gorm:"index;not null" json:"order_item_id"`
	Quantity     int       `gorm:"not null" json:"quantity"`
	RefundAmount float64   `gorm:"type:decimal(15,2);not null" json:"refund_amount"`
	Status       string    `gorm:"type:varchar(20)" json:"status,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	OrderItem OrderItem `gorm:"foreignKey:OrderItemID" json:"order_item,omitempty"`
}

func (ReturnItem) TableName() string {
	return "return_items"
}

