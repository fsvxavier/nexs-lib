package decimal

import (
	"context"
	"fmt"
	"sync"

	"github.com/fsvxavier/nexs-lib/decimal/config"
	"github.com/fsvxavier/nexs-lib/decimal/hooks"
	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
	cockroachProvider "github.com/fsvxavier/nexs-lib/decimal/providers/cockroach"
	shopspringProvider "github.com/fsvxavier/nexs-lib/decimal/providers/shopspring"
)

// Manager implements DecimalManager interface
type Manager struct {
	mu              sync.RWMutex
	config          *config.Config
	currentProvider interfaces.DecimalProvider
	providers       map[string]func(*config.Config) interfaces.DecimalProvider
	hookManager     interfaces.HookManager
}

// NewManager creates a new decimal manager with the given configuration
func NewManager(cfg *config.Config) *Manager {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	manager := &Manager{
		config:    cfg,
		providers: make(map[string]func(*config.Config) interfaces.DecimalProvider),
	}

	// Register built-in providers
	manager.registerBuiltinProviders()

	// Set the initial provider
	provider, err := manager.getProviderByName(cfg.GetProviderName())
	if err != nil {
		// Fallback to cockroach if configured provider fails
		provider, _ = manager.getProviderByName("cockroach")
	}
	manager.currentProvider = provider

	// Initialize hook manager if hooks are enabled
	if cfg.IsHooksEnabled() {
		manager.hookManager = hooks.NewHookManager()
	}

	return manager
}

// NewManagerWithProvider creates a new manager with a specific provider
func NewManagerWithProvider(provider interfaces.DecimalProvider, cfg *config.Config) *Manager {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	manager := &Manager{
		config:          cfg,
		currentProvider: provider,
		providers:       make(map[string]func(*config.Config) interfaces.DecimalProvider),
	}

	manager.registerBuiltinProviders()

	if cfg.IsHooksEnabled() {
		manager.hookManager = hooks.NewHookManager()
	}

	return manager
}

func (m *Manager) registerBuiltinProviders() {
	m.providers["cockroach"] = func(cfg *config.Config) interfaces.DecimalProvider {
		return cockroachProvider.NewProvider(cfg)
	}
	m.providers["shopspring"] = func(cfg *config.Config) interfaces.DecimalProvider {
		return shopspringProvider.NewProvider(cfg)
	}
}

func (m *Manager) getProviderByName(name string) (interfaces.DecimalProvider, error) {
	providerFactory, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}
	return providerFactory(m.config), nil
}

// SetProvider sets the current provider
func (m *Manager) SetProvider(provider interfaces.DecimalProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentProvider = provider
}

// GetProvider returns the current provider
func (m *Manager) GetProvider() interfaces.DecimalProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentProvider
}

// SwitchProvider switches to a different provider by name
func (m *Manager) SwitchProvider(providerName string) error {
	provider, err := m.getProviderByName(providerName)
	if err != nil {
		return err
	}

	m.SetProvider(provider)
	return nil
}

// Factory methods with current provider
func (m *Manager) NewFromString(value string) (interfaces.Decimal, error) {
	ctx := context.Background()

	if m.hookManager != nil {
		modifiedValue, err := m.hookManager.ExecutePreHooks(ctx, "NewFromString", value)
		if err != nil {
			m.hookManager.ExecuteErrorHooks(ctx, "NewFromString", err)
			return nil, err
		}
		if modifiedValue != nil {
			if str, ok := modifiedValue.(string); ok {
				value = str
			}
		}
	}

	result, err := m.GetProvider().NewFromString(value)

	if m.hookManager != nil {
		if err != nil {
			m.hookManager.ExecuteErrorHooks(ctx, "NewFromString", err)
		} else {
			m.hookManager.ExecutePostHooks(ctx, "NewFromString", result, err)
		}
	}

	return result, err
}

func (m *Manager) NewFromFloat(value float64) (interfaces.Decimal, error) {
	ctx := context.Background()

	if m.hookManager != nil {
		modifiedValue, err := m.hookManager.ExecutePreHooks(ctx, "NewFromFloat", value)
		if err != nil {
			m.hookManager.ExecuteErrorHooks(ctx, "NewFromFloat", err)
			return nil, err
		}
		if modifiedValue != nil {
			if f, ok := modifiedValue.(float64); ok {
				value = f
			}
		}
	}

	result, err := m.GetProvider().NewFromFloat(value)

	if m.hookManager != nil {
		if err != nil {
			m.hookManager.ExecuteErrorHooks(ctx, "NewFromFloat", err)
		} else {
			m.hookManager.ExecutePostHooks(ctx, "NewFromFloat", result, err)
		}
	}

	return result, err
}

func (m *Manager) NewFromInt(value int64) (interfaces.Decimal, error) {
	ctx := context.Background()

	if m.hookManager != nil {
		modifiedValue, err := m.hookManager.ExecutePreHooks(ctx, "NewFromInt", value)
		if err != nil {
			m.hookManager.ExecuteErrorHooks(ctx, "NewFromInt", err)
			return nil, err
		}
		if modifiedValue != nil {
			if i, ok := modifiedValue.(int64); ok {
				value = i
			}
		}
	}

	result, err := m.GetProvider().NewFromInt(value)

	if m.hookManager != nil {
		if err != nil {
			m.hookManager.ExecuteErrorHooks(ctx, "NewFromInt", err)
		} else {
			m.hookManager.ExecutePostHooks(ctx, "NewFromInt", result, err)
		}
	}

	return result, err
}

