package handler

import (
	"strconv"

	"github.com/florus/backend/internal/service"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type RecommendationHandler struct {
	recommendationService service.RecommendationService
}

func NewRecommendationHandler(recommendationService service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{recommendationService: recommendationService}
}

// GetColdStart godoc
// @Summary Get cold-start recommendations (popular products for new users)
// @Tags Recommendations
// @Produce json
// @Param limit query int false "Limit results" default(6)
// @Success 200 {array} dto.ProductResponse
// @Router /api/recommendations/cold-start [get]
func (h *RecommendationHandler) GetColdStart(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6"))

	products, err := h.recommendationService.GetColdStartRecommendations(limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Cold-start recommendations retrieved", products)
}

// GetWarmStart godoc
// @Summary Get warm-start recommendations (personalized for active users)
// @Tags Recommendations
// @Produce json
// @Param limit query int false "Limit results" default(4)
// @Success 200 {array} dto.ProductResponse
// @Router /api/recommendations/warm-start [get]
func (h *RecommendationHandler) GetWarmStart(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "4"))

	products, err := h.recommendationService.GetWarmStartRecommendations(limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Warm-start recommendations retrieved", products)
}

// GetSimilar godoc
// @Summary Get similar products (content-based filtering)
// @Tags Recommendations
// @Produce json
// @Param productId path int true "Product ID"
// @Param limit query int false "Limit results" default(4)
// @Success 200 {array} dto.ProductResponse
// @Router /api/recommendations/similar/{productId} [get]
func (h *RecommendationHandler) GetSimilar(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid product ID")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "4"))

	products, err := h.recommendationService.GetSimilarProducts(uint(productID), limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Similar products retrieved", products)
}

// GetCoViewed godoc
// @Summary Get co-viewed products (collaborative filtering)
// @Tags Recommendations
// @Produce json
// @Param productId path int true "Product ID"
// @Param limit query int false "Limit results" default(4)
// @Success 200 {array} dto.ProductResponse
// @Router /api/recommendations/co-viewed/{productId} [get]
func (h *RecommendationHandler) GetCoViewed(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid product ID")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "4"))

	products, err := h.recommendationService.GetCoViewedProducts(uint(productID), limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Co-viewed products retrieved", products)
}

// GetCrossSell godoc
// @Summary Get cross-sell products (frequently bought together)
// @Tags Recommendations
// @Produce json
// @Param productId path int true "Product ID"
// @Param limit query int false "Limit results" default(4)
// @Success 200 {array} dto.ProductResponse
// @Router /api/recommendations/cross-sell/{productId} [get]
func (h *RecommendationHandler) GetCrossSell(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid product ID")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "4"))

	products, err := h.recommendationService.GetCrossSellProducts(uint(productID), limit)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.OK(c, "Cross-sell products retrieved", products)
}
