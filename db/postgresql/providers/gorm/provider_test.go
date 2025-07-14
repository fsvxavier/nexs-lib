package gorm

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/providers/gorm/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Test helpers
func createMockGormDBWithPing(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	return gormDB, mock, sqlDB
}

func createMockGormDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	return gormDB, mock, sqlDB
}

func createValidConfig() *config.Config {
	return &config.Config{
		Host:               "localhost",
		Port:               5432,
		Database:           "testdb",
		Username:           "testuser",
		Password:           "testpass",
		MaxConns:           10,
		MinConns:           1,
		MaxConnLifetime:    time.Hour,
		MaxConnIdleTime:    time.Minute * 30,
		ConnectTimeout:     time.Second * 30,
		QueryTimeout:       time.Second * 30,
		ApplicationName:    "test-app",
		SearchPath:         []string{"public"},
		Timezone:           "UTC",
		DefaultSchema:      "public",
		MultiTenantEnabled: false,
	}
}

// TestProvider_BasicOperations tests basic provider operations
func TestProvider_BasicOperations(t *testing.T) {
	t.Parallel()

	provider := NewProvider()

	t.Run("Provider_Metadata", func(t *testing.T) {
		assert.Equal(t, postgresql.ProviderType("gorm"), provider.Type())
		assert.Equal(t, "gorm", provider.Name())
		assert.Equal(t, "gorm-v1.30", provider.Version())
	})

	t.Run("Provider_Health_And_Metrics", func(t *testing.T) {
		ctx := context.Background()

		// Test health check
		healthy := provider.IsHealthy(ctx)
		assert.True(t, healthy)

		// Test metrics
		metrics := provider.GetMetrics(ctx)
		assert.NotNil(t, metrics)

		// Test close
		provider.Close()
	})
}

// TestProvider_ConfigValidation tests configuration validation
func TestProvider_ConfigValidation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	provider := NewProvider()

	t.Run("Nil_Config", func(t *testing.T) {
		pool, err := provider.CreatePool(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, pool)
		assert.Contains(t, err.Error(), "config cannot be nil")

		conn, err := provider.CreateConnection(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, conn)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("Invalid_Config_Fields", func(t *testing.T) {
		invalidConfigs := []*config.Config{
			{MaxConns: -1, Host: "localhost", Database: "test", Username: "user", Password: "pass"},
			{MinConns: -1, Host: "localhost", Database: "test", Username: "user", Password: "pass"},
			{MaxConnLifetime: -1, Host: "localhost", Database: "test", Username: "user", Password: "pass"},
		}

		for i, cfg := range invalidConfigs {
			pool, err := provider.CreatePool(ctx, cfg)
			assert.Error(t, err, "Config %d should fail", i)
			assert.Nil(t, pool, "Pool should be nil for config %d", i)
		}
	})
}

// TestProvider_WithGomockMocks tests provider with gomock mocks
func TestProvider_WithGomockMocks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("Mock_Pool_Operations", func(t *testing.T) {
		mockPool := mocks.NewMockIPool(ctrl)
		mockConn := mocks.NewMockIConn(ctrl)

		// Setup expectations
		mockPool.EXPECT().Acquire(ctx).Return(mockConn, nil)
		mockPool.EXPECT().Stats().Return(postgresql.PoolStats{
			AcquiredConns: 2,
			IdleConns:     8,
			MaxConns:      10,
			TotalConns:    10,
		})
		mockPool.EXPECT().Close()

		// Test pool operations
		conn, err := mockPool.Acquire(ctx)
		assert.NoError(t, err)
		assert.Equal(t, mockConn, conn)

		stats := mockPool.Stats()
		assert.Equal(t, int32(2), stats.AcquiredConns)
		assert.Equal(t, int32(8), stats.IdleConns)
		assert.Equal(t, int32(10), stats.MaxConns)

		mockPool.Close()
	})

	t.Run("Mock_Connection_Operations", func(t *testing.T) {
		mockConn := mocks.NewMockIConn(ctrl)
		mockRows := mocks.NewMockIRows(ctrl)
		mockRow := mocks.NewMockIRow(ctrl)
		mockTx := mocks.NewMockITransaction(ctrl)

		// Setup query expectations
		mockConn.EXPECT().Query(ctx, "SELECT * FROM users", gomock.Any()).Return(mockRows, nil)
		mockConn.EXPECT().QueryRow(ctx, "SELECT COUNT(*) FROM users").Return(mockRow, nil)
		mockConn.EXPECT().Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "test").Return(nil)
		mockConn.EXPECT().BeginTransaction(ctx).Return(mockTx, nil)
		mockConn.EXPECT().Release(ctx)

		// Test operations
		rows, err := mockConn.Query(ctx, "SELECT * FROM users")
		assert.NoError(t, err)
		assert.Equal(t, mockRows, rows)

		row, err := mockConn.QueryRow(ctx, "SELECT COUNT(*) FROM users")
		assert.NoError(t, err)
		assert.Equal(t, mockRow, row)

		err = mockConn.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "test")
		assert.NoError(t, err)

		tx, err := mockConn.BeginTransaction(ctx)
		assert.NoError(t, err)
		assert.Equal(t, mockTx, tx)

		mockConn.Release(ctx)
	})

	t.Run("Mock_Transaction_Operations", func(t *testing.T) {
		mockTx := mocks.NewMockITransaction(ctrl)
		mockRows := mocks.NewMockIRows(ctrl)

		// Setup transaction expectations
		mockTx.EXPECT().Query(ctx, "SELECT * FROM users WHERE active = $1", true).Return(mockRows, nil)
		mockTx.EXPECT().Exec(ctx, "UPDATE users SET last_login = NOW() WHERE id = $1", 1).Return(nil)
		mockTx.EXPECT().Commit(ctx).Return(nil)

		// Test transaction operations
		rows, err := mockTx.Query(ctx, "SELECT * FROM users WHERE active = $1", true)
		assert.NoError(t, err)
		assert.Equal(t, mockRows, rows)

		err = mockTx.Exec(ctx, "UPDATE users SET last_login = NOW() WHERE id = $1", 1)
		assert.NoError(t, err)

		err = mockTx.Commit(ctx)
		assert.NoError(t, err)
	})
}

