package customers

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

type CreateCustomerRequest struct {
	Name    string `json:"name" binding:"required"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type UpdateCustomerRequest struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	IsActive *bool  `json:"is_active"`
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg config.Config) {
	h := Handler{DB: db}

	customers := rg.Group("/customers")
	customers.Use(middleware.AuthRequired(db, cfg.JWTSecret))
	customers.Use(middleware.ShopRequired(db))

	customers.GET("", h.List)
	customers.POST("", h.Create)
	customers.GET("/:id", h.Get)
	customers.PUT("/:id", h.Update)
	customers.DELETE("/:id", h.Delete)
}

func (h Handler) List(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)
	search := c.Query("search")

	var customers []models.Customer
	query := h.DB.Where("shop_id = ? AND is_active = ?", shopID, true)

	if search != "" {
		query = query.Where("name ILIKE ? OR phone ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Order("created_at desc").Find(&customers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load customers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": customers})
}

func (h Handler) Create(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)

	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	customer := models.Customer{
		ShopID:   shopID,
		Name:     req.Name,
		Phone:    req.Phone,
		Address:  req.Address,
		IsActive: true,
	}

	if err := h.DB.Create(&customer).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to create customer"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": customer})
}

func (h Handler) Get(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)
	id := c.Param("id")

	var customer models.Customer
	if err := h.DB.Where("id = ? AND shop_id = ?", id, shopID).First(&customer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "customer not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": customer})
}

func (h Handler) Update(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)
	id := c.Param("id")

	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var customer models.Customer
	if err := h.DB.Where("id = ? AND shop_id = ?", id, shopID).First(&customer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "customer not found"})
		return
	}

	if req.Name != "" {
		customer.Name = req.Name
	}
	if req.Phone != "" {
		customer.Phone = req.Phone
	}
	if req.Address != "" {
		customer.Address = req.Address
	}
	if req.IsActive != nil {
		customer.IsActive = *req.IsActive
	}

	if err := h.DB.Save(&customer).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to update customer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": customer})
}

func (h Handler) Delete(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)
	id := c.Param("id")

	var customer models.Customer
	if err := h.DB.Where("id = ? AND shop_id = ?", id, shopID).First(&customer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "customer not found"})
		return
	}

	customer.IsActive = false
	if err := h.DB.Save(&customer).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to delete customer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "customer deleted successfully"})
}
