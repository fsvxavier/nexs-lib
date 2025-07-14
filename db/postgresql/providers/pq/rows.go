package pq

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
)

// Rows represents query results for lib/pq
type Rows struct {
	rows *sql.Rows
}

// Next moves to the next row
func (r *Rows) Next() bool {
	if r.rows == nil {
		return false
	}
	return r.rows.Next()
}

// Scan scans the current row
func (r *Rows) Scan(dest ...interface{}) error {
	if r.rows == nil {
		return fmt.Errorf("rows is nil")
	}
	return r.rows.Scan(dest...)
}

// Close closes the rows
func (r *Rows) Close() {
	if r.rows != nil {
		r.rows.Close()
	}
}

// Err returns any error encountered during iteration
func (r *Rows) Err() error {
	if r.rows == nil {
		return nil
	}
	return r.rows.Err()
}

// RawValues returns raw byte values - lib/pq doesn't directly support this
func (r *Rows) RawValues() [][]byte {
	// lib/pq doesn't have direct raw values support like pgx
	return nil
}

// Row represents a single row result for lib/pq
type Row struct {
	row *sql.Row
}

// Scan scans the row into destination
func (r *Row) Scan(dest ...interface{}) error {
	if r.row == nil {
		return fmt.Errorf("row is nil")
	}
	return r.row.Scan(dest...)
}

// Transaction represents a lib/pq transaction
type Transaction struct {
	tx     *sql.Tx
	config *config.Config
}

// QueryOne executes a query and scans one result within the transaction
func (t *Transaction) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	row := t.tx.QueryRowContext(ctx, query, args...)
	return row.Scan(dst)
}

// QueryAll executes a query and scans all results within the transaction
func (t *Transaction) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Simplified implementation - would need reflection for proper scanning
	return fmt.Errorf("QueryAll not implemented for lib/pq transactions - use Query() instead")
}

// QueryCount executes a count query within the transaction
func (t *Transaction) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	var count int
	row := t.tx.QueryRowContext(ctx, query, args...)
	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}
	return &count, nil
}

// Query executes a query and returns rows within the transaction
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (postgresql.IRows, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows: rows}, nil
}

// QueryRow executes a query and returns a single row within the transaction
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) (postgresql.IRow, error) {
	row := t.tx.QueryRowContext(ctx, query, args...)
	return &Row{row: row}, nil
}

// Exec executes a command within the transaction
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := t.tx.ExecContext(ctx, query, args...)
	return err
}

// SendBatch executes a batch within the transaction - lib/pq doesn't support batches
func (t *Transaction) SendBatch(ctx context.Context, batch postgresql.IBatch) (postgresql.IBatchResults, error) {
	return nil, fmt.Errorf("lib/pq doesn't support batch operations")
}

// BeginTransaction starts a nested transaction (savepoint)
func (t *Transaction) BeginTransaction(ctx context.Context) (postgresql.ITransaction, error) {
	// Create a savepoint for nested transaction
	savepointName := fmt.Sprintf("sp_%d", len("savepoint"))
	_, err := t.tx.ExecContext(ctx, "SAVEPOINT "+savepointName)
	if err != nil {
		return nil, err
	}

	return &Transaction{
		tx:     t.tx,
		config: t.config,
	}, nil
}

// BeginTransactionWithOptions starts a transaction with options
func (t *Transaction) BeginTransactionWithOptions(ctx context.Context, opts postgresql.TxOptions) (postgresql.ITransaction, error) {
	// lib/pq doesn't support changing transaction options mid-transaction
	return t.BeginTransaction(ctx)
}

// Prepare prepares a statement within the transaction
func (t *Transaction) Prepare(ctx context.Context, name, query string) error {
	stmt, err := t.tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	stmt.Close() // Close immediately since we don't store it
	return nil
}

// BeforeReleaseHook executes before releasing the transaction
func (t *Transaction) BeforeReleaseHook(ctx context.Context) error {
	return nil
}

// AfterAcquireHook executes after acquiring the transaction
func (t *Transaction) AfterAcquireHook(ctx context.Context) error {
	return nil
}

// Release releases the transaction (commits if not already committed/rolled back)
func (t *Transaction) Release(ctx context.Context) {
	// In a transaction context, Release doesn't do anything
	// The transaction must be explicitly committed or rolled back
}

// Ping tests the transaction connection
func (t *Transaction) Ping(ctx context.Context) error {
	// Transactions don't support ping directly in lib/pq
	return nil
}

// Listen - not supported in transactions
func (t *Transaction) Listen(ctx context.Context, channel string) error {
	return fmt.Errorf("LISTEN not supported in transactions")
}

// Unlisten - not supported in transactions
func (t *Transaction) Unlisten(ctx context.Context, channel string) error {
	return fmt.Errorf("UNLISTEN not supported in transactions")
}

// WaitForNotification - not supported in transactions
func (t *Transaction) WaitForNotification(ctx context.Context, timeout time.Duration) (*postgresql.Notification, error) {
	return nil, fmt.Errorf("WaitForNotification not supported in transactions")
}

// Commit commits the transaction
func (t *Transaction) Commit(ctx context.Context) error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback()
}

// Savepoint creates a savepoint with the given name
func (t *Transaction) Savepoint(ctx context.Context, name string) error {
	_, err := t.tx.ExecContext(ctx, "SAVEPOINT "+name)
	return err
}

// RollbackToSavepoint rolls back to a savepoint
func (t *Transaction) RollbackToSavepoint(ctx context.Context, name string) error {
	_, err := t.tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+name)
	return err
}

// ReleaseSavepoint releases a savepoint
func (t *Transaction) ReleaseSavepoint(ctx context.Context, name string) error {
	_, err := t.tx.ExecContext(ctx, "RELEASE SAVEPOINT "+name)
	return err
}
