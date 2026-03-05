package migrations

import (
	"fmt"
	"log"
	"strings"

	"github.com/florus/backend/internal/models"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	log.Println("Starting database seeding...")

	// Check if data already exists
	var categoryCount int64
	db.Model(&models.Category{}).Count(&categoryCount)
	if categoryCount > 0 {
		log.Println("Database already seeded, skipping...")
		return nil
	}

	// Seed Admin User
	if err := seedAdminUser(db); err != nil {
		return err
	}

	// Seed Categories
	categoryBySlug, err := seedCategories(db)
	if err != nil {
		return err
	}

	// Seed Products
	productBySlug, err := seedProducts(db, categoryBySlug)
	if err != nil {
		return err
	}

	// Seed Product Relationships
	if err := seedProductRelationships(db, productBySlug); err != nil {
		return err
	}

	log.Println("Database seeding completed successfully")
	return nil
}

// ForceSeed wipes existing data then seeds again.
// This is safe to run multiple times and does not depend on identity values staying at 1..N.
func ForceSeed(db *gorm.DB) error {
	log.Println("Starting database reseeding (wipe + seed)...")
	if err := wipeSeedData(db); err != nil {
		return err
	}
	return Seed(db)
}

// wipeSeedData removes data in FK-safe order. It does NOT reset Oracle identity counters.
func wipeSeedData(db *gorm.DB) error {
	// Child tables first
	stmts := []string{
		"DELETE FROM product_relationships",
		"DELETE FROM product_images",
		"DELETE FROM product_ingredients",
		"DELETE FROM product_tags",
		"DELETE FROM reviews",
		"DELETE FROM wishlist_items",
		"DELETE FROM cart_items",
		"DELETE FROM order_items",
		"DELETE FROM payments",
		"DELETE FROM orders",
		"DELETE FROM coupons",
		"DELETE FROM user_events",
		// Parents
		"DELETE FROM products",
		"DELETE FROM categories",
		"DELETE FROM users",
	}

	for _, sql := range stmts {
		if err := db.Exec(sql).Error; err != nil {
			// Some tables may not exist yet in a partially-migrated Oracle schema.
			// Ignore and continue so reseed can proceed after migrations create them.
			if strings.Contains(err.Error(), "ORA-00942") {
				continue
			}
			return fmt.Errorf("wipe failed on %q: %w", sql, err)
		}
	}

	log.Println("Database wipe completed")
	return nil
}

func seedAdminUser(db *gorm.DB) error {
	admin := &models.User{
		Email: "admin@florus.com",
		Name:  "Admin",
		Role:  models.RoleAdmin,
	}
	admin.SetPassword("admin123")

	if err := db.Create(admin).Error; err != nil {
		log.Printf("Failed to seed admin user: %v", err)
		return err
	}

	log.Println("Admin user seeded: admin@florus.com / admin123")
	return nil
}

func seedCategories(db *gorm.DB) (map[string]models.Category, error) {
	categories := []models.Category{
		{Name: "Serum", Slug: "serum", Description: "Face serums for skincare", IsActive: true},
		{Name: "Moisturizer", Slug: "moisturizer", Description: "Face moisturizers", IsActive: true},
		{Name: "Lipstick", Slug: "lipstick", Description: "Lip colors and lipsticks", IsActive: true},
		{Name: "Eye Cream", Slug: "eye-cream", Description: "Eye care products", IsActive: true},
		{Name: "Foundation", Slug: "foundation", Description: "Face foundations", IsActive: true},
		{Name: "Cleanser", Slug: "cleanser", Description: "Face cleansers", IsActive: true},
		{Name: "Mascara", Slug: "mascara", Description: "Eye mascaras", IsActive: true},
		{Name: "Blush", Slug: "blush", Description: "Cheek blushes", IsActive: true},
		{Name: "Perfume", Slug: "perfume", Description: "Fragrances and perfumes", IsActive: true},
		{Name: "Sunscreen", Slug: "sunscreen", Description: "Sun protection products", IsActive: true},
		{Name: "Mask", Slug: "mask", Description: "Face masks", IsActive: true},
		{Name: "Nail Polish", Slug: "nail-polish", Description: "Nail polishes", IsActive: true},
		{Name: "Eyeshadow", Slug: "eyeshadow", Description: "Eye shadow palettes", IsActive: true},
	}

	for _, cat := range categories {
		if err := db.Create(&cat).Error; err != nil {
			log.Printf("Failed to seed category %s: %v", cat.Name, err)
			return nil, err
		}
	}

	log.Printf("Seeded %d categories", len(categories))

	var persisted []models.Category
	if err := db.Find(&persisted).Error; err != nil {
		return nil, err
	}
	bySlug := make(map[string]models.Category, len(persisted))
	for _, c := range persisted {
		bySlug[c.Slug] = c
	}
	return bySlug, nil
}

