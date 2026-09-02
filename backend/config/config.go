package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

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
	_ = godotenv.Load()

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

	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	return cfg
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.AppEnv) == "" {
		return errors.New("APP_ENV must not be empty")
	}
	if c.Port == "" || c.DBHost == "" || c.DBPort == "" || c.DBUser == "" || c.DBName == "" {
		return errors.New("database and server configuration must not be empty")
	}

	if c.AppEnv == "production" {
		if c.JWTSecret == "" || c.JWTSecret == "super-secret-key-change-me" {
			return errors.New("JWT_SECRET must be set to a secure value in production")
		}
		if c.DBPassword == "" || c.DBPassword == "postgres" {
			return errors.New("DB_PASSWORD must be set to a secure value in production")
		}
		if c.GoogleClientID == "" || c.GoogleClientSecret == "" || c.GoogleRedirectURL == "" {
			return errors.New("Google OAuth configuration must be set in production")
		}
		if c.FrontendURL == "" {
			return errors.New("FRONTEND_URL must be set in production")
		}
	}

	return nil
}

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
