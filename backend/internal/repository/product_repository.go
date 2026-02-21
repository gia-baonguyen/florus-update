package repository

import (
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/pkg/utils"
	"gorm.io/gorm"
)

type ProductFilter struct {
	CategoryID *uint
	Search     string
	MinPrice   *float64
	MaxPrice   *float64
	Brand      string
	SortBy     string // newest, price_asc, price_desc, popular, rating
}

type ProductRepository interface {
	Create(product *models.Product) error
	FindByID(id uint) (*models.Product, error)
	FindBySlug(slug string) (*models.Product, error)
	FindAll(pagination utils.Pagination, categoryID *uint, search string) ([]models.Product, int64, error)
	FindAllWithFilter(pagination utils.Pagination, filter ProductFilter) ([]models.Product, int64, error)
	FindByCategory(categoryID uint, pagination utils.Pagination) ([]models.Product, int64, error)
	Update(product *models.Product) error
	Delete(id uint) error
	UpdateStock(tx *gorm.DB, productID uint, quantity int) error
	FindByIDWithLock(tx *gorm.DB, id uint) (*models.Product, error)
	ExistsBySlug(slug string) bool
	GetAllBrands() ([]string, error)

	// Tags and Ingredients
	CreateTags(productID uint, tags []string) error
	CreateIngredients(productID uint, ingredients []string) error
	DeleteTags(productID uint) error
	DeleteIngredients(productID uint) error

	// Images
	CreateImages(productID uint, images []models.ProductImage) error
	DeleteImages(productID uint) error
	GetImages(productID uint) ([]models.ProductImage, error)

	// Recommendations
	FindByIDs(ids []uint) ([]models.Product, error)
	GetRelatedProducts(productID uint, relationType models.RelationshipType, limit int) ([]models.Product, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) FindByID(id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.Preload("Category").Preload("Tags").Preload("Ingredients").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) FindBySlug(slug string) (*models.Product, error) {
	var product models.Product
	err := r.db.Preload("Category").Preload("Tags").Preload("Ingredients").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Where("slug = ?", slug).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) FindAll(pagination utils.Pagination, categoryID *uint, search string) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.Model(&models.Product{}).Where("is_active = ?", true)

	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	if search != "" {
		query = query.Where("name LIKE ? OR brand LIKE ? OR description LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("Category").Preload("Tags").Preload("Ingredients").
		Offset(pagination.GetOffset()).
		Limit(pagination.Limit).
		Order("created_at DESC").
		Find(&products).Error

	return products, total, err
}

func (r *productRepository) FindByCategory(categoryID uint, pagination utils.Pagination) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.Model(&models.Product{}).
		Where("category_id = ? AND is_active = ?", categoryID, true)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("Category").Preload("Tags").Preload("Ingredients").
		Offset(pagination.GetOffset()).
		Limit(pagination.Limit).
		Order("created_at DESC").
		Find(&products).Error

	return products, total, err
}

func (r *productRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}

func (r *productRepository) Delete(id uint) error {
	return r.db.Delete(&models.Product{}, id).Error
}

func (r *productRepository) UpdateStock(tx *gorm.DB, productID uint, quantity int) error {
	return tx.Model(&models.Product{}).
		Where("id = ?", productID).
		Update("stock", gorm.Expr("stock - ?", quantity)).Error
}

func (r *productRepository) FindByIDWithLock(tx *gorm.DB, id uint) (*models.Product, error) {
	var product models.Product
	// SQLite uses database-level locking, so we just do a regular SELECT
	// For production with Oracle/PostgreSQL, use: tx.Clauses(clause.Locking{Strength: "UPDATE"})
	err := tx.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) ExistsBySlug(slug string) bool {
	var count int64
	r.db.Model(&models.Product{}).Where("slug = ?", slug).Count(&count)
	return count > 0
}

func (r *productRepository) CreateTags(productID uint, tags []string) error {
	for _, tag := range tags {
		if err := r.db.Create(&models.ProductTag{ProductID: productID, TagName: tag}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *productRepository) CreateIngredients(productID uint, ingredients []string) error {
	for _, ing := range ingredients {
		if err := r.db.Create(&models.ProductIngredient{ProductID: productID, IngredientName: ing}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *productRepository) DeleteTags(productID uint) error {
	return r.db.Where("product_id = ?", productID).Delete(&models.ProductTag{}).Error
}

func (r *productRepository) DeleteIngredients(productID uint) error {
	return r.db.Where("product_id = ?", productID).Delete(&models.ProductIngredient{}).Error
}

func (r *productRepository) FindByIDs(ids []uint) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Preload("Category").Preload("Tags").Preload("Ingredients").
		Where("id IN ? AND is_active = ?", ids, true).
		Find(&products).Error
	return products, err
}

func (r *productRepository) GetRelatedProducts(productID uint, relationType models.RelationshipType, limit int) ([]models.Product, error) {
	var relationships []models.ProductRelationship
	err := r.db.Where("source_product_id = ? AND relationship_type = ?", productID, relationType).
		Order("score DESC").
		Limit(limit).
		Find(&relationships).Error
	if err != nil {
		return nil, err
	}

	var productIDs []uint
	for _, rel := range relationships {
		productIDs = append(productIDs, rel.TargetProductID)
	}

	if len(productIDs) == 0 {
		return []models.Product{}, nil
	}

	return r.FindByIDs(productIDs)
}

func (r *productRepository) FindAllWithFilter(pagination utils.Pagination, filter ProductFilter) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.Model(&models.Product{}).Where("is_active = ?", true)

	// Category filter
	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", *filter.CategoryID)
	}

	// Search filter
	if filter.Search != "" {
		query = query.Where("name LIKE ? OR brand LIKE ? OR description LIKE ?",
			"%"+filter.Search+"%", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	// Price range filter
	if filter.MinPrice != nil {
		query = query.Where("price >= ?", *filter.MinPrice)
	}
	if filter.MaxPrice != nil {
		query = query.Where("price <= ?", *filter.MaxPrice)
	}

	// Brand filter
	if filter.Brand != "" {
		query = query.Where("brand = ?", filter.Brand)
	}

	// Count total
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Sorting
	orderClause := "created_at DESC"
	switch filter.SortBy {
	case "price_asc":
		orderClause = "price ASC"
	case "price_desc":
		orderClause = "price DESC"
	case "popular":
		orderClause = "review_count DESC, rating DESC"
	case "rating":
		orderClause = "rating DESC"
	case "newest":
		orderClause = "created_at DESC"
	}

	err = query.Preload("Category").Preload("Tags").Preload("Ingredients").
		Offset(pagination.GetOffset()).
		Limit(pagination.Limit).
		Order(orderClause).
		Find(&products).Error

	return products, total, err
}

func (r *productRepository) GetAllBrands() ([]string, error) {
	var brands []string
	err := r.db.Model(&models.Product{}).
		Where("is_active = ? AND brand IS NOT NULL AND brand != ''", true).
		Distinct().
		Pluck("brand", &brands).Error
	return brands, err
}

func (r *productRepository) CreateImages(productID uint, images []models.ProductImage) error {
	for i := range images {
		images[i].ProductID = productID
		if err := r.db.Create(&images[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *productRepository) DeleteImages(productID uint) error {
	return r.db.Where("product_id = ?", productID).Delete(&models.ProductImage{}).Error
}

func (r *productRepository) GetImages(productID uint) ([]models.ProductImage, error) {
	var images []models.ProductImage
	err := r.db.Where("product_id = ?", productID).Order("sort_order ASC").Find(&images).Error
	return images, err
}
