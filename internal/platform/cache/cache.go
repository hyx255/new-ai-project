// Package cache provides a cache abstraction with pluggable implementations.
// Per ADR-003, Redis is conditionally introduced only when there is a clear need.
// This package provides a Store interface and a default NoOp implementation.
package cache

import (
	"context"
	"time"

	"broadcast-platform/internal/platform/config"
	"broadcast-platform/internal/platform/logging"

	"go.uber.org/zap"
)

// Store defines the cache interface.
// All business modules must use this interface, not a concrete implementation.
type Store interface {
	// Get retrieves a value by key. Returns nil, nil if key does not exist.
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value with optional TTL. TTL of 0 means no expiration.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a key.
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists.
	Exists(ctx context.Context, key string) (bool, error)

	// Close closes the cache connection.
	Close() error

	// IsEnabled returns whether the cache is active.
	IsEnabled() bool
}

// New creates a new cache store based on configuration.
// Returns a NoOp store when cache is disabled.
func New(cfg config.CacheConfig, logger *logging.Logger) Store {
	if !cfg.Enabled {
		logger.Info("cache disabled, using NoOp store")
		return &noopStore{}
	}

	logger.Warn("cache enabled but no implementation available, using NoOp store",
		zap.String("driver", cfg.Driver))
	return &noopStore{}
}

// noopStore is a no-operation cache that does nothing.
type noopStore struct{}

func (n *noopStore) Get(_ context.Context, _ string) ([]byte, error) { return nil, nil }
func (n *noopStore) Set(_ context.Context, _ string, _ []byte, _ time.Duration) error { return nil }
func (n *noopStore) Delete(_ context.Context, _ string) error        { return nil }
func (n *noopStore) Exists(_ context.Context, _ string) (bool, error) { return false, nil }
func (n *noopStore) Close() error                                    { return nil }
func (n *noopStore) IsEnabled() bool                                 { return false }
