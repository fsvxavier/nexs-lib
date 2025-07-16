package pgx

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

// PGXConn implements the IConn interface using pgx
type PGXConn struct {
	conn        *pgxpool.Conn
	rawConn     *pgx.Conn // For single connections
	pool        *PGXPool
	hookManager interfaces.HookManager
	config      interfaces.Config
	stats       *ConnectionStatsImpl
	mu          sync.RWMutex
	acquired    bool
	closed      bool
	tenantID    string
}

// ConnectionStatsImpl implements ConnectionStats interface
type ConnectionStatsImpl struct {
	mu                 sync.RWMutex
	totalQueries       int64
	totalExecs         int64
	totalTransactions  int64
	totalBatches       int64
	failedQueries      int64
	failedExecs        int64
	failedTransactions int64
	totalQueryTime     time.Duration
	totalExecTime      time.Duration
	lastActivity       time.Time
	createdAt          time.Time
}

// NewConnectionStats creates a new ConnectionStats instance
func NewConnectionStats() *ConnectionStatsImpl {
	return &ConnectionStatsImpl{
		createdAt: time.Now(),
	}
}

// IncrementQueries increments the query counter
func (cs *ConnectionStatsImpl) IncrementQueries() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.totalQueries++
}

// IncrementExecs increments the exec counter
func (cs *ConnectionStatsImpl) IncrementExecs() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.totalExecs++
}

// IncrementTransactions increments the transaction counter
func (cs *ConnectionStatsImpl) IncrementTransactions() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.totalTransactions++
}

// IncrementBatches increments the batch counter
func (cs *ConnectionStatsImpl) IncrementBatches() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.totalBatches++
}

// IncrementFailedQueries increments the failed query counter
func (cs *ConnectionStatsImpl) IncrementFailedQueries() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.failedQueries++
}

// IncrementFailedExecs increments the failed exec counter
func (cs *ConnectionStatsImpl) IncrementFailedExecs() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.failedExecs++
}

// IncrementFailedTransactions increments the failed transaction counter
func (cs *ConnectionStatsImpl) IncrementFailedTransactions() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.failedTransactions++
}

// AddQueryTime adds query execution time
func (cs *ConnectionStatsImpl) AddQueryTime(duration time.Duration) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.totalQueryTime += duration
}

// AddExecTime adds exec execution time
func (cs *ConnectionStatsImpl) AddExecTime(duration time.Duration) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.totalExecTime += duration
}

// UpdateLastActivity updates the last activity timestamp
func (cs *ConnectionStatsImpl) UpdateLastActivity() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.lastActivity = time.Now()
}

// GetAverageQueryTime returns the average query time
func (cs *ConnectionStatsImpl) GetAverageQueryTime() time.Duration {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if cs.totalQueries == 0 {
		return 0
	}
	return cs.totalQueryTime / time.Duration(cs.totalQueries)
}

// GetAverageExecTime returns the average exec time
func (cs *ConnectionStatsImpl) GetAverageExecTime() time.Duration {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if cs.totalExecs == 0 {
		return 0
	}
	return cs.totalExecTime / time.Duration(cs.totalExecs)
}

// Stats returns the connection statistics
func (cs *ConnectionStatsImpl) Stats() interfaces.ConnectionStats {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	avgQueryTime := time.Duration(0)
	if cs.totalQueries > 0 {
		avgQueryTime = cs.totalQueryTime / time.Duration(cs.totalQueries)
	}

	avgExecTime := time.Duration(0)
	if cs.totalExecs > 0 {
		avgExecTime = cs.totalExecTime / time.Duration(cs.totalExecs)
	}

	return interfaces.ConnectionStats{
		TotalQueries:       cs.totalQueries,
		TotalExecs:         cs.totalExecs,
		TotalTransactions:  cs.totalTransactions,
		TotalBatches:       cs.totalBatches,
		FailedQueries:      cs.failedQueries,
		FailedExecs:        cs.failedExecs,
		FailedTransactions: cs.failedTransactions,
		AverageQueryTime:   avgQueryTime,
		AverageExecTime:    avgExecTime,
		LastActivity:       cs.lastActivity,
		CreatedAt:          cs.createdAt,
	}
}

// NewConn creates a new PGX connection
// NewConn creates a new PGX connection using the provider
func NewConn(ctx context.Context, config interfaces.Config) (interfaces.IConn, error) {
	provider := NewPGXProvider()
	return provider.NewConn(ctx, config)
}

