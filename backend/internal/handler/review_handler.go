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

type ReviewHandler struct {
	reviewService service.ReviewService
}

func NewReviewHandler(reviewService service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: reviewService}
}

// GetProductReviews godoc
// @Summary Get product reviews
// @Tags Reviews
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} dto.ReviewListResponse
// @Router /api/products/{id}/reviews [get]
func (h *ReviewHandler) GetProductReviews(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid product ID")
		return
	}

	resp, err := h.reviewService.GetProductReviews(uint(productID))
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Reviews retrieved successfully", resp)
}

// CreateReview godoc
// @Summary Create a product review
// @Tags Reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param request body dto.CreateReviewRequest true "Review data"
// @Success 201 {object} dto.ReviewResponse
// @Router /api/products/{id}/reviews [post]
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID := middleware.GetUserID(c)

	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid product ID")
		return
	}

	var req dto.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	resp, err := h.reviewService.CreateReview(userID, uint(productID), req)
	if err != nil {
		if errors.Is(err, service.ErrAlreadyReviewed) {
			utils.BadRequest(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrInvalidRating) {
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

	utils.Created(c, "Review created successfully", resp)
}

// UpdateReview godoc
// @Summary Update a review
// @Tags Reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Review ID"
// @Param request body dto.UpdateReviewRequest true "Review data"
// @Success 200 {object} dto.ReviewResponse
// @Router /api/reviews/{id} [put]
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	userID := middleware.GetUserID(c)

	reviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid review ID")
		return
	}

	var req dto.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	resp, err := h.reviewService.UpdateReview(userID, uint(reviewID), req)
	if err != nil {
		if errors.Is(err, service.ErrReviewNotFound) {
			utils.NotFound(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrNotReviewOwner) {
			utils.Forbidden(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrInvalidRating) {
			utils.BadRequest(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Review updated successfully", resp)
}

// DeleteReview godoc
// @Summary Delete a review
// @Tags Reviews
// @Produce json
// @Security BearerAuth
// @Param id path int true "Review ID"
// @Success 200 {object} map[string]string
// @Router /api/reviews/{id} [delete]
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	userID := middleware.GetUserID(c)

	reviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid review ID")
		return
	}

	err = h.reviewService.DeleteReview(userID, uint(reviewID))
	if err != nil {
		if errors.Is(err, service.ErrReviewNotFound) {
			utils.NotFound(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrNotReviewOwner) {
			utils.Forbidden(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Review deleted successfully", nil)
}

// GetUserReview godoc
// @Summary Get user's review for a product
// @Tags Reviews
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 200 {object} dto.ReviewResponse
// @Router /api/products/{id}/my-review [get]
func (h *ReviewHandler) GetUserReview(c *gin.Context) {
	userID := middleware.GetUserID(c)

	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid product ID")
		return
	}

	resp, err := h.reviewService.GetUserReview(userID, uint(productID))
	if err != nil {
		if errors.Is(err, service.ErrReviewNotFound) {
			utils.NotFound(c, "You haven't reviewed this product yet")
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Review retrieved successfully", resp)
}
