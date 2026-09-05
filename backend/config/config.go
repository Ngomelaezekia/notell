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
	B2Endpoint         string
	B2Bucket           string
	B2KeyID            string
	B2ApplicationKey   string
	B2Region           string
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
		B2Endpoint:         normalizeEndpoint(getEnv("B2_ENDPOINT", "")),
		B2Bucket:           strings.TrimSpace(getEnv("B2_BUCKET", "")),
		B2KeyID:            strings.TrimSpace(getEnv("B2_KEY_ID", "")),
		B2ApplicationKey:   getEnv("B2_APPLICATION_KEY", ""),
		B2Region:           strings.TrimSpace(getEnv("B2_REGION", "")),
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

	storageDriver := strings.ToLower(strings.TrimSpace(c.StorageDriver))
	if storageDriver == "" {
		storageDriver = "local"
	}
	if storageDriver != "local" && storageDriver != "r2" && storageDriver != "b2" {
		return errors.New("STORAGE_DRIVER must be local, r2, or b2")
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

		switch storageDriver {
		case "r2":
			return errors.New("R2 is not configured in this deployment; use STORAGE_DRIVER=b2 for Backblaze B2")
		case "b2":
			if err := validateAbsoluteURL(c.B2Endpoint); err != nil {
				return fmt.Errorf("B2_ENDPOINT: %w", err)
			}
			if c.B2Bucket == "" || c.B2KeyID == "" || c.B2ApplicationKey == "" || c.B2Region == "" {
				return errors.New("Backblaze B2 storage configuration must be set in production")
			}
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

func normalizeEndpoint(value string) string {
	value = strings.TrimRight(strings.TrimSpace(value), "/")
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err == nil && parsed.Scheme != "" {
		return value
	}
	return "https://" + value
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
