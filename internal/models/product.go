package models

import "github.com/google/uuid"

type Product struct {
	BaseModel
	ShopID       uuid.UUID `gorm:"type:uuid;index;not null" json:"shop_id"`
	Name         string    `gorm:"size:150;not null" json:"name"`
	SKU          string    `gorm:"size:100" json:"sku"`
	Category     string    `gorm:"size:100" json:"category"`
	Unit         string    `gorm:"size:50;default:piece" json:"unit"`
	SellingPrice float64   `gorm:"not null" json:"selling_price"`
	CostPrice    float64   `gorm:"default:0" json:"cost_price"`
	StockQty     float64   `gorm:"default:0" json:"stock_qty"`
	MinStockQty  float64   `gorm:"default:0" json:"min_stock_qty"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
}
