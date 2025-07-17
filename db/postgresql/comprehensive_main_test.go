//go:build unit
// +build unit

package postgresql

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

func TestDefaultFactoryOperations(t *testing.T) {
	factory := GetDefaultFactory()

	t.Run("GetDefaultFactory returns singleton", func(t *testing.T) {
		factory2 := GetDefaultFactory()
		if factory != factory2 {
			t.Error("GetDefaultFactory should return the same instance")
		}
	})

	t.Run("Create PGX through default factory", func(t *testing.T) {
		provider, err := factory.CreateProvider(interfaces.ProviderTypePGX)
		if err != nil {
			t.Errorf("CreateProvider() error = %v", err)
		}

		if provider == nil {
			t.Error("CreateProvider() returned nil")
		}
	})

	t.Run("Register and retrieve custom provider", func(t *testing.T) {
		customProvider, _ := NewPostgreSQLProvider(interfaces.ProviderTypePGX)
		customType := interfaces.ProviderType("custom-test")

		err := factory.RegisterProvider(customType, customProvider)
		if err != nil {
			t.Errorf("RegisterProvider() error = %v", err)
		}

		retrieved, exists := factory.GetProvider(customType)
		if !exists {
			t.Error("GetProvider() should return true for registered provider")
		}

		if retrieved != customProvider {
			t.Error("GetProvider() should return the same instance")
		}

		// Test listing includes custom provider
		providers := factory.ListProviders()
		found := false
		for _, p := range providers {
			if p == customType {
				found = true
				break
			}
		}
		if !found {
			t.Error("ListProviders() should include registered custom provider")
		}
	})

	t.Run("Register nil provider", func(t *testing.T) {
		err := factory.RegisterProvider("nil-test", nil)
		if err == nil {
			t.Error("RegisterProvider() should return error for nil provider")
		}
	})
}

func TestConfigurationHelpers(t *testing.T) {
	t.Run("NewDefaultConfig helper", func(t *testing.T) {
		connStr := "postgres://test:test@localhost/test"
		cfg := NewDefaultConfig(connStr)

		if cfg == nil {
			t.Error("NewDefaultConfig() should not return nil")
		}

		if cfg.GetConnectionString() != connStr {
			t.Errorf("Expected connection string %s, got %s", connStr, cfg.GetConnectionString())
		}
	})

	t.Run("Configuration option helpers", func(t *testing.T) {
		cfg := NewDefaultConfig("postgres://localhost/test")

		// Test type assertion for applying options
		if defaultCfg, ok := cfg.(*config.DefaultConfig); ok {
			err := defaultCfg.ApplyOptions(
				WithMaxConns(100),
				WithMinConns(10),
				WithMaxConnLifetime(time.Hour),
				WithMultiTenant(true),
				WithMaxRetries(5),
			)

			if err != nil {
				t.Errorf("ApplyOptions() error = %v", err)
			}

			poolConfig := cfg.GetPoolConfig()
			if poolConfig.MaxConns != 100 {
				t.Errorf("Expected MaxConns 100, got %d", poolConfig.MaxConns)
			}

			if poolConfig.MinConns != 10 {
				t.Errorf("Expected MinConns 10, got %d", poolConfig.MinConns)
			}

			if !cfg.IsMultiTenantEnabled() {
				t.Error("Multi-tenancy should be enabled")
			}

			retryConfig := cfg.GetRetryConfig()
			if retryConfig.MaxRetries != 5 {
				t.Errorf("Expected MaxRetries 5, got %d", retryConfig.MaxRetries)
			}
		} else {
			t.Error("Config should be of type *config.DefaultConfig")
		}
	})
}