func seedProducts(db *gorm.DB, categoriesBySlug map[string]models.Category) (map[string]models.Product, error) {
	// Helper function to create float pointer
	floatPtr := func(f float64) *float64 { return &f }

	products := []struct {
		Product       models.Product
		CategorySlug  string
		Tags        []string
		Ingredients []string
	}{
		{
			Product: models.Product{
				Name: "Vitamin C Brightening Serum", Slug: "vitamin-c-brightening-serum",
				Brand: "Radiance Labs", Price: 580000, OriginalPrice: floatPtr(650000),
				Description: "A powerful vitamin C serum that brightens skin and reduces dark spots",
				ImageURL: "https://images.unsplash.com/photo-1620916566398-39f1143ab7be?w=400",
				Stock: 100, Rating: 4.8, ReviewCount: 256,
				AIScore: floatPtr(94), AIRecommendType: "cold-start", IsActive: true,
			},
			CategorySlug: "serum",
			Tags:        []string{"Brightening", "Anti-aging", "Vitamin C"},
			Ingredients: []string{"Vitamin C", "Hyaluronic Acid", "Niacinamide"},
		},
		{
			Product: models.Product{
				Name: "Hyaluronic Acid Moisturizer", Slug: "hyaluronic-acid-moisturizer",
				Brand: "Aqua Glow", Price: 420000,
				Description: "Deep hydrating moisturizer with hyaluronic acid",
				ImageURL: "https://images.unsplash.com/photo-1611930022073-b7a4ba5fcccd?w=400",
				Stock: 150, Rating: 4.6, ReviewCount: 189,
				AIScore: floatPtr(91), AIRecommendType: "warm-start", IsActive: true,
			},
			CategorySlug: "moisturizer",
			Tags:        []string{"Hydrating", "Daily Use", "All Skin Types"},
			Ingredients: []string{"Hyaluronic Acid", "Glycerin", "Aloe Vera"},
		},
		{
			Product: models.Product{
				Name: "Matte Lipstick - Coral Dream", Slug: "matte-lipstick-coral-dream",
				Brand: "Velvet Kiss", Price: 290000, OriginalPrice: floatPtr(350000),
				Description: "Long-lasting matte lipstick in coral shade",
				ImageURL: "https://images.unsplash.com/photo-1586495777744-4413f21062fa?w=400",
				Stock: 200, Rating: 4.5, ReviewCount: 312,
				AIScore: floatPtr(88), AIRecommendType: "cold-start", IsActive: true,
			},
			CategorySlug: "lipstick",
			Tags:        []string{"Matte", "Long-lasting", "Coral"},
			Ingredients: []string{"Vitamin E", "Jojoba Oil"},
		},
		{
			Product: models.Product{
				Name: "Retinol Eye Cream", Slug: "retinol-eye-cream",
				Brand: "Youth Restore", Price: 650000,
				Description: "Anti-aging eye cream with retinol",
				ImageURL: "https://images.unsplash.com/photo-1570194065650-d99fb4b38b15?w=400",
				Stock: 80, Rating: 4.7, ReviewCount: 145,
				AIScore: floatPtr(92), AIRecommendType: "content-based", IsActive: true,
			},
			CategorySlug: "eye-cream",
			Tags:        []string{"Anti-aging", "Retinol", "Eye Care"},
			Ingredients: []string{"Retinol", "Peptides", "Caffeine"},
		},
		{
			Product: models.Product{
				Name: "Flawless Foundation SPF 30", Slug: "flawless-foundation-spf-30",
				Brand: "Perfect Canvas", Price: 520000,
				Description: "Full coverage foundation with sun protection",
				ImageURL: "https://images.unsplash.com/photo-1631214524020-7e18db9a8f92?w=400",
				Stock: 120, Rating: 4.4, ReviewCount: 278,
				AIScore: floatPtr(89), AIRecommendType: "warm-start", IsActive: true,
			},
			CategorySlug: "foundation",
			Tags:        []string{"Full Coverage", "SPF 30", "Long-wear"},
			Ingredients: []string{"SPF 30", "Hyaluronic Acid", "Vitamin E"},
		},
		{
			Product: models.Product{
				Name: "Gentle Foam Cleanser", Slug: "gentle-foam-cleanser",
				Brand: "Pure Clean", Price: 280000,
				Description: "Gentle foaming cleanser for all skin types",
				ImageURL: "https://images.unsplash.com/photo-1556228720-195a672e8a03?w=400",
				Stock: 180, Rating: 4.3, ReviewCount: 423,
				AIScore: floatPtr(86), AIRecommendType: "collaborative", IsActive: true,
			},
			CategorySlug: "cleanser",
			Tags:        []string{"Gentle", "Foaming", "Daily Use"},
			Ingredients: []string{"Aloe Vera", "Green Tea", "Chamomile"},
		},
		{
			Product: models.Product{
				Name: "Volume Mascara - Midnight Black", Slug: "volume-mascara-midnight-black",
				Brand: "Lash Perfect", Price: 320000, OriginalPrice: floatPtr(380000),
				Description: "Volumizing mascara for dramatic lashes",
				ImageURL: "https://images.unsplash.com/photo-1631214540553-ff044a3ff1ea?w=400",
				Stock: 160, Rating: 4.6, ReviewCount: 567,
				AIScore: floatPtr(90), AIRecommendType: "cold-start", IsActive: true,
			},
			CategorySlug: "mascara",
			Tags:        []string{"Volumizing", "Waterproof", "Black"},
			Ingredients: []string{"Beeswax", "Vitamin E", "Keratin"},
		},
		{
			Product: models.Product{
				Name: "Baked Blush - Rose Petal", Slug: "baked-blush-rose-petal",
				Brand: "Glow Studio", Price: 350000,
				Description: "Natural-looking baked blush in rose shade",
				ImageURL: "https://images.unsplash.com/photo-1596704017254-9b121068fb31?w=400",
				Stock: 140, Rating: 4.5, ReviewCount: 234,
				AIScore: floatPtr(87), AIRecommendType: "cold-start", IsActive: true,
			},
			CategorySlug: "blush",
			Tags:        []string{"Natural", "Rose", "Buildable"},
			Ingredients: []string{"Mineral Pigments", "Vitamin E"},
		},
		{
			Product: models.Product{
				Name: "Eau de Parfum - Garden Rose", Slug: "eau-de-parfum-garden-rose",
				Brand: "Essence Luxe", Price: 980000, OriginalPrice: floatPtr(1200000),
				Description: "Elegant floral perfume with rose notes",
				ImageURL: "https://images.unsplash.com/photo-1541643600914-78b084683601?w=400",
				Stock: 60, Rating: 4.9, ReviewCount: 89,
				AIScore: floatPtr(96), AIRecommendType: "cold-start", IsActive: true,
			},
			CategorySlug: "perfume",
			Tags:        []string{"Floral", "Long-lasting", "Elegant"},
			Ingredients: []string{"Rose Extract", "Jasmine", "Musk"},
		},
		{
			Product: models.Product{
				Name: "Mineral Sunscreen SPF 50+", Slug: "mineral-sunscreen-spf-50",
				Brand: "Sun Shield", Price: 380000,
				Description: "Lightweight mineral sunscreen with high protection",
				ImageURL: "https://images.unsplash.com/photo-1556227702-d1e4e7b5c232?w=400",
				Stock: 200, Rating: 4.7, ReviewCount: 345,
				AIScore: floatPtr(93), AIRecommendType: "warm-start", IsActive: true,
			},
			CategorySlug: "sunscreen",
			Tags:        []string{"SPF 50+", "Mineral", "Lightweight"},
			Ingredients: []string{"Zinc Oxide", "Titanium Dioxide", "Vitamin E"},
		},
		{
			Product: models.Product{
				Name: "Clay Detox Mask", Slug: "clay-detox-mask",
				Brand: "Skin Therapy", Price: 450000,
				Description: "Deep cleansing clay mask for pore refinement",
				ImageURL: "https://images.unsplash.com/photo-1596755389378-c31d21fd1273?w=400",
				Stock: 90, Rating: 4.4, ReviewCount: 178,
				AIScore: floatPtr(85), AIRecommendType: "content-based", IsActive: true,
			},
			CategorySlug: "mask",
			Tags:        []string{"Detox", "Clay", "Pore Refining"},
			Ingredients: []string{"Kaolin Clay", "Charcoal", "Tea Tree Oil"},
		},
		{
			Product: models.Product{
				Name: "Gel Nail Polish - Nude Rose", Slug: "gel-nail-polish-nude-rose",
				Brand: "Nail Couture", Price: 180000,
				Description: "Long-lasting gel nail polish in nude rose",
				ImageURL: "https://images.unsplash.com/photo-1604654894610-df63bc536371?w=400",
				Stock: 250, Rating: 4.2, ReviewCount: 156,
				AIScore: floatPtr(84), AIRecommendType: "cross-sell", IsActive: true,
			},
			CategorySlug: "nail-polish",
			Tags:        []string{"Gel", "Nude", "Long-lasting"},
			Ingredients: []string{"Gel Formula", "Vitamin E"},
		},
		{
			Product: models.Product{
				Name: "Eyeshadow Palette - Natural Nude", Slug: "eyeshadow-palette-natural-nude",
				Brand: "Color Art", Price: 620000, OriginalPrice: floatPtr(750000),
				Description: "12-shade eyeshadow palette with natural nude tones",
				ImageURL: "https://images.unsplash.com/photo-1512496015851-a90fb38ba796?w=400",
				Stock: 70, Rating: 4.8, ReviewCount: 289,
				AIScore: floatPtr(95), AIRecommendType: "warm-start", IsActive: true,
			},
			CategorySlug: "eyeshadow",
			Tags:        []string{"Palette", "Nude", "Matte & Shimmer"},
			Ingredients: []string{"Mineral Pigments", "Vitamin E", "Jojoba Oil"},
		},
	}

	for _, p := range products {
		cat, ok := categoriesBySlug[p.CategorySlug]
		if !ok {
			return nil, fmt.Errorf("unknown category slug %q for product %q", p.CategorySlug, p.Product.Slug)
		}
		p.Product.CategoryID = cat.ID

		if err := db.Create(&p.Product).Error; err != nil {
			log.Printf("Failed to seed product %s: %v", p.Product.Name, err)
			return nil, err
		}

		// Create tags
		for _, tag := range p.Tags {
			db.Create(&models.ProductTag{ProductID: p.Product.ID, TagName: tag})
		}

		// Create ingredients
		for _, ing := range p.Ingredients {
			db.Create(&models.ProductIngredient{ProductID: p.Product.ID, IngredientName: ing})
		}
	}

	log.Printf("Seeded %d products with tags and ingredients", len(products))

	var persisted []models.Product
	if err := db.Find(&persisted).Error; err != nil {
		return nil, err
	}
	bySlug := make(map[string]models.Product, len(persisted))
	for _, p := range persisted {
		bySlug[p.Slug] = p
	}
	return bySlug, nil
}

