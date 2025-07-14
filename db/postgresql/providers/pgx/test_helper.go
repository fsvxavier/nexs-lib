//go:build unit

package pgx

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/pashagolub/pgxmock/v4"
)

// TestHelper provides utilities for testing with pgxmock
type TestHelper struct {
	Mock   pgxmock.PgxConnIface
	Config *config.Config
}

// NewTestHelper creates a new test helper with a pgxmock connection
func NewTestHelper() (*TestHelper, error) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		return nil, err
	}

	cfg := config.DefaultConfig()

	return &TestHelper{
		Mock:   mock,
		Config: cfg,
	}, nil
}

// NewTestHelperWithConfig creates a new test helper with custom config
func NewTestHelperWithConfig(cfg *config.Config) (*TestHelper, error) {
	mock, err := pgxmock.NewConn()
	if err != nil {
		return nil, err
	}

	return &TestHelper{
		Mock:   mock,
		Config: cfg,
	}, nil
}

// NewConnWithMock creates a Conn instance with the mock connection
// Note: This creates a test connection that bypasses the real pgx.Conn
func (h *TestHelper) NewConnWithMock() *ConnMock {
	return &ConnMock{
		mock:               h.Mock,
		config:             h.Config,
		logger:             defaultLogger(),
		released:           false,
		isPooled:           false,
		multiTenantEnabled: h.Config.MultiTenantEnabled,
	}
}

// Close closes the mock connection
func (h *TestHelper) Close() error {
	return h.Mock.Close(context.Background())
}

// ExpectationsWereMet checks if all expectations were met
func (h *TestHelper) ExpectationsWereMet() error {
	return h.Mock.ExpectationsWereMet()
}

// ExpectQuery sets up an expectation for a query
func (h *TestHelper) ExpectQuery(query string) *pgxmock.ExpectedQuery {
	return h.Mock.ExpectQuery(query)
}

// ExpectExec sets up an expectation for an exec
func (h *TestHelper) ExpectExec(query string) *pgxmock.ExpectedExec {
	return h.Mock.ExpectExec(query)
}

// ExpectBegin sets up an expectation for begin transaction
func (h *TestHelper) ExpectBegin() *pgxmock.ExpectedBegin {
	return h.Mock.ExpectBegin()
}

// ExpectCommit sets up an expectation for commit
func (h *TestHelper) ExpectCommit() *pgxmock.ExpectedCommit {
	return h.Mock.ExpectCommit()
}

// ExpectRollback sets up an expectation for rollback
func (h *TestHelper) ExpectRollback() *pgxmock.ExpectedRollback {
	return h.Mock.ExpectRollback()
}

// ExpectPing sets up an expectation for ping
func (h *TestHelper) ExpectPing() *pgxmock.ExpectedPing {
	return h.Mock.ExpectPing()
}

// ExpectPrepare sets up an expectation for prepare
func (h *TestHelper) ExpectPrepare(name, query string) *pgxmock.ExpectedPrepare {
	return h.Mock.ExpectPrepare(name, query)
}

// defaultLogger returns a no-op logger for testing
func defaultLogger() *log.Logger {
	return log.New(&noopWriter{}, "", 0)
}

type noopWriter struct{}

func (w *noopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// ConnMock is a test version of Conn that uses pgxmock
type ConnMock struct {
	mock               pgxmock.PgxConnIface
	config             *config.Config
	logger             *log.Logger
	released           bool
	isPooled           bool
	multiTenantEnabled bool
}

// Implement all IConn methods for ConnMock to satisfy the interface
func (c *ConnMock) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}

	// Execute before query hook
	if c.config.Hooks != nil && c.config.Hooks.BeforeQuery != nil {
		if err := c.config.Hooks.BeforeQuery(ctx, query, args); err != nil {
			return fmt.Errorf("before query hook failed: %w", err)
		}
	}

	start := time.Now()
	rows, err := c.mock.Query(ctx, query, args...)
	duration := time.Since(start)

	// Execute after query hook
	if c.config.Hooks != nil && c.config.Hooks.AfterQuery != nil {
		hookErr := c.config.Hooks.AfterQuery(ctx, query, args, duration, err)
		if hookErr != nil && c.logger != nil {
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

func (c *ConnMock) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return fmt.Errorf("not fully implemented yet")
}

func (c *ConnMock) QueryCount(ctx context.Context, query string, args ...interface{}) (int, error) {
	var count int
	err := c.QueryOne(ctx, &count, query, args...)
	return count, err
}

func (c *ConnMock) Query(ctx context.Context, query string, args ...interface{}) (postgresql.IRows, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}
	rows, err := c.mock.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows: rows}, nil
}

func (c *ConnMock) QueryRow(ctx context.Context, query string, args ...interface{}) (postgresql.IRow, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}
	row := c.mock.QueryRow(ctx, query, args...)
	return &Row{row: row}, nil
}

func (c *ConnMock) Exec(ctx context.Context, query string, args ...interface{}) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}
	_, err := c.mock.Exec(ctx, query, args...)
	return err
}

func (c *ConnMock) SendBatch(ctx context.Context, batch postgresql.IBatch) (postgresql.IBatchResults, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}
	return nil, fmt.Errorf("batch not implemented in mock")
}

func (c *ConnMock) BeginTransaction(ctx context.Context) (postgresql.ITransaction, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}
	tx, err := c.mock.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx}, nil
}

func (c *ConnMock) BeginTransactionWithOptions(ctx context.Context, options postgresql.TxOptions) (postgresql.ITransaction, error) {
	return c.BeginTransaction(ctx)
}

func (c *ConnMock) Prepare(ctx context.Context, name, query string) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}
	_, err := c.mock.Prepare(ctx, name, query)
	return err
}

func (c *ConnMock) Ping(ctx context.Context) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}
	return c.mock.Ping(ctx)
}

func (c *ConnMock) Listen(ctx context.Context, channel string) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}
	return fmt.Errorf("listen not implemented in mock")
}

func (c *ConnMock) Unlisten(ctx context.Context, channel string) error {
	if c.released {
		return fmt.Errorf("connection is released")
	}
	return fmt.Errorf("unlisten not implemented in mock")
}

func (c *ConnMock) WaitForNotification(ctx context.Context, timeout time.Duration) (*postgresql.Notification, error) {
	if c.released {
		return nil, fmt.Errorf("connection is released")
	}
	return nil, fmt.Errorf("wait for notification not implemented in mock")
}

func (c *ConnMock) Close(ctx context.Context) error {
	return c.mock.Close(ctx)
}

func (c *ConnMock) Release(ctx context.Context) {
	c.released = true
	c.BeforeReleaseHook(ctx)
}

func (c *ConnMock) AfterAcquireHook(ctx context.Context) error {
	if !c.multiTenantEnabled {
		return nil
	}
	return fmt.Errorf("multi-tenant hook not implemented in mock")
}

func (c *ConnMock) BeforeReleaseHook(ctx context.Context) error {
	if c.config.Hooks != nil && c.config.Hooks.BeforeRelease != nil {
		return c.config.Hooks.BeforeRelease(ctx, c)
	}
	return nil
}
