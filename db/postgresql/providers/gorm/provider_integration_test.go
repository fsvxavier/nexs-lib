//go:build integration
// +build integration

package gorm

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProvider_Integration_Real_Database tests with real database connection
func TestProvider_Integration_Real_Database(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := context.Background()
	provider := NewProvider()

	// Use environment variables or test database config
	cfg := &config.Config{
		Host:               "localhost",
		Port:               5432,
		Database:           "test_db",
		Username:           "test_user",
		Password:           "test_pass",
		MaxConns:           5,
		MinConns:           1,
		MaxConnLifetime:    time.Hour,
		MaxConnIdleTime:    time.Minute * 30,
		ConnectTimeout:     time.Second * 10,
		QueryTimeout:       time.Second * 30,
		ApplicationName:    "gorm-provider-test",
		SearchPath:         []string{"public"},
		Timezone:           "UTC",
		DefaultSchema:      "public",
		MultiTenantEnabled: false,
	}

	t.Run("Real_Pool_Creation", func(t *testing.T) {
		pool, err := provider.CreatePool(ctx, cfg)
		if err != nil {
			t.Skipf("Cannot connect to test database: %v", err)
			return
		}
		defer pool.Close()

		// Test pool operations
		stats := pool.Stats()
		assert.GreaterOrEqual(t, stats.MaxConns, int32(1))

		// Test ping
		err = pool.Ping(ctx)
		assert.NoError(t, err)

		// Test health check
		err = pool.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("Real_Connection_Operations", func(t *testing.T) {
		pool, err := provider.CreatePool(ctx, cfg)
		if err != nil {
			t.Skipf("Cannot connect to test database: %v", err)
			return
		}
		defer pool.Close()

		conn, err := pool.Acquire(ctx)
		require.NoError(t, err)
		defer conn.Release(ctx)

		// Test simple query
		var result int
		row, err := conn.QueryRow(ctx, "SELECT 1")
		require.NoError(t, err)

		err = row.Scan(&result)
		assert.NoError(t, err)
		assert.Equal(t, 1, result)
	})

	t.Run("Real_Transaction_Operations", func(t *testing.T) {
		pool, err := provider.CreatePool(ctx, cfg)
		if err != nil {
			t.Skipf("Cannot connect to test database: %v", err)
			return
		}
		defer pool.Close()

		conn, err := pool.Acquire(ctx)
		require.NoError(t, err)
		defer conn.Release(ctx)

		// Test transaction
		tx, err := conn.BeginTransaction(ctx)
		require.NoError(t, err)

		// Simple transaction operation
		err = tx.Exec(ctx, "SELECT 1")
		assert.NoError(t, err)

		// Rollback transaction
		err = tx.Rollback(ctx)
		assert.NoError(t, err)
	})
}
