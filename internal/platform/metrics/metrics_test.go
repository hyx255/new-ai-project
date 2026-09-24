package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"broadcast-platform/internal/platform/config"

	"github.com/gin-gonic/gin"
)

func TestNewMetricsEnabled(t *testing.T) {
	cfg := config.MetricsConfig{Enabled: true, Path: "/metrics"}
	m := New(cfg)
	if !m.IsEnabled() {
		t.Error("expected metrics to be enabled")
	}
}

func TestNewMetricsDisabled(t *testing.T) {
	cfg := config.MetricsConfig{Enabled: false}
	m := New(cfg)
	if m.IsEnabled() {
		t.Error("expected metrics to be disabled")
	}
}

func TestMetricsMiddlewareDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.MetricsConfig{Enabled: false}
	m := New(cfg)
	router := gin.New()
	router.Use(m.Middleware())
	router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMetricsMiddlewareEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.MetricsConfig{Enabled: true, Path: "/metrics"}
	m := New(cfg)
	router := gin.New()
	router.Use(m.Middleware())
	router.GET("/test", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMetricsHandler(t *testing.T) {
	cfg := config.MetricsConfig{Enabled: true, Path: "/metrics"}
	m := New(cfg)
	handler := m.Handler()
	if handler == nil {
		t.Error("expected non-nil handler")
	}
}

func TestMultipleMetricsInstances(t *testing.T) {
	// Verify that creating multiple metrics instances does not panic
	cfg := config.MetricsConfig{Enabled: true, Path: "/metrics"}
	m1 := New(cfg)
	m2 := New(cfg)
	if m1 == nil || m2 == nil {
		t.Error("expected non-nil metrics instances")
	}
}

