package service

import (
	"errors"

	"github.com/florus/backend/internal/dto"
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/repository"
)

var (
	ErrReviewNotFound     = errors.New("review not found")
	ErrAlreadyReviewed    = errors.New("you have already reviewed this product")
	ErrInvalidRating      = errors.New("rating must be between 1 and 5")
	ErrNotReviewOwner     = errors.New("you can only modify your own reviews")
)

type ReviewService interface {
	GetProductReviews(productID uint) (*dto.ReviewListResponse, error)
	CreateReview(userID, productID uint, req dto.CreateReviewRequest) (*dto.ReviewResponse, error)
	UpdateReview(userID, reviewID uint, req dto.UpdateReviewRequest) (*dto.ReviewResponse, error)
	DeleteReview(userID, reviewID uint) error
	GetUserReview(userID, productID uint) (*dto.ReviewResponse, error)
}

type reviewService struct {
	reviewRepo  repository.ReviewRepository
	productRepo repository.ProductRepository
}

func NewReviewService(reviewRepo repository.ReviewRepository, productRepo repository.ProductRepository) ReviewService {
	return &reviewService{
		reviewRepo:  reviewRepo,
		productRepo: productRepo,
	}
}

func (s *reviewService) GetProductReviews(productID uint) (*dto.ReviewListResponse, error) {
	reviews, err := s.reviewRepo.FindByProductID(productID)
	if err != nil {
		return nil, err
	}

	avgRating, count, err := s.reviewRepo.GetAverageRating(productID)
	if err != nil {
		return nil, err
	}

	var reviewResponses []dto.ReviewResponse
	for _, review := range reviews {
		reviewResponses = append(reviewResponses, dto.ToReviewResponse(&review))
	}

	// Ensure JSON encodes empty array [] instead of null
	if reviewResponses == nil {
		reviewResponses = []dto.ReviewResponse{}
	}

	return &dto.ReviewListResponse{
		Reviews:       reviewResponses,
		TotalReviews:  count,
		AverageRating: avgRating,
	}, nil
}

func (s *reviewService) CreateReview(userID, productID uint, req dto.CreateReviewRequest) (*dto.ReviewResponse, error) {
	// Validate rating
	if req.Rating < 1 || req.Rating > 5 {
		return nil, ErrInvalidRating
	}

	// Check if product exists
	_, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, ErrProductNotFound
	}

	// Check if user already reviewed
	existing, _ := s.reviewRepo.FindByUserAndProduct(userID, productID)
	if existing != nil {
		return nil, ErrAlreadyReviewed
	}

	// Create review
	review := &models.Review{
		UserID:    userID,
		ProductID: productID,
		Rating:    req.Rating,
		Comment:   req.Comment,
	}
	if err := s.reviewRepo.Create(review); err != nil {
		return nil, err
	}

	// Add images if provided
	if len(req.Images) > 0 {
		if err := s.reviewRepo.AddImages(review.ID, req.Images); err != nil {
			// Log but don't fail the review creation
			// Images can be added later
		}
	}

	// Update product rating
	s.updateProductRating(productID)

	// Reload with user data and images
	review, _ = s.reviewRepo.FindByID(review.ID)
	resp := dto.ToReviewResponse(review)
	return &resp, nil
}

func (s *reviewService) UpdateReview(userID, reviewID uint, req dto.UpdateReviewRequest) (*dto.ReviewResponse, error) {
	review, err := s.reviewRepo.FindByID(reviewID)
	if err != nil {
		return nil, ErrReviewNotFound
	}

	// Check ownership
	if review.UserID != userID {
		return nil, ErrNotReviewOwner
	}

	// Update fields
	if req.Rating != nil {
		if *req.Rating < 1 || *req.Rating > 5 {
			return nil, ErrInvalidRating
		}
		review.Rating = *req.Rating
	}
	if req.Comment != nil {
		review.Comment = *req.Comment
	}

	if err := s.reviewRepo.Update(review); err != nil {
		return nil, err
	}

	// Update images if provided (replace existing using transaction)
	if req.Images != nil {
		if err := s.reviewRepo.ReplaceImages(reviewID, req.Images); err != nil {
			return nil, err
		}
	}

	// Update product rating
	s.updateProductRating(review.ProductID)

	// Reload to get updated images
	review, _ = s.reviewRepo.FindByID(reviewID)
	resp := dto.ToReviewResponse(review)
	return &resp, nil
}

func (s *reviewService) DeleteReview(userID, reviewID uint) error {
	review, err := s.reviewRepo.FindByID(reviewID)
	if err != nil {
		return ErrReviewNotFound
	}

	// Check ownership
	if review.UserID != userID {
		return ErrNotReviewOwner
	}

	productID := review.ProductID
	if err := s.reviewRepo.Delete(reviewID); err != nil {
		return err
	}

	// Update product rating
	s.updateProductRating(productID)

	return nil
}

func (s *reviewService) GetUserReview(userID, productID uint) (*dto.ReviewResponse, error) {
	review, err := s.reviewRepo.FindByUserAndProduct(userID, productID)
	if err != nil {
		return nil, ErrReviewNotFound
	}
	// Reload to get images
	review, _ = s.reviewRepo.FindByID(review.ID)
	resp := dto.ToReviewResponse(review)
	return &resp, nil
}

func (s *reviewService) updateProductRating(productID uint) {
	avgRating, count, err := s.reviewRepo.GetAverageRating(productID)
	if err != nil {
		return
	}

	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return
	}

	product.Rating = avgRating
	product.ReviewCount = count
	s.productRepo.Update(product)
}
