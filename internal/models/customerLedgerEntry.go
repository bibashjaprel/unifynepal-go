package models

import "github.com/google/uuid"

type CustomerLedgerEntry struct {
	BaseModel
	ShopID     uuid.UUID  `gorm:"type:uuid;index;not null" json:"shop_id"`
	CustomerID uuid.UUID  `gorm:"type:uuid;index;not null" json:"customer_id"`
	BillID     *uuid.UUID `gorm:"type:uuid;index" json:"bill_id"`

	Type        string  `gorm:"size:30;not null" json:"type"` // credit, payment, adjustment
	Amount      float64 `gorm:"not null" json:"amount"`
	Description string  `gorm:"size:255" json:"description"`
}
