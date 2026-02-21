package dto

// Auth DTOs
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required,min=2"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Category DTOs
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	ParentID    *uint  `json:"parent_id"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	ParentID    *uint  `json:"parent_id"`
	IsActive    *bool  `json:"is_active"`
}

// Product DTOs
type CreateProductRequest struct {
	Name            string   `json:"name" binding:"required"`
	Slug            string   `json:"slug" binding:"required"`
	Brand           string   `json:"brand"`
	Price           float64  `json:"price" binding:"required,gt=0"`
	OriginalPrice   *float64 `json:"original_price"`
	CategoryID      uint     `json:"category_id" binding:"required"`
	Description     string   `json:"description"`
	ImageURL        string   `json:"image_url"`
	Stock           int      `json:"stock"`
	Rating          float64  `json:"rating"`
	ReviewCount     int      `json:"review_count"`
	AIScore         *float64 `json:"ai_score"`
	AIRecommendType string   `json:"ai_recommend_type"`
	Tags            []string `json:"tags"`
	Ingredients     []string `json:"ingredients"`
}

type UpdateProductRequest struct {
	Name            string   `json:"name"`
	Slug            string   `json:"slug"`
	Brand           string   `json:"brand"`
	Price           *float64 `json:"price"`
	OriginalPrice   *float64 `json:"original_price"`
	CategoryID      *uint    `json:"category_id"`
	Description     string   `json:"description"`
	ImageURL        string   `json:"image_url"`
	Stock           *int     `json:"stock"`
	Rating          *float64 `json:"rating"`
	ReviewCount     *int     `json:"review_count"`
	AIScore         *float64 `json:"ai_score"`
	AIRecommendType string   `json:"ai_recommend_type"`
	IsActive        *bool    `json:"is_active"`
	Tags            []string `json:"tags"`
	Ingredients     []string `json:"ingredients"`
}

// Cart DTOs
type AddToCartRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

// Order DTOs
type CreateOrderRequest struct {
	ShippingAddress string `json:"shipping_address" binding:"required"`
	Note            string `json:"note"`
	CouponCode      string `json:"coupon_code"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// User DTOs
type UpdateUserRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

// Profile DTOs
type UpdateProfileRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// Review DTOs
type CreateReviewRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

type UpdateReviewRequest struct {
	Rating  *int    `json:"rating"`
	Comment *string `json:"comment"`
}
