// Package tracing provides lightweight trace ID propagation for HTTP requests.
package tracing

import (
	"context"
	"crypto/rand"
	"fmt"
)

type contextKey string

const (
	traceIDKey   contextKey = "trace_id"
	requestIDKey contextKey = "request_id"
)

// NewTraceID generates a new unique trace ID (32 hex chars).
func NewTraceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// WithTraceID stores a trace ID in the context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID extracts the trace ID from the context.
// Returns empty string if not found.
func GetTraceID(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}

// WithRequestID stores a request ID in the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID extracts the request ID from the context.
func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// ExtractOrGenerate retrieves an existing trace ID from the header value,
// or generates a new one if the header is empty.
func ExtractOrGenerate(headerValue string) string {
	if headerValue != "" {
		return headerValue
	}
	return NewTraceID()
}
