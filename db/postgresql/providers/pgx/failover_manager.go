package pgx

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

// FailoverManagerImpl implements the FailoverManager interface
type FailoverManagerImpl struct {
	config       interfaces.FailoverConfig
	activeNode   string
	healthyNodes map[string]bool
	nodeLastSeen map[string]time.Time
	stats        *FailoverStatsImpl
	mu           sync.RWMutex
	healthTicker *time.Ticker
	stopHealth   chan struct{}
	stopped      bool
}

// FailoverStatsImpl implements failover statistics tracking
type FailoverStatsImpl struct {
	totalFailovers      int64
	successfulFailovers int64
	failedFailovers     int64
	lastFailoverTime    time.Time
	mu                  sync.RWMutex
}

// NewFailoverManager creates a new failover manager
func NewFailoverManager(config interfaces.FailoverConfig) interfaces.FailoverManager {
	fm := &FailoverManagerImpl{
		config:       config,
		healthyNodes: make(map[string]bool),
		nodeLastSeen: make(map[string]time.Time),
		stats:        &FailoverStatsImpl{},
		stopHealth:   make(chan struct{}),
	}

	// Initialize all nodes as healthy
	for _, node := range config.FallbackNodes {
		fm.healthyNodes[node] = true
		fm.nodeLastSeen[node] = time.Now()
	}

	// Set the first healthy node as active
	if len(config.FallbackNodes) > 0 {
		fm.activeNode = config.FallbackNodes[0]
	}

	// Start health check routine if enabled
	if config.Enabled && config.HealthCheckInterval > 0 {
		fm.startHealthCheck()
	}

	return fm
}

// Execute performs failover logic for operations
func (fm *FailoverManagerImpl) Execute(ctx context.Context, operation func(conn interfaces.IConn) error) error {
	if !fm.config.Enabled {
		return fmt.Errorf("failover is not enabled")
	}

	maxAttempts := fm.config.MaxFailoverAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Get current active node
		node := fm.getActiveNode()
		if node == "" {
			return fmt.Errorf("no healthy nodes available")
		}

		// Create connection to the active node
		conn, err := fm.createConnection(ctx, node)
		if err != nil {
			fm.markNodeDown(node)
			lastErr = err

			// Try to failover to another node
			if !fm.performFailover() {
				// No more nodes available
				break
			}
			continue
		}

		// Execute operation
		err = operation(conn)
		conn.Release() // Always release connection

		if err == nil {
			atomic.AddInt64(&fm.stats.successfulFailovers, 1)
			return nil
		}

		// Check if error indicates node failure
		if fm.isNodeFailureError(err) {
			fm.markNodeDown(node)
			lastErr = err

			// Try to failover to another node
			if !fm.performFailover() {
				// No more nodes available
				break
			}
		} else {
			// Operation error, not a node failure
			return err
		}
	}

	atomic.AddInt64(&fm.stats.failedFailovers, 1)
	return fmt.Errorf("failover failed after %d attempts: %w", maxAttempts, lastErr)
}

// MarkNodeDown marks a node as unhealthy
func (fm *FailoverManagerImpl) MarkNodeDown(nodeID string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if _, exists := fm.healthyNodes[nodeID]; !exists {
		return fmt.Errorf("node %s not found in configuration", nodeID)
	}

	fm.healthyNodes[nodeID] = false

	// If this was the active node, trigger failover
	if fm.activeNode == nodeID {
		fm.performFailoverUnsafe()
	}

	return nil
}

// MarkNodeUp marks a node as healthy
func (fm *FailoverManagerImpl) MarkNodeUp(nodeID string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if _, exists := fm.healthyNodes[nodeID]; !exists {
		return fmt.Errorf("node %s not found in configuration", nodeID)
	}

	fm.healthyNodes[nodeID] = true
	fm.nodeLastSeen[nodeID] = time.Now()

	// If we don't have an active node, make this the active node
	if fm.activeNode == "" || !fm.healthyNodes[fm.activeNode] {
		fm.activeNode = nodeID
	}

	return nil
}

// GetHealthyNodes returns a list of healthy nodes
func (fm *FailoverManagerImpl) GetHealthyNodes() []string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	var healthy []string
	for node, isHealthy := range fm.healthyNodes {
		if isHealthy {
			healthy = append(healthy, node)
		}
	}
	return healthy
}

// GetUnhealthyNodes returns a list of unhealthy nodes
func (fm *FailoverManagerImpl) GetUnhealthyNodes() []string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	var unhealthy []string
	for node, isHealthy := range fm.healthyNodes {
		if !isHealthy {
			unhealthy = append(unhealthy, node)
		}
	}
	return unhealthy
}

