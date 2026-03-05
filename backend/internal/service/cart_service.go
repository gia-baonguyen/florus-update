package service

import (
	"errors"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrCartItemNotFound = errors.New("cart item not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type CartService interface {
	AddToCart(userID uint, req dto.AddToCartRequest) (*dto.CartResponse, error)
	GetCart(userID uint) (*dto.CartResponse, error)
	UpdateQuantity(userID uint, itemID uint, req dto.UpdateCartItemRequest) (*dto.CartResponse, error)
	RemoveItem(userID uint, itemID uint) error
	ClearCart(userID uint) error
}

type cartService struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

func NewCartService(cartRepo repository.CartRepository, productRepo repository.ProductRepository) CartService {
	return &cartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *cartService) AddToCart(userID uint, req dto.AddToCartRequest) (*dto.CartResponse, error) {
	// Check if product exists and has enough stock
	product, err := s.productRepo.FindByID(req.ProductID)
	if err != nil {
		return nil, ErrProductNotFound
	}

	if product.Stock < req.Quantity {
		return nil, ErrInsufficientStock
	}

	// Check if item already in cart
	existingItem, err := s.cartRepo.FindItem(userID, req.ProductID)
	if err == nil && existingItem != nil {
		// Update quantity
		newQuantity := existingItem.Quantity + req.Quantity
		if product.Stock < newQuantity {
			return nil, ErrInsufficientStock
		}
		if err := s.cartRepo.UpdateQuantity(existingItem.ID, newQuantity); err != nil {
			return nil, err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) || existingItem == nil {
		// Add new item
		cartItem := &models.CartItem{
			UserID:    userID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		}
		if err := s.cartRepo.AddItem(cartItem); err != nil {
			return nil, err
		}
	}

	return s.GetCart(userID)
}

func (s *cartService) GetCart(userID uint) (*dto.CartResponse, error) {
	items, err := s.cartRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	// In some older data, there may be multiple cart_items rows for the same product.
	// To make the UX consistent, we aggregate items by ProductID here so that
	// each product appears only once in the cart (with summed quantity/total).
	type aggItem struct {
		id       uint
		product  models.Product
		quantity int
	}

	aggregated := make(map[uint]*aggItem) // key: ProductID

	for _, item := range items {
		if existing, ok := aggregated[item.ProductID]; ok {
			existing.quantity += item.Quantity
			// id và product giữ nguyên từ bản ghi đầu tiên
		} else {
			aggregated[item.ProductID] = &aggItem{
				id:       item.ID,
				product:  item.Product,
				quantity: item.Quantity,
			}
		}
	}

	var cartItems []dto.CartItemResponse
	var subtotal float64
	var itemCount int

	for productID, ai := range aggregated {
		itemTotal := ai.product.Price * float64(ai.quantity)
		subtotal += itemTotal
		itemCount += ai.quantity

		cartItems = append(cartItems, dto.CartItemResponse{
			ID:        ai.id,
			ProductID: productID,
			Product:   dto.ToProductResponse(&ai.product),
			Quantity:  ai.quantity,
			Price:     ai.product.Price,
			Total:     itemTotal,
		})
	}

	shippingFee := models.CalculateShippingFee(subtotal)

	return &dto.CartResponse{
		Items:       cartItems,
		Subtotal:    subtotal,
		ShippingFee: shippingFee,
		Total:       subtotal + shippingFee,
		ItemCount:   itemCount,
	}, nil
}

func (s *cartService) UpdateQuantity(userID uint, itemID uint, req dto.UpdateCartItemRequest) (*dto.CartResponse, error) {
	item, err := s.cartRepo.FindItemByID(itemID)
	if err != nil {
		return nil, ErrCartItemNotFound
	}

	if item.UserID != userID {
		return nil, ErrCartItemNotFound
	}

	// Check stock
	product, err := s.productRepo.FindByID(item.ProductID)
	if err != nil {
		return nil, ErrProductNotFound
	}

	if product.Stock < req.Quantity {
		return nil, ErrInsufficientStock
	}

	if err := s.cartRepo.UpdateQuantity(itemID, req.Quantity); err != nil {
		return nil, err
	}

	return s.GetCart(userID)
}

func (s *cartService) RemoveItem(userID uint, itemID uint) error {
	item, err := s.cartRepo.FindItemByID(itemID)
	if err != nil {
		return ErrCartItemNotFound
	}

	if item.UserID != userID {
		return ErrCartItemNotFound
	}

	return s.cartRepo.DeleteItem(itemID)
}

func (s *cartService) ClearCart(userID uint) error {
	return s.cartRepo.ClearCart(userID)
}
