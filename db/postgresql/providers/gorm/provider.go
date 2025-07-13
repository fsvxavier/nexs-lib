package gorm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/interfaces"
)

// Provider implements the DatabaseProvider interface for GORM driver
type Provider struct {
	cfg    *config.Config
	db     *gorm.DB
	mu     sync.RWMutex
	closed bool
}

// NewProvider creates a new GORM provider
func NewProvider(cfg *config.Config) (interfaces.DatabaseProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	return &Provider{
		cfg: cfg,
	}, nil
}

// Connect establishes a connection to the database
func (p *Provider) Connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("provider is closed")
	}

	if p.db != nil {
		return nil // already connected
	}

	// Configure GORM logger
	logLevel := logger.Info
	if !p.cfg.LoggingEnabled {
		logLevel = logger.Silent
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(p.cfg.DSN()), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(p.cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(p.cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(p.cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(p.cfg.ConnMaxIdleTime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), p.cfg.ConnectTimeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	p.db = db
	return nil
}

// Close closes the database connection
func (p *Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	if p.db != nil {
		sqlDB, err := p.db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	p.closed = true
	return nil
}

// DB returns the underlying database instance
func (p *Provider) DB() any {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.db
}

// Pool returns the IPool interface
func (p *Provider) Pool() interfaces.IPool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.db == nil {
		return nil
	}
	return &Pool{db: p.db, multiTenantEnabled: p.cfg.MultiTenantEnabled}
}

// Pool implements the IPool interface for GORM
type Pool struct {
	db                 *gorm.DB
	multiTenantEnabled bool
}

// Acquire acquires a connection from the pool
func (p *Pool) Acquire(ctx context.Context) (interfaces.IConn, error) {
	// GORM doesn't expose individual connections like pgx,
	// so we wrap the GORM DB instance
	conn := &Conn{
		db:                 p.db.WithContext(ctx),
		multiTenantEnabled: p.multiTenantEnabled,
	}

	if p.multiTenantEnabled {
		if err := conn.AfterAcquireHook(ctx); err != nil {
			return nil, err
		}
	}

	return conn, nil
}

// Close closes the pool
func (p *Pool) Close() {
	if sqlDB, err := p.db.DB(); err == nil {
		sqlDB.Close()
	}
}

// Ping pings the database
func (p *Pool) Ping(ctx context.Context) error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// GetConnWithNotPresent gets a connection if not already present
func (p *Pool) GetConnWithNotPresent(ctx context.Context, conn interfaces.IConn) (interfaces.IConn, func(), error) {
	if conn != nil {
		return conn, func() {}, nil
	}

	newConn, err := p.Acquire(ctx)
	if err != nil {
		return nil, nil, err
	}

	releaseFunc := func() {
		newConn.Release(ctx)
	}

	return newConn, releaseFunc, nil
}

// Stats returns pool statistics
func (p *Pool) Stats() interfaces.PoolStats {
	if p.db == nil {
		return interfaces.PoolStats{}
	}

	sqlDB, err := p.db.DB()
	if err != nil {
		return interfaces.PoolStats{}
	}

	stats := sqlDB.Stats()
	return interfaces.PoolStats{
		MaxConns:     int32(stats.MaxOpenConnections),
		TotalConns:   int32(stats.OpenConnections),
		IdleConns:    int32(stats.Idle),
		AcquireCount: int64(stats.OpenConnections),
	}
}

// Conn implements the IConn interface for GORM
type Conn struct {
	db                 *gorm.DB
	multiTenantEnabled bool
}

// QueryOne executes a query that returns at most one row
func (c *Conn) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	result := c.db.WithContext(ctx).Raw(query, args...).Scan(dst)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// QueryAll executes a query that returns multiple rows
func (c *Conn) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	result := c.db.WithContext(ctx).Raw(query, args...).Scan(dst)
	return result.Error
}

// QueryCount executes a count query
func (c *Conn) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	var count int64
	result := c.db.WithContext(ctx).Raw(query, args...).Scan(&count)
	if result.Error != nil {
		return nil, result.Error
	}
	intCount := int(count)
	return &intCount, nil
}

// Query executes a query that returns rows
func (c *Conn) Query(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	// GORM doesn't provide direct access to sql.Rows
	// This is a limitation - in practice, you'd need to use GORM's native methods
	return nil, fmt.Errorf("raw query with rows not supported in GORM provider - use QueryAll instead")
}

// Exec executes a query without returning any rows
func (c *Conn) Exec(ctx context.Context, query string, args ...interface{}) error {
	result := c.db.WithContext(ctx).Exec(query, args...)
	return result.Error
}

// SendBatch sends a batch of queries
func (c *Conn) SendBatch(ctx context.Context, batch interfaces.IBatch) (interfaces.IBatchResults, error) {
	// GORM doesn't support batching like pgx
	return nil, fmt.Errorf("batch operations not supported in GORM provider")
}

// QueryRow executes a query that returns at most one row
func (c *Conn) QueryRow(ctx context.Context, query string, args ...interface{}) (interfaces.IRow, error) {
	row := &Row{
		db:    c.db.WithContext(ctx),
		query: query,
		args:  args,
	}
	return row, nil
}

// QueryRows executes a query that returns multiple rows
func (c *Conn) QueryRows(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	return c.Query(ctx, query, args...)
}

// BeforeReleaseHook is called before releasing the connection
func (c *Conn) BeforeReleaseHook(ctx context.Context) error {
	// Implementation for any cleanup before release
	return nil
}

// AfterAcquireHook is called after acquiring the connection
func (c *Conn) AfterAcquireHook(ctx context.Context) error {
	if c.multiTenantEnabled {
		// Implementation for multi-tenant setup
		// This could include setting RLS policies, etc.
	}
	return nil
}

// Release releases the connection back to the pool
func (c *Conn) Release(ctx context.Context) {
	if err := c.BeforeReleaseHook(ctx); err != nil {
		// Log error but don't prevent release
	}
	// GORM manages connections automatically
}

// Ping pings the database
func (c *Conn) Ping(ctx context.Context) error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// BeginTransaction begins a new transaction
func (c *Conn) BeginTransaction(ctx context.Context) (interfaces.ITransaction, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &Transaction{db: tx, multiTenantEnabled: c.multiTenantEnabled}, nil
}

// BeginTransactionWithOptions begins a new transaction with options
func (c *Conn) BeginTransactionWithOptions(ctx context.Context, options interfaces.TxOptions) (interfaces.ITransaction, error) {
	// GORM doesn't support detailed transaction options like pgx
	// This is a limitation of the GORM driver
	return c.BeginTransaction(ctx)
}

// Transaction implements the ITransaction interface for GORM
type Transaction struct {
	db                 *gorm.DB
	multiTenantEnabled bool
}

// Commit commits the transaction
func (t *Transaction) Commit(ctx context.Context) error {
	result := t.db.WithContext(ctx).Commit()
	return result.Error
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback(ctx context.Context) error {
	result := t.db.WithContext(ctx).Rollback()
	return result.Error
}

// QueryOne executes a query that returns at most one row
func (t *Transaction) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	result := t.db.WithContext(ctx).Raw(query, args...).Scan(dst)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// QueryAll executes a query that returns multiple rows
func (t *Transaction) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	result := t.db.WithContext(ctx).Raw(query, args...).Scan(dst)
	return result.Error
}

// QueryCount executes a count query
func (t *Transaction) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	var count int64
	result := t.db.WithContext(ctx).Raw(query, args...).Scan(&count)
	if result.Error != nil {
		return nil, result.Error
	}
	intCount := int(count)
	return &intCount, nil
}

