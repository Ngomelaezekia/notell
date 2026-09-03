package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
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
	PublicURL          string
	MediaPublicURL     string
	StorageDriver      string
	R2Endpoint         string
	R2Bucket           string
	R2AccessKeyID      string
	R2SecretAccessKey  string
	R2Region           string
}

func Load() *Config {
	_ = godotenv.Load()

	serverURL := strings.TrimRight(strings.TrimSpace(getEnv("SERVER_URL", "http://localhost:8080")), "/")
	cfg := &Config{
		AppEnv:             strings.ToLower(strings.TrimSpace(getEnv("APP_ENV", "development"))),
		Port:               strings.TrimSpace(getEnv("PORT", "8080")),
		DBHost:             strings.TrimSpace(getEnv("DB_HOST", "localhost")),
		DBPort:             strings.TrimSpace(getEnv("DB_PORT", "5432")),
		DBUser:             strings.TrimSpace(getEnv("DB_USER", "postgres")),
		DBPassword:         getEnv("DB_PASSWORD", "postgres"),
		DBName:             strings.TrimSpace(getEnv("DB_NAME", "notell_db")),
		JWTSecret:          getEnv("JWT_SECRET", "super-secret-key-change-me"),
		GoogleClientID:     strings.TrimSpace(getEnv("GOOGLE_CLIENT_ID", "")),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  strings.TrimSpace(getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback")),
		FrontendURL:        normalizeFrontendURL(getEnv("FRONTEND_URL", "http://localhost:5173")),
		PublicURL:          serverURL,
		MediaPublicURL:     strings.TrimRight(strings.TrimSpace(getEnv("MEDIA_PUBLIC_URL", serverURL)), "/"),
		StorageDriver:      strings.ToLower(strings.TrimSpace(getEnv("STORAGE_DRIVER", "local"))),
		R2Endpoint:         strings.TrimRight(strings.TrimSpace(getEnv("R2_ENDPOINT", "")), "/"),
		R2Bucket:           strings.TrimSpace(getEnv("R2_BUCKET", "")),
		R2AccessKeyID:      strings.TrimSpace(getEnv("R2_ACCESS_KEY_ID", "")),
		R2SecretAccessKey:  getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2Region:           strings.TrimSpace(getEnv("R2_REGION", "auto")),
	}

	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	return cfg
}

func (c *Config) Validate() error {
	appEnv := strings.ToLower(strings.TrimSpace(c.AppEnv))
	if appEnv == "" {
		return errors.New("APP_ENV must not be empty")
	}
	if c.Port == "" || c.DBHost == "" || c.DBPort == "" || c.DBUser == "" || c.DBName == "" {
		return errors.New("database and server configuration must not be empty")
	}

	if port, err := strconv.Atoi(c.Port); err != nil || port < 1 || port > 65535 {
		return errors.New("PORT must be a valid TCP port")
	}
	if port, err := strconv.Atoi(c.DBPort); err != nil || port < 1 || port > 65535 {
		return errors.New("DB_PORT must be a valid TCP port")
	}

	if !strings.EqualFold(c.StorageDriver, "local") && !strings.EqualFold(c.StorageDriver, "r2") {
		return errors.New("STORAGE_DRIVER must be local or r2")
	}

	if appEnv == "production" {
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
		if err := validateFrontendURL(c.FrontendURL); err != nil {
			return fmt.Errorf("FRONTEND_URL: %w", err)
		}
		if err := validateAbsoluteURL(c.GoogleRedirectURL); err != nil {
			return fmt.Errorf("GOOGLE_REDIRECT_URL: %w", err)
		}
		if err := validatePublicURL(c.PublicURL); err != nil {
			return fmt.Errorf("SERVER_URL: %w", err)
		}
		if !strings.EqualFold(c.StorageDriver, "r2") {
			return errors.New("STORAGE_DRIVER must be r2 in production")
		}
		if err := validatePublicURL(c.MediaPublicURL); err != nil {
			return fmt.Errorf("MEDIA_PUBLIC_URL: %w", err)
		}
		if c.R2Endpoint == "" || c.R2Bucket == "" || c.R2AccessKeyID == "" || c.R2SecretAccessKey == "" {
			return errors.New("R2 storage configuration must be set in production")
		}
		if err := validateAbsoluteURL(c.R2Endpoint); err != nil {
			return fmt.Errorf("R2_ENDPOINT: %w", err)
		}
		if c.R2Region == "" {
			return errors.New("R2_REGION must not be empty")
		}
	}

	return nil
}

func normalizeFrontendURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.TrimRight(value, "/")
}

func validateFrontendURL(value string) error {
	if err := validateAbsoluteURL(value); err != nil {
		return err
	}
	parsed, _ := url.Parse(strings.TrimSpace(value))
	if parsed.Path != "" && parsed.Path != "/" {
		return errors.New("must not include a URL path")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("must not include query or fragment components")
	}
	return nil
}

func validateAbsoluteURL(value string) error {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("must be an absolute http or https URL")
	}
	if parsed.User != nil {
		return errors.New("must not include user information")
	}
	return nil
}

func validatePublicURL(value string) error {
	if err := validateAbsoluteURL(value); err != nil {
		return err
	}
	parsed, _ := url.Parse(strings.TrimSpace(value))
	if parsed.Path != "" && parsed.Path != "/" {
		return errors.New("must not include a URL path")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("must not include query or fragment components")
	}
	return nil
}

func quoteConninfoValue(value string) string {
	if value != "" && !strings.ContainsAny(value, " \t\n\r'\\") {
		return value
	}
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "'", "\\'")
	return "'" + value + "'"
}

func (c *Config) GetDBDSN() string {
	sslMode := "disable"
	if strings.EqualFold(strings.TrimSpace(c.AppEnv), "production") {
		sslMode = "require"
	}

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		quoteConninfoValue(c.DBHost),
		quoteConninfoValue(c.DBUser),
		quoteConninfoValue(c.DBPassword),
		quoteConninfoValue(c.DBName),
		quoteConninfoValue(c.DBPort),
		quoteConninfoValue(sslMode),
	)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
