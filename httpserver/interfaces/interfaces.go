// Package interfaces defines the common interfaces for HTTP servers and observers.
package interfaces

import (
	"context"
	"net/http"
	"time"
)

// HTTPServer defines the common interface for all HTTP server implementations.
type HTTPServer interface {
	// Start starts the HTTP server.
	Start() error
	// Stop gracefully stops the HTTP server with the given context.
	Stop(ctx context.Context) error
	// SetHandler sets the HTTP handler for the server.
	SetHandler(handler http.Handler)
	// GetAddr returns the server address.
	GetAddr() string
	// IsRunning returns true if the server is currently running.
	IsRunning() bool
	// GracefulStop performs a graceful shutdown with connection draining.
	GracefulStop(ctx context.Context, drainTimeout time.Duration) error
	// Restart performs a zero-downtime restart.
	Restart(ctx context.Context) error
	// GetConnectionsCount returns the number of active connections.
	GetConnectionsCount() int64
	// GetHealthStatus returns the current health status.
	GetHealthStatus() HealthStatus
}

// HealthStatus represents the health status of the server.
type HealthStatus struct {
	Status      string                 `json:"status"`
	Version     string                 `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
	Uptime      time.Duration          `json:"uptime"`
	Connections int64                  `json:"connections"`
	Checks      map[string]HealthCheck `json:"checks"`
}

// HealthCheck represents an individual health check.
type HealthCheck struct {
	Status    string        `json:"status"`
	Message   string        `json:"message"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// GracefulServer extends HTTPServer with advanced graceful operations.
type GracefulServer interface {
	HTTPServer
	// PreShutdownHook registers a function to be called before shutdown.
	PreShutdownHook(hook func() error)
	// PostShutdownHook registers a function to be called after shutdown.
	PostShutdownHook(hook func() error)
	// SetDrainTimeout sets the timeout for connection draining.
	SetDrainTimeout(timeout time.Duration)
	// WaitForConnections waits for all connections to finish or timeout.
	WaitForConnections(ctx context.Context) error
}

// ServerObserver defines the interface for observing server lifecycle events.
type ServerObserver interface {
	// OnStart is called when the server starts.
	OnStart(name string)
	// OnStop is called when the server stops.
	OnStop(name string)
	// OnRequest is called for each HTTP request with timing and status information.
	OnRequest(name string, req *http.Request, status int, duration time.Duration)
	// OnBeforeRequest is called before processing a request.
	OnBeforeRequest(name string, req *http.Request)
	// OnAfterRequest is called after processing a request.
	OnAfterRequest(name string, req *http.Request, status int, duration time.Duration)
}

// ServerFactory defines the function signature for creating HTTP servers.
type ServerFactory func(config interface{}) (HTTPServer, error)

// Middleware defines the interface for all middleware implementations.
type Middleware interface {
	// Wrap wraps an http.Handler with the middleware functionality.
	Wrap(next http.Handler) http.Handler
	// Name returns the middleware name for identification.
	Name() string
	// Priority returns the middleware priority (lower numbers execute first).
	Priority() int
}

// MiddlewareChain represents a chain of middleware.
type MiddlewareChain interface {
	// Add adds middleware to the chain.
	Add(middleware Middleware) MiddlewareChain
	// Then wraps the given handler with all middleware in the chain.
	Then(h http.Handler) http.Handler
}

// MiddlewareConfig represents common middleware configuration.
type MiddlewareConfig interface {
	// IsEnabled returns whether the middleware is enabled.
	IsEnabled() bool
	// ShouldSkip checks if a path should be skipped.
	ShouldSkip(path string) bool
}

// HealthChecker defines the interface for health check implementations.
type HealthChecker interface {
	// Check performs a health check and returns the result.
	Check(ctx context.Context) HealthCheckResult
	// Name returns the name of the health check.
	Name() string
	// Type returns the type of health check (liveness, readiness, startup).
	Type() string
}

// HealthCheckResult represents the result of a health check.
type HealthCheckResult struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
	Critical  bool                   `json:"critical"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// RateLimiter defines the interface for rate limiting implementations.
type RateLimiter interface {
	// Allow checks if a request should be allowed.
	Allow(key string) bool
	// Reset resets the rate limit for a key.
	Reset(key string)
	// GetLimit returns the current limit.
	GetLimit() int
	// GetRemaining returns remaining requests for a key.
	GetRemaining(key string) int
}

// Compressor defines the interface for compression implementations.
type Compressor interface {
	// Compress compresses the response using the specified algorithm.
	Compress(algorithm string, data []byte) ([]byte, error)
	// SupportedAlgorithms returns list of supported compression algorithms.
	SupportedAlgorithms() []string
}

// HookEvent represents different types of hook events.
type HookEvent string

const (
	// Server lifecycle events
	HookEventServerStart    HookEvent = "server.start"
	HookEventServerStop     HookEvent = "server.stop"
	HookEventServerRestart  HookEvent = "server.restart"
	HookEventServerError    HookEvent = "server.error"
	HookEventServerReady    HookEvent = "server.ready"
	HookEventServerShutdown HookEvent = "server.shutdown"

	// Request lifecycle events
	HookEventRequestStart   HookEvent = "request.start"
	HookEventRequestEnd     HookEvent = "request.end"
	HookEventRequestError   HookEvent = "request.error"
	HookEventRequestTimeout HookEvent = "request.timeout"
	HookEventRequestPanic   HookEvent = "request.panic"
	HookEventRequestRetry   HookEvent = "request.retry"

	// Middleware events
	HookEventMiddlewareStart HookEvent = "middleware.start"
	HookEventMiddlewareEnd   HookEvent = "middleware.end"
	HookEventMiddlewareError HookEvent = "middleware.error"
	HookEventMiddlewareSkip  HookEvent = "middleware.skip"

	// Health check events
	HookEventHealthCheck HookEvent = "health.check"
	HookEventHealthPass  HookEvent = "health.pass"
	HookEventHealthFail  HookEvent = "health.fail"

	// Rate limiting events
	HookEventRateLimitHit   HookEvent = "ratelimit.hit"
	HookEventRateLimitAllow HookEvent = "ratelimit.allow"
	HookEventRateLimitDeny  HookEvent = "ratelimit.deny"

	// Compression events
	HookEventCompressionStart HookEvent = "compression.start"
	HookEventCompressionEnd   HookEvent = "compression.end"
	HookEventCompressionSkip  HookEvent = "compression.skip"

	// Authentication events
	HookEventAuthStart   HookEvent = "auth.start"
	HookEventAuthSuccess HookEvent = "auth.success"
	HookEventAuthFailure HookEvent = "auth.failure"
	HookEventAuthSkip    HookEvent = "auth.skip"

	// Custom events
	HookEventCustom HookEvent = "custom"
)

// HookContext provides context information for hook execution.
type HookContext struct {
	Event         HookEvent              `json:"event"`
	ServerName    string                 `json:"server_name"`
	Timestamp     time.Time              `json:"timestamp"`
	Request       *http.Request          `json:"-"`
	Response      http.ResponseWriter    `json:"-"`
	StatusCode    int                    `json:"status_code,omitempty"`
	Duration      time.Duration          `json:"duration,omitempty"`
	Error         error                  `json:"-"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Span          interface{}            `json:"-"` // For tracing spans
	Logger        interface{}            `json:"-"` // For structured logging
	TraceID       string                 `json:"trace_id,omitempty"`
	SpanID        string                 `json:"span_id,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
}

// Hook defines the interface for generic hooks.
type Hook interface {
	// Execute executes the hook with the given context.
	Execute(ctx *HookContext) error
	// Name returns the hook name for identification.
	Name() string
	// Events returns the events this hook handles.
	Events() []HookEvent
	// Priority returns the hook priority (lower numbers execute first).
	Priority() int
	// IsEnabled returns whether the hook is enabled.
	IsEnabled() bool
	// ShouldExecute determines if the hook should execute for the given context.
	ShouldExecute(ctx *HookContext) bool
}

// AsyncHook defines the interface for asynchronous hooks.
type AsyncHook interface {
	Hook
	// ExecuteAsync executes the hook asynchronously.
	ExecuteAsync(ctx *HookContext) <-chan error
	// BufferSize returns the buffer size for async execution.
	BufferSize() int
	// Timeout returns the timeout for async execution.
	Timeout() time.Duration
}

// ConditionalHook defines the interface for conditional hooks.
type ConditionalHook interface {
	Hook
	// Condition returns a function that determines if the hook should execute.
	Condition() func(ctx *HookContext) bool
}

// FilteredHook defines the interface for hooks with filtering capabilities.
type FilteredHook interface {
	Hook
	// PathFilter returns a function to filter by request path.
	PathFilter() func(path string) bool
	// MethodFilter returns a function to filter by HTTP method.
	MethodFilter() func(method string) bool
	// HeaderFilter returns a function to filter by request headers.
	HeaderFilter() func(headers http.Header) bool
}

// HookRegistry defines the interface for hook management.
type HookRegistry interface {
	// Register registers a hook for specific events.
	Register(hook Hook, events ...HookEvent) error
	// Unregister removes a hook from the registry.
	Unregister(name string) error
	// Execute executes all hooks for a specific event.
	Execute(ctx *HookContext) error
	// ExecuteAsync executes all hooks for a specific event asynchronously.
	ExecuteAsync(ctx *HookContext) <-chan error
	// GetHooks returns all hooks for a specific event.
	GetHooks(event HookEvent) []Hook
	// EnableHook enables a hook by name.
	EnableHook(name string) error
	// DisableHook disables a hook by name.
	DisableHook(name string) error
	// ListHooks returns all registered hooks.
	ListHooks() map[string]Hook
	// Clear removes all hooks.
	Clear()
}

// HookExecutor defines the interface for hook execution.
type HookExecutor interface {
	// ExecuteSequential executes hooks sequentially.
	ExecuteSequential(hooks []Hook, ctx *HookContext) error
	// ExecuteParallel executes hooks in parallel.
	ExecuteParallel(hooks []Hook, ctx *HookContext) error
	// ExecuteWithTimeout executes hooks with a timeout.
	ExecuteWithTimeout(hooks []Hook, ctx *HookContext, timeout time.Duration) error
	// ExecuteWithRetry executes hooks with retry logic.
	ExecuteWithRetry(hooks []Hook, ctx *HookContext, retries int) error
}

// HookChain defines the interface for chaining hooks.
type HookChain interface {
	// Add adds a hook to the chain.
	Add(hook Hook) HookChain
	// Execute executes all hooks in the chain.
	Execute(ctx *HookContext) error
	// ExecuteUntil executes hooks until a condition is met.
	ExecuteUntil(ctx *HookContext, condition func(*HookContext) bool) error
	// ExecuteIf executes hooks if a condition is met.
	ExecuteIf(ctx *HookContext, condition func(*HookContext) bool) error
}

// HookObserver defines the interface for observing hook execution.
type HookObserver interface {
	// OnHookStart is called when a hook starts executing.
	OnHookStart(name string, event HookEvent, ctx *HookContext)
	// OnHookEnd is called when a hook finishes executing.
	OnHookEnd(name string, event HookEvent, ctx *HookContext, err error, duration time.Duration)
	// OnHookError is called when a hook encounters an error.
	OnHookError(name string, event HookEvent, ctx *HookContext, err error)
	// OnHookSkip is called when a hook is skipped.
	OnHookSkip(name string, event HookEvent, ctx *HookContext, reason string)
}

// HookManager defines the comprehensive interface for hook management.
type HookManager interface {
	HookRegistry
	HookExecutor
	// SetObserver sets the hook observer.
	SetObserver(observer HookObserver)
	// GetMetrics returns hook execution metrics.
	GetMetrics() HookMetrics
	// Shutdown gracefully shuts down the hook manager.
	Shutdown(ctx context.Context) error
}

// HookMetrics provides metrics about hook execution.
type HookMetrics struct {
	TotalExecutions      int64                    `json:"total_executions"`
	SuccessfulExecutions int64                    `json:"successful_executions"`
	FailedExecutions     int64                    `json:"failed_executions"`
	AverageLatency       time.Duration            `json:"average_latency"`
	MaxLatency           time.Duration            `json:"max_latency"`
	MinLatency           time.Duration            `json:"min_latency"`
	ExecutionsByEvent    map[HookEvent]int64      `json:"executions_by_event"`
	ExecutionsByHook     map[string]int64         `json:"executions_by_hook"`
	ErrorsByHook         map[string]int64         `json:"errors_by_hook"`
	LatencyByHook        map[string]time.Duration `json:"latency_by_hook"`
}
