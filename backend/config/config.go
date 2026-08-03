package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	Port               string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	JWTSecret          string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	FrontendURL        string
}

func Load() *Config {
	// Attempt to load .env file (ignore error if file doesn't exist, e.g. in container environments)
	if err := godotenv.Load(); err != nil {
		log.Println("Config: No .env file found, relying on environment variables.")
	}

	cfg := &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		Port:               getEnv("PORT", "8080"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", "postgres"),
		DBName:             getEnv("DB_NAME", "notell_db"),
		JWTSecret:          getEnv("JWT_SECRET", "super-secret-key-change-me"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:5173"),
	}

	cfg.Validate()
	return cfg
}

// Validate checks for critical configuration values in non-development environments.
func (c *Config) Validate() {
	if c.AppEnv == "production" {
		if c.JWTSecret == "super-secret-key-change-me" || c.JWTSecret == "" {
			log.Fatal("CRITICAL: JWT_SECRET must be set to a secure string in production")
		}
		if c.GoogleClientID == "" || c.GoogleClientSecret == "" {
			log.Fatal("CRITICAL: GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set in production")
		}
	}
}

// GetDBDSN builds and returns the PostgreSQL connection string.
func (c *Config) GetDBDSN() string {
	sslMode := "disable"
	if c.AppEnv == "production" {
		sslMode = "require"
	}

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, sslMode,
	)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
