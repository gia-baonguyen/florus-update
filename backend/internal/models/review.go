package models

import (
	"time"
)

type Review struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	ProductID uint      `gorm:"index;not null" json:"product_id"`
	Rating    int       `gorm:"not null" json:"rating"` // 1-5
	Comment   string    `gorm:"column:COMMENT_TEXT;type:text" json:"comment"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	User    User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Product Product       `gorm:"foreignKey:ProductID" json:"-"`
	Images  []ReviewImage `gorm:"foreignKey:ReviewID" json:"images,omitempty"`
}

func (Review) TableName() string {
	return "reviews"
}

// ReviewImage stores images attached to a review
type ReviewImage struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ReviewID  uint      `gorm:"index;not null" json:"review_id"`
	URL       string    `gorm:"type:varchar(500);not null" json:"url"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ReviewImage) TableName() string {
	return "review_images"
}
