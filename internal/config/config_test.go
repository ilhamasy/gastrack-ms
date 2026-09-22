package config_test

import (
	"os"
	"testing"

	"github.com/ilhamasy/gastrack-ms/internal/config"
)

func TestLoad_Success(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://test@localhost/testdb")
	defer os.Unsetenv("DATABASE_URL")
	os.Setenv("PORT", "9090")
	defer os.Unsetenv("PORT")
	os.Setenv("JWT_SECRET", "supersecret")
	defer os.Unsetenv("JWT_SECRET")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.DatabaseURL != "postgres://test@localhost/testdb" {
		t.Errorf("expected DATABASE_URL postgres://test@localhost/testdb, got %s", cfg.DatabaseURL)
	}
	if cfg.Port != "9090" {
		t.Errorf("expected PORT 9090, got %s", cfg.Port)
	}
	if cfg.JWTSecret != "supersecret" {
		t.Errorf("expected JWT_SECRET supersecret, got %s", cfg.JWTSecret)
	}
}

func TestLoad_DefaultPort(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://test@localhost/testdb")
	defer os.Unsetenv("DATABASE_URL")
	os.Setenv("JWT_SECRET", "supersecret")
	defer os.Unsetenv("JWT_SECRET")
	os.Unsetenv("PORT")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected default PORT 8080, got %s", cfg.Port)
	}
}

func TestLoad_MissingDBURL(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Setenv("JWT_SECRET", "supersecret")
	defer os.Unsetenv("JWT_SECRET")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL, got nil")
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://test@localhost/testdb")
	defer os.Unsetenv("DATABASE_URL")
	os.Unsetenv("JWT_SECRET")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing JWT_SECRET, got nil")
	}
}
