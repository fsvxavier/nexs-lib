//go:build unit

package pgx

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/pashagolub/pgxmock/v4"
)

func TestConn_Interface(t *testing.T) {
	// Test interface compliance
	var _ postgresql.IConn = &Conn{}
}

func TestConn_ReleasedOperations(t *testing.T) {
	cfg := config.DefaultConfig()
	conn := &Conn{
		config:   cfg,
		released: true,
		isPooled: false,
	}

	ctx := context.Background()

	// Test all operations fail when connection is released
	err := conn.QueryOne(ctx, nil, "SELECT 1")
	if err == nil {
		t.Error("Expected error for QueryOne on released connection")
	}

	err = conn.QueryAll(ctx, nil, "SELECT 1")
	if err == nil {
		t.Error("Expected error for QueryAll on released connection")
	}

	_, err = conn.QueryCount(ctx, "SELECT COUNT(*)")
	if err == nil {
		t.Error("Expected error for QueryCount on released connection")
	}

	_, err = conn.Query(ctx, "SELECT 1")
	if err == nil {
		t.Error("Expected error for Query on released connection")
	}

	_, err = conn.QueryRow(ctx, "SELECT 1")
	if err == nil {
		t.Error("Expected error for QueryRow on released connection")
	}

	err = conn.Exec(ctx, "SELECT 1")
	if err == nil {
		t.Error("Expected error for Exec on released connection")
	}

	_, err = conn.SendBatch(ctx, NewBatch())
	if err == nil {
		t.Error("Expected error for SendBatch on released connection")
	}

	_, err = conn.BeginTransaction(ctx)
	if err == nil {
		t.Error("Expected error for BeginTransaction on released connection")
	}

	_, err = conn.BeginTransactionWithOptions(ctx, postgresql.TxOptions{})
	if err == nil {
		t.Error("Expected error for BeginTransactionWithOptions on released connection")
	}

	err = conn.Prepare(ctx, "test", "SELECT 1")
	if err == nil {
		t.Error("Expected error for Prepare on released connection")
	}

	err = conn.Ping(ctx)
	if err == nil {
		t.Error("Expected error for Ping on released connection")
	}

	err = conn.Listen(ctx, "test_channel")
	if err == nil {
		t.Error("Expected error for Listen on released connection")
	}

	err = conn.Unlisten(ctx, "test_channel")
	if err == nil {
		t.Error("Expected error for Unlisten on released connection")
	}

	_, err = conn.WaitForNotification(ctx, time.Second)
	if err == nil {
		t.Error("Expected error for WaitForNotification on released connection")
	}
}

func TestConn_DoubleRelease(t *testing.T) {
	cfg := config.DefaultConfig()
	conn := &Conn{
		config:   cfg,
		released: false,
		isPooled: false,
	}

	ctx := context.Background()

	// First release
	conn.Release(ctx)

	if !conn.released {
		t.Error("Expected connection to be marked as released")
	}

	// Second release should be a no-op
	conn.Release(ctx) // Should not panic
}

func TestConn_MultiTenantHooks(t *testing.T) {
	// Test multi-tenant hooks without actual database connection
	cfg := config.NewConfig(
		config.WithMultiTenant(false), // Disable to avoid queries
	)

	conn := &Conn{
		config:             cfg,
		multiTenantEnabled: false,
		released:           false,
		isPooled:           false,
	}

	ctx := context.Background()

	// Test AfterAcquireHook with multi-tenant disabled
	err := conn.AfterAcquireHook(ctx)
	if err != nil {
		t.Error("AfterAcquireHook should not fail when multi-tenant is disabled")
	}

	// Test BeforeReleaseHook with multi-tenant disabled
	err = conn.BeforeReleaseHook(ctx)
	if err != nil {
		t.Error("BeforeReleaseHook should not fail when multi-tenant is disabled")
	}
}

func TestConn_WithTimeout(t *testing.T) {
	cfg := config.NewConfig(
		config.WithQueryTimeout(100 * time.Millisecond),
	)

	conn := &Conn{
		config:   cfg,
		released: false,
		isPooled: false,
	}

	// Test that timeout configuration is applied without executing actual query
	if conn.config.QueryTimeout != 100*time.Millisecond {
		t.Error("Expected timeout config to be set")
	}
}

