// Package middlewares provides middleware implementations for HTTP server abstraction.
// It contains middleware examples that can be used with any HTTP server provider.
package middlewares

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// MiddlewareManager manages the lifecycle and execution of middlewares.
type MiddlewareManager struct {
	middlewares     []Middleware
	observerManager *ObserverManager
	logger          Logger
}

// Middleware defines the interface for HTTP middleware.
type Middleware interface {
	// Name returns the name of the middleware.
	Name() string

	// Process handles the middleware processing logic.
	// The next parameter is a function that represents the next middleware or handler in the chain.
	Process(ctx context.Context, req interface{}, next MiddlewareNext) (interface{}, error)

	// IsEnabled returns true if the middleware is enabled.
	IsEnabled() bool

	// SetEnabled sets the enabled state of the middleware.
	SetEnabled(enabled bool)

	// Priority returns the execution priority of the middleware.
	// Lower numbers execute first.
	Priority() int
}

// MiddlewareNext represents the next function in the middleware chain.
type MiddlewareNext func(ctx context.Context, req interface{}) (interface{}, error)

// Logger defines the interface for logging in middlewares.
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
}

// DefaultLogger provides a simple default logger implementation.
type DefaultLogger struct{}

// Info logs an info message.
func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	fmt.Printf("[INFO] "+msg+"\n", args...)
}

// Error logs an error message.
func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	fmt.Printf("[ERROR] "+msg+"\n", args...)
}

// Debug logs a debug message.
func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+msg+"\n", args...)
}

// Warn logs a warning message.
func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
	fmt.Printf("[WARN] "+msg+"\n", args...)
}

// ObserverManager manages middleware observers (simplified version for middlewares).
type ObserverManager struct {
	observers []interfaces.ServerObserver
}

// NewObserverManager creates a new observer manager for middlewares.
func NewObserverManager() *ObserverManager {
	return &ObserverManager{
		observers: make([]interfaces.ServerObserver, 0),
	}
}

// AttachObserver registers an observer for middleware events.
func (om *ObserverManager) AttachObserver(observer interfaces.ServerObserver) error {
	if observer == nil {
		return fmt.Errorf("observer cannot be nil")
	}

	om.observers = append(om.observers, observer)
	return nil
}

// NotifyObservers notifies all registered observers about middleware events.
func (om *ObserverManager) NotifyObservers(eventType interfaces.EventType, ctx context.Context, data interface{}) error {
	var lastError error
	for _, observer := range om.observers {
		switch eventType {
		case interfaces.EventRequest:
			if err := observer.OnRequest(ctx, data); err != nil {
				lastError = err
			}
		case interfaces.EventResponse:
			if respData, ok := data.(ResponseEventData); ok {
				if err := observer.OnResponse(ctx, respData.Request, respData.Response, respData.Duration); err != nil {
					lastError = err
				}
			}
		case interfaces.EventError:
			if err, ok := data.(error); ok {
				if err := observer.OnError(ctx, err); err != nil {
					lastError = err
				}
			}
		}
	}
	return lastError
}

// ResponseEventData contains data for response events.
type ResponseEventData struct {
	Request  interface{}
	Response interface{}
	Duration time.Duration
}

// NewMiddlewareManager creates a new middleware manager.
func NewMiddlewareManager() *MiddlewareManager {
	return &MiddlewareManager{
		middlewares:     make([]Middleware, 0),
		observerManager: NewObserverManager(),
		logger:          &DefaultLogger{},
	}
}

// AddMiddleware adds a middleware to the manager.
func (mm *MiddlewareManager) AddMiddleware(middleware Middleware) error {
	if middleware == nil {
		return fmt.Errorf("middleware cannot be nil")
	}

	mm.middlewares = append(mm.middlewares, middleware)
	mm.sortMiddlewares()
	return nil
}

