package pgx

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool implements postgresql.IPool for PGX driver
type Pool struct {
	pool               *pgxpool.Pool
	config             *config.Config
	multiTenantEnabled bool
	logger             *log.Logger
	metrics            map[string]interface{}
	mu                 sync.RWMutex
	closed             bool
}

// Acquire acquires a connection from the pool
func (p *Pool) Acquire(ctx context.Context) (postgresql.IConn, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return nil, fmt.Errorf("pool is closed")
	}

	// Apply timeout if configured
	if p.config.ConnectTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.config.ConnectTimeout)
		defer cancel()
	}

	poolConn, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}

	conn := &Conn{
		poolConn:           poolConn,
		conn:               poolConn.Conn(),
		config:             p.config,
		multiTenantEnabled: p.multiTenantEnabled,
		logger:             p.logger,
		isPooled:           true,
	}

	// Execute after acquire hook if multi-tenant is enabled
	if p.multiTenantEnabled {
		if err := conn.AfterAcquireHook(ctx); err != nil {
			poolConn.Release()
			return nil, fmt.Errorf("after acquire hook failed: %w", err)
		}
	}

	// Execute custom hook if configured
	if p.config.Hooks != nil && p.config.Hooks.AfterAcquire != nil {
		if err := p.config.Hooks.AfterAcquire(ctx, conn); err != nil {
			poolConn.Release()
			return nil, fmt.Errorf("custom after acquire hook failed: %w", err)
		}
	}

	return conn, nil
}

// AcquireWithTimeout acquires a connection with a specific timeout
func (p *Pool) AcquireWithTimeout(ctx context.Context, timeout time.Duration) (postgresql.IConn, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return p.Acquire(ctx)
}

// Close closes the connection pool
func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.closed = true

	if p.pool != nil {
		p.pool.Close()
		p.logger.Printf("Pool closed")
	}
}

// Ping checks if the pool is alive
func (p *Pool) Ping(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("pool is closed")
	}

	// Apply timeout if configured
	if p.config.HealthCheckTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.config.HealthCheckTimeout)
		defer cancel()
	}

	conn, err := p.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection for ping: %w", err)
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, "SELECT 1")
	if err != nil {
		return fmt.Errorf("ping query failed: %w", err)
	}
	defer rows.Close()

	return nil
}

// HealthCheck performs a comprehensive health check
func (p *Pool) HealthCheck(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return fmt.Errorf("pool is closed")
	}

	// Basic ping
	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Check pool statistics
	stats := p.pool.Stat()
	if stats.TotalConns() == 0 {
		return fmt.Errorf("no connections in pool")
	}

	// Check if we can acquire multiple connections
	conn1, err := p.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire first connection: %w", err)
	}
	defer conn1.Release(ctx)

	conn2, err := p.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire second connection: %w", err)
	}
	defer conn2.Release(ctx)

	return nil
}

// Stats returns pool statistics
func (p *Pool) Stats() postgresql.PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed || p.pool == nil {
		return postgresql.PoolStats{}
	}

	stats := p.pool.Stat()

	return postgresql.PoolStats{
		AcquireCount:            stats.AcquireCount(),
		AcquireDuration:         stats.AcquireDuration(),
		AcquiredConns:           stats.AcquiredConns(),
		CanceledAcquireCount:    stats.CanceledAcquireCount(),
		ConstructingConns:       stats.ConstructingConns(),
		EmptyAcquireCount:       stats.EmptyAcquireCount(),
		IdleConns:               stats.IdleConns(),
		MaxConns:                stats.MaxConns(),
		TotalConns:              stats.TotalConns(),
		NewConnsCount:           stats.NewConnsCount(),
		MaxLifetimeDestroyCount: stats.MaxLifetimeDestroyCount(),
		MaxIdleDestroyCount:     stats.MaxIdleDestroyCount(),
	}
}

// GetConnWithNotPresent checks if the provided connection is nil and acquires a new one if needed
func (p *Pool) GetConnWithNotPresent(ctx context.Context, conn postgresql.IConn) (postgresql.IConn, func(), error) {
	if conn == nil {
		newConn, err := p.Acquire(ctx)
		if err != nil {
			return nil, func() {}, fmt.Errorf("failed to acquire connection: %w", err)
		}

		return newConn, func() { newConn.Release(ctx) }, nil
	}

	return conn, func() {}, nil
}
