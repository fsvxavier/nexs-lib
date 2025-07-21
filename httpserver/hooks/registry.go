// Package hooks provides implementations for generic HTTP server hooks.
package hooks

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// DefaultHookRegistry implements the HookRegistry interface.
type DefaultHookRegistry struct {
	hooks    map[interfaces.HookEvent][]interfaces.Hook
	hooksMap map[string]interfaces.Hook
	observer interfaces.HookObserver
	metrics  *HookMetricsCollector
	mu       sync.RWMutex
}

// NewDefaultHookRegistry creates a new default hook registry.
func NewDefaultHookRegistry() *DefaultHookRegistry {
	return &DefaultHookRegistry{
		hooks:    make(map[interfaces.HookEvent][]interfaces.Hook),
		hooksMap: make(map[string]interfaces.Hook),
		metrics:  NewHookMetricsCollector(),
	}
}

// Register registers a hook for specific events.
func (r *DefaultHookRegistry) Register(hook interfaces.Hook, events ...interfaces.HookEvent) error {
	if hook == nil {
		return fmt.Errorf("hook cannot be nil")
	}

	if len(events) == 0 {
		events = hook.Events()
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if hook already exists
	if _, exists := r.hooksMap[hook.Name()]; exists {
		return fmt.Errorf("hook %s already registered", hook.Name())
	}

	// Register hook for each event
	for _, event := range events {
		r.hooks[event] = append(r.hooks[event], hook)
		// Sort hooks by priority
		sort.Slice(r.hooks[event], func(i, j int) bool {
			return r.hooks[event][i].Priority() < r.hooks[event][j].Priority()
		})
	}

	r.hooksMap[hook.Name()] = hook
	return nil
}

// Unregister removes a hook from the registry.
func (r *DefaultHookRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.hooksMap[name]
	if !exists {
		return fmt.Errorf("hook %s not found", name)
	}

	// Remove hook from all events
	for event, hooks := range r.hooks {
		for i, h := range hooks {
			if h.Name() == name {
				r.hooks[event] = append(hooks[:i], hooks[i+1:]...)
				break
			}
		}
		// Clean up empty slices
		if len(r.hooks[event]) == 0 {
			delete(r.hooks, event)
		}
	}

	delete(r.hooksMap, name)
	return nil
}

// Execute executes all hooks for a specific event.
func (r *DefaultHookRegistry) Execute(ctx *interfaces.HookContext) error {
	if ctx == nil {
		return fmt.Errorf("hook context cannot be nil")
	}

	r.mu.RLock()
	hooks := r.hooks[ctx.Event]
	r.mu.RUnlock()

	if len(hooks) == 0 {
		return nil
	}

	var lastError error
	for _, hook := range hooks {
		if !hook.IsEnabled() {
			if r.observer != nil {
				r.observer.OnHookSkip(hook.Name(), ctx.Event, ctx, "hook disabled")
			}
			continue
		}

		if !hook.ShouldExecute(ctx) {
			if r.observer != nil {
				r.observer.OnHookSkip(hook.Name(), ctx.Event, ctx, "condition not met")
			}
			continue
		}

		start := time.Now()
		if r.observer != nil {
			r.observer.OnHookStart(hook.Name(), ctx.Event, ctx)
		}

		err := hook.Execute(ctx)
		duration := time.Since(start)

		r.metrics.RecordExecution(hook.Name(), ctx.Event, duration, err)

		if r.observer != nil {
			r.observer.OnHookEnd(hook.Name(), ctx.Event, ctx, err, duration)
		}

		if err != nil {
			if r.observer != nil {
				r.observer.OnHookError(hook.Name(), ctx.Event, ctx, err)
			}
			lastError = err
			// Continue executing other hooks even if one fails
		}
	}

	return lastError
}

// ExecuteAsync executes all hooks for a specific event asynchronously.
func (r *DefaultHookRegistry) ExecuteAsync(ctx *interfaces.HookContext) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)
		if err := r.Execute(ctx); err != nil {
			errChan <- err
		}
	}()

	return errChan
}

// GetHooks returns all hooks for a specific event.
func (r *DefaultHookRegistry) GetHooks(event interfaces.HookEvent) []interfaces.Hook {
	r.mu.RLock()
	defer r.mu.RUnlock()

	hooks := r.hooks[event]
	result := make([]interfaces.Hook, len(hooks))
	copy(result, hooks)
	return result
}

