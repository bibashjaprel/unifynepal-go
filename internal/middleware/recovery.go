package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				requestID, _ := c.Get("request_id")

				log.Printf(
					`{"level":"error","request_id":"%v","panic":"%v","stack":"%s"}`,
					requestID,
					r,
					string(debug.Stack()),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success":    false,
					"message":    "internal server error",
					"request_id": requestID,
				})
			}
		}()

		c.Next()
	}
}
