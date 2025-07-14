package pq

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/lib/pq"
)

// Conn represents a lib/pq database connection
type Conn struct {
	db         *sql.DB
	config     *config.Config
	released   bool
	isFromPool bool
}

// QueryOne executes a query and scans one result
func (c *Conn) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}

	row := c.db.QueryRowContext(ctx, query, args...)
	return row.Scan(dst)
}

// QueryAll executes a query and scans all results
func (c *Conn) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// For lib/pq, we need to handle scanning manually
	// This is a simplified implementation
	return fmt.Errorf("QueryAll not implemented for lib/pq - use Query() instead")
}

// QueryCount executes a count query
func (c *Conn) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}

	var count int
	row := c.db.QueryRowContext(ctx, query, args...)
	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

// Query executes a query and returns rows
func (c *Conn) Query(ctx context.Context, query string, args ...interface{}) (postgresql.IRows, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return &Rows{rows: rows}, nil
}

// QueryRow executes a query and returns a single row
func (c *Conn) QueryRow(ctx context.Context, query string, args ...interface{}) (postgresql.IRow, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}

	row := c.db.QueryRowContext(ctx, query, args...)
	return &Row{row: row}, nil
}

// Exec executes a command
func (c *Conn) Exec(ctx context.Context, query string, args ...interface{}) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}

	_, err := c.db.ExecContext(ctx, query, args...)
	return err
}

// SendBatch executes a batch - lib/pq doesn't support traditional batches like pgx
func (c *Conn) SendBatch(ctx context.Context, batch postgresql.IBatch) (postgresql.IBatchResults, error) {
	return nil, fmt.Errorf("lib/pq doesn't support batch operations like pgx")
}

// BeginTransaction starts a new transaction
func (c *Conn) BeginTransaction(ctx context.Context) (postgresql.ITransaction, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &Transaction{
		tx:     tx,
		config: c.config,
	}, nil
}

// BeginTransactionWithOptions starts a transaction with options
func (c *Conn) BeginTransactionWithOptions(ctx context.Context, opts postgresql.TxOptions) (postgresql.ITransaction, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}

	// Convert TxOptions to sql.TxOptions
	sqlOpts := &sql.TxOptions{
		ReadOnly: opts.ReadOnly,
	}

	// Convert isolation level
	switch opts.IsoLevel {
	case postgresql.IsoLevelReadUncommitted:
		sqlOpts.Isolation = sql.LevelReadUncommitted
	case postgresql.IsoLevelReadCommitted:
		sqlOpts.Isolation = sql.LevelReadCommitted
	case postgresql.IsoLevelRepeatableRead:
		sqlOpts.Isolation = sql.LevelRepeatableRead
	case postgresql.IsoLevelSerializable:
		sqlOpts.Isolation = sql.LevelSerializable
	default:
		sqlOpts.Isolation = sql.LevelDefault
	}

	tx, err := c.db.BeginTx(ctx, sqlOpts)
	if err != nil {
		return nil, err
	}

	return &Transaction{
		tx:     tx,
		config: c.config,
	}, nil
}

// Prepare prepares a statement
func (c *Conn) Prepare(ctx context.Context, name, query string) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}

	stmt, err := c.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	// Close immediately since we don't have a way to store it by name
	stmt.Close()
	return nil
}

// BeforeReleaseHook executes before releasing the connection
func (c *Conn) BeforeReleaseHook(ctx context.Context) error {
	return nil
}

// AfterAcquireHook executes after acquiring the connection
func (c *Conn) AfterAcquireHook(ctx context.Context) error {
	return nil
}

// Release releases the connection
func (c *Conn) Release(ctx context.Context) {
	if c.released {
		return
	}

	c.released = true
	// For lib/pq, if not from pool, close the connection
	if !c.isFromPool && c.db != nil {
		c.db.Close()
	}
}

// Ping tests the connection
func (c *Conn) Ping(ctx context.Context) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}

	return c.db.PingContext(ctx)
}

// Listen starts listening on a channel (PostgreSQL LISTEN)
func (c *Conn) Listen(ctx context.Context, channel string) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}

	// lib/pq supports LISTEN/NOTIFY
	_, err := c.db.ExecContext(ctx, "LISTEN "+pq.QuoteIdentifier(channel))
	return err
}

// Unlisten stops listening on a channel (PostgreSQL UNLISTEN)
func (c *Conn) Unlisten(ctx context.Context, channel string) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}

	_, err := c.db.ExecContext(ctx, "UNLISTEN "+pq.QuoteIdentifier(channel))
	return err
}

// WaitForNotification waits for a notification
func (c *Conn) WaitForNotification(ctx context.Context, timeout time.Duration) (*postgresql.Notification, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}

	// lib/pq has limited support for notifications
	// This is a simplified implementation
	return nil, fmt.Errorf("WaitForNotification not fully implemented for lib/pq")
}
