// Package interfaces define as interfaces principais para o sistema de logging.
// Seguindo os princípios da Arquitetura Hexagonal, estas interfaces definem
// as portas (ports) do sistema de logging.
package interfaces

import (
	"context"
	"io"
	"time"
)

// Level representa os níveis de log disponíveis
type Level int8

const (
	TraceLevel Level = iota - 2
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

// String retorna a representação em string do nível
func (l Level) String() string {
	switch l {
	case TraceLevel:
		return "TRACE"
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

// Format representa os formatos de output disponíveis
type Format int

const (
	JSONFormat Format = iota
	TextFormat
	ConsoleFormat
)

// String retorna a representação em string do formato
func (f Format) String() string {
	switch f {
	case JSONFormat:
		return "json"
	case TextFormat:
		return "text"
	case ConsoleFormat:
		return "console"
	default:
		return "unknown"
	}
}

// Field representa um campo estruturado do log
type Field struct {
	Key   string
	Value interface{}
	Type  FieldType
}

// FieldType representa o tipo do campo para otimização
type FieldType int

const (
	StringType FieldType = iota
	IntType
	Int64Type
	Float64Type
	BoolType
	TimeType
	DurationType
	ErrorType
	ObjectType
	ArrayType
)

// Entry representa uma entrada de log completa
type Entry struct {
	Level      Level
	Message    string
	Fields     []Field
	Time       time.Time
	Caller     *Caller
	Context    context.Context
	TraceID    string
	SpanID     string
	StackTrace string
}

// Caller representa informações sobre quem chamou o log
type Caller struct {
	File     string
	Line     int
	Function string
}

// Logger interface principal do sistema de logging
// Define as operações básicas de logging com suporte a contexto
type Logger interface {
	// Métodos básicos de logging
	Trace(ctx context.Context, msg string, fields ...Field)
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field)
	Panic(ctx context.Context, msg string, fields ...Field)

	// Métodos com formatação
	Tracef(ctx context.Context, format string, args ...interface{})
	Debugf(ctx context.Context, format string, args ...interface{})
	Infof(ctx context.Context, format string, args ...interface{})
	Warnf(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
	Fatalf(ctx context.Context, format string, args ...interface{})
	Panicf(ctx context.Context, format string, args ...interface{})

	// Métodos com código de erro/evento
	TraceWithCode(ctx context.Context, code, msg string, fields ...Field)
	DebugWithCode(ctx context.Context, code, msg string, fields ...Field)
	InfoWithCode(ctx context.Context, code, msg string, fields ...Field)
	WarnWithCode(ctx context.Context, code, msg string, fields ...Field)
	ErrorWithCode(ctx context.Context, code, msg string, fields ...Field)

	// Métodos utilitários
	WithFields(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
	WithError(err error) Logger
	WithTraceID(traceID string) Logger
	WithSpanID(spanID string) Logger

	// Configuração
	SetLevel(level Level)
	GetLevel() Level
	IsLevelEnabled(level Level) bool

	// Clonagem
	Clone() Logger

	// Lifecycle
	Flush() error
	Close() error
}

// Provider interface para implementações específicas de cada biblioteca de logging
// Esta é a porta (port) para os adaptadores (adapters) externos
type Provider interface {
	Logger

	// Configuração
	Configure(config Config) error

	// Informações do provider
	Name() string
	Version() string

	// Health check
	HealthCheck() error
}

// Config representa a configuração completa do sistema de logging
type Config struct {
	// Configurações básicas
	Level      Level     `json:"level" yaml:"level"`
	Format     Format    `json:"format" yaml:"format"`
	Output     io.Writer `json:"-" yaml:"-"`
	TimeFormat string    `json:"time_format" yaml:"time_format"`

	// Informações do serviço
	ServiceName    string `json:"service_name" yaml:"service_name"`
	ServiceVersion string `json:"service_version" yaml:"service_version"`
	Environment    string `json:"environment" yaml:"environment"`

	// Configurações de contexto
	AddSource     bool `json:"add_source" yaml:"add_source"`
	AddStacktrace bool `json:"add_stacktrace" yaml:"add_stacktrace"`
	AddCaller     bool `json:"add_caller" yaml:"add_caller"`

	// Campos globais
	GlobalFields map[string]interface{} `json:"global_fields" yaml:"global_fields"`

	// Configurações avançadas
	Async       *AsyncConfig    `json:"async" yaml:"async"`
	Sampling    *SamplingConfig `json:"sampling" yaml:"sampling"`
	Hooks       []Hook          `json:"-" yaml:"-"`
	Middlewares []Middleware    `json:"-" yaml:"-"`

	// Configurações de performance
	BufferSize    int           `json:"buffer_size" yaml:"buffer_size"`
	FlushInterval time.Duration `json:"flush_interval" yaml:"flush_interval"`

	// Métricas
	EnableMetrics bool   `json:"enable_metrics" yaml:"enable_metrics"`
	MetricsPrefix string `json:"metrics_prefix" yaml:"metrics_prefix"`
}

// AsyncConfig configuração para logging assíncrono
type AsyncConfig struct {
	Enabled       bool          `json:"enabled" yaml:"enabled"`
	BufferSize    int           `json:"buffer_size" yaml:"buffer_size"`
	FlushInterval time.Duration `json:"flush_interval" yaml:"flush_interval"`
	Workers       int           `json:"workers" yaml:"workers"`
	DropOnFull    bool          `json:"drop_on_full" yaml:"drop_on_full"`
}

// SamplingConfig configuração de sampling para controle de volume
type SamplingConfig struct {
	Enabled    bool          `json:"enabled" yaml:"enabled"`
	Initial    int           `json:"initial" yaml:"initial"`
	Thereafter int           `json:"thereafter" yaml:"thereafter"`
	Tick       time.Duration `json:"tick" yaml:"tick"`
	Levels     []Level       `json:"levels" yaml:"levels"`
}

// Hook interface para interceptação de logs
type Hook interface {
	Fire(entry *Entry) error
	Levels() []Level
}

// Middleware interface para transformação de logs
type Middleware interface {
	Process(entry *Entry) *Entry
}

// MetricsCollector interface para coleta de métricas de logging
type MetricsCollector interface {
	IncrementCounter(name string, tags map[string]string)
	RecordHistogram(name string, value float64, tags map[string]string)
	RecordGauge(name string, value float64, tags map[string]string)
}

// Factory interface para criação de loggers
type Factory interface {
	CreateLogger(name string, config Config) (Logger, error)
	CreateProvider(providerType string, config Config) (Provider, error)
	RegisterProvider(name string, provider Provider)
	GetProvider(name string) (Provider, bool)
	ListProviders() []string
}

// Repository interface para persistência de configurações
type Repository interface {
	SaveConfig(name string, config Config) error
	LoadConfig(name string) (Config, error)
	DeleteConfig(name string) error
	ListConfigs() ([]string, error)
}
