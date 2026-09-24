package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	Logging   LoggingConfig   `yaml:"logging"`
	Metrics   MetricsConfig   `yaml:"metrics"`
	Cache     CacheConfig     `yaml:"cache"`
	CORS      CORSConfig      `yaml:"cors"`
	Tracing   TracingConfig   `yaml:"tracing"`
	Execution ExecutionConfig `yaml:"execution"`
	Auth      AuthConfig      `yaml:"auth"`
}

// ServerConfig represents server configuration.
type ServerConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Env             string        `yaml:"env"` // development, production, test
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

// DatabaseConfig represents database configuration.
type DatabaseConfig struct {
	Driver          string `yaml:"driver"` // sqlite3, mysql
	DSN             string `yaml:"dsn"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime_seconds"`
}

// LoggingConfig represents logging configuration.
type LoggingConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error, fatal
	Format string `yaml:"format"` // json, console
}

// MetricsConfig represents Prometheus metrics configuration.
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

// CacheConfig represents cache/Redis configuration.
type CacheConfig struct {
	Enabled bool   `yaml:"enabled"`
	Driver  string `yaml:"driver"` // redis, noop
	DSN     string `yaml:"dsn"`
}

// CORSConfig represents CORS configuration.
type CORSConfig struct {
	AllowOrigins []string `yaml:"allow_origins"`
	AllowMethods []string `yaml:"allow_methods"`
	AllowHeaders []string `yaml:"allow_headers"`
}

// TracingConfig represents tracing configuration.
type TracingConfig struct {
	Enabled bool `yaml:"enabled"`
}

// ExecutionConfig represents execution engine configuration.
type ExecutionConfig struct {
	MaxWorkers         int `yaml:"max_workers"`          // max concurrent device executions
	MaxBatchSize       int `yaml:"max_batch_size"`       // max devices per batch operation
	ShutdownTimeoutSec int `yaml:"shutdown_timeout_sec"` // seconds to wait for in-flight ops
}

// AuthConfig represents authentication configuration.
type AuthConfig struct {
	JWTSecret        string `yaml:"jwt_secret"`         // JWT signing secret (required)
	TokenExpiryHours int    `yaml:"token_expiry_hours"` // JWT expiry in hours (default: 2)
}

// Load loads configuration from file and environment variables.
func Load(configPath string) (*Config, error) {
	cfg := defaultConfig()

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	applyEnvOverrides(cfg)
	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			Env:             "development",
			ShutdownTimeout: 10 * time.Second,
		},
		Database: DatabaseConfig{
			Driver:          "sqlite3",
			DSN:             "file:broadcast.db?cache=shared&mode=rwc",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "console",
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
		},
		Cache: CacheConfig{
			Enabled: false,
			Driver:  "noop",
		},
		CORS: CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
		},
		Tracing: TracingConfig{
			Enabled: true,
		},
		Execution: ExecutionConfig{
			MaxWorkers:         10,
			MaxBatchSize:       100,
			ShutdownTimeoutSec: 30,
		},
		Auth: AuthConfig{
			JWTSecret:        "change-me-in-production",
			TokenExpiryHours: 2,
		},
	}
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("SERVER_ENV"); v != "" {
		cfg.Server.Env = v
	}
	if v := os.Getenv("DB_DRIVER"); v != "" {
		cfg.Database.Driver = v
	}
	if v := os.Getenv("DB_DSN"); v != "" {
		cfg.Database.DSN = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		cfg.Logging.Format = v
	}
	if v := os.Getenv("METRICS_ENABLED"); v != "" {
		cfg.Metrics.Enabled = strings.ToLower(v) == "true"
	}
	if v := os.Getenv("CACHE_ENABLED"); v != "" {
		cfg.Cache.Enabled = strings.ToLower(v) == "true"
	}
	if v := os.Getenv("CACHE_DSN"); v != "" {
		cfg.Cache.DSN = v
	}
	if v := os.Getenv("EXEC_MAX_WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Execution.MaxWorkers = n
		}
	}
	if v := os.Getenv("EXEC_MAX_BATCH_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Execution.MaxBatchSize = n
		}
	}
	if v := os.Getenv("AUTH_JWT_SECRET"); v != "" {
		cfg.Auth.JWTSecret = v
	}
	if v := os.Getenv("AUTH_TOKEN_EXPIRY_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Auth.TokenExpiryHours = n
		}
	}
}

// Address returns the server listen address.
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsDevelopment returns true if running in development environment.
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.Server.Env) == "development"
}

// IsProduction returns true if running in production environment.
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Server.Env) == "production"
}

// IsTest returns true if running in test environment.
func (c *Config) IsTest() bool {
	return strings.ToLower(c.Server.Env) == "test"
}