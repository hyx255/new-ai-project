package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"broadcast-platform/internal/platform/cache"
	"broadcast-platform/internal/platform/config"
	"broadcast-platform/internal/platform/database"
	"broadcast-platform/internal/platform/logging"
	"broadcast-platform/internal/platform/metrics"
	"broadcast-platform/internal/platform/response"
	"broadcast-platform/internal/platform/tracing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RouteRegistrar is a function that registers module routes on the router.
type RouteRegistrar func(router *gin.Engine)

// GlobalMiddlewareRegistrar is a function that registers global middleware on the router.
// These are applied before route-specific middleware and module routes.
type GlobalMiddlewareRegistrar func(router *gin.Engine)

// Server is the HTTP server for the broadcast platform.
type Server struct {
	cfg                *config.Config
	logger             *logging.Logger
	router             *gin.Engine
	httpSrv            *http.Server
	db                 *database.DB
	metrics            *metrics.Metrics
	cache              cache.Store
	globalMiddlewares  []GlobalMiddlewareRegistrar
	registrars         []RouteRegistrar
	shutdownCallbacks  []func()
}

// New creates a new Server instance.
func New(
	cfg *config.Config,
	logger *logging.Logger,
	db *database.DB,
	m *metrics.Metrics,
	c cache.Store,
	globalMiddlewares []GlobalMiddlewareRegistrar,
	registrars ...RouteRegistrar,
) *Server {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	s := &Server{
		cfg:               cfg,
		logger:            logger,
		router:            router,
		db:                db,
		metrics:           m,
		cache:             c,
		globalMiddlewares: globalMiddlewares,
		registrars:        registrars,
	}

	s.setupMiddleware()
	s.setupRoutes()

	s.httpSrv = &http.Server{
		Addr:    cfg.Address(),
		Handler: router,
	}

	return s
}

// OnShutdown registers a callback to be executed during graceful shutdown.
// Callbacks run after the HTTP server stops accepting new connections.
func (s *Server) OnShutdown(fn func()) {
	s.shutdownCallbacks = append(s.shutdownCallbacks, fn)
}

func (s *Server) setupMiddleware() {
	s.router.Use(RecoveryMiddleware(s.logger))
	s.router.Use(RequestIDMiddleware())
	s.router.Use(CORSMiddleware(s.cfg.CORS))
	s.router.Use(AccessLogMiddleware(s.logger))
	s.router.Use(s.metrics.Middleware())

	// Apply global middleware (auth, etc.)
	for _, mw := range s.globalMiddlewares {
		mw(s.router)
	}
}

func (s *Server) setupRoutes() {
	s.router.GET("/health", s.healthHandler)

	if s.cfg.Metrics.Enabled {
		s.router.GET(s.cfg.Metrics.Path, gin.WrapH(s.metrics.Handler()))
	}

	if s.cfg.IsDevelopment() {
		debugGroup := s.router.Group("/debug/pprof")
		{
			debugGroup.GET("/", gin.WrapH(http.HandlerFunc(pprof.Index)))
			debugGroup.GET("/cmdline", gin.WrapH(http.HandlerFunc(pprof.Cmdline)))
			debugGroup.GET("/profile", gin.WrapH(http.HandlerFunc(pprof.Profile)))
			debugGroup.GET("/symbol", gin.WrapH(http.HandlerFunc(pprof.Symbol)))
			debugGroup.GET("/trace", gin.WrapH(http.HandlerFunc(pprof.Trace)))
		}
	}

	// Register module routes
	for _, reg := range s.registrars {
		reg(s.router)
	}
}

func (s *Server) healthHandler(c *gin.Context) {
	healthData := map[string]interface{}{
		"status": "ok",
	}

	if s.db != nil {
		if err := s.db.HealthCheck(); err != nil {
			healthData["status"] = "degraded"
			healthData["database"] = "unhealthy"
		} else {
			healthData["database"] = "healthy"
		}
	} else {
		healthData["database"] = "not connected"
	}

	traceID := tracing.GetTraceID(c.Request.Context())
	if traceID != "" {
		healthData["trace_id"] = traceID
	}

	response.Success(c, healthData)
}

// Start starts the HTTP server and blocks until a shutdown signal is received.
func (s *Server) Start() error {
	go func() {
		s.logger.Info("server listening", zap.String("address", s.cfg.Address()))
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("server failed to start", zap.String("error", err.Error()))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	s.logger.Info("received shutdown signal", zap.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTimeout)
	defer cancel()

	s.logger.Info("shutting down server...")
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	// Run shutdown callbacks (execution engine, etc.)
	for _, cb := range s.shutdownCallbacks {
		cb()
	}

	s.logger.Info("server exited gracefully")
	return nil
}

// Router returns the underlying gin.Engine for testing purposes.
func (s *Server) Router() *gin.Engine {
	return s.router
}

// ErrorResponse is a convenience method for sending error responses.
func ErrorResponse(c *gin.Context, err error) {
	response.Error(c, err)
}

// NotFoundResponse sends a standard not found response.
func NotFoundResponse(c *gin.Context) {
	response.NotFound(c, "resource not found")
}
