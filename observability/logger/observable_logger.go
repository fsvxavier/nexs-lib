package logger

import (
	"context"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger/interfaces"
)

// observableLogger implementação do logger com observabilidade
type observableLogger struct {
	provider         interfaces.Provider
	metricsCollector interfaces.MetricsCollector
	hookManager      interfaces.HookManager
	config           *interfaces.Config
}

// NewObservableLogger cria um novo logger observável
func NewObservableLogger(provider interfaces.Provider) interfaces.ObservableLogger {
	metricsCollector := NewMetricsCollector()
	hookManager := NewHookManager()

	// Registra hook de métricas automaticamente
	metricsHook := NewMetricsHook(metricsCollector)
	hookManager.RegisterHook(interfaces.AfterHook, metricsHook)

	return &observableLogger{
		provider:         provider,
		metricsCollector: metricsCollector,
		hookManager:      hookManager,
	}
}

// ConfigureObservableLogger configura um provider existente para ser observável
func ConfigureObservableLogger(provider interfaces.Provider, config *interfaces.Config) interfaces.ObservableLogger {
	logger := NewObservableLogger(provider)

	if config != nil {
		provider.Configure(config)
		logger.(*observableLogger).config = config
	}

	return logger
}

// createLogEntry cria uma entrada de log padronizada
func (o *observableLogger) createLogEntry(ctx context.Context, level interfaces.Level, msg string, fields []interfaces.Field, code string) *interfaces.LogEntry {
	entry := &interfaces.LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Fields:    make(map[string]any),
		Context:   ctx,
		Code:      code,
	}

	// Adiciona campos estruturados
	for _, field := range fields {
		entry.Fields[field.Key] = field.Value
	}

	// Calcula tamanho estimado da entrada
	entry.Size = int64(len(msg) + len(code))
	for k := range entry.Fields {
		entry.Size += int64(len(k) + 20) // 20 bytes estimado por valor
	}

	return entry
}

// executeLog executa o log com hooks e métricas
func (o *observableLogger) executeLog(ctx context.Context, level interfaces.Level, msg string, fields []interfaces.Field, code string) {
	start := time.Now()

	// Cria entrada de log
	entry := o.createLogEntry(ctx, level, msg, fields, code)

	// Executa hooks "before"
	if err := o.hookManager.ExecuteBeforeHooks(ctx, entry); err != nil {
		// Log de erro do hook, mas não interrompe o log principal
		o.metricsCollector.RecordError(err)
		return
	}

	// Executa o log no provider
	switch level {
	case interfaces.DebugLevel:
		o.provider.Debug(ctx, entry.Message, fields...)
	case interfaces.InfoLevel:
		o.provider.Info(ctx, entry.Message, fields...)
	case interfaces.WarnLevel:
		o.provider.Warn(ctx, entry.Message, fields...)
	case interfaces.ErrorLevel:
		if code != "" {
			o.provider.ErrorWithCode(ctx, code, entry.Message, fields...)
		} else {
			o.provider.Error(ctx, entry.Message, fields...)
		}
	case interfaces.FatalLevel:
		o.provider.Fatal(ctx, entry.Message, fields...)
	case interfaces.PanicLevel:
		o.provider.Panic(ctx, entry.Message, fields...)
	}

	// Registra métricas de tempo
	duration := time.Since(start)
	o.metricsCollector.RecordLog(level, duration)

	// Executa hooks "after"
	if err := o.hookManager.ExecuteAfterHooks(ctx, entry); err != nil {
		o.metricsCollector.RecordError(err)
	}
}

// Debug log com nível debug
func (o *observableLogger) Debug(ctx context.Context, msg string, fields ...interfaces.Field) {
	o.executeLog(ctx, interfaces.DebugLevel, msg, fields, "")
}

// Info log com nível info
func (o *observableLogger) Info(ctx context.Context, msg string, fields ...interfaces.Field) {
	o.executeLog(ctx, interfaces.InfoLevel, msg, fields, "")
}

// Warn log com nível warn
func (o *observableLogger) Warn(ctx context.Context, msg string, fields ...interfaces.Field) {
	o.executeLog(ctx, interfaces.WarnLevel, msg, fields, "")
}

// Error log com nível error
func (o *observableLogger) Error(ctx context.Context, msg string, fields ...interfaces.Field) {
	o.executeLog(ctx, interfaces.ErrorLevel, msg, fields, "")
}

// Fatal log com nível fatal
func (o *observableLogger) Fatal(ctx context.Context, msg string, fields ...interfaces.Field) {
	o.executeLog(ctx, interfaces.FatalLevel, msg, fields, "")
}

