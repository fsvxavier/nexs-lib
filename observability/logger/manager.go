package logger

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// LoggerManager gerencia os diferentes providers de logging
type LoggerManager struct {
	providers map[string]Provider
	current   Logger
	mu        sync.RWMutex
}

var globalManager = &LoggerManager{
	providers: make(map[string]Provider),
	current:   &noopLogger{},
}

// RegisterProvider registra um provider de logging
func RegisterProvider(name string, provider Provider) {
	globalManager.mu.Lock()
	defer globalManager.mu.Unlock()
	globalManager.providers[name] = provider

	// Se o provider sendo registrado for zap, configure-o como padrão
	if name == "zap" {
		setDefaultZapProvider(provider)
	} else if name == "slog" && len(globalManager.providers) == 1 {
		// Se for o primeiro provider registrado e for slog, use como fallback
		setDefaultSlogProvider(provider)
	}
}

// setDefaultZapProvider configura zap como provider padrão
func setDefaultZapProvider(provider Provider) {
	defaultConfig := &Config{
		Level:          InfoLevel,
		Format:         JSONFormat,
		Output:         os.Stdout,
		AddSource:      false,
		AddStacktrace:  false,
		TimeFormat:     time.RFC3339,
		ServiceName:    "application",
		ServiceVersion: "1.0.0",
		Environment:    "development",
	}

	if err := provider.Configure(defaultConfig); err == nil {
		globalManager.current = provider
	}
}

// setDefaultSlogProvider configura slog como provider fallback
func setDefaultSlogProvider(provider Provider) {
	// Só configura slog se não houver provider atual configurado
	if globalManager.current == nil || fmt.Sprintf("%T", globalManager.current) == "*logger.noopLogger" {
		defaultConfig := &Config{
			Level:          InfoLevel,
			Format:         JSONFormat,
			Output:         os.Stdout,
			AddSource:      false,
			AddStacktrace:  false,
			TimeFormat:     time.RFC3339,
			ServiceName:    "application",
			ServiceVersion: "1.0.0",
			Environment:    "development",
		}

		if err := provider.Configure(defaultConfig); err == nil {
			globalManager.current = provider
		}
	}
}

// SetProvider define o provider ativo
func SetProvider(name string, config *Config) error {
	globalManager.mu.Lock()
	defer globalManager.mu.Unlock()

	provider, exists := globalManager.providers[name]
	if !exists {
		return fmt.Errorf("provider '%s' not found", name)
	}

	if err := provider.Configure(config); err != nil {
		return fmt.Errorf("failed to configure provider '%s': %w", name, err)
	}

	globalManager.current = provider
	return nil
}

// GetCurrentProvider retorna o provider atual
func GetCurrentProvider() Logger {
	globalManager.mu.RLock()
	defer globalManager.mu.RUnlock()
	return globalManager.current
}

// GetCurrentProviderName retorna o nome do provider atual
func GetCurrentProviderName() string {
	globalManager.mu.RLock()
	defer globalManager.mu.RUnlock()

	currentProvider := globalManager.current
	for name, provider := range globalManager.providers {
		if provider == currentProvider {
			return name
		}
	}
	return "unknown"
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

// ConfigureProvider configura um provider específico
func ConfigureProvider(name string, config *Config) error {
	globalManager.mu.Lock()
	defer globalManager.mu.Unlock()

	provider, exists := globalManager.providers[name]
	if !exists {
		return fmt.Errorf("provider '%s' not found", name)
	}

	if err := provider.Configure(config); err != nil {
		return fmt.Errorf("failed to configure provider '%s': %w", name, err)
	}

	return nil
}

// SetActiveProvider define o provider ativo (versão simplificada)
func SetActiveProvider(name string) error {
	globalManager.mu.Lock()
	defer globalManager.mu.Unlock()

	provider, exists := globalManager.providers[name]
	if !exists {
		return fmt.Errorf("provider '%s' not found", name)
	}

	globalManager.current = provider
	return nil
}

// Métodos globais que delegam para o logger atual
func Debug(ctx context.Context, msg string, fields ...Field) {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	logger.Debug(ctx, msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...Field) {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	logger.Info(ctx, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...Field) {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	logger.Warn(ctx, msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...Field) {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	logger.Error(ctx, msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...Field) {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	logger.Fatal(ctx, msg, fields...)
}

func Panic(ctx context.Context, msg string, fields ...Field) {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	logger.Panic(ctx, msg, fields...)
}

func Debugf(ctx context.Context, format string, args ...any) {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	logger.Debugf(ctx, format, args...)
}

func Infof(ctx context.Context, format string, args ...any) {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	logger.Infof(ctx, format, args...)
}

func Warnf(ctx context.Context, format string, args ...any) {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	logger.Warnf(ctx, format, args...)
}

func Errorf(ctx context.Context, format string, args ...any) {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	logger.Errorf(ctx, format, args...)
}

func Fatalf(ctx context.Context, format string, args ...any) {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	logger.Fatalf(ctx, format, args...)
}

func Panicf(ctx context.Context, format string, args ...any) {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	logger.Panicf(ctx, format, args...)
}

func WithFields(fields ...Field) Logger {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	return logger.WithFields(fields...)
}

func WithContext(ctx context.Context) Logger {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	return logger.WithContext(ctx)
}

// Métodos globais com código
func DebugWithCode(ctx context.Context, code, msg string, fields ...Field) {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	logger.DebugWithCode(ctx, code, msg, fields...)
}

func InfoWithCode(ctx context.Context, code, msg string, fields ...Field) {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	logger.InfoWithCode(ctx, code, msg, fields...)
}

func WarnWithCode(ctx context.Context, code, msg string, fields ...Field) {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	logger.WarnWithCode(ctx, code, msg, fields...)
}

func ErrorWithCode(ctx context.Context, code, msg string, fields ...Field) {
	globalManager.mu.RLock()
	logger := globalManager.current
	globalManager.mu.RUnlock()
	logger.ErrorWithCode(ctx, code, msg, fields...)
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
func (n *noopLogger) Close() error                                                         { return nil }
