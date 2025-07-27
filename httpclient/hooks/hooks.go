// Package hooks provides hook chain support for request/response processing.
package hooks

import (
	"context"
	"sync"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// Manager manages hook execution for HTTP operations.
type Manager struct {
	hooks []interfaces.Hook
	mu    sync.RWMutex
}

// NewManager creates a new hook manager.
func NewManager() *Manager {
	return &Manager{
		hooks: make([]interfaces.Hook, 0),
	}
}

// Add adds a hook to the manager.
func (m *Manager) Add(hook interfaces.Hook) *Manager {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.hooks = append(m.hooks, hook)
	return m
}

// Remove removes a hook from the manager.
func (m *Manager) Remove(hook interfaces.Hook) *Manager {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, h := range m.hooks {
		if h == hook {
			m.hooks = append(m.hooks[:i], m.hooks[i+1:]...)
			break
		}
	}
	return m
}

// ExecuteBeforeRequest executes all BeforeRequest hooks.
func (m *Manager) ExecuteBeforeRequest(ctx context.Context, req *interfaces.Request) error {
	m.mu.RLock()
	hooks := make([]interfaces.Hook, len(m.hooks))
	copy(hooks, m.hooks)
	m.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook.BeforeRequest(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteAfterResponse executes all AfterResponse hooks.
func (m *Manager) ExecuteAfterResponse(ctx context.Context, req *interfaces.Request, resp *interfaces.Response) error {
	m.mu.RLock()
	hooks := make([]interfaces.Hook, len(m.hooks))
	copy(hooks, m.hooks)
	m.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook.AfterResponse(ctx, req, resp); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteOnError executes all OnError hooks.
func (m *Manager) ExecuteOnError(ctx context.Context, req *interfaces.Request, err error) error {
	m.mu.RLock()
	hooks := make([]interfaces.Hook, len(m.hooks))
	copy(hooks, m.hooks)
	m.mu.RUnlock()

	for _, hook := range hooks {
		if hookErr := hook.OnError(ctx, req, err); hookErr != nil {
			return hookErr
		}
	}
	return nil
}

// Count returns the number of hooks in the manager.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.hooks)
}

// Clear removes all hooks from the manager.
func (m *Manager) Clear() *Manager {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks = m.hooks[:0]
	return m
}

// Built-in hook implementations

// TimingHook measures request timing.
type TimingHook struct {
	callback func(method, url string, duration int64) // duration in nanoseconds
}

// NewTimingHook creates a new timing hook.
func NewTimingHook(callback func(method, url string, duration int64)) *TimingHook {
	return &TimingHook{callback: callback}
}

// BeforeRequest implements the Hook interface.
func (h *TimingHook) BeforeRequest(ctx context.Context, req *interfaces.Request) error {
	if req.Context == nil {
		req.Context = ctx
	}
	// Store start time in context (would need custom context key in real implementation)
	return nil
}

// AfterResponse implements the Hook interface.
func (h *TimingHook) AfterResponse(ctx context.Context, req *interfaces.Request, resp *interfaces.Response) error {
	if h.callback != nil && resp != nil {
		duration := resp.Latency.Nanoseconds()
		h.callback(req.Method, req.URL, duration)
	}
	return nil
}

// OnError implements the Hook interface.
func (h *TimingHook) OnError(ctx context.Context, req *interfaces.Request, err error) error {
	// Could measure error timing here too
	return nil
}

// LoggingHook logs request lifecycle events.
type LoggingHook struct {
	logger func(level string, format string, args ...interface{})
}

// NewLoggingHook creates a new logging hook.
func NewLoggingHook(logger func(level string, format string, args ...interface{})) *LoggingHook {
	return &LoggingHook{logger: logger}
}

// BeforeRequest implements the Hook interface.
func (h *LoggingHook) BeforeRequest(ctx context.Context, req *interfaces.Request) error {
	if h.logger != nil {
		h.logger("INFO", "Starting request: %s %s", req.Method, req.URL)
	}
	return nil
}

// AfterResponse implements the Hook interface.
func (h *LoggingHook) AfterResponse(ctx context.Context, req *interfaces.Request, resp *interfaces.Response) error {
	if h.logger != nil {
		h.logger("INFO", "Request completed: %s %s - Status: %d - Duration: %v",
			req.Method, req.URL, resp.StatusCode, resp.Latency)
	}
	return nil
}

// OnError implements the Hook interface.
func (h *LoggingHook) OnError(ctx context.Context, req *interfaces.Request, err error) error {
	if h.logger != nil {
		h.logger("ERROR", "Request failed: %s %s - Error: %v", req.Method, req.URL, err)
	}
	return nil
}

// ValidationHook validates requests before execution.
type ValidationHook struct {
	validator func(*interfaces.Request) error
}

// NewValidationHook creates a new validation hook.
func NewValidationHook(validator func(*interfaces.Request) error) *ValidationHook {
	return &ValidationHook{validator: validator}
}

// BeforeRequest implements the Hook interface.
func (h *ValidationHook) BeforeRequest(ctx context.Context, req *interfaces.Request) error {
	if h.validator != nil {
		return h.validator(req)
	}
	return nil
}

// AfterResponse implements the Hook interface.
func (h *ValidationHook) AfterResponse(ctx context.Context, req *interfaces.Request, resp *interfaces.Response) error {
	return nil
}

// OnError implements the Hook interface.
func (h *ValidationHook) OnError(ctx context.Context, req *interfaces.Request, err error) error {
	return nil
}

// MetricsHook collects detailed metrics.
type MetricsHook struct {
	collector func(event string, data map[string]interface{})
}

// NewMetricsHook creates a new metrics hook.
func NewMetricsHook(collector func(event string, data map[string]interface{})) *MetricsHook {
	return &MetricsHook{collector: collector}
}

// BeforeRequest implements the Hook interface.
func (h *MetricsHook) BeforeRequest(ctx context.Context, req *interfaces.Request) error {
	if h.collector != nil {
		data := map[string]interface{}{
			"method": req.Method,
			"url":    req.URL,
		}
		h.collector("request_started", data)
	}
	return nil
}

// AfterResponse implements the Hook interface.
func (h *MetricsHook) AfterResponse(ctx context.Context, req *interfaces.Request, resp *interfaces.Response) error {
	if h.collector != nil {
		data := map[string]interface{}{
			"method":      req.Method,
			"url":         req.URL,
			"status_code": resp.StatusCode,
			"duration":    resp.Latency.Nanoseconds(),
			"success":     !resp.IsError,
		}
		h.collector("request_completed", data)
	}
	return nil
}

// OnError implements the Hook interface.
func (h *MetricsHook) OnError(ctx context.Context, req *interfaces.Request, err error) error {
	if h.collector != nil {
		data := map[string]interface{}{
			"method": req.Method,
			"url":    req.URL,
			"error":  err.Error(),
		}
		h.collector("request_failed", data)
	}
	return nil
}

// CircuitBreakerHook implements circuit breaker pattern.
type CircuitBreakerHook struct {
	failures    int
	maxFailures int
	isOpen      bool
	mu          sync.RWMutex
}

// NewCircuitBreakerHook creates a new circuit breaker hook.
func NewCircuitBreakerHook(maxFailures int) *CircuitBreakerHook {
	return &CircuitBreakerHook{
		maxFailures: maxFailures,
	}
}

// BeforeRequest implements the Hook interface.
func (h *CircuitBreakerHook) BeforeRequest(ctx context.Context, req *interfaces.Request) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.isOpen {
		return &CircuitBreakerError{Message: "Circuit breaker is open"}
	}
	return nil
}

// AfterResponse implements the Hook interface.
func (h *CircuitBreakerHook) AfterResponse(ctx context.Context, req *interfaces.Request, resp *interfaces.Response) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if resp.IsError || resp.StatusCode >= 500 {
		h.failures++
		if h.failures >= h.maxFailures {
			h.isOpen = true
		}
	} else {
		h.failures = 0
		h.isOpen = false
	}
	return nil
}

// OnError implements the Hook interface.
func (h *CircuitBreakerHook) OnError(ctx context.Context, req *interfaces.Request, err error) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.failures++
	if h.failures >= h.maxFailures {
		h.isOpen = true
	}
	return nil
}