// TestProvider_ErrorHandling tests error handling scenarios
func TestProvider_ErrorHandling(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("Connection_Errors", func(t *testing.T) {
		mockConn := mocks.NewMockIConn(ctrl)
		expectedErr := errors.New("connection failed")

		// Setup error expectations
		mockConn.EXPECT().Query(ctx, "SELECT * FROM invalid_table").Return(nil, expectedErr)
		mockConn.EXPECT().Exec(ctx, "INVALID SQL").Return(expectedErr)

		// Test error handling
		rows, err := mockConn.Query(ctx, "SELECT * FROM invalid_table")
		assert.Error(t, err)
		assert.Nil(t, rows)
		assert.Equal(t, expectedErr, err)

		err = mockConn.Exec(ctx, "INVALID SQL")
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("Transaction_Errors", func(t *testing.T) {
		mockTx := mocks.NewMockITransaction(ctrl)
		rollbackErr := errors.New("rollback failed")

		// Setup transaction error expectations
		mockTx.EXPECT().Rollback(ctx).Return(rollbackErr)

		// Test transaction error handling
		err := mockTx.Rollback(ctx)
		assert.Error(t, err)
		assert.Equal(t, rollbackErr, err)
	})
}

// TestProvider_RowOperations tests row operations with mocks
func TestProvider_RowOperations(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("Rows_Operations", func(t *testing.T) {
		mockRows := mocks.NewMockIRows(ctrl)

		// Setup rows expectations
		mockRows.EXPECT().Next().Return(true).Times(2)
		mockRows.EXPECT().Next().Return(false)
		mockRows.EXPECT().Scan(gomock.Any()).Return(nil).Times(2)
		mockRows.EXPECT().Close() // Close() has no return value
		mockRows.EXPECT().Err().Return(nil)

		// Test rows operations
		assert.True(t, mockRows.Next())
		assert.True(t, mockRows.Next())
		assert.False(t, mockRows.Next())

		var dest interface{}
		err := mockRows.Scan(&dest)
		assert.NoError(t, err)

		err = mockRows.Scan(&dest)
		assert.NoError(t, err)

		mockRows.Close() // Close() has no return value

		err = mockRows.Err()
		assert.NoError(t, err)
	})

	t.Run("Row_Operations", func(t *testing.T) {
		mockRow := mocks.NewMockIRow(ctrl)

		// Setup row expectations
		mockRow.EXPECT().Scan(gomock.Any()).Return(nil)

		// Test row operations
		var dest interface{}
		err := mockRow.Scan(&dest)
		assert.NoError(t, err)
	})
}

// TestProvider_Advanced_Scenarios tests advanced usage scenarios
func TestProvider_Advanced_Scenarios(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("Nested_Transactions", func(t *testing.T) {
		mockTx1 := mocks.NewMockITransaction(ctrl)
		mockTx2 := mocks.NewMockITransaction(ctrl)

		// Setup nested transaction expectations
		mockTx1.EXPECT().BeginTransaction(ctx).Return(mockTx2, nil)
		mockTx1.EXPECT().Savepoint(ctx, "sp1").Return(nil)
		mockTx2.EXPECT().Rollback(ctx).Return(nil)
		mockTx1.EXPECT().RollbackToSavepoint(ctx, "sp1").Return(nil)
		mockTx1.EXPECT().Commit(ctx).Return(nil)

		// Test nested transactions
		nestedTx, err := mockTx1.BeginTransaction(ctx)
		assert.NoError(t, err)
		assert.Equal(t, mockTx2, nestedTx)

		err = mockTx1.Savepoint(ctx, "sp1")
		assert.NoError(t, err)

		err = nestedTx.Rollback(ctx)
		assert.NoError(t, err)

		err = mockTx1.RollbackToSavepoint(ctx, "sp1")
		assert.NoError(t, err)

		err = mockTx1.Commit(ctx)
		assert.NoError(t, err)
	})

	t.Run("Batch_Operations", func(t *testing.T) {
		mockConn := mocks.NewMockIConn(ctrl)
		mockBatch := mocks.NewMockIBatch(ctrl)
		mockBatchResults := mocks.NewMockIBatchResults(ctrl)

		// Setup batch expectations
		mockConn.EXPECT().SendBatch(ctx, mockBatch).Return(mockBatchResults, nil)
		mockBatchResults.EXPECT().Close() // Close() has no return value

		// Test batch operations
		results, err := mockConn.SendBatch(ctx, mockBatch)
		assert.NoError(t, err)
		assert.Equal(t, mockBatchResults, results)

		results.Close() // Close() has no return value
	})

	t.Run("Connection_Hooks", func(t *testing.T) {
		mockConn := mocks.NewMockIConn(ctrl)

		// Setup hook expectations
		mockConn.EXPECT().BeforeReleaseHook(ctx).Return(nil)
		mockConn.EXPECT().AfterAcquireHook(ctx).Return(nil)

		// Test hooks
		err := mockConn.BeforeReleaseHook(ctx)
		assert.NoError(t, err)

		err = mockConn.AfterAcquireHook(ctx)
		assert.NoError(t, err)
	})
}

// TestProvider_Timeout_Scenarios tests timeout handling
func TestProvider_Timeout_Scenarios(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("Context_Timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()

		mockConn := mocks.NewMockIConn(ctrl)
		timeoutErr := context.DeadlineExceeded

		// Setup timeout expectations
		mockConn.EXPECT().Query(ctx, "SELECT SLEEP(1)").Return(nil, timeoutErr)

		// Test timeout handling
		rows, err := mockConn.Query(ctx, "SELECT SLEEP(1)")
		assert.Error(t, err)
		assert.Nil(t, rows)
		assert.Equal(t, timeoutErr, err)
	})

	t.Run("Pool_Acquire_Timeout", func(t *testing.T) {
		ctx := context.Background()
		mockPool := mocks.NewMockIPool(ctrl)
		timeout := time.Millisecond * 100

		// Setup timeout expectations
		mockPool.EXPECT().AcquireWithTimeout(ctx, timeout).Return(nil, context.DeadlineExceeded)

		// Test pool timeout
		conn, err := mockPool.AcquireWithTimeout(ctx, timeout)
		assert.Error(t, err)
		assert.Nil(t, conn)
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}
