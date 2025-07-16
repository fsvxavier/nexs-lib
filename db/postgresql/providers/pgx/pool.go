package pgx

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

// PGXPool implements the IPool interface using pgxpool
type PGXPool struct {
	pool          *pgxpool.Pool
	hookManager   interfaces.HookManager
	config        interfaces.Config
	mu            sync.RWMutex
	stats         *PoolStatsImpl
	closed        bool
	bufferPool    interfaces.BufferPool
	safetyMonitor interfaces.SafetyMonitor
}

// PoolStatsImpl implements PoolStats interface
type PoolStatsImpl struct {
	mu                      sync.RWMutex
	acquireCount            int64
	acquireDuration         time.Duration
	canceledAcquireCount    int64
	emptyAcquireCount       int64
	newConnsCount           int64
	maxLifetimeDestroyCount int64
	maxIdleDestroyCount     int64
	// Fields for current pool state
	acquiredConns     int32
	constructingConns int32
	idleConns         int32
	maxConns          int32
	totalConns        int32
}

// NewPool creates a new PGX pool
// NewPool creates a new PGX connection pool using the provider
func NewPool(ctx context.Context, config interfaces.Config) (interfaces.IPool, error) {
	provider := NewPGXProvider()
	return provider.NewPool(ctx, config)
}

// NewPoolWithManagers creates a new PGX connection pool with explicit managers
func NewPoolWithManagers(ctx context.Context, config interfaces.Config, hookManager interfaces.HookManager) (interfaces.IPool, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if hookManager == nil {
		return nil, fmt.Errorf("hook manager cannot be nil")
	}

	// Parse pool configuration
	poolConfig, err := parsePoolConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	// Create pgxpool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	pgxPool := &PGXPool{
		pool:        pool,
		hookManager: hookManager,
		config:      config,
		stats:       &PoolStatsImpl{},
		closed:      false,
	}

	// Test the connection
	if err := pgxPool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pgxPool, nil
}

// parsePoolConfig converts our config to pgxpool.Config
func parsePoolConfig(config interfaces.Config) (*pgxpool.Config, error) {
	poolConfig, err := pgxpool.ParseConfig(config.GetConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Apply pool configuration
	cfg := config.GetPoolConfig()
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	// Configure connection timeout if specified
	if cfg.ConnectTimeout > 0 {
		poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	}

	// Configure TLS if enabled
	tlsConfig := config.GetTLSConfig()
	if tlsConfig.Enabled {
		if tlsConfig.InsecureSkipVerify {
			poolConfig.ConnConfig.TLSConfig = nil // Will use InsecureSkipVerify
		}
		// Additional TLS configuration would go here
	}

	return poolConfig, nil
}

// Acquire gets a connection from the pool
func (p *PGXPool) Acquire(ctx context.Context) (interfaces.IConn, error) {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return nil, fmt.Errorf("pool is closed")
	}
	p.mu.RUnlock()

	// Execute before acquire hooks
	execCtx := &interfaces.ExecutionContext{
		Context:   ctx,
		Operation: "acquire",
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	if err := p.hookManager.ExecuteHooks(interfaces.BeforeAcquireHook, execCtx); err != nil {
		return nil, fmt.Errorf("before acquire hook failed: %w", err)
	}

	start := time.Now()
	pgxConn, err := p.pool.Acquire(ctx)
	duration := time.Since(start)

	// Update stats
	p.updateAcquireStats(duration, err != nil)

	execCtx.Duration = duration
	execCtx.Error = err

	if err != nil {
		// Execute error handling
		_ = p.hookManager.ExecuteHooks(interfaces.OnErrorHook, execCtx)
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}

	// Create our connection wrapper
	conn := &PGXConn{
		conn:        pgxConn,
		pool:        p,
		hookManager: p.hookManager,
		config:      p.config,
		stats:       NewConnectionStats(),
		acquired:    true,
	}

	// Execute after acquire hooks
	if err := p.hookManager.ExecuteHooks(interfaces.AfterAcquireHook, execCtx); err != nil {
		conn.Release()
		return nil, fmt.Errorf("after acquire hook failed: %w", err)
	}

	return conn, nil
}

// AcquireFunc executes a function with an acquired connection
func (p *PGXPool) AcquireFunc(ctx context.Context, f func(interfaces.IConn) error) error {
	conn, err := p.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	return f(conn)
}

// Close closes the pool
func (p *PGXPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.pool.Close()
	p.closed = true
}

// Reset resets the pool (closes and would need to be recreated)
func (p *PGXPool) Reset() {
	p.Close()
}

// Stats returns pool statistics
func (p *PGXPool) Stats() interfaces.PoolStats {
	p.stats.mu.RLock()
	defer p.stats.mu.RUnlock()

	pgxStats := p.pool.Stat()

	return interfaces.PoolStats{
		AcquireCount:            p.stats.acquireCount,
		AcquireDuration:         p.stats.acquireDuration,
		AcquiredConns:           pgxStats.AcquiredConns(),
		CanceledAcquireCount:    p.stats.canceledAcquireCount,
		ConstructingConns:       pgxStats.ConstructingConns(),
		EmptyAcquireCount:       p.stats.emptyAcquireCount,
		IdleConns:               pgxStats.IdleConns(),
		MaxConns:                pgxStats.MaxConns(),
		TotalConns:              pgxStats.TotalConns(),
		NewConnsCount:           p.stats.newConnsCount,
		MaxLifetimeDestroyCount: p.stats.maxLifetimeDestroyCount,
		MaxIdleDestroyCount:     p.stats.maxIdleDestroyCount,
	}
}

// Config returns pool configuration
func (p *PGXPool) Config() interfaces.PoolConfig {
	return p.config.GetPoolConfig()
}

// Ping tests connectivity to the database
func (p *PGXPool) Ping(ctx context.Context) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return fmt.Errorf("pool is closed")
	}
	p.mu.RUnlock()

	return p.pool.Ping(ctx)
}

