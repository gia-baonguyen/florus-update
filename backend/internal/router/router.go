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

	// Initialize services
	jwtConfig := utils.JWTConfig{
		Secret:          cfg.JWT.Secret,
		ExpirationHours: cfg.JWT.ExpirationHours,
		Issuer:          cfg.JWT.Issuer,
	}

	authService := service.NewAuthService(userRepo, jwtConfig)
	categoryService := service.NewCategoryService(categoryRepo)
	productService := service.NewProductService(productRepo, categoryRepo)
	cartService := service.NewCartService(cartRepo, productRepo)
	couponService := service.NewCouponService(couponRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo, userRepo, couponService)
	recommendationService := service.NewRecommendationService(productRepo)
	adminService := service.NewAdminService(db)
	wishlistService := service.NewWishlistService(wishlistRepo, productRepo)
	reviewService := service.NewReviewService(reviewRepo, productRepo)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler()
	authHandler := handler.NewAuthHandler(authService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	productHandler := handler.NewProductHandler(productService)
	cartHandler := handler.NewCartHandler(cartService)
	orderHandler := handler.NewOrderHandler(orderService)
	recommendationHandler := handler.NewRecommendationHandler(recommendationService)
	adminHandler := handler.NewAdminHandler(adminService)
	wishlistHandler := handler.NewWishlistHandler(wishlistService)
	reviewHandler := handler.NewReviewHandler(reviewService)
	couponHandler := handler.NewCouponHandler(couponService)

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

		// Coupon routes (public validation, admin for management)
		coupons := api.Group("/coupons")
		{
			coupons.POST("/validate", couponHandler.ValidateCoupon)
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
