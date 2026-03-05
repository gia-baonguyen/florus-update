package service

import (
	"errors"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
	"github.com/florus/backend/pkg/utils"
	"gorm.io/gorm"
)

var (
	ErrOrderNotFound    = errors.New("order not found")
	ErrEmptyCart        = errors.New("cart is empty")
	ErrCannotCancelOrder = errors.New("order cannot be cancelled")
)

type OrderService interface {
	CreateOrder(userID uint, req dto.CreateOrderRequest) (*dto.OrderResponse, error)
	BuyNow(userID uint, req dto.BuyNowRequest) (*dto.OrderResponse, error)
	GetOrderByID(userID uint, orderID uint) (*dto.OrderResponse, error)
	GetUserOrders(userID uint, pagination utils.Pagination) ([]dto.OrderResponse, *utils.Meta, error)
	CancelOrder(userID uint, orderID uint) error
}

type orderService struct {
	orderRepo     repository.OrderRepository
	cartRepo      repository.CartRepository
	productRepo   repository.ProductRepository
	userRepo      repository.UserRepository
	couponService CouponService
	loyaltyService LoyaltyService
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
	userRepo repository.UserRepository,
	couponService CouponService,
	loyaltyService LoyaltyService,
) OrderService {
	return &orderService{
		orderRepo:     orderRepo,
		cartRepo:      cartRepo,
		productRepo:   productRepo,
		userRepo:      userRepo,
		couponService: couponService,
		loyaltyService: loyaltyService,
	}
}

