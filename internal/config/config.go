package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName           string
	AppEnv            string
	AppPort           string
	DatabaseURL       string
	JWTSecret         string
	JWTExpiresInHours int
	FrontendURL       string
}

func Load() Config {
	_ = godotenv.Load()

	expires, err := strconv.Atoi(getEnv("JWT_EXPIRES_IN_HOURS", "24"))
	if err != nil {
		log.Fatal("Invalid JWT_EXPIRES_IN_HOURS")
	}

	return Config{
		AppName:           getEnv("APP_NAME", "Unify Nepal API"),
		AppEnv:            getEnv("APP_ENV", "development"),
		AppPort:           getEnv("APP_PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		JWTSecret:         getEnv("JWT_SECRET", ""),
		JWTExpiresInHours: expires,
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
