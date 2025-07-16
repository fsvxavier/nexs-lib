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
}

func (m *MockConnection) QueryRow(ctx context.Context, query string, args ...interface{}) interfaces.IRow {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, query, args...)
	}
	return &MockRow{}
}

func (m *MockConnection) Query(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, query, args...)
	}
	return &MockRows{}, nil
}

func (m *MockConnection) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if m.QueryOneFunc != nil {
		return m.QueryOneFunc(ctx, dst, query, args...)
	}
	return nil
}

func (m *MockConnection) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if m.QueryAllFunc != nil {
		return m.QueryAllFunc(ctx, dst, query, args...)
	}
	return nil
}

func (m *MockConnection) QueryCount(ctx context.Context, query string, args ...interface{}) (int64, error) {
	if m.QueryCountFunc != nil {
		return m.QueryCountFunc(ctx, query, args...)
	}
	return 0, nil
}

func (m *MockConnection) Exec(ctx context.Context, query string, args ...interface{}) (interfaces.CommandTag, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(ctx, query, args...)
	}
	return &MockCommandTag{}, nil
}

func (m *MockConnection) SendBatch(ctx context.Context, batch interfaces.IBatch) interfaces.IBatchResults {
	if m.SendBatchFunc != nil {
		return m.SendBatchFunc(ctx, batch)
	}
	return &MockBatchResults{}
}

func (m *MockConnection) Begin(ctx context.Context) (interfaces.ITransaction, error) {
	if m.BeginFunc != nil {
		return m.BeginFunc(ctx)
	}
	return &MockTransaction{}, nil
}

func (m *MockConnection) BeginTx(ctx context.Context, txOptions interfaces.TxOptions) (interfaces.ITransaction, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx, txOptions)
	}
	return &MockTransaction{}, nil
}

func (m *MockConnection) Release() {
	if m.ReleaseFunc != nil {
		m.ReleaseFunc()
	}
}

func (m *MockConnection) Close(ctx context.Context) error {
	if m.CloseFunc != nil {
		return m.CloseFunc(ctx)
	}
	return nil
}

func (m *MockConnection) Ping(ctx context.Context) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	return nil
}

func (m *MockConnection) IsClosed() bool {
	if m.IsClosedFunc != nil {
		return m.IsClosedFunc()
	}
	return false
}

func (m *MockConnection) Prepare(ctx context.Context, name, query string) error {
	if m.PrepareFunc != nil {
		return m.PrepareFunc(ctx, name, query)
	}
	return nil
}

func (m *MockConnection) Deallocate(ctx context.Context, name string) error {
	if m.DeallocateFunc != nil {
		return m.DeallocateFunc(ctx, name)
	}
	return nil
}

func (m *MockConnection) CopyFrom(ctx context.Context, tableName string, columnNames []string, rowSrc interfaces.CopyFromSource) (int64, error) {
	if m.CopyFromFunc != nil {
		return m.CopyFromFunc(ctx, tableName, columnNames, rowSrc)
	}
	return 0, nil
}

func (m *MockConnection) CopyTo(ctx context.Context, w interfaces.CopyToWriter, query string, args ...interface{}) error {
	if m.CopyToFunc != nil {
		return m.CopyToFunc(ctx, w, query, args...)
	}
	return nil
}

func (m *MockConnection) Listen(ctx context.Context, channel string) error {
	if m.ListenFunc != nil {
		return m.ListenFunc(ctx, channel)
	}
	return nil
}

func (m *MockConnection) Unlisten(ctx context.Context, channel string) error {
	if m.UnlistenFunc != nil {
		return m.UnlistenFunc(ctx, channel)
	}
	return nil
}

func (m *MockConnection) WaitForNotification(ctx context.Context, timeout time.Duration) (*interfaces.Notification, error) {
	if m.WaitForNotificationFunc != nil {
		return m.WaitForNotificationFunc(ctx, timeout)
	}
	return nil, nil
}

func (m *MockConnection) SetTenant(ctx context.Context, tenantID string) error {
	if m.SetTenantFunc != nil {
		return m.SetTenantFunc(ctx, tenantID)
	}
	return nil
}

