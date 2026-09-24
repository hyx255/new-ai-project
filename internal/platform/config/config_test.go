package config

import (
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.Env != "development" {
		t.Errorf("expected env development, got %s", cfg.Server.Env)
	}
	if cfg.Server.ShutdownTimeout != 10*time.Second {
		t.Errorf("expected shutdown timeout 10s, got %v", cfg.Server.ShutdownTimeout)
	}
	if cfg.Database.Driver != "sqlite3" {
		t.Errorf("expected driver sqlite3, got %s", cfg.Database.Driver)
	}
	if cfg.Database.MaxOpenConns != 25 {
		t.Errorf("expected max open conns 25, got %d", cfg.Database.MaxOpenConns)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("expected level info, got %s", cfg.Logging.Level)
	}
	if !cfg.Metrics.Enabled {
		t.Error("expected metrics enabled by default")
	}
	if cfg.Metrics.Path != "/metrics" {
		t.Errorf("expected metrics path /metrics, got %s", cfg.Metrics.Path)
	}
	if cfg.Cache.Enabled {
		t.Error("expected cache disabled by default")
	}
	if !cfg.Tracing.Enabled {
		t.Error("expected tracing enabled by default")
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	cfg, err := Load("nonexistent.yaml")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
}

func TestEnvironmentOverrides(t *testing.T) {
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("SERVER_ENV", "production")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("METRICS_ENABLED", "false")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SERVER_ENV")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("METRICS_ENABLED")
	}()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Server.Env != "production" {
		t.Errorf("expected env production, got %s", cfg.Server.Env)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("expected level debug, got %s", cfg.Logging.Level)
	}
	if cfg.Metrics.Enabled {
		t.Error("expected metrics disabled from env override")
	}
}

func TestConfigAddress(t *testing.T) {
	cfg := &Config{Server: ServerConfig{Host: "localhost", Port: 3000}}
	if cfg.Address() != "localhost:3000" {
		t.Errorf("expected localhost:3000, got %s", cfg.Address())
	}
}

func TestConfigIsDevelopment(t *testing.T) {
	cfg := &Config{Server: ServerConfig{Env: "development"}}
	if !cfg.IsDevelopment() {
		t.Error("expected IsDevelopment true")
	}
	cfg.Server.Env = "production"
	if cfg.IsDevelopment() {
		t.Error("expected IsDevelopment false for production")
	}
}

func TestConfigIsProduction(t *testing.T) {
	cfg := &Config{Server: ServerConfig{Env: "production"}}
	if !cfg.IsProduction() {
		t.Error("expected IsProduction true")
	}
}

func TestConfigIsTest(t *testing.T) {
	cfg := &Config{Server: ServerConfig{Env: "test"}}
	if !cfg.IsTest() {
		t.Error("expected IsTest true")
	}
	cfg.Server.Env = "development"
	if cfg.IsTest() {
		t.Error("expected IsTest false for development")
	}
}

