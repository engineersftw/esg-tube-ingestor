package logger

import (
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with application-specific functionality
type Logger struct {
	zap *zap.Logger
}

// New creates a new logger with specified level and format
func New(level, format string, output io.Writer) (*Logger, error) {
	// Parse log level
	zapLevel, err := parseLevel(level)
	if err != nil {
		return nil, err
	}

	// Parse format
	encoder, err := createEncoder(format)
	if err != nil {
		return nil, err
	}

	// Create writer syncer
	var writeSyncer zapcore.WriteSyncer
	if output == nil {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else {
		writeSyncer = zapcore.AddSync(output)
	}

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, zapLevel)

	// Create logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{zap: zapLogger}, nil
}

// NewFromConfig creates a logger from application configuration
func NewFromConfig(level, format string) (*Logger, error) {
	return New(level, format, nil)
}

// parseLevel converts string log level to zap level
func parseLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", level)
	}
}

// createEncoder creates encoder based on format
func createEncoder(format string) (zapcore.Encoder, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	switch format {
	case "json":
		return zapcore.NewJSONEncoder(encoderConfig), nil
	case "text", "console":
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return zapcore.NewConsoleEncoder(encoderConfig), nil
	default:
		return nil, fmt.Errorf("invalid log format: %s (must be json or text)", format)
	}
}

// WithCorrelationID adds a correlation ID to all log messages
func (l *Logger) WithCorrelationID(correlationID string) *Logger {
	return &Logger{
		zap: l.zap.With(zap.String("correlation_id", correlationID)),
	}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	return &Logger{
		zap: l.zap.With(zapFields...),
	}
}

// WithField adds a single field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		zap: l.zap.With(zap.Any(key, value)),
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

// Zap returns the underlying zap logger for advanced usage
func (l *Logger) Zap() *zap.Logger {
	return l.zap
}