func seedProductRelationships(db *gorm.DB, productsBySlug map[string]models.Product) error {
	floatPtr := func(f float64) *float64 { return &f }

	// Similar products (content-based)
	type relSeed struct {
		SourceSlug string
		TargetSlug string
		Type       models.RelationshipType
		Score      float64
	}

	similarSeeds := []relSeed{
		{SourceSlug: "vitamin-c-brightening-serum", TargetSlug: "eyeshadow-palette-natural-nude", Type: models.RelationSimilar, Score: 0.92},
		{SourceSlug: "vitamin-c-brightening-serum", TargetSlug: "eau-de-parfum-garden-rose", Type: models.RelationSimilar, Score: 0.88},
		{SourceSlug: "vitamin-c-brightening-serum", TargetSlug: "clay-detox-mask", Type: models.RelationSimilar, Score: 0.85},
		{SourceSlug: "hyaluronic-acid-moisturizer", TargetSlug: "retinol-eye-cream", Type: models.RelationSimilar, Score: 0.90},
		{SourceSlug: "hyaluronic-acid-moisturizer", TargetSlug: "mineral-sunscreen-spf-50", Type: models.RelationSimilar, Score: 0.87},
		{SourceSlug: "matte-lipstick-coral-dream", TargetSlug: "baked-blush-rose-petal", Type: models.RelationSimilar, Score: 0.91},
		{SourceSlug: "matte-lipstick-coral-dream", TargetSlug: "volume-mascara-midnight-black", Type: models.RelationSimilar, Score: 0.86},
	}

	// Co-viewed products (collaborative)
	coViewedSeeds := []relSeed{
		{SourceSlug: "vitamin-c-brightening-serum", TargetSlug: "hyaluronic-acid-moisturizer", Type: models.RelationCoViewed, Score: 0.94},
		{SourceSlug: "vitamin-c-brightening-serum", TargetSlug: "gentle-foam-cleanser", Type: models.RelationCoViewed, Score: 0.89},
		{SourceSlug: "hyaluronic-acid-moisturizer", TargetSlug: "vitamin-c-brightening-serum", Type: models.RelationCoViewed, Score: 0.93},
		{SourceSlug: "matte-lipstick-coral-dream", TargetSlug: "baked-blush-rose-petal", Type: models.RelationCoViewed, Score: 0.91},
		{SourceSlug: "flawless-foundation-spf-30", TargetSlug: "baked-blush-rose-petal", Type: models.RelationCoViewed, Score: 0.88},
	}

	// Cross-sell products (frequently bought together)
	crossSellSeeds := []relSeed{
		{SourceSlug: "vitamin-c-brightening-serum", TargetSlug: "gentle-foam-cleanser", Type: models.RelationFrequentlyBought, Score: 0.95},
		{SourceSlug: "vitamin-c-brightening-serum", TargetSlug: "hyaluronic-acid-moisturizer", Type: models.RelationFrequentlyBought, Score: 0.92},
		{SourceSlug: "matte-lipstick-coral-dream", TargetSlug: "baked-blush-rose-petal", Type: models.RelationFrequentlyBought, Score: 0.90},
		{SourceSlug: "flawless-foundation-spf-30", TargetSlug: "mineral-sunscreen-spf-50", Type: models.RelationFrequentlyBought, Score: 0.88},
		{SourceSlug: "volume-mascara-midnight-black", TargetSlug: "eyeshadow-palette-natural-nude", Type: models.RelationFrequentlyBought, Score: 0.87},
	}

	allSeeds := append(similarSeeds, coViewedSeeds...)
	allSeeds = append(allSeeds, crossSellSeeds...)

	var created int
	for _, s := range allSeeds {
		src, ok1 := productsBySlug[s.SourceSlug]
		dst, ok2 := productsBySlug[s.TargetSlug]
		if !ok1 || !ok2 {
			log.Printf("Skipping relationship (missing product): %s -> %s", s.SourceSlug, s.TargetSlug)
			continue
		}

		rel := models.ProductRelationship{
			SourceProductID:  src.ID,
			TargetProductID:  dst.ID,
			RelationshipType: s.Type,
			Score:            floatPtr(s.Score),
		}
		if err := db.Create(&rel).Error; err != nil {
			log.Printf("Failed to seed relationship: %v", err)
			continue
		}
		created++
	}

	log.Printf("Seeded %d product relationships", created)
	return nil
}
