package handler

import (
	"errors"
	"strconv"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/service"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// GetAll godoc
// @Summary Get all products with pagination and advanced filtering
// @Tags Products
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param category_id query int false "Filter by category ID"
// @Param search query string false "Search query"
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Param brand query string false "Filter by brand"
// @Param sort_by query string false "Sort by: newest, price_asc, price_desc, popular, rating"
// @Success 200 {array} dto.ProductResponse
// @Router /api/products [get]
func (h *ProductHandler) GetAll(c *gin.Context) {
	pagination := utils.GetPagination(c)

	var categoryID *uint
	if catIDStr := c.Query("category_id"); catIDStr != "" {
		if catID, err := strconv.ParseUint(catIDStr, 10, 32); err == nil {
			catIDVal := uint(catID)
			categoryID = &catIDVal
		}
	}

	var minPrice, maxPrice *float64
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if p, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			minPrice = &p
		}
	}
	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if p, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			maxPrice = &p
		}
	}

	search := c.Query("search")
	brand := c.Query("brand")
	sortBy := c.DefaultQuery("sort_by", "newest")

	products, meta, err := h.productService.GetAllWithFilter(pagination, categoryID, search, minPrice, maxPrice, brand, sortBy)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMeta(c, 200, "Products retrieved successfully", products, meta)
}

// GetBrands godoc
// @Summary Get all product brands
// @Tags Products
// @Produce json
// @Success 200 {array} string
// @Router /api/products/brands [get]
func (h *ProductHandler) GetBrands(c *gin.Context) {
	brands, err := h.productService.GetAllBrands()
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Brands retrieved successfully", brands)
}

// GetByID godoc
// @Summary Get product by ID
// @Tags Products
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} dto.ProductResponse
// @Router /api/products/{id} [get]
func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid product ID")
		return
	}

	product, err := h.productService.GetByID(uint(id))
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			utils.NotFound(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Product retrieved successfully", product)
}

// GetBySlug godoc
// @Summary Get product by slug
// @Tags Products
// @Produce json
// @Param slug path string true "Product slug"
// @Success 200 {object} dto.ProductResponse
// @Router /api/products/slug/{slug} [get]
func (h *ProductHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	product, err := h.productService.GetBySlug(slug)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			utils.NotFound(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Product retrieved successfully", product)
}

// GetByCategory godoc
// @Summary Get products by category ID
// @Tags Products
// @Produce json
// @Param categoryId path int true "Category ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {array} dto.ProductResponse
// @Router /api/products/category/{categoryId} [get]
func (h *ProductHandler) GetByCategory(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("categoryId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid category ID")
		return
	}

	pagination := utils.GetPagination(c)

	products, meta, err := h.productService.GetByCategory(uint(categoryID), pagination)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMeta(c, 200, "Products retrieved successfully", products, meta)
}

// Create godoc
// @Summary Create a new product (Admin only)
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateProductRequest true "Product data"
// @Success 201 {object} dto.ProductResponse
// @Router /api/products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	product, err := h.productService.Create(req)
	if err != nil {
		if errors.Is(err, service.ErrProductSlugExists) {
			utils.BadRequest(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrCategoryNotFound) {
			utils.BadRequest(c, "Category not found")
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Created(c, "Product created successfully", product)
}

// Update godoc
// @Summary Update a product (Admin only)
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param request body dto.UpdateProductRequest true "Product data"
// @Success 200 {object} dto.ProductResponse
// @Router /api/products/{id} [put]
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid product ID")
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	product, err := h.productService.Update(uint(id), req)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			utils.NotFound(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrProductSlugExists) {
			utils.BadRequest(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Product updated successfully", product)
}

// Delete godoc
// @Summary Delete a product (Admin only)
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 200 {object} utils.Response
// @Router /api/products/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid product ID")
		return
	}

	if err := h.productService.Delete(uint(id)); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			utils.NotFound(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Product deleted successfully", nil)
}
