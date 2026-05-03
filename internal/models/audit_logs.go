package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type AuditLog struct {
	BaseModel

	ShopID *uuid.UUID `gorm:"type:uuid;index" json:"shop_id"`
	UserID *uuid.UUID `gorm:"type:uuid;index" json:"user_id"`

	Action     string         `gorm:"size:100;not null" json:"action"`
	EntityType string         `gorm:"size:100;not null" json:"entity_type"`
	EntityID   *uuid.UUID     `gorm:"type:uuid;index" json:"entity_id"`
	Metadata   datatypes.JSON `gorm:"type:jsonb" json:"metadata"`
}
