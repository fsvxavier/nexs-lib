//go:build unit
// +build unit

package gorm

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

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
				Driver:          interfaces.DriverGORM,
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
			gormProvider, ok := provider.(*Provider)
			if !ok {
				t.Error("NewProvider() did not return a GORM provider")
			}

			// Verify configuration is set
			if gormProvider.cfg != tt.config {
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
				Driver:          interfaces.DriverGORM,
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
		Driver:          interfaces.DriverGORM,
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
		Driver:          interfaces.DriverGORM,
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
	if gormDB, ok := db.(*gorm.DB); ok && gormDB != nil {
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
		Driver:          interfaces.DriverGORM,
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

	// Test Ping (should be no-op)
	err = tx.Ping(context.Background())
	if err != nil {
		t.Errorf("Ping() error = %v", err)
	}

	// Test Release (should be no-op)
	tx.Release(context.Background())
}

func TestConn_UnsupportedOperations(t *testing.T) {
	conn := &Conn{}

	// Test Query (should return error)
	_, err := conn.Query(context.Background(), "SELECT 1")
	if err == nil {
		t.Error("Query() should return error for unsupported operation")
	}

	// Test SendBatch (should return error)
	_, err = conn.SendBatch(context.Background(), nil)
	if err == nil {
		t.Error("SendBatch() should return error for unsupported operation")
	}

	// Test QueryRows (should return error)
	_, err = conn.QueryRows(context.Background(), "SELECT 1")
	if err == nil {
		t.Error("QueryRows() should return error for unsupported operation")
	}
}

func TestTransaction_UnsupportedOperations(t *testing.T) {
	tx := &Transaction{}

	// Test Query (should return error)
	_, err := tx.Query(context.Background(), "SELECT 1")
	if err == nil {
		t.Error("Query() should return error for unsupported operation")
	}

	// Test SendBatch (should return error)
	_, err = tx.SendBatch(context.Background(), nil)
	if err == nil {
		t.Error("SendBatch() should return error for unsupported operation")
	}

	// Test QueryRows (should return error)
	_, err = tx.QueryRows(context.Background(), "SELECT 1")
	if err == nil {
		t.Error("QueryRows() should return error for unsupported operation")
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
		Driver:          interfaces.DriverGORM,
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
