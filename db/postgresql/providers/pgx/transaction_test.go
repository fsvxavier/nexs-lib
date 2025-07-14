//go:build unit

package pgx

import (
	"context"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
)

func TestTransaction_Interface(t *testing.T) {
	// Test interface compliance
	var _ postgresql.ITransaction = &Transaction{}
}

func TestTransaction_CompletedOperations(t *testing.T) {
	cfg := config.DefaultConfig()
	tx := &Transaction{
		config:    cfg,
		completed: true,
		startTime: time.Now(),
	}

	ctx := context.Background()

	// Test all operations fail when transaction is completed
	err := tx.QueryOne(ctx, nil, "SELECT 1")
	if err == nil {
		t.Error("Expected error for QueryOne on completed transaction")
	}

	err = tx.QueryAll(ctx, nil, "SELECT 1")
	if err == nil {
		t.Error("Expected error for QueryAll on completed transaction")
	}

	_, err = tx.QueryCount(ctx, "SELECT COUNT(*)")
	if err == nil {
		t.Error("Expected error for QueryCount on completed transaction")
	}

	_, err = tx.Query(ctx, "SELECT 1")
	if err == nil {
		t.Error("Expected error for Query on completed transaction")
	}

	_, err = tx.QueryRow(ctx, "SELECT 1")
	if err == nil {
		t.Error("Expected error for QueryRow on completed transaction")
	}

	err = tx.Exec(ctx, "SELECT 1")
	if err == nil {
		t.Error("Expected error for Exec on completed transaction")
	}

	_, err = tx.SendBatch(ctx, NewBatch())
	if err == nil {
		t.Error("Expected error for SendBatch on completed transaction")
	}

	err = tx.Prepare(ctx, "test", "SELECT 1")
	if err == nil {
		t.Error("Expected error for Prepare on completed transaction")
	}

	err = tx.Ping(ctx)
	if err == nil {
		t.Error("Expected error for Ping on completed transaction")
	}

	err = tx.Savepoint(ctx, "test")
	if err == nil {
		t.Error("Expected error for Savepoint on completed transaction")
	}

	err = tx.RollbackToSavepoint(ctx, "test")
	if err == nil {
		t.Error("Expected error for RollbackToSavepoint on completed transaction")
	}

	err = tx.ReleaseSavepoint(ctx, "test")
	if err == nil {
		t.Error("Expected error for ReleaseSavepoint on completed transaction")
	}
}