// Panic log com nível panic
func (o *observableLogger) Panic(ctx context.Context, msg string, fields ...interfaces.Field) {
	o.executeLog(ctx, interfaces.PanicLevel, msg, fields, "")
}

// Debugf log formatado com nível debug
func (o *observableLogger) Debugf(ctx context.Context, format string, args ...any) {
	o.provider.Debugf(ctx, format, args...)
	o.metricsCollector.RecordLog(interfaces.DebugLevel, 0)
}

// Infof log formatado com nível info
func (o *observableLogger) Infof(ctx context.Context, format string, args ...any) {
	o.provider.Infof(ctx, format, args...)
	o.metricsCollector.RecordLog(interfaces.InfoLevel, 0)
}

// Warnf log formatado com nível warn
func (o *observableLogger) Warnf(ctx context.Context, format string, args ...any) {
	o.provider.Warnf(ctx, format, args...)
	o.metricsCollector.RecordLog(interfaces.WarnLevel, 0)
}

// Errorf log formatado com nível error
func (o *observableLogger) Errorf(ctx context.Context, format string, args ...any) {
	o.provider.Errorf(ctx, format, args...)
	o.metricsCollector.RecordLog(interfaces.ErrorLevel, 0)
}

// Fatalf log formatado com nível fatal
func (o *observableLogger) Fatalf(ctx context.Context, format string, args ...any) {
	o.provider.Fatalf(ctx, format, args...)
	o.metricsCollector.RecordLog(interfaces.FatalLevel, 0)
}

// Panicf log formatado com nível panic
func (o *observableLogger) Panicf(ctx context.Context, format string, args ...any) {
	o.provider.Panicf(ctx, format, args...)
	o.metricsCollector.RecordLog(interfaces.PanicLevel, 0)
}

// DebugWithCode log com código de debug
func (o *observableLogger) DebugWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	o.executeLog(ctx, interfaces.DebugLevel, msg, fields, code)
}

// InfoWithCode log com código de info
func (o *observableLogger) InfoWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	o.executeLog(ctx, interfaces.InfoLevel, msg, fields, code)
}

// WarnWithCode log com código de warn
func (o *observableLogger) WarnWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	o.executeLog(ctx, interfaces.WarnLevel, msg, fields, code)
}

// ErrorWithCode log com código de error
func (o *observableLogger) ErrorWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
	o.executeLog(ctx, interfaces.ErrorLevel, msg, fields, code)
}

// WithFields cria logger com campos adicionais
func (o *observableLogger) WithFields(fields ...interfaces.Field) interfaces.Logger {
	newProvider := o.provider.WithFields(fields...)
	return &observableLogger{
		provider:         newProvider.(interfaces.Provider),
		metricsCollector: o.metricsCollector,
		hookManager:      o.hookManager,
		config:           o.config,
	}
}

// WithContext cria logger com contexto
func (o *observableLogger) WithContext(ctx context.Context) interfaces.Logger {
	newProvider := o.provider.WithContext(ctx)
	return &observableLogger{
		provider:         newProvider.(interfaces.Provider),
		metricsCollector: o.metricsCollector,
		hookManager:      o.hookManager,
		config:           o.config,
	}
}

// SetLevel define o nível de log
func (o *observableLogger) SetLevel(level interfaces.Level) {
	o.provider.SetLevel(level)
}

// GetLevel retorna o nível atual
func (o *observableLogger) GetLevel() interfaces.Level {
	return o.provider.GetLevel()
}

// Clone cria uma cópia do logger
func (o *observableLogger) Clone() interfaces.Logger {
	clonedProvider := o.provider.Clone()
	return &observableLogger{
		provider:         clonedProvider.(interfaces.Provider),
		metricsCollector: o.metricsCollector,
		hookManager:      o.hookManager,
		config:           o.config,
	}
}

// Close fecha o logger
func (o *observableLogger) Close() error {
	return o.provider.Close()
}

// GetMetrics retorna as métricas coletadas
func (o *observableLogger) GetMetrics() interfaces.Metrics {
	return o.metricsCollector.GetMetrics()
}

// GetMetricsCollector retorna o coletor de métricas
func (o *observableLogger) GetMetricsCollector() interfaces.MetricsCollector {
	return o.metricsCollector
}

// GetHookManager retorna o gerenciador de hooks
func (o *observableLogger) GetHookManager() interfaces.HookManager {
	return o.hookManager
}

// RegisterHook registra um novo hook
func (o *observableLogger) RegisterHook(hookType interfaces.HookType, hook interfaces.Hook) error {
	return o.hookManager.RegisterHook(hookType, hook)
}

// UnregisterHook remove um hook
func (o *observableLogger) UnregisterHook(hookType interfaces.HookType, name string) error {
	return o.hookManager.UnregisterHook(hookType, name)
}
