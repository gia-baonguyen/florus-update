package dto

import "github.com/florus/backend/internal/models"

// Auth Response
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// User Response
type UserResponse struct {
	ID           uint                `json:"id"`
	Email        string              `json:"email"`
	Name         string              `json:"name"`
	Phone        string              `json:"phone,omitempty"`
	Address      string              `json:"address,omitempty"`
	Role         models.UserRole     `json:"role"`
	UserStatus   models.UserStatus   `json:"user_status"`
	AuthProvider models.AuthProvider `json:"auth_provider"`
	AvatarURL    string              `json:"avatar_url,omitempty"`
	LoyaltyTier  string              `json:"loyalty_tier"`
	LoyaltyPoints int64              `json:"loyalty_points"`
}

func ToUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		Name:         user.Name,
		Phone:        user.Phone,
		Address:      user.Address,
		Role:         user.Role,
		UserStatus:   user.UserStatus,
		AuthProvider: user.AuthProvider,
		AvatarURL:    user.AvatarURL,
		LoyaltyTier:  user.LoyaltyTier,
		LoyaltyPoints: user.LoyaltyPoints,
	}
}

// Product Response
type ProductResponse struct {
	ID              uint                       `json:"id"`
	Name            string                     `json:"name"`
	Slug            string                     `json:"slug"`
	Brand           string                     `json:"brand,omitempty"`
	Price           float64                    `json:"price"`
	OriginalPrice   *float64                   `json:"original_price,omitempty"`
	Discount        float64                    `json:"discount,omitempty"`
	CategoryID      uint                       `json:"category_id"`
	Category        *CategoryResponse          `json:"category,omitempty"`
	Description     string                     `json:"description,omitempty"`
	ImageURL        string                     `json:"image_url,omitempty"`
	Images          []ProductImageResponse     `json:"images,omitempty"`
	Stock           int                        `json:"stock"`
	InStock         bool                       `json:"in_stock"`
	Rating          float64                    `json:"rating"`
	ReviewCount     int                        `json:"review_count"`
	AIScore         *float64                   `json:"ai_score,omitempty"`
	AIRecommendType string                     `json:"ai_recommend_type,omitempty"`
	Tags            []string                   `json:"tags,omitempty"`
	Ingredients     []string                   `json:"ingredients,omitempty"`
}

type ProductImageResponse struct {
	ID        uint   `json:"id"`
	URL       string `json:"url"`
	AltText   string `json:"alt_text,omitempty"`
	SortOrder int    `json:"sort_order"`
	IsPrimary bool   `json:"is_primary"`
}

func ToProductResponse(product *models.Product) ProductResponse {
	resp := ProductResponse{
		ID:              product.ID,
		Name:            product.Name,
		Slug:            product.Slug,
		Brand:           product.Brand,
		Price:           product.Price,
		OriginalPrice:   product.OriginalPrice,
		Discount:        product.CalculateDiscount(),
		CategoryID:      product.CategoryID,
		Description:     product.Description,
		ImageURL:        product.ImageURL,
		Stock:           product.Stock,
		InStock:         product.IsInStock(),
		Rating:          product.Rating,
		ReviewCount:     product.ReviewCount,
		AIScore:         product.AIScore,
		AIRecommendType: product.AIRecommendType,
	}

	if product.Category.ID != 0 {
		cat := ToCategoryResponse(&product.Category)
		resp.Category = &cat
	}

	for _, tag := range product.Tags {
		resp.Tags = append(resp.Tags, tag.TagName)
	}

	for _, ing := range product.Ingredients {
		resp.Ingredients = append(resp.Ingredients, ing.IngredientName)
	}

	for _, img := range product.Images {
		resp.Images = append(resp.Images, ProductImageResponse{
			ID:        img.ID,
			URL:       img.URL,
			AltText:   img.AltText,
			SortOrder: img.SortOrder,
			IsPrimary: img.IsPrimary,
		})
	}

	return resp
}

// Category Response
type CategoryResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	IsActive    bool   `json:"is_active"`
}

