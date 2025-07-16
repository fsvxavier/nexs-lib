package pgx

import (
	"context"
	"sync"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
	"github.com/jackc/pgx/v5"
)

// PGXTransaction implements the ITransaction interface
type PGXTransaction struct {
	tx          pgx.Tx
	conn        *PGXConn
	committed   bool
	rolledback  bool
	mutex       sync.RWMutex
	hookManager interfaces.HookManager
	stats       *interfaces.ConnectionStats
	tenant      string
}

// QueryRow implements IConn.QueryRow
func (t *PGXTransaction) QueryRow(ctx context.Context, query string, args ...interface{}) interfaces.IRow {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if t.committed || t.rolledback {
		return &PGXRow{row: nil, conn: t.conn, query: query, args: args}
	}

	// Execute hooks
	execCtx := &interfaces.ExecutionContext{
		Context:   ctx,
		Operation: "QueryRow",
		Query:     query,
		Args:      args,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Execute before hooks
	if t.hookManager != nil {
		if err := t.hookManager.ExecuteHooks(interfaces.BeforeQueryHook, execCtx); err != nil {
			return &PGXRow{row: nil, conn: t.conn, query: query, args: args}
		}
	}

	row := t.tx.QueryRow(ctx, query, args...)

	// Update stats
	execCtx.Duration = time.Since(execCtx.StartTime)
	if t.stats != nil {
		t.stats.TotalQueries++
		t.stats.LastActivity = time.Now()
	}

	// Execute after hooks
	if t.hookManager != nil {
		t.hookManager.ExecuteHooks(interfaces.AfterQueryHook, execCtx)
	}

	return newPGXRow(row, t.conn, query, args)
}

// Query implements IConn.Query
func (t *PGXTransaction) Query(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if t.committed || t.rolledback {
		return nil, pgx.ErrTxClosed
	}

	// Execute hooks
	execCtx := &interfaces.ExecutionContext{
		Context:   ctx,
		Operation: "Query",
		Query:     query,
		Args:      args,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Execute before hooks
	if t.hookManager != nil {
		if err := t.hookManager.ExecuteHooks(interfaces.BeforeQueryHook, execCtx); err != nil {
			execCtx.Error = err
			return nil, err
		}
	}

	rows, err := t.tx.Query(ctx, query, args...)
	execCtx.Duration = time.Since(execCtx.StartTime)

	if err != nil {
		execCtx.Error = err
		if t.stats != nil {
			t.stats.FailedQueries++
		}

		return nil, err
	}

	// Update stats
	if t.stats != nil {
		t.stats.TotalQueries++
		t.stats.LastActivity = time.Now()
	}

	// Execute after hooks
	if t.hookManager != nil {
		t.hookManager.ExecuteHooks(interfaces.AfterQueryHook, execCtx)
	}

	return newPGXRows(rows, t.conn, query, args), nil
}

// QueryOne implements IConn.QueryOne
func (t *PGXTransaction) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	rows, err := t.Query(ctx, query, args...)
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

// QueryAll implements IConn.QueryAll
func (t *PGXTransaction) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	// This would need reflection to populate slice/array dst
	// For now, return not implemented
	return pgx.ErrNoRows
}

// QueryCount implements IConn.QueryCount
func (t *PGXTransaction) QueryCount(ctx context.Context, query string, args ...interface{}) (int64, error) {
	var count int64
	err := t.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

// Exec implements IConn.Exec
func (t *PGXTransaction) Exec(ctx context.Context, query string, args ...interface{}) (interfaces.CommandTag, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if t.committed || t.rolledback {
		return nil, pgx.ErrTxClosed
	}

	// Execute hooks and middlewares
	execCtx := &interfaces.ExecutionContext{
		Context:   ctx,
		Operation: "Exec",
		Query:     query,
		Args:      args,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Execute before hooks
	if t.hookManager != nil {
		if err := t.hookManager.ExecuteHooks(interfaces.BeforeExecHook, execCtx); err != nil {
			execCtx.Error = err
			return nil, err
		}
	}

	commandTag, err := t.tx.Exec(ctx, query, args...)
	execCtx.Duration = time.Since(execCtx.StartTime)

	if err != nil {
		execCtx.Error = err
		if t.stats != nil {
			t.stats.FailedExecs++
		}

		return nil, err
	}

	// Update stats
	execCtx.RowsAffected = commandTag.RowsAffected()
	if t.stats != nil {
		t.stats.TotalExecs++
		t.stats.LastActivity = time.Now()
	}

	// Execute after hooks
	if t.hookManager != nil {
		t.hookManager.ExecuteHooks(interfaces.AfterExecHook, execCtx)
	}

	return &PGXCommandTag{tag: commandTag}, nil
}

// SendBatch implements IConn.SendBatch
func (t *PGXTransaction) SendBatch(ctx context.Context, batch interfaces.IBatch) interfaces.IBatchResults {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if t.committed || t.rolledback {
		return newPGXBatchResults(nil, t.conn)
	}

	pgxBatch := batch.(*PGXBatch)
	results := t.tx.SendBatch(ctx, pgxBatch.batch)
	return newPGXBatchResults(results, t.conn)
}

// Begin implements IConn.Begin - nested transactions not supported, returns self
func (t *PGXTransaction) Begin(ctx context.Context) (interfaces.ITransaction, error) {
	return t, nil
}

// BeginTx implements IConn.BeginTx - nested transactions not supported, returns self
func (t *PGXTransaction) BeginTx(ctx context.Context, txOptions interfaces.TxOptions) (interfaces.ITransaction, error) {
	return t, nil
}

// Release implements IConn.Release - no-op for transactions
func (t *PGXTransaction) Release() {
	// No-op for transactions
}

// Close implements IConn.Close
func (t *PGXTransaction) Close(ctx context.Context) error {
	return t.Rollback(ctx)
}

// Ping implements IConn.Ping - no-op for transactions
func (t *PGXTransaction) Ping(ctx context.Context) error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if t.committed || t.rolledback {
		return pgx.ErrTxClosed
	}
	return nil
}

// IsClosed implements IConn.IsClosed
func (t *PGXTransaction) IsClosed() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.committed || t.rolledback
}

// Prepare implements IConn.Prepare
func (t *PGXTransaction) Prepare(ctx context.Context, name, query string) error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if t.committed || t.rolledback {
		return pgx.ErrTxClosed
	}

	_, err := t.tx.Prepare(ctx, name, query)
	return err
}

