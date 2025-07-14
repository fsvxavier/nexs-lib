package pgx

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Conn implements postgresql.IConn for PGX driver
type Conn struct {
	poolConn           *pgxpool.Conn // Only set if connection is from pool
	conn               *pgx.Conn     // The underlying PGX connection
	config             *config.Config
	multiTenantEnabled bool
	logger             *log.Logger
	isPooled           bool
	released           bool
}

// QueryOne executes a query and scans the result into dst
func (c *Conn) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}

	// Apply query timeout if configured
	if c.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.QueryTimeout)
		defer cancel()
	}

	// Execute before query hook
	if c.config.Hooks != nil && c.config.Hooks.BeforeQuery != nil {
		if err := c.config.Hooks.BeforeQuery(ctx, query, args); err != nil {
			return fmt.Errorf("before query hook failed: %w", err)
		}
	}

	start := time.Now()
	rows, err := c.conn.Query(ctx, query, args...)
	duration := time.Since(start)

	// Execute after query hook
	if c.config.Hooks != nil && c.config.Hooks.AfterQuery != nil {
		hookErr := c.config.Hooks.AfterQuery(ctx, query, args, duration, err)
		if hookErr != nil {
			c.logger.Printf("After query hook failed: %v", hookErr)
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
func (c *Conn) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}

	// Apply query timeout if configured
	if c.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.QueryTimeout)
		defer cancel()
	}

	// Execute before query hook
	if c.config.Hooks != nil && c.config.Hooks.BeforeQuery != nil {
		if err := c.config.Hooks.BeforeQuery(ctx, query, args); err != nil {
			return fmt.Errorf("before query hook failed: %w", err)
		}
	}

	start := time.Now()
	rows, err := c.conn.Query(ctx, query, args...)
	duration := time.Since(start)

	// Execute after query hook
	if c.config.Hooks != nil && c.config.Hooks.AfterQuery != nil {
		hookErr := c.config.Hooks.AfterQuery(ctx, query, args, duration, err)
		if hookErr != nil {
			c.logger.Printf("After query hook failed: %v", hookErr)
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
func (c *Conn) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}

	var count int
	err := c.QueryOne(ctx, &count, query, args...)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

// Query executes a query and returns the rows
func (c *Conn) Query(ctx context.Context, query string, args ...interface{}) (postgresql.IRows, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}

	// Apply query timeout if configured
	if c.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.QueryTimeout)
		defer cancel()
	}

	// Execute before query hook
	if c.config.Hooks != nil && c.config.Hooks.BeforeQuery != nil {
		if err := c.config.Hooks.BeforeQuery(ctx, query, args); err != nil {
			return nil, fmt.Errorf("before query hook failed: %w", err)
		}
	}

	start := time.Now()
	rows, err := c.conn.Query(ctx, query, args...)
	duration := time.Since(start)

	// Execute after query hook
	if c.config.Hooks != nil && c.config.Hooks.AfterQuery != nil {
		hookErr := c.config.Hooks.AfterQuery(ctx, query, args, duration, err)
		if hookErr != nil {
			c.logger.Printf("After query hook failed: %v", hookErr)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return &Rows{rows: rows}, nil
}

// QueryRow executes a query and returns a single row
func (c *Conn) QueryRow(ctx context.Context, query string, args ...interface{}) (postgresql.IRow, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}

	// Apply query timeout if configured
	if c.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.QueryTimeout)
		defer cancel()
	}

	// Execute before query hook
	if c.config.Hooks != nil && c.config.Hooks.BeforeQuery != nil {
		if err := c.config.Hooks.BeforeQuery(ctx, query, args); err != nil {
			return nil, fmt.Errorf("before query hook failed: %w", err)
		}
	}

	start := time.Now()
	row := c.conn.QueryRow(ctx, query, args...)
	duration := time.Since(start)

	// Execute after query hook
	if c.config.Hooks != nil && c.config.Hooks.AfterQuery != nil {
		hookErr := c.config.Hooks.AfterQuery(ctx, query, args, duration, nil)
		if hookErr != nil {
			c.logger.Printf("After query hook failed: %v", hookErr)
		}
	}

	return &Row{row: row}, nil
}

