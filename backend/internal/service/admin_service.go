package service

import (
	"errors"
	"time"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/pkg/utils"
	"gorm.io/gorm"
)

var (
	ErrInvalidOrderStatus = errors.New("invalid order status transition")
	ErrCannotDemoteSelf   = errors.New("cannot demote yourself")
)

type AdminService interface {
	// Orders
	GetAllOrders(pagination utils.Pagination, status string) ([]dto.AdminOrderResponse, *utils.Meta, error)
	GetOrderDetail(orderID uint) (*dto.AdminOrderResponse, error)
	UpdateOrderStatus(orderID uint, status models.OrderStatus) (*dto.AdminOrderResponse, error)

	// Users
	GetAllUsers(pagination utils.Pagination) ([]dto.AdminUserResponse, *utils.Meta, error)
	UpdateUserRole(adminID uint, userID uint, role models.UserRole) (*dto.AdminUserResponse, error)

	// Stats
	GetDashboardStats() (*dto.DashboardStats, error)
}

type adminService struct {
	db *gorm.DB
}

func NewAdminService(db *gorm.DB) AdminService {
	return &adminService{db: db}
}

// GetAllOrders returns all orders with optional status filter
func (s *adminService) GetAllOrders(pagination utils.Pagination, status string) ([]dto.AdminOrderResponse, *utils.Meta, error) {
	var orders []models.Order
	var total int64

	query := s.db.Model(&models.Order{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	err := query.Preload("User").
		Preload("OrderItems").
		Preload("OrderItems.Product").
		Offset(pagination.GetOffset()).
		Limit(pagination.Limit).
		Order("created_at DESC").
		Find(&orders).Error

	if err != nil {
		return nil, nil, err
	}

	var responses []dto.AdminOrderResponse
	for _, order := range orders {
		responses = append(responses, dto.ToAdminOrderResponse(&order))
	}

	meta := utils.NewMeta(pagination.Page, pagination.Limit, total)
	return responses, meta, nil
}

// GetOrderDetail returns order with full details
func (s *adminService) GetOrderDetail(orderID uint) (*dto.AdminOrderResponse, error) {
	var order models.Order
	err := s.db.Preload("User").
		Preload("OrderItems").
		Preload("OrderItems.Product").
		Preload("OrderItems.Product.Category").
		First(&order, orderID).Error

	if err != nil {
		return nil, ErrOrderNotFound
	}

	resp := dto.ToAdminOrderResponse(&order)
	return &resp, nil
}

// UpdateOrderStatus updates order status with validation
func (s *adminService) UpdateOrderStatus(orderID uint, newStatus models.OrderStatus) (*dto.AdminOrderResponse, error) {
	var order models.Order
	if err := s.db.First(&order, orderID).Error; err != nil {
		return nil, ErrOrderNotFound
	}

	// Validate status transition
	if !isValidStatusTransition(order.Status, newStatus) {
		return nil, ErrInvalidOrderStatus
	}

	// Update status
	if err := s.db.Model(&order).Update("status", newStatus).Error; err != nil {
		return nil, err
	}

	// Reload with relations
	return s.GetOrderDetail(orderID)
}

// isValidStatusTransition validates order status transitions
func isValidStatusTransition(current, new models.OrderStatus) bool {
	transitions := map[models.OrderStatus][]models.OrderStatus{
		models.OrderStatusPending:    {models.OrderStatusConfirmed, models.OrderStatusCancelled},
		models.OrderStatusConfirmed:  {models.OrderStatusProcessing, models.OrderStatusCancelled},
		models.OrderStatusProcessing: {models.OrderStatusShipping},
		models.OrderStatusShipping:   {models.OrderStatusDelivered},
	}

	allowed, exists := transitions[current]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == new {
			return true
		}
	}
	return false
}

// GetAllUsers returns all users
func (s *adminService) GetAllUsers(pagination utils.Pagination) ([]dto.AdminUserResponse, *utils.Meta, error) {
	var users []models.User
	var total int64

	if err := s.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, nil, err
	}

	err := s.db.Offset(pagination.GetOffset()).
		Limit(pagination.Limit).
		Order("created_at DESC").
		Find(&users).Error

	if err != nil {
		return nil, nil, err
	}

	var responses []dto.AdminUserResponse
	for _, user := range users {
		// Get order count for each user
		var orderCount int64
		s.db.Model(&models.Order{}).Where("user_id = ?", user.ID).Count(&orderCount)

		responses = append(responses, dto.ToAdminUserResponse(&user, orderCount))
	}

	meta := utils.NewMeta(pagination.Page, pagination.Limit, total)
	return responses, meta, nil
}

// UpdateUserRole updates user role
func (s *adminService) UpdateUserRole(adminID uint, userID uint, role models.UserRole) (*dto.AdminUserResponse, error) {
	// Prevent admin from demoting themselves
	if adminID == userID && role != models.RoleAdmin {
		return nil, ErrCannotDemoteSelf
	}

	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	if err := s.db.Model(&user).Update("role", role).Error; err != nil {
		return nil, err
	}

	user.Role = role

	var orderCount int64
	s.db.Model(&models.Order{}).Where("user_id = ?", user.ID).Count(&orderCount)

	resp := dto.ToAdminUserResponse(&user, orderCount)
	return &resp, nil
}

// GetDashboardStats returns dashboard statistics
func (s *adminService) GetDashboardStats() (*dto.DashboardStats, error) {
	stats := &dto.DashboardStats{}

	// Total orders
	s.db.Model(&models.Order{}).Count(&stats.TotalOrders)

	// Total revenue (from delivered orders)
	s.db.Model(&models.Order{}).
		Where("status = ?", models.OrderStatusDelivered).
		Select("COALESCE(SUM(total), 0)").
		Scan(&stats.TotalRevenue)

	// Total users
	s.db.Model(&models.User{}).Count(&stats.TotalUsers)

	// Pending orders
	s.db.Model(&models.Order{}).
		Where("status = ?", models.OrderStatusPending).
		Count(&stats.PendingOrders)

	// Orders by status
	var statusCounts []dto.OrderStatusCount
	s.db.Model(&models.Order{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts)
	stats.OrdersByStatus = statusCounts

	// New users this month
	firstOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	s.db.Model(&models.User{}).
		Where("created_at >= ?", firstOfMonth).
		Count(&stats.NewUsersThisMonth)

	// Recent orders (last 10)
	var recentOrders []models.Order
	s.db.Preload("User").
		Preload("OrderItems").
		Order("created_at DESC").
		Limit(10).
		Find(&recentOrders)

	for _, order := range recentOrders {
		stats.RecentOrders = append(stats.RecentOrders, dto.ToAdminOrderResponse(&order))
	}

	// Top selling products (by quantity sold)
	var topProducts []dto.TopProduct
	s.db.Model(&models.OrderItem{}).
		Select("products.id, products.name, products.image_url, SUM(order_items.quantity) as total_sold").
		Joins("JOIN products ON products.id = order_items.product_id").
		Group("products.id, products.name, products.image_url").
		Order("total_sold DESC").
		Limit(5).
		Scan(&topProducts)
	stats.TopProducts = topProducts

	return stats, nil
}
