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

// Rows represents query results for GORM
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

// Columns returns column names
func (r *Rows) Columns() ([]string, error) {
	if r.rows == nil {
		return nil, fmt.Errorf("rows is nil")
	}
	return r.rows.Columns()
}

// ColumnTypes returns column types
func (r *Rows) ColumnTypes() ([]*sql.ColumnType, error) {
	if r.rows == nil {
		return nil, fmt.Errorf("rows is nil")
	}
	return r.rows.ColumnTypes()
}

// RawValues returns raw byte values
func (r *Rows) RawValues() [][]byte {
	// GORM/sql.Rows doesn't directly support raw values
	// This is a limitation of the GORM adapter
	return nil
}

// Values returns current row values
func (r *Rows) Values() ([]interface{}, error) {
	if r.rows == nil {
		return nil, fmt.Errorf("rows is nil")
	}

	cols, err := r.rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	err = r.rows.Scan(valuePtrs...)
	if err != nil {
		return nil, err
	}

	return values, nil
}

// Row represents a single row result for GORM
type Row struct {
	db    *gorm.DB
	query string
	args  []interface{}
}

// Scan scans the row into destination
func (r *Row) Scan(dest ...interface{}) error {
	if r.db == nil {
		return fmt.Errorf("db is nil")
	}

	row := r.db.Raw(r.query, r.args...).Row()
	return row.Scan(dest...)
}

// Err returns any error from the row
func (r *Row) Err() error {
	if r.db == nil {
		return fmt.Errorf("db is nil")
	}

	// Execute the query to check for errors
	result := r.db.Raw(r.query, r.args...)
	return result.Error
}

// Transaction represents a GORM transaction
type Transaction struct {
	tx     *gorm.DB
	config *config.Config
}

// QueryOne executes a query and scans one result within the transaction
func (t *Transaction) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	result := t.tx.WithContext(ctx).Raw(query, args...).First(dst)
	return result.Error
}

// QueryAll executes a query and scans all results within the transaction
func (t *Transaction) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	result := t.tx.WithContext(ctx).Raw(query, args...).Find(dst)
	return result.Error
}

// QueryCount executes a count query within the transaction
func (t *Transaction) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	var count int64
	result := t.tx.WithContext(ctx).Raw(query, args...).Count(&count)
	if result.Error != nil {
		return nil, result.Error
	}

	intCount := int(count)
	return &intCount, nil
}

// Query executes a query and returns rows within the transaction
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (postgresql.IRows, error) {
	rows, err := t.tx.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}

	return &Rows{rows: rows}, nil
}

// QueryRow executes a query and returns a single row within the transaction
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) (postgresql.IRow, error) {
	return &Row{
		db:    t.tx.WithContext(ctx),
		query: query,
		args:  args,
	}, nil
}

// Exec executes a command within the transaction
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) error {
	result := t.tx.WithContext(ctx).Exec(query, args...)
	return result.Error
}

// SendBatch executes a batch within the transaction - GORM doesn't support traditional batches
func (t *Transaction) SendBatch(ctx context.Context, batch postgresql.IBatch) (postgresql.IBatchResults, error) {
	return nil, fmt.Errorf("GORM doesn't support batch operations")
}

// Commit commits the transaction
func (t *Transaction) Commit(ctx context.Context) error {
	result := t.tx.Commit()
	return result.Error
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback(ctx context.Context) error {
	result := t.tx.Rollback()
	return result.Error
}

// BeginSavepoint creates a savepoint - GORM supports this through SavePoint
func (t *Transaction) BeginSavepoint(ctx context.Context, name string) error {
	result := t.tx.WithContext(ctx).SavePoint(name)
	return result.Error
}

// ReleaseSavepoint releases a savepoint
func (t *Transaction) ReleaseSavepoint(ctx context.Context, name string) error {
	// GORM doesn't have direct ReleaseSavepoint, but we can execute raw SQL
	result := t.tx.WithContext(ctx).Exec("RELEASE SAVEPOINT " + name)
	return result.Error
}

// BeginTransaction starts a nested transaction (savepoint)
func (t *Transaction) BeginTransaction(ctx context.Context) (postgresql.ITransaction, error) {
	// Create a savepoint for nested transaction
	savepointName := fmt.Sprintf("sp_%d", time.Now().UnixNano())
	err := t.BeginSavepoint(ctx, savepointName)
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
	// GORM doesn't support transaction options directly, fall back to basic transaction
	return t.BeginTransaction(ctx)
}

// Prepare prepares a statement within the transaction
func (t *Transaction) Prepare(ctx context.Context, name, query string) error {
	// GORM handles statement preparation internally
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
	sqlDB, err := t.tx.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// Listen - GORM doesn't support LISTEN/NOTIFY
func (t *Transaction) Listen(ctx context.Context, channel string) error {
	return fmt.Errorf("GORM doesn't support LISTEN/NOTIFY")
}

// Unlisten - GORM doesn't support LISTEN/NOTIFY
func (t *Transaction) Unlisten(ctx context.Context, channel string) error {
	return fmt.Errorf("GORM doesn't support LISTEN/NOTIFY")
}

// WaitForNotification - GORM doesn't support LISTEN/NOTIFY
func (t *Transaction) WaitForNotification(ctx context.Context, timeout time.Duration) (*postgresql.Notification, error) {
	return nil, fmt.Errorf("GORM doesn't support LISTEN/NOTIFY")
}

// Savepoint creates a savepoint with the given name
func (t *Transaction) Savepoint(ctx context.Context, name string) error {
	return t.BeginSavepoint(ctx, name)
}

// RollbackToSavepoint rolls back to a savepoint
func (t *Transaction) RollbackToSavepoint(ctx context.Context, name string) error {
	result := t.tx.WithContext(ctx).RollbackTo(name)
	return result.Error
}
