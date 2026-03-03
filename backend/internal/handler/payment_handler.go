package handler

import (
	"errors"
	"net/http"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/service"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService service.PaymentService
}

func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// CreatePayment godoc
// @Summary Create a payment for an order
// @Tags Payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreatePaymentRequest true "Create payment request"
// @Success 200 {object} dto.CreatePaymentResponse
// @Router /api/payments/create [post]
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req dto.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// Validate payment method
	method := models.PaymentMethod(req.PaymentMethod)
	if method != models.PaymentMethodZaloPay && method != models.PaymentMethodMoMo && method != models.PaymentMethodVNPay {
		utils.BadRequest(c, "Invalid payment method. Supported: zalopay, momo, vnpay")
		return
	}

	resp, err := h.paymentService.CreatePayment(req.OrderCode, method)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			utils.NotFound(c, "Order not found")
			return
		}
		if errors.Is(err, service.ErrPaymentAlreadyPaid) {
			utils.BadRequest(c, "Order has already been paid")
			return
		}
		if errors.Is(err, service.ErrPaymentFailed) {
			utils.InternalServerError(c, "Failed to create payment")
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Payment created successfully", resp)
}

// GetPaymentStatus godoc
// @Summary Get payment status for an order
// @Tags Payments
// @Produce json
// @Security BearerAuth
// @Param orderCode path string true "Order Code"
// @Success 200 {object} dto.PaymentResponse
// @Router /api/payments/status/{orderCode} [get]
func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	orderCode := c.Param("orderCode")

	payment, err := h.paymentService.GetPaymentByOrderCode(orderCode)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			utils.NotFound(c, "Order not found")
			return
		}
		if errors.Is(err, service.ErrPaymentNotFound) {
			utils.NotFound(c, "Payment not found")
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Payment status retrieved", dto.ToPaymentResponse(payment))
}

// ZaloPayCallback godoc
// @Summary Handle ZaloPay payment callback
// @Tags Payments
// @Accept json
// @Produce json
// @Router /api/payments/callback/zalopay [post]
func (h *PaymentHandler) ZaloPayCallback(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusOK, gin.H{"return_code": -1, "return_message": "invalid request"})
		return
	}

	if err := h.paymentService.HandleZaloPayCallback(data); err != nil {
		c.JSON(http.StatusOK, gin.H{"return_code": -1, "return_message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"return_code": 1, "return_message": "success"})
}

// MoMoCallback godoc
// @Summary Handle MoMo payment callback
// @Tags Payments
// @Accept json
// @Produce json
// @Router /api/payments/callback/momo [post]
func (h *PaymentHandler) MoMoCallback(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusOK, gin.H{"resultCode": -1, "message": "invalid request"})
		return
	}

	if err := h.paymentService.HandleMoMoCallback(data); err != nil {
		c.JSON(http.StatusOK, gin.H{"resultCode": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"resultCode": 0, "message": "success"})
}

// VNPayCallback godoc
// @Summary Handle VNPay payment callback (return URL)
// @Tags Payments
// @Produce json
// @Router /api/payments/callback/vnpay [get]
func (h *PaymentHandler) VNPayCallback(c *gin.Context) {
	params := c.Request.URL.Query()

	if err := h.paymentService.HandleVNPayCallback(params); err != nil {
		c.JSON(http.StatusOK, gin.H{"RspCode": "97", "Message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"RspCode": "00", "Message": "success"})
}
