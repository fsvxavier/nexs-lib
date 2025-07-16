package hooks

import (
	"fmt"
	"sync"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

// DefaultHookManager implements the HookManager interface
type DefaultHookManager struct {
	hooks       map[interfaces.HookType][]interfaces.Hook
	customHooks map[interfaces.HookType]map[string]interfaces.Hook
	mutex       sync.RWMutex
}

// NewDefaultHookManager creates a new hook manager
func NewDefaultHookManager() *DefaultHookManager {
	return &DefaultHookManager{
		hooks:       make(map[interfaces.HookType][]interfaces.Hook),
		customHooks: make(map[interfaces.HookType]map[string]interfaces.Hook),
	}
}

// RegisterHook registers a hook for a specific type
func (hm *DefaultHookManager) RegisterHook(hookType interfaces.HookType, hook interfaces.Hook) error {
	if hook == nil {
		return fmt.Errorf("hook cannot be nil")
	}

	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	if hm.hooks[hookType] == nil {
		hm.hooks[hookType] = make([]interfaces.Hook, 0)
	}

	hm.hooks[hookType] = append(hm.hooks[hookType], hook)
	return nil
}

// RegisterCustomHook registers a custom hook with a custom type
func (hm *DefaultHookManager) RegisterCustomHook(hookType interfaces.HookType, name string, hook interfaces.Hook) error {
	if hook == nil {
		return fmt.Errorf("hook cannot be nil")
	}

	if name == "" {
		return fmt.Errorf("custom hook name cannot be empty")
	}

	if hookType < interfaces.CustomHookBase {
		return fmt.Errorf("custom hook type must be >= CustomHookBase")
	}

	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	if hm.customHooks[hookType] == nil {
		hm.customHooks[hookType] = make(map[string]interfaces.Hook)
	}

	hm.customHooks[hookType][name] = hook
	return nil
}

// ExecuteHooks executes all hooks of a specific type
func (hm *DefaultHookManager) ExecuteHooks(hookType interfaces.HookType, ctx *interfaces.ExecutionContext) error {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	// Execute standard hooks
	if hooks, exists := hm.hooks[hookType]; exists {
		for _, hook := range hooks {
			if result := hook(ctx); result != nil {
				if result.Error != nil {
					return fmt.Errorf("hook execution failed: %w", result.Error)
				}
				if !result.Continue {
					return fmt.Errorf("hook requested execution stop")
				}
			}
		}
	}

	// Execute custom hooks
	if customHooks, exists := hm.customHooks[hookType]; exists {
		for name, hook := range customHooks {
			if result := hook(ctx); result != nil {
				if result.Error != nil {
					return fmt.Errorf("custom hook '%s' execution failed: %w", name, result.Error)
				}
				if !result.Continue {
					return fmt.Errorf("custom hook '%s' requested execution stop", name)
				}
			}
		}
	}

	return nil
}

// UnregisterHook removes a hook
func (hm *DefaultHookManager) UnregisterHook(hookType interfaces.HookType) error {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	delete(hm.hooks, hookType)
	return nil
}

// UnregisterCustomHook removes a custom hook
func (hm *DefaultHookManager) UnregisterCustomHook(hookType interfaces.HookType, name string) error {
	if name == "" {
		return fmt.Errorf("custom hook name cannot be empty")
	}

	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	if customHooks, exists := hm.customHooks[hookType]; exists {
		delete(customHooks, name)
		if len(customHooks) == 0 {
			delete(hm.customHooks, hookType)
		}
	}

	return nil
}

// ListHooks returns all registered hooks
func (hm *DefaultHookManager) ListHooks() map[interfaces.HookType][]interfaces.Hook {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	result := make(map[interfaces.HookType][]interfaces.Hook)

	// Copy standard hooks
	for hookType, hooks := range hm.hooks {
		result[hookType] = make([]interfaces.Hook, len(hooks))
		copy(result[hookType], hooks)
	}

	// Add custom hooks
	for hookType, customHooks := range hm.customHooks {
		if result[hookType] == nil {
			result[hookType] = make([]interfaces.Hook, 0)
		}
		for _, hook := range customHooks {
			result[hookType] = append(result[hookType], hook)
		}
	}

	return result
}

// GetHookCount returns the number of registered hooks for a specific type
func (hm *DefaultHookManager) GetHookCount(hookType interfaces.HookType) int {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	count := 0

	if hooks, exists := hm.hooks[hookType]; exists {
		count += len(hooks)
	}

	if customHooks, exists := hm.customHooks[hookType]; exists {
		count += len(customHooks)
	}

	return count
}

// HasHooks returns true if there are registered hooks for the specified type
func (hm *DefaultHookManager) HasHooks(hookType interfaces.HookType) bool {
	return hm.GetHookCount(hookType) > 0
}

// ClearHooks removes all hooks
func (hm *DefaultHookManager) ClearHooks() {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	hm.hooks = make(map[interfaces.HookType][]interfaces.Hook)
	hm.customHooks = make(map[interfaces.HookType]map[string]interfaces.Hook)
}

// ClearHooksOfType removes all hooks of a specific type
func (hm *DefaultHookManager) ClearHooksOfType(hookType interfaces.HookType) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	delete(hm.hooks, hookType)
	delete(hm.customHooks, hookType)
}
