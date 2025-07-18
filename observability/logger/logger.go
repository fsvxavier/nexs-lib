package logger

import (
	"io"
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger/interfaces"
)

// Re-export das principais interfaces e tipos
type (
	Logger         = interfaces.Logger
	Provider       = interfaces.Provider
	Field          = interfaces.Field
	Level          = interfaces.Level
	Format         = interfaces.Format
	Config         = interfaces.Config
	SamplingConfig = interfaces.SamplingConfig
	ContextKey     = interfaces.ContextKey
)

// Re-export das constantes
const (
	DebugLevel Level = interfaces.DebugLevel
	InfoLevel  Level = interfaces.InfoLevel
	WarnLevel  Level = interfaces.WarnLevel
	ErrorLevel Level = interfaces.ErrorLevel
	FatalLevel Level = interfaces.FatalLevel
	PanicLevel Level = interfaces.PanicLevel

	JSONFormat    Format = interfaces.JSONFormat
	ConsoleFormat Format = interfaces.ConsoleFormat
	TextFormat    Format = interfaces.TextFormat

	TraceIDKey   ContextKey = interfaces.TraceIDKey
	SpanIDKey    ContextKey = interfaces.SpanIDKey
	UserIDKey    ContextKey = interfaces.UserIDKey
	RequestIDKey ContextKey = interfaces.RequestIDKey
)

// DefaultConfig retorna uma configuração padrão
func DefaultConfig() *Config {
	return &Config{
		Level:      InfoLevel,
		Format:     JSONFormat,
		Output:     os.Stdout,
		TimeFormat: time.RFC3339,
		Fields:     make(map[string]any),
	}
}

// Funções para criação de campos estruturados
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

// Funções de configuração baseadas em ambiente
func EnvironmentConfig() *Config {
	config := DefaultConfig()

	// LOG_LEVEL
	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		switch levelStr {
		case "debug", "DEBUG":
			config.Level = DebugLevel
		case "info", "INFO":
			config.Level = InfoLevel
		case "warn", "WARN", "warning", "WARNING":
			config.Level = WarnLevel
		case "error", "ERROR":
			config.Level = ErrorLevel
		case "fatal", "FATAL":
			config.Level = FatalLevel
		case "panic", "PANIC":
			config.Level = PanicLevel
		}
	}

	// LOG_FORMAT
	if formatStr := os.Getenv("LOG_FORMAT"); formatStr != "" {
		switch formatStr {
		case "json", "JSON":
			config.Format = JSONFormat
		case "console", "CONSOLE":
			config.Format = ConsoleFormat
		case "text", "TEXT":
			config.Format = TextFormat
		}
	}

	// SERVICE_NAME
	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		config.ServiceName = serviceName
	}

	// SERVICE_VERSION
	if serviceVersion := os.Getenv("SERVICE_VERSION"); serviceVersion != "" {
		config.ServiceVersion = serviceVersion
	}

	// ENVIRONMENT
	if environment := os.Getenv("ENVIRONMENT"); environment != "" {
		config.Environment = environment
	} else if env := os.Getenv("ENV"); env != "" {
		config.Environment = env
	}

	return config
}

// DevelopmentConfig retorna uma configuração otimizada para desenvolvimento
func DevelopmentConfig() *Config {
	return &Config{
		Level:         DebugLevel,
		Format:        ConsoleFormat,
		Output:        os.Stdout,
		TimeFormat:    time.RFC3339,
		AddSource:     true,
		AddStacktrace: false,
		Fields:        make(map[string]any),
	}
}

// ProductionConfig retorna uma configuração otimizada para produção
func ProductionConfig() *Config {
	return &Config{
		Level:         InfoLevel,
		Format:        JSONFormat,
		Output:        os.Stdout,
		TimeFormat:    time.RFC3339,
		AddSource:     false,
		AddStacktrace: true,
		Fields:        make(map[string]any),
		SamplingConfig: &SamplingConfig{
			Initial:    1000,
			Thereafter: 100,
			Tick:       time.Second,
		},
	}
}

// TestingConfig retorna uma configuração otimizada para testes
func TestingConfig() *Config {
	return &Config{
		Level:         DebugLevel,
		Format:        JSONFormat,
		Output:        io.Discard,
		TimeFormat:    time.RFC3339,
		AddSource:     false,
		AddStacktrace: false,
		Fields:        make(map[string]any),
	}
}
