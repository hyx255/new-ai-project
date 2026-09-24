package bootstrap

import (
	"fmt"
	"time"

	"broadcast-platform/internal/platform/auth"
	"broadcast-platform/internal/platform/cache"
	"broadcast-platform/internal/platform/config"
	"broadcast-platform/internal/platform/database"
	"broadcast-platform/internal/platform/logging"
	"broadcast-platform/internal/platform/metrics"

	httpserver "broadcast-platform/internal/platform/http"

	// Device module
	deviceadapter "broadcast-platform/internal/modules/device/adapter"
	deviceapi "broadcast-platform/internal/modules/device/api"
	deviceengine "broadcast-platform/internal/modules/device/engine"
	devicerepo "broadcast-platform/internal/modules/device/repository"
	devicesvc "broadcast-platform/internal/modules/device/service"

	// User module
	userapi "broadcast-platform/internal/modules/user/api"
	userrepo "broadcast-platform/internal/modules/user/repository"
	usersvc "broadcast-platform/internal/modules/user/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Run is the unified application startup flow.
func Run(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	logger, err := logging.New(cfg.Logging)
	if err != nil {
		return fmt.Errorf("logging: %w", err)
	}
	defer logger.Sync()

	logger.Info("starting broadcast platform",
		zap.String("env", cfg.Server.Env),
		zap.String("address", cfg.Address()),
	)

	db := initDatabase(cfg, logger)
	if db != nil {
		defer db.Close()
	}

	m := metrics.New(cfg.Metrics)
	logger.Info("metrics initialized", zap.Bool("enabled", m.IsEnabled()))

	cacheStore := cache.New(cfg.Cache, logger.WithModule("cache"))
	defer cacheStore.Close()

	// Initialize JWT token provider
	tokenProvider := auth.NewTokenProvider(cfg.Auth.JWTSecret, cfg.Auth.TokenExpiryHours)
	logger.Info("JWT token provider initialized",
		zap.Int("token_expiry_hours", cfg.Auth.TokenExpiryHours),
	)

	// Build global middleware and route registrars
	var globalMiddlewares []httpserver.GlobalMiddlewareRegistrar
	var registrars []httpserver.RouteRegistrar
	var engine *deviceengine.ExecutionEngine

	if db != nil {
		// Initialize User module (needed for auth middleware)
		userRepo := userrepo.NewUserRepository(db)

		// Auth middleware (global)
		globalMiddlewares = append(globalMiddlewares, func(router *gin.Engine) {
			router.Use(userapi.AuthMiddleware(tokenProvider, userRepo))
			router.Use(userapi.RequirePasswordChangedMiddleware())
		})

		// User module registrar
		authService := usersvc.NewAuthService(userRepo, tokenProvider, logger.WithModule("auth"))
		userService := usersvc.NewUserService(userRepo, logger.WithModule("user"))
		authHandler := userapi.NewAuthHandler(authService)
		userHandler := userapi.NewUserHandler(userService)

		registrars = append(registrars, func(router *gin.Engine) {
			userapi.RegisterRoutes(router, authHandler, userHandler)
		})

		// Device module registrar
		var deviceReg httpserver.RouteRegistrar
		deviceReg, engine = initDeviceModule(cfg, db, logger)
		registrars = append(registrars, deviceReg)
	}

	srv := httpserver.New(cfg, logger, db, m, cacheStore, globalMiddlewares, registrars...)

	// Register shutdown hook for execution engine
	if engine != nil {
		shutdownTimeout := time.Duration(cfg.Execution.ShutdownTimeoutSec) * time.Second
		srv.OnShutdown(func() {
			logger.Info("shutting down execution engine")
			engine.Shutdown(shutdownTimeout)
		})
	}

	return srv.Start()
}

// initDeviceModule initializes the Device module and returns a route registrar and the execution engine.
func initDeviceModule(cfg *config.Config, db *database.DB, logger *logging.Logger) (httpserver.RouteRegistrar, *deviceengine.ExecutionEngine) {
	// Repositories
	dtRepo := devicerepo.NewDeviceTypeRepository(db)
	devRepo := devicerepo.NewDeviceRepository(db)
	opRepo := devicerepo.NewOperationRepository(db)
	execRepo := devicerepo.NewExecutionRepository(db)

	// Gateway (Simulator for MVP)
	gw := deviceadapter.NewSimulatorAdapter(logger.WithModule("simulator"))

	// Execution Engine
	eng := deviceengine.NewExecutionEngine(
		cfg.Execution.MaxWorkers,
		opRepo, execRepo, devRepo, gw,
		logger.WithModule("execution-engine"),
	)

	logger.Info("execution engine initialized",
		zap.Int("max_workers", cfg.Execution.MaxWorkers),
		zap.Int("max_batch_size", cfg.Execution.MaxBatchSize),
	)

	// Services
	dtService := devicesvc.NewDeviceTypeService(dtRepo, logger.WithModule("device-type"))
	devService := devicesvc.NewDeviceService(devRepo, dtRepo, logger.WithModule("device"))
	opService := devicesvc.NewOperationService(devRepo, dtRepo, opRepo, execRepo, gw, logger.WithModule("operation"))
	batchService := devicesvc.NewBatchOperationService(
		devRepo, dtRepo, opRepo, execRepo, eng,
		cfg.Execution.MaxBatchSize,
		logger.WithModule("batch-operation"),
	)

	// Handlers
	dtHandler := deviceapi.NewDeviceTypeHandler(dtService)
	devHandler := deviceapi.NewDeviceHandler(devService)
	opHandler := deviceapi.NewOperationHandler(opService)
	batchHandler := deviceapi.NewBatchOperationHandler(batchService)

	registrar := func(router *gin.Engine) {
		deviceapi.RegisterRoutes(router, dtHandler, devHandler, opHandler, batchHandler)
	}

	return registrar, eng
}

func initDatabase(cfg *config.Config, logger *logging.Logger) *database.DB {
	if cfg.Database.DSN == "" {
		logger.Info("database DSN empty, skipping")
		return nil
	}

	db, err := database.New(cfg.Database, logger.WithModule("database"))
	if err != nil {
		logger.Warn("database connection failed, running in degraded mode",
			zap.String("error", err.Error()))
		return nil
	}
	return db
}
