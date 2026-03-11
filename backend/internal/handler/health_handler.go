package handler

import (
	"net/http"

	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	utils.OK(c, "OK", gin.H{
		"status":  "healthy",
		"service": "florus-api",
		"version": "1.1.0",
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