// NewListenConn creates a new PGX listen connection using the provider
func NewListenConn(ctx context.Context, config interfaces.Config) (interfaces.IConn, error) {
	provider := NewPGXProvider()
	return provider.NewListenConn(ctx, config)
}

// NewConnWithManagers creates a new PGX connection with explicit managers
func NewConnWithManagers(ctx context.Context, config interfaces.Config, hookManager interfaces.HookManager) (interfaces.IConn, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if hookManager == nil {
		return nil, fmt.Errorf("hook manager cannot be nil")
	}

	// Parse connection configuration
	connConfig, err := pgx.ParseConfig(config.GetConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure TLS if enabled
	tlsConfig := config.GetTLSConfig()
	if tlsConfig.Enabled {
		if tlsConfig.InsecureSkipVerify {
			connConfig.TLSConfig = nil
		}
	}

	// Create connection
	rawConn, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	conn := &PGXConn{
		rawConn:     rawConn,
		hookManager: hookManager,
		config:      config,
		stats:       NewConnectionStats(),
		acquired:    false,
		closed:      false,
	}

	// Test the connection
	if err := conn.Ping(ctx); err != nil {
		rawConn.Close(ctx)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return conn, nil
}

// NewListenConnWithManagers creates a new PGX connection for LISTEN/NOTIFY with explicit managers
func NewListenConnWithManagers(ctx context.Context, config interfaces.Config, hookManager interfaces.HookManager) (interfaces.IConn, error) {
	// Listen connections are regular connections with specific usage patterns
	conn, err := NewConnWithManagers(ctx, config, hookManager)
	if err != nil {
		return nil, fmt.Errorf("failed to create listen connection: %w", err)
	}

	return conn, nil
}

// getActiveConn returns the active connection (either pool or raw)
func (c *PGXConn) getActiveConn() interface{} {
	if c.conn != nil {
		return c.conn
	}
	return c.rawConn
}

// executeWithHooksAndMiddleware executes an operation with hooks and middleware
func (c *PGXConn) executeWithHooksAndMiddleware(ctx context.Context, operation string, query string, args []interface{}, execFunc func() error) error {
	// Create execution context
	execCtx := &interfaces.ExecutionContext{
		Context:   ctx,
		Operation: operation,
		Query:     query,
		Args:      args,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Execute before hooks
	if err := c.hookManager.ExecuteHooks(getBeforeHook(operation), execCtx); err != nil {
		return fmt.Errorf("before hook failed: %w", err)
	}

	// Execute the operation
	start := time.Now()
	err := execFunc()
	execCtx.Duration = time.Since(start)
	execCtx.Error = err

	// Update stats
	c.updateStats(operation, execCtx.Duration, err != nil)

	if err != nil {
		// Execute error handling
		_ = c.hookManager.ExecuteHooks(interfaces.OnErrorHook, execCtx)
		return err
	}

	// Execute after hooks
	if err := c.hookManager.ExecuteHooks(getAfterHook(operation), execCtx); err != nil {
		return fmt.Errorf("after hook failed: %w", err)
	}

	return nil
}

// getBeforeHook returns the appropriate before hook for an operation
func getBeforeHook(operation string) interfaces.HookType {
	switch operation {
	case "query":
		return interfaces.BeforeQueryHook
	case "exec":
		return interfaces.BeforeExecHook
	case "transaction":
		return interfaces.BeforeTransactionHook
	case "batch":
		return interfaces.BeforeBatchHook
	default:
		return interfaces.BeforeQueryHook
	}
}

// getAfterHook returns the appropriate after hook for an operation
func getAfterHook(operation string) interfaces.HookType {
	switch operation {
	case "query":
		return interfaces.AfterQueryHook
	case "exec":
		return interfaces.AfterExecHook
	case "transaction":
		return interfaces.AfterTransactionHook
	case "batch":
		return interfaces.AfterBatchHook
	default:
		return interfaces.AfterQueryHook
	}
}

// updateStats updates connection statistics
func (c *PGXConn) updateStats(operation string, duration time.Duration, failed bool) {
	c.stats.mu.Lock()
	defer c.stats.mu.Unlock()

	c.stats.lastActivity = time.Now()

	switch operation {
	case "query":
		c.stats.totalQueries++
		c.stats.totalQueryTime += duration
		if failed {
			c.stats.failedQueries++
		}
	case "exec":
		c.stats.totalExecs++
		c.stats.totalExecTime += duration
		if failed {
			c.stats.failedExecs++
		}
	case "transaction":
		c.stats.totalTransactions++
		if failed {
			c.stats.failedTransactions++
		}
	case "batch":
		c.stats.totalBatches++
	}
}

// QueryRow executes a query that returns at most one row
func (c *PGXConn) QueryRow(ctx context.Context, query string, args ...interface{}) interfaces.IRow {
	var row pgx.Row

	_ = c.executeWithHooksAndMiddleware(ctx, "query", query, args, func() error {
		if c.conn != nil {
			row = c.conn.QueryRow(ctx, query, args...)
		} else {
			row = c.rawConn.QueryRow(ctx, query, args...)
		}
		return nil
	})

	return &PGXRow{row: row}
}

// Query executes a query that returns rows
func (c *PGXConn) Query(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	var rows pgx.Rows
	var err error

	opErr := c.executeWithHooksAndMiddleware(ctx, "query", query, args, func() error {
		if c.conn != nil {
			rows, err = c.conn.Query(ctx, query, args...)
		} else {
			rows, err = c.rawConn.Query(ctx, query, args...)
		}
		return err
	})

	if opErr != nil {
		return nil, opErr
	}

	return &PGXRows{rows: rows}, nil
}

// QueryOne executes a query and scans the result into dst
func (c *PGXConn) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return c.executeWithHooksAndMiddleware(ctx, "query", query, args, func() error {
		var row pgx.Row
		if c.conn != nil {
			row = c.conn.QueryRow(ctx, query, args...)
		} else {
			row = c.rawConn.QueryRow(ctx, query, args...)
		}
		return row.Scan(dst)
	})
}

// QueryAll executes a query and scans all results into dst
func (c *PGXConn) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return c.executeWithHooksAndMiddleware(ctx, "query", query, args, func() error {
		if c.conn != nil {
			return c.conn.QueryRow(ctx, "SELECT count(*) FROM ("+query+") AS count_query", args...).Scan(dst)
		} else {
			return c.rawConn.QueryRow(ctx, "SELECT count(*) FROM ("+query+") AS count_query", args...).Scan(dst)
		}
	})
}

// QueryCount executes a query and returns the count
func (c *PGXConn) QueryCount(ctx context.Context, query string, args ...interface{}) (int64, error) {
	var count int64
	err := c.executeWithHooksAndMiddleware(ctx, "query", query, args, func() error {
		countQuery := "SELECT count(*) FROM (" + query + ") AS count_query"
		var row pgx.Row
		if c.conn != nil {
			row = c.conn.QueryRow(ctx, countQuery, args...)
		} else {
			row = c.rawConn.QueryRow(ctx, countQuery, args...)
		}
		return row.Scan(&count)
	})

	return count, err
}

// Exec executes a query that doesn't return rows
func (c *PGXConn) Exec(ctx context.Context, query string, args ...interface{}) (interfaces.CommandTag, error) {
	var tag pgconn.CommandTag
	var err error

	opErr := c.executeWithHooksAndMiddleware(ctx, "exec", query, args, func() error {
		if c.conn != nil {
			tag, err = c.conn.Exec(ctx, query, args...)
		} else {
			tag, err = c.rawConn.Exec(ctx, query, args...)
		}
		return err
	})

	if opErr != nil {
		return nil, opErr
	}

	return &PGXCommandTag{tag: tag}, nil
}

// SendBatch sends a batch of queries
func (c *PGXConn) SendBatch(ctx context.Context, batch interfaces.IBatch) interfaces.IBatchResults {
	pgxBatch := batch.(*PGXBatch)

	var results pgx.BatchResults
	_ = c.executeWithHooksAndMiddleware(ctx, "batch", "", nil, func() error {
		if c.conn != nil {
			results = c.conn.SendBatch(ctx, pgxBatch.batch)
		} else {
			results = c.rawConn.SendBatch(ctx, pgxBatch.batch)
		}
		return nil
	})

	return &PGXBatchResults{results: results}
}

// Begin starts a transaction
func (c *PGXConn) Begin(ctx context.Context) (interfaces.ITransaction, error) {
	return c.BeginTx(ctx, interfaces.TxOptions{})
}

// BeginTx starts a transaction with options
func (c *PGXConn) BeginTx(ctx context.Context, txOptions interfaces.TxOptions) (interfaces.ITransaction, error) {
	var tx pgx.Tx
	var err error

	opErr := c.executeWithHooksAndMiddleware(ctx, "transaction", "", nil, func() error {
		pgxOptions := convertTxOptions(txOptions)
		if c.conn != nil {
			tx, err = c.conn.BeginTx(ctx, pgxOptions)
		} else {
			tx, err = c.rawConn.BeginTx(ctx, pgxOptions)
		}
		return err
	})

	if opErr != nil {
		return nil, opErr
	}

	return newPGXTransaction(tx, c), nil
}

// Release releases the connection back to the pool
func (c *PGXConn) Release() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.acquired || c.closed {
		return
	}

	// Execute before release hooks
	execCtx := &interfaces.ExecutionContext{
		Context:   context.Background(),
		Operation: "release",
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	_ = c.hookManager.ExecuteHooks(interfaces.BeforeReleaseHook, execCtx)

	if c.conn != nil {
		c.conn.Release()
	}

	c.acquired = false

	// Execute after release hooks
	_ = c.hookManager.ExecuteHooks(interfaces.AfterReleaseHook, execCtx)
}

// Close closes the connection
func (c *PGXConn) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	var err error
	if c.rawConn != nil {
		err = c.rawConn.Close(ctx)
	} else if c.conn != nil {
		c.conn.Release()
	}

	c.closed = true
	return err
}

