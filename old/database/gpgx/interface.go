package gpgx

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IConn interface {
	QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error)
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, query string, args ...interface{}) error

	SendBatch(ctx context.Context, batch IBatch) (IBatchResults, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) (IRow, error)
	QueryRows(ctx context.Context, query string, args ...interface{}) (IRows, error)

	BeforeReleaseHook(ctx context.Context) (err error)
	AfterAcquireHook(ctx context.Context) (err error)

	Release(ctx context.Context)
	Ping(ctx context.Context) error

	BeginTransaction(ctx context.Context) (ITransaction, error)
}

type IPool interface {
	Acquire(ctx context.Context) (IConn, error)
	Close()
	Ping(ctx context.Context) error
	GetConnWithNotPresent(ctx context.Context, conn IConn) (IConn, func(), error)
}

type Conn interface {
	Pool() *pgxpool.Pool
}

type ITransaction interface {
	IConn
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type ITxOptions struct {
	IsoLevel       pgx.TxIsoLevel
	AccessMode     pgx.TxAccessMode
	DeferrableMode pgx.TxDeferrableMode
}

type IBatch interface {
	Queue(query string, arguments ...any)
	getBatch() *pgx.Batch
}

type IBatchResults interface {
	QueryOne(dst interface{}) error
	QueryAll(dst interface{}) error
	Exec() error
	Close()
}

type IRow interface {
	Scan(dest ...any) error
}

type IRows interface {
	Scan(dest ...any) error
	Close()
	Next() bool
	RawValues() [][]byte
	Err() error
}
