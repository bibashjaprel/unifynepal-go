package admin

import (
	"net/http"

	"github.com/bibashjaprel/unifynepal-api/internal/config"
	"github.com/bibashjaprel/unifynepal-api/internal/middleware"
	"github.com/bibashjaprel/unifynepal-api/internal/models"
	"github.com/bibashjaprel/unifynepal-api/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

type UpdateShopStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg config.Config) {
	h := Handler{DB: db}

	group := rg.Group("/admin")
	group.Use(middleware.AuthRequired(db, cfg.JWTSecret))
	group.Use(middleware.PlatformAdminRequired())

	group.GET("/shops", h.ListShops)
	group.GET("/users", h.ListUsers)
	group.PATCH("/shops/:id/status", h.UpdateShopStatus)
}

func (h Handler) ListShops(c *gin.Context) {
	var shops []models.Shop

	if err := h.DB.Order("created_at desc").Find(&shops).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load shops"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": shops})
}

func (h Handler) ListUsers(c *gin.Context) {
	var users []models.User

	if err := h.DB.Order("created_at desc").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h Handler) UpdateShopStatus(c *gin.Context) {
	id := c.Param("id")

	userValue, _ := c.Get("user")
	user := userValue.(models.User)

	var req UpdateShopStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if req.Status != "active" && req.Status != "paused" && req.Status != "suspended" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid status"})
		return
	}

	var shop models.Shop
	if err := h.DB.Where("id = ?", id).First(&shop).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "shop not found"})
		return
	}

	oldStatus := shop.Status
	shop.Status = req.Status

	if err := h.DB.Save(&shop).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update shop status"})
		return
	}

	utils.CreateAuditLog(h.DB, utils.AuditInput{
		ShopID:     &shop.ID,
		UserID:     &user.ID,
		Action:     "shop_status_updated",
		EntityType: "shop",
		EntityID:   &shop.ID,
		Metadata: map[string]interface{}{
			"old_status": oldStatus,
			"new_status": req.Status,
		},
	})

	c.JSON(http.StatusOK, gin.H{"data": shop})
}
