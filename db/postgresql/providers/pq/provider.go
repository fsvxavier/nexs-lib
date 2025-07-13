package pq

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/interfaces"
)

// Provider implements the DatabaseProvider interface for lib/pq driver
type Provider struct {
	cfg    *config.Config
	db     *sql.DB
	mu     sync.RWMutex
	closed bool
}

// NewProvider creates a new lib/pq provider
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

	// Open database connection
	db, err := sql.Open("postgres", p.cfg.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(p.cfg.MaxOpenConns)
	db.SetMaxIdleConns(p.cfg.MaxIdleConns)
	db.SetConnMaxLifetime(p.cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(p.cfg.ConnMaxIdleTime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), p.cfg.ConnectTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
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
		p.db.Close()
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

// Pool implements the IPool interface for lib/pq
type Pool struct {
	db                 *sql.DB
	multiTenantEnabled bool
}

// Acquire acquires a connection from the pool
func (p *Pool) Acquire(ctx context.Context) (interfaces.IConn, error) {
	// lib/pq doesn't expose individual connections like pgx,
	// so we wrap the sql.DB instance
	conn := &Conn{
		db:                 p.db,
		multiTenantEnabled: p.multiTenantEnabled,
		ctx:                ctx,
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
	p.db.Close()
}

// Ping pings the database
func (p *Pool) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
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

	stats := p.db.Stats()
	return interfaces.PoolStats{
		MaxConns:     int32(stats.MaxOpenConnections),
		TotalConns:   int32(stats.OpenConnections),
		IdleConns:    int32(stats.Idle),
		AcquireCount: int64(stats.OpenConnections),
	}
}

// Conn implements the IConn interface for lib/pq
type Conn struct {
	db                 *sql.DB
	multiTenantEnabled bool
	ctx                context.Context
}

// QueryOne executes a query that returns at most one row
func (c *Conn) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	ctx = c.getContext(ctx)
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	return rows.Scan(dst)
}

// QueryAll executes a query that returns multiple rows
func (c *Conn) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	ctx = c.getContext(ctx)
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// For QueryAll, we expect the caller to handle the iteration
	// This is a simplified implementation - in practice, you'd need reflection
	// or a specific scanning strategy based on your needs
	for rows.Next() {
		if err := rows.Scan(dst); err != nil {
			return err
		}
	}

	return rows.Err()
}

// QueryCount executes a count query
func (c *Conn) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	ctx = c.getContext(ctx)
	var count int
	err := c.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return nil, err
	}
	return &count, nil
}

// Query executes a query that returns rows
func (c *Conn) Query(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	ctx = c.getContext(ctx)
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows: rows}, nil
}

// Exec executes a query without returning any rows
func (c *Conn) Exec(ctx context.Context, query string, args ...interface{}) error {
	ctx = c.getContext(ctx)
	_, err := c.db.ExecContext(ctx, query, args...)
	return err
}

// SendBatch sends a batch of queries
func (c *Conn) SendBatch(ctx context.Context, batch interfaces.IBatch) (interfaces.IBatchResults, error) {
	// lib/pq doesn't support batching like pgx
	// We simulate it by executing queries sequentially
	pqBatch, ok := batch.(*Batch)
	if !ok {
		return nil, fmt.Errorf("invalid batch type")
	}

	return &BatchResults{
		conn:    c,
		ctx:     c.getContext(ctx),
		queries: pqBatch.queries,
		current: 0,
	}, nil
}

// QueryRow executes a query that returns at most one row
func (c *Conn) QueryRow(ctx context.Context, query string, args ...interface{}) (interfaces.IRow, error) {
	ctx = c.getContext(ctx)
	row := c.db.QueryRowContext(ctx, query, args...)
	return &Row{row: row}, nil
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
	// lib/pq manages connections automatically
}

// Ping pings the database
func (c *Conn) Ping(ctx context.Context) error {
	ctx = c.getContext(ctx)
	return c.db.PingContext(ctx)
}

// BeginTransaction begins a new transaction
func (c *Conn) BeginTransaction(ctx context.Context) (interfaces.ITransaction, error) {
	ctx = c.getContext(ctx)
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx, multiTenantEnabled: c.multiTenantEnabled}, nil
}

// BeginTransactionWithOptions begins a new transaction with options
func (c *Conn) BeginTransactionWithOptions(ctx context.Context, options interfaces.TxOptions) (interfaces.ITransaction, error) {
	ctx = c.getContext(ctx)
	txOptions := &sql.TxOptions{}

	// Map isolation level
	switch options.IsolationLevel {
	case "READ_UNCOMMITTED":
		txOptions.Isolation = sql.LevelReadUncommitted
	case "READ_COMMITTED":
		txOptions.Isolation = sql.LevelReadCommitted
	case "REPEATABLE_READ":
		txOptions.Isolation = sql.LevelRepeatableRead
	case "SERIALIZABLE":
		txOptions.Isolation = sql.LevelSerializable
	default:
		txOptions.Isolation = sql.LevelReadCommitted
	}

	// Map read-only
	txOptions.ReadOnly = options.ReadOnly

	tx, err := c.db.BeginTx(ctx, txOptions)
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx, multiTenantEnabled: c.multiTenantEnabled}, nil
}

// getContext returns the context to use, preferring the passed context over the conn context
func (c *Conn) getContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return c.ctx
}

