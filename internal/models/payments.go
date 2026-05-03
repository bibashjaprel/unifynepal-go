package models

import "github.com/google/uuid"

type Payment struct {
	BaseModel

	ShopID     uuid.UUID  `gorm:"type:uuid;index;not null" json:"shop_id"`
	BillID     *uuid.UUID `gorm:"type:uuid;index" json:"bill_id"`
	CustomerID *uuid.UUID `gorm:"type:uuid;index" json:"customer_id"`

	Amount      float64 `gorm:"not null" json:"amount"`
	Method      string  `gorm:"size:30;not null" json:"method"` // cash, online, card
	ReferenceNo string  `gorm:"size:100" json:"reference_no"`
	Note        string  `gorm:"size:255" json:"note"`
}
