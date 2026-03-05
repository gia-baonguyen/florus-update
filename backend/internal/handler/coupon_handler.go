package handler

import (
	"net/http"
	"strconv"

	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/service"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type CouponHandler struct {
	couponService service.CouponService
}

func NewCouponHandler(couponService service.CouponService) *CouponHandler {
	return &CouponHandler{
		couponService: couponService,
	}
}

type CreateCouponRequest struct {
	Code              string  `json:"code" binding:"required"`
	DiscountType      string  `json:"discount_type" binding:"required,oneof=percent fixed"`
	DiscountValue     float64 `json:"discount_value" binding:"required,gt=0"`
	MinOrderAmount    float64 `json:"min_order_amount"`
	MaxDiscountAmount float64 `json:"max_discount_amount"`
	UsageLimit        int     `json:"usage_limit"`
	StartDate         string  `json:"start_date" binding:"required"`
	EndDate           string  `json:"end_date" binding:"required"`
	Description       string  `json:"description"`
	IsActive          bool    `json:"is_active"`
}

type ValidateCouponRequest struct {
	Code       string  `json:"code" binding:"required"`
	OrderTotal float64 `json:"order_total" binding:"required,gt=0"`
}

// GetAvailableCouponsRequest is used for query params when listing coupons for cart
type GetAvailableCouponsRequest struct {
	OrderTotal float64 `form:"order_total"`
}

func (h *CouponHandler) Create(c *gin.Context) {
	var req CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	startDate, err := utils.ParseDate(req.StartDate)
	if err != nil {
		utils.BadRequest(c, "Invalid start date format")
		return
	}

	endDate, err := utils.ParseDate(req.EndDate)
	if err != nil {
		utils.BadRequest(c, "Invalid end date format")
		return
	}

	coupon := &models.Coupon{
		Code:              req.Code,
		DiscountType:      models.DiscountType(req.DiscountType),
		DiscountValue:     req.DiscountValue,
		MinOrderAmount:    req.MinOrderAmount,
		MaxDiscountAmount: req.MaxDiscountAmount,
		UsageLimit:        req.UsageLimit,
		StartDate:         startDate,
		EndDate:           endDate,
		Description:       req.Description,
		IsActive:          req.IsActive,
	}

	if err := h.couponService.Create(coupon); err != nil {
		if err == service.ErrCouponCodeExists {
			utils.ErrorResponse(c, http.StatusConflict, "Conflict", "Coupon code already exists")
			return
		}
		utils.InternalServerError(c, "Failed to create coupon")
		return
	}

	utils.Created(c, "Coupon created successfully", coupon)
}

func (h *CouponHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	coupons, total, err := h.couponService.GetAll(page, limit)
	if err != nil {
		utils.InternalServerError(c, "Failed to get coupons")
		return
	}

	utils.OK(c, "Coupons retrieved successfully", gin.H{
		"coupons": coupons,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetAvailable returns a public list of currently valid coupons.
// Optional query param: order_total – when provided, only coupons whose
// MinOrderAmount <= order_total are returned.
func (h *CouponHandler) GetAvailable(c *gin.Context) {
	var req GetAvailableCouponsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	coupons, err := h.couponService.GetAvailable(req.OrderTotal)
	if err != nil {
		utils.InternalServerError(c, "Failed to get available coupons")
		return
	}

	utils.OK(c, "Available coupons retrieved successfully", gin.H{
		"coupons": coupons,
	})
}

func (h *CouponHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid coupon ID")
		return
	}

	coupon, err := h.couponService.GetByID(uint(id))
	if err != nil {
		utils.NotFound(c, "Coupon not found")
		return
	}

	utils.OK(c, "Coupon retrieved successfully", coupon)
}

func (h *CouponHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid coupon ID")
		return
	}

	var req CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	startDate, err := utils.ParseDate(req.StartDate)
	if err != nil {
		utils.BadRequest(c, "Invalid start date format")
		return
	}

	endDate, err := utils.ParseDate(req.EndDate)
	if err != nil {
		utils.BadRequest(c, "Invalid end date format")
		return
	}

	coupon := &models.Coupon{
		ID:                uint(id),
		Code:              req.Code,
		DiscountType:      models.DiscountType(req.DiscountType),
		DiscountValue:     req.DiscountValue,
		MinOrderAmount:    req.MinOrderAmount,
		MaxDiscountAmount: req.MaxDiscountAmount,
		UsageLimit:        req.UsageLimit,
		StartDate:         startDate,
		EndDate:           endDate,
		Description:       req.Description,
		IsActive:          req.IsActive,
	}

	if err := h.couponService.Update(coupon); err != nil {
		if err == service.ErrCouponNotFound {
			utils.NotFound(c, "Coupon not found")
			return
		}
		if err == service.ErrCouponCodeExists {
			utils.ErrorResponse(c, http.StatusConflict, "Conflict", "Coupon code already exists")
			return
		}
		utils.InternalServerError(c, "Failed to update coupon")
		return
	}

	utils.OK(c, "Coupon updated successfully", coupon)
}

func (h *CouponHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid coupon ID")
		return
	}

	if err := h.couponService.Delete(uint(id)); err != nil {
		if err == service.ErrCouponNotFound {
			utils.NotFound(c, "Coupon not found")
			return
		}
		utils.InternalServerError(c, "Failed to delete coupon")
		return
	}

	utils.OK(c, "Coupon deleted successfully", nil)
}

func (h *CouponHandler) ValidateCoupon(c *gin.Context) {
	var req ValidateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	coupon, discount, err := h.couponService.ValidateCoupon(req.Code, req.OrderTotal)
	if err != nil {
		switch err {
		case service.ErrCouponNotFound:
			utils.NotFound(c, "Coupon not found")
		case service.ErrCouponInvalid:
			utils.BadRequest(c, "Coupon is invalid or expired")
		case service.ErrCouponMinAmount:
			utils.BadRequest(c, "Order total does not meet minimum amount")
		default:
			utils.InternalServerError(c, "Failed to validate coupon")
		}
		return
	}

	utils.OK(c, "Coupon is valid", gin.H{
		"coupon":      coupon,
		"discount":    discount,
		"final_total": req.OrderTotal - discount,
	})
}
