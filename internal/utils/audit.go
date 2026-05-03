package utils

import (
	"encoding/json"

	"github.com/bibashjaprel/unifynepal-api/internal/models"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AuditInput struct {
	ShopID     *uuid.UUID
	UserID     *uuid.UUID
	Action     string
	EntityType string
	EntityID   *uuid.UUID
	Metadata   interface{}
}

func CreateAuditLog(db *gorm.DB, input AuditInput) {
	var metadata datatypes.JSON

	if input.Metadata != nil {
		bytes, err := json.Marshal(input.Metadata)
		if err == nil {
			metadata = datatypes.JSON(bytes)
		}
	}

	log := models.AuditLog{
		ShopID:     input.ShopID,
		UserID:     input.UserID,
		Action:     input.Action,
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		Metadata:   metadata,
	}

	_ = db.Create(&log).Error
}
