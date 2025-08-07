package hooks

import (
	"context"
	"sync"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// ErrorHookManager gerencia hooks de erro
type ErrorHookManager struct {
	hooks []interfaces.ErrorHookFunc
	mu    sync.RWMutex
}

// NewErrorHookManager cria um novo gerenciador de hooks de erro
func NewErrorHookManager() *ErrorHookManager {
	return &ErrorHookManager{
		hooks: make([]interfaces.ErrorHookFunc, 0),
	}
}

// Register registra um hook de erro
func (m *ErrorHookManager) Register(hook interfaces.ErrorHookFunc) {
	if hook == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.hooks = append(m.hooks, hook)
}

// Execute executa todos os hooks de erro registrados
func (m *ErrorHookManager) Execute(ctx context.Context, err interfaces.DomainErrorInterface) error {
	if err == nil {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, hook := range m.hooks {
		if hookErr := hook(ctx, err); hookErr != nil {
			return hookErr
		}
	}

	return nil
}

// Count retorna o número de hooks registrados
func (m *ErrorHookManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.hooks)
}

// Clear remove todos os hooks registrados
func (m *ErrorHookManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.hooks = make([]interfaces.ErrorHookFunc, 0)
}

// Instância global para uso em toda a aplicação
var GlobalErrorHookManager = NewErrorHookManager()

// RegisterErrorHook registra um hook de erro globalmente
func RegisterErrorHook(hook interfaces.ErrorHookFunc) {
	GlobalErrorHookManager.Register(hook)
}

// ExecuteErrorHooks executa todos os hooks de erro globais
func ExecuteErrorHooks(ctx context.Context, err interfaces.DomainErrorInterface) error {
	return GlobalErrorHookManager.Execute(ctx, err)
}

// GetErrorHookCount retorna o número de hooks de erro globais
func GetErrorHookCount() int {
	return GlobalErrorHookManager.Count()
}

// ClearErrorHooks limpa todos os hooks de erro globais
func ClearErrorHooks() {
	GlobalErrorHookManager.Clear()
}
