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
	// AI extraction settings
	AIProvider string // "openai" (default), "anthropic", "ollama"
	AIModel    string // e.g. "gpt-4o-mini", "claude-3-haiku-20240307", "llama3"
	AIAPIKey   string
	AIBaseURL  string // optional: custom base URL for OpenAI-compatible APIs (Groq, Together, etc.)
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
		AIProvider:         getEnv("AI_PROVIDER", "openai"),
		AIModel:            getEnv("AI_MODEL", "gpt-4o-mini"),
		AIAPIKey:           getEnv("AI_API_KEY", getEnv("OPENAI_API_KEY", "")),
		AIBaseURL:          getEnv("AI_BASE_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
