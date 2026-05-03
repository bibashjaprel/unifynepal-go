package udharo

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

type CustomerDueRow struct {
	CustomerID   uuid.UUID `json:"customer_id"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	TotalCredit  float64   `json:"total_credit"`
	TotalPayment float64   `json:"total_payment"`
	CurrentDue   float64   `json:"current_due"`
}

type UdharoSummary struct {
	TotalDue             float64 `json:"total_due"`
	TotalCredit          float64 `json:"total_credit"`
	TotalPayment         float64 `json:"total_payment"`
	CustomersWithDue     int64   `json:"customers_with_due"`
	TotalLedgerCustomers int64   `json:"total_ledger_customers"`
}

type ApplyUdharoPaymentRequest struct {
	Amount      float64    `json:"amount" binding:"required"`
	Method      string     `json:"method"`
	ReferenceNo string     `json:"reference_no"`
	Note        string     `json:"note"`
	BillID      *uuid.UUID `json:"bill_id"`
}

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg config.Config) {
	h := Handler{DB: db}

	group := rg.Group("")
	group.Use(middleware.AuthRequired(db, cfg.JWTSecret))
	group.Use(middleware.ShopRequired(db))

	group.GET("/udharo/summary", h.Summary)
	group.GET("/udharo/customers", h.CustomersWithDue)
	group.GET("/customers/:id/udharo/ledger", h.CustomerLedger)
	group.POST("/customers/:id/udharo/payments", h.ApplyPayment)
}

func (h Handler) Summary(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)

	var summary UdharoSummary

	err := h.DB.Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN type = 'credit' THEN amount ELSE 0 END), 0) AS total_credit,
			COALESCE(SUM(CASE WHEN type = 'payment' THEN amount ELSE 0 END), 0) AS total_payment,
			COALESCE(SUM(CASE WHEN type = 'credit' THEN amount ELSE 0 END), 0)
			-
			COALESCE(SUM(CASE WHEN type = 'payment' THEN amount ELSE 0 END), 0) AS total_due,
			COUNT(DISTINCT customer_id) AS total_ledger_customers
		FROM customer_ledger_entries
		WHERE shop_id = ?
	`, shopID).Scan(&summary).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load udharo summary"})
		return
	}

	var customersWithDue int64
	err = h.DB.Raw(`
		SELECT COUNT(*) FROM (
			SELECT
				customer_id,
				COALESCE(SUM(CASE WHEN type = 'credit' THEN amount ELSE 0 END), 0)
				-
				COALESCE(SUM(CASE WHEN type = 'payment' THEN amount ELSE 0 END), 0) AS current_due
			FROM customer_ledger_entries
			WHERE shop_id = ?
			GROUP BY customer_id
			HAVING
				COALESCE(SUM(CASE WHEN type = 'credit' THEN amount ELSE 0 END), 0)
				-
				COALESCE(SUM(CASE WHEN type = 'payment' THEN amount ELSE 0 END), 0) > 0
		) x
	`, shopID).Scan(&customersWithDue).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to count due customers"})
		return
	}

	summary.CustomersWithDue = customersWithDue

	c.JSON(http.StatusOK, gin.H{"data": summary})
}

func (h Handler) CustomersWithDue(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)

	var rows []CustomerDueRow

	err := h.DB.Raw(`
		SELECT
			c.id AS customer_id,
			c.name,
			c.phone,
			COALESCE(SUM(CASE WHEN cle.type = 'credit' THEN cle.amount ELSE 0 END), 0) AS total_credit,
			COALESCE(SUM(CASE WHEN cle.type = 'payment' THEN cle.amount ELSE 0 END), 0) AS total_payment,
			COALESCE(SUM(CASE WHEN cle.type = 'credit' THEN cle.amount ELSE 0 END), 0)
			-
			COALESCE(SUM(CASE WHEN cle.type = 'payment' THEN cle.amount ELSE 0 END), 0) AS current_due
		FROM customers c
		JOIN customer_ledger_entries cle
			ON cle.customer_id = c.id
			AND cle.shop_id = c.shop_id
		WHERE c.shop_id = ?
		GROUP BY c.id, c.name, c.phone
		HAVING
			COALESCE(SUM(CASE WHEN cle.type = 'credit' THEN cle.amount ELSE 0 END), 0)
			-
			COALESCE(SUM(CASE WHEN cle.type = 'payment' THEN cle.amount ELSE 0 END), 0) > 0
		ORDER BY current_due DESC
	`, shopID).Scan(&rows).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load udharo customers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rows})
}

