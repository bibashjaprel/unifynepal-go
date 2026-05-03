package models

import "gorm.io/datatypes"

type SubscriptionPlan struct {
	BaseModel

	Name             string         `gorm:"size:100;uniqueIndex;not null" json:"name"`
	PriceMonthly     float64        `gorm:"not null;default:0" json:"price_monthly"`
	MaxUsers         int            `gorm:"not null;default:1" json:"max_users"`
	MaxProducts      int            `gorm:"not null;default:100" json:"max_products"`
	MaxBillsPerMonth int            `gorm:"not null;default:100" json:"max_bills_per_month"`
	Features         datatypes.JSON `gorm:"type:jsonb" json:"features"`
	IsActive         bool           `gorm:"default:true" json:"is_active"`
}
