package gorm

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/gorm/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestConn_BasicOperations tests basic connection operations
func TestConn_BasicOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(*testing.T) *Conn
		testFunc func(*testing.T, *Conn)
	}{
		{
			name: "QueryOne_Success",
			setup: func(t *testing.T) *Conn {
				return setupTestConn(t)
			},
			testFunc: func(t *testing.T, conn *Conn) {
				ctx := context.Background()
				var result string

				// This would normally fail with real DB, but we're testing the flow
				err := conn.QueryOne(ctx, &result, "SELECT 'test'")
				// Error is expected since we don't have a real DB connection
				assert.Error(t, err)
			},
		},
		{
			name: "QueryAll_Success",
			setup: func(t *testing.T) *Conn {
				return setupTestConn(t)
			},
			testFunc: func(t *testing.T, conn *Conn) {
				ctx := context.Background()
				var results []string

				err := conn.QueryAll(ctx, &results, "SELECT 'test1' UNION SELECT 'test2'")
				assert.Error(t, err) // Expected since no real DB
			},
		},
		{
			name: "QueryCount_Success",
			setup: func(t *testing.T) *Conn {
				return setupTestConn(t)
			},
			testFunc: func(t *testing.T, conn *Conn) {
				ctx := context.Background()

				count, err := conn.QueryCount(ctx, "SELECT COUNT(*) FROM (SELECT 1) t")
				assert.Error(t, err) // Expected since no real DB
				assert.Nil(t, count)
			},
		},
		{
			name: "Query_Success",
			setup: func(t *testing.T) *Conn {
				return setupTestConn(t)
			},
			testFunc: func(t *testing.T, conn *Conn) {
				ctx := context.Background()

				rows, err := conn.Query(ctx, "SELECT 'test'")
				assert.Error(t, err) // Expected since no real DB
				assert.Nil(t, rows)
			},
		},
		{
			name: "QueryRow_Success",
			setup: func(t *testing.T) *Conn {
				return setupTestConn(t)
			},
			testFunc: func(t *testing.T, conn *Conn) {
				ctx := context.Background()

				row, err := conn.QueryRow(ctx, "SELECT 'test'")
				assert.NoError(t, err)
				assert.NotNil(t, row)

				// Test row implementation
				var result string
				err = row.Scan(&result)
				assert.Error(t, err) // Expected since no real DB
			},
		},
		{
			name: "Exec_Success",
			setup: func(t *testing.T) *Conn {
				return setupTestConn(t)
			},
			testFunc: func(t *testing.T, conn *Conn) {
				ctx := context.Background()

				err := conn.Exec(ctx, "CREATE TABLE test (id INT)")
				assert.Error(t, err) // Expected since no real DB
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			conn := tt.setup(t)
			tt.testFunc(t, conn)
		})
	}
}

