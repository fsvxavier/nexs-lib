package mocks

import (
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

// MockRows implements interfaces.IRows for testing
type MockRows struct {
	NextFunc              func() bool
	ScanFunc              func(dest ...interface{}) error
	CloseFunc             func() error
	ErrFunc               func() error
	CommandTagFunc        func() interfaces.CommandTag
	FieldDescriptionsFunc func() []interfaces.FieldDescription
	RawValuesFunc         func() [][]byte
}

// NewMockRows creates a new mock rows
func NewMockRows() *MockRows {
	return &MockRows{}
}

func (m *MockRows) Next() bool {
	if m.NextFunc != nil {
		return m.NextFunc()
	}
	return false
}

func (m *MockRows) Scan(dest ...interface{}) error {
	if m.ScanFunc != nil {
		return m.ScanFunc(dest...)
	}
	return nil
}

func (m *MockRows) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockRows) Err() error {
	if m.ErrFunc != nil {
		return m.ErrFunc()
	}
	return nil
}

func (m *MockRows) CommandTag() interfaces.CommandTag {
	if m.CommandTagFunc != nil {
		return m.CommandTagFunc()
	}
	return NewMockCommandTag()
}

func (m *MockRows) FieldDescriptions() []interfaces.FieldDescription {
	if m.FieldDescriptionsFunc != nil {
		return m.FieldDescriptionsFunc()
	}
	return nil
}

func (m *MockRows) RawValues() [][]byte {
	if m.RawValuesFunc != nil {
		return m.RawValuesFunc()
	}
	return nil
}
