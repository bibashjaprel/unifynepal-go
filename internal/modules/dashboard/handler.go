package dashboard

import (
	"net/http"

	"github.com/bibashjaprel/unifynepal-api/internal/config"
	"github.com/bibashjaprel/unifynepal-api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

type StatsResponse struct {
	TodaySales     float64 `json:"today_sales"`
	MonthSales     float64 `json:"month_sales"`
	TotalDue       float64 `json:"total_due"`
	TotalProducts  int64   `json:"total_products"`
	LowStockCount  int64   `json:"low_stock_count"`
	TotalCustomers int64   `json:"total_customers"`
	TotalBills     int64   `json:"total_bills"`
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg config.Config) {
	h := Handler{DB: db}

	group := rg.Group("/dashboard")
	group.Use(middleware.AuthRequired(db, cfg.JWTSecret))
	group.Use(middleware.ShopRequired(db))

	group.GET("/stats", h.Stats)
}

func (h Handler) Stats(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)

	var stats StatsResponse

	if err := h.DB.Raw(`
		SELECT COALESCE(SUM(total_amount), 0)
		FROM bills
		WHERE shop_id = ?
		AND DATE(created_at) = CURRENT_DATE
	`, shopID).Scan(&stats.TodaySales).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load today sales"})
		return
	}

	if err := h.DB.Raw(`
		SELECT COALESCE(SUM(total_amount), 0)
		FROM bills
		WHERE shop_id = ?
		AND DATE_TRUNC('month', created_at) = DATE_TRUNC('month', CURRENT_DATE)
	`, shopID).Scan(&stats.MonthSales).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load month sales"})
		return
	}

	if err := h.DB.Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN type = 'credit' THEN amount ELSE 0 END), 0)
			-
			COALESCE(SUM(CASE WHEN type = 'payment' THEN amount ELSE 0 END), 0)
		FROM customer_ledger_entries
		WHERE shop_id = ?
	`, shopID).Scan(&stats.TotalDue).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load total due"})
		return
	}

	if err := h.DB.Raw(`
		SELECT COUNT(*)
		FROM products
		WHERE shop_id = ?
		AND is_active = true
	`, shopID).Scan(&stats.TotalProducts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to count products"})
		return
	}

	if err := h.DB.Raw(`
		SELECT COUNT(*)
		FROM products
		WHERE shop_id = ?
		AND is_active = true
		AND stock_qty <= min_stock_qty
	`, shopID).Scan(&stats.LowStockCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to count low stock products"})
		return
	}

	if err := h.DB.Raw(`
		SELECT COUNT(*)
		FROM customers
		WHERE shop_id = ?
		AND is_active = true
	`, shopID).Scan(&stats.TotalCustomers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to count customers"})
		return
	}

	if err := h.DB.Raw(`
		SELECT COUNT(*)
		FROM bills
		WHERE shop_id = ?
	`, shopID).Scan(&stats.TotalBills).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to count bills"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}
