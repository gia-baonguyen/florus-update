package handler

import (
	"errors"
	"strconv"

	"github.com/florus/backend/internal/middleware"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/service"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService service.AdminService
}

func NewAdminHandler(adminService service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// GetAllOrders godoc
// @Summary Get all orders (Admin)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status"
// @Success 200 {array} dto.AdminOrderResponse
// @Router /api/admin/orders [get]
func (h *AdminHandler) GetAllOrders(c *gin.Context) {
	pagination := utils.GetPagination(c)
	status := c.Query("status")

	orders, meta, err := h.adminService.GetAllOrders(pagination, status)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMeta(c, 200, "Orders retrieved successfully", orders, meta)
}

// GetOrderDetail godoc
// @Summary Get order detail (Admin)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} dto.AdminOrderResponse
// @Router /api/admin/orders/{id} [get]
func (h *AdminHandler) GetOrderDetail(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid order ID")
		return
	}

	order, err := h.adminService.GetOrderDetail(uint(orderID))
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

// UpdateOrderStatus godoc
// @Summary Update order status (Admin)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body UpdateOrderStatusRequest true "New status"
// @Success 200 {object} dto.AdminOrderResponse
// @Router /api/admin/orders/{id}/status [put]
func (h *AdminHandler) UpdateOrderStatus(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid order ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// Validate status value
	validStatuses := map[string]models.OrderStatus{
		"pending":    models.OrderStatusPending,
		"confirmed":  models.OrderStatusConfirmed,
		"processing": models.OrderStatusProcessing,
		"shipping":   models.OrderStatusShipping,
		"delivered":  models.OrderStatusDelivered,
		"cancelled":  models.OrderStatusCancelled,
	}

	status, ok := validStatuses[req.Status]
	if !ok {
		utils.BadRequest(c, "Invalid status value")
		return
	}

	order, err := h.adminService.UpdateOrderStatus(uint(orderID), status)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			utils.NotFound(c, "Order not found")
			return
		}
		if errors.Is(err, service.ErrInvalidOrderStatus) {
			utils.BadRequest(c, "Invalid status transition")
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Order status updated successfully", order)
}

// GetAllUsers godoc
// @Summary Get all users (Admin)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {array} dto.AdminUserResponse
// @Router /api/admin/users [get]
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	pagination := utils.GetPagination(c)

	users, meta, err := h.adminService.GetAllUsers(pagination)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMeta(c, 200, "Users retrieved successfully", users, meta)
}

// UpdateUserRole godoc
// @Summary Update user role (Admin)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body UpdateUserRoleRequest true "New role"
// @Success 200 {object} dto.AdminUserResponse
// @Router /api/admin/users/{id}/role [put]
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	adminID := middleware.GetUserID(c)

	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid user ID")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	// Validate role value
	var role models.UserRole
	switch req.Role {
	case "admin":
		role = models.RoleAdmin
	case "user":
		role = models.RoleUser
	default:
		utils.BadRequest(c, "Invalid role value. Must be 'admin' or 'user'")
		return
	}

	user, err := h.adminService.UpdateUserRole(adminID, uint(userID), role)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			utils.NotFound(c, "User not found")
			return
		}
		if errors.Is(err, service.ErrCannotDemoteSelf) {
			utils.BadRequest(c, "Cannot demote yourself")
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "User role updated successfully", user)
}

// GetDashboardStats godoc
// @Summary Get dashboard statistics (Admin)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.DashboardStats
// @Router /api/admin/stats [get]
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.adminService.GetDashboardStats()
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Dashboard stats retrieved successfully", stats)
}
