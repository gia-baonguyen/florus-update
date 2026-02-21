package middleware

import (
	"strings"

	"github.com/florus/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserIDKey           = "userID"
	UserEmailKey        = "userEmail"
	UserRoleKey         = "userRole"
)

func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			utils.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			utils.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)

		claims, err := utils.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			if err == utils.ErrExpiredToken {
				utils.Unauthorized(c, "Token has expired")
			} else {
				utils.Unauthorized(c, "Invalid token")
			}
			c.Abort()
			return
		}

		// Set user info in context
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Set(UserRoleKey, claims.Role)

		c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0
	}
	return userID.(uint)
}

// GetUserRole extracts user role from context
func GetUserRole(c *gin.Context) string {
	role, exists := c.Get(UserRoleKey)
	if !exists {
		return ""
	}
	return role.(string)
}
