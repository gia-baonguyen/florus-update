package models

import "database/sql/driver"

type RelationshipType string

const (
	RelationSimilar            RelationshipType = "similar"
	RelationCoViewed           RelationshipType = "co_viewed"
	RelationFrequentlyBought   RelationshipType = "frequently_bought_with"
)

// Value implements driver.Valuer for Oracle compatibility
func (r RelationshipType) Value() (driver.Value, error) {
	return string(r), nil
}

// Scan implements sql.Scanner for Oracle compatibility
func (r *RelationshipType) Scan(value interface{}) error {
	if value == nil {
		*r = RelationSimilar
		return nil
	}
	switch v := value.(type) {
	case string:
		*r = RelationshipType(v)
	case []byte:
		*r = RelationshipType(string(v))
	}
	return nil
}

type ProductRelationship struct {
	ID               uint             `gorm:"primaryKey;autoIncrement" json:"id"`
	SourceProductID  uint             `gorm:"index;not null" json:"source_product_id"`
	TargetProductID  uint             `gorm:"not null" json:"target_product_id"`
	RelationshipType RelationshipType `gorm:"type:varchar(50);not null" json:"relationship_type"`
	Score            *float64         `gorm:"type:decimal(5,2)" json:"score,omitempty"`

	// Relations
	SourceProduct Product `gorm:"foreignKey:SourceProductID" json:"source_product,omitempty"`
	TargetProduct Product `gorm:"foreignKey:TargetProductID" json:"target_product,omitempty"`
}

func (ProductRelationship) TableName() string {
	return "product_relationships"
}
