package interfaces

import (
	"context"
	"time"
)

// DatabaseProvider represents a generic database provider that can work with multiple drivers
type DatabaseProvider interface {
	Connect() error
	Close() error
	DB() any // Returns the underlying driver instance
	Pool() IPool
}

// IConn defines the connection interface compatible with multiple drivers
type IConn interface {
	QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error)
	Query(ctx context.Context, query string, args ...interface{}) (IRows, error)
	Exec(ctx context.Context, query string, args ...interface{}) error

	SendBatch(ctx context.Context, batch IBatch) (IBatchResults, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) (IRow, error)
	QueryRows(ctx context.Context, query string, args ...interface{}) (IRows, error)

	BeforeReleaseHook(ctx context.Context) error
	AfterAcquireHook(ctx context.Context) error

	Release(ctx context.Context)
	Ping(ctx context.Context) error

	BeginTransaction(ctx context.Context) (ITransaction, error)
	BeginTransactionWithOptions(ctx context.Context, options TxOptions) (ITransaction, error)
}

// IPool defines the connection pool interface
type IPool interface {
	Acquire(ctx context.Context) (IConn, error)
	Close()
	Ping(ctx context.Context) error
	GetConnWithNotPresent(ctx context.Context, conn IConn) (IConn, func(), error)
	Stats() PoolStats
}

// ITransaction defines the transaction interface
type ITransaction interface {
	IConn
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// TxOptions defines transaction options
type TxOptions struct {
	IsolationLevel string
	AccessMode     string
	Deferrable     bool
	ReadOnly       bool
}

// IBatch defines the batch interface
type IBatch interface {
	Queue(query string, arguments ...any)
	Len() int
}

// IBatchResults defines the batch results interface
type IBatchResults interface {
	QueryOne(dst interface{}) error
	QueryAll(dst interface{}) error
	Exec() error
	Close()
}

// IRow defines the row interface
type IRow interface {
	Scan(dest ...any) error
}

// IRows defines the rows interface
type IRows interface {
	Scan(dest ...any) error
	Close()
	Next() bool
	RawValues() [][]byte
	Err() error
}

// PoolStats represents connection pool statistics
type PoolStats struct {
	AcquireCount         int64
	AcquireDuration      time.Duration
	AcquiredConns        int32
	CanceledAcquireCount int64
	ConstructingConns    int32
	EmptyAcquireCount    int64
	IdleConns            int32
	MaxConns             int32
	TotalConns           int32
}

// DriverType represents the database driver type
type DriverType string

const (
	DriverPGX  DriverType = "pgx"
	DriverGORM DriverType = "gorm"
	DriverPQ   DriverType = "pq"
)
