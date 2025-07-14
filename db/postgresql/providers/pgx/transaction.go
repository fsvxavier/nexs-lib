package pgx

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/jackc/pgx/v5"
)

// Transaction implements postgresql.ITransaction
type Transaction struct {
	tx                 pgx.Tx
	config             *config.Config
	multiTenantEnabled bool
	logger             *log.Logger
	startTime          time.Time
	completed          bool
}

// QueryOne executes a query and scans the result into dst
func (t *Transaction) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if t.completed {
		return fmt.Errorf("transaction is completed")
	}

	// Apply query timeout if configured
	if t.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, t.config.QueryTimeout)
		defer cancel()
	}

	// Execute before query hook
	if t.config.Hooks != nil && t.config.Hooks.BeforeQuery != nil {
		if err := t.config.Hooks.BeforeQuery(ctx, query, args); err != nil {
			return fmt.Errorf("before query hook failed: %w", err)
		}
	}

	start := time.Now()
	rows, err := t.tx.Query(ctx, query, args...)
	duration := time.Since(start)

	// Execute after query hook
	if t.config.Hooks != nil && t.config.Hooks.AfterQuery != nil {
		hookErr := t.config.Hooks.AfterQuery(ctx, query, args, duration, err)
		if hookErr != nil {
			t.logger.Printf("After query hook failed: %v", hookErr)
		}
	}

	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if rows.Err() != nil {
			return fmt.Errorf("rows error: %w", rows.Err())
		}
		return fmt.Errorf("no rows returned")
	}

	return rows.Scan(dst)
}

// QueryAll executes a query and scans all results into dst
func (t *Transaction) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if t.completed {
		return fmt.Errorf("transaction is completed")
	}

	// Apply query timeout if configured
	if t.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, t.config.QueryTimeout)
		defer cancel()
	}

	// Execute before query hook
	if t.config.Hooks != nil && t.config.Hooks.BeforeQuery != nil {
		if err := t.config.Hooks.BeforeQuery(ctx, query, args); err != nil {
			return fmt.Errorf("before query hook failed: %w", err)
		}
	}

	start := time.Now()
	rows, err := t.tx.Query(ctx, query, args...)
	duration := time.Since(start)

	// Execute after query hook
	if t.config.Hooks != nil && t.config.Hooks.AfterQuery != nil {
		hookErr := t.config.Hooks.AfterQuery(ctx, query, args, duration, err)
		if hookErr != nil {
			t.logger.Printf("After query hook failed: %v", hookErr)
		}
	}

	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// TODO: Implement proper scanning to dst slice/array
	// For now, this is a placeholder implementation
	return fmt.Errorf("QueryAll not fully implemented yet")
}

// QueryCount executes a query and returns the count
func (t *Transaction) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	if t.completed {
		return nil, fmt.Errorf("transaction is completed")
	}

	var count int
	err := t.QueryOne(ctx, &count, query, args...)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

// Query executes a query and returns the rows
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (postgresql.IRows, error) {
	if t.completed {
		return nil, fmt.Errorf("transaction is completed")
	}

	// Apply query timeout if configured
	if t.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, t.config.QueryTimeout)
		defer cancel()
	}

	// Execute before query hook
	if t.config.Hooks != nil && t.config.Hooks.BeforeQuery != nil {
		if err := t.config.Hooks.BeforeQuery(ctx, query, args); err != nil {
			return nil, fmt.Errorf("before query hook failed: %w", err)
		}
	}

	start := time.Now()
	rows, err := t.tx.Query(ctx, query, args...)
	duration := time.Since(start)

	// Execute after query hook
	if t.config.Hooks != nil && t.config.Hooks.AfterQuery != nil {
		hookErr := t.config.Hooks.AfterQuery(ctx, query, args, duration, err)
		if hookErr != nil {
			t.logger.Printf("After query hook failed: %v", hookErr)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return &Rows{rows: rows}, nil
}

// QueryRow executes a query and returns a single row
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) (postgresql.IRow, error) {
	if t.completed {
		return nil, fmt.Errorf("transaction is completed")
	}

	// Apply query timeout if configured
	if t.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, t.config.QueryTimeout)
		defer cancel()
	}

	// Execute before query hook
	if t.config.Hooks != nil && t.config.Hooks.BeforeQuery != nil {
		if err := t.config.Hooks.BeforeQuery(ctx, query, args); err != nil {
			return nil, fmt.Errorf("before query hook failed: %w", err)
		}
	}

	start := time.Now()
	row := t.tx.QueryRow(ctx, query, args...)
	duration := time.Since(start)

	// Execute after query hook
	if t.config.Hooks != nil && t.config.Hooks.AfterQuery != nil {
		hookErr := t.config.Hooks.AfterQuery(ctx, query, args, duration, nil)
		if hookErr != nil {
			t.logger.Printf("After query hook failed: %v", hookErr)
		}
	}

	return &Row{row: row}, nil
}

