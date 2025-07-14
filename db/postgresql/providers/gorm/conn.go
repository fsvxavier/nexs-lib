package gorm

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"gorm.io/gorm"
)

// Conn represents a GORM database connection
type Conn struct {
	db         *gorm.DB
	config     *config.Config
	released   bool
	isFromPool bool
}

// QueryOne executes a query and scans one result
func (c *Conn) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}
	if c.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	result := c.db.WithContext(ctx).Raw(query, args...).First(dst)
	return result.Error
}

// QueryAll executes a query and scans all results
func (c *Conn) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}
	if c.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	result := c.db.WithContext(ctx).Raw(query, args...).Find(dst)
	return result.Error
}

// QueryCount executes a count query
func (c *Conn) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}
	if c.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	var count int64
	result := c.db.WithContext(ctx).Raw(query, args...).Count(&count)
	if result.Error != nil {
		return nil, result.Error
	}

	intCount := int(count)
	return &intCount, nil
}

// Query executes a query and returns rows
func (c *Conn) Query(ctx context.Context, query string, args ...interface{}) (postgresql.IRows, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}
	if c.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	rows, err := c.db.WithContext(ctx).Raw(query, args...).Rows()
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

	// GORM doesn't have direct QueryRow, so we use Raw with First
	return &Row{
		db:    c.db,
		query: query,
		args:  args,
	}, nil
}

// Exec executes a command
func (c *Conn) Exec(ctx context.Context, query string, args ...interface{}) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}
	if c.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	result := c.db.WithContext(ctx).Exec(query, args...)
	return result.Error
}

// SendBatch executes a batch - GORM doesn't support traditional batches
func (c *Conn) SendBatch(ctx context.Context, batch postgresql.IBatch) (postgresql.IBatchResults, error) {
	return nil, fmt.Errorf("GORM doesn't support batch operations")
}

// BeginTransaction starts a new transaction
func (c *Conn) BeginTransaction(ctx context.Context) (postgresql.ITransaction, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}
	if c.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &Transaction{
		tx:     tx,
		config: c.config,
	}, nil
}

// BeginTransactionWithOptions starts a transaction with options
func (c *Conn) BeginTransactionWithOptions(ctx context.Context, opts postgresql.TxOptions) (postgresql.ITransaction, error) {
	// GORM doesn't support transaction options like isolation levels directly
	// We'll start a basic transaction for now
	return c.BeginTransaction(ctx)
}

// Prepare prepares a statement - GORM handles this internally
func (c *Conn) Prepare(ctx context.Context, name, query string) error {
	// GORM handles statement preparation internally
	return nil
}

// BeforeReleaseHook executes before releasing the connection
func (c *Conn) BeforeReleaseHook(ctx context.Context) error {
	if c.config.Hooks != nil {
		// Execute hooks if they exist
	}
	return nil
}

// AfterAcquireHook executes after acquiring the connection
func (c *Conn) AfterAcquireHook(ctx context.Context) error {
	if c.config.Hooks != nil {
		// Execute hooks if they exist
	}
	return nil
}

// Release releases the connection back to the pool
func (c *Conn) Release(ctx context.Context) {
	if c.released {
		return
	}

	c.released = true
	// GORM connections don't need explicit release
}

// Ping tests the connection
func (c *Conn) Ping(ctx context.Context) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}
	if c.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.PingContext(ctx)
}

// Listen - GORM doesn't support LISTEN/NOTIFY
func (c *Conn) Listen(ctx context.Context, channel string) error {
	return fmt.Errorf("GORM doesn't support LISTEN/NOTIFY")
}

// Unlisten - GORM doesn't support LISTEN/NOTIFY
func (c *Conn) Unlisten(ctx context.Context, channel string) error {
	return fmt.Errorf("GORM doesn't support LISTEN/NOTIFY")
}

// WaitForNotification - GORM doesn't support LISTEN/NOTIFY
func (c *Conn) WaitForNotification(ctx context.Context, timeout time.Duration) (*postgresql.Notification, error) {
	return nil, fmt.Errorf("GORM doesn't support LISTEN/NOTIFY")
}
