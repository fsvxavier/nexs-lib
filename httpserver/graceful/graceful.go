// Package graceful provides utilities for graceful shutdown, restart, and rolling updates.
package graceful

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// Manager handles graceful operations for HTTP servers.
type Manager struct {
	mu                sync.RWMutex
	servers           map[string]interfaces.HTTPServer
	connections       int64
	drainTimeout      time.Duration
	shutdownTimeout   time.Duration
	preShutdownHooks  []func() error
	postShutdownHooks []func() error
	healthChecks      map[string]func() interfaces.HealthCheck
	version           string
	startTime         time.Time
	shutdownChannel   chan os.Signal
	isShuttingDown    int32
}

// NewManager creates a new graceful operations manager.
func NewManager() *Manager {
	return &Manager{
		servers:         make(map[string]interfaces.HTTPServer),
		drainTimeout:    30 * time.Second,
		shutdownTimeout: 60 * time.Second,
		healthChecks:    make(map[string]func() interfaces.HealthCheck),
		version:         "1.0.0",
		startTime:       time.Now(),
		shutdownChannel: make(chan os.Signal, 1),
	}
}

// RegisterServer registers a server for graceful management.
func (m *Manager) RegisterServer(name string, server interfaces.HTTPServer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.servers[name] = server
}

// UnregisterServer removes a server from graceful management.
func (m *Manager) UnregisterServer(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.servers, name)
}

// SetDrainTimeout sets the timeout for connection draining.
func (m *Manager) SetDrainTimeout(timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.drainTimeout = timeout
}

// SetShutdownTimeout sets the timeout for shutdown operations.
func (m *Manager) SetShutdownTimeout(timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shutdownTimeout = timeout
}

// AddPreShutdownHook adds a hook to be executed before shutdown.
func (m *Manager) AddPreShutdownHook(hook func() error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.preShutdownHooks = append(m.preShutdownHooks, hook)
}

// AddPostShutdownHook adds a hook to be executed after shutdown.
func (m *Manager) AddPostShutdownHook(hook func() error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.postShutdownHooks = append(m.postShutdownHooks, hook)
}

// AddHealthCheck adds a health check function.
func (m *Manager) AddHealthCheck(name string, check func() interfaces.HealthCheck) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.healthChecks[name] = check
}

// GetHealthStatus returns the current health status.
func (m *Manager) GetHealthStatus() interfaces.HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := "healthy"
	if atomic.LoadInt32(&m.isShuttingDown) == 1 {
		status = "shutting_down"
	}

	checks := make(map[string]interfaces.HealthCheck)
	hasWarning := false
	hasError := false

	for name, checkFunc := range m.healthChecks {
		check := checkFunc()
		checks[name] = check

		switch check.Status {
		case "warning":
			hasWarning = true
		case "error", "unhealthy", "critical":
			hasError = true
		}
	}

	// Determine overall status based on individual checks
	if hasError {
		status = "unhealthy"
	} else if hasWarning {
		status = "warning"
	}

	return interfaces.HealthStatus{
		Status:      status,
		Version:     m.version,
		Timestamp:   time.Now(),
		Uptime:      time.Since(m.startTime),
		Connections: atomic.LoadInt64(&m.connections),
		Checks:      checks,
	}
}

// GracefulShutdown performs a graceful shutdown of all registered servers.
func (m *Manager) GracefulShutdown(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&m.isShuttingDown, 0, 1) {
		return fmt.Errorf("shutdown already in progress")
	}

	log.Println("Starting graceful shutdown...")

	// Execute pre-shutdown hooks
	if err := m.executePreShutdownHooks(); err != nil {
		log.Printf("Pre-shutdown hook failed: %v", err)
	}

	// Create timeout context for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, m.shutdownTimeout)
	defer cancel()

	// Stop accepting new connections
	log.Println("Stopping acceptance of new connections...")

	var wg sync.WaitGroup
	errors := make(chan error, len(m.servers))

	m.mu.RLock()
	for name, server := range m.servers {
		wg.Add(1)
		go func(serverName string, srv interfaces.HTTPServer) {
			defer wg.Done()

			log.Printf("Shutting down server: %s", serverName)

			// Try graceful stop first if server supports it
			if gracefulSrv, ok := srv.(interfaces.GracefulServer); ok {
				if err := gracefulSrv.GracefulStop(shutdownCtx, m.drainTimeout); err != nil {
					log.Printf("Graceful stop failed for %s: %v", serverName, err)
					// Fallback to regular stop
					if err := srv.Stop(shutdownCtx); err != nil {
						errors <- fmt.Errorf("failed to stop server %s: %w", serverName, err)
						return
					}
				}
			} else {
				if err := srv.Stop(shutdownCtx); err != nil {
					errors <- fmt.Errorf("failed to stop server %s: %w", serverName, err)
					return
				}
			}

			log.Printf("Server %s stopped successfully", serverName)
		}(name, server)
	}
	m.mu.RUnlock()

	// Wait for all servers to stop
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All servers stopped successfully")
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout reached, forcing stop")
		return fmt.Errorf("shutdown timeout reached")
	case err := <-errors:
		log.Printf("Shutdown error: %v", err)
		return err
	}

	// Wait for connections to drain
	log.Println("Waiting for connections to drain...")
	drainCtx, drainCancel := context.WithTimeout(context.Background(), m.drainTimeout)
	defer drainCancel()

	if err := m.waitForConnections(drainCtx); err != nil {
		log.Printf("Connection drain timeout: %v", err)
	}

	// Execute post-shutdown hooks
	if err := m.executePostShutdownHooks(); err != nil {
		log.Printf("Post-shutdown hook failed: %v", err)
	}

	log.Println("Graceful shutdown completed")
	return nil
}