func (h Handler) CustomerLedger(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)

	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid customer id"})
		return
	}

	var customer models.Customer
	if err := h.DB.Where("id = ? AND shop_id = ? AND is_active = ?", customerID, shopID, true).First(&customer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "customer not found"})
		return
	}

	var entries []models.CustomerLedgerEntry
	if err := h.DB.
		Where("shop_id = ? AND customer_id = ?", shopID, customerID).
		Order("created_at asc").
		Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load customer ledger"})
		return
	}

	var totalCredit float64
	var totalPayment float64

	for _, entry := range entries {
		switch entry.Type {
		case "credit":
			totalCredit += entry.Amount
		case "payment":
			totalPayment += entry.Amount
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"customer":      customer,
			"entries":       entries,
			"total_credit":  totalCredit,
			"total_payment": totalPayment,
			"current_due":   totalCredit - totalPayment,
		},
	})
}

func (h Handler) ApplyPayment(c *gin.Context) {
	shopID := c.MustGet("shop_id").(uuid.UUID)

	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid customer id"})
		return
	}

	var req ApplyUdharoPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "amount must be greater than zero"})
		return
	}

	if req.Method == "" {
		req.Method = "cash"
	}

	var createdPayment models.Payment
	var createdLedger models.CustomerLedgerEntry

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		var customer models.Customer
		if err := tx.Where("id = ? AND shop_id = ? AND is_active = ?", customerID, shopID, true).First(&customer).Error; err != nil {
			return fmt.Errorf("customer not found")
		}

		currentDue, err := h.getCustomerDue(tx, shopID, customerID)
		if err != nil {
			return err
		}

		if currentDue <= 0 {
			return fmt.Errorf("customer has no due amount")
		}

		if req.Amount > currentDue {
			return fmt.Errorf("payment amount cannot be greater than current due")
		}

		if req.BillID != nil {
			var bill models.Bill
			if err := tx.Where("id = ? AND shop_id = ? AND customer_id = ?", *req.BillID, shopID, customerID).First(&bill).Error; err != nil {
				return fmt.Errorf("bill not found for this customer")
			}

			if bill.DueAmount <= 0 {
				return fmt.Errorf("bill has no due amount")
			}

			if req.Amount > bill.DueAmount {
				return fmt.Errorf("payment amount cannot be greater than bill due amount")
			}

			newPaid := bill.PaidAmount + req.Amount
			newDue := bill.DueAmount - req.Amount
			status := "partial"
			if newDue <= 0 {
				newDue = 0
				status = "paid"
			}

			if err := tx.Model(&bill).Updates(map[string]interface{}{
				"paid_amount":    newPaid,
				"due_amount":     newDue,
				"payment_status": status,
				"payment_method": req.Method,
			}).Error; err != nil {
				return err
			}
		}

		createdPayment = models.Payment{
			ShopID:      shopID,
			BillID:      req.BillID,
			CustomerID:  &customerID,
			Amount:      req.Amount,
			Method:      req.Method,
			ReferenceNo: req.ReferenceNo,
			Note:        req.Note,
		}

		if createdPayment.Note == "" {
			createdPayment.Note = "Udharo payment received"
		}

		if err := tx.Create(&createdPayment).Error; err != nil {
			return err
		}

		createdLedger = models.CustomerLedgerEntry{
			ShopID:      shopID,
			CustomerID:  customerID,
			BillID:      req.BillID,
			Type:        "payment",
			Amount:      req.Amount,
			Description: "Udharo payment received",
		}

		if req.Note != "" {
			createdLedger.Description = req.Note
		}

		if err := tx.Create(&createdLedger).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"payment": createdPayment,
			"ledger":  createdLedger,
		},
	})
}

func (h Handler) getCustomerDue(tx *gorm.DB, shopID uuid.UUID, customerID uuid.UUID) (float64, error) {
	var due float64

	err := tx.Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN type = 'credit' THEN amount ELSE 0 END), 0)
			-
			COALESCE(SUM(CASE WHEN type = 'payment' THEN amount ELSE 0 END), 0) AS current_due
		FROM customer_ledger_entries
		WHERE shop_id = ? AND customer_id = ?
	`, shopID, customerID).Scan(&due).Error

	return due, err
}
