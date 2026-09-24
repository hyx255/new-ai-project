package database

import (
	"database/sql"
	"fmt"
	"time"

	"broadcast-platform/internal/platform/config"
	"broadcast-platform/internal/platform/logging"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// DB wraps sql.DB with lifecycle management.
type DB struct {
	*sql.DB
	logger *logging.Logger
	driver string
}

// New creates a new database connection with pool configuration.
func New(cfg config.DatabaseConfig, logger *logging.Logger) (*DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("database connected",
		zap.String("driver", cfg.Driver),
		zap.Int("max_open", cfg.MaxOpenConns),
		zap.Int("max_idle", cfg.MaxIdleConns),
	)

	return &DB{
		DB:     db,
		logger: logger,
		driver: cfg.Driver,
	}, nil
}

// HealthCheck performs a health check on the database.
func (db *DB) HealthCheck() error {
	return db.Ping()
}

// Driver returns the database driver name.
func (db *DB) Driver() string {
	return db.driver
}

// Close closes the database connection.
func (db *DB) Close() error {
	db.logger.Info("closing database connection")
	return db.DB.Close()
}