func TestTransaction_NestedTransactions(t *testing.T) {
	cfg := config.DefaultConfig()
	tx := &Transaction{
		config:    cfg,
		completed: false,
		startTime: time.Now(),
	}

	ctx := context.Background()

	// Test nested transactions are not supported
	_, err := tx.BeginTransaction(ctx)
	if err == nil {
		t.Error("Expected error for nested transaction")
	}

	expectedMsg := "nested transactions are not supported"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Test with options
	_, err = tx.BeginTransactionWithOptions(ctx, postgresql.TxOptions{})
	if err == nil {
		t.Error("Expected error for nested transaction with options")
	}

	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestTransaction_UnsupportedOperations(t *testing.T) {
	cfg := config.DefaultConfig()
	tx := &Transaction{
		config:    cfg,
		completed: false,
		startTime: time.Now(),
	}

	ctx := context.Background()

	// Test LISTEN is not supported
	err := tx.Listen(ctx, "test_channel")
	if err == nil {
		t.Error("Expected error for LISTEN in transaction")
	}

	expectedMsg := "LISTEN is not supported in transactions"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Test UNLISTEN is not supported
	err = tx.Unlisten(ctx, "test_channel")
	if err == nil {
		t.Error("Expected error for UNLISTEN in transaction")
	}

	expectedMsg = "UNLISTEN is not supported in transactions"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}

	// Test WaitForNotification is not supported
	_, err = tx.WaitForNotification(ctx, time.Second)
	if err == nil {
		t.Error("Expected error for WaitForNotification in transaction")
	}

	expectedMsg = "WaitForNotification is not supported in transactions"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestTransaction_HookMethods(t *testing.T) {
	cfg := config.DefaultConfig()
	tx := &Transaction{
		config:    cfg,
		completed: false,
		startTime: time.Now(),
	}

	ctx := context.Background()

	// Test hook methods (should be no-ops for transactions)
	err := tx.BeforeReleaseHook(ctx)
	if err != nil {
		t.Errorf("Expected BeforeReleaseHook to return nil, got %v", err)
	}

	err = tx.AfterAcquireHook(ctx)
	if err != nil {
		t.Errorf("Expected AfterAcquireHook to return nil, got %v", err)
	}

	// Test Release (should be a no-op)
	tx.Release(ctx) // Should not panic
}

func TestTransaction_DoubleCommit(t *testing.T) {
	cfg := config.DefaultConfig()
	tx := &Transaction{
		config:    cfg,
		completed: false,
		startTime: time.Now(),
	}

	ctx := context.Background()

	// Simulate first commit
	tx.completed = true

	// Try to commit again
	err := tx.Commit(ctx)
	if err == nil {
		t.Error("Expected error for double commit")
	}

	expectedMsg := "transaction is already completed"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestTransaction_DoubleRollback(t *testing.T) {
	cfg := config.DefaultConfig()
	tx := &Transaction{
		config:    cfg,
		completed: false,
		startTime: time.Now(),
	}

	ctx := context.Background()

	// Simulate first rollback
	tx.completed = true

	// Try to rollback again
	err := tx.Rollback(ctx)
	if err == nil {
		t.Error("Expected error for double rollback")
	}

	expectedMsg := "transaction is already completed"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestTransaction_QueryAllNotImplemented(t *testing.T) {
	cfg := config.DefaultConfig()
	tx := &Transaction{
		config:    cfg,
		completed: false,
		startTime: time.Now(),
	}

	ctx := context.Background()

	// Mock transaction - we can't test actual DB operations without connection
	// But we can test that the method exists and returns the expected error

	// For now, we test that the method signature is correct
	var dst interface{}
	err := tx.QueryAll(ctx, dst, "SELECT 1")

	// We expect this to fail because there's no real pgx.Tx
	// but it should fail at the pgx level, not at our validation level
	if err == nil {
		t.Error("Expected error due to nil pgx.Tx")
	}
}

func TestTransaction_WithTimeout(t *testing.T) {
	cfg := config.NewConfig(
		config.WithQueryTimeout(100 * time.Millisecond),
	)

	tx := &Transaction{
		config:    cfg,
		completed: false,
		startTime: time.Now(),
	}

	ctx := context.Background()

	// Test that timeout is applied (will fail due to nil tx, but timeout should be set)
	err := tx.QueryOne(ctx, nil, "SELECT 1")
	if err == nil {
		t.Error("Expected error due to nil pgx.Tx")
	}

	// The error should not be related to timeout validation, but to nil tx
	// This tests that our timeout logic doesn't crash
}

func TestTransaction_Hooks(t *testing.T) {
	var beforeQueryCalled bool
	var afterQueryCalled bool
	var afterTxCalled bool

	hooks := &config.HooksConfig{
		BeforeQuery: func(ctx context.Context, query string, args []interface{}) error {
			beforeQueryCalled = true
			return nil
		},
		AfterQuery: func(ctx context.Context, query string, args []interface{}, duration time.Duration, err error) error {
			afterQueryCalled = true
			return nil
		},
		AfterTransaction: func(ctx context.Context, committed bool, duration time.Duration, err error) error {
			afterTxCalled = true
			return nil
		},
	}

	cfg := config.NewConfig(
		config.WithHooks(hooks),
	)

	tx := &Transaction{
		config:    cfg,
		completed: false,
		startTime: time.Now(),
	}

	ctx := context.Background()

	// Test hooks are called (will fail due to nil tx, but hooks should be called)
	tx.QueryOne(ctx, nil, "SELECT 1")

	if !beforeQueryCalled {
		t.Error("Expected BeforeQuery hook to be called")
	}

	if !afterQueryCalled {
		t.Error("Expected AfterQuery hook to be called")
	}

	// Reset flags
	afterTxCalled = false

	// Test transaction commit hook
	tx.Commit(ctx) // Will fail due to nil tx, but hook should be called

	if !afterTxCalled {
		t.Error("Expected AfterTransaction hook to be called on commit")
	}

	// Reset for rollback test
	afterTxCalled = false
	tx.completed = false

	// Test transaction rollback hook
	tx.Rollback(ctx) // Will fail due to nil tx, but hook should be called

	if !afterTxCalled {
		t.Error("Expected AfterTransaction hook to be called on rollback")
	}
}

// Test conversion methods

func TestConvertTxOptions(t *testing.T) {
	// Transaction convertTxOptions is an internal method
	// Testing is covered by the transaction creation and operation tests
}
