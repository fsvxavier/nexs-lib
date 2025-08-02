// Package interfaces defines the core contracts for HTTP server abstraction.
// It provides interfaces that enable extensible HTTP server implementations
// with standardized lifecycle management, event handling, and middleware support.
package interfaces

import (
	"context"
	"time"
)

// HTTPServer defines the core interface that all HTTP server providers must implement.
// It provides standardized methods for server lifecycle management and configuration.
type HTTPServer interface {
	// Start initiates the HTTP server with the provided configuration.
	// Returns an error if the server fails to start.
	Start(ctx context.Context) error

	// Stop gracefully shuts down the HTTP server.
	// It waits for ongoing requests to complete within the timeout.
	Stop(ctx context.Context) error

	// RegisterRoute adds a new route to the server with the specified method, path, and handler.
	// Supports standard HTTP methods (GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD).
	// The handler parameter is interface{} to support different provider-specific handler types.
	RegisterRoute(method, path string, handler interface{}) error

	// RegisterMiddleware adds middleware to the server's middleware chain.
	// Middleware is executed in the order it was registered.
	// The middleware parameter is interface{} to support different provider-specific middleware types.
	RegisterMiddleware(middleware interface{}) error

	// AttachObserver registers an observer for server lifecycle events.
	// Multiple observers can be attached to receive event notifications.
	AttachObserver(observer ServerObserver) error

	// DetachObserver removes an observer from the server's observer list.
	DetachObserver(observer ServerObserver) error

	// GetAddr returns the address the server is listening on.
	// Returns empty string if server is not started.
	GetAddr() string

	// IsRunning returns true if the server is currently running.
	IsRunning() bool

	// GetStats returns runtime statistics about the server.
	GetStats() ServerStats
}

// ServerObserver defines the interface for observing HTTP server lifecycle events.
// Implementations can react to server state changes and request processing events.
type ServerObserver interface {
	// OnStart is called when the server starts successfully.
	OnStart(ctx context.Context, addr string) error

	// OnStop is called when the server stops.
	OnStop(ctx context.Context) error

	// OnError is called when an error occurs during server operations.
	OnError(ctx context.Context, err error) error

	// OnRequest is called before processing each request.
	// The req parameter is interface{} to support different provider-specific request types.
	OnRequest(ctx context.Context, req interface{}) error

	// OnResponse is called after processing each request.
	// The req and resp parameters are interface{} to support different provider-specific types.
	OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error

	// OnRouteEnter is called when entering a route handler.
	// The req parameter is interface{} to support different provider-specific request types.
	OnRouteEnter(ctx context.Context, method, path string, req interface{}) error

	// OnRouteExit is called when exiting a route handler.
	// The req parameter is interface{} to support different provider-specific request types.
	OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error
}

// HookFunc defines the signature for hook functions that can be executed
// at specific points in the server lifecycle.
type HookFunc func(ctx context.Context, data interface{}) error

// ServerStats contains runtime statistics about the HTTP server.
type ServerStats struct {
	// StartTime is when the server was started.
	StartTime time.Time

	// RequestCount is the total number of requests processed.
	RequestCount int64

	// ErrorCount is the total number of errors encountered.
	ErrorCount int64

	// AverageResponseTime is the average response time for requests.
	AverageResponseTime time.Duration

	// ActiveConnections is the current number of active connections.
	ActiveConnections int64

	// Provider is the name of the HTTP server provider (gin, echo, fiber, etc.).
	Provider string
}

// EventType represents the type of server lifecycle event.
type EventType string

const (
	// EventStart represents server start event.
	EventStart EventType = "start"

	// EventStop represents server stop event.
	EventStop EventType = "stop"

	// EventError represents server error event.
	EventError EventType = "error"

	// EventRequest represents request processing event.
	EventRequest EventType = "request"

	// EventResponse represents response processing event.
	EventResponse EventType = "response"

	// EventRouteEnter represents route entry event.
	EventRouteEnter EventType = "route_enter"

	// EventRouteExit represents route exit event.
	EventRouteExit EventType = "route_exit"
)

// Event represents a server lifecycle event with associated data.
type Event struct {
	// Type is the type of event.
	Type EventType

	// Timestamp is when the event occurred.
	Timestamp time.Time

	// Data contains event-specific data.
	Data interface{}

	// Context is the request context if applicable.
	Context context.Context
}

// ProviderFactory defines the interface for creating HTTP server instances.
// Each provider (gin, echo, fiber, etc.) must implement this factory.
type ProviderFactory interface {
	// Create creates a new HTTP server instance with the provided configuration.
	Create(config interface{}) (HTTPServer, error)

	// GetName returns the name of the provider (e.g., "gin", "echo", "fiber").
	GetName() string

	// GetDefaultConfig returns the default configuration for this provider.
	GetDefaultConfig() interface{}

	// ValidateConfig validates the provided configuration.
	ValidateConfig(config interface{}) error
}

// Registry defines the interface for managing HTTP server providers.
// It acts as a central repository for all available server implementations.
type Registry interface {
	// Register adds a new provider factory to the registry.
	Register(name string, factory ProviderFactory) error

	// Unregister removes a provider factory from the registry.
	Unregister(name string) error

	// Create creates a new HTTP server instance using the specified provider.
	Create(providerName string, config interface{}) (HTTPServer, error)

	// List returns a list of all registered provider names.
	List() []string

	// IsRegistered checks if a provider is registered.
	IsRegistered(name string) bool

	// GetProvider returns the factory for the specified provider.
	GetProvider(name string) (ProviderFactory, error)
}

// Config defines the base configuration interface for HTTP servers.
// Provider-specific configurations should embed this interface.
type Config interface {
	// GetAddr returns the address to bind the server to.
	GetAddr() string

	// GetPort returns the port to bind the server to.
	GetPort() int

	// GetReadTimeout returns the read timeout for requests.
	GetReadTimeout() time.Duration

	// GetWriteTimeout returns the write timeout for responses.
	GetWriteTimeout() time.Duration

	// GetIdleTimeout returns the idle timeout for connections.
	GetIdleTimeout() time.Duration

	// IsGracefulShutdown returns true if graceful shutdown is enabled.
	IsGracefulShutdown() bool

	// GetShutdownTimeout returns the timeout for graceful shutdown.
	GetShutdownTimeout() time.Duration
}
