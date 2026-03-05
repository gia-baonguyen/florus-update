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

// Google OAuth DTO
type GoogleLoginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
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
	// Legacy free-text address (kept for backward compatibility)
	ShippingAddress string `json:"shipping_address"`
	// New structured fields (preferred)
	ShippingAddressID  *uint  `json:"shipping_address_id"`
	ShippingMethodCode string `json:"shipping_method_code"`
	Note               string `json:"note"`
	CouponCode         string `json:"coupon_code"`
	PaymentMethod      string `json:"payment_method"` // cod, zalopay, momo, vnpay
	// Optional loyalty points the user wants to redeem on this order
	LoyaltyPointsToUse *int64 `json:"loyalty_points_to_use"`
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

// Address DTOs
type CreateAddressRequest struct {
	FullName    string `json:"full_name" binding:"required"`
	Phone       string `json:"phone" binding:"required"`
	Street      string `json:"street" binding:"required"`
	City        string `json:"city" binding:"required"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
	IsDefault   bool   `json:"is_default"`
}

type UpdateAddressRequest struct {
	FullName    string `json:"full_name"`
	Phone       string `json:"phone"`
	Street      string `json:"street"`
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
	IsDefault   *bool  `json:"is_default"`
}

// Returns DTOs
type CreateReturnRequest struct {
	OrderID uint                    `json:"order_id" binding:"required"`
	Items   []CreateReturnItemEntry `json:"items" binding:"required,dive"`
	Reason  string                  `json:"reason"`
	Note    string                  `json:"note"`
}

type CreateReturnItemEntry struct {
	OrderItemID uint `json:"order_item_id" binding:"required"`
	Quantity    int  `json:"quantity" binding:"required,gt=0"`
}

// Payment DTOs
type CreatePaymentRequest struct {
	OrderCode     string `json:"order_code" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"` // zalopay, momo, vnpay
}

type PaymentCallbackRequest struct {
	// Common fields
	TransactionID string `json:"transaction_id"`
	OrderCode     string `json:"order_code"`
	Status        string `json:"status"`
	Amount        int64  `json:"amount"`

	// ZaloPay specific
	AppTransID string `json:"app_trans_id"`
	ZpTransID  string `json:"zp_trans_id"`

	// MoMo specific
	RequestID   string `json:"requestId"`
	PartnerCode string `json:"partnerCode"`
	OrderID     string `json:"orderId"`
	ResultCode  int    `json:"resultCode"`

	// VNPay specific
	VnpTxnRef        string `json:"vnp_TxnRef"`
	VnpTransactionNo string `json:"vnp_TransactionNo"`
	VnpResponseCode  string `json:"vnp_ResponseCode"`
}
