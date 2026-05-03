package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/bibashjaprel/unifynepal-api/internal/models"
	"github.com/bibashjaprel/unifynepal-api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func AuthRequired(db *gorm.DB, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "missing token"})
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")

		claims := &utils.JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
			return
		}

		var session models.UserSession
		if err := db.Where(
			"token_id = ? AND user_id = ? AND is_active = ? AND revoked_at IS NULL",
			claims.TokenID,
			claims.UserID,
			true,
		).First(&session).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "session expired or revoked"})
			return
		}

		if time.Now().After(session.ExpiresAt) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "session expired"})
			return
		}

		var user models.User
		if err := db.First(&user, "id = ?", claims.UserID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "user not found"})
			return
		}

		db.Model(&session).Update("last_seen_at", time.Now())

		c.Set("user", user)
		c.Set("session", session)
		c.Next()
	}
}
