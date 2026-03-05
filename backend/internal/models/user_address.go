package models

import "time"

// UserAddress represents a saved shipping address for a user.
// This is intentionally generic so it can work for both VN and
// international addresses if needed later.
type UserAddress struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	FullName    string    `gorm:"type:varchar(100);not null" json:"full_name"`
	Phone       string    `gorm:"type:varchar(20);not null" json:"phone"`
	Street      string    `gorm:"type:varchar(255);not null" json:"street"`
	City        string    `gorm:"type:varchar(100);not null" json:"city"`
	State       string    `gorm:"type:varchar(100)" json:"state"`
	PostalCode  string    `gorm:"type:varchar(20)" json:"postal_code"`
	CountryCode string    `gorm:"type:varchar(10);default:'VN'" json:"country_code"`
	IsDefault   bool      `gorm:"type:number(1);default:0" json:"is_default"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserAddress) TableName() string {
	return "user_addresses"
}