// Exec executes a query without returning rows
func (c *Conn) Exec(ctx context.Context, query string, args ...interface{}) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}

	// Apply query timeout if configured
	if c.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.QueryTimeout)
		defer cancel()
	}

	// Execute before query hook
	if c.config.Hooks != nil && c.config.Hooks.BeforeQuery != nil {
		if err := c.config.Hooks.BeforeQuery(ctx, query, args); err != nil {
			return fmt.Errorf("before query hook failed: %w", err)
		}
	}

	start := time.Now()
	_, err := c.conn.Exec(ctx, query, args...)
	duration := time.Since(start)

	// Execute after query hook
	if c.config.Hooks != nil && c.config.Hooks.AfterQuery != nil {
		hookErr := c.config.Hooks.AfterQuery(ctx, query, args, duration, err)
		if hookErr != nil {
			c.logger.Printf("After query hook failed: %v", hookErr)
		}
	}

	if err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}

	return nil
}

// SendBatch sends a batch of queries
func (c *Conn) SendBatch(ctx context.Context, batch postgresql.IBatch) (postgresql.IBatchResults, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}

	pgxBatch, ok := batch.(*Batch)
	if !ok {
		return nil, fmt.Errorf("invalid batch type")
	}

	// Apply query timeout if configured
	if c.config.QueryTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.config.QueryTimeout)
		defer cancel()
	}

	batchResults := c.conn.SendBatch(ctx, pgxBatch.batch)

	return &BatchResults{results: batchResults}, nil
}

// BeginTransaction begins a new transaction
func (c *Conn) BeginTransaction(ctx context.Context) (postgresql.ITransaction, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}

	// Execute before transaction hook
	if c.config.Hooks != nil && c.config.Hooks.BeforeTransaction != nil {
		if err := c.config.Hooks.BeforeTransaction(ctx); err != nil {
			return nil, fmt.Errorf("before transaction hook failed: %w", err)
		}
	}

	tx, err := c.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &Transaction{
		tx:                 tx,
		config:             c.config,
		multiTenantEnabled: c.multiTenantEnabled,
		logger:             c.logger,
		startTime:          time.Now(),
	}, nil
}

// BeginTransactionWithOptions begins a new transaction with options
func (c *Conn) BeginTransactionWithOptions(ctx context.Context, opts postgresql.TxOptions) (postgresql.ITransaction, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}

	// Convert our options to PGX options
	pgxOpts := c.convertTxOptions(opts)

	// Execute before transaction hook
	if c.config.Hooks != nil && c.config.Hooks.BeforeTransaction != nil {
		if err := c.config.Hooks.BeforeTransaction(ctx); err != nil {
			return nil, fmt.Errorf("before transaction hook failed: %w", err)
		}
	}

	tx, err := c.conn.BeginTx(ctx, pgxOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &Transaction{
		tx:                 tx,
		config:             c.config,
		multiTenantEnabled: c.multiTenantEnabled,
		logger:             c.logger,
		startTime:          time.Now(),
	}, nil
}

// Prepare prepares a statement
func (c *Conn) Prepare(ctx context.Context, name, query string) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}

	_, err := c.conn.Prepare(ctx, name, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	return nil
}

// BeforeReleaseHook executes before connection release
func (c *Conn) BeforeReleaseHook(ctx context.Context) error {
	if c.multiTenantEnabled {
		// Reset any tenant-specific settings
		if err := c.Exec(ctx, "RESET search_path"); err != nil {
			c.logger.Printf("Failed to reset search_path: %v", err)
		}

		if err := c.Exec(ctx, "RESET role"); err != nil {
			c.logger.Printf("Failed to reset role: %v", err)
		}
	}

	// Execute custom hook if configured
	if c.config.Hooks != nil && c.config.Hooks.BeforeRelease != nil {
		if err := c.config.Hooks.BeforeRelease(ctx, c); err != nil {
			return fmt.Errorf("custom before release hook failed: %w", err)
		}
	}

	return nil
}

