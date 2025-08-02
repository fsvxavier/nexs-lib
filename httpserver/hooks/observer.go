// Package hooks provides the Observer pattern implementation for HTTP server lifecycle events.
// It enables extensible event handling through hooks that can be registered for specific events.
package hooks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// EventData contains the data associated with a server event.
type EventData struct {
	EventType interfaces.EventType
	Timestamp time.Time
	Data      interface{}
	Context   context.Context
}

// ObserverManager manages the registration and notification of observers for server events.
// It implements the Observer pattern to allow multiple components to react to server lifecycle events.
type ObserverManager struct {
	observers []interfaces.ServerObserver
	hooks     map[interfaces.EventType][]interfaces.HookFunc
	mutex     sync.RWMutex
}

// NewObserverManager creates a new observer manager.
func NewObserverManager() *ObserverManager {
	return &ObserverManager{
		observers: make([]interfaces.ServerObserver, 0),
		hooks:     make(map[interfaces.EventType][]interfaces.HookFunc),
		mutex:     sync.RWMutex{},
	}
}

// AttachObserver registers an observer for server lifecycle events.
func (om *ObserverManager) AttachObserver(observer interfaces.ServerObserver) error {
	if observer == nil {
		return fmt.Errorf("observer cannot be nil")
	}

	om.mutex.Lock()
	defer om.mutex.Unlock()

	// Check if observer is already attached
	for _, existing := range om.observers {
		if existing == observer {
			return fmt.Errorf("observer already attached")
		}
	}

	om.observers = append(om.observers, observer)
	return nil
}

// DetachObserver removes an observer from the server's observer list.
func (om *ObserverManager) DetachObserver(observer interfaces.ServerObserver) error {
	if observer == nil {
		return fmt.Errorf("observer cannot be nil")
	}

	om.mutex.Lock()
	defer om.mutex.Unlock()

	for i, existing := range om.observers {
		if existing == observer {
			// Remove observer by slicing
			om.observers = append(om.observers[:i], om.observers[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("observer not found")
}

// AttachHook registers a hook function for a specific event type.
func (om *ObserverManager) AttachHook(eventType interfaces.EventType, hook interfaces.HookFunc) error {
	if hook == nil {
		return fmt.Errorf("hook cannot be nil")
	}

	om.mutex.Lock()
	defer om.mutex.Unlock()

	if om.hooks[eventType] == nil {
		om.hooks[eventType] = make([]interfaces.HookFunc, 0)
	}

	om.hooks[eventType] = append(om.hooks[eventType], hook)
	return nil
}

// DetachHook removes all hooks for a specific event type.
func (om *ObserverManager) DetachHook(eventType interfaces.EventType) error {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	delete(om.hooks, eventType)
	return nil
}

// NotifyObservers notifies all registered observers about an event.
func (om *ObserverManager) NotifyObservers(eventType interfaces.EventType, ctx context.Context, data interface{}) error {
	om.mutex.RLock()
	observers := make([]interfaces.ServerObserver, len(om.observers))
	copy(observers, om.observers)
	hooks := make([]interfaces.HookFunc, 0)
	if om.hooks[eventType] != nil {
		hooks = make([]interfaces.HookFunc, len(om.hooks[eventType]))
		copy(hooks, om.hooks[eventType])
	}
	om.mutex.RUnlock()

	// Execute hooks first
	for _, hook := range hooks {
		if err := hook(ctx, data); err != nil {
			return fmt.Errorf("hook execution failed: %w", err)
		}
	}

	// Notify observers
	var lastError error
	for _, observer := range observers {
		if err := om.notifyObserver(observer, eventType, ctx, data); err != nil {
			lastError = err
		}
	}

	return lastError
}

// notifyObserver notifies a specific observer about an event.
func (om *ObserverManager) notifyObserver(observer interfaces.ServerObserver, eventType interfaces.EventType, ctx context.Context, data interface{}) error {
	switch eventType {
	case interfaces.EventStart:
		if addr, ok := data.(string); ok {
			return observer.OnStart(ctx, addr)
		}
		return fmt.Errorf("invalid data type for start event: expected string, got %T", data)

	case interfaces.EventStop:
		return observer.OnStop(ctx)

	case interfaces.EventError:
		if err, ok := data.(error); ok {
			return observer.OnError(ctx, err)
		}
		return fmt.Errorf("invalid data type for error event: expected error, got %T", data)

	case interfaces.EventRequest:
		return observer.OnRequest(ctx, data)

	case interfaces.EventResponse:
		if respData, ok := data.(ResponseEventData); ok {
			return observer.OnResponse(ctx, respData.Request, respData.Response, respData.Duration)
		}
		return fmt.Errorf("invalid data type for response event: expected ResponseEventData, got %T", data)

	case interfaces.EventRouteEnter:
		if routeData, ok := data.(RouteEventData); ok {
			return observer.OnRouteEnter(ctx, routeData.Method, routeData.Path, routeData.Request)
		}
		return fmt.Errorf("invalid data type for route enter event: expected RouteEventData, got %T", data)

	case interfaces.EventRouteExit:
		if routeData, ok := data.(RouteExitEventData); ok {
			return observer.OnRouteExit(ctx, routeData.Method, routeData.Path, routeData.Request, routeData.Duration)
		}
		return fmt.Errorf("invalid data type for route exit event: expected RouteExitEventData, got %T", data)

	default:
		return fmt.Errorf("unknown event type: %s", eventType)
	}
}

// GetObserverCount returns the number of registered observers.
func (om *ObserverManager) GetObserverCount() int {
	om.mutex.RLock()
	defer om.mutex.RUnlock()
	return len(om.observers)
}

// GetHookCount returns the number of registered hooks for a specific event type.
func (om *ObserverManager) GetHookCount(eventType interfaces.EventType) int {
	om.mutex.RLock()
	defer om.mutex.RUnlock()
	return len(om.hooks[eventType])
}

// Clear removes all observers and hooks.
func (om *ObserverManager) Clear() {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	om.observers = make([]interfaces.ServerObserver, 0)
	om.hooks = make(map[interfaces.EventType][]interfaces.HookFunc)
}

// ResponseEventData contains data for response events.
type ResponseEventData struct {
	Request  interface{}
	Response interface{}
	Duration time.Duration
}

// RouteEventData contains data for route enter events.
type RouteEventData struct {
	Method  string
	Path    string
	Request interface{}
}

// RouteExitEventData contains data for route exit events.
type RouteExitEventData struct {
	Method   string
	Path     string
	Request  interface{}
	Duration time.Duration
}
