package models

import "github.com/google/uuid"

type Bill struct {
	BaseModel
	ShopID     uuid.UUID  `gorm:"type:uuid;index;not null" json:"shop_id"`
	CustomerID *uuid.UUID `gorm:"type:uuid;index" json:"customer_id"`
	BillNumber string     `gorm:"size:50;index;not null" json:"bill_number"`

	Subtotal    float64 `gorm:"not null" json:"subtotal"`
	Discount    float64 `gorm:"default:0" json:"discount"`
	TotalAmount float64 `gorm:"not null" json:"total_amount"`
	PaidAmount  float64 `gorm:"default:0" json:"paid_amount"`
	DueAmount   float64 `gorm:"default:0" json:"due_amount"`

	PaymentStatus string `gorm:"size:30;default:unpaid" json:"payment_status"` // paid, partial, unpaid
	PaymentMethod string `gorm:"size:30" json:"payment_method"`                // cash, online, card

	Items []BillItem `gorm:"foreignKey:BillID" json:"items"`
}
