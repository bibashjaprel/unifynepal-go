package middleware

import (
	"log"
	"time"

	"github.com/bibashjaprel/unifynepal-api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)

		requestID, _ := c.Get("request_id")

		userID := ""
		if userValue, exists := c.Get("user"); exists {
			if user, ok := userValue.(models.User); ok {
				userID = user.ID.String()
			}
		}

		shopID := ""
		if shopValue, exists := c.Get("shop_id"); exists {
			if id, ok := shopValue.(uuid.UUID); ok {
				shopID = id.String()
			}
		}

		log.Printf(
			`{"level":"info","request_id":"%v","method":"%s","path":"%s","status":%d,"latency_ms":%d,"client_ip":"%s","user_agent":"%s","user_id":"%s","shop_id":"%s"}`,
			requestID,
			c.Request.Method,
			c.FullPath(),
			c.Writer.Status(),
			latency.Milliseconds(),
			c.ClientIP(),
			c.GetHeader("User-Agent"),
			userID,
			shopID,
		)
	}
}