// Deallocate implements IConn.Deallocate
func (t *PGXTransaction) Deallocate(ctx context.Context, name string) error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if t.committed || t.rolledback {
		return pgx.ErrTxClosed
	}

	_, err := t.tx.Exec(ctx, "DEALLOCATE "+name)
	return err
}

// CopyFrom implements IConn.CopyFrom
func (t *PGXTransaction) CopyFrom(ctx context.Context, tableName string, columnNames []string, rowSrc interfaces.CopyFromSource) (int64, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if t.committed || t.rolledback {
		return 0, pgx.ErrTxClosed
	}

	copyFromSrc := &pgxCopyFromSourceAdapter{src: rowSrc}
	return t.tx.CopyFrom(ctx, pgx.Identifier{tableName}, columnNames, copyFromSrc)
}

// CopyTo implements IConn.CopyTo
func (t *PGXTransaction) CopyTo(ctx context.Context, w interfaces.CopyToWriter, query string, args ...interface{}) error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if t.committed || t.rolledback {
		return pgx.ErrTxClosed
	}

	// PGX doesn't directly support CopyTo in transactions, so we implement it differently
	rows, err := t.Query(ctx, "COPY ("+query+") TO STDOUT", args...)
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

// Listen implements IConn.Listen - not supported in transactions
func (t *PGXTransaction) Listen(ctx context.Context, channel string) error {
	return pgx.ErrTxClosed
}

// Unlisten implements IConn.Unlisten - not supported in transactions
func (t *PGXTransaction) Unlisten(ctx context.Context, channel string) error {
	return pgx.ErrTxClosed
}

// WaitForNotification implements IConn.WaitForNotification - not supported in transactions
func (t *PGXTransaction) WaitForNotification(ctx context.Context, timeout time.Duration) (*interfaces.Notification, error) {
	return nil, pgx.ErrTxClosed
}

