package inventory

import (
	"fmt"
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

type CreateStockMovementRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Type      string    `json:"type" binding:"required"` // stock_in, stock_out, adjustment
	Quantity  float64   `json:"quantity" binding:"required"`
	Note      string    `json:"note"`
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg config.Config) {
	h := Handler{DB: db}

	group := rg.Group("")
	group.Use(middleware.AuthRequired(db, cfg.JWTSecret))
	group.Use(middleware.ShopRequired(db))

	group.GET("/stock-movements", h.ListStockMovements)
	group.POST("/stock-movements", h.CreateStockMovement)
	group.GET("/inventory/low-stock", h.LowStock)
}

func (h Handler) ListStockMovements(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)

	var movements []models.StockMovement
	if err := h.DB.
		Where("shop_id = ?", shopID).
		Order("created_at desc").
		Find(&movements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load stock movements"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movements})
}

func (h Handler) LowStock(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)

	var products []models.Product
	if err := h.DB.
		Where("shop_id = ? AND is_active = ? AND stock_qty <= min_stock_qty", shopID, true).
		Order("stock_qty asc").
		Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load low stock products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": products})
}

func (h Handler) CreateStockMovement(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)

	var req CreateStockMovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "quantity must be greater than zero"})
		return
	}

	if req.Type != "stock_in" && req.Type != "stock_out" && req.Type != "adjustment" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid movement type"})
		return
	}

	var createdMovement models.StockMovement

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var product models.Product
		if err := tx.Where("id = ? AND shop_id = ? AND is_active = ?", req.ProductID, shopID, true).First(&product).Error; err != nil {
			return fmt.Errorf("product not found")
		}

		previousQty := product.StockQty
		newQty := previousQty

		switch req.Type {
		case "stock_in":
			newQty = previousQty + req.Quantity
		case "stock_out":
			if previousQty < req.Quantity {
				return fmt.Errorf("insufficient stock")
			}
			newQty = previousQty - req.Quantity
		case "adjustment":
			newQty = req.Quantity
		}

		if err := tx.Model(&product).Update("stock_qty", newQty).Error; err != nil {
			return err
		}

		movementQty := req.Quantity
		if req.Type == "adjustment" {
			movementQty = newQty - previousQty
			if movementQty < 0 {
				movementQty = -movementQty
			}
		}

		createdMovement = models.StockMovement{
			ShopID:        shopID,
			ProductID:     product.ID,
			Type:          req.Type,
			Quantity:      movementQty,
			PreviousQty:   previousQty,
			NewQty:        newQty,
			ReferenceType: "manual",
			ReferenceID:   nil,
			Note:          req.Note,
		}

		if createdMovement.Note == "" {
			createdMovement.Note = "Manual stock movement"
		}

		if err := tx.Create(&createdMovement).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": createdMovement})
}