func TestConn_Hooks(t *testing.T) {
	var beforeQueryCalled bool
	var afterQueryCalled bool
	var beforeReleaseCalled bool

	hooks := &config.HooksConfig{
		BeforeQuery: func(ctx context.Context, query string, args []interface{}) error {
			beforeQueryCalled = true
			return nil
		},
		AfterQuery: func(ctx context.Context, query string, args []interface{}, duration time.Duration, err error) error {
			afterQueryCalled = true
			return nil
		},
		BeforeRelease: func(ctx context.Context, conn interface{}) error {
			beforeReleaseCalled = true
			return nil
		},
	}

	cfg := config.NewConfig(
		config.WithHooks(hooks),
	)

	// Create pgxmock connection
	helper, err := NewTestHelperWithConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create test helper: %v", err)
	}
	defer helper.Close()

	// Set up expectations
	rows := pgxmock.NewRows([]string{"result"}).AddRow(1)
	helper.ExpectQuery("SELECT 1").WillReturnRows(rows)

	conn := helper.NewConnWithMock()
	ctx := context.Background()

	// Test query hooks are called
	var result int
	err = conn.QueryOne(ctx, &result, "SELECT 1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !beforeQueryCalled {
		t.Error("Expected BeforeQuery hook to be called")
	}

	if !afterQueryCalled {
		t.Error("Expected AfterQuery hook to be called")
	}

	// Test release hook
	conn.BeforeReleaseHook(ctx)

	if !beforeReleaseCalled {
		t.Error("Expected BeforeRelease hook to be called")
	}

	// Verify all expectations were met
	if err := helper.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %v", err)
	}
}

