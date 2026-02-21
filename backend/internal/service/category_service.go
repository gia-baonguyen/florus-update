package service

import (
	"errors"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrCategorySlugExists = errors.New("category slug already exists")
)

type CategoryService interface {
	Create(req dto.CreateCategoryRequest) (*dto.CategoryResponse, error)
	GetByID(id uint) (*dto.CategoryResponse, error)
	GetBySlug(slug string) (*dto.CategoryResponse, error)
	GetAll(activeOnly bool) ([]dto.CategoryResponse, error)
	Update(id uint, req dto.UpdateCategoryRequest) (*dto.CategoryResponse, error)
	Delete(id uint) error
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *categoryService) Create(req dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	if s.categoryRepo.ExistsBySlug(req.Slug) {
		return nil, ErrCategorySlugExists
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	category := &models.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		ParentID:    req.ParentID,
		IsActive:    isActive,
	}

	if err := s.categoryRepo.Create(category); err != nil {
		return nil, err
	}

	resp := dto.ToCategoryResponse(category)
	return &resp, nil
}

func (s *categoryService) GetByID(id uint) (*dto.CategoryResponse, error) {
	category, err := s.categoryRepo.FindByID(id)
	if err != nil {
		return nil, ErrCategoryNotFound
	}

	resp := dto.ToCategoryResponse(category)
	return &resp, nil
}

func (s *categoryService) GetBySlug(slug string) (*dto.CategoryResponse, error) {
	category, err := s.categoryRepo.FindBySlug(slug)
	if err != nil {
		return nil, ErrCategoryNotFound
	}

	resp := dto.ToCategoryResponse(category)
	return &resp, nil
}

func (s *categoryService) GetAll(activeOnly bool) ([]dto.CategoryResponse, error) {
	categories, err := s.categoryRepo.FindAll(activeOnly)
	if err != nil {
		return nil, err
	}

	var responses []dto.CategoryResponse
	for _, cat := range categories {
		responses = append(responses, dto.ToCategoryResponse(&cat))
	}

	return responses, nil
}

func (s *categoryService) Update(id uint, req dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	category, err := s.categoryRepo.FindByID(id)
	if err != nil {
		return nil, ErrCategoryNotFound
	}

	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Slug != "" && req.Slug != category.Slug {
		if s.categoryRepo.ExistsBySlug(req.Slug) {
			return nil, ErrCategorySlugExists
		}
		category.Slug = req.Slug
	}
	if req.Description != "" {
		category.Description = req.Description
	}
	if req.ImageURL != "" {
		category.ImageURL = req.ImageURL
	}
	if req.ParentID != nil {
		category.ParentID = req.ParentID
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := s.categoryRepo.Update(category); err != nil {
		return nil, err
	}

	resp := dto.ToCategoryResponse(category)
	return &resp, nil
}

func (s *categoryService) Delete(id uint) error {
	_, err := s.categoryRepo.FindByID(id)
	if err != nil {
		return ErrCategoryNotFound
	}

	return s.categoryRepo.Delete(id)
}
