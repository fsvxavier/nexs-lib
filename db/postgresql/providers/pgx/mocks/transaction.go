package mocks

import (
	"context"
)

// MockTransaction implements interfaces.ITransaction for testing
type MockTransaction struct {
	// Embed MockConnection for transaction operations
	*MockConnection

	// Transaction-specific function fields
	CommitFunc   func(ctx context.Context) error
	RollbackFunc func(ctx context.Context) error

	// State tracking
	isCommitted  bool
	isRolledBack bool
}

// NewMockTransaction creates a new mock transaction
func NewMockTransaction() *MockTransaction {
	return &MockTransaction{
		MockConnection: NewMockConnection(),
		isCommitted:    false,
		isRolledBack:   false,
	}
}

func (m *MockTransaction) Commit(ctx context.Context) error {
	m.callCount["Commit"]++
	m.isCommitted = true
	if m.CommitFunc != nil {
		return m.CommitFunc(ctx)
	}
	return nil
}

func (m *MockTransaction) Rollback(ctx context.Context) error {
	m.callCount["Rollback"]++
	m.isRolledBack = true
	if m.RollbackFunc != nil {
		return m.RollbackFunc(ctx)
	}
	return nil
}

// IsCommitted returns whether the transaction was committed
func (m *MockTransaction) IsCommitted() bool {
	return m.isCommitted
}

// IsRolledBack returns whether the transaction was rolled back
func (m *MockTransaction) IsRolledBack() bool {
	return m.isRolledBack
}

// Reset resets the transaction state for reuse in tests
func (m *MockTransaction) Reset() {
	m.isCommitted = false
	m.isRolledBack = false
	m.ResetCallCounts()
}
