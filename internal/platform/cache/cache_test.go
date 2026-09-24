package cache

import (
	"context"
	"testing"
	"time"

	"broadcast-platform/internal/platform/config"
	"broadcast-platform/internal/platform/logging"
)

func TestNoopStore(t *testing.T) {
	store := &noopStore{}

	if store.IsEnabled() {
		t.Error("expected noop store to be disabled")
	}

	ctx := context.Background()

	if err := store.Set(ctx, "key", []byte("value"), time.Minute); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	val, err := store.Get(ctx, "key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != nil {
		t.Errorf("expected nil value from noop store, got %v", val)
	}

	exists, err := store.Exists(ctx, "key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected noop Exists to return false")
	}

	if err := store.Delete(ctx, "key"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := store.Close(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewDisabledCache(t *testing.T) {
	cfg := config.CacheConfig{Enabled: false}
	logger, _ := logging.New(config.LoggingConfig{Level: "info", Format: "console"})
	store := New(cfg, logger)

	if store.IsEnabled() {
		t.Error("expected disabled cache")
	}
}

