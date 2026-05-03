package audit

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

	group := rg.Group("/audit")
	group.Use(middleware.AuthRequired(db, cfg.JWTSecret))
	group.Use(middleware.ShopRequired(db))

	group.GET("/logs", h.List)
}

func (h Handler) List(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)

	var logs []models.AuditLog
	if err := h.DB.
		Where("shop_id = ? OR shop_id IS NULL", shopID).
		Order("created_at desc").
		Limit(100).
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}
