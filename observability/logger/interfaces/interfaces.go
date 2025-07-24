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

	// Métodos de buffer
	GetBuffer() Buffer
	SetBuffer(buffer Buffer) error
	FlushBuffer() error
	GetBufferStats() BufferStats
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
	BufferConfig   *BufferConfig   `json:"buffer" yaml:"buffer"`
}

// SamplingConfig configuração de sampling para logs de alto volume
type SamplingConfig struct {
	Initial    int           `json:"initial" yaml:"initial"`
	Thereafter int           `json:"thereafter" yaml:"thereafter"`
	Tick       time.Duration `json:"tick" yaml:"tick"`
}

// BufferConfig configuração de buffer para alta performance
type BufferConfig struct {
	Enabled      bool          `json:"enabled" yaml:"enabled"`             // Habilita ou desabilita o buffer
	Size         int           `json:"size" yaml:"size"`                   // Tamanho do buffer circular (número de entradas)
	BatchSize    int           `json:"batch_size" yaml:"batch_size"`       // Tamanho do lote para flush
	FlushTimeout time.Duration `json:"flush_timeout" yaml:"flush_timeout"` // Timeout para flush automático
	MemoryLimit  int64         `json:"memory_limit" yaml:"memory_limit"`   // Limite de memória em bytes
	AutoFlush    bool          `json:"auto_flush" yaml:"auto_flush"`       // Habilita flush automático
	ForceSync    bool          `json:"force_sync" yaml:"force_sync"`       // Força sincronização após flush
}

// BufferStats estatísticas do buffer
type BufferStats struct {
	TotalEntries   int64         `json:"total_entries"`
	DroppedEntries int64         `json:"dropped_entries"`
	FlushCount     int64         `json:"flush_count"`
	BufferSize     int           `json:"buffer_size"`
	UsedSize       int           `json:"used_size"`
	LastFlush      time.Time     `json:"last_flush"`
	MemoryUsage    int64         `json:"memory_usage"`
	FlushDuration  time.Duration `json:"flush_duration"`
}

// Buffer interface para gerenciamento de buffer
type Buffer interface {
	// Write adiciona uma entrada no buffer
	Write(entry *LogEntry) error

	// Flush força o flush de todas as entradas pendentes
	Flush() error

	// Close fecha o buffer e faz flush final
	Close() error

	// Stats retorna estatísticas do buffer
	Stats() BufferStats

	// IsFull verifica se o buffer está cheio
	IsFull() bool

	// Size retorna o número atual de entradas no buffer
	Size() int

	// Clear limpa o buffer sem fazer flush
	Clear()
}

// LogEntry representa uma entrada de log para buffer
type LogEntry struct {
	Timestamp time.Time       `json:"timestamp"`
	Level     Level           `json:"level"`
	Message   string          `json:"message"`
	Fields    map[string]any  `json:"fields"`
	Context   context.Context `json:"-"`
	Code      string          `json:"code,omitempty"`
	Source    string          `json:"source,omitempty"`
	Stack     string          `json:"stack,omitempty"`
	Size      int64           `json:"-"` // Tamanho estimado da entrada em bytes
}

// ContextKey tipo para chaves de contexto
type ContextKey string

const (
	TraceIDKey   ContextKey = "trace_id"
	SpanIDKey    ContextKey = "span_id"
	UserIDKey    ContextKey = "user_id"
	RequestIDKey ContextKey = "request_id"
)

// Metrics interface para métricas de logging
type Metrics interface {
	// Contadores por nível
	GetLogCount(level Level) int64
	GetTotalLogCount() int64

	// Tempo de processamento
	GetAverageProcessingTime() time.Duration
	GetProcessingTimeByLevel(level Level) time.Duration

	// Taxa de erro
	GetErrorRate() float64
	GetSamplingRate() float64

	// Estatísticas de provider
	GetProviderStats(provider string) *ProviderStats

	// Reset metrics
	Reset()

	// Export metrics para sistemas externos
	Export() map[string]interface{}
}

// ProviderStats estatísticas específicas de um provider
type ProviderStats struct {
	ProviderName      string          `json:"provider_name"`
	LogCount          map[Level]int64 `json:"log_count"`
	TotalLogs         int64           `json:"total_logs"`
	ErrorCount        int64           `json:"error_count"`
	AverageLatency    time.Duration   `json:"average_latency"`
	LastLogTime       time.Time       `json:"last_log_time"`
	ConfigurationTime time.Time       `json:"configuration_time"`
	BufferStats       *BufferStats    `json:"buffer_stats,omitempty"`
}

// Hook interface para hooks customizados
type Hook interface {
	// Execute é chamado antes ou depois do processamento do log
	Execute(ctx context.Context, entry *LogEntry) error

	// GetName retorna o nome do hook
	GetName() string

	// IsEnabled verifica se o hook está habilitado
	IsEnabled() bool

	// SetEnabled habilita/desabilita o hook
	SetEnabled(enabled bool)
}

// HookType define os tipos de hooks
type HookType string

const (
	BeforeHook HookType = "before" // Executado antes do log
	AfterHook  HookType = "after"  // Executado depois do log
)

// HookManager interface para gerenciamento de hooks
type HookManager interface {
	// Registro de hooks
	RegisterHook(hookType HookType, hook Hook) error
	UnregisterHook(hookType HookType, name string) error

	// Execução de hooks
	ExecuteBeforeHooks(ctx context.Context, entry *LogEntry) error
	ExecuteAfterHooks(ctx context.Context, entry *LogEntry) error

	// Listagem e gerenciamento
	ListHooks(hookType HookType) []Hook
	GetHook(hookType HookType, name string) Hook
	ClearHooks(hookType HookType)

	// Estado dos hooks
	EnableAllHooks()
	DisableAllHooks()
	GetHookCount(hookType HookType) int
}

// MetricsCollector interface para coleta de métricas
type MetricsCollector interface {
	// Coleta de métricas básicas
	RecordLog(level Level, duration time.Duration)
	RecordError(err error)
	RecordSample(sampled bool)

	// Coleta de métricas de provider
	RecordProviderOperation(provider string, operation string, duration time.Duration)
	RecordProviderError(provider string, err error)

	// Coleta de métricas de buffer
	RecordBufferOperation(operation string, size int, duration time.Duration)

	// Getter para métricas
	GetMetrics() Metrics
}

// ObservableLogger interface que combina logging com observabilidade
type ObservableLogger interface {
	Logger

	// Métricas
	GetMetrics() Metrics
	GetMetricsCollector() MetricsCollector

	// Hooks
	GetHookManager() HookManager
	RegisterHook(hookType HookType, hook Hook) error
	UnregisterHook(hookType HookType, name string) error
}
