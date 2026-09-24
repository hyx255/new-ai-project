package metrics

import (
	"net/http"
	"strconv"
	"time"

	"broadcast-platform/internal/platform/config"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all application metrics.
type Metrics struct {
	registry            *prometheus.Registry
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	appInfo             *prometheus.GaugeVec
	enabled             bool
}

// New creates a new Metrics instance with its own registry.
func New(cfg config.MetricsConfig) *Metrics {
	m := &Metrics{
		enabled:  cfg.Enabled,
		registry: prometheus.NewRegistry(),
	}
	if !cfg.Enabled {
		return m
	}

	m.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	m.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	m.appInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_info",
			Help: "Application information",
		},
		[]string{"version", "env"},
	)

	m.registry.MustRegister(m.httpRequestsTotal, m.httpRequestDuration, m.appInfo)
	m.appInfo.WithLabelValues("0.1.0-skeleton", "development").Set(1)

	return m
}

// Handler returns an HTTP handler for the metrics endpoint.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// Middleware returns a Gin middleware that records HTTP metrics.
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.enabled {
			c.Next()
			return
		}
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		c.Next()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		m.httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		m.httpRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}

// IsEnabled returns whether metrics collection is enabled.
func (m *Metrics) IsEnabled() bool {
	return m.enabled
}

