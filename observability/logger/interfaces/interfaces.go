package interfaces

import (
	"context"
	"time"
)

// Logger define a interface principal para logging
type Logger interface {
	// Métodos com campos estruturados
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field)
	Panic(ctx context.Context, msg string, fields ...Field)

	// Métodos formatados
	Debugf(ctx context.Context, format string, args ...any)
	Infof(ctx context.Context, format string, args ...any)
	Warnf(ctx context.Context, format string, args ...any)
	Errorf(ctx context.Context, format string, args ...any)
	Fatalf(ctx context.Context, format string, args ...any)
	Panicf(ctx context.Context, format string, args ...any)

	// Métodos com códigos de erro/evento
	DebugWithCode(ctx context.Context, code, msg string, fields ...Field)
	InfoWithCode(ctx context.Context, code, msg string, fields ...Field)
	WarnWithCode(ctx context.Context, code, msg string, fields ...Field)
	ErrorWithCode(ctx context.Context, code, msg string, fields ...Field)

	// Métodos utilitários
	WithFields(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
	SetLevel(level Level)
	GetLevel() Level
	Clone() Logger
	Close() error
}

// Provider define a interface para providers de logging
type Provider interface {
	Logger
	Configure(config *Config) error
}

// Field representa um campo estruturado de log
type Field struct {
	Key   string
	Value any
}

// Level representa os níveis de log
type Level int8

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

// String retorna a representação em string do nível
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	case PanicLevel:
		return "PANIC"
	default:
		return "UNKNOWN"
	}
}

// Format representa o formato de saída dos logs
type Format string

const (
	JSONFormat    Format = "json"
	ConsoleFormat Format = "console"
	TextFormat    Format = "text"
)

// Config representa a configuração do logger
type Config struct {
	Level          Level           `json:"level" yaml:"level"`
	Format         Format          `json:"format" yaml:"format"`
	Output         any             `json:"-" yaml:"-"` // io.Writer
	TimeFormat     string          `json:"time_format" yaml:"time_format"`
	ServiceName    string          `json:"service_name" yaml:"service_name"`
	ServiceVersion string          `json:"service_version" yaml:"service_version"`
	Environment    string          `json:"environment" yaml:"environment"`
	AddSource      bool            `json:"add_source" yaml:"add_source"`
	AddStacktrace  bool            `json:"add_stacktrace" yaml:"add_stacktrace"`
	Fields         map[string]any  `json:"fields" yaml:"fields"`
	SamplingConfig *SamplingConfig `json:"sampling" yaml:"sampling"`
}

// SamplingConfig configuração de sampling para logs de alto volume
type SamplingConfig struct {
	Initial    int           `json:"initial" yaml:"initial"`
	Thereafter int           `json:"thereafter" yaml:"thereafter"`
	Tick       time.Duration `json:"tick" yaml:"tick"`
}

// ContextKey tipo para chaves de contexto
type ContextKey string

const (
	TraceIDKey   ContextKey = "trace_id"
	SpanIDKey    ContextKey = "span_id"
	UserIDKey    ContextKey = "user_id"
	RequestIDKey ContextKey = "request_id"
)
