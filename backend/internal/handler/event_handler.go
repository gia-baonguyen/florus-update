package handler

import (
	"net/http"
	"strconv"

	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EventHandler struct {
	eventService service.EventService
}

func NewEventHandler(eventService service.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

// EventRequest represents the request body for logging events
type EventRequest struct {
	EventType   string  `json:"event_type" binding:"required"`
	ProductID   *uint   `json:"product_id"`
	CategoryID  *uint   `json:"category_id"`
	OrderID     *uint   `json:"order_id"`
	SearchQuery string  `json:"search_query"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Rating      *int    `json:"rating"`
	Metadata    string  `json:"metadata"`
	UserID      *uint   `json:"user_id"` // Optional: for data import purposes
}

// LogEvent handles POST /api/events - logs a user event
// @Summary Log user event
// @Description Log user interaction event for analytics and recommendations
// @Tags Events
// @Accept json
// @Produce json
// @Param event body EventRequest true "Event data"
// @Success 201 {object} map[string]interface{}
// @Router /events [post]
func (h *EventHandler) LogEvent(c *gin.Context) {
	var req EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (may be nil for anonymous users)
	var userID *uint
	if id, exists := c.Get("user_id"); exists {
		uid := id.(uint)
		userID = &uid
	}
	// Allow user_id from request body for data import purposes
	if userID == nil && req.UserID != nil {
		userID = req.UserID
	}

	// Get or create session ID
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// Create event
	event := &models.UserEvent{
		UserID:      userID,
		SessionID:   sessionID,
		EventType:   models.EventType(req.EventType),
		ProductID:   req.ProductID,
		CategoryID:  req.CategoryID,
		OrderID:     req.OrderID,
		SearchQuery: req.SearchQuery,
		Quantity:    req.Quantity,
		Price:       req.Price,
		Rating:      req.Rating,
		Metadata:    req.Metadata,
		UserAgent:   c.GetHeader("User-Agent"),
		IPAddress:   c.ClientIP(),
		Referrer:    c.GetHeader("Referer"),
	}

	if err := h.eventService.LogEvent(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log event"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Event logged successfully",
		"event_id":   event.ID,
		"session_id": sessionID,
	})
}

// LogProductView handles POST /api/events/product-view
func (h *EventHandler) LogProductView(c *gin.Context) {
	var req struct {
		ProductID  uint  `json:"product_id" binding:"required"`
		CategoryID *uint `json:"category_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID *uint
	if id, exists := c.Get("user_id"); exists {
		uid := id.(uint)
		userID = &uid
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	if err := h.eventService.LogProductView(userID, sessionID, req.ProductID, req.CategoryID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log event"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Product view logged", "session_id": sessionID})
}

// LogAddToCart handles POST /api/events/add-to-cart
func (h *EventHandler) LogAddToCart(c *gin.Context) {
	var req struct {
		ProductID uint    `json:"product_id" binding:"required"`
		Quantity  int     `json:"quantity" binding:"required"`
		Price     float64 `json:"price"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID *uint
	if id, exists := c.Get("user_id"); exists {
		uid := id.(uint)
		userID = &uid
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	if err := h.eventService.LogAddToCart(userID, sessionID, req.ProductID, req.Quantity, req.Price); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log event"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Add to cart logged", "session_id": sessionID})
}

// LogSearch handles POST /api/events/search
func (h *EventHandler) LogSearch(c *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID *uint
	if id, exists := c.Get("user_id"); exists {
		uid := id.(uint)
		userID = &uid
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	if err := h.eventService.LogSearch(userID, sessionID, req.Query); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log event"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Search logged", "session_id": sessionID})
}

// GetUserBehavior handles GET /api/events/user/:id/behavior
func (h *EventHandler) GetUserBehavior(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	profile, err := h.eventService.GetUserBehaviorProfile(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user behavior"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// GetProductStats handles GET /api/events/product/:id/stats
func (h *EventHandler) GetProductStats(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	stats, err := h.eventService.GetProductEventStats(uint(productID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetEventStats handles GET /api/admin/events/stats
func (h *EventHandler) GetEventStats(c *gin.Context) {
	days := 30 // Default last 30 days
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			days = parsed
		}
	}

	stats, err := h.eventService.GetEventStats(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get event stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats, "period_days": days})
}

// GetPopularProducts handles GET /api/events/popular-products
func (h *EventHandler) GetPopularProducts(c *gin.Context) {
	limit := 10
	days := 7

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			days = parsed
		}
	}

	productIDs, err := h.eventService.GetPopularProducts(limit, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get popular products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product_ids": productIDs})
}
