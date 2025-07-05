package logger

import (
	"context"
	"fmt"
	"os"
	"sync"
)

// LoggerManager gerencia os diferentes providers de logging
type LoggerManager struct {
	providers     map[string]Provider
	current       Logger
	defaultLogger Logger
	mu            sync.RWMutex
}

var globalManager = &LoggerManager{
	providers:     make(map[string]Provider),
	current:       &noopLogger{},
	defaultLogger: &noopLogger{},
}

// RegisterProvider registra um provider de logging
func RegisterProvider(name string, provider Provider) {
	globalManager.mu.Lock()
	defer globalManager.mu.Unlock()
	globalManager.providers[name] = provider
}

// SetProvider define o provider ativo
func SetProvider(name string, config *Config) error {
	globalManager.mu.Lock()
	defer globalManager.mu.Unlock()

	provider, exists := globalManager.providers[name]
	if !exists {
		return fmt.Errorf("provider '%s' not found", name)
	}

	if config == nil {
		config = DefaultConfig()
	}

	if config.Output == nil {
		config.Output = os.Stdout
	}

	if err := provider.Configure(config); err != nil {
		return fmt.Errorf("failed to configure provider '%s': %w", name, err)
	}

	globalManager.current = provider
	globalManager.defaultLogger = provider

	return nil
}

// GetCurrentProvider retorna o provider atual
func GetCurrentProvider() Logger {
	globalManager.mu.RLock()
	defer globalManager.mu.RUnlock()
	return globalManager.current
}

// ListProviders lista todos os providers registrados
func ListProviders() []string {
	globalManager.mu.RLock()
	defer globalManager.mu.RUnlock()

	providers := make([]string, 0, len(globalManager.providers))
	for name := range globalManager.providers {
		providers = append(providers, name)
	}
	return providers
}

// Métodos globais que delegam para o logger atual
func Debug(ctx context.Context, msg string, fields ...Field) {
	globalManager.mu.RLock()
	logger := globalManager.defaultLogger
	globalManager.mu.RUnlock()
	logger.Debug(ctx, msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...Field) {
	globalManager.mu.RLock()
	logger := globalManager.defaultLogger
	globalManager.mu.RUnlock()
	logger.Info(ctx, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...Field) {
	globalManager.mu.RLock()
	logger := globalManager.defaultLogger
	globalManager.mu.RUnlock()
	logger.Warn(ctx, msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...Field) {
	globalManager.mu.RLock()
	logger := globalManager.defaultLogger
	globalManager.mu.RUnlock()
	logger.Error(ctx, msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...Field) {
	globalManager.mu.RLock()
	logger := globalManager.defaultLogger
	globalManager.mu.RUnlock()
	logger.Fatal(ctx, msg, fields...)
}

func Panic(ctx context.Context, msg string, fields ...Field) {
	globalManager.mu.RLock()
	logger := globalManager.defaultLogger
	globalManager.mu.RUnlock()
	logger.Panic(ctx, msg, fields...)
}

func Errorf(ctx context.Context, format string, args ...any) {
	globalManager.mu.RLock()
	logger := globalManager.defaultLogger
	globalManager.mu.RUnlock()
	logger.Errorf(ctx, format, args...)
}

func Infof(ctx context.Context, format string, args ...any) {
	globalManager.mu.RLock()
	logger := globalManager.defaultLogger
	globalManager.mu.RUnlock()
	logger.Infof(ctx, format, args...)
}

func Debugf(ctx context.Context, format string, args ...any) {
	globalManager.mu.RLock()
	logger := globalManager.defaultLogger
	globalManager.mu.RUnlock()
	logger.Debugf(ctx, format, args...)
}

func Warnf(ctx context.Context, format string, args ...any) {
	globalManager.mu.RLock()
	logger := globalManager.defaultLogger
	globalManager.mu.RUnlock()
	logger.Warnf(ctx, format, args...)
}

func Fatalf(ctx context.Context, format string, args ...any) {
	globalManager.mu.RLock()
	logger := globalManager.defaultLogger
	globalManager.mu.RUnlock()
	logger.Fatalf(ctx, format, args...)
}

func Panicf(ctx context.Context, format string, args ...any) {
	globalManager.mu.RLock()
	logger := globalManager.defaultLogger
	globalManager.mu.RUnlock()
	logger.Panicf(ctx, format, args...)
}

func WithFields(fields ...Field) Logger {
	globalManager.mu.RLock()
	logger := globalManager.defaultLogger
	globalManager.mu.RUnlock()
	return logger.WithFields(fields...)
}

func WithContext(ctx context.Context) Logger {
	globalManager.mu.RLock()
	logger := globalManager.defaultLogger
	globalManager.mu.RUnlock()
	return logger.WithContext(ctx)
}

// noopLogger implementação vazia para quando nenhum provider está configurado
type noopLogger struct{}

func (n *noopLogger) Debug(ctx context.Context, msg string, fields ...Field)               {}
func (n *noopLogger) Info(ctx context.Context, msg string, fields ...Field)                {}
func (n *noopLogger) Warn(ctx context.Context, msg string, fields ...Field)                {}
func (n *noopLogger) Error(ctx context.Context, msg string, fields ...Field)               {}
func (n *noopLogger) Fatal(ctx context.Context, msg string, fields ...Field)               {}
func (n *noopLogger) Panic(ctx context.Context, msg string, fields ...Field)               {}
func (n *noopLogger) Debugf(ctx context.Context, format string, args ...any)               {}
func (n *noopLogger) Infof(ctx context.Context, format string, args ...any)                {}
func (n *noopLogger) Warnf(ctx context.Context, format string, args ...any)                {}
func (n *noopLogger) Errorf(ctx context.Context, format string, args ...any)               {}
func (n *noopLogger) Fatalf(ctx context.Context, format string, args ...any)               {}
func (n *noopLogger) Panicf(ctx context.Context, format string, args ...any)               {}
func (n *noopLogger) DebugWithCode(ctx context.Context, code, msg string, fields ...Field) {}
func (n *noopLogger) InfoWithCode(ctx context.Context, code, msg string, fields ...Field)  {}
func (n *noopLogger) WarnWithCode(ctx context.Context, code, msg string, fields ...Field)  {}
func (n *noopLogger) ErrorWithCode(ctx context.Context, code, msg string, fields ...Field) {}
func (n *noopLogger) WithFields(fields ...Field) Logger                                    { return n }
func (n *noopLogger) WithContext(ctx context.Context) Logger                               { return n }
func (n *noopLogger) SetLevel(level Level)                                                 {}
func (n *noopLogger) GetLevel() Level                                                      { return InfoLevel }
func (n *noopLogger) Clone() Logger                                                        { return &noopLogger{} }