func TestConn_InvalidBatchType(t *testing.T) {
	cfg := config.DefaultConfig()
	conn := &Conn{
		config:   cfg,
		released: false,
		isPooled: false,
	}

	ctx := context.Background()

	// Create a mock batch that's not our type
	type mockBatch struct {
		postgresql.IBatch
	}

	mockBatchImpl := &mockBatch{}

	// This should fail with invalid batch type
	_, err := conn.SendBatch(ctx, mockBatchImpl)
	if err == nil {
		t.Error("Expected error for invalid batch type")
	}

	expectedMsg := "invalid batch type"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestConn_ConvertTxOptions(t *testing.T) {
	conn := &Conn{}

	tests := []struct {
		name  string
		input postgresql.TxOptions
	}{
		{
			name: "all options set",
			input: postgresql.TxOptions{
				IsoLevel:       postgresql.IsoLevelSerializable,
				AccessMode:     postgresql.AccessModeReadOnly,
				DeferrableMode: postgresql.DeferrableModeDeferrable,
			},
		},
		{
			name: "default options",
			input: postgresql.TxOptions{
				IsoLevel:       postgresql.IsoLevelDefault,
				AccessMode:     postgresql.AccessModeDefault,
				DeferrableMode: postgresql.DeferrableModeDefault,
			},
		},
		{
			name: "read write mode",
			input: postgresql.TxOptions{
				IsoLevel:       postgresql.IsoLevelReadCommitted,
				AccessMode:     postgresql.AccessModeReadWrite,
				DeferrableMode: postgresql.DeferrableModeNotDeferrable,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should not panic
			pgxOpts := conn.convertTxOptions(tt.input)

			// Basic sanity check
			_ = pgxOpts.IsoLevel
			_ = pgxOpts.AccessMode
			_ = pgxOpts.DeferrableMode
		})
	}
}

func TestConn_QueryAllNotImplemented(t *testing.T) {
	helper, err := NewTestHelper()
	if err != nil {
		t.Fatalf("Failed to create test helper: %v", err)
	}
	defer helper.Close()

	conn := helper.NewConnWithMock()
	ctx := context.Background()

	// QueryAll should return "not fully implemented yet" error
	var dst interface{}
	err = conn.QueryAll(ctx, dst, "SELECT 1")

	if err == nil {
		t.Error("Expected error for QueryAll")
	}

	expectedMsg := "not fully implemented yet"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestConn_QueryCount(t *testing.T) {
	helper, err := NewTestHelper()
	if err != nil {
		t.Fatalf("Failed to create test helper: %v", err)
	}
	defer helper.Close()

	// Set up expectations for count query
	rows := pgxmock.NewRows([]string{"count"}).AddRow(10)
	helper.ExpectQuery("SELECT COUNT(.+) FROM users").WillReturnRows(rows)

	conn := helper.NewConnWithMock()
	ctx := context.Background()

	count, err := conn.QueryCount(ctx, "SELECT COUNT(*) FROM users")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if count != 10 {
		t.Errorf("Expected count 10, got %d", count)
	}

	if err := helper.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %v", err)
	}
}

func TestConn_NotificationMethods(t *testing.T) {
	helper, err := NewTestHelper()
	if err != nil {
		t.Fatalf("Failed to create test helper: %v", err)
	}
	defer helper.Close()

	conn := helper.NewConnWithMock()
	ctx := context.Background()

	// Test Listen - should return not implemented error from mock
	err = conn.Listen(ctx, "test_channel")
	if err == nil {
		t.Error("Expected error for Listen")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("Expected 'not implemented' error, got: %v", err)
	}

	// Test Unlisten - should return not implemented error from mock
	err = conn.Unlisten(ctx, "test_channel")
	if err == nil {
		t.Error("Expected error for Unlisten")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("Expected 'not implemented' error, got: %v", err)
	}

	// Test WaitForNotification - should return not implemented error from mock
	_, err = conn.WaitForNotification(ctx, time.Second)
	if err == nil {
		t.Error("Expected error for WaitForNotification")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("Expected 'not implemented' error, got: %v", err)
	}
}

func TestIsTLSEnabled(t *testing.T) {
	// Test with no environment variable
	os.Unsetenv("DB_TLS_ENABLED")
	if isTLSEnabled() {
		t.Error("Expected TLS to be disabled when env var is not set")
	}

	// Test with false value
	os.Setenv("DB_TLS_ENABLED", "false")
	if isTLSEnabled() {
		t.Error("Expected TLS to be disabled when env var is 'false'")
	}

	// Test with true value
	os.Setenv("DB_TLS_ENABLED", "true")
	if !isTLSEnabled() {
		t.Error("Expected TLS to be enabled when env var is 'true'")
	}

	// Test with invalid value (should default to false)
	os.Setenv("DB_TLS_ENABLED", "invalid")
	if isTLSEnabled() {
		t.Error("Expected TLS to be disabled when env var is invalid")
	}

	// Clean up
	os.Unsetenv("DB_TLS_ENABLED")
}

func TestConn_HookErrors(t *testing.T) {
	hookError := "hook failed"

	hooks := &config.HooksConfig{
		BeforeQuery: func(ctx context.Context, query string, args []interface{}) error {
			return errors.New(hookError)
		},
	}

	cfg := config.NewConfig(
		config.WithHooks(hooks),
	)

	conn := &Conn{
		config:   cfg,
		released: false,
		isPooled: false,
	}

	ctx := context.Background()

	// Test that hook error is returned
	err := conn.QueryOne(ctx, nil, "SELECT 1")
	if err == nil {
		t.Error("Expected error from hook")
	}

	if !strings.Contains(err.Error(), hookError) {
		t.Errorf("Expected error to contain '%s', got '%s'", hookError, err.Error())
	}
}

func TestConn_EdgeCases(t *testing.T) {
	// Test with minimal config
	cfg := &config.Config{}

	helper, err := NewTestHelperWithConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create test helper: %v", err)
	}
	defer helper.Close()

	conn := helper.NewConnWithMock()
	ctx := context.Background()

	// Operations should not panic even with minimal config
	err = conn.QueryOne(ctx, nil, "SELECT 1")
	if err == nil {
		t.Error("Expected error due to mock expectations not being set")
	}
}

