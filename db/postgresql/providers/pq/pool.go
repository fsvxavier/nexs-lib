package pq

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
)

// Pool represents a lib/pq database pool
type Pool struct {
	db     *sql.DB
	config *config.Config
	closed bool
}

// Acquire gets a connection from the pool
func (p *Pool) Acquire(ctx context.Context) (postgresql.IConn, error) {
	if p.closed {
		return nil, fmt.Errorf("pool is closed")
	}

	// lib/pq doesn't have explicit connection acquisition
	// We return a wrapper around the database
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
	if p.db == nil {
		return postgresql.PoolStats{}
	}

	stats := p.db.Stats()
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
	if p.db != nil {
		p.db.Close()
	}
}

// Ping checks if the pool is healthy
func (p *Pool) Ping(ctx context.Context) error {
	if p.closed {
		return fmt.Errorf("pool is closed")
	}

	if p.db == nil {
		return fmt.Errorf("database is nil")
	}

	return p.db.PingContext(ctx)
}

// HealthCheck checks if the pool is healthy
func (p *Pool) HealthCheck(ctx context.Context) error {
	return p.Ping(ctx)
}

// GetConnWithNotPresent gets a connection with notification presence check
func (p *Pool) GetConnWithNotPresent(ctx context.Context, conn postgresql.IConn) (postgresql.IConn, func(), error) {
	// For lib/pq with LISTEN/NOTIFY support
	acquiredConn, err := p.Acquire(ctx)
	if err != nil {
		return nil, nil, err
	}

	releaseFunc := func() {
		acquiredConn.Release(ctx)
	}

	return acquiredConn, releaseFunc, nil
}