func ToCategoryResponse(category *models.Category) CategoryResponse {
	return CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		ImageURL:    category.ImageURL,
		ParentID:    category.ParentID,
		IsActive:    category.IsActive,
	}
}

// Cart Response
type CartResponse struct {
	Items       []CartItemResponse `json:"items"`
	Subtotal    float64            `json:"subtotal"`
	ShippingFee float64            `json:"shipping_fee"`
	Total       float64            `json:"total"`
	ItemCount   int                `json:"item_count"`
}

type CartItemResponse struct {
	ID        uint            `json:"id"`
	ProductID uint            `json:"product_id"`
	Product   ProductResponse `json:"product"`
	Quantity  int             `json:"quantity"`
	Price     float64         `json:"price"`
	Total     float64         `json:"total"`
}

// Order Response
type OrderResponse struct {
	ID              uint                 `json:"id"`
	OrderCode       string               `json:"order_code"`
	Subtotal        float64              `json:"subtotal"`
	ShippingFee     float64              `json:"shipping_fee"`
	Discount        float64              `json:"discount"`
	Total           float64              `json:"total"`
	Status          models.OrderStatus   `json:"status"`
	PaymentMethod   models.PaymentMethod `json:"payment_method"`
	PaymentStatus   models.PaymentStatus `json:"payment_status"`
	ShippingAddress string               `json:"shipping_address"`
	Note            string               `json:"note,omitempty"`
	Items           []OrderItemResponse  `json:"items,omitempty"`
	Payment         *PaymentResponse     `json:"payment,omitempty"`
	CreatedAt       string               `json:"created_at"`
}

type OrderItemResponse struct {
	ID         uint            `json:"id"`
	ProductID  uint            `json:"product_id"`
	Product    ProductResponse `json:"product,omitempty"`
	Quantity   int             `json:"quantity"`
	UnitPrice  float64         `json:"unit_price"`
	TotalPrice float64         `json:"total_price"`
}

