package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"broadcast-platform/internal/platform/cache"
	"broadcast-platform/internal/platform/config"
	"broadcast-platform/internal/platform/logging"
	"broadcast-platform/internal/platform/metrics"
	"broadcast-platform/internal/platform/response"
	"broadcast-platform/internal/platform/tracing"

	"github.com/gin-gonic/gin"
)

func newTestServer() *Server {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8080,
			Env:  "test",
		},
		Metrics: config.MetricsConfig{Enabled: false},
		CORS: config.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET"},
			AllowHeaders: []string{"Origin"},
		},
	}
	logger, _ := logging.New(config.LoggingConfig{Level: "error", Format: "console"})
	m := metrics.New(cfg.Metrics)
	c := cache.New(config.CacheConfig{Enabled: false}, logger)

	return New(cfg, logger, nil, m, c, nil)
}

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Code != 0 {
		t.Errorf("expected code 0, got %d", resp.Code)
	}
	if resp.Message != "success" {
		t.Errorf("expected message success, got %s", resp.Message)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be a map")
	}
	if data["status"] != "ok" {
		t.Errorf("expected status ok, got %v", data["status"])
	}
	if data["database"] != "not connected" {
		t.Errorf("expected database not connected, got %v", data["database"])
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	traceID := w.Header().Get(HeaderXTraceID)
	if traceID == "" {
		t.Error("expected X-Trace-ID header to be set")
	}

	requestID := w.Header().Get(HeaderXRequestID)
	if requestID == "" {
		t.Error("expected X-Request-ID header to be set")
	}
}

func TestRequestIDPropagation(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set(HeaderXTraceID, "existing-trace-id")
	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	if got := w.Header().Get(HeaderXTraceID); got != "existing-trace-id" {
		t.Errorf("expected trace ID to be propagated, got %s", got)
	}
}

func TestTracingContextPropagation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var capturedTraceID string
	var capturedRequestID string

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		capturedTraceID = tracing.GetTraceID(c.Request.Context())
		capturedRequestID = tracing.GetRequestID(c.Request.Context())
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedTraceID == "" {
		t.Error("expected trace ID in context")
	}
	if capturedRequestID == "" {
		t.Error("expected request ID in context")
	}
}

func TestNotFoundRoute(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

