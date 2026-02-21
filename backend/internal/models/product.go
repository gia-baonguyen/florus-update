package models

import (
	"time"
)

type Product struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string    `gorm:"type:varchar(255);not null" json:"name"`
	Slug            string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Brand           string    `gorm:"type:varchar(100)" json:"brand,omitempty"`
	Price           float64   `gorm:"type:decimal(15,2);not null" json:"price"`
	OriginalPrice   *float64  `gorm:"type:decimal(15,2)" json:"original_price,omitempty"`
	CategoryID      uint      `gorm:"index;not null" json:"category_id"`
	Description     string    `gorm:"type:text" json:"description,omitempty"`
	ImageURL        string    `gorm:"type:varchar(500)" json:"image_url,omitempty"`
	Stock           int       `gorm:"default:0" json:"stock"`
	Rating          float64   `gorm:"type:decimal(3,2);default:0" json:"rating"`
	ReviewCount     int       `gorm:"default:0" json:"review_count"`
	AIScore         *float64  `gorm:"type:decimal(5,2)" json:"ai_score,omitempty"`
	AIRecommendType string    `gorm:"type:varchar(50)" json:"ai_recommend_type,omitempty"`
	IsActive        bool      `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Category    Category            `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tags        []ProductTag        `gorm:"foreignKey:ProductID" json:"tags,omitempty"`
	Ingredients []ProductIngredient `gorm:"foreignKey:ProductID" json:"ingredients,omitempty"`
	Images      []ProductImage      `gorm:"foreignKey:ProductID" json:"images,omitempty"`
}

type ProductImage struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID uint      `gorm:"index;not null" json:"product_id"`
	URL       string    `gorm:"type:varchar(500);not null" json:"url"`
	AltText   string    `gorm:"type:varchar(255)" json:"alt_text,omitempty"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	IsPrimary bool      `gorm:"default:false" json:"is_primary"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ProductImage) TableName() string {
	return "product_images"
}

func (Product) TableName() string {
	return "products"
}

// CalculateDiscount returns discount percentage if original price exists
func (p *Product) CalculateDiscount() float64 {
	if p.OriginalPrice == nil || *p.OriginalPrice <= 0 {
		return 0
	}
	return ((*p.OriginalPrice - p.Price) / *p.OriginalPrice) * 100
}

// IsInStock checks if product is available
func (p *Product) IsInStock() bool {
	return p.Stock > 0
}
