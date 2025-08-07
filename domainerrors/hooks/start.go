package hooks

import (
	"context"
	"sync"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// StartHookManager gerencia hooks de inicialização
type StartHookManager struct {
	hooks []interfaces.StartHookFunc
	mu    sync.RWMutex
}

// NewStartHookManager cria um novo gerenciador de hooks de start
func NewStartHookManager() *StartHookManager {
	return &StartHookManager{
		hooks: make([]interfaces.StartHookFunc, 0),
	}
}

// Register registra um hook de start
func (m *StartHookManager) Register(hook interfaces.StartHookFunc) {
	if hook == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.hooks = append(m.hooks, hook)
}

// Execute executa todos os hooks de start registrados
func (m *StartHookManager) Execute(ctx context.Context) error {
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
func (m *StartHookManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.hooks)
}

// Clear remove todos os hooks registrados
func (m *StartHookManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.hooks = make([]interfaces.StartHookFunc, 0)
}

// Instância global para uso em toda a aplicação
var GlobalStartHookManager = NewStartHookManager()

// RegisterStartHook registra um hook de start globalmente
func RegisterStartHook(hook interfaces.StartHookFunc) {
	GlobalStartHookManager.Register(hook)
}

// ExecuteStartHooks executa todos os hooks de start globais
func ExecuteStartHooks(ctx context.Context) error {
	return GlobalStartHookManager.Execute(ctx)
}

// GetStartHookCount retorna o número de hooks de start globais
func GetStartHookCount() int {
	return GlobalStartHookManager.Count()
}

// ClearStartHooks limpa todos os hooks de start globais
func ClearStartHooks() {
	GlobalStartHookManager.Clear()
}