// EnableHook enables a hook by name.
func (r *DefaultHookRegistry) EnableHook(name string) error {
	r.mu.RLock()
	hook, exists := r.hooksMap[name]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("hook %s not found", name)
	}

	if enableableHook, ok := hook.(interface{ SetEnabled(bool) }); ok {
		enableableHook.SetEnabled(true)
		return nil
	}

	return fmt.Errorf("hook %s does not support enable/disable", name)
}

// DisableHook disables a hook by name.
func (r *DefaultHookRegistry) DisableHook(name string) error {
	r.mu.RLock()
	hook, exists := r.hooksMap[name]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("hook %s not found", name)
	}

	if enableableHook, ok := hook.(interface{ SetEnabled(bool) }); ok {
		enableableHook.SetEnabled(false)
		return nil
	}

	return fmt.Errorf("hook %s does not support enable/disable", name)
}

// ListHooks returns all registered hooks.
func (r *DefaultHookRegistry) ListHooks() map[string]interfaces.Hook {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]interfaces.Hook)
	for name, hook := range r.hooksMap {
		result[name] = hook
	}
	return result
}

// Clear removes all hooks.
func (r *DefaultHookRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.hooks = make(map[interfaces.HookEvent][]interfaces.Hook)
	r.hooksMap = make(map[string]interfaces.Hook)
	r.metrics.Reset()
}

// SetObserver sets the hook observer.
func (r *DefaultHookRegistry) SetObserver(observer interfaces.HookObserver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.observer = observer
}

// GetMetrics returns hook execution metrics.
func (r *DefaultHookRegistry) GetMetrics() interfaces.HookMetrics {
	return r.metrics.GetMetrics()
}

// Shutdown gracefully shuts down the hook registry.
func (r *DefaultHookRegistry) Shutdown(ctx context.Context) error {
	r.Clear()
	return nil
}

// ExecuteSequential executes hooks sequentially.
func (r *DefaultHookRegistry) ExecuteSequential(hooks []interfaces.Hook, ctx *interfaces.HookContext) error {
	var lastError error
	for _, hook := range hooks {
		if !hook.IsEnabled() || !hook.ShouldExecute(ctx) {
			continue
		}

		start := time.Now()
		if r.observer != nil {
			r.observer.OnHookStart(hook.Name(), ctx.Event, ctx)
		}

		err := hook.Execute(ctx)
		duration := time.Since(start)

		r.metrics.RecordExecution(hook.Name(), ctx.Event, duration, err)

		if r.observer != nil {
			r.observer.OnHookEnd(hook.Name(), ctx.Event, ctx, err, duration)
		}

		if err != nil {
			if r.observer != nil {
				r.observer.OnHookError(hook.Name(), ctx.Event, ctx, err)
			}
			lastError = err
		}
	}
	return lastError
}

// ExecuteParallel executes hooks in parallel.
func (r *DefaultHookRegistry) ExecuteParallel(hooks []interfaces.Hook, ctx *interfaces.HookContext) error {
	if len(hooks) == 0 {
		return nil
	}

	errorChan := make(chan error, len(hooks))
	var wg sync.WaitGroup

	for _, hook := range hooks {
		if !hook.IsEnabled() || !hook.ShouldExecute(ctx) {
			continue
		}

		wg.Add(1)
		go func(h interfaces.Hook) {
			defer wg.Done()

			start := time.Now()
			if r.observer != nil {
				r.observer.OnHookStart(h.Name(), ctx.Event, ctx)
			}

			err := h.Execute(ctx)
			duration := time.Since(start)

			r.metrics.RecordExecution(h.Name(), ctx.Event, duration, err)

			if r.observer != nil {
				r.observer.OnHookEnd(h.Name(), ctx.Event, ctx, err, duration)
			}

			if err != nil {
				if r.observer != nil {
					r.observer.OnHookError(h.Name(), ctx.Event, ctx, err)
				}
				errorChan <- err
			}
		}(hook)
	}

	go func() {
		wg.Wait()
		close(errorChan)
	}()

	// Collect any errors
	var lastError error
	for err := range errorChan {
		if err != nil {
			lastError = err
		}
	}

	return lastError
}

