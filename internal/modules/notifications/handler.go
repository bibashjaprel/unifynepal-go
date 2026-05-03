package notifications

import (
	"net/http"

	"github.com/bibashjaprel/unifynepal-api/internal/config"
	"github.com/bibashjaprel/unifynepal-api/internal/middleware"
	"github.com/bibashjaprel/unifynepal-api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg config.Config) {
	h := Handler{DB: db}

	group := rg.Group("/notifications")
	group.Use(middleware.AuthRequired(db, cfg.JWTSecret))

	group.GET("", h.List)
	group.POST("/:id/read", h.MarkRead)
	group.POST("/read-all", h.MarkAllRead)
}

func (h Handler) List(c *gin.Context) {
	userValue, _ := c.Get("user")
	user := userValue.(models.User)

	shopIDHeader := c.GetHeader("X-Shop-ID")

	var notifications []models.Notification
	query := h.DB.
		Where("user_id = ?", user.ID).
		Order("created_at desc").
		Limit(100)

	if shopIDHeader != "" {
		shopID, err := uuid.Parse(shopIDHeader)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid shop id"})
			return
		}

		var member models.ShopMember
		if err := h.DB.Where("shop_id = ? AND user_id = ?", shopID, user.ID).First(&member).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"message": "you do not have access to this shop"})
			return
		}

		query = query.Where("shop_id = ? OR shop_id IS NULL", shopID)
	}

	if err := query.Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load notifications"})
		return
	}

	var unreadCount int64
	countQuery := h.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", user.ID, false)

	if shopIDHeader != "" {
		shopID, _ := uuid.Parse(shopIDHeader)
		countQuery = countQuery.Where("shop_id = ? OR shop_id IS NULL", shopID)
	}

	_ = countQuery.Count(&unreadCount).Error

	c.JSON(http.StatusOK, gin.H{
		"data": notifications,
		"meta": gin.H{
			"unread_count": unreadCount,
		},
	})
}

func (h Handler) MarkRead(c *gin.Context) {
	userValue, _ := c.Get("user")
	user := userValue.(models.User)

	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid notification id"})
		return
	}

	var notification models.Notification
	if err := h.DB.
		Where("id = ? AND user_id = ?", notificationID, user.ID).
		First(&notification).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "notification not found"})
		return
	}

	if err := h.DB.Model(&notification).Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to mark notification as read"})
		return
	}

	notification.IsRead = true

	c.JSON(http.StatusOK, gin.H{
		"message": "notification marked as read",
		"data":    notification,
	})
}

func (h Handler) MarkAllRead(c *gin.Context) {
	userValue, _ := c.Get("user")
	user := userValue.(models.User)

	shopIDHeader := c.GetHeader("X-Shop-ID")

	query := h.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", user.ID, false)

	if shopIDHeader != "" {
		shopID, err := uuid.Parse(shopIDHeader)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid shop id"})
			return
		}

		var member models.ShopMember
		if err := h.DB.Where("shop_id = ? AND user_id = ?", shopID, user.ID).First(&member).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"message": "you do not have access to this shop"})
			return
		}

		query = query.Where("shop_id = ? OR shop_id IS NULL", shopID)
	}

	if err := query.Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "all notifications marked as read",
	})
}
