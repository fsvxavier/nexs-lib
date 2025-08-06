package hooks

import (
	"errors"
	"sort"
	"sync"

	interfaces "github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

var (
	// ErrHookExecutionFailed é retornado quando a execução de um hook falha
	ErrHookExecutionFailed = errors.New("hook execution failed")
	// ErrInvalidHookChain é retornado quando a cadeia de hooks é inválida
	ErrInvalidHookChain = errors.New("invalid hook chain")
)

// DefaultHookChain implementação padrão da cadeia de hooks
type DefaultHookChain struct {
	hooks []interfaces.Hook
	mutex sync.RWMutex
}

// NewDefaultHookChain cria uma nova instância da cadeia padrão
func NewDefaultHookChain() *DefaultHookChain {
	return &DefaultHookChain{
		hooks: make([]interfaces.Hook, 0),
	}
}

// Add adiciona um hook à cadeia
func (c *DefaultHookChain) Add(hook interfaces.Hook) interfaces.HookChain {
	if hook == nil {
		return c
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Verifica se o hook já existe
	for _, existing := range c.hooks {
		if existing.Name() == hook.Name() {
			return c // Ignora silenciosamente se já existe
		}
	}

	// Adiciona o hook
	c.hooks = append(c.hooks, hook)

	// Ordena por prioridade (0 = maior prioridade)
	sort.Slice(c.hooks, func(i, j int) bool {
		return c.hooks[i].Priority() < c.hooks[j].Priority()
	})

	return c
}

// Execute executa todos os hooks na cadeia
func (c *DefaultHookChain) Execute(ctx *interfaces.HookContext) error {
	c.mutex.RLock()
	hooks := c.getEnabledHooks()
	c.mutex.RUnlock()

	for _, hook := range hooks {
		if err := hook.Execute(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Size retorna o número de hooks na cadeia
func (c *DefaultHookChain) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.hooks)
}

// Clear limpa a cadeia
func (c *DefaultHookChain) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.hooks = make([]interfaces.Hook, 0)
}

// getEnabledHooks retorna apenas hooks habilitados (método interno)
func (c *DefaultHookChain) getEnabledHooks() []interfaces.Hook {
	enabledHooks := make([]interfaces.Hook, 0, len(c.hooks))
	for _, hook := range c.hooks {
		if hook.Enabled() {
			enabledHooks = append(enabledHooks, hook)
		}
	}
	return enabledHooks
}

// GlobalHookChain instância global da cadeia de hooks
var GlobalHookChain interfaces.HookChain = NewDefaultHookChain()

// AddHook adiciona um hook à cadeia global
func AddHook(hook interfaces.Hook) interfaces.HookChain {
	return GlobalHookChain.Add(hook)
}

// ExecuteHookChain executa toda a cadeia global de hooks
func ExecuteHookChain(ctx *interfaces.HookContext) error {
	return GlobalHookChain.Execute(ctx)
}

// ClearHookChain limpa todos os hooks da cadeia global
func ClearHookChain() {
	GlobalHookChain.Clear()
}

// SizeHookChain retorna o número de hooks na cadeia global
func SizeHookChain() int {
	return GlobalHookChain.Size()
}
