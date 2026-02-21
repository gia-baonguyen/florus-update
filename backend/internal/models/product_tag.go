package models

type ProductTag struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID uint   `gorm:"index;not null" json:"product_id"`
	TagName   string `gorm:"type:varchar(100);not null" json:"tag_name"`
}

func (ProductTag) TableName() string {
	return "product_tags"
}
