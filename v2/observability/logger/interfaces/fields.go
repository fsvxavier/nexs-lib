package interfaces

import (
	"errors"
	"time"
)

// Field creation helper functions for type-safe logging

// String cria um campo de string
func String(key, value string) Field {
	return Field{
		Key:   key,
		Value: value,
		Type:  StringType,
	}
}

// Int cria um campo de int
func Int(key string, value int) Field {
	return Field{
		Key:   key,
		Value: value,
		Type:  IntType,
	}
}

// Int64 cria um campo de int64
func Int64(key string, value int64) Field {
	return Field{
		Key:   key,
		Value: value,
		Type:  Int64Type,
	}
}

// Float64 cria um campo de float64
func Float64(key string, value float64) Field {
	return Field{
		Key:   key,
		Value: value,
		Type:  Float64Type,
	}
}

// Bool cria um campo de bool
func Bool(key string, value bool) Field {
	return Field{
		Key:   key,
		Value: value,
		Type:  BoolType,
	}
}

// Time cria um campo de time.Time
func Time(key string, value time.Time) Field {
	return Field{
		Key:   key,
		Value: value,
		Type:  TimeType,
	}
}

// Duration cria um campo de time.Duration
func Duration(key string, value time.Duration) Field {
	return Field{
		Key:   key,
		Value: value,
		Type:  DurationType,
	}
}

// Error cria um campo de error
func Error(err error) Field {
	if err == nil {
		return Field{
			Key:   "error",
			Value: nil,
			Type:  ErrorType,
		}
	}
	return Field{
		Key:   "error",
		Value: err.Error(),
		Type:  ErrorType,
	}
}

// ErrorNamed cria um campo de error com nome customizado
func ErrorNamed(key string, err error) Field {
	if err == nil {
		return Field{
			Key:   key,
			Value: nil,
			Type:  ErrorType,
		}
	}
	return Field{
		Key:   key,
		Value: err.Error(),
		Type:  ErrorType,
	}
}

// Object cria um campo de objeto
func Object(key string, value interface{}) Field {
	return Field{
		Key:   key,
		Value: value,
		Type:  ObjectType,
	}
}

// Array cria um campo de array
func Array(key string, value interface{}) Field {
	return Field{
		Key:   key,
		Value: value,
		Type:  ArrayType,
	}
}

// Common predefined fields

// TraceID cria um campo para trace ID
func TraceID(value string) Field {
	return String("trace_id", value)
}

// SpanID cria um campo para span ID
func SpanID(value string) Field {
	return String("span_id", value)
}

// UserID cria um campo para user ID
func UserID(value string) Field {
	return String("user_id", value)
}

// RequestID cria um campo para request ID
func RequestID(value string) Field {
	return String("request_id", value)
}

// CorrelationID cria um campo para correlation ID
func CorrelationID(value string) Field {
	return String("correlation_id", value)
}

// Method cria um campo para HTTP method
func Method(value string) Field {
	return String("method", value)
}

// Path cria um campo para HTTP path
func Path(value string) Field {
	return String("path", value)
}

// StatusCode cria um campo para HTTP status code
func StatusCode(value int) Field {
	return Int("status_code", value)
}

// Latency cria um campo para latência
func Latency(value time.Duration) Field {
	return Duration("latency", value)
}

// Operation cria um campo para operação
func Operation(value string) Field {
	return String("operation", value)
}

// Component cria um campo para componente
func Component(value string) Field {
	return String("component", value)
}

// Version cria um campo para versão
func Version(value string) Field {
	return String("version", value)
}

// Environment cria um campo para ambiente
func Environment(value string) Field {
	return String("environment", value)
}

// Validation functions

// ValidateLevel verifica se o nível é válido
func ValidateLevel(level Level) error {
	if level < TraceLevel || level > PanicLevel {
		return errors.New("invalid log level")
	}
	return nil
}

// ValidateFormat verifica se o formato é válido
func ValidateFormat(format Format) error {
	if format < JSONFormat || format > ConsoleFormat {
		return errors.New("invalid log format")
	}
	return nil
}

// ValidateConfig valida uma configuração
func ValidateConfig(config Config) error {
	if err := ValidateLevel(config.Level); err != nil {
		return err
	}

	if err := ValidateFormat(config.Format); err != nil {
		return err
	}

	if config.ServiceName == "" {
		return errors.New("service name is required")
	}

	// Valida configurações diretas de buffer e flush
	if config.BufferSize < 0 {
		return errors.New("buffer size cannot be negative")
	}

	if config.FlushInterval < 0 {
		return errors.New("flush interval cannot be negative")
	}

	if config.Async != nil {
		if config.Async.BufferSize < 0 {
			return errors.New("async buffer size cannot be negative")
		}
		if config.Async.BufferSize == 0 {
			return errors.New("async buffer size must be positive")
		}
		if config.Async.Workers <= 0 {
			return errors.New("async workers must be positive")
		}
		if config.Async.FlushInterval < 0 {
			return errors.New("async flush interval cannot be negative")
		}
	}

	if config.Sampling != nil {
		if config.Sampling.Initial < 0 {
			return errors.New("sampling initial must be non-negative")
		}
		if config.Sampling.Thereafter < 0 {
			return errors.New("sampling thereafter must be non-negative")
		}
	}

	return nil
}
