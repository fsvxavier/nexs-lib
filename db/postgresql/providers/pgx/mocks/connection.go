package mocks

import (
	"context"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

// MockConnection implements interfaces.IConn for testing
type MockConnection struct {
	QueryRowFunc            func(ctx context.Context, query string, args ...interface{}) interfaces.IRow
	QueryFunc               func(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error)
	QueryOneFunc            func(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryAllFunc            func(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryCountFunc          func(ctx context.Context, query string, args ...interface{}) (int64, error)
	ExecFunc                func(ctx context.Context, query string, args ...interface{}) (interfaces.CommandTag, error)
	SendBatchFunc           func(ctx context.Context, batch interfaces.IBatch) interfaces.IBatchResults
	BeginFunc               func(ctx context.Context) (interfaces.ITransaction, error)
	BeginTxFunc             func(ctx context.Context, txOptions interfaces.TxOptions) (interfaces.ITransaction, error)
	ReleaseFunc             func()
	CloseFunc               func(ctx context.Context) error
	PingFunc                func(ctx context.Context) error
	IsClosedFunc            func() bool
	PrepareFunc             func(ctx context.Context, name, query string) error
	DeallocateFunc          func(ctx context.Context, name string) error
	CopyFromFunc            func(ctx context.Context, tableName string, columnNames []string, rowSrc interfaces.CopyFromSource) (int64, error)
	CopyToFunc              func(ctx context.Context, w interfaces.CopyToWriter, query string, args ...interface{}) error
	ListenFunc              func(ctx context.Context, channel string) error
	UnlistenFunc            func(ctx context.Context, channel string) error
	WaitForNotificationFunc func(ctx context.Context, timeout time.Duration) (*interfaces.Notification, error)
	SetTenantFunc           func(ctx context.Context, tenantID string) error
	GetTenantFunc           func(ctx context.Context) (string, error)
	GetHookManagerFunc      func() interfaces.HookManager
	HealthCheckFunc         func(ctx context.Context) error
	StatsFunc               func() interfaces.ConnectionStats

	// State tracking
	callCount     map[string]int
	isClosedState bool
}

// NewMockConnection creates a new mock connection with default implementations
func NewMockConnection() *MockConnection {
	return &MockConnection{
		callCount:     make(map[string]int),
		isClosedState: false,
	}
}

func (m *MockConnection) QueryRow(ctx context.Context, query string, args ...interface{}) interfaces.IRow {
	m.callCount["QueryRow"]++
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, query, args...)
	}
	return NewMockRow()
}

func (m *MockConnection) Query(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	m.callCount["Query"]++
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, query, args...)
	}
	return NewMockRows(), nil
}

func (m *MockConnection) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	m.callCount["QueryOne"]++
	if m.QueryOneFunc != nil {
		return m.QueryOneFunc(ctx, dst, query, args...)
	}
	return nil
}

func (m *MockConnection) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	m.callCount["QueryAll"]++
	if m.QueryAllFunc != nil {
		return m.QueryAllFunc(ctx, dst, query, args...)
	}
	return nil
}

func (m *MockConnection) QueryCount(ctx context.Context, query string, args ...interface{}) (int64, error) {
	m.callCount["QueryCount"]++
	if m.QueryCountFunc != nil {
		return m.QueryCountFunc(ctx, query, args...)
	}
	return 0, nil
}

func (m *MockConnection) Exec(ctx context.Context, query string, args ...interface{}) (interfaces.CommandTag, error) {
	m.callCount["Exec"]++
	if m.ExecFunc != nil {
		return m.ExecFunc(ctx, query, args...)
	}
	return NewMockCommandTag(), nil
}

func (m *MockConnection) SendBatch(ctx context.Context, batch interfaces.IBatch) interfaces.IBatchResults {
	m.callCount["SendBatch"]++
	if m.SendBatchFunc != nil {
		return m.SendBatchFunc(ctx, batch)
	}
	return NewMockBatchResults()
}

func (m *MockConnection) Begin(ctx context.Context) (interfaces.ITransaction, error) {
	m.callCount["Begin"]++
	if m.BeginFunc != nil {
		return m.BeginFunc(ctx)
	}
	return NewMockTransaction(), nil
}

func (m *MockConnection) BeginTx(ctx context.Context, txOptions interfaces.TxOptions) (interfaces.ITransaction, error) {
	m.callCount["BeginTx"]++
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx, txOptions)
	}
	return NewMockTransaction(), nil
}

