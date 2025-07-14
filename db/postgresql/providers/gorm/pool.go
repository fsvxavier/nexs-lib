package gorm

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"gorm.io/gorm"
)

// Pool represents a GORM database pool
type Pool struct {
	db     *gorm.DB
	sqlDB  *sql.DB
	config *config.Config
	closed bool
}

// Acquire gets a connection from the pool (in GORM, this returns a reference to the main DB)
func (p *Pool) Acquire(ctx context.Context) (postgresql.IConn, error) {
	if p.closed {
		return nil, fmt.Errorf("pool is closed")
	}

	return &Conn{
		db:         p.db,
		config:     p.config,
		released:   false,
		isFromPool: true,
	}, nil
}

// AcquireWithTimeout gets a connection with timeout
func (p *Pool) AcquireWithTimeout(ctx context.Context, timeout time.Duration) (postgresql.IConn, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return p.Acquire(ctxWithTimeout)
}

// Stats returns pool statistics
func (p *Pool) Stats() postgresql.PoolStats {
	if p.sqlDB == nil {
		return postgresql.PoolStats{}
	}

	stats := p.sqlDB.Stats()
	return postgresql.PoolStats{
		AcquiredConns: int32(stats.InUse),
		IdleConns:     int32(stats.Idle),
		MaxConns:      int32(stats.MaxOpenConnections),
		TotalConns:    int32(stats.OpenConnections),
	}
}

// Close closes the pool
func (p *Pool) Close() {
	if p.closed {
		return
	}

	p.closed = true
	if p.sqlDB != nil {
		p.sqlDB.Close()
	}
}

// Config returns the pool configuration
func (p *Pool) Config() *config.Config {
	return p.config
}

// Ping checks if the pool is healthy
func (p *Pool) Ping(ctx context.Context) error {
	if p.closed {
		return fmt.Errorf("pool is closed")
	}

	if p.sqlDB == nil {
		return fmt.Errorf("underlying sql.DB is nil")
	}

	return p.sqlDB.PingContext(ctx)
}

// HealthCheck checks if the pool is healthy
func (p *Pool) HealthCheck(ctx context.Context) error {
	return p.Ping(ctx)
}

// GetConnWithNotPresent gets a connection with notification presence check
func (p *Pool) GetConnWithNotPresent(ctx context.Context, conn postgresql.IConn) (postgresql.IConn, func(), error) {
	// GORM doesn't support notification presence checks
	acquiredConn, err := p.Acquire(ctx)
	if err != nil {
		return nil, nil, err
	}

	releaseFunc := func() {
		acquiredConn.Release(ctx)
	}

	return acquiredConn, releaseFunc, nil
}

// BeforeAcquireHook executes before acquiring a connection
func (p *Pool) BeforeAcquireHook(ctx context.Context) error {
	// Simplified hook execution
	return nil
}

// AfterReleaseHook executes after releasing a connection
func (p *Pool) AfterReleaseHook(ctx context.Context) error {
	// Simplified hook execution
	return nil
}
