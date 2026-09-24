package database

import (
	"os"
	"testing"

	"broadcast-platform/internal/platform/config"
	"broadcast-platform/internal/platform/logging"
)

func TestNewSQLiteConnection(t *testing.T) {
	dbPath := "file:test.db?cache=shared&mode=memory"
	cfg := config.DatabaseConfig{
		Driver:          "sqlite3",
		DSN:             dbPath,
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 60,
	}
	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "console"})

	db, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	if db.Driver() != "sqlite3" {
		t.Errorf("expected driver sqlite3, got %s", db.Driver())
	}

	if err := db.HealthCheck(); err != nil {
		t.Errorf("health check failed: %v", err)
	}
}

func TestNewDatabaseInvalidDSN(t *testing.T) {
	cfg := config.DatabaseConfig{
		Driver: "sqlite3",
		DSN:    "/nonexistent/path/that/does/not/exist/test.db",
	}
	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "console"})

	_, err := New(cfg, logger)
	if err == nil {
		t.Error("expected error for invalid DSN")
	}
}

func TestNewDatabaseInvalidDriver(t *testing.T) {
	cfg := config.DatabaseConfig{
		Driver: "nonexistent",
		DSN:    "test",
	}
	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "console"})

	_, err := New(cfg, logger)
	if err == nil {
		t.Error("expected error for invalid driver")
	}
}

func TestClose(t *testing.T) {
	dbPath := "file:test_close.db?cache=shared&mode=memory"
	cfg := config.DatabaseConfig{
		Driver: "sqlite3",
		DSN:    dbPath,
	}
	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "console"})

	db, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Errorf("close failed: %v", err)
	}
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Remove("test.db")
	os.Exit(code)
}

