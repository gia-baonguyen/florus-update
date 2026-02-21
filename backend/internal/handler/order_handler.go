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

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// GetOrders godoc
// @Summary Get current user's orders
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {array} dto.OrderResponse
// @Router /api/orders [get]
func (h *OrderHandler) GetOrders(c *gin.Context) {
	userID := middleware.GetUserID(c)
	pagination := utils.GetPagination(c)

	orders, meta, err := h.orderService.GetUserOrders(userID, pagination)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMeta(c, 200, "Orders retrieved successfully", orders, meta)
}

// GetOrderByID godoc
// @Summary Get order by ID
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} dto.OrderResponse
// @Router /api/orders/{id} [get]
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	userID := middleware.GetUserID(c)

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid order ID")
		return
	}

	order, err := h.orderService.GetOrderByID(userID, uint(orderID))
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			utils.NotFound(c, "Order not found")
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Order retrieved successfully", order)
}

// CreateOrder godoc
// @Summary Create a new order from cart (Transaction)
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateOrderRequest true "Order request"
// @Success 201 {object} dto.OrderResponse
// @Router /api/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	order, err := h.orderService.CreateOrder(userID, req)
	if err != nil {
		if errors.Is(err, service.ErrEmptyCart) {
			utils.BadRequest(c, "Cart is empty")
			return
		}
		if errors.Is(err, service.ErrInsufficientStock) {
			utils.BadRequest(c, "Insufficient stock for one or more items")
			return
		}
		if errors.Is(err, service.ErrProductNotFound) {
			utils.BadRequest(c, "One or more products not found")
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Created(c, "Order created successfully", order)
}

// CancelOrder godoc
// @Summary Cancel an order
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} utils.Response
// @Router /api/orders/{id}/cancel [put]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid order ID")
		return
	}

	if err := h.orderService.CancelOrder(userID, uint(orderID)); err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			utils.NotFound(c, "Order not found")
			return
		}
		if errors.Is(err, service.ErrCannotCancelOrder) {
			utils.BadRequest(c, "Order cannot be cancelled")
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Order cancelled successfully", nil)
}