// ExecuteWithTimeout executes hooks with a timeout.
func (r *DefaultHookRegistry) ExecuteWithTimeout(hooks []interfaces.Hook, ctx *interfaces.HookContext, timeout time.Duration) error {
	if len(hooks) == 0 {
		return nil
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- r.ExecuteSequential(hooks, ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		return fmt.Errorf("hook execution timed out after %v", timeout)
	}
}

// ExecuteWithRetry executes hooks with retry logic.
func (r *DefaultHookRegistry) ExecuteWithRetry(hooks []interfaces.Hook, ctx *interfaces.HookContext, retries int) error {
	var lastError error
	for attempt := 0; attempt <= retries; attempt++ {
		err := r.ExecuteSequential(hooks, ctx)
		if err == nil {
			return nil
		}
		lastError = err
		if attempt < retries {
			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond) // Exponential backoff
		}
	}
	return fmt.Errorf("hook execution failed after %d retries: %w", retries, lastError)
}

// HookMetricsCollector collects metrics about hook execution.
type HookMetricsCollector struct {
	totalExecutions      int64
	successfulExecutions int64
	failedExecutions     int64
	executionLatencies   []time.Duration
	executionsByEvent    map[interfaces.HookEvent]int64
	executionsByHook     map[string]int64
	errorsByHook         map[string]int64
	latencyByHook        map[string]time.Duration
	mu                   sync.RWMutex
}

// NewHookMetricsCollector creates a new metrics collector.
func NewHookMetricsCollector() *HookMetricsCollector {
	return &HookMetricsCollector{
		executionsByEvent: make(map[interfaces.HookEvent]int64),
		executionsByHook:  make(map[string]int64),
		errorsByHook:      make(map[string]int64),
		latencyByHook:     make(map[string]time.Duration),
	}
}

// RecordExecution records a hook execution.
func (c *HookMetricsCollector) RecordExecution(hookName string, event interfaces.HookEvent, duration time.Duration, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalExecutions++
	c.executionLatencies = append(c.executionLatencies, duration)
	c.executionsByEvent[event]++
	c.executionsByHook[hookName]++
	c.latencyByHook[hookName] = duration

	if err != nil {
		c.failedExecutions++
		c.errorsByHook[hookName]++
	} else {
		c.successfulExecutions++
	}
}

// GetMetrics returns the collected metrics.
func (c *HookMetricsCollector) GetMetrics() interfaces.HookMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var avgLatency, maxLatency, minLatency time.Duration
	if len(c.executionLatencies) > 0 {
		var total time.Duration
		maxLatency = c.executionLatencies[0]
		minLatency = c.executionLatencies[0]

		for _, latency := range c.executionLatencies {
			total += latency
			if latency > maxLatency {
				maxLatency = latency
			}
			if latency < minLatency {
				minLatency = latency
			}
		}
		avgLatency = total / time.Duration(len(c.executionLatencies))
	}

	// Copy maps to avoid race conditions
	executionsByEvent := make(map[interfaces.HookEvent]int64)
	for k, v := range c.executionsByEvent {
		executionsByEvent[k] = v
	}

	executionsByHook := make(map[string]int64)
	for k, v := range c.executionsByHook {
		executionsByHook[k] = v
	}

	errorsByHook := make(map[string]int64)
	for k, v := range c.errorsByHook {
		errorsByHook[k] = v
	}

	latencyByHook := make(map[string]time.Duration)
	for k, v := range c.latencyByHook {
		latencyByHook[k] = v
	}

	return interfaces.HookMetrics{
		TotalExecutions:      c.totalExecutions,
		SuccessfulExecutions: c.successfulExecutions,
		FailedExecutions:     c.failedExecutions,
		AverageLatency:       avgLatency,
		MaxLatency:           maxLatency,
		MinLatency:           minLatency,
		ExecutionsByEvent:    executionsByEvent,
		ExecutionsByHook:     executionsByHook,
		ErrorsByHook:         errorsByHook,
		LatencyByHook:        latencyByHook,
	}
}

// Reset resets all metrics.
func (c *HookMetricsCollector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalExecutions = 0
	c.successfulExecutions = 0
	c.failedExecutions = 0
	c.executionLatencies = nil
	c.executionsByEvent = make(map[interfaces.HookEvent]int64)
	c.executionsByHook = make(map[string]int64)
	c.errorsByHook = make(map[string]int64)
	c.latencyByHook = make(map[string]time.Duration)
}

// Verify DefaultHookRegistry implements HookManager interface.
var _ interfaces.HookManager = (*DefaultHookRegistry)(nil)
