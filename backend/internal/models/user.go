package models

import (
	"database/sql/driver"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// Value implements driver.Valuer for Oracle compatibility
func (r UserRole) Value() (driver.Value, error) {
	return string(r), nil
}

// Scan implements sql.Scanner for Oracle compatibility
func (r *UserRole) Scan(value interface{}) error {
	if value == nil {
		*r = RoleUser
		return nil
	}
	switch v := value.(type) {
	case string:
		*r = UserRole(v)
	case []byte:
		*r = UserRole(string(v))
	}
	return nil
}

type UserStatus string

const (
	StatusCold UserStatus = "cold"
	StatusWarm UserStatus = "warm"
)

// Value implements driver.Valuer for Oracle compatibility
func (s UserStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan implements sql.Scanner for Oracle compatibility
func (s *UserStatus) Scan(value interface{}) error {
	if value == nil {
		*s = StatusCold
		return nil
	}
	switch v := value.(type) {
	case string:
		*s = UserStatus(v)
	case []byte:
		*s = UserStatus(string(v))
	}
	return nil
}

type AuthProvider string

const (
	AuthProviderLocal  AuthProvider = "local"
	AuthProviderGoogle AuthProvider = "google"
)

// Value implements driver.Valuer for Oracle compatibility
func (a AuthProvider) Value() (driver.Value, error) {
	return string(a), nil
}

// Scan implements sql.Scanner for Oracle compatibility
func (a *AuthProvider) Scan(value interface{}) error {
	if value == nil {
		*a = AuthProviderLocal
		return nil
	}
	switch v := value.(type) {
	case string:
		*a = AuthProvider(v)
	case []byte:
		*a = AuthProvider(string(v))
	}
	return nil
}

type User struct {
	ID           uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        string       `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string       `gorm:"type:varchar(255)" json:"-"`
	Name         string       `gorm:"type:varchar(100);not null" json:"name"`
	Phone        string       `gorm:"type:varchar(20)" json:"phone,omitempty"`
	Address      string       `gorm:"type:varchar(500)" json:"address,omitempty"`
	Role         UserRole     `gorm:"type:varchar(20);default:user" json:"role"`
	UserStatus   UserStatus   `gorm:"type:varchar(20);default:cold" json:"user_status"`
	AuthProvider AuthProvider `gorm:"type:varchar(20);default:local" json:"auth_provider"`
	GoogleID     string       `gorm:"type:varchar(255);index" json:"-"`
	AvatarURL    string       `gorm:"type:varchar(500)" json:"avatar_url,omitempty"`
	LoyaltyTier  string       `gorm:"type:varchar(20);default:'Bronze'" json:"loyalty_tier"`
	LoyaltyPoints int64       `gorm:"type:number(10);default:0" json:"loyalty_points"`
	CreatedAt    time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time    `gorm:"autoUpdateTime" json:"updated_at"`

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
	if u.AuthProvider == "" {
		u.AuthProvider = AuthProviderLocal
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

func (u *User) IsOAuthUser() bool {
	return u.AuthProvider != AuthProviderLocal
}
