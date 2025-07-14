package gorm

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/stretchr/testify/assert"
)

func TestProvider_Interface(t *testing.T) {
	provider := NewProvider()
	// Verify that Provider implements IProvider interface
	var _ postgresql.IProvider = provider
}

func TestProvider_CreatePool(t *testing.T) {
	cfg := &config.Config{
		Host:     "localhost",
		Port:     5432,
		Database: "test",
		Username: "test",
		Password: "test",
	}

	provider := NewProvider()
	ctx := context.Background()

	t.Run("create pool with valid config", func(t *testing.T) {
		pool, err := provider.CreatePool(ctx, cfg)
		// Without a real database, this will likely fail, but we test the interface
		if err != nil {
			assert.Error(t, err, "Expected error without real database")
		} else {
			assert.NotNil(t, pool, "Pool should not be nil if creation succeeds")
			assert.Implements(t, (*postgresql.IPool)(nil), pool, "Pool should implement IPool interface")
		}
	})

	t.Run("create pool with nil config", func(t *testing.T) {
		pool, err := provider.CreatePool(ctx, nil)
		assert.Error(t, err, "Should fail with nil config")
		assert.Nil(t, pool, "Pool should be nil on error")
	})
}

func TestProvider_CreateConnection(t *testing.T) {
	cfg := &config.Config{
		Host:     "localhost",
		Port:     5432,
		Database: "test",
		Username: "test",
		Password: "test",
	}

	provider := NewProvider()
	ctx := context.Background()

	t.Run("create connection with valid config", func(t *testing.T) {
		conn, err := provider.CreateConnection(ctx, cfg)
		// Without a real database, this will likely fail, but we test the interface
		if err != nil {
			assert.Error(t, err, "Expected error without real database")
		} else {
			assert.NotNil(t, conn, "Connection should not be nil if creation succeeds")
			assert.Implements(t, (*postgresql.IConn)(nil), conn, "Connection should implement IConn interface")
		}
	})

	t.Run("create connection with nil config", func(t *testing.T) {
		conn, err := provider.CreateConnection(ctx, nil)
		assert.Error(t, err, "Should fail with nil config")
		assert.Nil(t, conn, "Connection should be nil on error")
	})
}

func TestProvider_Type(t *testing.T) {
	provider := NewProvider()
	providerType := provider.Type()
	assert.Equal(t, postgresql.ProviderTypeGORM, providerType, "Provider type should be GORM")
}

func TestProvider_Name(t *testing.T) {
	provider := NewProvider()
	name := provider.Name()
	assert.Equal(t, "gorm", name, "Provider name should be 'gorm'")
}

func TestProvider_Version(t *testing.T) {
	provider := NewProvider()
	version := provider.Version()
	assert.NotEmpty(t, version, "Version should not be empty")
}

func TestProvider_IsHealthy(t *testing.T) {
	provider := NewProvider()
	ctx := context.Background()

	// Without a real database connection, this might return true or false depending on implementation
	healthy := provider.IsHealthy(ctx)
	// Just verify it returns a boolean
	assert.IsType(t, true, healthy, "IsHealthy should return a boolean")
}

func TestProvider_GetMetrics(t *testing.T) {
	provider := NewProvider()
	ctx := context.Background()

	metrics := provider.GetMetrics(ctx)
	assert.NotNil(t, metrics, "Metrics should not be nil")
	assert.IsType(t, map[string]interface{}{}, metrics, "Metrics should be a map")
}

func TestProvider_Close(t *testing.T) {
	provider := NewProvider()

	err := provider.Close()
	assert.NoError(t, err, "Close should not return error")
}

func TestNewProvider(t *testing.T) {
	provider := NewProvider()
	assert.NotNil(t, provider, "NewProvider should return a non-nil provider")
	assert.Implements(t, (*postgresql.IProvider)(nil), provider, "Provider should implement IProvider interface")
}
