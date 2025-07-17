package mocks

import (
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

// MockBatchResults implements interfaces.IBatchResults for testing
type MockBatchResults struct {
	QueryRowFunc func() interfaces.IRow
	QueryFunc    func() (interfaces.IRows, error)
	ExecFunc     func() (interfaces.CommandTag, error)
	CloseFunc    func() error
}

// NewMockBatchResults creates a new mock batch results
func NewMockBatchResults() *MockBatchResults {
	return &MockBatchResults{}
}

func (m *MockBatchResults) QueryRow() interfaces.IRow {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc()
	}
	return NewMockRow()
}

func (m *MockBatchResults) Query() (interfaces.IRows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc()
	}
	return NewMockRows(), nil
}

func (m *MockBatchResults) Exec() (interfaces.CommandTag, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc()
	}
	return NewMockCommandTag(), nil
}

func (m *MockBatchResults) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockBatchResults) Err() error {
	// For mock purposes, return nil
	return nil
}