func TestProviderSpecificMethods(t *testing.T) {
	provider, err := NewPostgreSQLProvider(interfaces.ProviderTypePGX)
	if err != nil {
		t.Fatalf("NewPostgreSQLProvider() error = %v", err)
	}

	t.Run("Schema operations", func(t *testing.T) {
		mockConn := &mockConnection{
			execResults:     make(map[string]interfaces.CommandTag),
			queryAllResults: make(map[string]interface{}),
		}

		// Mock successful exec for schema operations
		mockConn.execResults["CREATE SCHEMA"] = &mockCommandTag{rowsAffected: 0}
		mockConn.execResults["DROP SCHEMA"] = &mockCommandTag{rowsAffected: 0}

		// Mock schema list
		schemas := []string{"public", "test_schema"}
		mockConn.queryAllResults["SELECT schema_name"] = &schemas

		ctx := context.Background()

		// Test CreateSchema
		err := provider.CreateSchema(ctx, mockConn, "test_schema")
		if err != nil {
			t.Errorf("CreateSchema() error = %v", err)
		}

		// Test CreateSchema with empty name
		err = provider.CreateSchema(ctx, mockConn, "")
		if err == nil {
			t.Error("CreateSchema() should return error for empty schema name")
		}

		// Test DropSchema
		err = provider.DropSchema(ctx, mockConn, "test_schema")
		if err != nil {
			t.Errorf("DropSchema() error = %v", err)
		}

		// Test DropSchema with empty name
		err = provider.DropSchema(ctx, mockConn, "")
		if err == nil {
			t.Error("DropSchema() should return error for empty schema name")
		}

		// Test ListSchemas
		resultSchemas, err := provider.ListSchemas(ctx, mockConn)
		if err != nil {
			t.Errorf("ListSchemas() error = %v", err)
		}

		if len(resultSchemas) != 2 {
			t.Errorf("Expected 2 schemas, got %d", len(resultSchemas))
		}
	})

	t.Run("Database operations", func(t *testing.T) {
		mockConn := &mockConnection{
			execResults:     make(map[string]interfaces.CommandTag),
			queryAllResults: make(map[string]interface{}),
		}

		// Mock successful exec for database operations
		mockConn.execResults["CREATE DATABASE"] = &mockCommandTag{rowsAffected: 0}
		mockConn.execResults["DROP DATABASE"] = &mockCommandTag{rowsAffected: 0}

		// Mock database list
		databases := []string{"postgres", "testdb"}
		mockConn.queryAllResults["SELECT datname"] = &databases

		ctx := context.Background()

		// Test CreateDatabase
		err := provider.CreateDatabase(ctx, mockConn, "new_db")
		if err != nil {
			t.Errorf("CreateDatabase() error = %v", err)
		}

		// Test CreateDatabase with empty name
		err = provider.CreateDatabase(ctx, mockConn, "")
		if err == nil {
			t.Error("CreateDatabase() should return error for empty database name")
		}

		// Test DropDatabase
		err = provider.DropDatabase(ctx, mockConn, "old_db")
		if err != nil {
			t.Errorf("DropDatabase() error = %v", err)
		}

		// Test DropDatabase with empty name
		err = provider.DropDatabase(ctx, mockConn, "")
		if err == nil {
			t.Error("DropDatabase() should return error for empty database name")
		}

		// Test ListDatabases
		resultDatabases, err := provider.ListDatabases(ctx, mockConn)
		if err != nil {
			t.Errorf("ListDatabases() error = %v", err)
		}

		if len(resultDatabases) != 2 {
			t.Errorf("Expected 2 databases, got %d", len(resultDatabases))
		}
	})

	t.Run("Retry and Failover with managers", func(t *testing.T) {
		ctx := context.Background()

		// Test WithRetry with successful operation
		callCount := 0
		operation := func() error {
			callCount++
			return nil
		}

		err := provider.WithRetry(ctx, operation)
		if err != nil {
			t.Errorf("WithRetry() error = %v", err)
		}

		if callCount != 1 {
			t.Errorf("Expected operation to be called once, got %d", callCount)
		}

		// Test WithFailover (will fail due to incomplete implementation)
		failoverOp := func(conn interfaces.IConn) error {
			return nil
		}

		err = provider.WithFailover(ctx, failoverOp)
		if err == nil {
			t.Error("WithFailover() should return error for incomplete implementation")
		}
	})

	t.Run("Manager accessors", func(t *testing.T) {
		retryManager := provider.GetRetryManager()
		if retryManager == nil {
			t.Error("GetRetryManager() should not return nil")
		}

		failoverManager := provider.GetFailoverManager()
		if failoverManager == nil {
			t.Error("GetFailoverManager() should not return nil")
		}
	})
}

func TestNewPGXProviderHelper(t *testing.T) {
	provider, err := NewPGXProvider()
	if err != nil {
		t.Errorf("NewPGXProvider() error = %v", err)
	}

	if provider == nil {
		t.Error("NewPGXProvider() should not return nil")
	}

	if provider.GetDriverName() != "pgx" {
		t.Errorf("Expected driver name 'pgx', got %s", provider.GetDriverName())
	}
}

// Mock implementations for testing
type mockConnection struct {
	execResults     map[string]interfaces.CommandTag
	queryAllResults map[string]interface{}
}

func (m *mockConnection) QueryRow(ctx context.Context, query string, args ...interface{}) interfaces.IRow {
	return &mockRow{}
}

func (m *mockConnection) Query(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	return &mockRows{}, nil
}

func (m *mockConnection) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	return nil
}

