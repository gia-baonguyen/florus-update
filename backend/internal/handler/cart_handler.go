package handler

import (
	"errors"
	"strconv"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/middleware"
	"github.com/florus/backend/internal/service"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	cartService service.CartService
}

func NewCartHandler(cartService service.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

// GetCart godoc
// @Summary Get current user's cart
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.CartResponse
// @Router /api/cart [get]
func (h *CartHandler) GetCart(c *gin.Context) {
	userID := middleware.GetUserID(c)

	cart, err := h.cartService.GetCart(userID)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Cart retrieved successfully", cart)
}

// AddToCart godoc
// @Summary Add product to cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.AddToCartRequest true "Add to cart request"
// @Success 200 {object} dto.CartResponse
// @Router /api/cart [post]
func (h *CartHandler) AddToCart(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	cart, err := h.cartService.AddToCart(userID, req)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			utils.NotFound(c, "Product not found")
			return
		}
		if errors.Is(err, service.ErrInsufficientStock) {
			utils.BadRequest(c, "Insufficient stock")
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Product added to cart", cart)
}

// UpdateQuantity godoc
// @Summary Update cart item quantity
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param itemId path int true "Cart item ID"
// @Param request body dto.UpdateCartItemRequest true "Update quantity request"
// @Success 200 {object} dto.CartResponse
// @Router /api/cart/{itemId} [put]
func (h *CartHandler) UpdateQuantity(c *gin.Context) {
	userID := middleware.GetUserID(c)

	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid item ID")
		return
	}

	var req dto.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	cart, err := h.cartService.UpdateQuantity(userID, uint(itemID), req)
	if err != nil {
		if errors.Is(err, service.ErrCartItemNotFound) {
			utils.NotFound(c, "Cart item not found")
			return
		}
		if errors.Is(err, service.ErrInsufficientStock) {
			utils.BadRequest(c, "Insufficient stock")
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Cart updated successfully", cart)
}

// RemoveItem godoc
// @Summary Remove item from cart
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Param itemId path int true "Cart item ID"
// @Success 200 {object} utils.Response
// @Router /api/cart/{itemId} [delete]
func (h *CartHandler) RemoveItem(c *gin.Context) {
	userID := middleware.GetUserID(c)

	itemID, err := strconv.ParseUint(c.Param("itemId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid item ID")
		return
	}

	if err := h.cartService.RemoveItem(userID, uint(itemID)); err != nil {
		if errors.Is(err, service.ErrCartItemNotFound) {
			utils.NotFound(c, "Cart item not found")
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Item removed from cart", nil)
}

// ClearCart godoc
// @Summary Clear all items from cart
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/cart [delete]
func (h *CartHandler) ClearCart(c *gin.Context) {
	userID := middleware.GetUserID(c)

	if err := h.cartService.ClearCart(userID); err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Cart cleared successfully", nil)
}