func (m *MockConnection) GetTenant(ctx context.Context) (string, error) {
	if m.GetTenantFunc != nil {
		return m.GetTenantFunc(ctx)
	}
	return "", nil
}

func (m *MockConnection) GetHookManager() interfaces.HookManager {
	if m.GetHookManagerFunc != nil {
		return m.GetHookManagerFunc()
	}
	return nil
}

func (m *MockConnection) HealthCheck(ctx context.Context) error {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}
	return nil
}

func (m *MockConnection) Stats() interfaces.ConnectionStats {
	if m.StatsFunc != nil {
		return m.StatsFunc()
	}
	return interfaces.ConnectionStats{}
}

// MockRow implements interfaces.IRow for testing
type MockRow struct {
	ScanFunc func(dest ...interface{}) error
}

func (m *MockRow) Scan(dest ...interface{}) error {
	if m.ScanFunc != nil {
		return m.ScanFunc(dest...)
	}
	return nil
}

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
	return &MockCommandTag{}
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

// MockTransaction implements interfaces.ITransaction for testing
type MockTransaction struct {
	CommitFunc   func(ctx context.Context) error
	RollbackFunc func(ctx context.Context) error
	// Embed MockConnection for transaction operations
	*MockConnection
}

func (m *MockTransaction) Commit(ctx context.Context) error {
	if m.CommitFunc != nil {
		return m.CommitFunc(ctx)
	}
	return nil
}

func (m *MockTransaction) Rollback(ctx context.Context) error {
	if m.RollbackFunc != nil {
		return m.RollbackFunc(ctx)
	}
	return nil
}

// MockBatchResults implements interfaces.IBatchResults for testing
type MockBatchResults struct {
	QueryRowFunc func() interfaces.IRow
	QueryFunc    func() (interfaces.IRows, error)
	ExecFunc     func() (interfaces.CommandTag, error)
	CloseFunc    func() error
}

func (m *MockBatchResults) QueryRow() interfaces.IRow {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc()
	}
	return &MockRow{}
}

func (m *MockBatchResults) Query() (interfaces.IRows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc()
	}
	return &MockRows{}, nil
}

func (m *MockBatchResults) Exec() (interfaces.CommandTag, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc()
	}
	return &MockCommandTag{}, nil
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

// MockCommandTag implements interfaces.CommandTag for testing
type MockCommandTag struct {
	StringFunc       func() string
	RowsAffectedFunc func() int64
	InsertFunc       func() bool
	UpdateFunc       func() bool
	DeleteFunc       func() bool
	SelectFunc       func() bool
}

func (m *MockCommandTag) String() string {
	if m.StringFunc != nil {
		return m.StringFunc()
	}
	return ""
}

func (m *MockCommandTag) RowsAffected() int64 {
	if m.RowsAffectedFunc != nil {
		return m.RowsAffectedFunc()
	}
	return 0
}

func (m *MockCommandTag) Insert() bool {
	if m.InsertFunc != nil {
		return m.InsertFunc()
	}
	return false
}

func (m *MockCommandTag) Update() bool {
	if m.UpdateFunc != nil {
		return m.UpdateFunc()
	}
	return false
}

func (m *MockCommandTag) Delete() bool {
	if m.DeleteFunc != nil {
		return m.DeleteFunc()
	}
	return false
}

func (m *MockCommandTag) Select() bool {
	if m.SelectFunc != nil {
		return m.SelectFunc()
	}
	return false
}

// NewMockConnection creates a new mock connection with default implementations
func NewMockConnection() *MockConnection {
	return &MockConnection{}
}

// NewMockRow creates a new mock row
func NewMockRow() *MockRow {
	return &MockRow{}
}

// NewMockRows creates a new mock rows
func NewMockRows() *MockRows {
	return &MockRows{}
}

// NewMockTransaction creates a new mock transaction
func NewMockTransaction() *MockTransaction {
	return &MockTransaction{
		MockConnection: NewMockConnection(),
	}
}

// NewMockBatchResults creates a new mock batch results
func NewMockBatchResults() *MockBatchResults {
	return &MockBatchResults{}
}

// NewMockCommandTag creates a new mock command tag
func NewMockCommandTag() *MockCommandTag {
	return &MockCommandTag{}
}
