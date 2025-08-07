package hooks

import (
	"context"
	"sync"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// StopHookManager gerencia hooks de parada
type StopHookManager struct {
	hooks []interfaces.StopHookFunc
	mu    sync.RWMutex
}

// NewStopHookManager cria um novo gerenciador de hooks de stop
func NewStopHookManager() *StopHookManager {
	return &StopHookManager{
		hooks: make([]interfaces.StopHookFunc, 0),
	}
}

// Register registra um hook de stop
func (m *StopHookManager) Register(hook interfaces.StopHookFunc) {
	if hook == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.hooks = append(m.hooks, hook)
}

// Execute executa todos os hooks de stop registrados
func (m *StopHookManager) Execute(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, hook := range m.hooks {
		if err := hook(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Count retorna o número de hooks registrados
func (m *StopHookManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.hooks)
}

// Clear remove todos os hooks registrados
func (m *StopHookManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.hooks = make([]interfaces.StopHookFunc, 0)
}

// Instância global para uso em toda a aplicação
var GlobalStopHookManager = NewStopHookManager()

// RegisterStopHook registra um hook de stop globalmente
func RegisterStopHook(hook interfaces.StopHookFunc) {
	GlobalStopHookManager.Register(hook)
}

// ExecuteStopHooks executa todos os hooks de stop globais
func ExecuteStopHooks(ctx context.Context) error {
	return GlobalStopHookManager.Execute(ctx)
}

// GetStopHookCount retorna o número de hooks de stop globais
func GetStopHookCount() int {
	return GlobalStopHookManager.Count()
}

// ClearStopHooks limpa todos os hooks de stop globais
func ClearStopHooks() {
	GlobalStopHookManager.Clear()
}
