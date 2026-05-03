package models

import "github.com/google/uuid"

type Notification struct {
	BaseModel

	ShopID *uuid.UUID `gorm:"type:uuid;index" json:"shop_id"`
	UserID uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`

	Title   string `gorm:"size:150;not null" json:"title"`
	Message string `gorm:"size:500;not null" json:"message"`
	Type    string `gorm:"size:50;not null" json:"type"` // system, billing, udharo, stock, subscription
	IsRead  bool   `gorm:"default:false" json:"is_read"`
}