func ToOrderResponse(order *models.Order) OrderResponse {
	resp := OrderResponse{
		ID:              order.ID,
		OrderCode:       order.OrderCode,
		Subtotal:        order.Subtotal,
		ShippingFee:     order.ShippingFee,
		Discount:        order.Discount,
		Total:           order.Total,
		Status:          order.Status,
		PaymentMethod:   order.PaymentMethod,
		PaymentStatus:   order.PaymentStatus,
		ShippingAddress: order.ShippingAddress,
		Note:            order.Note,
		CreatedAt:       order.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	for _, item := range order.OrderItems {
		resp.Items = append(resp.Items, OrderItemResponse{
			ID:         item.ID,
			ProductID:  item.ProductID,
			Product:    ToProductResponse(&item.Product),
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
		})
	}

	if order.Payment != nil {
		payment := ToPaymentResponse(order.Payment)
		resp.Payment = &payment
	}

	return resp
}

// Admin Order Response (includes user info)
type AdminOrderResponse struct {
	ID              uint                `json:"id"`
	OrderCode       string              `json:"order_code"`
	User            UserResponse        `json:"user"`
	Subtotal        float64             `json:"subtotal"`
	ShippingFee     float64             `json:"shipping_fee"`
	Discount        float64             `json:"discount"`
	Total           float64             `json:"total"`
	Status          models.OrderStatus  `json:"status"`
	ShippingAddress string              `json:"shipping_address"`
	Note            string              `json:"note,omitempty"`
	Items           []OrderItemResponse `json:"items,omitempty"`
	CreatedAt       string              `json:"created_at"`
	UpdatedAt       string              `json:"updated_at"`
}

func ToAdminOrderResponse(order *models.Order) AdminOrderResponse {
	resp := AdminOrderResponse{
		ID:              order.ID,
		OrderCode:       order.OrderCode,
		User:            ToUserResponse(&order.User),
		Subtotal:        order.Subtotal,
		ShippingFee:     order.ShippingFee,
		Discount:        order.Discount,
		Total:           order.Total,
		Status:          order.Status,
		ShippingAddress: order.ShippingAddress,
		Note:            order.Note,
		CreatedAt:       order.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       order.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	for _, item := range order.OrderItems {
		resp.Items = append(resp.Items, OrderItemResponse{
			ID:         item.ID,
			ProductID:  item.ProductID,
			Product:    ToProductResponse(&item.Product),
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			TotalPrice: item.TotalPrice,
		})
	}

	return resp
}

// Admin User Response (includes stats)
type AdminUserResponse struct {
	ID         uint              `json:"id"`
	Email      string            `json:"email"`
	Name       string            `json:"name"`
	Phone      string            `json:"phone,omitempty"`
	Address    string            `json:"address,omitempty"`
	Role       models.UserRole   `json:"role"`
	UserStatus models.UserStatus `json:"user_status"`
	OrderCount int64             `json:"order_count"`
	CreatedAt  string            `json:"created_at"`
}

func ToAdminUserResponse(user *models.User, orderCount int64) AdminUserResponse {
	return AdminUserResponse{
		ID:         user.ID,
		Email:      user.Email,
		Name:       user.Name,
		Phone:      user.Phone,
		Address:    user.Address,
		Role:       user.Role,
		UserStatus: user.UserStatus,
		OrderCount: orderCount,
		CreatedAt:  user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// Address responses
type UserAddressResponse struct {
	ID          uint   `json:"id"`
	FullName    string `json:"full_name"`
	Phone       string `json:"phone"`
	Street      string `json:"street"`
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	CountryCode string `json:"country_code"`
	IsDefault   bool   `json:"is_default"`
}

func ToUserAddressResponse(a *models.UserAddress) UserAddressResponse {
	return UserAddressResponse{
		ID:          a.ID,
		FullName:    a.FullName,
		Phone:       a.Phone,
		Street:      a.Street,
		City:        a.City,
		State:       a.State,
		PostalCode:  a.PostalCode,
		CountryCode: a.CountryCode,
		IsDefault:   a.IsDefault,
	}
}

// Shipping responses
type ShippingMethodResponse struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Return / RMA responses
type ReturnItemResponse struct {
	ID           uint    `json:"id"`
	OrderItemID  uint    `json:"order_item_id"`
	Product      ProductResponse `json:"product,omitempty"`
	Quantity     int     `json:"quantity"`
	RefundAmount float64 `json:"refund_amount"`
	Status       string  `json:"status,omitempty"`
}

type ReturnResponse struct {
	ID        uint                `json:"id"`
	OrderID   uint                `json:"order_id"`
	Status    string              `json:"status"`
	Reason    string              `json:"reason,omitempty"`
	Note      string              `json:"note,omitempty"`
	Items     []ReturnItemResponse `json:"items"`
	CreatedAt string              `json:"created_at"`
	UpdatedAt string              `json:"updated_at"`
}

func ToReturnResponse(r *models.Return) ReturnResponse {
	resp := ReturnResponse{
		ID:        r.ID,
		OrderID:   r.OrderID,
		Status:    string(r.Status),
		Reason:    r.Reason,
		Note:      r.Note,
		CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: r.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	for _, it := range r.Items {
		itemResp := ReturnItemResponse{
			ID:           it.ID,
			OrderItemID:  it.OrderItemID,
			Quantity:     it.Quantity,
			RefundAmount: it.RefundAmount,
			Status:       it.Status,
		}
		if it.OrderItem.ID != 0 {
			itemResp.Product = ToProductResponse(&it.OrderItem.Product)
		}
		resp.Items = append(resp.Items, itemResp)
	}
	return resp
}

// Dashboard Stats
type DashboardStats struct {
	TotalOrders       int64                `json:"total_orders"`
	TotalRevenue      float64              `json:"total_revenue"`
	TotalUsers        int64                `json:"total_users"`
	PendingOrders     int64                `json:"pending_orders"`
	OrdersByStatus    []OrderStatusCount   `json:"orders_by_status"`
	NewUsersThisMonth int64                `json:"new_users_this_month"`
	RecentOrders      []AdminOrderResponse `json:"recent_orders"`
	TopProducts       []TopProduct         `json:"top_products"`
}

type OrderStatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type TopProduct struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	ImageURL  string `json:"image_url"`
	TotalSold int64  `json:"total_sold"`
}

// Wishlist Response
type WishlistResponse struct {
	Items      []WishlistItemResponse `json:"items"`
	TotalItems int                    `json:"total_items"`
}

type WishlistItemResponse struct {
	ID        uint            `json:"id"`
	ProductID uint            `json:"product_id"`
	Product   ProductResponse `json:"product"`
	CreatedAt string          `json:"created_at"`
}

func ToWishlistItemResponse(item *models.WishlistItem) WishlistItemResponse {
	return WishlistItemResponse{
		ID:        item.ID,
		ProductID: item.ProductID,
		Product:   ToProductResponse(&item.Product),
		CreatedAt: item.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// Review Response
type ReviewResponse struct {
	ID        uint                  `json:"id"`
	UserID    uint                  `json:"user_id"`
	ProductID uint                  `json:"product_id"`
	Rating    int                   `json:"rating"`
	Comment   string                `json:"comment"`
	Images    []ReviewImageResponse `json:"images,omitempty"`
	User      UserResponse          `json:"user"`
	CreatedAt string                `json:"created_at"`
	UpdatedAt string                `json:"updated_at"`
}

type ReviewImageResponse struct {
	ID        uint   `json:"id"`
	URL       string `json:"url"`
	SortOrder int    `json:"sort_order"`
}

type ReviewListResponse struct {
	Reviews       []ReviewResponse `json:"reviews"`
	TotalReviews  int              `json:"total_reviews"`
	AverageRating float64          `json:"average_rating"`
}

func ToReviewResponse(review *models.Review) ReviewResponse {
	resp := ReviewResponse{
		ID:        review.ID,
		UserID:    review.UserID,
		ProductID: review.ProductID,
		Rating:    review.Rating,
		Comment:   review.Comment,
		User:      ToUserResponse(&review.User),
		CreatedAt: review.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: review.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	for _, img := range review.Images {
		resp.Images = append(resp.Images, ReviewImageResponse{
			ID:        img.ID,
			URL:       img.URL,
			SortOrder: img.SortOrder,
		})
	}
	return resp
}

// Payment Response
type PaymentResponse struct {
	ID              uint                 `json:"id"`
	OrderID         uint                 `json:"order_id"`
	TransactionID   string               `json:"transaction_id"`
	Method          models.PaymentMethod `json:"method"`
	Amount          float64              `json:"amount"`
	Status          models.PaymentStatus `json:"status"`
	ProviderTransID string               `json:"provider_trans_id,omitempty"`
	PaymentURL      string               `json:"payment_url,omitempty"`
	ErrorMessage    string               `json:"error_message,omitempty"`
	PaidAt          string               `json:"paid_at,omitempty"`
	CreatedAt       string               `json:"created_at"`
}

func ToPaymentResponse(payment *models.Payment) PaymentResponse {
	resp := PaymentResponse{
		ID:              payment.ID,
		OrderID:         payment.OrderID,
		TransactionID:   payment.TransactionID,
		Method:          payment.Method,
		Amount:          payment.Amount,
		Status:          payment.Status,
		ProviderTransID: payment.ProviderTransID,
		PaymentURL:      payment.PaymentURL,
		ErrorMessage:    payment.ErrorMessage,
		CreatedAt:       payment.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if payment.PaidAt != nil {
		resp.PaidAt = payment.PaidAt.Format("2006-01-02 15:04:05")
	}
	return resp
}

// Create Payment Response (includes payment URL)
type CreatePaymentResponse struct {
	OrderCode  string `json:"order_code"`
	PaymentURL string `json:"payment_url"`
	Method     string `json:"method"`
}
