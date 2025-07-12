package logger

import (
	"context"
	"fmt"
	"sync"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/observability/logger/providers/slog"
	"github.com/fsvxavier/nexs-lib/v2/observability/logger/providers/zap"
	"github.com/fsvxavier/nexs-lib/v2/observability/logger/providers/zerolog"
)

// Factory implementa o padrão Factory para criação de loggers e providers
type Factory struct {
	providers     map[string]interfaces.Provider
	mu            sync.RWMutex
	defaultConfig interfaces.Config
}

// NewFactory cria uma nova instância da Factory
func NewFactory() *Factory {
	return &Factory{
		providers:     make(map[string]interfaces.Provider),
		defaultConfig: DefaultConfig(),
	}
}

// CreateLogger cria um novo logger com o provider padrão
func (f *Factory) CreateLogger(name string, config interfaces.Config) (interfaces.Logger, error) {
	if err := interfaces.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Se não há provider registrado, retorna erro
	f.mu.RLock()
	defer f.mu.RUnlock()

	if len(f.providers) == 0 {
		return nil, fmt.Errorf("no providers registered")
	}

	// Usa o primeiro provider disponível ou um específico se definido no config
	var provider interfaces.Provider

	// Se há apenas um provider, usa ele
	if len(f.providers) == 1 {
		for _, p := range f.providers {
			provider = p
			break
		}
	} else {
		// Se há múltiplos providers, precisa especificar qual usar
		// Por ora, vamos usar o primeiro disponível
		for _, p := range f.providers {
			provider = p
			break
		}
	}

	if provider == nil {
		return nil, fmt.Errorf("no suitable provider found")
	}

	// Configura o provider
	if err := provider.Configure(config); err != nil {
		return nil, fmt.Errorf("failed to configure provider: %w", err)
	}

	// Cria o core logger
	coreLogger := NewCoreLogger(provider, config)

	return coreLogger, nil
}

// CreateProvider cria um novo provider do tipo especificado
func (f *Factory) CreateProvider(providerType string, config interfaces.Config) (interfaces.Provider, error) {
	f.mu.RLock()
	provider, exists := f.providers[providerType]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider type '%s' not registered", providerType)
	}

	// Configura o provider
	if err := provider.Configure(config); err != nil {
		return nil, fmt.Errorf("failed to configure provider: %w", err)
	}

	return provider, nil
}

// RegisterDefaultProviders registra todos os providers padrão
func (f *Factory) RegisterDefaultProviders() {
	f.RegisterProvider("zap", zap.NewProvider())
	f.RegisterProvider("slog", slog.NewProvider())
	f.RegisterProvider("zerolog", zerolog.NewProvider())
}

// RegisterProvider registra um novo provider na factory
func (f *Factory) RegisterProvider(name string, provider interfaces.Provider) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.providers[name] = provider
}

// GetProvider retorna um provider registrado
func (f *Factory) GetProvider(name string) (interfaces.Provider, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	provider, exists := f.providers[name]
	return provider, exists
}

// ListProviders retorna a lista de nomes de providers registrados
func (f *Factory) ListProviders() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	names := make([]string, 0, len(f.providers))
	for name := range f.providers {
		names = append(names, name)
	}

	return names
}

// SetDefaultConfig define uma nova configuração padrão
func (f *Factory) SetDefaultConfig(config interfaces.Config) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.defaultConfig = config
}

// GetDefaultConfig retorna a configuração padrão
func (f *Factory) GetDefaultConfig() interfaces.Config {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.defaultConfig
}

// Manager global singleton para facilitar o uso
type Manager struct {
	factory *Factory
	current interfaces.Logger
	mu      sync.RWMutex
}

var globalManager = &Manager{
	factory: func() *Factory {
		f := NewFactory()
		f.RegisterDefaultProviders()
		return f
	}(),
}

// GetGlobalManager retorna a instância global do manager
func GetGlobalManager() *Manager {
	return globalManager
}

// RegisterProvider registra um provider globalmente
func RegisterProvider(name string, provider interfaces.Provider) {
	globalManager.factory.RegisterProvider(name, provider)
}

