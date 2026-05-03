package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, status int, message string, data interface{}) {
	if status == 0 {
		status = http.StatusOK
	}

	c.JSON(status, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func Error(c *gin.Context, status int, message string, errors interface{}) {
	if status == 0 {
		status = http.StatusInternalServerError
	}

	requestID, _ := c.Get("request_id")

	c.JSON(status, gin.H{
		"success":    false,
		"message":    message,
		"errors":     errors,
		"request_id": requestID,
	})
}
