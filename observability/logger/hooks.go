package logger

import (
	"context"
	"fmt"
	"sync"

	"github.com/fsvxavier/nexs-lib/observability/logger/interfaces"
)

// hookManager implementação thread-safe do gerenciador de hooks
type hookManager struct {
	mu          sync.RWMutex
	beforeHooks map[string]interfaces.Hook
	afterHooks  map[string]interfaces.Hook
	enabled     bool
}

// NewHookManager cria um novo gerenciador de hooks
func NewHookManager() interfaces.HookManager {
	return &hookManager{
		beforeHooks: make(map[string]interfaces.Hook),
		afterHooks:  make(map[string]interfaces.Hook),
		enabled:     true,
	}
}

// RegisterHook registra um novo hook
func (h *hookManager) RegisterHook(hookType interfaces.HookType, hook interfaces.Hook) error {
	if hook == nil {
		return fmt.Errorf("hook cannot be nil")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	name := hook.GetName()
	if name == "" {
		return fmt.Errorf("hook name cannot be empty")
	}

	switch hookType {
	case interfaces.BeforeHook:
		if _, exists := h.beforeHooks[name]; exists {
			return fmt.Errorf("before hook '%s' already registered", name)
		}
		h.beforeHooks[name] = hook
	case interfaces.AfterHook:
		if _, exists := h.afterHooks[name]; exists {
			return fmt.Errorf("after hook '%s' already registered", name)
		}
		h.afterHooks[name] = hook
	default:
		return fmt.Errorf("invalid hook type: %s", hookType)
	}

	return nil
}

// UnregisterHook remove um hook registrado
func (h *hookManager) UnregisterHook(hookType interfaces.HookType, name string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch hookType {
	case interfaces.BeforeHook:
		if _, exists := h.beforeHooks[name]; !exists {
			return fmt.Errorf("before hook '%s' not found", name)
		}
		delete(h.beforeHooks, name)
	case interfaces.AfterHook:
		if _, exists := h.afterHooks[name]; !exists {
			return fmt.Errorf("after hook '%s' not found", name)
		}
		delete(h.afterHooks, name)
	default:
		return fmt.Errorf("invalid hook type: %s", hookType)
	}

	return nil
}

// ExecuteBeforeHooks executa todos os hooks "before"
func (h *hookManager) ExecuteBeforeHooks(ctx context.Context, entry *interfaces.LogEntry) error {
	if !h.enabled {
		return nil
	}

	h.mu.RLock()
	hooks := make([]interfaces.Hook, 0, len(h.beforeHooks))
	for _, hook := range h.beforeHooks {
		if hook.IsEnabled() {
			hooks = append(hooks, hook)
		}
	}
	h.mu.RUnlock()

	// Executa hooks sem segurar o lock para evitar deadlock
	for _, hook := range hooks {
		if err := hook.Execute(ctx, entry); err != nil {
			return fmt.Errorf("before hook '%s' failed: %w", hook.GetName(), err)
		}
	}

	return nil
}

// ExecuteAfterHooks executa todos os hooks "after"
func (h *hookManager) ExecuteAfterHooks(ctx context.Context, entry *interfaces.LogEntry) error {
	if !h.enabled {
		return nil
	}

	h.mu.RLock()
	hooks := make([]interfaces.Hook, 0, len(h.afterHooks))
	for _, hook := range h.afterHooks {
		if hook.IsEnabled() {
			hooks = append(hooks, hook)
		}
	}
	h.mu.RUnlock()

	// Executa hooks sem segurar o lock
	var lastError error
	for _, hook := range hooks {
		if err := hook.Execute(ctx, entry); err != nil {
			// Para hooks "after", coletamos erros mas não interrompemos a execução
			lastError = fmt.Errorf("after hook '%s' failed: %w", hook.GetName(), err)
		}
	}

	return lastError
}

// ListHooks retorna lista de hooks de um tipo
func (h *hookManager) ListHooks(hookType interfaces.HookType) []interfaces.Hook {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var hooks []interfaces.Hook
	switch hookType {
	case interfaces.BeforeHook:
		hooks = make([]interfaces.Hook, 0, len(h.beforeHooks))
		for _, hook := range h.beforeHooks {
			hooks = append(hooks, hook)
		}
	case interfaces.AfterHook:
		hooks = make([]interfaces.Hook, 0, len(h.afterHooks))
		for _, hook := range h.afterHooks {
			hooks = append(hooks, hook)
		}
	}

	return hooks
}

// GetHook retorna um hook específico
func (h *hookManager) GetHook(hookType interfaces.HookType, name string) interfaces.Hook {
	h.mu.RLock()
	defer h.mu.RUnlock()

	switch hookType {
	case interfaces.BeforeHook:
		return h.beforeHooks[name]
	case interfaces.AfterHook:
		return h.afterHooks[name]
	default:
		return nil
	}
}

// ClearHooks remove todos os hooks de um tipo
func (h *hookManager) ClearHooks(hookType interfaces.HookType) {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch hookType {
	case interfaces.BeforeHook:
		h.beforeHooks = make(map[string]interfaces.Hook)
	case interfaces.AfterHook:
		h.afterHooks = make(map[string]interfaces.Hook)
	}
}

// EnableAllHooks habilita todos os hooks
func (h *hookManager) EnableAllHooks() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.enabled = true

	for _, hook := range h.beforeHooks {
		hook.SetEnabled(true)
	}
	for _, hook := range h.afterHooks {
		hook.SetEnabled(true)
	}
}