// HealthCheck performs a comprehensive health check
func (p *PGXPool) HealthCheck(ctx context.Context) error {
	// Check if pool is closed
	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Check pool stats to ensure we have healthy connections
	stats := p.pool.Stat()
	if stats.TotalConns() == 0 && stats.MaxConns() > 0 {
		return fmt.Errorf("no connections available in pool")
	}

	// Try to acquire and immediately release a connection
	conn, err := p.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection for health check: %w", err)
	}
	defer conn.Release()

	// Perform a simple query
	_, err = conn.Exec(ctx, "SELECT 1")
	if err != nil {
		return fmt.Errorf("health check query failed: %w", err)
	}

	return nil
}

// GetHookManager returns the hook manager
func (p *PGXPool) GetHookManager() interfaces.HookManager {
	return p.hookManager
}

// GetBufferPool returns the buffer pool for memory optimization
func (p *PGXPool) GetBufferPool() interfaces.BufferPool {
	if p.bufferPool == nil {
		p.bufferPool = NewBufferPool()
	}
	return p.bufferPool
}

// GetSafetyMonitor returns the safety monitor for thread-safety monitoring
func (p *PGXPool) GetSafetyMonitor() interfaces.SafetyMonitor {
	if p.safetyMonitor == nil {
		p.safetyMonitor = NewSafetyMonitor()
	}
	return p.safetyMonitor
}

// NewPoolStats creates a new PoolStatsImpl instance
func NewPoolStats() *PoolStatsImpl {
	return &PoolStatsImpl{}
}

// IncrementAcquireCount increments the acquire count
func (ps *PoolStatsImpl) IncrementAcquireCount() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.acquireCount++
}

// AddAcquireDuration adds to the total acquire duration
func (ps *PoolStatsImpl) AddAcquireDuration(duration time.Duration) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.acquireDuration += duration
}

// IncrementCanceledAcquireCount increments the canceled acquire count
func (ps *PoolStatsImpl) IncrementCanceledAcquireCount() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.canceledAcquireCount++
}

// IncrementEmptyAcquireCount increments the empty acquire count
func (ps *PoolStatsImpl) IncrementEmptyAcquireCount() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.emptyAcquireCount++
}

// IncrementNewConnsCount increments the new connections count
func (ps *PoolStatsImpl) IncrementNewConnsCount() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.newConnsCount++
}

// IncrementMaxLifetimeDestroyCount increments the max lifetime destroy count
func (ps *PoolStatsImpl) IncrementMaxLifetimeDestroyCount() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.maxLifetimeDestroyCount++
}

// IncrementMaxIdleDestroyCount increments the max idle destroy count
func (ps *PoolStatsImpl) IncrementMaxIdleDestroyCount() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.maxIdleDestroyCount++
}

// SetAcquiredConns sets the acquired connections count
func (ps *PoolStatsImpl) SetAcquiredConns(count int32) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.acquiredConns = count
}

// SetConstructingConns sets the constructing connections count
func (ps *PoolStatsImpl) SetConstructingConns(count int32) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.constructingConns = count
}

// SetIdleConns sets the idle connections count
func (ps *PoolStatsImpl) SetIdleConns(count int32) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.idleConns = count
}

// SetMaxConns sets the max connections count
func (ps *PoolStatsImpl) SetMaxConns(count int32) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.maxConns = count
}

// SetTotalConns sets the total connections count
func (ps *PoolStatsImpl) SetTotalConns(count int32) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.totalConns = count
}

// Stats returns the pool statistics
func (ps *PoolStatsImpl) Stats() interfaces.PoolStats {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return interfaces.PoolStats{
		AcquireCount:            ps.acquireCount,
		AcquireDuration:         ps.acquireDuration,
		AcquiredConns:           ps.acquiredConns,
		CanceledAcquireCount:    ps.canceledAcquireCount,
		ConstructingConns:       ps.constructingConns,
		EmptyAcquireCount:       ps.emptyAcquireCount,
		IdleConns:               ps.idleConns,
		MaxConns:                ps.maxConns,
		TotalConns:              ps.totalConns,
		NewConnsCount:           ps.newConnsCount,
		MaxLifetimeDestroyCount: ps.maxLifetimeDestroyCount,
		MaxIdleDestroyCount:     ps.maxIdleDestroyCount,
	}
}

// updateAcquireStats updates acquisition statistics
func (p *PGXPool) updateAcquireStats(duration time.Duration, failed bool) {
	p.stats.mu.Lock()
	defer p.stats.mu.Unlock()

	p.stats.acquireCount++
	p.stats.acquireDuration += duration

	if failed {
		p.stats.canceledAcquireCount++
	}
}
