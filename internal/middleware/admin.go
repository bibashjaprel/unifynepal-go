package middleware

import (
	"net/http"

	"github.com/bibashjaprel/unifynepal-api/internal/models"
	"github.com/gin-gonic/gin"
)

func PlatformAdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userValue, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
			return
		}

		user := userValue.(models.User)

		if user.Role != "platform_admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "platform admin access required"})
			return
		}

		c.Next()
	}
}
