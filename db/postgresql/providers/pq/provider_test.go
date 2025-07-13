//go:build unit
// +build unit

package pq

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/interfaces"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.Config
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid configuration",
			config: &config.Config{
				Host:            "localhost",
				Port:            5432,
				User:            "postgres",
				Password:        "password",
				Database:        "testdb",
				SSLMode:         "disable",
				Driver:          interfaces.DriverPQ,
				MaxOpenConns:    25,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
				ConnMaxIdleTime: 5 * time.Minute,
				QueryTimeout:    30 * time.Second,
				ConnectTimeout:  10 * time.Second,
			},
			wantError: false,
		},
		{
			name:      "nil configuration",
			config:    nil,
			wantError: true,
			errorMsg:  "config cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.config)

			if tt.wantError {
				if err == nil {
					t.Errorf("NewProvider() error = nil, wantError %v", tt.wantError)
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("NewProvider() error = %v, want %v", err.Error(), tt.errorMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewProvider() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if provider == nil {
				t.Error("NewProvider() returned nil provider")
			}

			// Verify provider type
			pqProvider, ok := provider.(*Provider)
			if !ok {
				t.Error("NewProvider() did not return a PQ provider")
			}

			// Verify configuration is set
			if pqProvider.cfg != tt.config {
				t.Error("NewProvider() did not set configuration correctly")
			}
		})
	}
}

func TestProvider_Connect_InvalidConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.Config
		wantError bool
	}{
		{
			name: "invalid connection string",
			config: &config.Config{
				Host:            "",
				Port:            5432,
				User:            "postgres",
				Password:        "password",
				Database:        "testdb",
				SSLMode:         "disable",
				Driver:          interfaces.DriverPQ,
				MaxOpenConns:    25,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
				ConnMaxIdleTime: 5 * time.Minute,
				QueryTimeout:    30 * time.Second,
				ConnectTimeout:  100 * time.Millisecond, // Very short timeout
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.config)
			if err != nil {
				t.Fatalf("NewProvider() error = %v", err)
			}

			err = provider.Connect()
			if tt.wantError {
				if err == nil {
					t.Errorf("Connect() error = nil, wantError %v", tt.wantError)
				}
			} else {
				if err != nil {
					t.Errorf("Connect() error = %v, wantError %v", err, tt.wantError)
				}
			}

			// Clean up
			provider.Close()
		})
	}
}

func TestProvider_Close(t *testing.T) {
	cfg := &config.Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "password",
		Database:        "testdb",
		SSLMode:         "disable",
		Driver:          interfaces.DriverPQ,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		QueryTimeout:    30 * time.Second,
		ConnectTimeout:  10 * time.Second,
	}

	provider, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	// Test close without connection
	err = provider.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Test multiple closes
	err = provider.Close()
	if err != nil {
		t.Errorf("Close() second call error = %v", err)
	}

	// Test connect after close
	err = provider.Connect()
	if err == nil {
		t.Error("Connect() after close should return error")
	}
}

func TestProvider_DB(t *testing.T) {
	cfg := &config.Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "password",
		Database:        "testdb",
		SSLMode:         "disable",
		Driver:          interfaces.DriverPQ,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		QueryTimeout:    30 * time.Second,
		ConnectTimeout:  10 * time.Second,
	}

	provider, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	defer provider.Close()

	// Test DB() before connection (db should be nil)
	db := provider.DB()
	if sqlDB, ok := db.(*sql.DB); ok && sqlDB != nil {
		t.Error("DB() should return nil before connection")
	}
}

func TestProvider_Pool(t *testing.T) {
	cfg := &config.Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "password",
		Database:        "testdb",
		SSLMode:         "disable",
		Driver:          interfaces.DriverPQ,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		QueryTimeout:    30 * time.Second,
		ConnectTimeout:  10 * time.Second,
	}

	provider, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	defer provider.Close()

	// Test Pool() before connection
	pool := provider.Pool()
	if pool != nil {
		t.Error("Pool() should return nil before connection")
	}
}

func TestNewBatch(t *testing.T) {
	batch := NewBatch()
	if batch == nil {
		t.Error("NewBatch() returned nil")
	}

	// Test initial length
	if batch.Len() != 0 {
		t.Errorf("NewBatch() Len() = %d, want 0", batch.Len())
	}

	// Test adding queries
	batch.Queue("SELECT 1", nil)
	if batch.Len() != 1 {
		t.Errorf("After Queue() Len() = %d, want 1", batch.Len())
	}

	batch.Queue("SELECT 2", "arg1")
	if batch.Len() != 2 {
		t.Errorf("After second Queue() Len() = %d, want 2", batch.Len())
	}
}

func TestConn_Methods(t *testing.T) {
	// This is a basic test that verifies the Conn struct exists and has the required methods
	// In a real test, you'd need a mock or actual database connection
	conn := &Conn{}
	if conn == nil {
		t.Error("Conn struct creation failed")
	}

	// Test BeforeReleaseHook
	err := conn.BeforeReleaseHook(context.Background())
	if err != nil {
		t.Errorf("BeforeReleaseHook() error = %v", err)
	}

	// Test AfterAcquireHook without multi-tenant
	conn.multiTenantEnabled = false
	err = conn.AfterAcquireHook(context.Background())
	if err != nil {
		t.Errorf("AfterAcquireHook() error = %v", err)
	}

	// Test AfterAcquireHook with multi-tenant
	conn.multiTenantEnabled = true
	err = conn.AfterAcquireHook(context.Background())
	if err != nil {
		t.Errorf("AfterAcquireHook() with multi-tenant error = %v", err)
	}

	// Test getContext
	ctx := context.Background()
	resultCtx := conn.getContext(ctx)
	if resultCtx != ctx {
		t.Error("getContext() should return passed context")
	}

	// Test getContext with nil
	conn.ctx = ctx
	resultCtx = conn.getContext(nil)
	if resultCtx != ctx {
		t.Error("getContext() should return conn context when passed nil")
	}
}

