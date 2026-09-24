package logging

import (
	"broadcast-platform/internal/platform/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with additional context
type Logger struct {
	*zap.Logger
	module string
}

// New creates a new logger with the given configuration
func New(cfg config.LoggingConfig) (*Logger, error) {
	level := parseLevel(cfg.Level)

	var zapConfig zap.Config
	if cfg.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zapConfig.Level = zap.NewAtomicLevelAt(level)
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapLogger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{
		Logger: zapLogger,
		module: "platform",
	}, nil
}

// WithModule creates a new logger with the specified module name
func (l *Logger) WithModule(module string) *Logger {
	return &Logger{
		Logger: l.Logger.With(zap.String("module", module)),
		module: module,
	}
}

// WithTraceID creates a new logger with the specified trace ID
func (l *Logger) WithTraceID(traceID string) *Logger {
	return &Logger{
		Logger: l.Logger.With(zap.String("trace_id", traceID)),
		module: l.module,
	}
}

// WithOperationID creates a new logger with the specified operation ID
func (l *Logger) WithOperationID(operationID string) *Logger {
	return &Logger{
		Logger: l.Logger.With(zap.String("operation_id", operationID)),
		module: l.module,
	}
}

// WithExecutionID creates a new logger with the specified execution ID
func (l *Logger) WithExecutionID(executionID string) *Logger {
	return &Logger{
		Logger: l.Logger.With(zap.String("execution_id", executionID)),
		module: l.module,
	}
}

// WithAgentSessionID creates a new logger with the specified agent session ID
func (l *Logger) WithAgentSessionID(sessionID string) *Logger {
	return &Logger{
		Logger: l.Logger.With(zap.String("agent_session_id", sessionID)),
		module: l.module,
	}
}

// parseLevel converts string level to zapcore.Level
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
