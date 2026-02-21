package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type UserStatus string

const (
	StatusCold UserStatus = "cold"
	StatusWarm UserStatus = "warm"
)

type User struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	Name         string     `gorm:"type:varchar(100);not null" json:"name"`
	Phone        string     `gorm:"type:varchar(20)" json:"phone,omitempty"`
	Address      string     `gorm:"type:varchar(500)" json:"address,omitempty"`
	Role         UserRole   `gorm:"type:varchar(20);default:user" json:"role"`
	UserStatus   UserStatus `gorm:"type:varchar(20);default:cold" json:"user_status"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Orders    []Order    `gorm:"foreignKey:UserID" json:"orders,omitempty"`
	CartItems []CartItem `gorm:"foreignKey:UserID" json:"cart_items,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Role == "" {
		u.Role = RoleUser
	}
	if u.UserStatus == "" {
		u.UserStatus = StatusCold
	}
	return nil
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
