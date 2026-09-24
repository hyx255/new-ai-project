package httpserver

import (
	"net/http"
	"time"

	"broadcast-platform/internal/platform/config"
	"broadcast-platform/internal/platform/logging"
	"broadcast-platform/internal/platform/tracing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	HeaderXRequestID = "X-Request-ID"
	HeaderXTraceID   = "X-Trace-ID"
)

// RequestIDMiddleware generates or propagates a request/trace ID.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(HeaderXTraceID)
		traceID = tracing.ExtractOrGenerate(traceID)
		requestID := tracing.NewTraceID()
		ctx := tracing.WithTraceID(c.Request.Context(), traceID)
		ctx = tracing.WithRequestID(ctx, requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Header(HeaderXTraceID, traceID)
		c.Header(HeaderXRequestID, requestID)
		c.Next()
	}
}

// AccessLogMiddleware logs HTTP requests in a structured format.
func AccessLogMiddleware(logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		traceID := tracing.GetTraceID(c.Request.Context())
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.String("trace_id", traceID),
		}
		if status >= 500 {
			logger.Error("request completed with server error", fields...)
		} else if status >= 400 {
			logger.Warn("request completed with client error", fields...)
		} else {
			logger.Info("request completed", fields...)
		}
	}
}

// RecoveryMiddleware provides enhanced panic recovery with logging.
func RecoveryMiddleware(logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				traceID := tracing.GetTraceID(c.Request.Context())
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("trace_id", traceID),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    1999,
					"message": "internal server error",
				})
			}
		}()
		c.Next()
	}
}

// CORSMiddleware creates a CORS middleware based on configuration.
func CORSMiddleware(cfg config.CORSConfig) gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins: cfg.AllowOrigins,
		AllowMethods: cfg.AllowMethods,
		AllowHeaders: cfg.AllowHeaders,
	}
	return cors.New(corsConfig)
}

