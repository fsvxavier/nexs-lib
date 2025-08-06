package hooks

import (
	"errors"
	"sort"
	"sync"

	interfaces "github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

var (
	// ErrHookAlreadyExists é retornado quando tentamos registrar um hook que já existe
	ErrHookAlreadyExists = errors.New("hook already exists")
	// ErrHookNotFound é retornado quando um hook não é encontrado
	ErrHookNotFound = errors.New("hook not found")
	// ErrInvalidHook é retornado quando um hook é inválido
	ErrInvalidHook = errors.New("invalid hook")
)

// DefaultHookRegistry implementação padrão do HookRegistry
type DefaultHookRegistry struct {
	hooks map[interfaces.HookType][]interfaces.Hook
	mutex sync.RWMutex
}

// NewDefaultHookRegistry cria uma nova instância do registry padrão
func NewDefaultHookRegistry() *DefaultHookRegistry {
	return &DefaultHookRegistry{
		hooks: make(map[interfaces.HookType][]interfaces.Hook),
	}
}

// Register registra um hook
func (r *DefaultHookRegistry) Register(hook interfaces.Hook) error {
	if hook == nil {
		return ErrInvalidHook
	}

	if hook.Name() == "" {
		return errors.New("hook name cannot be empty")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	hookType := hook.Type()
	hooks := r.hooks[hookType]

	// Verifica se o hook já existe
	for _, existingHook := range hooks {
		if existingHook.Name() == hook.Name() {
			return ErrHookAlreadyExists
		}
	}

	// Adiciona o hook
	hooks = append(hooks, hook)

	// Ordena por prioridade (0 = maior prioridade)
	sort.Slice(hooks, func(i, j int) bool {
		return hooks[i].Priority() < hooks[j].Priority()
	})

	r.hooks[hookType] = hooks
	return nil
}

// Unregister remove um hook pelo nome
func (r *DefaultHookRegistry) Unregister(name string) error {
	if name == "" {
		return errors.New("hook name cannot be empty")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	found := false
	for hookType, hooks := range r.hooks {
		for i, hook := range hooks {
			if hook.Name() == name {
				// Remove o hook
				r.hooks[hookType] = append(hooks[:i], hooks[i+1:]...)
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return ErrHookNotFound
	}

	return nil
}

// GetHooks retorna hooks por tipo ordenados por prioridade
func (r *DefaultHookRegistry) GetHooks(hookType interfaces.HookType) []interfaces.Hook {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	hooks := r.hooks[hookType]
	if hooks == nil {
		return []interfaces.Hook{}
	}

	// Retorna apenas hooks habilitados
	enabledHooks := make([]interfaces.Hook, 0, len(hooks))
	for _, hook := range hooks {
		if hook.Enabled() {
			enabledHooks = append(enabledHooks, hook)
		}
	}

	return enabledHooks
}

// ExecuteHooks executa todos os hooks de um tipo
func (r *DefaultHookRegistry) ExecuteHooks(ctx *interfaces.HookContext, hookType interfaces.HookType) error {
	hooks := r.GetHooks(hookType)

	for _, hook := range hooks {
		if err := hook.Execute(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Clear remove todos os hooks
func (r *DefaultHookRegistry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.hooks = make(map[interfaces.HookType][]interfaces.Hook)
}

// Count retorna o número de hooks registrados
func (r *DefaultHookRegistry) Count() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	count := 0
	for _, hooks := range r.hooks {
		count += len(hooks)
	}

	return count
}

// ListAll retorna todos os hooks registrados
func (r *DefaultHookRegistry) ListAll() map[interfaces.HookType][]interfaces.Hook {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result := make(map[interfaces.HookType][]interfaces.Hook)
	for hookType, hooks := range r.hooks {
		result[hookType] = make([]interfaces.Hook, len(hooks))
		copy(result[hookType], hooks)
	}

	return result
}

// GlobalHookRegistry instância global do registry
var GlobalHookRegistry interfaces.HookRegistry = NewDefaultHookRegistry()

// RegisterHook registra um hook no registry global
func RegisterHook(hook interfaces.Hook) error {
	return GlobalHookRegistry.Register(hook)
}

// UnregisterHook remove um hook do registry global
func UnregisterHook(name string) error {
	return GlobalHookRegistry.Unregister(name)
}

// GetHooks retorna hooks do registry global
func GetHooks(hookType interfaces.HookType) []interfaces.Hook {
	return GlobalHookRegistry.GetHooks(hookType)
}

// ExecuteHooks executa hooks do registry global
func ExecuteHooks(ctx *interfaces.HookContext, hookType interfaces.HookType) error {
	return GlobalHookRegistry.ExecuteHooks(ctx, hookType)
}

// ClearHooks limpa todos os hooks do registry global
func ClearHooks() {
	GlobalHookRegistry.Clear()
}

// CountHooks retorna o número total de hooks no registry global
func CountHooks() int {
	return GlobalHookRegistry.Count()
}

// ListAllHooks retorna todos os hooks do registry global
func ListAllHooks() map[interfaces.HookType][]interfaces.Hook {
	return GlobalHookRegistry.ListAll()
}
