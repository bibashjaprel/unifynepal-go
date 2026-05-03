package models

import (
	"time"

	"github.com/google/uuid"
)

type ShopSubscription struct {
	BaseModel

	ShopID uuid.UUID `gorm:"type:uuid;index;not null" json:"shop_id"`
	PlanID uuid.UUID `gorm:"type:uuid;index;not null" json:"plan_id"`

	Status string `gorm:"size:30;not null;default:trialing" json:"status"` // trialing, active, expired, cancelled

	TrialEndsAt        *time.Time `json:"trial_ends_at"`
	CurrentPeriodStart *time.Time `json:"current_period_start"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end"`
}
