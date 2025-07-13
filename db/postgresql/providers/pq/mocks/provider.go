package mocks

import (
	"context"

	"github.com/fsvxavier/nexs-lib/db/postgresql/interfaces"
)

// MockProvider is a mock implementation of interfaces.DatabaseProvider for lib/pq
type MockProvider struct {
	ConnectFunc func() error
	CloseFunc   func() error
	DBFunc      func() any
	PoolFunc    func() interfaces.IPool
}

// Connect calls ConnectFunc if set, otherwise returns nil
func (m *MockProvider) Connect() error {
	if m.ConnectFunc != nil {
		return m.ConnectFunc()
	}
	return nil
}

// Close calls CloseFunc if set, otherwise returns nil
func (m *MockProvider) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// DB calls DBFunc if set, otherwise returns nil
func (m *MockProvider) DB() any {
	if m.DBFunc != nil {
		return m.DBFunc()
	}
	return nil
}

// Pool calls PoolFunc if set, otherwise returns nil
func (m *MockProvider) Pool() interfaces.IPool {
	if m.PoolFunc != nil {
		return m.PoolFunc()
	}
	return nil
}

// MockPool is a mock implementation of interfaces.IPool for lib/pq
type MockPool struct {
	AcquireFunc               func(ctx context.Context) (interfaces.IConn, error)
	CloseFunc                 func()
	PingFunc                  func(ctx context.Context) error
	GetConnWithNotPresentFunc func(ctx context.Context, conn interfaces.IConn) (interfaces.IConn, func(), error)
	StatsFunc                 func() interfaces.PoolStats
}

// Acquire calls AcquireFunc if set, otherwise returns nil
func (m *MockPool) Acquire(ctx context.Context) (interfaces.IConn, error) {
	if m.AcquireFunc != nil {
		return m.AcquireFunc(ctx)
	}
	return nil, nil
}

// Close calls CloseFunc if set
func (m *MockPool) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}

// Ping calls PingFunc if set, otherwise returns nil
func (m *MockPool) Ping(ctx context.Context) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	return nil
}

// GetConnWithNotPresent calls GetConnWithNotPresentFunc if set, otherwise returns nil
func (m *MockPool) GetConnWithNotPresent(ctx context.Context, conn interfaces.IConn) (interfaces.IConn, func(), error) {
	if m.GetConnWithNotPresentFunc != nil {
		return m.GetConnWithNotPresentFunc(ctx, conn)
	}
	return nil, func() {}, nil
}

// Stats calls StatsFunc if set, otherwise returns empty stats
func (m *MockPool) Stats() interfaces.PoolStats {
	if m.StatsFunc != nil {
		return m.StatsFunc()
	}
	return interfaces.PoolStats{}
}

// MockConn is a mock implementation of interfaces.IConn for lib/pq
type MockConn struct {
	QueryOneFunc                    func(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryAllFunc                    func(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryCountFunc                  func(ctx context.Context, query string, args ...interface{}) (*int, error)
	QueryFunc                       func(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error)
	ExecFunc                        func(ctx context.Context, query string, args ...interface{}) error
	SendBatchFunc                   func(ctx context.Context, batch interfaces.IBatch) (interfaces.IBatchResults, error)
	QueryRowFunc                    func(ctx context.Context, query string, args ...interface{}) (interfaces.IRow, error)
	QueryRowsFunc                   func(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error)
	BeforeReleaseHookFunc           func(ctx context.Context) error
	AfterAcquireHookFunc            func(ctx context.Context) error
	ReleaseFunc                     func(ctx context.Context)
	PingFunc                        func(ctx context.Context) error
	BeginTransactionFunc            func(ctx context.Context) (interfaces.ITransaction, error)
	BeginTransactionWithOptionsFunc func(ctx context.Context, options interfaces.TxOptions) (interfaces.ITransaction, error)
}

// QueryOne calls QueryOneFunc if set, otherwise returns nil
func (m *MockConn) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if m.QueryOneFunc != nil {
		return m.QueryOneFunc(ctx, dst, query, args...)
	}
	return nil
}

// QueryAll calls QueryAllFunc if set, otherwise returns nil
func (m *MockConn) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	if m.QueryAllFunc != nil {
		return m.QueryAllFunc(ctx, dst, query, args...)
	}
	return nil
}

