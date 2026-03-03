package models

import (
	"database/sql/driver"
	"time"

	"gorm.io/gorm"
)

// EventType defines the types of user events
type EventType string

const (
	EventProductView    EventType = "product_view"
	EventProductClick   EventType = "product_click"
	EventAddToCart      EventType = "add_to_cart"
	EventRemoveFromCart EventType = "remove_from_cart"
	EventAddToWishlist  EventType = "add_to_wishlist"
	EventPurchase       EventType = "purchase"
	EventSearch         EventType = "search"
	EventCategoryView   EventType = "category_view"
	EventReview         EventType = "review"
	EventRating         EventType = "rating"
)

// Value implements driver.Valuer for Oracle compatibility
func (e EventType) Value() (driver.Value, error) {
	return string(e), nil
}

// Scan implements sql.Scanner for Oracle compatibility
func (e *EventType) Scan(value interface{}) error {
	if value == nil {
		*e = EventProductView
		return nil
	}
	switch v := value.(type) {
	case string:
		*e = EventType(v)
	case []byte:
		*e = EventType(string(v))
	}
	return nil
}

// UserEvent represents a user interaction event for recommendation system
type UserEvent struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     *uint          `gorm:"index" json:"user_id"`
	SessionID  string         `gorm:"index;size:100" json:"session_id"`
	EventType  EventType      `gorm:"index;size:50;not null" json:"event_type"`
	ProductID  *uint          `gorm:"index" json:"product_id"`
	CategoryID *uint          `gorm:"index" json:"category_id"`
	OrderID    *uint          `gorm:"index" json:"order_id"`
	SearchQuery string        `gorm:"size:255" json:"search_query,omitempty"`
	Quantity   int            `json:"quantity,omitempty"`
	Price      float64        `json:"price,omitempty"`
	Rating     *int           `json:"rating,omitempty"`
	Metadata   string         `gorm:"type:text" json:"metadata,omitempty"` // JSON string for extra data
	UserAgent  string         `gorm:"size:500" json:"user_agent,omitempty"`
	IPAddress  string         `gorm:"size:45" json:"ip_address,omitempty"`
	Referrer   string         `gorm:"size:500" json:"referrer,omitempty"`
	CreatedAt  time.Time      `gorm:"index" json:"created_at"`

	// Relations (optional, for eager loading)
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Product  *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// TableName specifies the table name
func (UserEvent) TableName() string {
	return "user_events"
}

// BeforeCreate hook
func (e *UserEvent) BeforeCreate(tx *gorm.DB) error {
	e.CreatedAt = time.Now()
	return nil
}

// UserEventStats represents aggregated event statistics
type UserEventStats struct {
	EventType    EventType `json:"event_type"`
	Count        int64     `json:"count"`
	UniqueUsers  int64     `json:"unique_users"`
	LastOccurred time.Time `json:"last_occurred"`
}

// ProductEventStats represents product-level event statistics
type ProductEventStats struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	Views       int64   `json:"views"`
	Clicks      int64   `json:"clicks"`
	AddToCarts  int64   `json:"add_to_carts"`
	Purchases   int64   `json:"purchases"`
	ConversionRate float64 `json:"conversion_rate"`
}

// UserBehaviorProfile represents a user's behavior summary for recommendations
type UserBehaviorProfile struct {
	UserID              uint      `json:"user_id"`
	TotalEvents         int64     `json:"total_events"`
	ProductViews        int64     `json:"product_views"`
	Purchases           int64     `json:"purchases"`
	AverageRating       float64   `json:"average_rating"`
	PreferredCategories []uint    `json:"preferred_categories"`
	RecentProducts      []uint    `json:"recent_products"`
	LastActivity        time.Time `json:"last_activity"`
}