// SetTenant implements IConn.SetTenant
func (t *PGXTransaction) SetTenant(ctx context.Context, tenantID string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.tenant = tenantID
	return nil
}

// GetTenant implements IConn.GetTenant
func (t *PGXTransaction) GetTenant(ctx context.Context) (string, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.tenant, nil
}

// GetHookManager implements IConn.GetHookManager
func (t *PGXTransaction) GetHookManager() interfaces.HookManager {
	return t.hookManager
}

// HealthCheck implements IConn.HealthCheck
func (t *PGXTransaction) HealthCheck(ctx context.Context) error {
	return t.Ping(ctx)
}

// Stats implements IConn.Stats
func (t *PGXTransaction) Stats() interfaces.ConnectionStats {
	if t.stats == nil {
		return interfaces.ConnectionStats{}
	}
	return *t.stats
}

// Commit implements ITransaction.Commit
func (t *PGXTransaction) Commit(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.committed {
		return nil
	}
	if t.rolledback {
		return pgx.ErrTxClosed
	}

	// Execute hooks and middlewares
	execCtx := &interfaces.ExecutionContext{
		Context:   ctx,
		Operation: "Commit",
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Execute before hooks
	if t.hookManager != nil {
		if err := t.hookManager.ExecuteHooks(interfaces.BeforeTransactionHook, execCtx); err != nil {
			execCtx.Error = err
			return err
		}
	}

	err := t.tx.Commit(ctx)
	execCtx.Duration = time.Since(execCtx.StartTime)

	if err != nil {
		execCtx.Error = err
		if t.stats != nil {
			t.stats.FailedTransactions++
		}

		return err
	}

	t.committed = true

	// Update stats
	if t.stats != nil {
		t.stats.TotalTransactions++
		t.stats.LastActivity = time.Now()
	}

	// Execute after hooks
	if t.hookManager != nil {
		t.hookManager.ExecuteHooks(interfaces.AfterTransactionHook, execCtx)
	}

	return nil
}

// Rollback implements ITransaction.Rollback
func (t *PGXTransaction) Rollback(ctx context.Context) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.rolledback {
		return nil
	}
	if t.committed {
		return pgx.ErrTxClosed
	}

	// Execute hooks and middlewares
	execCtx := &interfaces.ExecutionContext{
		Context:   ctx,
		Operation: "Rollback",
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Execute before hooks
	if t.hookManager != nil {
		if err := t.hookManager.ExecuteHooks(interfaces.BeforeTransactionHook, execCtx); err != nil {
			execCtx.Error = err
		}
	}

	err := t.tx.Rollback(ctx)
	execCtx.Duration = time.Since(execCtx.StartTime)

	if err != nil {
		execCtx.Error = err
		if t.stats != nil {
			t.stats.FailedTransactions++
		}

		return err
	}

	t.rolledback = true

	// Update stats
	if t.stats != nil {
		t.stats.LastActivity = time.Now()
	}

	// Execute after hooks
	if t.hookManager != nil {
		t.hookManager.ExecuteHooks(interfaces.AfterTransactionHook, execCtx)
	}

	return nil
}

// Helper function to create a new PGXTransaction
func newPGXTransaction(tx pgx.Tx, conn *PGXConn) *PGXTransaction {
	stats := &interfaces.ConnectionStats{
		CreatedAt: time.Now(),
	}

	return &PGXTransaction{
		tx:          tx,
		conn:        conn,
		hookManager: conn.hookManager,
		stats:       stats,
		tenant:      conn.tenantID,
	}
}

// pgxCopyFromSourceAdapter adapts interfaces.CopyFromSource to pgx.CopyFromSource
type pgxCopyFromSourceAdapter struct {
	src interfaces.CopyFromSource
}

func (a *pgxCopyFromSourceAdapter) Next() bool {
	return a.src.Next()
}

func (a *pgxCopyFromSourceAdapter) Values() ([]interface{}, error) {
	return a.src.Values()
}

func (a *pgxCopyFromSourceAdapter) Err() error {
	return a.src.Err()
}