// DisableAllHooks desabilita todos os hooks
func (h *hookManager) DisableAllHooks() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.enabled = false

	for _, hook := range h.beforeHooks {
		hook.SetEnabled(false)
	}
	for _, hook := range h.afterHooks {
		hook.SetEnabled(false)
	}
}

// GetHookCount retorna número de hooks de um tipo
func (h *hookManager) GetHookCount(hookType interfaces.HookType) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	switch hookType {
	case interfaces.BeforeHook:
		return len(h.beforeHooks)
	case interfaces.AfterHook:
		return len(h.afterHooks)
	default:
		return 0
	}
}

// baseHook implementação base para hooks personalizados
type baseHook struct {
	name    string
	enabled bool
	mu      sync.RWMutex
}

// NewBaseHook cria um hook base
func NewBaseHook(name string) *baseHook {
	return &baseHook{
		name:    name,
		enabled: true,
	}
}

// GetName retorna o nome do hook
func (b *baseHook) GetName() string {
	return b.name
}

// IsEnabled verifica se o hook está habilitado
func (b *baseHook) IsEnabled() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.enabled
}

// SetEnabled habilita/desabilita o hook
func (b *baseHook) SetEnabled(enabled bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.enabled = enabled
}

// Execute deve ser implementado pelos hooks específicos
func (b *baseHook) Execute(ctx context.Context, entry *interfaces.LogEntry) error {
	return fmt.Errorf("execute method must be implemented by specific hook")
}

// MetricsHook hook para coleta de métricas
type MetricsHook struct {
	*baseHook
	collector interfaces.MetricsCollector
}

// NewMetricsHook cria um hook de métricas
func NewMetricsHook(collector interfaces.MetricsCollector) *MetricsHook {
	return &MetricsHook{
		baseHook:  NewBaseHook("metrics_collector"),
		collector: collector,
	}
}

// Execute coleta métricas da entrada de log
func (m *MetricsHook) Execute(ctx context.Context, entry *interfaces.LogEntry) error {
	if m.collector == nil {
		return nil
	}

	// Registra o log (assumindo que o tempo já foi calculado)
	m.collector.RecordLog(entry.Level, 0)

	return nil
}

// ValidationHook hook para validação de entradas
type ValidationHook struct {
	*baseHook
	validators []func(*interfaces.LogEntry) error
}

// NewValidationHook cria um hook de validação
func NewValidationHook(validators ...func(*interfaces.LogEntry) error) *ValidationHook {
	return &ValidationHook{
		baseHook:   NewBaseHook("entry_validator"),
		validators: validators,
	}
}

// Execute valida a entrada de log
func (v *ValidationHook) Execute(ctx context.Context, entry *interfaces.LogEntry) error {
	for _, validator := range v.validators {
		if err := validator(entry); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}
	return nil
}

// FilterHook hook para filtragem de logs
type FilterHook struct {
	*baseHook
	filter func(*interfaces.LogEntry) bool
}

// NewFilterHook cria um hook de filtro
func NewFilterHook(filter func(*interfaces.LogEntry) bool) *FilterHook {
	return &FilterHook{
		baseHook: NewBaseHook("log_filter"),
		filter:   filter,
	}
}

// Execute filtra a entrada de log
func (f *FilterHook) Execute(ctx context.Context, entry *interfaces.LogEntry) error {
	if f.filter != nil && !f.filter(entry) {
		return fmt.Errorf("log entry filtered out")
	}
	return nil
}

// TransformHook hook para transformação de dados
type TransformHook struct {
	*baseHook
	transformer func(*interfaces.LogEntry) error
}

// NewTransformHook cria um hook de transformação
func NewTransformHook(transformer func(*interfaces.LogEntry) error) *TransformHook {
	return &TransformHook{
		baseHook:    NewBaseHook("data_transformer"),
		transformer: transformer,
	}
}

// Execute transforma a entrada de log
func (t *TransformHook) Execute(ctx context.Context, entry *interfaces.LogEntry) error {
	if t.transformer != nil {
		return t.transformer(entry)
	}
	return nil
}
