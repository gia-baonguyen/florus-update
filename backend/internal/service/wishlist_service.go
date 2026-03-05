package service

import (
	"errors"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
)

var (
	ErrProductAlreadyInWishlist = errors.New("product already in wishlist")
	ErrWishlistItemNotFound     = errors.New("wishlist item not found")
)

type WishlistService interface {
	GetWishlist(userID uint) (*dto.WishlistResponse, error)
	AddToWishlist(userID, productID uint) (*dto.WishlistItemResponse, error)
	RemoveFromWishlist(userID, productID uint) error
	IsInWishlist(userID, productID uint) (bool, error)
	GetWishlistProductIDs(userID uint) ([]uint, error)
}

type wishlistService struct {
	wishlistRepo repository.WishlistRepository
	productRepo  repository.ProductRepository
}

func NewWishlistService(wishlistRepo repository.WishlistRepository, productRepo repository.ProductRepository) WishlistService {
	return &wishlistService{
		wishlistRepo: wishlistRepo,
		productRepo:  productRepo,
	}
}

func (s *wishlistService) GetWishlist(userID uint) (*dto.WishlistResponse, error) {
	items, err := s.wishlistRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Initialize as empty slice to avoid null in JSON response
	itemResponses := make([]dto.WishlistItemResponse, 0)
	for _, item := range items {
		itemResponses = append(itemResponses, dto.ToWishlistItemResponse(&item))
	}

	return &dto.WishlistResponse{
		Items:      itemResponses,
		TotalItems: len(itemResponses),
	}, nil
}

func (s *wishlistService) AddToWishlist(userID, productID uint) (*dto.WishlistItemResponse, error) {
	// Check if product exists
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, ErrProductNotFound
	}

	// Check if already in wishlist
	exists, err := s.wishlistRepo.Exists(userID, productID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProductAlreadyInWishlist
	}

	// Add to wishlist
	item := &models.WishlistItem{
		UserID:    userID,
		ProductID: productID,
	}
	if err := s.wishlistRepo.Create(item); err != nil {
		return nil, err
	}

	item.Product = *product
	resp := dto.ToWishlistItemResponse(item)
	return &resp, nil
}

func (s *wishlistService) RemoveFromWishlist(userID, productID uint) error {
	exists, err := s.wishlistRepo.Exists(userID, productID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrWishlistItemNotFound
	}

	return s.wishlistRepo.DeleteByUserAndProduct(userID, productID)
}

func (s *wishlistService) IsInWishlist(userID, productID uint) (bool, error) {
	return s.wishlistRepo.Exists(userID, productID)
}

func (s *wishlistService) GetWishlistProductIDs(userID uint) ([]uint, error) {
	return s.wishlistRepo.GetProductIDs(userID)
}
