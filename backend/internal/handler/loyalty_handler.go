package handler

import (
	"github.com/florus/backend/internal/middleware"
	"github.com/florus/backend/internal/service"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type LoyaltyHandler struct {
	loyaltyService service.LoyaltyService
}

func NewLoyaltyHandler(loyaltyService service.LoyaltyService) *LoyaltyHandler {
	return &LoyaltyHandler{loyaltyService: loyaltyService}
}

// Get current user's loyalty summary (tier + points)
func (h *LoyaltyHandler) GetLoyalty(c *gin.Context) {
	userID := middleware.GetUserID(c)
	tier, points, err := h.loyaltyService.GetUserLoyalty(userID)
	if err != nil {
		// If loyalty columns or data are not ready yet, gracefully return defaults
		utils.OK(c, "Loyalty info not available, returning defaults", gin.H{
			"tier":   "Silver",
			"points": 0,
		})
		return
	}
	utils.OK(c, "Loyalty info retrieved successfully", gin.H{
		"tier":   tier,
		"points": points,
	})
}

