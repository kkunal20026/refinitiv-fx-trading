package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Logger wraps zap logger for consistent logging
type Logger struct {
	*zap.Logger
}

// Config represents logger configuration
type Config struct {
	Level      string
	Encoding   string
	OutputPath string
}

// New creates a new logger instance
func New(cfg Config) (*Logger, error) {
	var config zap.Config

	switch cfg.Level {
	case "debug":
		config = zap.NewDevelopmentConfig()
	case "production":
		config = zap.NewProductionConfig()
	default:
		config = zap.NewProductionConfig()
	}

	if cfg.Encoding != "" {
		config.Encoding = cfg.Encoding
	}

	if cfg.OutputPath != "" {
		config.OutputPaths = []string{cfg.OutputPath}
	}

	zapLogger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &Logger{zapLogger}, nil
}

// WithContext adds context to logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract trace ID if present
	if traceID := ctx.Value("trace-id"); traceID != nil {
		return &Logger{l.With(zap.String("trace_id", traceID.(string)))}
	}
	return l
}

// WithFields adds fields to logger
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{l.With(fields...)}
}

// InfoWithContext logs info with context
func (l *Logger) InfoWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Info(msg, fields...)
}

// ErrorWithContext logs error with context
func (l *Logger) ErrorWithContext(ctx context.Context, msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	l.WithContext(ctx).Error(msg, fields...)
}

// WarnWithContext logs warning with context
func (l *Logger) WarnWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Warn(msg, fields...)
}

// DebugWithContext logs debug with context
func (l *Logger) DebugWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Debug(msg, fields...)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
