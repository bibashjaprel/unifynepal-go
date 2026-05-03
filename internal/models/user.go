package models

type User struct {
	BaseModel
	Name         string `gorm:"size:150;not null" json:"name"`
	Email        string `gorm:"size:150;uniqueIndex;not null" json:"email"`
	PasswordHash string `gorm:"not null" json:"-"`
	Role         string `gorm:"size:50;default:user" json:"role"` // user, platform_admin
	IsActive     bool   `gorm:"default:true" json:"is_active"`
	IsVerified   bool   `gorm:"default:false" json:"is_verified"`
}
