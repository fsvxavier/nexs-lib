package gorm

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/golang/mock/gomock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// BenchmarkConn_Operations benchmarks connection operations
func BenchmarkConn_Operations(b *testing.B) {
	conn := setupBenchmarkConn(b)
	ctx := context.Background()

	b.Run("QueryOne", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result string
			_ = conn.QueryOne(ctx, &result, "SELECT 'benchmark'")
		}
	})

	b.Run("QueryAll", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var results []string
			_ = conn.QueryAll(ctx, &results, "SELECT 'benchmark'")
		}
	})

	b.Run("QueryCount", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = conn.QueryCount(ctx, "SELECT COUNT(*) FROM (SELECT 1) t")
		}
	})

	b.Run("QueryRow", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = conn.QueryRow(ctx, "SELECT 'benchmark'")
		}
	})

	b.Run("Exec", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = conn.Exec(ctx, "SELECT 1")
		}
	})

	b.Run("BeginTransaction", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = conn.BeginTransaction(ctx)
		}
	})

	b.Run("Ping", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = conn.Ping(ctx)
		}
	})

	b.Run("Release", func(b *testing.B) {
		// Create fresh connection for each release test
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			testConn := setupBenchmarkConn(b)
			testConn.Release(ctx)
		}
	})
}

// BenchmarkConn_HookOperations benchmarks connection hook operations
func BenchmarkConn_HookOperations(b *testing.B) {
	conn := setupBenchmarkConn(b)
	ctx := context.Background()

	b.Run("BeforeReleaseHook", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = conn.BeforeReleaseHook(ctx)
		}
	})

	b.Run("AfterAcquireHook", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = conn.AfterAcquireHook(ctx)
		}
	})

	b.Run("Prepare", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = conn.Prepare(ctx, "bench_stmt", "SELECT $1")
		}
	})
}

// BenchmarkPool_Operations benchmarks pool operations
func BenchmarkPool_Operations(b *testing.B) {
	pool := setupBenchmarkPool(b)
	defer pool.Close()
	ctx := context.Background()

	b.Run("Acquire", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			conn, _ := pool.Acquire(ctx)
			if conn != nil {
				conn.Release(ctx)
			}
		}
	})

	b.Run("AcquireWithTimeout", func(b *testing.B) {
		timeout := 5 * time.Second
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			conn, _ := pool.AcquireWithTimeout(ctx, timeout)
			if conn != nil {
				conn.Release(ctx)
			}
		}
	})

	b.Run("Stats", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = pool.Stats()
		}
	})

	b.Run("Ping", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = pool.Ping(ctx)
		}
	})

	b.Run("HealthCheck", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = pool.HealthCheck(ctx)
		}
	})

	b.Run("Config", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = pool.Config()
		}
	})
}

// BenchmarkPool_HookOperations benchmarks pool hook operations
func BenchmarkPool_HookOperations(b *testing.B) {
	pool := setupBenchmarkPool(b)
	defer pool.Close()
	ctx := context.Background()

	b.Run("BeforeAcquireHook", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = pool.BeforeAcquireHook(ctx)
		}
	})

	b.Run("AfterReleaseHook", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = pool.AfterReleaseHook(ctx)
		}
	})

	b.Run("GetConnWithNotPresent", func(b *testing.B) {
		ctrl := gomock.NewController(&testing.T{})
		defer ctrl.Finish()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			conn, releaseFunc, _ := pool.GetConnWithNotPresent(ctx, nil)
			if releaseFunc != nil {
				releaseFunc()
			}
			if conn != nil {
				conn.Release(ctx)
			}
		}
	})
}

// BenchmarkRows_Operations benchmarks rows operations
func BenchmarkRows_Operations(b *testing.B) {
	b.Run("Next_WithNilRows", func(b *testing.B) {
		rows := &Rows{rows: nil}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = rows.Next()
		}
	})

	b.Run("Scan_WithNilRows", func(b *testing.B) {
		rows := &Rows{rows: nil}
		var dest interface{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = rows.Scan(&dest)
		}
	})

	b.Run("Close_WithNilRows", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rows := &Rows{rows: nil}
			rows.Close()
		}
	})

	b.Run("Err_WithNilRows", func(b *testing.B) {
		rows := &Rows{rows: nil}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = rows.Err()
		}
	})

	b.Run("Columns_WithNilRows", func(b *testing.B) {
		rows := &Rows{rows: nil}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = rows.Columns()
		}
	})

	b.Run("ColumnTypes_WithNilRows", func(b *testing.B) {
		rows := &Rows{rows: nil}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = rows.ColumnTypes()
		}
	})

	b.Run("RawValues", func(b *testing.B) {
		rows := &Rows{rows: nil}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = rows.RawValues()
		}
	})

	b.Run("Values_WithNilRows", func(b *testing.B) {
		rows := &Rows{rows: nil}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = rows.Values()
		}
	})
}