// SetProvider define o provider ativo globalmente
func SetProvider(name string, config interfaces.Config) error {
	provider, exists := globalManager.factory.GetProvider(name)
	if !exists {
		return fmt.Errorf("provider '%s' not found", name)
	}

	// Configura o provider
	if err := provider.Configure(config); err != nil {
		return fmt.Errorf("failed to configure provider: %w", err)
	}

	logger, err := globalManager.factory.CreateLogger(name, config)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	globalManager.mu.Lock()
	defer globalManager.mu.Unlock()
	globalManager.current = logger

	return nil
}

// SetCurrentLogger define o logger atual
func (m *Manager) SetCurrentLogger(logger interfaces.Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.current = logger
}

// GetCurrentLogger retorna o logger atual
func GetCurrentLogger() interfaces.Logger {
	globalManager.mu.RLock()
	defer globalManager.mu.RUnlock()

	if globalManager.current == nil {
		// Retorna um logger no-op se nenhum foi configurado
		return &noopLogger{}
	}

	return globalManager.current
}

// ListProviders lista todos os providers registrados globalmente
func ListProviders() []string {
	return globalManager.factory.ListProviders()
}

// CreateLogger cria um novo logger usando a factory global
func CreateLogger(name string, config interfaces.Config) (interfaces.Logger, error) {
	return globalManager.factory.CreateLogger(name, config)
}

// noopLogger implementação no-op para casos onde nenhum provider foi configurado
type noopLogger struct{}

func (n *noopLogger) Trace(ctx context.Context, msg string, fields ...interfaces.Field) {}
func (n *noopLogger) Debug(ctx context.Context, msg string, fields ...interfaces.Field) {}
func (n *noopLogger) Info(ctx context.Context, msg string, fields ...interfaces.Field)  {}
func (n *noopLogger) Warn(ctx context.Context, msg string, fields ...interfaces.Field)  {}
func (n *noopLogger) Error(ctx context.Context, msg string, fields ...interfaces.Field) {}
func (n *noopLogger) Fatal(ctx context.Context, msg string, fields ...interfaces.Field) {}
func (n *noopLogger) Panic(ctx context.Context, msg string, fields ...interfaces.Field) {}
func (n *noopLogger) Tracef(ctx context.Context, format string, args ...interface{})    {}
func (n *noopLogger) Debugf(ctx context.Context, format string, args ...interface{})    {}
func (n *noopLogger) Infof(ctx context.Context, format string, args ...interface{})     {}
func (n *noopLogger) Warnf(ctx context.Context, format string, args ...interface{})     {}
func (n *noopLogger) Errorf(ctx context.Context, format string, args ...interface{})    {}
func (n *noopLogger) Fatalf(ctx context.Context, format string, args ...interface{})    {}
func (n *noopLogger) Panicf(ctx context.Context, format string, args ...interface{})    {}
func (n *noopLogger) TraceWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
}
func (n *noopLogger) DebugWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
}
func (n *noopLogger) InfoWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
}
func (n *noopLogger) WarnWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
}
func (n *noopLogger) ErrorWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
}
func (n *noopLogger) WithFields(fields ...interfaces.Field) interfaces.Logger { return n }
func (n *noopLogger) WithContext(ctx context.Context) interfaces.Logger       { return n }
func (n *noopLogger) WithError(err error) interfaces.Logger                   { return n }
func (n *noopLogger) WithTraceID(traceID string) interfaces.Logger            { return n }
func (n *noopLogger) WithSpanID(spanID string) interfaces.Logger              { return n }
func (n *noopLogger) SetLevel(level interfaces.Level)                         {}
func (n *noopLogger) GetLevel() interfaces.Level                              { return interfaces.InfoLevel }
func (n *noopLogger) IsLevelEnabled(level interfaces.Level) bool              { return false }
func (n *noopLogger) Clone() interfaces.Logger                                { return n }
func (n *noopLogger) Flush() error                                            { return nil }
func (n *noopLogger) Close() error                                            { return nil }