// GetStats returns failover statistics
func (fm *FailoverManagerImpl) GetStats() interfaces.FailoverStats {
	fm.mu.RLock()
	activeNode := fm.activeNode
	downNodes := fm.GetUnhealthyNodes()
	fm.mu.RUnlock()

	fm.stats.mu.RLock()
	lastFailoverTime := fm.stats.lastFailoverTime
	fm.stats.mu.RUnlock()

	return interfaces.FailoverStats{
		TotalFailovers:      atomic.LoadInt64(&fm.stats.totalFailovers),
		SuccessfulFailovers: atomic.LoadInt64(&fm.stats.successfulFailovers),
		FailedFailovers:     atomic.LoadInt64(&fm.stats.failedFailovers),
		CurrentActiveNode:   activeNode,
		DownNodes:           downNodes,
		LastFailoverTime:    lastFailoverTime,
	}
}

// Internal methods

func (fm *FailoverManagerImpl) getActiveNode() string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.activeNode
}

func (fm *FailoverManagerImpl) markNodeDown(nodeID string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if _, exists := fm.healthyNodes[nodeID]; exists {
		fm.healthyNodes[nodeID] = false
	}
}

func (fm *FailoverManagerImpl) performFailover() bool {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	return fm.performFailoverUnsafe()
}

func (fm *FailoverManagerImpl) performFailoverUnsafe() bool {
	atomic.AddInt64(&fm.stats.totalFailovers, 1)

	// Find next healthy node
	for _, node := range fm.config.FallbackNodes {
		if fm.healthyNodes[node] {
			fm.activeNode = node
			fm.stats.mu.Lock()
			fm.stats.lastFailoverTime = time.Now()
			fm.stats.mu.Unlock()
			return true
		}
	}

	// No healthy nodes found
	fm.activeNode = ""
	return false
}

func (fm *FailoverManagerImpl) createConnection(ctx context.Context, nodeID string) (interfaces.IConn, error) {
	// This is a placeholder implementation
	// In real implementation, this would create a connection to the specific node
	// For now, we'll simulate connection creation

	// TODO: Implement actual connection creation to specific node
	// This would typically involve creating a new config with the node's connection string
	// and using the PGX provider to create a connection

	return nil, fmt.Errorf("connection creation not implemented for node %s", nodeID)
}

func (fm *FailoverManagerImpl) isNodeFailureError(err error) bool {
	if err == nil {
		return false
	}

	// Check for connection-related errors that indicate node failure
	errStr := err.Error()
	nodeFailurePatterns := []string{
		"connection refused",
		"connection reset",
		"network is unreachable",
		"no route to host",
		"connection timeout",
		"server closed the connection",
		"connection lost",
		"broken pipe",
	}

	for _, pattern := range nodeFailurePatterns {
		if contains(errStr, pattern) {
			return true
		}
	}

	return false
}

func (fm *FailoverManagerImpl) startHealthCheck() {
	fm.healthTicker = time.NewTicker(fm.config.HealthCheckInterval)

	go func() {
		for {
			select {
			case <-fm.healthTicker.C:
				fm.performHealthCheck()
			case <-fm.stopHealth:
				return
			}
		}
	}()
}

func (fm *FailoverManagerImpl) performHealthCheck() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	now := time.Now()

	// Check each node's health
	for _, node := range fm.config.FallbackNodes {
		// Simple health check: if we haven't seen the node for too long, mark it as down
		lastSeen, exists := fm.nodeLastSeen[node]
		if !exists {
			fm.nodeLastSeen[node] = now
			continue
		}

		// If node hasn't been seen for 3x the health check interval, mark it as down
		threshold := fm.config.HealthCheckInterval * 3
		if now.Sub(lastSeen) > threshold {
			fm.healthyNodes[node] = false
		}

		// TODO: Implement actual health check by attempting to connect/ping the node
		// This is a simplified implementation
	}

	// If active node is down, try to failover
	if fm.activeNode != "" && !fm.healthyNodes[fm.activeNode] {
		fm.performFailoverUnsafe()
	}
}

func (fm *FailoverManagerImpl) Stop() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.stopped {
		return
	}

	fm.stopped = true

	if fm.healthTicker != nil {
		fm.healthTicker.Stop()
	}

	if fm.stopHealth != nil {
		close(fm.stopHealth)
	}
}

// Helper function to check if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
