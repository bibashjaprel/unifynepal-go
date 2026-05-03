package routes

import (
	"net/http"

	"github.com/bibashjaprel/unifynepal-api/internal/config"
	"github.com/bibashjaprel/unifynepal-api/internal/modules/admin"
	"github.com/bibashjaprel/unifynepal-api/internal/modules/audit"
	"github.com/bibashjaprel/unifynepal-api/internal/modules/auth"
	"github.com/bibashjaprel/unifynepal-api/internal/modules/billing"
	"github.com/bibashjaprel/unifynepal-api/internal/modules/customers"
	"github.com/bibashjaprel/unifynepal-api/internal/modules/dashboard"
	"github.com/bibashjaprel/unifynepal-api/internal/modules/inventory"
	"github.com/bibashjaprel/unifynepal-api/internal/modules/notifications"
	"github.com/bibashjaprel/unifynepal-api/internal/modules/products"
	"github.com/bibashjaprel/unifynepal-api/internal/modules/subscriptions"
	"github.com/bibashjaprel/unifynepal-api/internal/modules/udharo"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Register(r *gin.Engine, db *gorm.DB, cfg config.Config) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":    cfg.AppName,
			"status":  "running",
			"version": "1.0.0",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")

	auth.RegisterRoutes(api, db, cfg)
	products.RegisterRoutes(api, db, cfg)
	customers.RegisterRoutes(api, db, cfg)
	billing.RegisterRoutes(api, db, cfg)
	udharo.RegisterRoutes(api, db, cfg)
	dashboard.RegisterRoutes(api, db, cfg)
	inventory.RegisterRoutes(api, db, cfg)
	audit.RegisterRoutes(api, db, cfg)
	subscriptions.RegisterRoutes(api, db, cfg)
	admin.RegisterRoutes(api, db, cfg)
	notifications.RegisterRoutes(api, db, cfg)
}
