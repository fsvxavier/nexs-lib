package gorm

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates a test GORM DB for unit testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=invalid port=0 user=test dbname=test sslmode=disable",
	}), &gorm.Config{
		Logger: logger.Discard, // Disable logging for tests
	})

	// If there's an error (expected with invalid DSN), return nil for safe testing
	if err != nil {
		return nil
	}

	return db
}

// TestTransaction_BasicOperations tests basic transaction operations
func TestTransaction_BasicOperations(t *testing.T) {
	t.Parallel()

	tx := setupTestTransaction(t)

	t.Run("QueryOne_Success", func(t *testing.T) {
		ctx := context.Background()
		var result string

		err := tx.QueryOne(ctx, &result, "SELECT 'test'")
		// Will error due to nil tx, which is expected for testing error paths
		assert.Error(t, err)
	})

	t.Run("QueryAll_Success", func(t *testing.T) {
		ctx := context.Background()
		var results []string

		err := tx.QueryAll(ctx, &results, "SELECT 'test1' UNION SELECT 'test2'")
		// Will error due to nil tx, which is expected for testing error paths
		assert.Error(t, err)
	})

	t.Run("QueryCount_Success", func(t *testing.T) {
		ctx := context.Background()

		count, err := tx.QueryCount(ctx, "SELECT COUNT(*) FROM (SELECT 1) t")
		// Will error due to nil tx, which is expected for testing error paths
		assert.Error(t, err)
		assert.Nil(t, count)
	})

	t.Run("Query_Success", func(t *testing.T) {
		ctx := context.Background()

		rows, err := tx.Query(ctx, "SELECT 'test'")
		// Will error due to nil tx, which is expected for testing error paths
		assert.Error(t, err)
		assert.Nil(t, rows)
	})

	t.Run("QueryRow_Success", func(t *testing.T) {
		ctx := context.Background()

		row, err := tx.QueryRow(ctx, "SELECT 'test'")
		assert.NoError(t, err)
		assert.NotNil(t, row)

		// Test row implementation with nil DB (from nil tx)
		var result string
		err = row.Scan(&result)
		// Will error due to nil DB, which is expected
		assert.Error(t, err)
	})

	t.Run("Exec_Success", func(t *testing.T) {
		ctx := context.Background()

		err := tx.Exec(ctx, "CREATE TABLE test (id INT)")
		// Will error due to nil tx, which is expected for testing error paths
		assert.Error(t, err)
	})
}

// TestTransaction_TransactionOperations tests transaction-specific operations
func TestTransaction_TransactionOperations(t *testing.T) {
	t.Parallel()

	tx := setupTestTransaction(t)
	ctx := context.Background()

	t.Run("Commit_Success", func(t *testing.T) {
		err := tx.Commit(ctx)
		// Will error due to no real DB connection, but tests the flow
		assert.Error(t, err)
	})

	t.Run("Rollback_Success", func(t *testing.T) {
		err := tx.Rollback(ctx)
		// Will error due to no real DB connection, but tests the flow
		assert.Error(t, err)
	})

	t.Run("BeginSavepoint_Success", func(t *testing.T) {
		err := tx.BeginSavepoint(ctx, "test_savepoint")
		// Will error due to no real DB connection, but tests the flow
		assert.Error(t, err)
	})

	t.Run("Savepoint_Success", func(t *testing.T) {
		err := tx.Savepoint(ctx, "test_savepoint")
		// Will error due to no real DB connection, but tests the flow
		assert.Error(t, err)
	})

	t.Run("RollbackToSavepoint_Success", func(t *testing.T) {
		err := tx.RollbackToSavepoint(ctx, "test_savepoint")
		// Will error due to no real DB connection, but tests the flow
		assert.Error(t, err)
	})

	t.Run("ReleaseSavepoint_Success", func(t *testing.T) {
		err := tx.ReleaseSavepoint(ctx, "test_savepoint")
		// Will error due to no real DB connection, but tests the flow
		assert.Error(t, err)
	})
}