func (m *MockConnection) Release() {
	m.callCount["Release"]++
	if m.ReleaseFunc != nil {
		m.ReleaseFunc()
	}
}

func (m *MockConnection) Close(ctx context.Context) error {
	m.callCount["Close"]++
	m.isClosedState = true
	if m.CloseFunc != nil {
		return m.CloseFunc(ctx)
	}
	return nil
}

func (m *MockConnection) Ping(ctx context.Context) error {
	m.callCount["Ping"]++
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	return nil
}

func (m *MockConnection) IsClosed() bool {
	if m.IsClosedFunc != nil {
		return m.IsClosedFunc()
	}
	return m.isClosedState
}

func (m *MockConnection) Prepare(ctx context.Context, name, query string) error {
	m.callCount["Prepare"]++
	if m.PrepareFunc != nil {
		return m.PrepareFunc(ctx, name, query)
	}
	return nil
}

func (m *MockConnection) Deallocate(ctx context.Context, name string) error {
	m.callCount["Deallocate"]++
	if m.DeallocateFunc != nil {
		return m.DeallocateFunc(ctx, name)
	}
	return nil
}

func (m *MockConnection) CopyFrom(ctx context.Context, tableName string, columnNames []string, rowSrc interfaces.CopyFromSource) (int64, error) {
	m.callCount["CopyFrom"]++
	if m.CopyFromFunc != nil {
		return m.CopyFromFunc(ctx, tableName, columnNames, rowSrc)
	}
	return 0, nil
}

func (m *MockConnection) CopyTo(ctx context.Context, w interfaces.CopyToWriter, query string, args ...interface{}) error {
	m.callCount["CopyTo"]++
	if m.CopyToFunc != nil {
		return m.CopyToFunc(ctx, w, query, args...)
	}
	return nil
}

func (m *MockConnection) Listen(ctx context.Context, channel string) error {
	m.callCount["Listen"]++
	if m.ListenFunc != nil {
		return m.ListenFunc(ctx, channel)
	}
	return nil
}

func (m *MockConnection) Unlisten(ctx context.Context, channel string) error {
	m.callCount["Unlisten"]++
	if m.UnlistenFunc != nil {
		return m.UnlistenFunc(ctx, channel)
	}
	return nil
}

func (m *MockConnection) WaitForNotification(ctx context.Context, timeout time.Duration) (*interfaces.Notification, error) {
	m.callCount["WaitForNotification"]++
	if m.WaitForNotificationFunc != nil {
		return m.WaitForNotificationFunc(ctx, timeout)
	}
	return nil, nil
}

func (m *MockConnection) SetTenant(ctx context.Context, tenantID string) error {
	m.callCount["SetTenant"]++
	if m.SetTenantFunc != nil {
		return m.SetTenantFunc(ctx, tenantID)
	}
	return nil
}

func (m *MockConnection) GetTenant(ctx context.Context) (string, error) {
	m.callCount["GetTenant"]++
	if m.GetTenantFunc != nil {
		return m.GetTenantFunc(ctx)
	}
	return "default", nil
}

func (m *MockConnection) GetHookManager() interfaces.HookManager {
	m.callCount["GetHookManager"]++
	if m.GetHookManagerFunc != nil {
		return m.GetHookManagerFunc()
	}
	return NewMockHookManager()
}

func (m *MockConnection) HealthCheck(ctx context.Context) error {
	m.callCount["HealthCheck"]++
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}
	return nil
}

func (m *MockConnection) Stats() interfaces.ConnectionStats {
	m.callCount["Stats"]++
	if m.StatsFunc != nil {
		return m.StatsFunc()
	}
	return interfaces.ConnectionStats{
		TotalQueries:      int64(m.callCount["Query"] + m.callCount["QueryRow"] + m.callCount["QueryAll"]),
		TotalExecs:        int64(m.callCount["Exec"]),
		TotalTransactions: int64(m.callCount["Begin"] + m.callCount["BeginTx"]),
		CreatedAt:         time.Now(),
	}
}

// GetCallCount returns the number of times a method was called
func (m *MockConnection) GetCallCount(method string) int {
	return m.callCount[method]
}

// ResetCallCounts resets all call counters
func (m *MockConnection) ResetCallCounts() {
	m.callCount = make(map[string]int)
}
