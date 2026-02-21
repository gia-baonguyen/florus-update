package models

type ProductIngredient struct {
	ID             uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID      uint   `gorm:"index;not null" json:"product_id"`
	IngredientName string `gorm:"type:varchar(255);not null" json:"ingredient_name"`
}

func (ProductIngredient) TableName() string {
	return "product_ingredients"
}
