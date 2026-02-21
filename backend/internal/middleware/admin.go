package middleware

import (
	"github.com/florus/backend/internal/models"
	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetUserRole(c)
		if role != string(models.RoleAdmin) {
			utils.Forbidden(c, "Admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}
