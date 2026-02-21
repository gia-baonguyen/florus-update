package models

type OrderItem struct {
	ID         uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID    uint    `gorm:"index;not null" json:"order_id"`
	ProductID  uint    `gorm:"not null" json:"product_id"`
	Quantity   int     `gorm:"not null" json:"quantity"`
	UnitPrice  float64 `gorm:"type:decimal(15,2);not null" json:"unit_price"`
	TotalPrice float64 `gorm:"type:decimal(15,2);not null" json:"total_price"`

	// Relations
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

// CalculateTotalPrice calculates total price for this item
func (oi *OrderItem) CalculateTotalPrice() float64 {
	return oi.UnitPrice * float64(oi.Quantity)
}