// RemoveMiddleware removes a middleware by name.
func (mm *MiddlewareManager) RemoveMiddleware(name string) error {
	for i, middleware := range mm.middlewares {
		if middleware.Name() == name {
			// Remove middleware by slicing
			mm.middlewares = append(mm.middlewares[:i], mm.middlewares[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("middleware with name %s not found", name)
}

// GetMiddleware returns a middleware by name.
func (mm *MiddlewareManager) GetMiddleware(name string) (Middleware, error) {
	for _, middleware := range mm.middlewares {
		if middleware.Name() == name {
			return middleware, nil
		}
	}
	return nil, fmt.Errorf("middleware with name %s not found", name)
}

// ListMiddlewares returns a list of all middleware names.
func (mm *MiddlewareManager) ListMiddlewares() []string {
	names := make([]string, 0, len(mm.middlewares))
	for _, middleware := range mm.middlewares {
		names = append(names, middleware.Name())
	}
	return names
}

// ProcessRequest processes a request through all enabled middlewares.
func (mm *MiddlewareManager) ProcessRequest(ctx context.Context, req interface{}) (interface{}, error) {
	startTime := time.Now()

	// Notify observers about request start
	mm.observerManager.NotifyObservers(interfaces.EventRequest, ctx, req)

	// Create the middleware chain
	handler := mm.createMiddlewareChain(ctx)

	// Process the request
	resp, err := handler(ctx, req)

	// Calculate duration
	duration := time.Since(startTime)

	// Notify observers about response
	respData := ResponseEventData{
		Request:  req,
		Response: resp,
		Duration: duration,
	}
	mm.observerManager.NotifyObservers(interfaces.EventResponse, ctx, respData)

	// Notify observers about errors if any
	if err != nil {
		mm.observerManager.NotifyObservers(interfaces.EventError, ctx, err)
	}

	return resp, err
}

// createMiddlewareChain creates a chain of middleware functions.
func (mm *MiddlewareManager) createMiddlewareChain(ctx context.Context) MiddlewareNext {
	// Filter enabled middlewares
	enabledMiddlewares := make([]Middleware, 0, len(mm.middlewares))
	for _, middleware := range mm.middlewares {
		if middleware.IsEnabled() {
			enabledMiddlewares = append(enabledMiddlewares, middleware)
		}
	}

	// Create the final handler (identity function)
	finalHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return req, nil
	}

	// Build the chain from right to left
	handler := finalHandler
	for i := len(enabledMiddlewares) - 1; i >= 0; i-- {
		middleware := enabledMiddlewares[i]
		nextHandler := handler

		// Capture variables in closure
		currentMiddleware := middleware
		currentNext := nextHandler

		handler = func(ctx context.Context, req interface{}) (interface{}, error) {
			return currentMiddleware.Process(ctx, req, currentNext)
		}
	}

	return handler
}

// sortMiddlewares sorts middlewares by priority (lower numbers first).
func (mm *MiddlewareManager) sortMiddlewares() {
	// Simple bubble sort for small lists
	n := len(mm.middlewares)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if mm.middlewares[j].Priority() > mm.middlewares[j+1].Priority() {
				mm.middlewares[j], mm.middlewares[j+1] = mm.middlewares[j+1], mm.middlewares[j]
			}
		}
	}
}

// AttachObserver attaches an observer to the middleware manager.
func (mm *MiddlewareManager) AttachObserver(observer interfaces.ServerObserver) error {
	return mm.observerManager.AttachObserver(observer)
}

// SetLogger sets a custom logger for the middleware manager.
func (mm *MiddlewareManager) SetLogger(logger Logger) {
	if logger != nil {
		mm.logger = logger
	}
}

// GetMiddlewareCount returns the number of registered middlewares.
func (mm *MiddlewareManager) GetMiddlewareCount() int {
	return len(mm.middlewares)
}

// GetEnabledMiddlewareCount returns the number of enabled middlewares.
func (mm *MiddlewareManager) GetEnabledMiddlewareCount() int {
	count := 0
	for _, middleware := range mm.middlewares {
		if middleware.IsEnabled() {
			count++
		}
	}
	return count
}

// Clear removes all middlewares.
func (mm *MiddlewareManager) Clear() {
	mm.middlewares = make([]Middleware, 0)
}

// BaseMiddleware provides a base implementation for common middleware functionality.
type BaseMiddleware struct {
	name     string
	enabled  bool
	priority int
	logger   Logger
}

// NewBaseMiddleware creates a new base middleware.
func NewBaseMiddleware(name string, priority int) *BaseMiddleware {
	return &BaseMiddleware{
		name:     name,
		enabled:  true,
		priority: priority,
		logger:   &DefaultLogger{},
	}
}

// Name returns the middleware name.
func (bm *BaseMiddleware) Name() string {
	return bm.name
}

// IsEnabled returns true if the middleware is enabled.
func (bm *BaseMiddleware) IsEnabled() bool {
	return bm.enabled
}

// SetEnabled sets the enabled state of the middleware.
func (bm *BaseMiddleware) SetEnabled(enabled bool) {
	bm.enabled = enabled
}

// Priority returns the execution priority of the middleware.
func (bm *BaseMiddleware) Priority() int {
	return bm.priority
}

// SetPriority sets the execution priority of the middleware.
func (bm *BaseMiddleware) SetPriority(priority int) {
	bm.priority = priority
}

// SetLogger sets a custom logger for the middleware.
func (bm *BaseMiddleware) SetLogger(logger Logger) {
	if logger != nil {
		bm.logger = logger
	}
}

// GetLogger returns the middleware logger.
func (bm *BaseMiddleware) GetLogger() Logger {
	return bm.logger
}

// Process provides a default implementation that just calls the next handler.
func (bm *BaseMiddleware) Process(ctx context.Context, req interface{}, next MiddlewareNext) (interface{}, error) {
	bm.logger.Debug("BaseMiddleware %s: Processing request", bm.name)
	return next(ctx, req)
}