// AfterAcquireHook executes after connection acquire
func (c *Conn) AfterAcquireHook(ctx context.Context) error {
	if c.multiTenantEnabled {
		// Set default schema if configured
		if c.config.DefaultSchema != "" && c.config.DefaultSchema != "public" {
			query := fmt.Sprintf("SET search_path TO %s", c.config.DefaultSchema)
			if err := c.Exec(ctx, query); err != nil {
				return fmt.Errorf("failed to set search_path: %w", err)
			}
		}
	}

	return nil
}

// Release releases the connection back to the pool or closes it
func (c *Conn) Release(ctx context.Context) {
	if c.released {
		return
	}

	c.released = true

	// Execute before release hook
	if err := c.BeforeReleaseHook(ctx); err != nil {
		c.logger.Printf("Before release hook failed: %v", err)
	}

	if c.isPooled && c.poolConn != nil {
		c.poolConn.Release()
	} else if !c.isPooled && c.conn != nil {
		c.conn.Close(ctx)
	}
}

// Ping checks if the connection is alive
func (c *Conn) Ping(ctx context.Context) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}

	return c.conn.Ping(ctx)
}

// Listen starts listening to a PostgreSQL notification channel
func (c *Conn) Listen(ctx context.Context, channel string) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}

	query := fmt.Sprintf("LISTEN %s", channel)
	return c.Exec(ctx, query)
}

// Unlisten stops listening to a PostgreSQL notification channel
func (c *Conn) Unlisten(ctx context.Context, channel string) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}

	query := fmt.Sprintf("UNLISTEN %s", channel)
	return c.Exec(ctx, query)
}

// WaitForNotification waits for a PostgreSQL notification
func (c *Conn) WaitForNotification(ctx context.Context, timeout time.Duration) (*postgresql.Notification, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	notification, err := c.conn.WaitForNotification(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for notification: %w", err)
	}

	return &postgresql.Notification{
		PID:     notification.PID,
		Channel: notification.Channel,
		Payload: notification.Payload,
	}, nil
}

// convertTxOptions converts our transaction options to PGX options
func (c *Conn) convertTxOptions(opts postgresql.TxOptions) pgx.TxOptions {
	pgxOpts := pgx.TxOptions{}

	// Isolation level
	switch opts.IsoLevel {
	case postgresql.IsoLevelReadUncommitted:
		pgxOpts.IsoLevel = pgx.ReadUncommitted
	case postgresql.IsoLevelReadCommitted:
		pgxOpts.IsoLevel = pgx.ReadCommitted
	case postgresql.IsoLevelRepeatableRead:
		pgxOpts.IsoLevel = pgx.RepeatableRead
	case postgresql.IsoLevelSerializable:
		pgxOpts.IsoLevel = pgx.Serializable
	default:
		pgxOpts.IsoLevel = pgx.ReadCommitted
	}

	// Access mode
	switch opts.AccessMode {
	case postgresql.AccessModeReadOnly:
		pgxOpts.AccessMode = pgx.ReadOnly
	case postgresql.AccessModeReadWrite:
		pgxOpts.AccessMode = pgx.ReadWrite
	default:
		pgxOpts.AccessMode = pgx.ReadWrite
	}

	// Deferrable mode
	switch opts.DeferrableMode {
	case postgresql.DeferrableModeDeferrable:
		pgxOpts.DeferrableMode = pgx.Deferrable
	case postgresql.DeferrableModeNotDeferrable:
		pgxOpts.DeferrableMode = pgx.NotDeferrable
	default:
		pgxOpts.DeferrableMode = pgx.NotDeferrable
	}

	return pgxOpts
}

// Helper function to check if TLS is enabled from environment
func isTLSEnabled() bool {
	enabled, _ := strconv.ParseBool(os.Getenv("DB_TLS_ENABLED"))
	return enabled
}
