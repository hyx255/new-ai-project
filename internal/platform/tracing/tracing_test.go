package tracing

import (
	"context"
	"testing"
)

func TestNewTraceID(t *testing.T) {
	id1 := NewTraceID()
	id2 := NewTraceID()

	if len(id1) != 32 {
		t.Errorf("expected 32 hex chars, got %d", len(id1))
	}
	if id1 == id2 {
		t.Error("expected unique trace IDs")
	}
}

func TestWithAndGetTraceID(t *testing.T) {
	ctx := context.Background()

	if got := GetTraceID(ctx); got != "" {
		t.Errorf("expected empty trace ID, got %q", got)
	}

	ctx = WithTraceID(ctx, "test-trace-123")
	if got := GetTraceID(ctx); got != "test-trace-123" {
		t.Errorf("expected test-trace-123, got %q", got)
	}
}

func TestWithAndGetRequestID(t *testing.T) {
	ctx := context.Background()

	if got := GetRequestID(ctx); got != "" {
		t.Errorf("expected empty request ID, got %q", got)
	}

	ctx = WithRequestID(ctx, "req-456")
	if got := GetRequestID(ctx); got != "req-456" {
		t.Errorf("expected req-456, got %q", got)
	}
}

func TestExtractOrGenerate(t *testing.T) {
	if got := ExtractOrGenerate("existing-id"); got != "existing-id" {
		t.Errorf("expected existing-id, got %q", got)
	}

	generated := ExtractOrGenerate("")
	if len(generated) != 32 {
		t.Errorf("expected generated trace ID with 32 chars, got %d", len(generated))
	}
}

