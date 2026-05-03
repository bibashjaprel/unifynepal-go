package models

type Shop struct {
	BaseModel
	Name        string `gorm:"size:150;not null" json:"name"`
	Phone       string `gorm:"size:30" json:"phone"`
	Address     string `gorm:"size:255" json:"address"`
	Status      string `gorm:"size:30;default:active" json:"status"` // active, paused
	OwnerUserID string `gorm:"type:uuid" json:"owner_user_id"`
}
