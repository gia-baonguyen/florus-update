package repository

import (
	"github.com/florus/backend/internal/models"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(category *models.Category) error
	FindByID(id uint) (*models.Category, error)
	FindBySlug(slug string) (*models.Category, error)
	FindAll(activeOnly bool) ([]models.Category, error)
	Update(category *models.Category) error
	Delete(id uint) error
	ExistsBySlug(slug string) bool
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(category *models.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepository) FindByID(id uint) (*models.Category, error) {
	var category models.Category
	err := r.db.Preload("Children").First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) FindBySlug(slug string) (*models.Category, error) {
	var category models.Category
	err := r.db.Preload("Children").Where("slug = ?", slug).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) FindAll(activeOnly bool) ([]models.Category, error) {
	var categories []models.Category
	query := r.db.Preload("Children")
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	err := query.Where("parent_id IS NULL").Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) Update(category *models.Category) error {
	return r.db.Save(category).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&models.Category{}, id).Error
}

func (r *categoryRepository) ExistsBySlug(slug string) bool {
	var count int64
	r.db.Model(&models.Category{}).Where("slug = ?", slug).Count(&count)
	return count > 0
}
