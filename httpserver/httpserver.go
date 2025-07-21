// Package httpserver provides a Registry + Factory + Orchestrator for HTTP servers.
package httpserver

import (
	"context"
	"fmt"
	"sync"

	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// Registry manages HTTP server factories and observers.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]interfaces.ServerFactory
	observers []interfaces.ServerObserver
}

// NewRegistry creates a new server registry.
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]interfaces.ServerFactory),
		observers: make([]interfaces.ServerObserver, 0),
	}
}

// Register registers a new server factory with the given name.
func (r *Registry) Register(name string, factory interfaces.ServerFactory) error {
	if name == "" {
		return fmt.Errorf("server name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("server factory with name '%s' already registered", name)
	}

	r.factories[name] = factory
	return nil
}

// Unregister removes a server factory.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; !exists {
		return fmt.Errorf("server factory with name '%s' not found", name)
	}

	delete(r.factories, name)
	return nil
}

// Create creates a new HTTP server instance using the registered factory.
func (r *Registry) Create(name string, cfg *config.Config) (interfaces.HTTPServer, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	r.mu.RLock()
	factory, exists := r.factories[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("server factory with name '%s' not found", name)
	}

	server, err := factory(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create server '%s': %w", name, err)
	}

	// Wrap the server to add observer notifications
	return &observableServer{
		HTTPServer: server,
		name:       name,
		registry:   r,
	}, nil
}

// AttachObserver attaches an observer to receive server lifecycle events.
func (r *Registry) AttachObserver(observer interfaces.ServerObserver) {
	if observer == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.observers = append(r.observers, observer)
}

// DetachObserver removes an observer from the registry.
func (r *Registry) DetachObserver(observer interfaces.ServerObserver) {
	if observer == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for i, obs := range r.observers {
		if obs == observer {
			r.observers = append(r.observers[:i], r.observers[i+1:]...)
			break
		}
	}
}

// ListFactories returns a list of registered factory names.
func (r *Registry) ListFactories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// GetObservers returns a copy of the observers list.
func (r *Registry) GetObservers() []interfaces.ServerObserver {
	r.mu.RLock()
	defer r.mu.RUnlock()

	observers := make([]interfaces.ServerObserver, len(r.observers))
	copy(observers, r.observers)
	return observers
}

// notifyStart notifies all observers when a server starts.
func (r *Registry) notifyStart(name string, addr string) {
	r.mu.RLock()
	observers := make([]interfaces.ServerObserver, len(r.observers))
	copy(observers, r.observers)
	r.mu.RUnlock()

	for _, observer := range observers {
		observer.OnStart(name)
	}
}

// notifyStop notifies all observers when a server stops.
func (r *Registry) notifyStop(name string, addr string) {
	r.mu.RLock()
	observers := make([]interfaces.ServerObserver, len(r.observers))
	copy(observers, r.observers)
	r.mu.RUnlock()

	for _, observer := range observers {
		observer.OnStop(name)
	}
}

// observableServer wraps an HTTPServer to add observer notifications.
type observableServer struct {
	interfaces.HTTPServer
	name     string
	registry *Registry
}

// Start wraps the underlying server's Start method with observer notifications.
func (o *observableServer) Start() error {
	err := o.HTTPServer.Start()
	if err != nil {
		return err
	}
	o.registry.notifyStart(o.name, o.HTTPServer.GetAddr())
	return nil
}

// Stop wraps the underlying server's Stop method with observer notifications.
func (o *observableServer) Stop(ctx context.Context) error {
	err := o.HTTPServer.Stop(ctx)
	o.registry.notifyStop(o.name, o.HTTPServer.GetAddr())
	return err
}

// Global registry instance
var defaultRegistry = NewRegistry()

// Register registers a server factory with the default registry.
func Register(name string, factory interfaces.ServerFactory) error {
	return defaultRegistry.Register(name, factory)
}

// Unregister removes a server factory from the default registry.
func Unregister(name string) error {
	return defaultRegistry.Unregister(name)
}

// Create creates a new HTTP server using the default registry.
func Create(name string, cfg *config.Config) (interfaces.HTTPServer, error) {
	return defaultRegistry.Create(name, cfg)
}

// AttachObserver attaches an observer to the default registry.
func AttachObserver(observer interfaces.ServerObserver) {
	defaultRegistry.AttachObserver(observer)
}

// DetachObserver removes an observer from the default registry.
func DetachObserver(observer interfaces.ServerObserver) {
	defaultRegistry.DetachObserver(observer)
}

// ListFactories returns registered factory names from the default registry.
func ListFactories() []string {
	return defaultRegistry.ListFactories()
}

// GetDefaultRegistry returns the default registry instance.
func GetDefaultRegistry() *Registry {
	return defaultRegistry
}
