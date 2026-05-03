package models

import "github.com/google/uuid"

type Customer struct {
	BaseModel

	ShopID  uuid.UUID `gorm:"type:uuid;index;not null" json:"shop_id"`
	Name    string    `gorm:"size:150;not null" json:"name"`
	Phone   string    `gorm:"size:30;index" json:"phone"`
	Address string    `gorm:"size:255" json:"address"`

	IsActive bool `gorm:"default:true" json:"is_active"`
}
