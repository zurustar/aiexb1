package config

import (
	"log"
	"os"
)

// Config holds the application configuration.
type Config struct {
	JWTSecret string
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() *Config {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("WARNING: JWT_SECRET environment variable not set. Using a default, insecure key. Please set a strong secret in production.")
		secret = "a-very-insecure-default-secret-key" // This should not be used in production
	}

	return &Config{
		JWTSecret: secret,
	}
}