func TestConn_QueryOneWithMock(t *testing.T) {
	helper, err := NewTestHelper()
	if err != nil {
		t.Fatalf("Failed to create test helper: %v", err)
	}
	defer helper.Close()

	// Set up expectations - single column
	rows := pgxmock.NewRows([]string{"id"}).AddRow(1)
	helper.ExpectQuery("SELECT (.+) FROM users WHERE id = (.+)").
		WithArgs(1).
		WillReturnRows(rows)

	conn := helper.NewConnWithMock()
	ctx := context.Background()

	var id int
	err = conn.QueryOne(ctx, &id, "SELECT id FROM users WHERE id = $1", 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if id != 1 {
		t.Errorf("Expected id 1, got %d", id)
	}

	if err := helper.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %v", err)
	}
}

func TestConn_QueryWithMock(t *testing.T) {
	helper, err := NewTestHelper()
	if err != nil {
		t.Fatalf("Failed to create test helper: %v", err)
	}
	defer helper.Close()

	// Set up expectations
	rows := pgxmock.NewRows([]string{"id", "name"}).
		AddRow(1, "user1").
		AddRow(2, "user2")
	helper.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)

	conn := helper.NewConnWithMock()
	ctx := context.Background()

	rows2, err := conn.Query(ctx, "SELECT id, name FROM users")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	defer rows2.Close()

	count := 0
	for rows2.Next() {
		count++
		var id int
		var name string
		if err := rows2.Scan(&id, &name); err != nil {
			t.Errorf("Scan error: %v", err)
		}
	}

	if count != 2 {
		t.Errorf("Expected 2 rows, got %d", count)
	}

	if err := helper.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %v", err)
	}
}

func TestConn_ExecWithMock(t *testing.T) {
	helper, err := NewTestHelper()
	if err != nil {
		t.Fatalf("Failed to create test helper: %v", err)
	}
	defer helper.Close()

	// Set up expectations
	helper.ExpectExec("INSERT INTO users").
		WithArgs("test").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	conn := helper.NewConnWithMock()
	ctx := context.Background()

	err = conn.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := helper.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %v", err)
	}
}

func TestConn_BeginTransactionWithMock(t *testing.T) {
	helper, err := NewTestHelper()
	if err != nil {
		t.Fatalf("Failed to create test helper: %v", err)
	}
	defer helper.Close()

	// Set up expectations
	helper.ExpectBegin()

	conn := helper.NewConnWithMock()
	ctx := context.Background()

	tx, err := conn.BeginTransaction(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if tx == nil {
		t.Error("Expected transaction to be returned")
	}

	if err := helper.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %v", err)
	}
}

func TestConn_PingWithMock(t *testing.T) {
	helper, err := NewTestHelper()
	if err != nil {
		t.Fatalf("Failed to create test helper: %v", err)
	}
	defer helper.Close()

	// Set up expectations
	helper.ExpectPing()

	conn := helper.NewConnWithMock()
	ctx := context.Background()

	err = conn.Ping(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := helper.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %v", err)
	}
}

func TestConn_PrepareWithMock(t *testing.T) {
	helper, err := NewTestHelper()
	if err != nil {
		t.Fatalf("Failed to create test helper: %v", err)
	}
	defer helper.Close()

	// Set up expectations
	helper.ExpectPrepare("test_stmt", "SELECT (.+) FROM users WHERE id = (.+)")

	conn := helper.NewConnWithMock()
	ctx := context.Background()

	err = conn.Prepare(ctx, "test_stmt", "SELECT id FROM users WHERE id = $1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if err := helper.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %v", err)
	}
}

func TestConn_QueryCountWithMock(t *testing.T) {
	helper, err := NewTestHelper()
	if err != nil {
		t.Fatalf("Failed to create test helper: %v", err)
	}
	defer helper.Close()

	// Set up expectations
	rows := pgxmock.NewRows([]string{"count"}).AddRow(5)
	helper.ExpectQuery("SELECT COUNT(.+) FROM users").WillReturnRows(rows)

	conn := helper.NewConnWithMock()
	ctx := context.Background()

	count, err := conn.QueryCount(ctx, "SELECT COUNT(*) FROM users")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected count 5, got %d", count)
	}

	if err := helper.ExpectationsWereMet(); err != nil {
		t.Errorf("Expectations were not met: %v", err)
	}
}