func (m *Manager) Zero() interfaces.Decimal {
	return m.GetProvider().Zero()
}

// Batch operations
func (m *Manager) Sum(decimals ...interfaces.Decimal) (interfaces.Decimal, error) {
	if len(decimals) == 0 {
		return m.Zero(), nil
	}

	result := decimals[0]
	for i := 1; i < len(decimals); i++ {
		var err error
		result, err = result.Add(decimals[i])
		if err != nil {
			return nil, fmt.Errorf("error during sum at index %d: %w", i, err)
		}
	}

	return result, nil
}

func (m *Manager) Average(decimals ...interfaces.Decimal) (interfaces.Decimal, error) {
	if len(decimals) == 0 {
		return nil, fmt.Errorf("cannot calculate average of empty slice")
	}

	sum, err := m.Sum(decimals...)
	if err != nil {
		return nil, fmt.Errorf("error calculating sum for average: %w", err)
	}

	count, err := m.NewFromInt(int64(len(decimals)))
	if err != nil {
		return nil, fmt.Errorf("error creating count decimal: %w", err)
	}

	return sum.Div(count)
}

func (m *Manager) Max(decimals ...interfaces.Decimal) (interfaces.Decimal, error) {
	if len(decimals) == 0 {
		return nil, fmt.Errorf("cannot find max of empty slice")
	}

	max := decimals[0]
	for i := 1; i < len(decimals); i++ {
		if decimals[i].IsGreaterThan(max) {
			max = decimals[i]
		}
	}

	return max, nil
}

func (m *Manager) Min(decimals ...interfaces.Decimal) (interfaces.Decimal, error) {
	if len(decimals) == 0 {
		return nil, fmt.Errorf("cannot find min of empty slice")
	}

	min := decimals[0]
	for i := 1; i < len(decimals); i++ {
		if decimals[i].IsLessThan(min) {
			min = decimals[i]
		}
	}

	return min, nil
}

// Utility operations
func (m *Manager) Parse(value interface{}) (interfaces.Decimal, error) {
	switch v := value.(type) {
	case string:
		return m.NewFromString(v)
	case float64:
		return m.NewFromFloat(v)
	case float32:
		return m.NewFromFloat(float64(v))
	case int:
		return m.NewFromInt(int64(v))
	case int32:
		return m.NewFromInt(int64(v))
	case int64:
		return m.NewFromInt(v)
	case interfaces.Decimal:
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported type for decimal parsing: %T", value)
	}
}

func (m *Manager) MarshalJSON(decimal interfaces.Decimal) ([]byte, error) {
	return decimal.MarshalJSON()
}

func (m *Manager) UnmarshalJSON(data []byte) (interfaces.Decimal, error) {
	// Try to parse as string first
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	return m.NewFromString(str)
}

// Hook management
func (m *Manager) GetHookManager() interfaces.HookManager {
	return m.hookManager
}

func (m *Manager) SetHookManager(hm interfaces.HookManager) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hookManager = hm
}

// Configuration access
func (m *Manager) GetConfig() *config.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

func (m *Manager) UpdateConfig(cfg *config.Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = cfg

	// Update provider if needed
	if cfg.GetProviderName() != m.currentProvider.Name() {
		provider, err := m.getProviderByName(cfg.GetProviderName())
		if err != nil {
			return fmt.Errorf("failed to switch to provider '%s': %w", cfg.GetProviderName(), err)
		}
		m.currentProvider = provider
	}

	return nil
}

// Default global manager instance
var defaultManager *Manager
var defaultManagerOnce sync.Once

// GetDefaultManager returns the default global manager instance
func GetDefaultManager() *Manager {
	defaultManagerOnce.Do(func() {
		defaultManager = NewManager(nil)
	})
	return defaultManager
}

// SetDefaultManager sets the global default manager
func SetDefaultManager(manager *Manager) {
	defaultManager = manager
}

// Convenience functions using the default manager
func NewFromString(value string) (interfaces.Decimal, error) {
	return GetDefaultManager().NewFromString(value)
}

func NewFromFloat(value float64) (interfaces.Decimal, error) {
	return GetDefaultManager().NewFromFloat(value)
}

func NewFromInt(value int64) (interfaces.Decimal, error) {
	return GetDefaultManager().NewFromInt(value)
}

func Zero() interfaces.Decimal {
	return GetDefaultManager().Zero()
}

func Sum(decimals ...interfaces.Decimal) (interfaces.Decimal, error) {
	return GetDefaultManager().Sum(decimals...)
}

func Average(decimals ...interfaces.Decimal) (interfaces.Decimal, error) {
	return GetDefaultManager().Average(decimals...)
}

func Max(decimals ...interfaces.Decimal) (interfaces.Decimal, error) {
	return GetDefaultManager().Max(decimals...)
}

func Min(decimals ...interfaces.Decimal) (interfaces.Decimal, error) {
	return GetDefaultManager().Min(decimals...)
}

func Parse(value interface{}) (interfaces.Decimal, error) {
	return GetDefaultManager().Parse(value)
}
