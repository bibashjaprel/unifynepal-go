package models

import "github.com/google/uuid"

type BillItem struct {
	BaseModel
	BillID    uuid.UUID `gorm:"type:uuid;index;not null" json:"bill_id"`
	ProductID uuid.UUID `gorm:"type:uuid;index;not null" json:"product_id"`

	ProductName string  `gorm:"size:150;not null" json:"product_name"`
	Quantity    float64 `gorm:"not null" json:"quantity"`
	UnitPrice   float64 `gorm:"not null" json:"unit_price"`
	CostPrice   float64 `gorm:"default:0" json:"cost_price"`
	Total       float64 `gorm:"not null" json:"total"`
}
