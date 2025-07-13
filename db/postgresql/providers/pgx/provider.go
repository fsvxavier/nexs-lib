package pgx

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/interfaces"
)

// Provider implements the DatabaseProvider interface for PGX driver
type Provider struct {
	cfg    *config.Config
	pool   *pgxpool.Pool
	mu     sync.RWMutex
	closed bool
}

// NewProvider creates a new PGX provider
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

	if p.pool != nil {
		return nil // already connected
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.cfg.ConnectTimeout)
	defer cancel()

	config, err := pgxpool.ParseConfig(p.cfg.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure TLS if enabled
	if p.cfg.TLSEnabled {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		config.ConnConfig.TLSConfig = tlsConfig
	}

	// Configure runtime parameters
	config.ConnConfig.RuntimeParams["timezone"] = p.cfg.Timezone
	config.ConnConfig.RuntimeParams["application_name"] = p.cfg.ApplicationName
	config.ConnConfig.RuntimeParams["search_path"] = p.cfg.SearchPath

	// Configure statement timeout if set
	if p.cfg.StatementTimeout > 0 {
		config.ConnConfig.RuntimeParams["statement_timeout"] = fmt.Sprintf("%dms", p.cfg.StatementTimeout.Milliseconds())
	}

	// Configure lock timeout if set
	if p.cfg.LockTimeout > 0 {
		config.ConnConfig.RuntimeParams["lock_timeout"] = fmt.Sprintf("%dms", p.cfg.LockTimeout.Milliseconds())
	}

	// Configure idle in transaction timeout if set
	if p.cfg.IdleInTransaction > 0 {
		config.ConnConfig.RuntimeParams["idle_in_transaction_session_timeout"] = fmt.Sprintf("%dms", p.cfg.IdleInTransaction.Milliseconds())
	}

	// Configure query execution mode
	switch p.cfg.QueryMode {
	case "CACHE_STATEMENT":
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement
	case "CACHE_DESCRIBE":
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe
	case "DESCRIBE_EXEC":
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeDescribeExec
	case "SIMPLE_PROTOCOL":
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	default:
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	}

	// Configure pool settings
	config.MaxConns = int32(p.cfg.MaxOpenConns)
	config.MinConns = int32(p.cfg.MinConns)
	config.MaxConnLifetime = p.cfg.ConnMaxLifetime
	config.MaxConnIdleTime = p.cfg.ConnMaxIdleTime

	// Create pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	p.pool = pool
	return nil
}

// Close closes the database connection
func (p *Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	if p.pool != nil {
		p.pool.Close()
	}

	p.closed = true
	return nil
}

// DB returns the underlying database instance
func (p *Provider) DB() any {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.pool
}

// Pool returns the IPool interface
func (p *Provider) Pool() interfaces.IPool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.pool == nil {
		return nil
	}
	return &Pool{pool: p.pool, multiTenantEnabled: p.cfg.MultiTenantEnabled}
}

// Pool implements the IPool interface for PGX
type Pool struct {
	pool               *pgxpool.Pool
	multiTenantEnabled bool
}

// Acquire acquires a connection from the pool
func (p *Pool) Acquire(ctx context.Context) (interfaces.IConn, error) {
	poolConn, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		conn:               poolConn,
		multiTenantEnabled: p.multiTenantEnabled,
	}

	if p.multiTenantEnabled {
		if err := conn.AfterAcquireHook(ctx); err != nil {
			poolConn.Release()
			return nil, err
		}
	}

	return conn, nil
}

// Close closes the pool
func (p *Pool) Close() {
	p.pool.Close()
}

// Ping pings the database
func (p *Pool) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
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
	stats := p.pool.Stat()
	return interfaces.PoolStats{
		AcquireCount:         stats.AcquireCount(),
		AcquireDuration:      stats.AcquireDuration(),
		AcquiredConns:        stats.AcquiredConns(),
		CanceledAcquireCount: stats.CanceledAcquireCount(),
		ConstructingConns:    stats.ConstructingConns(),
		EmptyAcquireCount:    stats.EmptyAcquireCount(),
		IdleConns:            stats.IdleConns(),
		MaxConns:             stats.MaxConns(),
		TotalConns:           stats.TotalConns(),
	}
}

// Conn implements the IConn interface for PGX
type Conn struct {
	conn               *pgxpool.Conn
	multiTenantEnabled bool
}

// QueryOne executes a query that returns at most one row
func (c *Conn) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return pgx.ErrNoRows
	}

	return rows.Scan(dst)
}

// QueryAll executes a query that returns multiple rows
func (c *Conn) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	rows, err := c.conn.Query(ctx, query, args...)
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
	var count int
	err := c.conn.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return nil, err
	}
	return &count, nil
}

// Query executes a query that returns rows
func (c *Conn) Query(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	rows, err := c.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows: rows}, nil
}

// Exec executes a query without returning any rows
func (c *Conn) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := c.conn.Exec(ctx, query, args...)
	return err
}

