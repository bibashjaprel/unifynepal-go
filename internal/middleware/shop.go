package middleware

import (
	"net/http"

	"github.com/bibashjaprel/unifynepal-api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ShopRequired(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		shopIDHeader := c.GetHeader("X-Shop-ID")
		if shopIDHeader == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "missing X-Shop-ID header"})
			return
		}

		shopID, err := uuid.Parse(shopIDHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid shop id"})
			return
		}

		userValue, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
			return
		}

		user := userValue.(models.User)

		var member models.ShopMember
		if err := db.Where("shop_id = ? AND user_id = ?", shopID, user.ID).First(&member).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "you do not have access to this shop"})
			return
		}

		c.Set("shop_id", shopID)
		c.Set("shop_role", member.Role)
		c.Next()
	}
}
