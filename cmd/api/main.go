package main

import (
	"log"

	"github.com/bibashjaprel/unifynepal-api/internal/config"
	"github.com/bibashjaprel/unifynepal-api/internal/database"
	"github.com/bibashjaprel/unifynepal-api/internal/middleware"
	"github.com/bibashjaprel/unifynepal-api/internal/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	db := database.Connect(cfg.DatabaseURL)
	database.AutoMigrate(db)

	router := gin.New()

	router.Use(middleware.RequestID())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.Recovery())

	routes.Register(router, db, cfg)

	log.Println("Starting", cfg.AppName, "on port", cfg.AppPort)

	err := router.Run(":" + cfg.AppPort)
	if err != nil {
		log.Fatal(err)
	}
}