// Ping tests the connection
func (c *PGXConn) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("connection is closed")
	}

	if c.conn != nil {
		return c.conn.Ping(ctx)
	}
	return c.rawConn.Ping(ctx)
}

// IsClosed returns whether the connection is closed
func (c *PGXConn) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// Prepare prepares a statement
func (c *PGXConn) Prepare(ctx context.Context, name, query string) error {
	return c.executeWithHooksAndMiddleware(ctx, "prepare", query, nil, func() error {
		if c.conn != nil {
			// pgxpool.Conn doesn't have Prepare, use the underlying connection
			_, err := c.conn.Conn().Prepare(ctx, name, query)
			return err
		}
		_, err := c.rawConn.Prepare(ctx, name, query)
		return err
	})
}

// Deallocate deallocates a prepared statement
func (c *PGXConn) Deallocate(ctx context.Context, name string) error {
	return c.executeWithHooksAndMiddleware(ctx, "deallocate", "", nil, func() error {
		if c.conn != nil {
			_, err := c.conn.Exec(ctx, "DEALLOCATE "+name)
			return err
		}
		_, err := c.rawConn.Exec(ctx, "DEALLOCATE "+name)
		return err
	})
}

// GetHookManager returns the hook manager
func (c *PGXConn) GetHookManager() interfaces.HookManager {
	return c.hookManager
}

