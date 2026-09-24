package logging

import (
	"testing"

	"broadcast-platform/internal/platform/config"
)

func TestNewLogger(t *testing.T) {
	cfg := config.LoggingConfig{Level: "info", Format: "console"}
	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Sync()

	if logger.module != "platform" {
		t.Errorf("expected default module platform, got %s", logger.module)
	}
}

func TestWithModule(t *testing.T) {
	cfg := config.LoggingConfig{Level: "info", Format: "console"}
	logger, _ := New(cfg)
	defer logger.Sync()

	modLogger := logger.WithModule("database")
	if modLogger.module != "database" {
		t.Errorf("expected module database, got %s", modLogger.module)
	}
}

func TestWithTraceID(t *testing.T) {
	cfg := config.LoggingConfig{Level: "info", Format: "json"}
	logger, _ := New(cfg)
	defer logger.Sync()

	traceLogger := logger.WithTraceID("trace-123")
	if traceLogger == nil {
		t.Error("expected non-nil logger")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"error", "error"},
		{"fatal", "fatal"},
		{"unknown", "info"},
	}
	for _, tt := range tests {
		level := parseLevel(tt.input)
		if level.String() != tt.want {
			t.Errorf("parseLevel(%q) = %s, want %s", tt.input, level.String(), tt.want)
		}
	}
}