// QueryCount calls QueryCountFunc if set, otherwise returns nil
func (m *MockConn) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	if m.QueryCountFunc != nil {
		return m.QueryCountFunc(ctx, query, args...)
	}
	count := 0
	return &count, nil
}

// Query calls QueryFunc if set, otherwise returns nil
func (m *MockConn) Query(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, query, args...)
	}
	return nil, nil
}

// Exec calls ExecFunc if set, otherwise returns nil
func (m *MockConn) Exec(ctx context.Context, query string, args ...interface{}) error {
	if m.ExecFunc != nil {
		return m.ExecFunc(ctx, query, args...)
	}
	return nil
}

// SendBatch calls SendBatchFunc if set, otherwise returns nil
func (m *MockConn) SendBatch(ctx context.Context, batch interfaces.IBatch) (interfaces.IBatchResults, error) {
	if m.SendBatchFunc != nil {
		return m.SendBatchFunc(ctx, batch)
	}
	return nil, nil
}

// QueryRow calls QueryRowFunc if set, otherwise returns nil
func (m *MockConn) QueryRow(ctx context.Context, query string, args ...interface{}) (interfaces.IRow, error) {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, query, args...)
	}
	return nil, nil
}

// QueryRows calls QueryRowsFunc if set, otherwise returns nil
func (m *MockConn) QueryRows(ctx context.Context, query string, args ...interface{}) (interfaces.IRows, error) {
	if m.QueryRowsFunc != nil {
		return m.QueryRowsFunc(ctx, query, args...)
	}
	return nil, nil
}

// BeforeReleaseHook calls BeforeReleaseHookFunc if set, otherwise returns nil
func (m *MockConn) BeforeReleaseHook(ctx context.Context) error {
	if m.BeforeReleaseHookFunc != nil {
		return m.BeforeReleaseHookFunc(ctx)
	}
	return nil
}

// AfterAcquireHook calls AfterAcquireHookFunc if set, otherwise returns nil
func (m *MockConn) AfterAcquireHook(ctx context.Context) error {
	if m.AfterAcquireHookFunc != nil {
		return m.AfterAcquireHookFunc(ctx)
	}
	return nil
}

// Release calls ReleaseFunc if set
func (m *MockConn) Release(ctx context.Context) {
	if m.ReleaseFunc != nil {
		m.ReleaseFunc(ctx)
	}
}

// Ping calls PingFunc if set, otherwise returns nil
func (m *MockConn) Ping(ctx context.Context) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	return nil
}

// BeginTransaction calls BeginTransactionFunc if set, otherwise returns nil
func (m *MockConn) BeginTransaction(ctx context.Context) (interfaces.ITransaction, error) {
	if m.BeginTransactionFunc != nil {
		return m.BeginTransactionFunc(ctx)
	}
	return nil, nil
}

// BeginTransactionWithOptions calls BeginTransactionWithOptionsFunc if set, otherwise returns nil
func (m *MockConn) BeginTransactionWithOptions(ctx context.Context, options interfaces.TxOptions) (interfaces.ITransaction, error) {
	if m.BeginTransactionWithOptionsFunc != nil {
		return m.BeginTransactionWithOptionsFunc(ctx, options)
	}
	return nil, nil
}

// MockTransaction is a mock implementation of interfaces.ITransaction for lib/pq
type MockTransaction struct {
	MockConn
	CommitFunc   func(ctx context.Context) error
	RollbackFunc func(ctx context.Context) error
}

