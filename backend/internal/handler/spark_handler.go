package handler

import (
	"strconv"

	"github.com/florus/backend/internal/service"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type SparkHandler struct {
	sparkService service.SparkRecommendationService
}

func NewSparkHandler(sparkService service.SparkRecommendationService) *SparkHandler {
	return &SparkHandler{sparkService: sparkService}
}

// GetSparkRecommendations godoc
// @Summary Get ML-based recommendations from Spark ALS model
// @Tags Recommendations
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit results" default(20)
// @Success 200 {array} dto.ProductResponse
// @Router /api/recommendations/spark/user [get]
func (h *SparkHandler) GetSparkRecommendations(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID required for Spark recommendations")
		return
	}

	var uid uint
	switch v := userID.(type) {
	case uint:
		uid = v
	case float64:
		uid = uint(v)
	case int:
		uid = uint(v)
	default:
		utils.BadRequest(c, "Invalid user ID")
		return
	}

	products, err := h.sparkService.GetSparkUserRecommendations(uid, limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Spark recommendations retrieved", products)
}

// GetSparkSimilarProducts godoc
// @Summary Get ML-based similar products from Spark ALS model
// @Tags Recommendations
// @Produce json
// @Param productId path int true "Product ID"
// @Param limit query int false "Limit results" default(10)
// @Success 200 {array} dto.ProductResponse
// @Router /api/recommendations/spark/similar/{productId} [get]
func (h *SparkHandler) GetSparkSim
ilarProducts(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid product ID")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	products, err := h.sparkService.GetSparkSimilarProducts(uint(productID), limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Spark similar products retrieved", products)
}

// GetSparkStatus godoc
// @Summary Get Spark recommendations status
// @Tags Recommendations
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/recommendations/spark/status [get]
func (h *SparkHandler) GetSparkStatus(c *gin.Context) {
	available := h.sparkService.IsAvailable()

	response := map[string]interface{}{
		"available": available,
	}

	if available {
		lastUpdate, err := h.sparkService.GetLastUpdateTime()
		if err == nil {
			response["last_updated"] = lastUpdate
		}
	}

	utils.OK(c, "Spark status retrieved", response)
}
