package mocks

import (
	"context"
	"database/sql"
	"errors"

	"github.com/DATA-DOG/go-sqlmock"
)

// GetMock retorna um novo mock para testes de sql
func GetMock() (*sql.DB, sqlmock.Sqlmock, error) {
	return sqlmock.New()
}

// CloseDB fecha a conexão mock
func CloseDB(db *sql.DB) {
	db.Close()
}

// MockDB encapsula um DB mockado para testes
type MockDB struct {
	DB *sql.DB
}

// MockConn representa uma conexão mockada para testes
type MockConn struct {
	// Mock para os métodos de conexão
	PingFunc         func(ctx context.Context) error
	QueryContextFunc func(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContextFunc  func(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	BeginTxFunc      func(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	CloseFunc        func() error
}

// Ping implementa a interface de conexão
func (m *MockConn) Ping(ctx context.Context) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	return errors.New("mock não implementado: Ping")
}

// QueryContext implementa a interface de conexão
func (m *MockConn) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if m.QueryContextFunc != nil {
		return m.QueryContextFunc(ctx, query, args...)
	}
	return nil, errors.New("mock não implementado: QueryContext")
}

// ExecContext implementa a interface de conexão
func (m *MockConn) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if m.ExecContextFunc != nil {
		return m.ExecContextFunc(ctx, query, args...)
	}
	return nil, errors.New("mock não implementado: ExecContext")
}

// BeginTx implementa a interface de conexão
func (m *MockConn) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx, opts)
	}
	return nil, errors.New("mock não implementado: BeginTx")
}

// Close implementa a interface de conexão
func (m *MockConn) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return errors.New("mock não implementado: Close")
}

// MockResult implementa sql.Result para testes
type MockResult struct {
	LastInsertIDFunc func() (int64, error)
	RowsAffectedFunc func() (int64, error)
}

// LastInsertId implementa a interface sql.Result
func (m *MockResult) LastInsertId() (int64, error) {
	if m.LastInsertIDFunc != nil {
		return m.LastInsertIDFunc()
	}
	return 0, errors.New("mock não implementado: LastInsertId")
}

// RowsAffected implementa a interface sql.Result
func (m *MockResult) RowsAffected() (int64, error) {
	if m.RowsAffectedFunc != nil {
		return m.RowsAffectedFunc()
	}
	return 0, errors.New("mock não implementado: RowsAffected")
}
