package models

import (
	"time"

	"github.com/google/uuid"
)

type UserSession struct {
	BaseModel

	UserID uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	ShopID *uuid.UUID `gorm:"type:uuid;index" json:"shop_id"`

	TokenID string `gorm:"size:100;uniqueIndex;not null" json:"token_id"`

	IPAddress string `gorm:"size:100" json:"ip_address"`
	UserAgent string `gorm:"size:500" json:"user_agent"`
	Device    string `gorm:"size:100" json:"device"`
	Browser   string `gorm:"size:100" json:"browser"`
	OS        string `gorm:"size:100" json:"os"`

	LastSeenAt time.Time  `json:"last_seen_at"`
	ExpiresAt  time.Time  `gorm:"index;not null" json:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at"`

	IsActive bool `gorm:"default:true" json:"is_active"`
}