// BenchmarkRow_Operations benchmarks row operations
func BenchmarkRow_Operations(b *testing.B) {
	b.Run("Scan_WithNilDB", func(b *testing.B) {
		row := &Row{
			db:    nil,
			query: "SELECT 'benchmark'",
			args:  []interface{}{},
		}
		var dest string
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = row.Scan(&dest)
		}
	})

	b.Run("Err_WithNilDB", func(b *testing.B) {
		row := &Row{
			db:    nil,
			query: "SELECT 'benchmark'",
			args:  []interface{}{},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = row.Err()
		}
	})

	b.Run("Scan_WithValidDB", func(b *testing.B) {
		db := setupBenchmarkDB(b)
		row := &Row{
			db:    db,
			query: "SELECT 'benchmark'",
			args:  []interface{}{},
		}
		var dest string
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = row.Scan(&dest)
		}
	})

	b.Run("Err_WithValidDB", func(b *testing.B) {
		db := setupBenchmarkDB(b)
		row := &Row{
			db:    db,
			query: "SELECT 'benchmark'",
			args:  []interface{}{},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = row.Err()
		}
	})
}

// BenchmarkTransaction_Operations benchmarks transaction operations
func BenchmarkTransaction_Operations(b *testing.B) {
	tx := setupBenchmarkTransaction(b)
	ctx := context.Background()

	b.Run("QueryOne", func(b *testing.B) {
		var result string
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tx.QueryOne(ctx, &result, "SELECT 'benchmark'")
		}
	})

	b.Run("QueryAll", func(b *testing.B) {
		var results []string
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tx.QueryAll(ctx, &results, "SELECT 'benchmark'")
		}
	})

	b.Run("QueryCount", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tx.QueryCount(ctx, "SELECT COUNT(*) FROM (SELECT 1) t")
		}
	})

	b.Run("QueryRow", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tx.QueryRow(ctx, "SELECT 'benchmark'")
		}
	})

	b.Run("Exec", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tx.Exec(ctx, "SELECT 1")
		}
	})

	b.Run("Commit", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tx.Commit(ctx)
		}
	})

	b.Run("Rollback", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tx.Rollback(ctx)
		}
	})

	b.Run("Ping", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tx.Ping(ctx)
		}
	})
}

// BenchmarkTransaction_SavepointOperations benchmarks transaction savepoint operations
func BenchmarkTransaction_SavepointOperations(b *testing.B) {
	tx := setupBenchmarkTransaction(b)
	ctx := context.Background()
	savepointName := "bench_savepoint"

	b.Run("BeginSavepoint", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tx.BeginSavepoint(ctx, savepointName)
		}
	})

	b.Run("Savepoint", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tx.Savepoint(ctx, savepointName)
		}
	})

	b.Run("RollbackToSavepoint", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tx.RollbackToSavepoint(ctx, savepointName)
		}
	})

	b.Run("ReleaseSavepoint", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tx.ReleaseSavepoint(ctx, savepointName)
		}
	})

	b.Run("BeginTransaction", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tx.BeginTransaction(ctx)
		}
	})

	b.Run("BeginTransactionWithOptions", func(b *testing.B) {
		opts := postgresql.TxOptions{
			IsoLevel:   postgresql.IsoLevelReadCommitted,
			AccessMode: postgresql.AccessModeReadWrite,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tx.BeginTransactionWithOptions(ctx, opts)
		}
	})
}

// BenchmarkTransaction_HookOperations benchmarks transaction hook operations
func BenchmarkTransaction_HookOperations(b *testing.B) {
	tx := setupBenchmarkTransaction(b)
	ctx := context.Background()

	b.Run("BeforeReleaseHook", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tx.BeforeReleaseHook(ctx)
		}
	})

	b.Run("AfterAcquireHook", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tx.AfterAcquireHook(ctx)
		}
	})

	b.Run("Release", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tx.Release(ctx)
		}
	})

	b.Run("Prepare", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tx.Prepare(ctx, "bench_stmt", "SELECT $1")
		}
	})
}

// setupBenchmarkConn creates a benchmark connection
func setupBenchmarkConn(b *testing.B) *Conn {
	db := setupBenchmarkDB(b)

	cfg := &config.Config{
		Host:     "localhost",
		Port:     5432,
		Database: "benchmark",
		Username: "benchmark",
		Password: "benchmark",
	}

	return &Conn{
		db:       db,
		config:   cfg,
		released: false,
	}
}

// setupBenchmarkPool creates a benchmark pool
func setupBenchmarkPool(b *testing.B) *Pool {
	db := setupBenchmarkDB(b)

	cfg := &config.Config{
		Host:     "localhost",
		Port:     5432,
		Database: "benchmark",
		Username: "benchmark",
		Password: "benchmark",
	}

	pool := &Pool{
		db:     db,
		config: cfg,
		closed: false,
	}

	// Try to get the underlying sql.DB for stats
	if db != nil {
		if sqlDB, err := db.DB(); err == nil {
			pool.sqlDB = sqlDB
		}
	}

	return pool
}

// setupBenchmarkDB creates a benchmark GORM DB
func setupBenchmarkDB(b *testing.B) *gorm.DB {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=invalid port=0 user=benchmark dbname=benchmark sslmode=disable",
	}), &gorm.Config{
		Logger: logger.Discard, // Disable logging for benchmarks
	})

	// If there's an error (expected with invalid DSN), return nil for safe benchmarking
	if err != nil {
		return nil
	}

	return db
}

// setupBenchmarkTransaction creates a benchmark transaction
func setupBenchmarkTransaction(b *testing.B) *Transaction {
	cfg := &config.Config{
		Host:     "localhost",
		Port:     5432,
		Database: "benchmark",
		Username: "benchmark",
		Password: "benchmark",
	}

	// Use nil tx to avoid panics in benchmarks
	return &Transaction{
		tx:     nil,
		config: cfg,
	}
}
