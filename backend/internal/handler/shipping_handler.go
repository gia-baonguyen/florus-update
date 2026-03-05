package handler

import (
	"github.com/florus/backend/internal/service"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type ShippingHandler struct {
	shippingService service.ShippingService
}

func NewShippingHandler(shippingService service.ShippingService) *ShippingHandler {
	return &ShippingHandler{shippingService: shippingService}
}

func (h *ShippingHandler) GetMethods(c *gin.Context) {
	methods, err := h.shippingService.GetShippingMethods()
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}
	utils.OK(c, "Shipping methods retrieved successfully", methods)
}

