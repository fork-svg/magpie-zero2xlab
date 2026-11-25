package config

import (
	"os"
)

type Config struct {
	// Server
	Port string
	Host string

	// JWT
	JWTSecret string

	// Zero2XLab
	Zero2XLabBaseURL string
	Zero2XLabAPIKey  string // Master API key for zero2xlab

	// Database
	DBPath string
}

func Load() *Config {
	return &Config{
		Port: getEnv("PORT", "8080"),
		Host: getEnv("HOST", "0.0.0.0"),

		JWTSecret: getEnv("JWT_SECRET", "change-this-secret-in-production"),

		Zero2XLabBaseURL: getEnv("ZERO2XLAB_BASE_URL", "https://oneastronomy.zero2x.org/api/platform/v1"),
		Zero2XLabAPIKey:  getEnv("ZERO2XLAB_API_KEY", ""),

		DBPath: getEnv("DB_PATH", "./data/gateway.db"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
