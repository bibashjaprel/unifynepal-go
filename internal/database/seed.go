package database

import (
	"log"

	"github.com/bibashjaprel/unifynepal-api/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) {
	plans := []models.SubscriptionPlan{
		{
			Name:             "Starter",
			PriceMonthly:     499,
			MaxUsers:         1,
			MaxProducts:      100,
			MaxBillsPerMonth: 300,
			Features:         datatypes.JSON([]byte(`["billing","products","customers","udharo"]`)),
			IsActive:         true,
		},
		{
			Name:             "Business",
			PriceMonthly:     999,
			MaxUsers:         3,
			MaxProducts:      1000,
			MaxBillsPerMonth: 3000,
			Features:         datatypes.JSON([]byte(`["billing","products","customers","udharo","reports","inventory"]`)),
			IsActive:         true,
		},
		{
			Name:             "Pro",
			PriceMonthly:     1999,
			MaxUsers:         10,
			MaxProducts:      10000,
			MaxBillsPerMonth: 10000,
			Features:         datatypes.JSON([]byte(`["billing","products","customers","udharo","reports","inventory","admin","priority_support"]`)),
			IsActive:         true,
		},
	}

	for _, plan := range plans {
		var existing models.SubscriptionPlan
		if err := db.Where("name = ?", plan.Name).First(&existing).Error; err == nil {
			continue
		}

		if err := db.Create(&plan).Error; err != nil {
			log.Println("failed to seed plan:", plan.Name, err)
		}
	}
}
