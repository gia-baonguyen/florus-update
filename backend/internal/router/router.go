package router

import (
	"github.com/florus/backend/internal/config"
	"github.com/florus/backend/internal/handler"
	"github.com/florus/backend/internal/middleware"
	"github.com/florus/backend/internal/repository"
	"github.com/florus/backend/internal/service"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, cfg *config.Config) *gin.Engine {
	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	router := gin.New()

	// Global middleware
	router.Use(middleware.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	cartRepo := repository.NewCartRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	wishlistRepo := repository.NewWishlistRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	couponRepo := repository.NewCouponRepository(db)
	eventRepo := repository.NewEventRepository(db)

	// Initialize services
	jwtConfig := utils.JWTConfig{
		Secret:          cfg.JWT.Secret,
		ExpirationHours: cfg.JWT.ExpirationHours,
		Issuer:          cfg.JWT.Issuer,
	}

	authService := service.NewAuthService(userRepo, jwtConfig)
	googleAuthService := service.NewGoogleAuthService(userRepo, jwtConfig, cfg.Google)
	categoryService := service.NewCategoryService(categoryRepo)
	productService := service.NewProductService(productRepo, categoryRepo)
	cartService := service.NewCartService(cartRepo, productRepo)
	couponService := service.NewCouponService(couponRepo)
	loyaltyRepo := repository.NewLoyaltyRepository(db)
	loyaltyService := service.NewLoyaltyService(loyaltyRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo, userRepo, couponService, loyaltyService)
	recommendationService := service.NewRecommendationService(productRepo)
	adminService := service.NewAdminService(db)
	wishlistService := service.NewWishlistService(wishlistRepo, productRepo)
	reviewService := service.NewReviewService(reviewRepo, productRepo)
	paymentService := service.NewPaymentService(db, orderRepo, cfg.Payment)
	eventService := service.NewEventService(eventRepo, nil)
	addressService := service.NewAddressService(repository.NewUserAddressRepository(db))
	shippingService := service.NewShippingService(repository.NewShippingRepository(db))
	returnService := service.NewReturnService(repository.NewReturnRepository(db), orderRepo)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler()
	authHandler := handler.NewAuthHandler(authService, googleAuthService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	productHandler := handler.NewProductHandler(productService)
	cartHandler := handler.NewCartHandler(cartService)
	orderHandler := handler.NewOrderHandler(orderService)
	recommendationHandler := handler.NewRecommendationHandler(recommendationService)
	adminHandler := handler.NewAdminHandler(adminService)
	wishlistHandler := handler.NewWishlistHandler(wishlistService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	couponHandler := handler.NewCouponHandler(couponService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	eventHandler := handler.NewEventHandler(eventService)
	addressHandler := handler.NewAddressHandler(addressService)
	shippingHandler := handler.NewShippingHandler(shippingService)
	loyaltyHandler := handler.NewLoyaltyHandler(loyaltyService)
	returnHandler := handler.NewReturnHandler(returnService)

	// Routes
	api := router.Group("/api")
	{
		// Health check
		api.GET("/health", healthHandler.HealthCheck)
		api.GET("/ready", healthHandler.Ready)

		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/google", authHandler.GoogleLogin)
			// Demo forgot/reset password flows
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
			auth.GET("/me", middleware.Auth(cfg.JWT.Secret), authHandler.GetMe)
			auth.PUT("/profile", middleware.Auth(cfg.JWT.Secret), authHandler.UpdateProfile)
			auth.PUT("/password", middleware.Auth(cfg.JWT.Secret), authHandler.ChangePassword)
		}

		// Category routes
		categories := api.Group("/categories")
		{
			categories.GET("", categoryHandler.GetAll)
			categories.GET("/:id", categoryHandler.GetByID)

			// Admin only
			adminCategories := categories.Group("")
			adminCategories.Use(middleware.Auth(cfg.JWT.Secret), middleware.AdminOnly())
			{
				adminCategories.POST("", categoryHandler.Create)
				adminCategories.PUT("/:id", categoryHandler.Update)
				adminCategories.DELETE("/:id", categoryHandler.Delete)
			}
		}

		// Product routes
		products := api.Group("/products")
		{
			products.GET("", productHandler.GetAll)
			products.GET("/brands", productHandler.GetBrands)
			products.GET("/:id", productHandler.GetByID)
			products.GET("/slug/:slug", productHandler.GetBySlug)
			products.GET("/category/:categoryId", productHandler.GetByCategory)

			// Product reviews (public get, auth for create)
			products.GET("/:id/reviews", reviewHandler.GetProductReviews)
			products.POST("/:id/reviews", middleware.Auth(cfg.JWT.Secret), reviewHandler.CreateReview)
			products.GET("/:id/my-review", middleware.Auth(cfg.JWT.Secret), reviewHandler.GetUserReview)

			// Admin only
			adminProducts := products.Group("")
			adminProducts.Use(middleware.Auth(cfg.JWT.Secret), middleware.AdminOnly())
			{
				adminProducts.POST("", productHandler.Create)
				adminProducts.PUT("/:id", productHandler.Update)
				adminProducts.DELETE("/:id", productHandler.Delete)
			}
		}

		// Review routes (for update/delete)
		reviews := api.Group("/reviews")
		reviews.Use(middleware.Auth(cfg.JWT.Secret))
		{
			reviews.PUT("/:id", reviewHandler.UpdateReview)
			reviews.DELETE("/:id", reviewHandler.DeleteReview)
		}

		// Cart routes (authenticated)
		cart := api.Group("/cart")
		cart.Use(middleware.Auth(cfg.JWT.Secret))
		{
			cart.GET("", cartHandler.GetCart)
			cart.POST("", cartHandler.AddToCart)
			cart.PUT("/:itemId", cartHandler.UpdateQuantity)
			cart.DELETE("/:itemId", cartHandler.RemoveItem)
			cart.DELETE("", cartHandler.ClearCart)
		}

		// Address routes (authenticated)
		addresses := api.Group("/addresses")
		addresses.Use(middleware.Auth(cfg.JWT.Secret))
		{
			addresses.GET("", addressHandler.GetAddresses)
			addresses.POST("", addressHandler.CreateAddress)
			addresses.PUT("/:id", addressHandler.UpdateAddress)
			addresses.DELETE("/:id", addressHandler.DeleteAddress)
			addresses.PUT("/:id/default", addressHandler.SetDefaultAddress)
		}

		// Shipping routes (public)
		shipping := api.Group("/shipping")
		{
			shipping.GET("/methods", shippingHandler.GetMethods)
		}

		// Loyalty routes (authenticated)
		loyalty := api.Group("/loyalty")
		loyalty.Use(middleware.Auth(cfg.JWT.Secret))
		{
			loyalty.GET("", loyaltyHandler.GetLoyalty)
		}

		// Return / RMA routes (authenticated)
		returns := api.Group("/returns")
		returns.Use(middleware.Auth(cfg.JWT.Secret))
		{
			returns.GET("", returnHandler.GetReturns)
			returns.GET("/:id", returnHandler.GetReturnByID)
			returns.POST("", returnHandler.CreateReturn)
		}

		// Order routes (authenticated)
		orders := api.Group("/orders")
		orders.Use(middleware.Auth(cfg.JWT.Secret))
		{
			orders.GET("", orderHandler.GetOrders)
			orders.GET("/:id", orderHandler.GetOrderByID)
			orders.POST("", orderHandler.CreateOrder)
			orders.PUT("/:id/cancel", orderHandler.CancelOrder)
		}

		// Wishlist routes (authenticated)
		wishlist := api.Group("/wishlist")
		wishlist.Use(middleware.Auth(cfg.JWT.Secret))
		{
			wishlist.GET("", wishlistHandler.GetWishlist)
			wishlist.POST("", wishlistHandler.AddToWishlist)
			wishlist.DELETE("/:productId", wishlistHandler.RemoveFromWishlist)
			wishlist.GET("/check/:productId", wishlistHandler.CheckWishlist)
			wishlist.GET("/ids", wishlistHandler.GetWishlistIDs)
		}

		// Recommendation routes (public)
		recommendations := api.Group("/recommendations")
		{
			recommendations.GET("/cold-start", recommendationHandler.GetColdStart)
			recommendations.GET("/warm-start", recommendationHandler.GetWarmStart)
			recommendations.GET("/similar/:productId", recommendationHandler.GetSimilar)
			recommendations.GET("/co-viewed/:productId", recommendationHandler.GetCoViewed)
			recommendations.GET("/cross-sell/:productId", recommendationHandler.GetCrossSell)
		}

		// Coupon routes (public validation, list available, admin for management)
		coupons := api.Group("/coupons")
		{
			// Public
			coupons.POST("/validate", couponHandler.ValidateCoupon)
			coupons.GET("/available", couponHandler.GetAvailable)
		}

		// Payment routes
		payments := api.Group("/payments")
		{
			// Create payment (authenticated)
			payments.POST("/create", middleware.Auth(cfg.JWT.Secret), paymentHandler.CreatePayment)
			// Get payment status (authenticated)
			payments.GET("/status/:orderCode", middleware.Auth(cfg.JWT.Secret), paymentHandler.GetPaymentStatus)
			// Payment callbacks (public - called by payment providers)
			payments.POST("/callback/zalopay", paymentHandler.ZaloPayCallback)
			payments.POST("/callback/momo", paymentHandler.MoMoCallback)
			payments.GET("/callback/vnpay", paymentHandler.VNPayCallback)
		}

		// Event routes (public logging, some analytics)
		events := api.Group("/events")
		{
			// Generic event logging
			events.POST("", eventHandler.LogEvent)
			// Specialized helpers (optional for FE)
			events.POST("/product-view", eventHandler.LogProductView)
			events.POST("/add-to-cart", eventHandler.LogAddToCart)
			events.POST("/search", eventHandler.LogSearch)
			// Analytics / recommendation helpers
			events.GET("/popular-products", eventHandler.GetPopularProducts)
			events.GET("/product/:id/stats", eventHandler.GetProductStats)
			// User behavior profile (could be used for personalization dashboard)
			events.GET("/user/:id/behavior", eventHandler.GetUserBehavior)
		}

		// Admin routes (admin only)
		admin := api.Group("/admin")
		admin.Use(middleware.Auth(cfg.JWT.Secret), middleware.AdminOnly())
		{
			// Orders
			admin.GET("/orders", adminHandler.GetAllOrders)
			admin.GET("/orders/:id", adminHandler.GetOrderDetail)
			admin.PUT("/orders/:id/status", adminHandler.UpdateOrderStatus)

			// Users
			admin.GET("/users", adminHandler.GetAllUsers)
			admin.PUT("/users/:id/role", adminHandler.UpdateUserRole)

			// Stats
			admin.GET("/stats", adminHandler.GetDashboardStats)

			// Event stats for admin dashboards
			admin.GET("/events/stats", eventHandler.GetEventStats)

			// Coupons
			admin.GET("/coupons", couponHandler.GetAll)
			admin.GET("/coupons/:id", couponHandler.GetByID)
			admin.POST("/coupons", couponHandler.Create)
			admin.PUT("/coupons/:id", couponHandler.Update)
			admin.DELETE("/coupons/:id", couponHandler.Delete)
		}
	}

	return router
}
