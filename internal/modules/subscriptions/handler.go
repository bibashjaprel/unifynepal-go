package subscriptions

import (
	"net/http"
	"time"

	"github.com/bibashjaprel/unifynepal-api/internal/config"
	"github.com/bibashjaprel/unifynepal-api/internal/middleware"
	"github.com/bibashjaprel/unifynepal-api/internal/models"
	"github.com/bibashjaprel/unifynepal-api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

type UpgradeRequest struct {
	PlanID uuid.UUID `json:"plan_id" binding:"required"`
	Note   string    `json:"note"`
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg config.Config) {
	h := Handler{DB: db}

	group := rg.Group("")
	group.Use(middleware.AuthRequired(db, cfg.JWTSecret))
	group.Use(middleware.ShopRequired(db))

	group.GET("/plans", h.ListPlans)
	group.GET("/subscription", h.CurrentSubscription)
	group.POST("/subscription/upgrade-request", h.UpgradeRequest)
}

func (h Handler) ListPlans(c *gin.Context) {
	var plans []models.SubscriptionPlan

	if err := h.DB.
		Where("is_active = ?", true).
		Order("price_monthly asc").
		Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load plans"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": plans})
}

func (h Handler) CurrentSubscription(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)

	var subscription models.ShopSubscription
	if err := h.DB.
		Where("shop_id = ?", shopID).
		Order("created_at desc").
		First(&subscription).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "subscription not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": subscription})
}

func (h Handler) UpgradeRequest(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)

	userValue, _ := c.Get("user")
	user := userValue.(models.User)

	var req UpgradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var plan models.SubscriptionPlan
	if err := h.DB.Where("id = ? AND is_active = ?", req.PlanID, true).First(&plan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "plan not found"})
		return
	}

	notification := models.Notification{
		ShopID:  &shopID,
		UserID:  user.ID,
		Title:   "Subscription upgrade requested",
		Message: "Upgrade requested for plan: " + plan.Name,
		Type:    "subscription",
		IsRead:  false,
	}

	if err := h.DB.Create(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create upgrade request"})
		return
	}

	utils.CreateAuditLog(h.DB, utils.AuditInput{
		ShopID:     &shopID,
		UserID:     &user.ID,
		Action:     "subscription_upgrade_requested",
		EntityType: "subscription_plan",
		EntityID:   &plan.ID,
		Metadata: map[string]interface{}{
			"plan_name": plan.Name,
			"note":      req.Note,
			"time":      time.Now(),
		},
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "upgrade request submitted",
		"data":    notification,
	})
}