// IsOpen returns whether the circuit breaker is open.
func (h *CircuitBreakerHook) IsOpen() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isOpen
}

// Reset resets the circuit breaker.
func (h *CircuitBreakerHook) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.failures = 0
	h.isOpen = false
}

// CircuitBreakerError represents a circuit breaker error.
type CircuitBreakerError struct {
	Message string
}

func (e *CircuitBreakerError) Error() string {
	return e.Message
}

// HookFunc is a function adapter for simple hooks.
type HookFunc struct {
	BeforeFunc func(context.Context, *interfaces.Request) error
	AfterFunc  func(context.Context, *interfaces.Request, *interfaces.Response) error
	ErrorFunc  func(context.Context, *interfaces.Request, error) error
}

// NewHookFunc creates a new function-based hook.
func NewHookFunc(
	before func(context.Context, *interfaces.Request) error,
	after func(context.Context, *interfaces.Request, *interfaces.Response) error,
	onError func(context.Context, *interfaces.Request, error) error,
) *HookFunc {
	return &HookFunc{
		BeforeFunc: before,
		AfterFunc:  after,
		ErrorFunc:  onError,
	}
}

// BeforeRequest implements the Hook interface.
func (h *HookFunc) BeforeRequest(ctx context.Context, req *interfaces.Request) error {
	if h.BeforeFunc != nil {
		return h.BeforeFunc(ctx, req)
	}
	return nil
}

// AfterResponse implements the Hook interface.
func (h *HookFunc) AfterResponse(ctx context.Context, req *interfaces.Request, resp *interfaces.Response) error {
	if h.AfterFunc != nil {
		return h.AfterFunc(ctx, req, resp)
	}
	return nil
}

// OnError implements the Hook interface.
func (h *HookFunc) OnError(ctx context.Context, req *interfaces.Request, err error) error {
	if h.ErrorFunc != nil {
		return h.ErrorFunc(ctx, req, err)
	}
	return nil
}