// Exec executes a query without returning rows
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) error {
	if t.completed {
		return fmt.Errorf("transaction is completed")
	}

	// Apply query timeout if configured
	if t.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, t.config.QueryTimeout)
		defer cancel()
	}

	// Execute before query hook
	if t.config.Hooks != nil && t.config.Hooks.BeforeQuery != nil {
		if err := t.config.Hooks.BeforeQuery(ctx, query, args); err != nil {
			return fmt.Errorf("before query hook failed: %w", err)
		}
	}

	start := time.Now()
	_, err := t.tx.Exec(ctx, query, args...)
	duration := time.Since(start)

	// Execute after query hook
	if t.config.Hooks != nil && t.config.Hooks.AfterQuery != nil {
		hookErr := t.config.Hooks.AfterQuery(ctx, query, args, duration, err)
		if hookErr != nil {
			t.logger.Printf("After query hook failed: %v", hookErr)
		}
	}

	if err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}

	return nil
}

// SendBatch sends a batch of queries
func (t *Transaction) SendBatch(ctx context.Context, batch postgresql.IBatch) (postgresql.IBatchResults, error) {
	if t.completed {
		return nil, fmt.Errorf("transaction is completed")
	}

	pgxBatch, ok := batch.(*Batch)
	if !ok {
		return nil, fmt.Errorf("invalid batch type")
	}

	// Apply query timeout if configured
	if t.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, t.config.QueryTimeout)
		defer cancel()
	}

	batchResults := t.tx.SendBatch(ctx, pgxBatch.batch)

	return &BatchResults{results: batchResults}, nil
}

// BeginTransaction is not supported within a transaction
func (t *Transaction) BeginTransaction(ctx context.Context) (postgresql.ITransaction, error) {
	return nil, fmt.Errorf("nested transactions are not supported")
}

// BeginTransactionWithOptions is not supported within a transaction
func (t *Transaction) BeginTransactionWithOptions(ctx context.Context, opts postgresql.TxOptions) (postgresql.ITransaction, error) {
	return nil, fmt.Errorf("nested transactions are not supported")
}

// Prepare prepares a statement within the transaction
func (t *Transaction) Prepare(ctx context.Context, name, query string) error {
	if t.completed {
		return fmt.Errorf("transaction is completed")
	}

	_, err := t.tx.Prepare(ctx, name, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	return nil
}

// BeforeReleaseHook is not applicable for transactions
func (t *Transaction) BeforeReleaseHook(ctx context.Context) error {
	// Not applicable for transactions
	return nil
}

// AfterAcquireHook is not applicable for transactions
func (t *Transaction) AfterAcquireHook(ctx context.Context) error {
	// Not applicable for transactions
	return nil
}

// Release is not applicable for transactions (use Commit or Rollback)
func (t *Transaction) Release(ctx context.Context) {
	// Not applicable for transactions
	// Transactions should be explicitly committed or rolled back
}

// Ping checks if the transaction is still valid
func (t *Transaction) Ping(ctx context.Context) error {
	if t.completed {
		return fmt.Errorf("transaction is completed")
	}

	// For transactions, we check if we can execute a simple query
	_, err := t.tx.Exec(ctx, "SELECT 1")
	return err
}

// Listen is not supported in transactions
func (t *Transaction) Listen(ctx context.Context, channel string) error {
	return fmt.Errorf("LISTEN is not supported in transactions")
}

// Unlisten is not supported in transactions
func (t *Transaction) Unlisten(ctx context.Context, channel string) error {
	return fmt.Errorf("UNLISTEN is not supported in transactions")
}

// WaitForNotification is not supported in transactions
func (t *Transaction) WaitForNotification(ctx context.Context, timeout time.Duration) (*postgresql.Notification, error) {
	return nil, fmt.Errorf("WaitForNotification is not supported in transactions")
}

// Commit commits the transaction
func (t *Transaction) Commit(ctx context.Context) error {
	if t.completed {
		return fmt.Errorf("transaction is already completed")
	}

	t.completed = true
	duration := time.Since(t.startTime)

	err := t.tx.Commit(ctx)

	// Execute after transaction hook
	if t.config.Hooks != nil && t.config.Hooks.AfterTransaction != nil {
		hookErr := t.config.Hooks.AfterTransaction(ctx, err == nil, duration, err)
		if hookErr != nil {
			t.logger.Printf("After transaction hook failed: %v", hookErr)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback(ctx context.Context) error {
	if t.completed {
		return fmt.Errorf("transaction is already completed")
	}

	t.completed = true
	duration := time.Since(t.startTime)

	err := t.tx.Rollback(ctx)

	// Execute after transaction hook
	if t.config.Hooks != nil && t.config.Hooks.AfterTransaction != nil {
		hookErr := t.config.Hooks.AfterTransaction(ctx, false, duration, err)
		if hookErr != nil {
			t.logger.Printf("After transaction hook failed: %v", hookErr)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	return nil
}

// Savepoint creates a savepoint
func (t *Transaction) Savepoint(ctx context.Context, name string) error {
	if t.completed {
		return fmt.Errorf("transaction is completed")
	}

	query := fmt.Sprintf("SAVEPOINT %s", name)
	return t.Exec(ctx, query)
}

// RollbackToSavepoint rolls back to a savepoint
func (t *Transaction) RollbackToSavepoint(ctx context.Context, name string) error {
	if t.completed {
		return fmt.Errorf("transaction is completed")
	}

	query := fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name)
	return t.Exec(ctx, query)
}

// ReleaseSavepoint releases a savepoint
func (t *Transaction) ReleaseSavepoint(ctx context.Context, name string) error {
	if t.completed {
		return fmt.Errorf("transaction is completed")
	}

	query := fmt.Sprintf("RELEASE SAVEPOINT %s", name)
	return t.Exec(ctx, query)
}
