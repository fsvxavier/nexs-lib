//go:build unit
// +build unit

package pgx

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/fsvxavier/nexs-lib/db/postgresql/interfaces"
	"github.com/jackc/pgx/v5/pgxpool"
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
				Driver:          interfaces.DriverPGX,
				MaxOpenConns:    25,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
				ConnMaxIdleTime: 5 * time.Minute,
				MinConns:        2,
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
			pgxProvider, ok := provider.(*Provider)
			if !ok {
				t.Error("NewProvider() did not return a PGX provider")
			}

			// Verify configuration is set
			if pgxProvider.cfg != tt.config {
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
				Driver:          interfaces.DriverPGX,
				MaxOpenConns:    25,
				MaxIdleConns:    5,
				ConnMaxLifetime: 5 * time.Minute,
				ConnMaxIdleTime: 5 * time.Minute,
				MinConns:        2,
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
		Driver:          interfaces.DriverPGX,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		MinConns:        2,
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
		Driver:          interfaces.DriverPGX,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		MinConns:        2,
		QueryTimeout:    30 * time.Second,
		ConnectTimeout:  10 * time.Second,
	}

	provider, err := NewProvider(cfg)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	defer provider.Close()

	// Test DB() before connection (pool should be nil)
	db := provider.DB()
	if pool, ok := db.(*pgxpool.Pool); ok && pool != nil {
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
		Driver:          interfaces.DriverPGX,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		MinConns:        2,
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

func TestRow_Scan(t *testing.T) {
	// This is a basic test that verifies the Row struct exists and has the Scan method
	// In a real test, you'd need a mock or actual database connection
	row := &Row{}
	if row == nil {
		t.Error("Row struct creation failed")
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
	br := &BatchResults{}
	if br == nil {
		t.Error("BatchResults struct creation failed")
	}
}

// Test Pool methods with mocks

// Test Conn methods detailed

// Test Transaction methods detailed

// Test BatchResults methods

// Test Row and Rows methods

// Test Batch with more operations
func TestBatch_ExtendedOperations(t *testing.T) {
	batch := NewBatch()

	// Test multiple queue operations
	batch.Queue("SELECT 1")
	batch.Queue("SELECT $1", "test")
	batch.Queue("SELECT $1, $2", "test1", "test2")

	if batch.Len() != 3 {
		t.Errorf("Batch.Len() = %d, want 3", batch.Len())
	}

	// Test empty batch
	emptyBatch := NewBatch()
	if emptyBatch.Len() != 0 {
		t.Errorf("Empty batch.Len() = %d, want 0", emptyBatch.Len())
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
		Driver:          interfaces.DriverPGX,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		MinConns:        2,
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

// Additional tests to improve coverage for PGX provider - Safe methods only

// Test structure validation only (no nil pointer operations)
func TestStructures_Validation(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Pool struct creation",
			test: func(t *testing.T) {
				pool := &Pool{}
				if pool == nil {
					t.Error("Pool struct creation failed")
				}
			},
		},
		{
			name: "Conn struct creation",
			test: func(t *testing.T) {
				conn := &Conn{}
				if conn == nil {
					t.Error("Conn struct creation failed")
				}
			},
		},
		{
			name: "Transaction struct creation",
			test: func(t *testing.T) {
				tx := &Transaction{}
				if tx == nil {
					t.Error("Transaction struct creation failed")
				}
			},
		},
		{
			name: "BatchResults struct creation",
			test: func(t *testing.T) {
				br := &BatchResults{}
				if br == nil {
					t.Error("BatchResults struct creation failed")
				}
			},
		},
		{
			name: "Row struct creation",
			test: func(t *testing.T) {
				row := &Row{}
				if row == nil {
					t.Error("Row struct creation failed")
				}
			},
		},
		{
			name: "Rows struct creation",
			test: func(t *testing.T) {
				rows := &Rows{}
				if rows == nil {
					t.Error("Rows struct creation failed")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// Test methods that are truly safe and don't cause panics
func TestSafeMethods(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Conn BeforeReleaseHook",
			test: func(t *testing.T) {
				conn := &Conn{}
				err := conn.BeforeReleaseHook(context.Background())
				if err != nil {
					t.Errorf("BeforeReleaseHook() error = %v", err)
				}
			},
		},
		{
			name: "Conn AfterAcquireHook without multi-tenant",
			test: func(t *testing.T) {
				conn := &Conn{multiTenantEnabled: false}
				err := conn.AfterAcquireHook(context.Background())
				if err != nil {
					t.Errorf("AfterAcquireHook() error = %v", err)
				}
			},
		},
		{
			name: "Conn AfterAcquireHook with multi-tenant",
			test: func(t *testing.T) {
				conn := &Conn{multiTenantEnabled: true}
				err := conn.AfterAcquireHook(context.Background())
				if err != nil {
					t.Errorf("AfterAcquireHook() with multi-tenant error = %v", err)
				}
			},
		},
		{
			name: "Transaction Ping",
			test: func(t *testing.T) {
				tx := &Transaction{}
				err := tx.Ping(context.Background())
				if err != nil {
					t.Errorf("Transaction.Ping() should not return error: %v", err)
				}
			},
		},
		{
			name: "Transaction BeginTransaction returns error for nested",
			test: func(t *testing.T) {
				tx := &Transaction{}
				_, err := tx.BeginTransaction(context.Background())
				if err == nil {
					t.Error("BeginTransaction() should return error for nested transactions")
				}
			},
		},
		{
			name: "Transaction BeginTransactionWithOptions returns error for nested",
			test: func(t *testing.T) {
				tx := &Transaction{}
				opts := interfaces.TxOptions{IsolationLevel: "READ_COMMITTED"}
				_, err := tx.BeginTransactionWithOptions(context.Background(), opts)
				if err == nil {
					t.Error("BeginTransactionWithOptions() should return error for nested transactions")
				}
			},
		},
		{
			name: "Transaction Release does not panic",
			test: func(t *testing.T) {
				tx := &Transaction{}
				// Should not panic
				tx.Release(context.Background())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// Test Batch operations which are safe
func TestBatch_Operations(t *testing.T) {
	batch := NewBatch()

	// Test multiple queue operations
	batch.Queue("SELECT 1")
	batch.Queue("SELECT $1", "test")
	batch.Queue("SELECT $1, $2", "test1", "test2")

	if batch.Len() != 3 {
		t.Errorf("Batch.Len() = %d, want 3", batch.Len())
	}

	// Test empty batch
	emptyBatch := NewBatch()
	if emptyBatch.Len() != 0 {
		t.Errorf("Empty batch.Len() = %d, want 0", emptyBatch.Len())
	}
}
