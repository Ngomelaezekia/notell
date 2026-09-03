package config

import (
	"errors"
	"fmt"
	"net"
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
}

func Load() *Config {
	_ = godotenv.Load()

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
		FrontendURL:        strings.TrimSpace(getEnv("FRONTEND_URL", "http://localhost:5173")),
		PublicURL:          strings.TrimRight(strings.TrimSpace(getEnv("SERVER_URL", "http://localhost:8080")), "/"),
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
		if err := validateAbsoluteURL(c.FrontendURL); err != nil {
			return fmt.Errorf("FRONTEND_URL: %w", err)
		}
		if err := validateAbsoluteURL(c.GoogleRedirectURL); err != nil {
			return fmt.Errorf("GOOGLE_REDIRECT_URL: %w", err)
		}
		if err := validatePublicURL(c.PublicURL); err != nil {
			return fmt.Errorf("SERVER_URL: %w", err)
		}
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

func (c *Config) GetDBDSN() string {
	sslMode := "disable"
	if strings.EqualFold(strings.TrimSpace(c.AppEnv), "production") {
		sslMode = "require"
	}

	host := strings.TrimSpace(c.DBHost)
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		host = net.JoinHostPort(host, c.DBPort)
	} else {
		host = net.JoinHostPort(host, c.DBPort)
	}

	dsn := url.URL{
		Scheme: "postgres",
		Host:   host,
		Path:   "/" + url.PathEscape(c.DBName),
		User:   url.UserPassword(c.DBUser, c.DBPassword),
	}
	query := dsn.Query()
	query.Set("sslmode", sslMode)
	dsn.RawQuery = query.Encode()
	return dsn.String()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