// Restart performs a zero-downtime restart.
func (m *Manager) Restart(ctx context.Context) error {
	log.Println("Starting zero-downtime restart...")

	// For zero-downtime restart, we need to:
	// 1. Start new server instances on different ports
	// 2. Update load balancer/proxy to point to new instances
	// 3. Drain connections from old instances
	// 4. Stop old instances

	return fmt.Errorf("restart functionality requires external orchestration (load balancer/proxy)")
}

// WaitForShutdownSignal waits for shutdown signals and triggers graceful shutdown.
func (m *Manager) WaitForShutdownSignal() {
	signal.Notify(m.shutdownChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-m.shutdownChannel
	log.Printf("Received signal: %v", sig)

	ctx, cancel := context.WithTimeout(context.Background(), m.shutdownTimeout)
	defer cancel()

	if err := m.GracefulShutdown(ctx); err != nil {
		log.Printf("Graceful shutdown failed: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}

// IncrementConnections increments the active connections counter.
func (m *Manager) IncrementConnections() {
	atomic.AddInt64(&m.connections, 1)
}

// DecrementConnections decrements the active connections counter.
func (m *Manager) DecrementConnections() {
	atomic.AddInt64(&m.connections, -1)
}

// GetConnectionsCount returns the current number of active connections.
func (m *Manager) GetConnectionsCount() int64 {
	return atomic.LoadInt64(&m.connections)
}

// executePreShutdownHooks executes all pre-shutdown hooks.
func (m *Manager) executePreShutdownHooks() error {
	m.mu.RLock()
	hooks := make([]func() error, len(m.preShutdownHooks))
	copy(hooks, m.preShutdownHooks)
	m.mu.RUnlock()

	for i, hook := range hooks {
		if err := hook(); err != nil {
			return fmt.Errorf("pre-shutdown hook %d failed: %w", i, err)
		}
	}
	return nil
}

// executePostShutdownHooks executes all post-shutdown hooks.
func (m *Manager) executePostShutdownHooks() error {
	m.mu.RLock()
	hooks := make([]func() error, len(m.postShutdownHooks))
	copy(hooks, m.postShutdownHooks)
	m.mu.RUnlock()

	for i, hook := range hooks {
		if err := hook(); err != nil {
			return fmt.Errorf("post-shutdown hook %d failed: %w", i, err)
		}
	}
	return nil
}

// waitForConnections waits for all connections to finish or timeout.
func (m *Manager) waitForConnections(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			remaining := atomic.LoadInt64(&m.connections)
			if remaining > 0 {
				return fmt.Errorf("timeout waiting for %d connections to drain", remaining)
			}
			return nil
		case <-ticker.C:
			if atomic.LoadInt64(&m.connections) == 0 {
				return nil
			}
		}
	}
}

// DefaultManager is the global graceful manager instance.
var DefaultManager = NewManager()

// RegisterServer registers a server with the default manager.
func RegisterServer(name string, server interfaces.HTTPServer) {
	DefaultManager.RegisterServer(name, server)
}

// GracefulShutdown performs graceful shutdown using the default manager.
func GracefulShutdown(ctx context.Context) error {
	return DefaultManager.GracefulShutdown(ctx)
}

// WaitForShutdownSignal waits for shutdown signals using the default manager.
func WaitForShutdownSignal() {
	DefaultManager.WaitForShutdownSignal()
}
