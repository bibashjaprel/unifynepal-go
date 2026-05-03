package database

import (
	"log"

	"github.com/bibashjaprel/unifynepal-api/internal/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.User{},
		&models.Shop{},
		&models.ShopMember{},
		&models.Product{},
		&models.Customer{},
		&models.Bill{},
		&models.BillItem{},
		&models.Payment{},
		&models.CustomerLedgerEntry{},
		&models.StockMovement{},
		&models.SubscriptionPlan{},
		&models.ShopSubscription{},
		&models.Notification{},
		&models.AuditLog{},
		&models.UserSession{},
	)
	if err != nil {
		log.Fatal("Migration failed: ", err)
	}
}
