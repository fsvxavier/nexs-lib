package logger

import (
	"context"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

// Funções globais de conveniência que delegam para o logger atual

// Trace logs a trace-level message
func Trace(ctx context.Context, msg string, fields ...interfaces.Field) {
	GetCurrentLogger().Trace(ctx, msg, fields...)
}

// Debug logs a debug-level message
func Debug(ctx context.Context, msg string, fields ...interfaces.Field) {
	GetCurrentLogger().Debug(ctx, msg, fields...)
}

// Info logs an info-level message
func Info(ctx context.Context, msg string, fields ...interfaces.Field) {
	GetCurrentLogger().Info(ctx, msg, fields...)
}

// Warn logs a warning-level message
func Warn(ctx context.Context, msg string, fields ...interfaces.Field) {
	GetCurrentLogger().Warn(ctx, msg, fields...)
}

// Error logs an error-level message
func Error(ctx context.Context, msg string, fields ...interfaces.Field) {
	GetCurrentLogger().Error(ctx, msg, fields...)
}

// Fatal logs a fatal-level message and exits
func Fatal(ctx context.Context, msg string, fields ...interfaces.Field) {
	GetCurrentLogger().Fatal(ctx, msg, fields...)
}

// Panic logs a panic-level message and panics
func Panic(ctx context.Context, msg string, fields ...interfaces.Field) {
	GetCurrentLogger().Panic(ctx, msg, fields...)
}

// Formatted logging functions

// Tracef logs a trace-level formatted message
func Tracef(ctx context.Context, format string, args ...interface{}) {
	GetCurrentLogger().Tracef(ctx, format, args...)
}

// Debugf logs a debug-level formatted message
func Debugf(ctx context.Context, format string, args ...interface{}) {
	GetCurrentLogger().Debugf(ctx, format, args...)
}

// Infof logs an info-level formatted message
func Infof(ctx context.Context, format string, args ...interface{}) {
	GetCurrentLogger().Infof(ctx, format, args...)
}

// Warnf logs a warning-level formatted message
func Warnf(ctx context.Context, format string, args ...interface{}) {
	GetCurrentLogger().Warnf(ctx, format, args...)
}

// Errorf logs an error-level formatted message
func Errorf(ctx context.Context, format string, args ...interface{}) {
	GetCurrentLogger().Errorf(ctx, format, args...)
}

// Fatalf logs a fatal-level formatted message and exits
func Fatalf(ctx context.Context, format string, args ...interface{}) {
	GetCurrentLogger().Fatalf(ctx, format, args...)
}

// Panicf logs a panic-level formatted message and panics
func Panicf(ctx context.Context, format string, args ...interface{}) {
	GetCurrentLogger().Panicf(ctx, format, args...)
}

// Code-based logging functions

// TraceWithCode logs a trace-level message with an event code
func TraceWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	GetCurrentLogger().TraceWithCode(ctx, code, msg, fields...)
}

// DebugWithCode logs a debug-level message with an event code
func DebugWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	GetCurrentLogger().DebugWithCode(ctx, code, msg, fields...)
}

// InfoWithCode logs an info-level message with an event code
func InfoWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	GetCurrentLogger().InfoWithCode(ctx, code, msg, fields...)
}

// WarnWithCode logs a warning-level message with an event code
func WarnWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	GetCurrentLogger().WarnWithCode(ctx, code, msg, fields...)
}

// ErrorWithCode logs an error-level message with an event code
func ErrorWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	GetCurrentLogger().ErrorWithCode(ctx, code, msg, fields...)
}

// Utility functions

// WithFields returns a logger with the given fields
func WithFields(fields ...interfaces.Field) interfaces.Logger {
	return GetCurrentLogger().WithFields(fields...)
}

// WithContext returns a logger with context information
func WithContext(ctx context.Context) interfaces.Logger {
	return GetCurrentLogger().WithContext(ctx)
}

// WithError returns a logger with an error field
func WithError(err error) interfaces.Logger {
	return GetCurrentLogger().WithError(err)
}

// WithTraceID returns a logger with a trace ID
func WithTraceID(traceID string) interfaces.Logger {
	return GetCurrentLogger().WithTraceID(traceID)
}

