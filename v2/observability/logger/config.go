// Package logger implementa o sistema principal de logging v2 da nexs-lib.
// Segue os princípios da Arquitetura Hexagonal e Clean Architecture.
package logger

import (
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

// DefaultConfig retorna uma configuração padrão otimizada
func DefaultConfig() interfaces.Config {
	return interfaces.Config{
		Level:          interfaces.InfoLevel,
		Format:         interfaces.JSONFormat,
		Output:         os.Stdout,
		TimeFormat:     time.RFC3339Nano,
		ServiceName:    "unknown",
		ServiceVersion: "unknown",
		Environment:    "development",
		AddSource:      false,
		AddStacktrace:  false,
		AddCaller:      true,
		GlobalFields:   make(map[string]interface{}),
		BufferSize:     1024,
		FlushInterval:  100 * time.Millisecond,
		EnableMetrics:  false,
		MetricsPrefix:  "logger",
	}
}

// ProductionConfig retorna uma configuração otimizada para produção
func ProductionConfig() interfaces.Config {
	config := DefaultConfig()
	config.Level = interfaces.InfoLevel
	config.Format = interfaces.JSONFormat
	config.Environment = "production"
	config.AddSource = false
	config.AddStacktrace = true
	config.AddCaller = false
	config.EnableMetrics = true

	// Configuração assíncrona para alta performance
	config.Async = &interfaces.AsyncConfig{
		Enabled:       true,
		BufferSize:    4096,
		FlushInterval: 50 * time.Millisecond,
		Workers:       2,
		DropOnFull:    false,
	}

	// Sampling para reduzir volume em produção
	config.Sampling = &interfaces.SamplingConfig{
		Enabled:    true,
		Initial:    100,
		Thereafter: 1000,
		Tick:       time.Second,
		Levels:     []interfaces.Level{interfaces.DebugLevel, interfaces.TraceLevel},
	}

	return config
}

// DevelopmentConfig retorna uma configuração otimizada para desenvolvimento
func DevelopmentConfig() interfaces.Config {
	config := DefaultConfig()
	config.Level = interfaces.DebugLevel
	config.Format = interfaces.ConsoleFormat
	config.Environment = "development"
	config.AddSource = true
	config.AddStacktrace = false
	config.AddCaller = true
	config.EnableMetrics = false

	// Sem async em desenvolvimento para debug mais fácil
	config.Async = nil
	config.Sampling = nil

	return config
}

// TestConfig retorna uma configuração otimizada para testes
func TestConfig() interfaces.Config {
	config := DefaultConfig()
	config.Level = interfaces.DebugLevel
	config.Format = interfaces.JSONFormat
	config.Environment = "test"
	config.AddSource = false
	config.AddStacktrace = false
	config.AddCaller = false
	config.EnableMetrics = false

	// Buffer pequeno para testes
	config.BufferSize = 64
	config.FlushInterval = 10 * time.Millisecond

	return config
}

// EnvironmentConfig cria configuração baseada em variáveis de ambiente
func EnvironmentConfig() interfaces.Config {
	config := DefaultConfig()

	// LOG_LEVEL
	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		switch levelStr {
		case "trace", "TRACE":
			config.Level = interfaces.TraceLevel
		case "debug", "DEBUG":
			config.Level = interfaces.DebugLevel
		case "info", "INFO":
			config.Level = interfaces.InfoLevel
		case "warn", "WARN", "warning", "WARNING":
			config.Level = interfaces.WarnLevel
		case "error", "ERROR":
			config.Level = interfaces.ErrorLevel
		case "fatal", "FATAL":
			config.Level = interfaces.FatalLevel
		case "panic", "PANIC":
			config.Level = interfaces.PanicLevel
		}
	}

	// LOG_FORMAT
	if formatStr := os.Getenv("LOG_FORMAT"); formatStr != "" {
		switch formatStr {
		case "json", "JSON":
			config.Format = interfaces.JSONFormat
		case "text", "TEXT":
			config.Format = interfaces.TextFormat
		case "console", "CONSOLE":
			config.Format = interfaces.ConsoleFormat
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

	// Configurações boooleanas
	if os.Getenv("LOG_ADD_SOURCE") == "true" {
		config.AddSource = true
	}
	if os.Getenv("LOG_ADD_STACKTRACE") == "true" {
		config.AddStacktrace = true
	}
	if os.Getenv("LOG_ADD_CALLER") == "true" {
		config.AddCaller = true
	}
	if os.Getenv("LOG_ENABLE_METRICS") == "true" {
		config.EnableMetrics = true
	}

	// TIME_FORMAT
	if timeFormat := os.Getenv("LOG_TIME_FORMAT"); timeFormat != "" {
		config.TimeFormat = timeFormat
	}

	// Configuração automática baseada no ambiente
	switch config.Environment {
	case "production", "prod":
		prodConfig := ProductionConfig()
		config.Async = prodConfig.Async
		config.Sampling = prodConfig.Sampling
		config.EnableMetrics = true
	case "development", "dev":
		config.Level = interfaces.DebugLevel
		config.Format = interfaces.ConsoleFormat
		config.AddSource = true
		config.AddCaller = true
	case "test", "testing":
		testConfig := TestConfig()
		config.BufferSize = testConfig.BufferSize
		config.FlushInterval = testConfig.FlushInterval
	}

	return config
}

// ConfigBuilder builder pattern para configuração fluente
type ConfigBuilder struct {
	config interfaces.Config
}

// NewConfigBuilder cria um novo builder com configuração padrão
func NewConfigBuilder() *ConfigBuilder {
	config := DefaultConfig()
	return &ConfigBuilder{config: config}
}

// Level define o nível de log
func (b *ConfigBuilder) Level(level interfaces.Level) *ConfigBuilder {
	b.config.Level = level
	return b
}

// Format define o formato de output
func (b *ConfigBuilder) Format(format interfaces.Format) *ConfigBuilder {
	b.config.Format = format
	return b
}

// ServiceName define o nome do serviço
func (b *ConfigBuilder) ServiceName(name string) *ConfigBuilder {
	b.config.ServiceName = name
	return b
}

// ServiceVersion define a versão do serviço
func (b *ConfigBuilder) ServiceVersion(version string) *ConfigBuilder {
	b.config.ServiceVersion = version
	return b
}

// Environment define o ambiente
func (b *ConfigBuilder) Environment(env string) *ConfigBuilder {
	b.config.Environment = env
	return b
}

// AddSource habilita/desabilita source code info
func (b *ConfigBuilder) AddSource(add bool) *ConfigBuilder {
	b.config.AddSource = add
	return b
}

// AddStacktrace habilita/desabilita stack traces
func (b *ConfigBuilder) AddStacktrace(add bool) *ConfigBuilder {
	b.config.AddStacktrace = add
	return b
}

// AddCaller habilita/desabilita caller info
func (b *ConfigBuilder) AddCaller(add bool) *ConfigBuilder {
	b.config.AddCaller = add
	return b
}

// WithGlobalField adiciona um campo global
func (b *ConfigBuilder) WithGlobalField(key string, value interface{}) *ConfigBuilder {
	if b.config.GlobalFields == nil {
		b.config.GlobalFields = make(map[string]interface{})
	}
	b.config.GlobalFields[key] = value
	return b
}

// WithAsync configura logging assíncrono
func (b *ConfigBuilder) WithAsync(bufferSize int, workers int, flushInterval time.Duration) *ConfigBuilder {
	b.config.Async = &interfaces.AsyncConfig{
		Enabled:       true,
		BufferSize:    bufferSize,
		FlushInterval: flushInterval,
		Workers:       workers,
		DropOnFull:    false,
	}
	return b
}

// WithSampling configura sampling
func (b *ConfigBuilder) WithSampling(initial, thereafter int, tick time.Duration, levels ...interfaces.Level) *ConfigBuilder {
	b.config.Sampling = &interfaces.SamplingConfig{
		Enabled:    true,
		Initial:    initial,
		Thereafter: thereafter,
		Tick:       tick,
		Levels:     levels,
	}
	return b
}

// EnableMetrics habilita coleta de métricas
func (b *ConfigBuilder) EnableMetrics(prefix string) *ConfigBuilder {
	b.config.EnableMetrics = true
	b.config.MetricsPrefix = prefix
	return b
}

// Build constrói a configuração final
func (b *ConfigBuilder) Build() interfaces.Config {
	return b.config
}