// Commit calls CommitFunc if set, otherwise returns nil
func (m *MockTransaction) Commit(ctx context.Context) error {
	if m.CommitFunc != nil {
		return m.CommitFunc(ctx)
	}
	return nil
}

// Rollback calls RollbackFunc if set, otherwise returns nil
func (m *MockTransaction) Rollback(ctx context.Context) error {
	if m.RollbackFunc != nil {
		return m.RollbackFunc(ctx)
	}
	return nil
}

// MockBatch is a mock implementation of interfaces.IBatch for lib/pq
type MockBatch struct {
	QueueFunc func(query string, arguments ...any)
	LenFunc   func() int
	queries   []string
}

// Queue calls QueueFunc if set, otherwise adds to internal slice
func (m *MockBatch) Queue(query string, arguments ...any) {
	if m.QueueFunc != nil {
		m.QueueFunc(query, arguments...)
		return
	}
	m.queries = append(m.queries, query)
}

// Len calls LenFunc if set, otherwise returns length of internal slice
func (m *MockBatch) Len() int {
	if m.LenFunc != nil {
		return m.LenFunc()
	}
	return len(m.queries)
}

// MockBatchResults is a mock implementation of interfaces.IBatchResults for lib/pq
type MockBatchResults struct {
	QueryOneFunc func(dst interface{}) error
	QueryAllFunc func(dst interface{}) error
	ExecFunc     func() error
	CloseFunc    func()
}

// QueryOne calls QueryOneFunc if set, otherwise returns nil
func (m *MockBatchResults) QueryOne(dst interface{}) error {
	if m.QueryOneFunc != nil {
		return m.QueryOneFunc(dst)
	}
	return nil
}

// QueryAll calls QueryAllFunc if set, otherwise returns nil
func (m *MockBatchResults) QueryAll(dst interface{}) error {
	if m.QueryAllFunc != nil {
		return m.QueryAllFunc(dst)
	}
	return nil
}

// Exec calls ExecFunc if set, otherwise returns nil
func (m *MockBatchResults) Exec() error {
	if m.ExecFunc != nil {
		return m.ExecFunc()
	}
	return nil
}

// Close calls CloseFunc if set
func (m *MockBatchResults) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}

// MockRow is a mock implementation of interfaces.IRow for lib/pq
type MockRow struct {
	ScanFunc func(dest ...any) error
}

// Scan calls ScanFunc if set, otherwise returns nil
func (m *MockRow) Scan(dest ...any) error {
	if m.ScanFunc != nil {
		return m.ScanFunc(dest...)
	}
	return nil
}

// MockRows is a mock implementation of interfaces.IRows for lib/pq
type MockRows struct {
	ScanFunc      func(dest ...any) error
	CloseFunc     func()
	NextFunc      func() bool
	RawValuesFunc func() [][]byte
	ErrFunc       func() error
	hasNext       bool
}

// Scan calls ScanFunc if set, otherwise returns nil
func (m *MockRows) Scan(dest ...any) error {
	if m.ScanFunc != nil {
		return m.ScanFunc(dest...)
	}
	return nil
}

// Close calls CloseFunc if set
func (m *MockRows) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}

// Next calls NextFunc if set, otherwise returns hasNext
func (m *MockRows) Next() bool {
	if m.NextFunc != nil {
		return m.NextFunc()
	}
	return m.hasNext
}

// SetHasNext sets the hasNext value for testing
func (m *MockRows) SetHasNext(hasNext bool) {
	m.hasNext = hasNext
}

// RawValues calls RawValuesFunc if set, otherwise returns nil
func (m *MockRows) RawValues() [][]byte {
	if m.RawValuesFunc != nil {
		return m.RawValuesFunc()
	}
	return nil
}

// Err calls ErrFunc if set, otherwise returns nil
func (m *MockRows) Err() error {
	if m.ErrFunc != nil {
		return m.ErrFunc()
	}
	return nil
}
