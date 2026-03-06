package handler

import (
	"strconv"

	"github.com/florus/backend/internal/service"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type GraphRecommendationHandler struct {
	graphService service.GraphRecommendationService
}

func NewGraphRecommendationHandler(graphService service.GraphRecommendationService) *GraphRecommendationHandler {
	return &GraphRecommendationHandler{graphService: graphService}
}

// GetGraphSimilar godoc
// @Summary Get similar products using graph-based relationships
// @Description Returns products connected via SIMILAR_TO relationships in Neo4j
// @Tags Graph Recommendations
// @Produce json
// @Param productId path int true "Product ID"
// @Param limit query int false "Limit results" default(10)
// @Success 200 {array} dto.ProductResponse
// @Router /api/recommendations/graph/similar/{productId} [get]
func (h *GraphRecommendationHandler) GetGraphSimilar(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid product ID")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	products, err := h.graphService.GetGraphSimilarProducts(c.Request.Context(), uint(productID), limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Graph-based similar products retrieved", products)
}

// GetFrequentlyBoughtTogether godoc
// @Summary Get frequently bought together products
// @Description Returns products connected via FREQUENTLY_BOUGHT_WITH relationships
// @Tags Graph Recommendations
// @Produce json
// @Param productId path int true "Product ID"
// @Param limit query int false "Limit results" default(6)
// @Success 200 {array} dto.ProductResponse
// @Router /api/recommendations/graph/bought-together/{productId} [get]
func (h *GraphRecommendationHandler) GetFrequentlyBoughtTogether(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid product ID")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6"))

	products, err := h.graphService.GetFrequentlyBoughtTogether(c.Request.Context(), uint(productID), limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Frequently bought together products retrieved", products)
}

// GetUserBased godoc
// @Summary Get user-based collaborative filtering recommendations
// @Description Returns products based on similar user behavior patterns in the graph
// @Tags Graph Recommendations
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit results" default(20)
// @Success 200 {array} dto.ProductResponse
// @Router /api/recommendations/graph/user-based [get]
func (h *GraphRecommendationHandler) GetUserBased(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User authentication required")
		return
	}

	// Convert userID to uint
	var uid uint
	switch v := userID.(type) {
	case uint:
		uid = v
	case float64:
		uid = uint(v)
	case int:
		uid = uint(v)
	default:
		utils.BadRequest(c, "Invalid user ID format")
		return
	}

	products, err := h.graphService.GetUserBasedRecommendations(c.Request.Context(), uid, limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "User-based graph recommendations retrieved", products)
}

// GetCategoryBased godoc
// @Summary Get popular products in a category
// @Description Returns popular products in the same category from the graph
// @Tags Graph Recommendations
// @Produce json
// @Param categoryId path int true "Category ID"
// @Param excludeProductId query int false "Product ID to exclude" default(0)
// @Param limit query int false "Limit results" default(10)
// @Success 200 {array} dto.ProductResponse
// @Router /api/recommendations/graph/category/{categoryId} [get]
func (h *GraphRecommendationHandler) GetCategoryBased(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("categoryId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid category ID")
		return
	}

	excludeProductID, _ := strconv.ParseUint(c.DefaultQuery("excludeProductId", "0"), 10, 32)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	products, err := h.graphService.GetCategoryRecommendations(c.Request.Context(), uint(categoryID), uint(excludeProductID), limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Category-based graph recommendations retrieved", products)
}

// GetSerendipity godoc
// @Summary Get serendipity recommendations for discovery
// @Description Returns diverse recommendations from unexplored categories
// @Tags Graph Recommendations
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit results" default(10)
// @Success 200 {array} dto.ProductResponse
// @Router /api/recommendations/graph/serendipity [get]
func (h *GraphRecommendationHandler) GetSerendipity(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User authentication required")
		return
	}

	// Convert userID to uint
	var uid uint
	switch v := userID.(type) {
	case uint:
		uid = v
	case float64:
		uid = uint(v)
	case int:
		uid = uint(v)
	default:
		utils.BadRequest(c, "Invalid user ID format")
		return
	}

	products, err := h.graphService.GetSerendipityRecommendations(c.Request.Context(), uid, limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Serendipity recommendations retrieved", products)
}

// GetGraphStatus godoc
// @Summary Check Neo4j graph database status
// @Description Returns whether the graph recommendation system is available
// @Tags Graph Recommendations
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/recommendations/graph/status [get]
func (h *GraphRecommendationHandler) GetGraphStatus(c *gin.Context) {
	available := h.graphService.IsAvailable()

	utils.OK(c, "Graph status retrieved", gin.H{
		"neo4j_available": available,
		"service":         "graph_recommendation",
	})
}
