package handler

import (
	"errors"
	"strconv"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/service"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryService service.CategoryService
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// GetAll godoc
// @Summary Get all categories
// @Tags Categories
// @Produce json
// @Success 200 {array} dto.CategoryResponse
// @Router /api/categories [get]
func (h *CategoryHandler) GetAll(c *gin.Context) {
	activeOnly := c.DefaultQuery("active", "true") == "true"

	categories, err := h.categoryService.GetAll(activeOnly)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Categories retrieved successfully", categories)
}

// GetByID godoc
// @Summary Get category by ID
// @Tags Categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} dto.CategoryResponse
// @Router /api/categories/{id} [get]
func (h *CategoryHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid category ID")
		return
	}

	category, err := h.categoryService.GetByID(uint(id))
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			utils.NotFound(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Category retrieved successfully", category)
}

// Create godoc
// @Summary Create a new category (Admin only)
// @Tags Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCategoryRequest true "Category data"
// @Success 201 {object} dto.CategoryResponse
// @Router /api/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	category, err := h.categoryService.Create(req)
	if err != nil {
		if errors.Is(err, service.ErrCategorySlugExists) {
			utils.BadRequest(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Created(c, "Category created successfully", category)
}

// Update godoc
// @Summary Update a category (Admin only)
// @Tags Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Param request body dto.UpdateCategoryRequest true "Category data"
// @Success 200 {object} dto.CategoryResponse
// @Router /api/categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid category ID")
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	category, err := h.categoryService.Update(uint(id), req)
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			utils.NotFound(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrCategorySlugExists) {
			utils.BadRequest(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Category updated successfully", category)
}

// Delete godoc
// @Summary Delete a category (Admin only)
// @Tags Categories
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Success 200 {object} utils.Response
// @Router /api/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid category ID")
		return
	}

	if err := h.categoryService.Delete(uint(id)); err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			utils.NotFound(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Category deleted successfully", nil)
}