// HealthCheck performs a health check on the connection
func (c *PGXConn) HealthCheck(ctx context.Context) error {
	return c.Ping(ctx)
}

// Stats returns connection statistics
func (c *PGXConn) Stats() interfaces.ConnectionStats {
	c.stats.mu.RLock()
	defer c.stats.mu.RUnlock()

	avgQueryTime := time.Duration(0)
	if c.stats.totalQueries > 0 {
		avgQueryTime = c.stats.totalQueryTime / time.Duration(c.stats.totalQueries)
	}

	avgExecTime := time.Duration(0)
	if c.stats.totalExecs > 0 {
		avgExecTime = c.stats.totalExecTime / time.Duration(c.stats.totalExecs)
	}

	return interfaces.ConnectionStats{
		TotalQueries:       c.stats.totalQueries,
		TotalExecs:         c.stats.totalExecs,
		TotalTransactions:  c.stats.totalTransactions,
		TotalBatches:       c.stats.totalBatches,
		FailedQueries:      c.stats.failedQueries,
		FailedExecs:        c.stats.failedExecs,
		FailedTransactions: c.stats.failedTransactions,
		AverageQueryTime:   avgQueryTime,
		AverageExecTime:    avgExecTime,
		LastActivity:       c.stats.lastActivity,
		CreatedAt:          c.stats.createdAt,
	}
}

// convertTxOptions converts our TxOptions to pgx.TxOptions
func convertTxOptions(opts interfaces.TxOptions) pgx.TxOptions {
	pgxOpts := pgx.TxOptions{}

	switch opts.IsoLevel {
	case interfaces.TxIsoLevelReadUncommitted:
		pgxOpts.IsoLevel = pgx.ReadUncommitted
	case interfaces.TxIsoLevelReadCommitted:
		pgxOpts.IsoLevel = pgx.ReadCommitted
	case interfaces.TxIsoLevelRepeatableRead:
		pgxOpts.IsoLevel = pgx.RepeatableRead
	case interfaces.TxIsoLevelSerializable:
		pgxOpts.IsoLevel = pgx.Serializable
	}

	switch opts.AccessMode {
	case interfaces.TxAccessModeReadOnly:
		pgxOpts.AccessMode = pgx.ReadOnly
	case interfaces.TxAccessModeReadWrite:
		pgxOpts.AccessMode = pgx.ReadWrite
	}

	switch opts.DeferrableMode {
	case interfaces.TxDeferrableModeDeferrable:
		pgxOpts.DeferrableMode = pgx.Deferrable
	case interfaces.TxDeferrableModeNotDeferrable:
		pgxOpts.DeferrableMode = pgx.NotDeferrable
	}

	if opts.BeginQuery != "" {
		pgxOpts.BeginQuery = opts.BeginQuery
	}

	return pgxOpts
}