// Query executes a query that returns rows
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	return nil, fmt.Errorf("raw query with rows not supported in GORM provider - use QueryAll instead")
}

// Exec executes a query without returning any rows
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) error {
	result := t.db.WithContext(ctx).Exec(query, args...)
	return result.Error
}

// SendBatch sends a batch of queries
func (t *Transaction) SendBatch(ctx context.Context, batch interfaces.IBatch) (interfaces.IBatchResults, error) {
	return nil, fmt.Errorf("batch operations not supported in GORM provider")
}

// QueryRow executes a query that returns at most one row
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) (interfaces.IRow, error) {
	row := &Row{
		db:    t.db.WithContext(ctx),
		query: query,
		args:  args,
	}
	return row, nil
}

// QueryRows executes a query that returns multiple rows
func (t *Transaction) QueryRows(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	return t.Query(ctx, query, args...)
}

// BeforeReleaseHook is called before releasing the connection
func (t *Transaction) BeforeReleaseHook(ctx context.Context) error {
	return nil
}

// AfterAcquireHook is called after acquiring the connection
func (t *Transaction) AfterAcquireHook(ctx context.Context) error {
	return nil
}

// Release releases the connection (no-op for transactions)
func (t *Transaction) Release(ctx context.Context) {
	// No-op for transactions
}

// Ping pings the database
func (t *Transaction) Ping(ctx context.Context) error {
	// No-op for transactions
	return nil
}

// BeginTransaction begins a nested transaction
func (t *Transaction) BeginTransaction(ctx context.Context) (interfaces.ITransaction, error) {
	// GORM supports nested transactions through SavePoint
	tx := t.db.WithContext(ctx).SavePoint("nested_" + fmt.Sprintf("%d", time.Now().UnixNano()))
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &Transaction{db: tx, multiTenantEnabled: t.multiTenantEnabled}, nil
}

// BeginTransactionWithOptions begins a nested transaction with options
func (t *Transaction) BeginTransactionWithOptions(ctx context.Context, options interfaces.TxOptions) (interfaces.ITransaction, error) {
	return t.BeginTransaction(ctx)
}

// Row implements the IRow interface for GORM
type Row struct {
	db    *gorm.DB
	query string
	args  []interface{}
}

// Scan scans the row into dest
func (r *Row) Scan(dest ...any) error {
	result := r.db.Raw(r.query, r.args...).Scan(dest)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Note: GORM doesn't easily support IBatch and IRows interfaces
// due to its ORM nature. These are marked as not supported.
// In a real implementation, you might want to either:
// 1. Use GORM's native methods instead of these interfaces
// 2. Use the underlying sql.DB for raw operations
// 3. Consider using pgx or pq drivers for more low-level control