func (m *mockConnection) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	// Find matching result based on query content
	for queryPattern, result := range m.queryAllResults {
		if indexOfSubstring(query, queryPattern) >= 0 {
			// Try to copy result to destination
			if schemas, ok := result.(*[]string); ok {
				if dstSchemas, ok := dst.(*[]string); ok {
					*dstSchemas = *schemas
					return nil
				}
			}
		}
	}
	return nil
}

func (m *mockConnection) QueryCount(ctx context.Context, query string, args ...interface{}) (int64, error) {
	return 0, nil
}

func (m *mockConnection) Exec(ctx context.Context, query string, args ...interface{}) (interfaces.CommandTag, error) {
	// Find matching result based on query content
	for queryPattern, result := range m.execResults {
		if indexOfSubstring(query, queryPattern) >= 0 {
			return result, nil
		}
	}
	return &mockCommandTag{}, nil
}

func (m *mockConnection) SendBatch(ctx context.Context, batch interfaces.IBatch) interfaces.IBatchResults {
	return &mockBatchResults{}
}

func (m *mockConnection) Begin(ctx context.Context) (interfaces.ITransaction, error) {
	return &mockTransaction{}, nil
}

func (m *mockConnection) BeginTx(ctx context.Context, txOptions interfaces.TxOptions) (interfaces.ITransaction, error) {
	return &mockTransaction{}, nil
}

func (m *mockConnection) Release() {}

func (m *mockConnection) Close(ctx context.Context) error {
	return nil
}

func (m *mockConnection) Ping(ctx context.Context) error {
	return nil
}

func (m *mockConnection) IsClosed() bool {
	return false
}

func (m *mockConnection) Prepare(ctx context.Context, name, query string) error {
	return nil
}

func (m *mockConnection) Deallocate(ctx context.Context, name string) error {
	return nil
}

func (m *mockConnection) CopyFrom(ctx context.Context, tableName string, columnNames []string, rowSrc interfaces.CopyFromSource) (int64, error) {
	return 0, nil
}

func (m *mockConnection) CopyTo(ctx context.Context, w interfaces.CopyToWriter, query string, args ...interface{}) error {
	return nil
}

func (m *mockConnection) Listen(ctx context.Context, channel string) error {
	return nil
}

func (m *mockConnection) Unlisten(ctx context.Context, channel string) error {
	return nil
}

func (m *mockConnection) WaitForNotification(ctx context.Context, timeout time.Duration) (*interfaces.Notification, error) {
	return nil, nil
}

func (m *mockConnection) SetTenant(ctx context.Context, tenantID string) error {
	return nil
}

func (m *mockConnection) GetTenant(ctx context.Context) (string, error) {
	return "", nil
}

func (m *mockConnection) GetHookManager() interfaces.HookManager {
	return nil
}

func (m *mockConnection) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *mockConnection) Stats() interfaces.ConnectionStats {
	return interfaces.ConnectionStats{}
}

type mockRow struct{}

func (m *mockRow) Scan(dest ...any) error {
	return nil
}

type mockRows struct{}

func (m *mockRows) Next() bool                                       { return false }
func (m *mockRows) Scan(dest ...any) error                           { return nil }
func (m *mockRows) Close() error                                     { return nil }
func (m *mockRows) Err() error                                       { return nil }
func (m *mockRows) CommandTag() interfaces.CommandTag                { return &mockCommandTag{} }
func (m *mockRows) FieldDescriptions() []interfaces.FieldDescription { return nil }
func (m *mockRows) RawValues() [][]byte                              { return nil }

type mockCommandTag struct {
	rowsAffected int64
}

func (m *mockCommandTag) String() string      { return "SELECT 0" }
func (m *mockCommandTag) RowsAffected() int64 { return m.rowsAffected }
func (m *mockCommandTag) Insert() bool        { return false }
func (m *mockCommandTag) Update() bool        { return false }
func (m *mockCommandTag) Delete() bool        { return false }
func (m *mockCommandTag) Select() bool        { return true }

type mockBatchResults struct{}

func (m *mockBatchResults) QueryRow() interfaces.IRow            { return &mockRow{} }
func (m *mockBatchResults) Query() (interfaces.IRows, error)     { return &mockRows{}, nil }
func (m *mockBatchResults) Exec() (interfaces.CommandTag, error) { return &mockCommandTag{}, nil }
func (m *mockBatchResults) Close() error                         { return nil }
func (m *mockBatchResults) Err() error                           { return nil }

type mockTransaction struct {
	*mockConnection
}

func (m *mockTransaction) Commit(ctx context.Context) error {
	return nil
}

func (m *mockTransaction) Rollback(ctx context.Context) error {
	return nil
}
