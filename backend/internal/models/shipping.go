package models

import "time"

// ShippingZone represents a geographical shipping area (e.g. HCM inner city).
type ShippingZone struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Description string    `gorm:"type:varchar(255)" json:"description,omitempty"`
	CountryCode string    `gorm:"type:varchar(10);default:'VN'" json:"country_code"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ShippingZone) TableName() string {
	return "shipping_zones"
}

// ShippingMethod represents a shipping method such as standard / express.
type ShippingMethod struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Code      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Description string  `gorm:"type:varchar(255)" json:"description,omitempty"`
	IsActive  bool      `gorm:"type:number(1);default:1" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ShippingMethod) TableName() string {
	return "shipping_methods"
}

// ShippingZoneMethod configures fees per zone/method combination.
type ShippingZoneMethod struct {
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ZoneID           uint      `gorm:"index;not null" json:"zone_id"`
	MethodID         uint      `gorm:"index;not null" json:"method_id"`
	BaseFee          float64   `gorm:"type:decimal(15,2);not null" json:"base_fee"`
	EstimatedDaysMin int       `json:"estimated_days_min"`
	EstimatedDaysMax int       `json:"estimated_days_max"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ShippingZoneMethod) TableName() string {
	return "shipping_zone_methods"
}

