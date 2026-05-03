package models

import "github.com/google/uuid"

type StockMovement struct {
	BaseModel

	ShopID    uuid.UUID `gorm:"type:uuid;index;not null" json:"shop_id"`
	ProductID uuid.UUID `gorm:"type:uuid;index;not null" json:"product_id"`

	Type        string  `gorm:"size:30;not null" json:"type"` // opening, stock_in, stock_out, sale, adjustment
	Quantity    float64 `gorm:"not null" json:"quantity"`
	PreviousQty float64 `gorm:"not null" json:"previous_qty"`
	NewQty      float64 `gorm:"not null" json:"new_qty"`

	ReferenceType string     `gorm:"size:50" json:"reference_type"` // bill, manual, product_create
	ReferenceID   *uuid.UUID `gorm:"type:uuid;index" json:"reference_id"`

	Note string `gorm:"size:255" json:"note"`
}
