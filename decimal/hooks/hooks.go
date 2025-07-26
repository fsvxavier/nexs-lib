package hooks

import (
	"context"
	"fmt"
	"sync"

	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
)

// HookManager implements the HookManager interface
type HookManager struct {
	mu         sync.RWMutex
	preHooks   []interfaces.PreHook
	postHooks  []interfaces.PostHook
	errorHooks []interfaces.ErrorHook
}

// NewHookManager creates a new hook manager
func NewHookManager() interfaces.HookManager {
	return &HookManager{
		preHooks:   make([]interfaces.PreHook, 0),
		postHooks:  make([]interfaces.PostHook, 0),
		errorHooks: make([]interfaces.ErrorHook, 0),
	}
}

// RegisterPreHook registers a pre-execution hook
func (hm *HookManager) RegisterPreHook(hook interfaces.PreHook) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.preHooks = append(hm.preHooks, hook)
}

// RegisterPostHook registers a post-execution hook
func (hm *HookManager) RegisterPostHook(hook interfaces.PostHook) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.postHooks = append(hm.postHooks, hook)
}

// RegisterErrorHook registers an error handling hook
func (hm *HookManager) RegisterErrorHook(hook interfaces.ErrorHook) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.errorHooks = append(hm.errorHooks, hook)
}

// ExecutePreHooks executes all registered pre-hooks
func (hm *HookManager) ExecutePreHooks(ctx context.Context, operation string, args ...interface{}) (interface{}, error) {
	hm.mu.RLock()
	hooks := make([]interfaces.PreHook, len(hm.preHooks))
	copy(hooks, hm.preHooks)
	hm.mu.RUnlock()

	var result interface{}
	for _, hook := range hooks {
		var err error
		result, err = hook.Execute(ctx, operation, args...)
		if err != nil {
			return nil, fmt.Errorf("pre-hook failed for operation '%s': %w", operation, err)
		}
	}

	return result, nil
}

// ExecutePostHooks executes all registered post-hooks
func (hm *HookManager) ExecutePostHooks(ctx context.Context, operation string, result interface{}, err error) error {
	hm.mu.RLock()
	hooks := make([]interfaces.PostHook, len(hm.postHooks))
	copy(hooks, hm.postHooks)
	hm.mu.RUnlock()

	for _, hook := range hooks {
		hookErr := hook.Execute(ctx, operation, result, err)
		if hookErr != nil {
			return fmt.Errorf("post-hook failed for operation '%s': %w", operation, hookErr)
		}
	}

	return nil
}

// ExecuteErrorHooks executes all registered error hooks
func (hm *HookManager) ExecuteErrorHooks(ctx context.Context, operation string, err error) error {
	hm.mu.RLock()
	hooks := make([]interfaces.ErrorHook, len(hm.errorHooks))
	copy(hooks, hm.errorHooks)
	hm.mu.RUnlock()

	for _, hook := range hooks {
		hookErr := hook.Execute(ctx, operation, err)
		if hookErr != nil {
			return fmt.Errorf("error-hook failed for operation '%s': %w", operation, hookErr)
		}
	}

	return nil
}

// ClearHooks removes all registered hooks
func (hm *HookManager) ClearHooks() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.preHooks = make([]interfaces.PreHook, 0)
	hm.postHooks = make([]interfaces.PostHook, 0)
	hm.errorHooks = make([]interfaces.ErrorHook, 0)
}

// GetHookCounts returns the number of registered hooks of each type
func (hm *HookManager) GetHookCounts() (preHooks, postHooks, errorHooks int) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	return len(hm.preHooks), len(hm.postHooks), len(hm.errorHooks)
}

// BasicLoggingHook is a simple hook that logs operations
type BasicLoggingHook struct {
	Logger func(message string)
}

// NewBasicLoggingHook creates a new basic logging hook
func NewBasicLoggingHook(logger func(string)) *BasicLoggingHook {
	return &BasicLoggingHook{Logger: logger}
}