// SetTenant implements IConn.SetTenant
func (c *PGXConn) SetTenant(ctx context.Context, tenantID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tenantID = tenantID
	return nil
}

// GetTenant implements IConn.GetTenant
func (c *PGXConn) GetTenant(ctx context.Context) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tenantID, nil
}

// CopyFrom implements IConn.CopyFrom
func (c *PGXConn) CopyFrom(ctx context.Context, tableName string, columnNames []string, rowSrc interfaces.CopyFromSource) (int64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return 0, ErrConnectionClosed
	}

	startTime := time.Now()

	// Execute hooks and middlewares
	execCtx := &interfaces.ExecutionContext{
		Context:   ctx,
		Operation: "CopyFrom",
		StartTime: startTime,
		Metadata:  make(map[string]interface{}),
	}

	// Execute before hooks
	if c.hookManager != nil {
		if err := c.hookManager.ExecuteHooks(interfaces.BeforeExecHook, execCtx); err != nil {
			return 0, err
		}
	}

	// Create adapter for pgx.CopyFromSource
	copyFromSrc := &pgxCopyFromSourceAdapter{src: rowSrc}

	var rowsAffected int64
	var err error

	if c.conn != nil {
		rowsAffected, err = c.conn.CopyFrom(ctx, pgx.Identifier{tableName}, columnNames, copyFromSrc)
	} else if c.rawConn != nil {
		rowsAffected, err = c.rawConn.CopyFrom(ctx, pgx.Identifier{tableName}, columnNames, copyFromSrc)
	} else {
		err = ErrConnectionClosed
	}

	// Update statistics
	execCtx.Duration = time.Since(startTime)
	execCtx.RowsAffected = rowsAffected
	execCtx.Error = err

	c.updateStats("CopyFrom", time.Since(startTime), false)

	// Execute after hooks
	if c.hookManager != nil {
		c.hookManager.ExecuteHooks(interfaces.AfterExecHook, execCtx)
	}

	return rowsAffected, nil
}

// CopyTo implements IConn.CopyTo
func (c *PGXConn) CopyTo(ctx context.Context, w interfaces.CopyToWriter, query string, args ...interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrConnectionClosed
	}

	// For CopyTo, we typically use Query and then write the results
	rows, err := c.Query(ctx, "COPY ("+query+") TO STDOUT", args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		values := rows.(*PGXRows).rows.RawValues()
		interfaceValues := make([]interface{}, len(values))
		for i, v := range values {
			interfaceValues[i] = v
		}

		if err := w.Write(interfaceValues); err != nil {
			return err
		}
	}

	return w.Close()
}

// Listen implements IConn.Listen
func (c *PGXConn) Listen(ctx context.Context, channel string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrConnectionClosed
	}

	var err error
	if c.conn != nil {
		_, err = c.conn.Exec(ctx, "LISTEN "+channel)
	} else if c.rawConn != nil {
		_, err = c.rawConn.Exec(ctx, "LISTEN "+channel)
	} else {
		err = ErrConnectionClosed
	}

	return err
}

// Unlisten implements IConn.Unlisten
func (c *PGXConn) Unlisten(ctx context.Context, channel string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrConnectionClosed
	}

	var err error
	if c.conn != nil {
		_, err = c.conn.Exec(ctx, "UNLISTEN "+channel)
	} else if c.rawConn != nil {
		_, err = c.rawConn.Exec(ctx, "UNLISTEN "+channel)
	} else {
		err = ErrConnectionClosed
	}

	return err
}

// WaitForNotification implements IConn.WaitForNotification
func (c *PGXConn) WaitForNotification(ctx context.Context, timeout time.Duration) (*interfaces.Notification, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, ErrConnectionClosed
	}

	// Only raw connections support WaitForNotification
	if c.rawConn == nil {
		return nil, ErrUnsupportedFeature
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	notification, err := c.rawConn.WaitForNotification(timeoutCtx)
	if err != nil {
		return nil, err
	}

	return &interfaces.Notification{
		PID:     notification.PID,
		Channel: notification.Channel,
		Payload: notification.Payload,
	}, nil
}
