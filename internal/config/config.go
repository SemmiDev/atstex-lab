package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// AppConfig holds centralized environment properties.
type AppConfig struct {
	Port               string
	DatabaseURL        string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleCallbackURL  string
}

// Load reads .env and returns a populated AppConfig.
func Load() *AppConfig {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found or couldn't load it. Using system env variables.")
	}

	return &AppConfig{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://atstex:password@localhost:5432/atstex?sslmode=disable"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleCallbackURL:  getEnv("GOOGLE_CALLBACK_URL", "http://localhost:8080/auth/google/callback"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