// Execute implements PreHook interface for logging
func (h *BasicLoggingHook) Execute(ctx context.Context, operation string, args ...interface{}) (interface{}, error) {
	if h.Logger != nil {
		h.Logger(fmt.Sprintf("PRE: Operation '%s' with args: %v", operation, args))
	}
	return nil, nil
}

// BasicPostLoggingHook logs post-execution information
type BasicPostLoggingHook struct {
	Logger func(message string)
}

// NewBasicPostLoggingHook creates a new basic post logging hook
func NewBasicPostLoggingHook(logger func(string)) *BasicPostLoggingHook {
	return &BasicPostLoggingHook{Logger: logger}
}

// Execute implements PostHook interface for logging
func (h *BasicPostLoggingHook) Execute(ctx context.Context, operation string, result interface{}, err error) error {
	if h.Logger != nil {
		if err != nil {
			h.Logger(fmt.Sprintf("POST: Operation '%s' failed with error: %v", operation, err))
		} else {
			h.Logger(fmt.Sprintf("POST: Operation '%s' completed successfully", operation))
		}
	}
	return nil
}

// BasicErrorLoggingHook logs error information
type BasicErrorLoggingHook struct {
	Logger func(message string)
}

// NewBasicErrorLoggingHook creates a new basic error logging hook
func NewBasicErrorLoggingHook(logger func(string)) *BasicErrorLoggingHook {
	return &BasicErrorLoggingHook{Logger: logger}
}

// Execute implements ErrorHook interface for logging
func (h *BasicErrorLoggingHook) Execute(ctx context.Context, operation string, err error) error {
	if h.Logger != nil {
		h.Logger(fmt.Sprintf("ERROR: Operation '%s' encountered error: %v", operation, err))
	}
	return nil
}

// ValidationHook validates input parameters before execution
type ValidationHook struct {
	ValidateString func(string) error
	ValidateFloat  func(float64) error
	ValidateInt    func(int64) error
}

// NewValidationHook creates a new validation hook
func NewValidationHook() *ValidationHook {
	return &ValidationHook{}
}

// Execute implements PreHook interface for validation
func (h *ValidationHook) Execute(ctx context.Context, operation string, args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, nil
	}

	switch operation {
	case "NewFromString":
		if len(args) > 0 {
			if str, ok := args[0].(string); ok && h.ValidateString != nil {
				if err := h.ValidateString(str); err != nil {
					return nil, fmt.Errorf("string validation failed: %w", err)
				}
			}
		}
	case "NewFromFloat":
		if len(args) > 0 {
			if f, ok := args[0].(float64); ok && h.ValidateFloat != nil {
				if err := h.ValidateFloat(f); err != nil {
					return nil, fmt.Errorf("float validation failed: %w", err)
				}
			}
		}
	case "NewFromInt":
		if len(args) > 0 {
			if i, ok := args[0].(int64); ok && h.ValidateInt != nil {
				if err := h.ValidateInt(i); err != nil {
					return nil, fmt.Errorf("int validation failed: %w", err)
				}
			}
		}
	}

	return nil, nil
}

// MetricsHook tracks operation metrics
type MetricsHook struct {
	mu      sync.RWMutex
	metrics map[string]int64
}

// NewMetricsHook creates a new metrics tracking hook
func NewMetricsHook() *MetricsHook {
	return &MetricsHook{
		metrics: make(map[string]int64),
	}
}

// Execute implements PostHook interface for metrics
func (h *MetricsHook) Execute(ctx context.Context, operation string, result interface{}, err error) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := operation
	if err != nil {
		key += "_error"
	} else {
		key += "_success"
	}

	h.metrics[key]++
	return nil
}

// GetMetrics returns a copy of the current metrics
func (h *MetricsHook) GetMetrics() map[string]int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[string]int64)
	for k, v := range h.metrics {
		result[k] = v
	}
	return result
}

// ResetMetrics clears all metrics
func (h *MetricsHook) ResetMetrics() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.metrics = make(map[string]int64)
}
