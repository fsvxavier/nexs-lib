package hooks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgres/interfaces"
)

// HookManager implementa IHookManager
type HookManager struct {
	mu          sync.RWMutex
	hooks       map[interfaces.HookType][]interfaces.Hook
	customHooks map[interfaces.HookType]map[string]interfaces.Hook
	hookTimeout time.Duration
	enabled     bool
}

// NewHookManager cria um novo hook manager
func NewHookManager(timeout time.Duration) interfaces.IHookManager {
	return &HookManager{
		hooks:       make(map[interfaces.HookType][]interfaces.Hook),
		customHooks: make(map[interfaces.HookType]map[string]interfaces.Hook),
		hookTimeout: timeout,
		enabled:     true,
	}
}

// RegisterHook registra um hook para um tipo específico
func (hm *HookManager) RegisterHook(hookType interfaces.HookType, hook interfaces.Hook) error {
	if hook == nil {
		return fmt.Errorf("hook cannot be nil")
	}

	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.hooks[hookType] == nil {
		hm.hooks[hookType] = []interfaces.Hook{}
	}

	hm.hooks[hookType] = append(hm.hooks[hookType], hook)
	return nil
}

// RegisterCustomHook registra um hook customizado
func (hm *HookManager) RegisterCustomHook(hookType interfaces.HookType, name string, hook interfaces.Hook) error {
	if hook == nil {
		return fmt.Errorf("hook cannot be nil")
	}
	if name == "" {
		return fmt.Errorf("hook name cannot be empty")
	}

	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.customHooks[hookType] == nil {
		hm.customHooks[hookType] = make(map[string]interfaces.Hook)
	}

	hm.customHooks[hookType][name] = hook
	return nil
}

// ExecuteHooks executa todos os hooks de um tipo específico
func (hm *HookManager) ExecuteHooks(hookType interfaces.HookType, ctx *interfaces.ExecutionContext) error {
	if !hm.enabled {
		return nil
	}

	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Executar hooks padrão
	if hooks, exists := hm.hooks[hookType]; exists {
		for _, hook := range hooks {
			if err := hm.executeHookWithTimeout(hook, ctx); err != nil {
				return err
			}
		}
	}

	// Executar hooks customizados
	if customHooks, exists := hm.customHooks[hookType]; exists {
		for _, hook := range customHooks {
			if err := hm.executeHookWithTimeout(hook, ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

// UnregisterHook remove um hook
func (hm *HookManager) UnregisterHook(hookType interfaces.HookType) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	delete(hm.hooks, hookType)
	return nil
}

// UnregisterCustomHook remove um hook customizado
func (hm *HookManager) UnregisterCustomHook(hookType interfaces.HookType, name string) error {
	if name == "" {
		return fmt.Errorf("hook name cannot be empty")
	}

	hm.mu.Lock()
	defer hm.mu.Unlock()

	if customHooks, exists := hm.customHooks[hookType]; exists {
		delete(customHooks, name)
		if len(customHooks) == 0 {
			delete(hm.customHooks, hookType)
		}
	}

	return nil
}

// ListHooks retorna todos os hooks registrados
func (hm *HookManager) ListHooks() map[interfaces.HookType][]interfaces.Hook {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Criar cópia para evitar race conditions
	result := make(map[interfaces.HookType][]interfaces.Hook)

	// Copiar hooks padrão
	for hookType, hooks := range hm.hooks {
		hooksCopy := make([]interfaces.Hook, len(hooks))
		copy(hooksCopy, hooks)
		result[hookType] = hooksCopy
	}

	// Adicionar hooks customizados
	for hookType, customHooks := range hm.customHooks {
		if result[hookType] == nil {
			result[hookType] = []interfaces.Hook{}
		}
		for _, hook := range customHooks {
			result[hookType] = append(result[hookType], hook)
		}
	}

	return result
}

// executeHookWithTimeout executa um hook com timeout
func (hm *HookManager) executeHookWithTimeout(hook interfaces.Hook, ctx *interfaces.ExecutionContext) error {
	// Criar contexto com timeout
	timeoutCtx, cancel := context.WithTimeout(ctx.Context, hm.hookTimeout)
	defer cancel()

	// Canal para resultado
	resultChan := make(chan *interfaces.HookResult, 1)
	errorChan := make(chan error, 1)

	// Executar hook em goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errorChan <- fmt.Errorf("hook panicked: %v", r)
			}
		}()

		result := hook(ctx)
		resultChan <- result
	}()

	// Aguardar resultado ou timeout
	select {
	case result := <-resultChan:
		if result != nil && result.Error != nil {
			return result.Error
		}
		if result != nil && !result.Continue {
			return fmt.Errorf("hook requested operation to stop")
		}
		return nil
	case err := <-errorChan:
		return err
	case <-timeoutCtx.Done():
		return fmt.Errorf("hook execution timed out after %v", hm.hookTimeout)
	}
}

// SetEnabled habilita/desabilita execução de hooks
func (hm *HookManager) SetEnabled(enabled bool) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.enabled = enabled
}

// IsEnabled verifica se hooks estão habilitados
func (hm *HookManager) IsEnabled() bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.enabled
}

// ClearHooks limpa todos os hooks
func (hm *HookManager) ClearHooks() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.hooks = make(map[interfaces.HookType][]interfaces.Hook)
	hm.customHooks = make(map[interfaces.HookType]map[string]interfaces.Hook)
}

// DefaultHookManager implementa hooks padrão
type DefaultHookManager struct {
	*HookManager
}

// NewDefaultHookManager cria um hook manager com hooks padrão
func NewDefaultHookManager() interfaces.IHookManager {
	hm := &DefaultHookManager{
		HookManager: NewHookManager(5 * time.Second).(*HookManager),
	}

	// Registrar hooks padrão
	hm.registerDefaultHooks()

	return hm
}

// registerDefaultHooks registra hooks padrão do sistema
func (dhm *DefaultHookManager) registerDefaultHooks() {
	// Hook de log de erros
	dhm.RegisterHook(interfaces.OnErrorHook, func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		if ctx.Error != nil {
			// Log do erro (em produção, usar logger apropriado)
			fmt.Printf("Error in operation %s: %v\n", ctx.Operation, ctx.Error)
		}
		return &interfaces.HookResult{Continue: true}
	})

	// Hook de métricas de performance
	dhm.RegisterHook(interfaces.AfterQueryHook, func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		if ctx.Duration > 5*time.Second {
			// Log de query lenta
			fmt.Printf("Slow query detected: %s took %v\n", ctx.Query, ctx.Duration)
		}
		return &interfaces.HookResult{Continue: true}
	})

	// Hook de auditoria
	dhm.RegisterHook(interfaces.BeforeExecHook, func(ctx *interfaces.ExecutionContext) *interfaces.HookResult {
		// Log de operações de modificação
		if ctx.Operation == "exec" {
			fmt.Printf("Executing: %s\n", ctx.Query)
		}
		return &interfaces.HookResult{Continue: true}
	})
}

// CreateExecutionContext cria um contexto de execução
func CreateExecutionContext(ctx context.Context, operation, query string, args []interface{}) *interfaces.ExecutionContext {
	return &interfaces.ExecutionContext{
		Context:   ctx,
		Operation: operation,
		Query:     query,
		Args:      args,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

// FinishExecutionContext finaliza um contexto de execução
func FinishExecutionContext(execCtx *interfaces.ExecutionContext, err error, rowsAffected int64) {
	execCtx.Duration = time.Since(execCtx.StartTime)
	execCtx.Error = err
	execCtx.RowsAffected = rowsAffected
}
