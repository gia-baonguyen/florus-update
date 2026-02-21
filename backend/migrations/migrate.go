package migrations

import (
	"log"

	"github.com/florus/backend/internal/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	log.Println("Running auto-migration...")

	err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.ProductTag{},
		&models.ProductIngredient{},
		&models.ProductRelationship{},
		&models.ProductImage{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.WishlistItem{},
		&models.Review{},
		&models.Coupon{},
	)

	if err != nil {
		log.Printf("Migration failed: %v", err)
		return err
	}

	log.Println("Migration completed successfully")
	return nil
}