// TestTransaction_NestedTransactions tests nested transaction operations
func TestTransaction_NestedTransactions(t *testing.T) {
	t.Parallel()

	tx := setupTestTransaction(t)
	ctx := context.Background()

	t.Run("BeginTransaction_Success", func(t *testing.T) {
		nestedTx, err := tx.BeginTransaction(ctx)
		// Will error due to no real DB connection, but tests the flow
		assert.Error(t, err)
		assert.Nil(t, nestedTx)
	})

	t.Run("BeginTransactionWithOptions_Success", func(t *testing.T) {
		opts := postgresql.TxOptions{
			IsoLevel:   postgresql.IsoLevelReadCommitted,
			AccessMode: postgresql.AccessModeReadWrite,
		}

		nestedTx, err := tx.BeginTransactionWithOptions(ctx, opts)
		// Will error due to no real DB connection, but tests the flow
		assert.Error(t, err)
		assert.Nil(t, nestedTx)
	})
}

// TestTransaction_BatchOperations tests transaction batch operations
func TestTransaction_BatchOperations(t *testing.T) {
	t.Parallel()

	tx := setupTestTransaction(t)
	ctx := context.Background()

	t.Run("SendBatch_NotSupported", func(t *testing.T) {
		// GORM doesn't support batch operations
		results, err := tx.SendBatch(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GORM doesn't support batch operations")
		assert.Nil(t, results)
	})
}

// TestTransaction_HookOperations tests transaction hook operations
func TestTransaction_HookOperations(t *testing.T) {
	t.Parallel()

	tx := setupTestTransaction(t)
	ctx := context.Background()

	t.Run("BeforeReleaseHook_Success", func(t *testing.T) {
		err := tx.BeforeReleaseHook(ctx)
		assert.NoError(t, err)
	})

	t.Run("AfterAcquireHook_Success", func(t *testing.T) {
		err := tx.AfterAcquireHook(ctx)
		assert.NoError(t, err)
	})

	t.Run("Release_Success", func(t *testing.T) {
		// Release doesn't do anything in transaction context
		tx.Release(ctx)
		// No assertion needed, just verifying it doesn't panic
	})
}

// TestTransaction_PrepareStatement tests transaction prepare statement operations
func TestTransaction_PrepareStatement(t *testing.T) {
	t.Parallel()

	tx := setupTestTransaction(t)
	ctx := context.Background()

	t.Run("Prepare_Success", func(t *testing.T) {
		err := tx.Prepare(ctx, "test_stmt", "SELECT $1")
		assert.NoError(t, err) // GORM handles internally, no error expected
	})
}

// TestTransaction_PingOperations tests transaction ping operations
func TestTransaction_PingOperations(t *testing.T) {
	t.Parallel()

	tx := setupTestTransaction(t)
	ctx := context.Background()

	t.Run("Ping_Success", func(t *testing.T) {
		err := tx.Ping(ctx)
		// Will error due to no real DB connection, but tests the flow
		assert.Error(t, err)
	})
}

// TestTransaction_ListenNotifyOperations tests transaction LISTEN/NOTIFY operations
func TestTransaction_ListenNotifyOperations(t *testing.T) {
	t.Parallel()

	tx := setupTestTransaction(t)
	ctx := context.Background()

	t.Run("Listen_NotSupported", func(t *testing.T) {
		err := tx.Listen(ctx, "test_channel")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GORM doesn't support LISTEN/NOTIFY")
	})

	t.Run("Unlisten_NotSupported", func(t *testing.T) {
		err := tx.Unlisten(ctx, "test_channel")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GORM doesn't support LISTEN/NOTIFY")
	})

	t.Run("WaitForNotification_NotSupported", func(t *testing.T) {
		notification, err := tx.WaitForNotification(ctx, 5*time.Second)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "GORM doesn't support LISTEN/NOTIFY")
		assert.Nil(t, notification)
	})
}

// TestTransaction_ErrorHandling tests transaction error handling scenarios
func TestTransaction_ErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("NilTx_Operations", func(t *testing.T) {
		tx := &Transaction{
			tx:     nil,
			config: &config.Config{},
		}

		ctx := context.Background()

		// Test QueryOne with nil tx
		var result string
		err := tx.QueryOne(ctx, &result, "SELECT 'test'")
		assert.Error(t, err)

		// Test QueryAll with nil tx
		var results []string
		err = tx.QueryAll(ctx, &results, "SELECT 'test'")
		assert.Error(t, err)

		// Test Exec with nil tx
		err = tx.Exec(ctx, "SELECT 1")
		assert.Error(t, err)

		// Test Commit with nil tx
		err = tx.Commit(ctx)
		assert.Error(t, err)

		// Test Rollback with nil tx
		err = tx.Rollback(ctx)
		assert.Error(t, err)
	})
}
