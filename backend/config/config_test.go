package config

import "testing"

func TestValidateRequiresCoreConfiguration(t *testing.T) {
	cfg := Config{
		AppEnv:     "development",
		Port:       "8080",
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBName:     "notell_db",
		JWTSecret:  "development-secret",
		DBPassword: "postgres",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	cfg.Port = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() accepted an empty port")
	}
}

func TestValidateRejectsInsecureProductionDefaults(t *testing.T) {
	cfg := Config{
		AppEnv:             "production",
		Port:               "8080",
		DBHost:             "db",
		DBPort:             "5432",
		DBUser:             "postgres",
		DBPassword:         "postgres",
		DBName:             "notell_db",
		JWTSecret:          "super-secret-key-change-me",
		GoogleClientID:     "client-id",
		GoogleClientSecret: "client-secret",
		GoogleRedirectURL:  "https://example.com/callback",
		FrontendURL:        "https://example.com",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() accepted insecure production defaults")
	}
}

func TestGetDBDSNUsesProductionSSL(t *testing.T) {
	cfg := Config{
		AppEnv:      "production",
		DBHost:      "db.example",
		DBPort:      "5432",
		DBUser:      "app",
		DBPassword:  "secret",
		DBName:      "notell",
	}

	want := "host=db.example user=app password=secret dbname=notell port=5432 sslmode=require"
	if got := cfg.GetDBDSN(); got != want {
		t.Fatalf("GetDBDSN() = %q, want %q", got, want)
	}
}