func TestTransaction_Methods(t *testing.T) {
	// This is a basic test that verifies the Transaction struct exists and has the required methods
	// In a real test, you'd need a mock or actual database connection
	tx := &Transaction{}
	if tx == nil {
		t.Error("Transaction struct creation failed")
	}

	// Test BeforeReleaseHook
	err := tx.BeforeReleaseHook(context.Background())
	if err != nil {
		t.Errorf("BeforeReleaseHook() error = %v", err)
	}

	// Test AfterAcquireHook
	err = tx.AfterAcquireHook(context.Background())
	if err != nil {
		t.Errorf("AfterAcquireHook() error = %v", err)
	}

	// Test nested transactions (should fail)
	_, err = tx.BeginTransaction(context.Background())
	if err == nil {
		t.Error("BeginTransaction() should return error for nested transactions")
	}

	_, err = tx.BeginTransactionWithOptions(context.Background(), interfaces.TxOptions{})
	if err == nil {
		t.Error("BeginTransactionWithOptions() should return error for nested transactions")
	}

	// Test Ping (should be no-op)
	err = tx.Ping(context.Background())
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}

	// Test Release (should be no-op)
	tx.Release(context.Background())
}

func TestBatchResults_Methods(t *testing.T) {
	// This is a basic test that verifies the BatchResults struct exists and has the required methods
	// In a real test, you'd need a mock or actual database connection
	br := &BatchResults{
		queries: []batchQuery{
			{query: "SELECT 1", args: nil},
			{query: "SELECT 2", args: []interface{}{"arg1"}},
		},
		current: 0,
	}

	if br == nil {
		t.Error("BatchResults struct creation failed")
	}

	// Test QueryOne without connection (should fail)
	err := br.QueryOne(nil)
	if err == nil {
		t.Error("QueryOne() should fail without connection")
	}

	// Test QueryAll without connection (should fail)
	err = br.QueryAll(nil)
	if err == nil {
		t.Error("QueryAll() should fail without connection")
	}

	// Test Exec without connection (should fail)
	err = br.Exec()
	if err == nil {
		t.Error("Exec() should fail without connection")
	}

	// Test Close
	br.Close()

	// Test beyond queries
	br.current = 10
	err = br.QueryOne(nil)
	if err == nil {
		t.Error("QueryOne() should fail when beyond queries")
	}

	err = br.QueryAll(nil)
	if err == nil {
		t.Error("QueryAll() should fail when beyond queries")
	}

	err = br.Exec()
	if err == nil {
		t.Error("Exec() should fail when beyond queries")
	}
}

func TestRow_Methods(t *testing.T) {
	// This is a basic test that verifies the Row struct exists and has the Scan method
	// In a real test, you'd need a mock or actual database connection
	row := &Row{}
	if row == nil {
		t.Error("Row struct creation failed")
	}
}

func TestRows_Methods(t *testing.T) {
	// This is a basic test that verifies the Rows struct exists and has the required methods
	// In a real test, you'd need a mock or actual database connection
	rows := &Rows{}
	if rows == nil {
		t.Error("Rows struct creation failed")
	}

	// Test RawValues (should return nil for lib/pq)
	rawValues := rows.RawValues()
	if rawValues != nil {
		t.Error("RawValues() should return nil for lib/pq")
	}
}

func TestPool_GetConnWithNotPresent(t *testing.T) {
	pool := &Pool{}

	// Test with existing connection
	mockConn := &Conn{}
	conn, releaseFunc, err := pool.GetConnWithNotPresent(context.Background(), mockConn)
	if err != nil {
		t.Errorf("GetConnWithNotPresent() with existing conn error = %v", err)
	}
	if conn != mockConn {
		t.Error("GetConnWithNotPresent() should return existing connection")
	}
	if releaseFunc == nil {
		t.Error("GetConnWithNotPresent() should return release function")
	}

	// Test release function
	releaseFunc()
}

func TestPool_Stats(t *testing.T) {
	pool := &Pool{}

	// Test Stats() with nil db (should return empty stats)
	stats := pool.Stats()
	if stats.MaxConns != 0 || stats.TotalConns != 0 {
		t.Error("Stats() should return empty stats when db is nil")
	}
}

// Benchmark tests
func BenchmarkNewProvider(b *testing.B) {
	cfg := &config.Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "password",
		Database:        "testdb",
		SSLMode:         "disable",
		Driver:          interfaces.DriverPQ,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		QueryTimeout:    30 * time.Second,
		ConnectTimeout:  10 * time.Second,
	}

	for i := 0; i < b.N; i++ {
		provider, err := NewProvider(cfg)
		if err != nil {
			b.Fatalf("NewProvider() error = %v", err)
		}
		provider.Close()
	}
}

func BenchmarkNewBatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		batch := NewBatch()
		batch.Queue("SELECT 1")
	}
}
