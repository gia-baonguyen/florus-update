package handler

import (
	"errors"
	"strconv"

	"github.com/florus/backend/internal/middleware"
	"github.com/florus/backend/internal/service"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type WishlistHandler struct {
	wishlistService service.WishlistService
}

func NewWishlistHandler(wishlistService service.WishlistService) *WishlistHandler {
	return &WishlistHandler{wishlistService: wishlistService}
}

// GetWishlist godoc
// @Summary Get user's wishlist
// @Tags Wishlist
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.WishlistResponse
// @Router /api/wishlist [get]
func (h *WishlistHandler) GetWishlist(c *gin.Context) {
	userID := middleware.GetUserID(c)

	resp, err := h.wishlistService.GetWishlist(userID)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Wishlist retrieved successfully", resp)
}

// AddToWishlist godoc
// @Summary Add product to wishlist
// @Tags Wishlist
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]uint true "Product ID"
// @Success 201 {object} dto.WishlistItemResponse
// @Router /api/wishlist [post]
func (h *WishlistHandler) AddToWishlist(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		ProductID uint `json:"product_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	resp, err := h.wishlistService.AddToWishlist(userID, req.ProductID)
	if err != nil {
		if errors.Is(err, service.ErrProductAlreadyInWishlist) {
			utils.BadRequest(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrProductNotFound) {
			utils.NotFound(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Created(c, "Product added to wishlist", resp)
}

// RemoveFromWishlist godoc
// @Summary Remove product from wishlist
// @Tags Wishlist
// @Produce json
// @Security BearerAuth
// @Param productId path int true "Product ID"
// @Success 200 {object} map[string]string
// @Router /api/wishlist/{productId} [delete]
func (h *WishlistHandler) RemoveFromWishlist(c *gin.Context) {
	userID := middleware.GetUserID(c)

	productID, err := strconv.ParseUint(c.Param("productId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid product ID")
		return
	}

	err = h.wishlistService.RemoveFromWishlist(userID, uint(productID))
	if err != nil {
		if errors.Is(err, service.ErrWishlistItemNotFound) {
			utils.NotFound(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Product removed from wishlist", nil)
}

// CheckWishlist godoc
// @Summary Check if product is in wishlist
// @Tags Wishlist
// @Produce json
// @Security BearerAuth
// @Param productId path int true "Product ID"
// @Success 200 {object} map[string]bool
// @Router /api/wishlist/check/{productId} [get]
func (h *WishlistHandler) CheckWishlist(c *gin.Context) {
	userID := middleware.GetUserID(c)

	productID, err := strconv.ParseUint(c.Param("productId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid product ID")
		return
	}

	inWishlist, err := h.wishlistService.IsInWishlist(userID, uint(productID))
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Wishlist status checked", gin.H{"in_wishlist": inWishlist})
}

// GetWishlistIDs godoc
// @Summary Get all product IDs in wishlist
// @Tags Wishlist
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string][]uint
// @Router /api/wishlist/ids [get]
func (h *WishlistHandler) GetWishlistIDs(c *gin.Context) {
	userID := middleware.GetUserID(c)

	productIDs, err := h.wishlistService.GetWishlistProductIDs(userID)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Wishlist IDs retrieved", gin.H{"product_ids": productIDs})
}
