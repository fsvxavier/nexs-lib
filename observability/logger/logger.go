package logger

import (
	"context"
	"io"
	"time"
)

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

var levelNames = map[Level]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
	FatalLevel: "FATAL",
	PanicLevel: "PANIC",
}

func (l Level) String() string {
	if name, ok := levelNames[l]; ok {
		return name
	}
	return "UNKNOWN"
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
	Output         io.Writer       `json:"-" yaml:"-"`
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

// DefaultConfig retorna uma configuração padrão
func DefaultConfig() *Config {
	return &Config{
		Level:         InfoLevel,
		Format:        JSONFormat,
		TimeFormat:    time.RFC3339,
		AddSource:     false,
		AddStacktrace: false,
		Fields:        make(map[string]any),
	}
}

// Field representa um campo estruturado do log
type Field struct {
	Key   string
	Value any
}

// Logger interface principal para logging estruturado
type Logger interface {
	// Métodos básicos de logging
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field)
	Panic(ctx context.Context, msg string, fields ...Field)

	// Métodos com formatação
	Debugf(ctx context.Context, format string, args ...any)
	Infof(ctx context.Context, format string, args ...any)
	Warnf(ctx context.Context, format string, args ...any)
	Errorf(ctx context.Context, format string, args ...any)
	Fatalf(ctx context.Context, format string, args ...any)
	Panicf(ctx context.Context, format string, args ...any)

	// Métodos com código de erro/evento
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
}

// Provider interface para implementação específica de cada biblioteca
type Provider interface {
	Logger
	Configure(config *Config) error
	Close() error
}

// Helper functions para criar fields
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Any(key string, value any) Field {
	return Field{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

func ErrorField(err error) Field {
	return Field{Key: "error", Value: err}
}

func Stack(key string) Field {
	return Field{Key: key, Value: "stack_trace"}
}
