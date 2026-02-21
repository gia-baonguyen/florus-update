package service

import (
	"errors"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
	"github.com/florus/backend/pkg/utils"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrProductSlugExists = errors.New("product slug already exists")
)

type ProductService interface {
	Create(req dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetByID(id uint) (*dto.ProductResponse, error)
	GetBySlug(slug string) (*dto.ProductResponse, error)
	GetAll(pagination utils.Pagination, categoryID *uint, search string) ([]dto.ProductResponse, *utils.Meta, error)
	GetAllWithFilter(pagination utils.Pagination, categoryID *uint, search string, minPrice, maxPrice *float64, brand, sortBy string) ([]dto.ProductResponse, *utils.Meta, error)
	GetByCategory(categoryID uint, pagination utils.Pagination) ([]dto.ProductResponse, *utils.Meta, error)
	Update(id uint, req dto.UpdateProductRequest) (*dto.ProductResponse, error)
	Delete(id uint) error
	GetAllBrands() ([]string, error)
}

type productService struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
}

func NewProductService(productRepo repository.ProductRepository, categoryRepo repository.CategoryRepository) ProductService {
	return &productService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *productService) Create(req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	if s.productRepo.ExistsBySlug(req.Slug) {
		return nil, ErrProductSlugExists
	}

	// Verify category exists
	_, err := s.categoryRepo.FindByID(req.CategoryID)
	if err != nil {
		return nil, ErrCategoryNotFound
	}

	product := &models.Product{
		Name:            req.Name,
		Slug:            req.Slug,
		Brand:           req.Brand,
		Price:           req.Price,
		OriginalPrice:   req.OriginalPrice,
		CategoryID:      req.CategoryID,
		Description:     req.Description,
		ImageURL:        req.ImageURL,
		Stock:           req.Stock,
		Rating:          req.Rating,
		ReviewCount:     req.ReviewCount,
		AIScore:         req.AIScore,
		AIRecommendType: req.AIRecommendType,
		IsActive:        true,
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, err
	}

	// Create tags
	if len(req.Tags) > 0 {
		if err := s.productRepo.CreateTags(product.ID, req.Tags); err != nil {
			return nil, err
		}
	}

	// Create ingredients
	if len(req.Ingredients) > 0 {
		if err := s.productRepo.CreateIngredients(product.ID, req.Ingredients); err != nil {
			return nil, err
		}
	}

	// Reload product with relations
	product, _ = s.productRepo.FindByID(product.ID)
	resp := dto.ToProductResponse(product)
	return &resp, nil
}

func (s *productService) GetByID(id uint) (*dto.ProductResponse, error) {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, ErrProductNotFound
	}

	resp := dto.ToProductResponse(product)
	return &resp, nil
}

func (s *productService) GetBySlug(slug string) (*dto.ProductResponse, error) {
	product, err := s.productRepo.FindBySlug(slug)
	if err != nil {
		return nil, ErrProductNotFound
	}

	resp := dto.ToProductResponse(product)
	return &resp, nil
}

func (s *productService) GetAll(pagination utils.Pagination, categoryID *uint, search string) ([]dto.ProductResponse, *utils.Meta, error) {
	products, total, err := s.productRepo.FindAll(pagination, categoryID, search)
	if err != nil {
		return nil, nil, err
	}

	var responses []dto.ProductResponse
	for _, p := range products {
		responses = append(responses, dto.ToProductResponse(&p))
	}

	meta := utils.NewMeta(pagination.Page, pagination.Limit, total)
	return responses, meta, nil
}

func (s *productService) GetByCategory(categoryID uint, pagination utils.Pagination) ([]dto.ProductResponse, *utils.Meta, error) {
	products, total, err := s.productRepo.FindByCategory(categoryID, pagination)
	if err != nil {
		return nil, nil, err
	}

	var responses []dto.ProductResponse
	for _, p := range products {
		responses = append(responses, dto.ToProductResponse(&p))
	}

	meta := utils.NewMeta(pagination.Page, pagination.Limit, total)
	return responses, meta, nil
}

func (s *productService) Update(id uint, req dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, ErrProductNotFound
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Slug != "" && req.Slug != product.Slug {
		if s.productRepo.ExistsBySlug(req.Slug) {
			return nil, ErrProductSlugExists
		}
		product.Slug = req.Slug
	}
	if req.Brand != "" {
		product.Brand = req.Brand
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.OriginalPrice != nil {
		product.OriginalPrice = req.OriginalPrice
	}
	if req.CategoryID != nil {
		product.CategoryID = *req.CategoryID
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.ImageURL != "" {
		product.ImageURL = req.ImageURL
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.Rating != nil {
		product.Rating = *req.Rating
	}
	if req.ReviewCount != nil {
		product.ReviewCount = *req.ReviewCount
	}
	if req.AIScore != nil {
		product.AIScore = req.AIScore
	}
	if req.AIRecommendType != "" {
		product.AIRecommendType = req.AIRecommendType
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := s.productRepo.Update(product); err != nil {
		return nil, err
	}

	// Update tags
	if req.Tags != nil {
		s.productRepo.DeleteTags(product.ID)
		if len(req.Tags) > 0 {
			s.productRepo.CreateTags(product.ID, req.Tags)
		}
	}

	// Update ingredients
	if req.Ingredients != nil {
		s.productRepo.DeleteIngredients(product.ID)
		if len(req.Ingredients) > 0 {
			s.productRepo.CreateIngredients(product.ID, req.Ingredients)
		}
	}

	// Reload product
	product, _ = s.productRepo.FindByID(product.ID)
	resp := dto.ToProductResponse(product)
	return &resp, nil
}

func (s *productService) Delete(id uint) error {
	_, err := s.productRepo.FindByID(id)
	if err != nil {
		return ErrProductNotFound
	}

	return s.productRepo.Delete(id)
}

func (s *productService) GetAllWithFilter(pagination utils.Pagination, categoryID *uint, search string, minPrice, maxPrice *float64, brand, sortBy string) ([]dto.ProductResponse, *utils.Meta, error) {
	filter := repository.ProductFilter{
		CategoryID: categoryID,
		Search:     search,
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
		Brand:      brand,
		SortBy:     sortBy,
	}

	products, total, err := s.productRepo.FindAllWithFilter(pagination, filter)
	if err != nil {
		return nil, nil, err
	}

	var responses []dto.ProductResponse
	for _, p := range products {
		responses = append(responses, dto.ToProductResponse(&p))
	}

	meta := utils.NewMeta(pagination.Page, pagination.Limit, total)
	return responses, meta, nil
}

func (s *productService) GetAllBrands() ([]string, error) {
	return s.productRepo.GetAllBrands()
}