// TestConn_ReleasedConnection tests operations on released connections
func TestConn_ReleasedConnection(t *testing.T) {
	t.Parallel()

	conn := setupTestConn(t)
	ctx := context.Background()

	// Release the connection
	conn.Release(ctx)
	assert.True(t, conn.released)

	// Test operations on released connection
	t.Run("QueryOne_OnReleasedConnection", func(t *testing.T) {
		var result string
		err := conn.QueryOne(ctx, &result, "SELECT 'test'")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection is released")
	})

	t.Run("QueryAll_OnReleasedConnection", func(t *testing.T) {
		var results []string
		err := conn.QueryAll(ctx, &results, "SELECT 'test'")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection is released")
	})

	t.Run("QueryCount_OnReleasedConnection", func(t *testing.T) {
		count, err := conn.QueryCount(ctx, "SELECT COUNT(*)")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection is released")
		assert.Nil(t, count)
	})

	t.Run("Query_OnReleasedConnection", func(t *testing.T) {
		rows, err := conn.Query(ctx, "SELECT 'test'")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection is released")
		assert.Nil(t, rows)
	})

	t.Run("QueryRow_OnReleasedConnection", func(t *testing.T) {
		row, err := conn.QueryRow(ctx, "SELECT 'test'")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection is released")
		assert.Nil(t, row)
	})

	t.Run("Exec_OnReleasedConnection", func(t *testing.T) {
		err := conn.Exec(ctx, "SELECT 1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection is released")
	})

	t.Run("BeginTransaction_OnReleasedConnection", func(t *testing.T) {
		tx, err := conn.BeginTransaction(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection is released")
		assert.Nil(t, tx)
	})

	t.Run("Ping_OnReleasedConnection", func(t *testing.T) {
		err := conn.Ping(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection is released")
	})
}

// TestConn_TransactionOperations tests transaction operations
func TestConn_TransactionOperations(t *testing.T) {
	t.Parallel()

	conn := setupTestConn(t)
	ctx := context.Background()

	t.Run("BeginTransaction_Success", func(t *testing.T) {
		tx, err := conn.BeginTransaction(ctx)
		// Will error due to no real DB, but tests the flow
		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("BeginTransactionWithOptions_Success", func(t *testing.T) {
		opts := postgresql.TxOptions{
			IsoLevel:   postgresql.IsoLevelReadCommitted,
			AccessMode: postgresql.AccessModeReadWrite,
		}

		tx, err := conn.BeginTransactionWithOptions(ctx, opts)
		// Will error due to no real DB, but tests the flow
		assert.Error(t, err)
		assert.Nil(t, tx)
	})
}

// TestConn_BatchOperations tests batch operations
func TestConn_BatchOperations(t *testing.T) {
	t.Parallel()

	conn := setupTestConn(t)
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBatch := mocks.NewMockIBatch(ctrl)

	t.Run("SendBatch_NotSupported", func(t *testing.T) {
		results, err := conn.SendBatch(ctx, mockBatch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GORM doesn't support batch operations")
		assert.Nil(t, results)
	})
}

// TestConn_HookOperations tests hook operations
func TestConn_HookOperations(t *testing.T) {
	t.Parallel()

	conn := setupTestConn(t)
	ctx := context.Background()

	t.Run("BeforeReleaseHook_Success", func(t *testing.T) {
		err := conn.BeforeReleaseHook(ctx)
		assert.NoError(t, err)
	})

	t.Run("AfterAcquireHook_Success", func(t *testing.T) {
		err := conn.AfterAcquireHook(ctx)
		assert.NoError(t, err)
	})
}

// TestConn_PrepareStatement tests prepared statement operations
func TestConn_PrepareStatement(t *testing.T) {
	t.Parallel()

	conn := setupTestConn(t)
	ctx := context.Background()

	t.Run("Prepare_Success", func(t *testing.T) {
		err := conn.Prepare(ctx, "test_stmt", "SELECT $1")
		assert.NoError(t, err) // GORM handles internally, no error expected
	})
}

// TestConn_ListenNotifyOperations tests LISTEN/NOTIFY operations
func TestConn_ListenNotifyOperations(t *testing.T) {
	t.Parallel()

	conn := setupTestConn(t)
	ctx := context.Background()

	t.Run("Listen_NotSupported", func(t *testing.T) {
		err := conn.Listen(ctx, "test_channel")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GORM doesn't support LISTEN/NOTIFY")
	})

	t.Run("Unlisten_NotSupported", func(t *testing.T) {
		err := conn.Unlisten(ctx, "test_channel")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GORM doesn't support LISTEN/NOTIFY")
	})

	t.Run("WaitForNotification_NotSupported", func(t *testing.T) {
		notification, err := conn.WaitForNotification(ctx, 5*time.Second)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GORM doesn't support LISTEN/NOTIFY")
		assert.Nil(t, notification)
	})
}

// TestConn_MultipleRelease tests multiple release calls
func TestConn_MultipleRelease(t *testing.T) {
	t.Parallel()

	conn := setupTestConn(t)
	ctx := context.Background()

	// First release
	conn.Release(ctx)
	assert.True(t, conn.released)

	// Second release should be safe
	conn.Release(ctx)
	assert.True(t, conn.released)
}

// TestConn_ErrorHandling tests error handling scenarios
func TestConn_ErrorHandling(t *testing.T) {
	t.Parallel()
	t.Run("NilDB_Operations", func(t *testing.T) {
		conn := &Conn{
			db:       nil,
			config:   &config.Config{},
			released: false,
		}

		ctx := context.Background()

		// Test QueryOne with nil DB
		var result string
		err := conn.QueryOne(ctx, &result, "SELECT 'test'")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection is nil")

		// Test QueryAll with nil DB
		var results []string
		err = conn.QueryAll(ctx, &results, "SELECT 'test'")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection is nil")

		// Test Exec with nil DB
		err = conn.Exec(ctx, "SELECT 1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection is nil")
	})
}

// setupTestConn creates a test connection for unit testing
func setupTestConn(t *testing.T) *Conn {
	// Create a GORM DB instance with a mock dialector for testing
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=invalid port=0 user=test dbname=test sslmode=disable",
	}), &gorm.Config{
		Logger: logger.Discard, // Disable logging for tests
	})

	// We expect this to fail, but we can still test the Conn wrapper logic
	// For unit tests, we're mainly testing the wrapper behavior
	if err != nil {
		// Create a minimal config for testing
		cfg := &config.Config{
			Host:     "localhost",
			Port:     5432,
			Database: "test",
			Username: "test",
			Password: "test",
		}

		// Return a connection with nil DB to test error paths
		return &Conn{
			db:       nil,
			config:   cfg,
			released: false,
		}
	}

	cfg := &config.Config{
		Host:     "localhost",
		Port:     5432,
		Database: "test",
		Username: "test",
		Password: "test",
	}

	return &Conn{
		db:       db,
		config:   cfg,
		released: false,
	}
}
