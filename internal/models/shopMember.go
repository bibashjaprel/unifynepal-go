package models

import "github.com/google/uuid"

type ShopMember struct {
	BaseModel
	ShopID uuid.UUID `gorm:"type:uuid;index;not null" json:"shop_id"`
	UserID uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Role   string    `gorm:"size:50;not null" json:"role"` // owner, admin, cashier, staff
}
