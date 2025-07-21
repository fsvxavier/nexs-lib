// Package bulkhead provides bulkhead pattern middleware implementation for resource isolation.
package bulkhead

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// Context key for storing the bulkhead resource key
const bulkheadResourceKey = "bulkhead:resource"

// Config represents bulkhead configuration.
type Config struct {
	// Enabled indicates if the middleware is enabled.
	Enabled bool
	// SkipPaths contains paths to skip for this middleware.
	SkipPaths []string
	// Headers contains custom headers to add.
	Headers map[string]string
	// MaxConcurrent is the maximum number of concurrent requests.
	MaxConcurrent int
	// QueueSize is the size of the waiting queue.
	QueueSize int
	// Timeout is the maximum wait time in queue.
	Timeout time.Duration
	// ResourceKey generates the resource key for bulkhead isolation.
	ResourceKey func(*http.Request) string
	// OnRejected is called when a request is rejected.
	OnRejected func(http.ResponseWriter, *http.Request, string)
}

// IsEnabled returns whether the middleware is enabled.
func (c *Config) IsEnabled() bool {
	return c.Enabled
}

// ShouldSkip checks if a path should be skipped.
func (c *Config) ShouldSkip(path string) bool {
	for _, skipPath := range c.SkipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// DefaultConfig returns a default bulkhead configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:       true,
		MaxConcurrent: 100,
		QueueSize:     50,
		Timeout:       10 * time.Second,
		ResourceKey:   defaultResourceKey,
		OnRejected:    defaultOnRejected,
	}
}

// Middleware implements bulkhead middleware.
type Middleware struct {
	config    Config
	resources map[string]*resource
	mu        sync.RWMutex
}

// resource represents a bulkhead resource with its own limits.
type resource struct {
	semaphore chan struct{}
	queue     chan *request
	active    int64
	mu        sync.RWMutex
}

// request represents a queued request.
type request struct {
	w      http.ResponseWriter
	r      *http.Request
	next   http.Handler
	done   chan struct{}
	result chan error
}

// NewMiddleware creates a new bulkhead middleware.
func NewMiddleware(config Config) *Middleware {
	if config.ResourceKey == nil {
		config.ResourceKey = defaultResourceKey
	}
	if config.OnRejected == nil {
		config.OnRejected = defaultOnRejected
	}

	return &Middleware{
		config:    config,
		resources: make(map[string]*resource),
	}
}

// Wrap implements the interfaces.Middleware interface.
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.config.IsEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		if m.config.ShouldSkip(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		resourceKey := m.config.ResourceKey(r)

		// Add resource key to context
		ctx := context.WithValue(r.Context(), bulkheadResourceKey, resourceKey)
		r = r.WithContext(ctx)

		res := m.getOrCreateResource(resourceKey)

		// Try to acquire semaphore immediately
		select {
		case res.semaphore <- struct{}{}:
			// Got permission immediately
			m.executeRequest(res, w, r, next)
			return
		default:
			// No immediate permission, try to queue
		}

		// Try to queue the request
		req := &request{
			w:      w,
			r:      r,
			next:   next,
			done:   make(chan struct{}),
			result: make(chan error, 1),
		}

		select {
		case res.queue <- req:
			// Successfully queued, wait for execution or timeout
			select {
			case <-req.done:
				// Request was processed
				return
			case <-time.After(m.config.Timeout):
				// Request timed out in queue
				m.config.OnRejected(w, r, "queue timeout")
				return
			case <-r.Context().Done():
				// Request context cancelled
				return
			}
		default:
			// Queue is full, reject immediately
			m.config.OnRejected(w, r, "queue full")
			return
		}
	})
}

// Name returns the middleware name.
func (m *Middleware) Name() string {
	return "bulkhead"
}

// Priority returns the middleware priority.
func (m *Middleware) Priority() int {
	return 300 // Bulkhead should happen after initial processing
}

// getOrCreateResource gets or creates a resource for the given key.
func (m *Middleware) getOrCreateResource(key string) *resource {
	m.mu.RLock()
	res, exists := m.resources[key]
	m.mu.RUnlock()

	if exists {
		return res
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if res, exists := m.resources[key]; exists {
		return res
	}

	// Create new resource
	res = &resource{
		semaphore: make(chan struct{}, m.config.MaxConcurrent),
		queue:     make(chan *request, m.config.QueueSize),
	}

	// Start queue processor
	go m.processQueue(res)

	m.resources[key] = res
	return res
}

// processQueue processes queued requests for a resource.
func (m *Middleware) processQueue(res *resource) {
	for req := range res.queue {
		// Wait for semaphore
		select {
		case res.semaphore <- struct{}{}:
			// Got permission, execute request
			go func(r *request) {
				m.executeRequest(res, r.w, r.r, r.next)
				close(r.done)
			}(req)
		case <-req.r.Context().Done():
			// Request cancelled while waiting
			close(req.done)
		}
	}
}

// executeRequest executes a request and releases the semaphore.
func (m *Middleware) executeRequest(res *resource, w http.ResponseWriter, r *http.Request, next http.Handler) {
	defer func() {
		<-res.semaphore // Release semaphore
	}()

	// Increment active counter
	res.mu.Lock()
	res.active++
	res.mu.Unlock()

	defer func() {
		// Decrement active counter
		res.mu.Lock()
		res.active--
		res.mu.Unlock()
	}()

	next.ServeHTTP(w, r)
}

// GetResourceStats returns statistics for a resource.
func (m *Middleware) GetResourceStats(key string) (active int64, queued int, available int) {
	m.mu.RLock()
	res, exists := m.resources[key]
	m.mu.RUnlock()

	if !exists {
		return 0, 0, m.config.MaxConcurrent
	}

	res.mu.RLock()
	active = res.active
	res.mu.RUnlock()

	queued = len(res.queue)
	available = m.config.MaxConcurrent - len(res.semaphore)

	return active, queued, available
}

// defaultResourceKey generates a default resource key (all requests use same resource).
func defaultResourceKey(r *http.Request) string {
	return "default"
}

// defaultOnRejected sends a service unavailable response.
func defaultOnRejected(w http.ResponseWriter, r *http.Request, reason string) {
	http.Error(w, "Service temporarily unavailable: "+reason, http.StatusServiceUnavailable)
}
