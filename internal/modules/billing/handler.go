package billing

import (
	"fmt"
	"math"
	"net/http"
	"time"

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

type CreateBillItemRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  float64   `json:"quantity" binding:"required"`
	UnitPrice float64   `json:"unit_price"`
}

type CreateBillRequest struct {
	CustomerID    *uuid.UUID              `json:"customer_id"`
	Discount      float64                 `json:"discount"`
	PaidAmount    float64                 `json:"paid_amount"`
	PaymentMethod string                  `json:"payment_method"`
	Items         []CreateBillItemRequest `json:"items" binding:"required,min=1"`
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg config.Config) {
	h := Handler{DB: db}

	bills := rg.Group("/bills")
	bills.Use(middleware.AuthRequired(db, cfg.JWTSecret))
	bills.Use(middleware.ShopRequired(db))

	bills.GET("", h.List)
	bills.POST("", h.Create)
	bills.GET("/:id", h.Get)
}

func (h Handler) List(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)

	var bills []models.Bill
	if err := h.DB.
		Preload("Items").
		Where("shop_id = ?", shopID).
		Order("created_at desc").
		Find(&bills).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load bills"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": bills})
}

func (h Handler) Get(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)
	id := c.Param("id")

	var bill models.Bill
	if err := h.DB.
		Preload("Items").
		Where("id = ? AND shop_id = ?", id, shopID).
		First(&bill).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "bill not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": bill})
}

func (h Handler) Create(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)

	var req CreateBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if req.Discount < 0 || req.PaidAmount < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "discount and paid_amount cannot be negative"})
		return
	}

	if req.CustomerID == nil && req.PaidAmount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "customer is required for unpaid bills"})
		return
	}

	var createdBill models.Bill

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if req.CustomerID != nil {
			var customer models.Customer
			if err := tx.Where("id = ? AND shop_id = ? AND is_active = ?", *req.CustomerID, shopID, true).First(&customer).Error; err != nil {
				return fmt.Errorf("customer not found")
			}
		}

		var subtotal float64
		var billItems []models.BillItem

		for _, itemReq := range req.Items {
			if itemReq.Quantity <= 0 {
				return fmt.Errorf("quantity must be greater than zero")
			}

			var product models.Product
			if err := tx.Where("id = ? AND shop_id = ? AND is_active = ?", itemReq.ProductID, shopID, true).First(&product).Error; err != nil {
				return fmt.Errorf("product not found")
			}

			if product.StockQty < itemReq.Quantity {
				return fmt.Errorf("insufficient stock for product: %s", product.Name)
			}

			unitPrice := itemReq.UnitPrice
			if unitPrice <= 0 {
				unitPrice = product.SellingPrice
			}

			lineTotal := unitPrice * itemReq.Quantity
			subtotal += lineTotal

			billItems = append(billItems, models.BillItem{
				ProductID:   product.ID,
				ProductName: product.Name,
				Quantity:    itemReq.Quantity,
				UnitPrice:   unitPrice,
				CostPrice:   product.CostPrice,
				Total:       lineTotal,
			})
		}

		if req.Discount > subtotal {
			return fmt.Errorf("discount cannot be greater than subtotal")
		}

		totalAmount := roundMoney(subtotal - req.Discount)
		paidAmount := roundMoney(req.PaidAmount)

		if paidAmount > totalAmount {
			return fmt.Errorf("paid amount cannot be greater than total amount")
		}

		dueAmount := roundMoney(totalAmount - paidAmount)

		paymentStatus := "unpaid"
		if paidAmount >= totalAmount {
			paymentStatus = "paid"
		} else if paidAmount > 0 {
			paymentStatus = "partial"
		}

		billNumber, err := h.generateBillNumber(tx, shopID)
		if err != nil {
			return err
		}

		bill := models.Bill{
			ShopID:        shopID,
			CustomerID:    req.CustomerID,
			BillNumber:    billNumber,
			Subtotal:      roundMoney(subtotal),
			Discount:      roundMoney(req.Discount),
			TotalAmount:   totalAmount,
			PaidAmount:    paidAmount,
			DueAmount:     dueAmount,
			PaymentStatus: paymentStatus,
			PaymentMethod: req.PaymentMethod,
		}

		if bill.PaymentMethod == "" && paidAmount > 0 {
			bill.PaymentMethod = "cash"
		}

		if err := tx.Create(&bill).Error; err != nil {
			return err
		}

		for i := range billItems {
			billItems[i].BillID = bill.ID

			if err := tx.Create(&billItems[i]).Error; err != nil {
				return err
			}

			var product models.Product
			if err := tx.Where("id = ? AND shop_id = ?", billItems[i].ProductID, shopID).First(&product).Error; err != nil {
				return fmt.Errorf("product not found while updating stock")
			}

			previousQty := product.StockQty
			newQty := previousQty - billItems[i].Quantity

			if err := tx.Model(&product).Update("stock_qty", newQty).Error; err != nil {
				return err
			}

			movement := models.StockMovement{
				ShopID:        shopID,
				ProductID:     product.ID,
				Type:          "sale",
				Quantity:      billItems[i].Quantity,
				PreviousQty:   previousQty,
				NewQty:        newQty,
				ReferenceType: "bill",
				ReferenceID:   &bill.ID,
				Note:          "Stock reduced from bill " + bill.BillNumber,
			}

			if err := tx.Create(&movement).Error; err != nil {
				return err
			}
		}

		if paidAmount > 0 {
			payment := models.Payment{
				ShopID:      shopID,
				BillID:      &bill.ID,
				CustomerID:  req.CustomerID,
				Amount:      paidAmount,
				Method:      bill.PaymentMethod,
				ReferenceNo: "",
				Note:        "Payment received for bill " + bill.BillNumber,
			}

			if err := tx.Create(&payment).Error; err != nil {
				return err
			}
		}

		if dueAmount > 0 {
			if req.CustomerID == nil {
				return fmt.Errorf("customer is required when due amount exists")
			}

			ledger := models.CustomerLedgerEntry{
				ShopID:      shopID,
				CustomerID:  *req.CustomerID,
				BillID:      &bill.ID,
				Type:        "credit",
				Amount:      dueAmount,
				Description: "Udharo created from bill " + bill.BillNumber,
			}

			if err := tx.Create(&ledger).Error; err != nil {
				return err
			}
		}

		if err := tx.Preload("Items").First(&createdBill, "id = ?", bill.ID).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": createdBill})
}

func (h Handler) generateBillNumber(tx *gorm.DB, shopID uuid.UUID) (string, error) {
	datePart := time.Now().Format("20060102")

	var count int64
	if err := tx.Model(&models.Bill{}).
		Where("shop_id = ? AND DATE(created_at) = CURRENT_DATE", shopID).
		Count(&count).Error; err != nil {
		return "", err
	}

	return fmt.Sprintf("BILL-%s-%04d", datePart, count+1), nil
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}