// WithSpanID returns a logger with a span ID
func WithSpanID(spanID string) interfaces.Logger {
	return GetCurrentLogger().WithSpanID(spanID)
}

// SetLevel sets the logging level for the current logger
func SetLevel(level interfaces.Level) {
	GetCurrentLogger().SetLevel(level)
}

// GetLevel returns the current logging level
func GetLevel() interfaces.Level {
	return GetCurrentLogger().GetLevel()
}

// IsLevelEnabled checks if a log level is enabled
func IsLevelEnabled(level interfaces.Level) bool {
	return GetCurrentLogger().IsLevelEnabled(level)
}

// Flush forces any buffered log entries to be written
func Flush() error {
	return GetCurrentLogger().Flush()
}

// Close closes the logger and releases resources
func Close() error {
	return GetCurrentLogger().Close()
}

// Clone creates a copy of the current logger
func Clone() interfaces.Logger {
	return GetCurrentLogger().Clone()
}

// SetCurrentLogger sets the global logger instance
func SetCurrentLogger(logger interfaces.Logger) {
	globalManager.SetCurrentLogger(logger)
}

// Convenience functions for common field types

// String creates a string field
func String(key, value string) interfaces.Field {
	return interfaces.String(key, value)
}

// Int creates an int field
func Int(key string, value int) interfaces.Field {
	return interfaces.Int(key, value)
}

// Int64 creates an int64 field
func Int64(key string, value int64) interfaces.Field {
	return interfaces.Int64(key, value)
}

// Float64 creates a float64 field
func Float64(key string, value float64) interfaces.Field {
	return interfaces.Float64(key, value)
}

// Bool creates a bool field
func Bool(key string, value bool) interfaces.Field {
	return interfaces.Bool(key, value)
}

// Time creates a time field
func Time(key string, value time.Time) interfaces.Field {
	return interfaces.Time(key, value)
}

// Duration creates a duration field
func Duration(key string, value time.Duration) interfaces.Field {
	return interfaces.Duration(key, value)
}

// Error creates an error field
func Err(err error) interfaces.Field {
	return interfaces.Error(err)
}

// ErrorNamed creates a named error field
func ErrorNamed(key string, err error) interfaces.Field {
	return interfaces.ErrorNamed(key, err)
}

// Object creates an object field
func Object(key string, value interface{}) interfaces.Field {
	return interfaces.Object(key, value)
}

// Array creates an array field
func Array(key string, value interface{}) interfaces.Field {
	return interfaces.Array(key, value)
}

// Common predefined fields

// TraceID creates a trace_id field
func TraceID(value string) interfaces.Field {
	return interfaces.TraceID(value)
}

// SpanID creates a span_id field
func SpanID(value string) interfaces.Field {
	return interfaces.SpanID(value)
}

// UserID creates a user_id field
func UserID(value string) interfaces.Field {
	return interfaces.UserID(value)
}

// RequestID creates a request_id field
func RequestID(value string) interfaces.Field {
	return interfaces.RequestID(value)
}

// CorrelationID creates a correlation_id field
func CorrelationID(value string) interfaces.Field {
	return interfaces.CorrelationID(value)
}

// Method creates an HTTP method field
func Method(value string) interfaces.Field {
	return interfaces.Method(value)
}

// Path creates an HTTP path field
func Path(value string) interfaces.Field {
	return interfaces.Path(value)
}

// StatusCode creates an HTTP status code field
func StatusCode(value int) interfaces.Field {
	return interfaces.StatusCode(value)
}

// Latency creates a latency field
func Latency(value time.Duration) interfaces.Field {
	return interfaces.Latency(value)
}

// Operation creates an operation field
func Operation(value string) interfaces.Field {
	return interfaces.Operation(value)
}

// Component creates a component field
func Component(value string) interfaces.Field {
	return interfaces.Component(value)
}

// Version creates a version field
func Version(value string) interfaces.Field {
	return interfaces.Version(value)
}

// Environment creates an environment field
func Environment(value string) interfaces.Field {
	return interfaces.Environment(value)
}
