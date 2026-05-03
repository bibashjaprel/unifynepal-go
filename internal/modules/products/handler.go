package products

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

type CreateProductRequest struct {
	Name         string  `json:"name" binding:"required"`
	SKU          string  `json:"sku"`
	Category     string  `json:"category"`
	Unit         string  `json:"unit"`
	SellingPrice float64 `json:"selling_price" binding:"required"`
	CostPrice    float64 `json:"cost_price"`
	StockQty     float64 `json:"stock_qty"`
	MinStockQty  float64 `json:"min_stock_qty"`
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg config.Config) {
	h := Handler{DB: db}

	products := rg.Group("/products")
	products.Use(middleware.AuthRequired(db, cfg.JWTSecret))
	products.Use(middleware.ShopRequired(db))

	products.GET("", h.List)
	products.POST("", h.Create)
}

func (h Handler) List(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)

	var products []models.Product
	if err := h.DB.Where("shop_id = ? AND is_active = ?", shopID, true).Order("created_at desc").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": products})
}

func (h Handler) Create(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	product := models.Product{
		ShopID:       shopID,
		Name:         req.Name,
		SKU:          req.SKU,
		Category:     req.Category,
		Unit:         req.Unit,
		SellingPrice: req.SellingPrice,
		CostPrice:    req.CostPrice,
		StockQty:     req.StockQty,
		MinStockQty:  req.MinStockQty,
		IsActive:     true,
	}

	if product.Unit == "" {
		product.Unit = "piece"
	}

	if err := h.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": product})
}