func (s *orderService) CreateOrder(userID uint, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	// Get cart items
	cartItems, err := s.cartRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	if len(cartItems) == 0 {
		return nil, ErrEmptyCart
	}

	// Start transaction
	db := s.orderRepo.GetDB()
	var order *models.Order
	var couponID uint

	err = db.Transaction(func(tx *gorm.DB) error {
		// 1. Validate stock and lock products
		var subtotal float64
		var orderItems []models.OrderItem

		for _, cartItem := range cartItems {
			// Lock product row for update
			product, err := s.productRepo.FindByIDWithLock(tx, cartItem.ProductID)
			if err != nil {
				return ErrProductNotFound
			}

			if product.Stock < cartItem.Quantity {
				return ErrInsufficientStock
			}

			// Calculate item total
			itemTotal := product.Price * float64(cartItem.Quantity)
			subtotal += itemTotal

			orderItems = append(orderItems, models.OrderItem{
				ProductID:  cartItem.ProductID,
				Quantity:   cartItem.Quantity,
				UnitPrice:  product.Price,
				TotalPrice: itemTotal,
			})

			// Deduct stock
			if err := s.productRepo.UpdateStock(tx, product.ID, cartItem.Quantity); err != nil {
				return err
			}
		}

		// 2. Calculate shipping
		shippingFee := models.CalculateShippingFee(subtotal)

		// 3. Validate and apply coupon if provided
		var discount float64
		if req.CouponCode != "" {
			coupon, discountAmount, err := s.couponService.ValidateCoupon(req.CouponCode, subtotal)
			if err != nil {
				return err
			}
			discount = discountAmount
			couponID = coupon.ID
		}

		// 4. Apply loyalty points if requested
		var loyaltyDiscount float64
		if s.loyaltyService != nil && req.LoyaltyPointsToUse != nil && *req.LoyaltyPointsToUse > 0 {
			_, loyaltyDiscount, err = s.loyaltyService.RedeemPoints(userID, *req.LoyaltyPointsToUse, subtotal-discount)
			if err != nil {
				return err
			}
		}

		// 5. Calculate total
		total := subtotal + shippingFee - discount - loyaltyDiscount

		// 6. Create order
		var shippingAddressID *uint
		if req.ShippingAddressID != nil {
			shippingAddressID = req.ShippingAddressID
		}

		shippingAddressText := req.ShippingAddress

		order = &models.Order{
			UserID:             userID,
			Subtotal:           subtotal,
			ShippingFee:        shippingFee,
			Discount:           discount + loyaltyDiscount,
			Total:              total,
			Status:             models.OrderStatusPending,
			ShippingAddress:    shippingAddressText,
			ShippingAddressID:  shippingAddressID,
			ShippingMethodCode: req.ShippingMethodCode,
			Note:               req.Note,
		}

		if err := s.orderRepo.Create(tx, order); err != nil {
			return err
		}

		// 4. Create order items
		for i := range orderItems {
			orderItems[i].OrderID = order.ID
		}

		if err := s.orderRepo.CreateItems(tx, orderItems); err != nil {
			return err
		}

		// 5. Clear cart
		if err := tx.Where("user_id = ?", userID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}

		// 6. Update user status to warm (use tx to stay in same transaction)
		if err := tx.Model(&models.User{}).Where("id = ?", userID).Update("user_status", models.StatusWarm).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Apply coupon usage count (after successful transaction)
	if couponID > 0 {
		_ = s.couponService.ApplyCoupon(couponID)
	}

	// Load complete order with items
	order, err = s.orderRepo.FindByID(order.ID)
	if err != nil {
		return nil, err
	}

	resp := dto.ToOrderResponse(order)
	return &resp, nil
}

// BuyNow creates an order directly from a single product without using the cart
func (s *orderService) BuyNow(userID uint, req dto.BuyNowRequest) (*dto.OrderResponse, error) {
	db := s.orderRepo.GetDB()
	var order *models.Order
	var couponID uint

	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. Validate stock and lock product
		product, err := s.productRepo.FindByIDWithLock(tx, req.ProductID)
		if err != nil {
			return ErrProductNotFound
		}

		if product.Stock < req.Quantity {
			return ErrInsufficientStock
		}

		// Calculate item total
		itemTotal := product.Price * float64(req.Quantity)
		subtotal := itemTotal

		orderItem := models.OrderItem{
			ProductID:  req.ProductID,
			Quantity:   req.Quantity,
			UnitPrice:  product.Price,
			TotalPrice: itemTotal,
		}

		// Deduct stock
		if err := s.productRepo.UpdateStock(tx, product.ID, req.Quantity); err != nil {
			return err
		}

		// 2. Calculate shipping
		shippingFee := models.CalculateShippingFee(subtotal)

		// 3. Validate and apply coupon if provided
		var discount float64
		if req.CouponCode != "" {
			coupon, discountAmount, err := s.couponService.ValidateCoupon(req.CouponCode, subtotal)
			if err != nil {
				return err
			}
			discount = discountAmount
			couponID = coupon.ID
		}

		// 4. Apply loyalty points if requested
		var loyaltyDiscount float64
		if s.loyaltyService != nil && req.LoyaltyPointsToUse != nil && *req.LoyaltyPointsToUse > 0 {
			_, loyaltyDiscount, err = s.loyaltyService.RedeemPoints(userID, *req.LoyaltyPointsToUse, subtotal-discount)
			if err != nil {
				return err
			}
		}

		// 5. Calculate total
		total := subtotal + shippingFee - discount - loyaltyDiscount

		// 6. Create order
		var shippingAddressID *uint
		if req.ShippingAddressID != nil {
			shippingAddressID = req.ShippingAddressID
		}

		order = &models.Order{
			UserID:             userID,
			Subtotal:           subtotal,
			ShippingFee:        shippingFee,
			Discount:           discount + loyaltyDiscount,
			Total:              total,
			Status:             models.OrderStatusPending,
			ShippingAddress:    req.ShippingAddress,
			ShippingAddressID:  shippingAddressID,
			ShippingMethodCode: req.ShippingMethodCode,
			Note:               req.Note,
		}

		if err := s.orderRepo.Create(tx, order); err != nil {
			return err
		}

		// 7. Create order item
		orderItem.OrderID = order.ID
		if err := s.orderRepo.CreateItems(tx, []models.OrderItem{orderItem}); err != nil {
			return err
		}

		// 8. Update user status to warm
		if err := tx.Model(&models.User{}).Where("id = ?", userID).Update("user_status", models.StatusWarm).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Apply coupon usage count (after successful transaction)
	if couponID > 0 {
		_ = s.couponService.ApplyCoupon(couponID)
	}

	// Load complete order with items
	order, err = s.orderRepo.FindByID(order.ID)
	if err != nil {
		return nil, err
	}

	resp := dto.ToOrderResponse(order)
	return &resp, nil
}

func (s *orderService) GetOrderByID(userID uint, orderID uint) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	if order.UserID != userID {
		return nil, ErrOrderNotFound
	}

	resp := dto.ToOrderResponse(order)
	return &resp, nil
}

func (s *orderService) GetUserOrders(userID uint, pagination utils.Pagination) ([]dto.OrderResponse, *utils.Meta, error) {
	orders, total, err := s.orderRepo.FindByUserID(userID, pagination)
	if err != nil {
		return nil, nil, err
	}

	var responses []dto.OrderResponse
	for _, order := range orders {
		responses = append(responses, dto.ToOrderResponse(&order))
	}

	meta := utils.NewMeta(pagination.Page, pagination.Limit, total)
	return responses, meta, nil
}

func (s *orderService) CancelOrder(userID uint, orderID uint) error {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return ErrOrderNotFound
	}

	if order.UserID != userID {
		return ErrOrderNotFound
	}

	if !order.CanCancel() {
		return ErrCannotCancelOrder
	}

	// Restore stock
	db := s.orderRepo.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, item := range order.OrderItems {
			if err := tx.Model(&models.Product{}).
				Where("id = ?", item.ProductID).
				Update("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				return err
			}
		}

		return s.orderRepo.UpdateStatus(orderID, models.OrderStatusCancelled)
	})

	return err
}