// SendBatch sends a batch of queries
func (c *Conn) SendBatch(ctx context.Context, batch interfaces.IBatch) (interfaces.IBatchResults, error) {
	pgxBatch, ok := batch.(*Batch)
	if !ok {
		return nil, fmt.Errorf("invalid batch type")
	}

	batchResults := c.conn.SendBatch(ctx, pgxBatch.batch)
	return &BatchResults{results: batchResults}, nil
}

// QueryRow executes a query that returns at most one row
func (c *Conn) QueryRow(ctx context.Context, query string, args ...interface{}) (interfaces.IRow, error) {
	row := c.conn.QueryRow(ctx, query, args...)
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
	c.conn.Release()
}

// Ping pings the database
func (c *Conn) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx)
}

// BeginTransaction begins a new transaction
func (c *Conn) BeginTransaction(ctx context.Context) (interfaces.ITransaction, error) {
	tx, err := c.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx, multiTenantEnabled: c.multiTenantEnabled}, nil
}

// BeginTransactionWithOptions begins a new transaction with options
func (c *Conn) BeginTransactionWithOptions(ctx context.Context, options interfaces.TxOptions) (interfaces.ITransaction, error) {
	txOptions := pgx.TxOptions{}

	// Map isolation level
	switch options.IsolationLevel {
	case "READ_UNCOMMITTED":
		txOptions.IsoLevel = pgx.ReadUncommitted
	case "READ_COMMITTED":
		txOptions.IsoLevel = pgx.ReadCommitted
	case "REPEATABLE_READ":
		txOptions.IsoLevel = pgx.RepeatableRead
	case "SERIALIZABLE":
		txOptions.IsoLevel = pgx.Serializable
	default:
		txOptions.IsoLevel = pgx.ReadCommitted
	}

	// Map access mode
	if options.ReadOnly {
		txOptions.AccessMode = pgx.ReadOnly
	} else {
		txOptions.AccessMode = pgx.ReadWrite
	}

	// Map deferrable
	if options.Deferrable {
		txOptions.DeferrableMode = pgx.Deferrable
	} else {
		txOptions.DeferrableMode = pgx.NotDeferrable
	}

	tx, err := c.conn.BeginTx(ctx, txOptions)
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx, multiTenantEnabled: c.multiTenantEnabled}, nil
}

// Transaction implements the ITransaction interface for PGX
type Transaction struct {
	tx                 pgx.Tx
	multiTenantEnabled bool
}

// Commit commits the transaction
func (t *Transaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

// QueryOne executes a query that returns at most one row
func (t *Transaction) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	rows, err := t.tx.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return pgx.ErrNoRows
	}

	return rows.Scan(dst)
}

// QueryAll executes a query that returns multiple rows
func (t *Transaction) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	rows, err := t.tx.Query(ctx, query, args...)
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
	err := t.tx.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return nil, err
	}
	return &count, nil
}

// Query executes a query that returns rows
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	rows, err := t.tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows: rows}, nil
}

// Exec executes a query without returning any rows
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := t.tx.Exec(ctx, query, args...)
	return err
}

// SendBatch sends a batch of queries
func (t *Transaction) SendBatch(ctx context.Context, batch interfaces.IBatch) (interfaces.IBatchResults, error) {
	pgxBatch, ok := batch.(*Batch)
	if !ok {
		return nil, fmt.Errorf("invalid batch type")
	}

	batchResults := t.tx.SendBatch(ctx, pgxBatch.batch)
	return &BatchResults{results: batchResults}, nil
}

// QueryRow executes a query that returns at most one row
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) (interfaces.IRow, error) {
	row := t.tx.QueryRow(ctx, query, args...)
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

// Batch implements the IBatch interface for PGX
type Batch struct {
	batch *pgx.Batch
}

// NewBatch creates a new batch
func NewBatch() interfaces.IBatch {
	return &Batch{batch: &pgx.Batch{}}
}

// Queue adds a query to the batch
func (b *Batch) Queue(query string, arguments ...any) {
	b.batch.Queue(query, arguments...)
}

// Len returns the number of queries in the batch
func (b *Batch) Len() int {
	return b.batch.Len()
}

// BatchResults implements the IBatchResults interface for PGX
type BatchResults struct {
	results pgx.BatchResults
}

// QueryOne executes the next query and scans the result into dst
func (br *BatchResults) QueryOne(dst interface{}) error {
	rows, err := br.results.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return pgx.ErrNoRows
	}

	return rows.Scan(dst)
}

// QueryAll executes the next query and collects all results
func (br *BatchResults) QueryAll(dst interface{}) error {
	rows, err := br.results.Query()
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

// Exec executes the next query
func (br *BatchResults) Exec() error {
	_, err := br.results.Exec()
	return err
}

// Close closes the batch results
func (br *BatchResults) Close() {
	br.results.Close()
}

// Row implements the IRow interface for PGX
type Row struct {
	row pgx.Row
}

// Scan scans the row into dest
func (r *Row) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}

// Rows implements the IRows interface for PGX
type Rows struct {
	rows pgx.Rows
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
	return r.rows.RawValues()
}

// Err returns any error that occurred during iteration
func (r *Rows) Err() error {
	return r.rows.Err()
}
