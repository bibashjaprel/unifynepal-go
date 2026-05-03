package auth

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
	DB  *gorm.DB
	Cfg config.Config
}

type SignupRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	ShopName string `json:"shop_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg config.Config) {
	h := Handler{DB: db, Cfg: cfg}

	auth := rg.Group("/auth")
	auth.POST("/signup", h.Signup)
	auth.POST("/login", h.Login)
	auth.POST("/logout", middleware.AuthRequired(db, cfg.JWTSecret), h.Logout)
}

func (h Handler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to hash password"})
		return
	}

	var user models.User
	var shop models.Shop

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		user = models.User{
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: passwordHash,
			Role:         "user",
			IsActive:     true,
			IsVerified:   true,
		}

		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		shop = models.Shop{
			Name:        req.ShopName,
			Status:      "active",
			OwnerUserID: user.ID.String(),
		}

		if err := tx.Create(&shop).Error; err != nil {
			return err
		}

		member := models.ShopMember{
			ShopID: shop.ID,
			UserID: user.ID,
			Role:   "owner",
		}

		return tx.Create(&member).Error
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	token, tokenID, expiresAt, err := utils.GenerateToken(
		user.ID,
		user.Email,
		h.Cfg.JWTSecret,
		h.Cfg.JWTExpiresInHours,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to generate token"})
		return
	}

	session := models.UserSession{
		UserID:     user.ID,
		ShopID:     &shop.ID,
		TokenID:    tokenID,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		LastSeenAt: time.Now(),
		ExpiresAt:  expiresAt,
		IsActive:   true,
	}

	if err := h.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create session"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "account created",
		"token":   token,
		"user":    user,
		"shop":    shop,
		"session": session,
	})
}

func (h Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid email or password"})
		return
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid email or password"})
		return
	}

	token, tokenID, expiresAt, err := utils.GenerateToken(
		user.ID,
		user.Email,
		h.Cfg.JWTSecret,
		h.Cfg.JWTExpiresInHours,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to generate token"})
		return
	}

	var member models.ShopMember
	var shopIDPtr *uuid.UUID

	if err := h.DB.Where("user_id = ?", user.ID).Order("created_at asc").First(&member).Error; err == nil {
		shopIDPtr = &member.ShopID
	}

	session := models.UserSession{
		UserID:     user.ID,
		ShopID:     shopIDPtr,
		TokenID:    tokenID,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		LastSeenAt: time.Now(),
		ExpiresAt:  expiresAt,
		IsActive:   true,
	}

	if err := h.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"token":   token,
		"user":    user,
		"session": session,
	})
}

func (h Handler) Logout(c *gin.Context) {
	sessionValue, exists := c.Get("session")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "session not found"})
		return
	}

	session := sessionValue.(models.UserSession)
	now := time.Now()

	if err := h.DB.Model(&session).Updates(map[string]interface{}{
		"is_active":  false,
		"revoked_at": now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
