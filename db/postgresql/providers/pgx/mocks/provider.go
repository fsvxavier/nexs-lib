package mocks

// This file provides convenient constructors and utilities for all mocks

// NewMockProvider creates a new mock provider for testing
func NewMockProvider() *MockProvider {
	return &MockProvider{
		GetConnectionFunc: func() (*MockConnection, error) {
			return NewMockConnection(), nil
		},
		GetPoolFunc: func() (*MockPool, error) {
			return NewMockPool(), nil
		},
		CloseFunc: func() error {
			return nil
		},
	}
}

// MockProvider provides a simple mock for provider-level testing
type MockProvider struct {
	GetConnectionFunc func() (*MockConnection, error)
	GetPoolFunc       func() (*MockPool, error)
	CloseFunc         func() error
}

func (m *MockProvider) GetConnection() (*MockConnection, error) {
	if m.GetConnectionFunc != nil {
		return m.GetConnectionFunc()
	}
	return NewMockConnection(), nil
}

func (m *MockProvider) GetPool() (*MockPool, error) {
	if m.GetPoolFunc != nil {
		return m.GetPoolFunc()
	}
	return NewMockPool(), nil
}

func (m *MockProvider) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// MockPool provides a simple mock for pool testing
type MockPool struct {
	AcquireFunc func() (*MockConnection, error)
	ReleaseFunc func(*MockConnection)
	CloseFunc   func() error
	StatsFunc   func() interface{}
}

// NewMockPool creates a new mock pool
func NewMockPool() *MockPool {
	return &MockPool{}
}

func (m *MockPool) Acquire() (*MockConnection, error) {
	if m.AcquireFunc != nil {
		return m.AcquireFunc()
	}
	return NewMockConnection(), nil
}

func (m *MockPool) Release(conn *MockConnection) {
	if m.ReleaseFunc != nil {
		m.ReleaseFunc(conn)
	}
}

func (m *MockPool) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockPool) Stats() interface{} {
	if m.StatsFunc != nil {
		return m.StatsFunc()
	}
	return nil
}