// Transaction implements the ITransaction interface for lib/pq
type Transaction struct {
	tx                 *sql.Tx
	multiTenantEnabled bool
}

// Commit commits the transaction
func (t *Transaction) Commit(ctx context.Context) error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback()
}

// QueryOne executes a query that returns at most one row
func (t *Transaction) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	return rows.Scan(dst)
}

// QueryAll executes a query that returns multiple rows
func (t *Transaction) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// For QueryAll, we expect the caller to handle the iteration
	// This is a simplified implementation - in practice, you'd need reflection
	// or a specific scanning strategy based on your needs
	for rows.Next() {
		if err := rows.Scan(dst); err != nil {
			return err
		}
	}

	return rows.Err()
}

// QueryCount executes a count query
func (t *Transaction) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	var count int
	err := t.tx.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return nil, err
	}
	return &count, nil
}

// Query executes a query that returns rows
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows: rows}, nil
}

// Exec executes a query without returning any rows
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := t.tx.ExecContext(ctx, query, args...)
	return err
}

// SendBatch sends a batch of queries
func (t *Transaction) SendBatch(ctx context.Context, batch interfaces.IBatch) (interfaces.IBatchResults, error) {
	// lib/pq doesn't support batching like pgx
	// We simulate it by executing queries sequentially
	pqBatch, ok := batch.(*Batch)
	if !ok {
		return nil, fmt.Errorf("invalid batch type")
	}

	return &BatchResults{
		tx:      t,
		ctx:     ctx,
		queries: pqBatch.queries,
		current: 0,
	}, nil
}

// QueryRow executes a query that returns at most one row
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) (interfaces.IRow, error) {
	row := t.tx.QueryRowContext(ctx, query, args...)
	return &Row{row: row}, nil
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

// BeginTransaction begins a nested transaction (not supported)
func (t *Transaction) BeginTransaction(ctx context.Context) (interfaces.ITransaction, error) {
	return nil, fmt.Errorf("nested transactions not supported")
}

// BeginTransactionWithOptions begins a nested transaction with options (not supported)
func (t *Transaction) BeginTransactionWithOptions(ctx context.Context, options interfaces.TxOptions) (interfaces.ITransaction, error) {
	return nil, fmt.Errorf("nested transactions not supported")
}

// Batch implements the IBatch interface for lib/pq
type Batch struct {
	queries []batchQuery
}

type batchQuery struct {
	query string
	args  []interface{}
}

// NewBatch creates a new batch
func NewBatch() interfaces.IBatch {
	return &Batch{queries: make([]batchQuery, 0)}
}

// Queue adds a query to the batch
func (b *Batch) Queue(query string, arguments ...any) {
	b.queries = append(b.queries, batchQuery{
		query: query,
		args:  arguments,
	})
}

// Len returns the number of queries in the batch
func (b *Batch) Len() int {
	return len(b.queries)
}

// BatchResults implements the IBatchResults interface for lib/pq
type BatchResults struct {
	conn    *Conn
	tx      *Transaction
	ctx     context.Context
	queries []batchQuery
	current int
}

// QueryOne executes the next query and scans the result into dst
func (br *BatchResults) QueryOne(dst interface{}) error {
	if br.current >= len(br.queries) {
		return fmt.Errorf("no more queries in batch")
	}

	query := br.queries[br.current]
	br.current++

	if br.tx != nil {
		return br.tx.QueryOne(br.ctx, dst, query.query, query.args...)
	}

	if br.conn == nil {
		return fmt.Errorf("no connection available")
	}

	return br.conn.QueryOne(br.ctx, dst, query.query, query.args...)
}

// QueryAll executes the next query and collects all results
func (br *BatchResults) QueryAll(dst interface{}) error {
	if br.current >= len(br.queries) {
		return fmt.Errorf("no more queries in batch")
	}

	query := br.queries[br.current]
	br.current++

	if br.tx != nil {
		return br.tx.QueryAll(br.ctx, dst, query.query, query.args...)
	}

	if br.conn == nil {
		return fmt.Errorf("no connection available")
	}

	return br.conn.QueryAll(br.ctx, dst, query.query, query.args...)
}

// Exec executes the next query
func (br *BatchResults) Exec() error {
	if br.current >= len(br.queries) {
		return fmt.Errorf("no more queries in batch")
	}

	query := br.queries[br.current]
	br.current++

	if br.tx != nil {
		return br.tx.Exec(br.ctx, query.query, query.args...)
	}

	if br.conn == nil {
		return fmt.Errorf("no connection available")
	}

	return br.conn.Exec(br.ctx, query.query, query.args...)
}

// Close closes the batch results
func (br *BatchResults) Close() {
	// No-op for lib/pq
}

// Row implements the IRow interface for lib/pq
type Row struct {
	row *sql.Row
}

// Scan scans the row into dest
func (r *Row) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}

// Rows implements the IRows interface for lib/pq
type Rows struct {
	rows *sql.Rows
}

// Scan scans the current row into dest
func (r *Rows) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

// Close closes the rows
func (r *Rows) Close() {
	r.rows.Close()
}

// Next advances to the next row
func (r *Rows) Next() bool {
	return r.rows.Next()
}

// RawValues returns the raw values of the current row
func (r *Rows) RawValues() [][]byte {
	// lib/pq doesn't provide RawValues like pgx
	// This is a limitation of the driver
	return nil
}

// Err returns any error that occurred during iteration
func (r *Rows) Err() error {
	return r.rows.Err()
}
