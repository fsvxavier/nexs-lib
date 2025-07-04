package common

import (
	"context"
)

// IConn define a interface para uma conexão com banco de dados PostgreSQL
type IConn interface {
	// Métodos para consultas
	QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error
	QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error)
	Query(ctx context.Context, query string, args ...interface{}) (IRows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) (IRow, error)

	// Execução de comandos
	Exec(ctx context.Context, query string, args ...interface{}) error

	// Gerenciamento de lotes (batch)
	SendBatch(ctx context.Context, batch IBatch) (IBatchResults, error)

	// Gerenciamento de conexão
	Ping(ctx context.Context) error
	Close(ctx context.Context) error

	// Gerenciamento de transação
	BeginTransaction(ctx context.Context) (ITransaction, error)
	BeginTransactionWithOptions(ctx context.Context, opts *TxOptions) (ITransaction, error)
}

// IPool define a interface para um pool de conexões com banco de dados PostgreSQL
type IPool interface {
	// Adquire uma conexão do pool
	Acquire(ctx context.Context) (IConn, error)

	// Fecha o pool de conexões
	Close() error

	// Verifica conexão com o banco de dados
	Ping(ctx context.Context) error

	// Retorna estatísticas do pool
	Stats() *PoolStats
}

// ITransaction define a interface para transações no PostgreSQL
type ITransaction interface {
	IConn
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// TxOptions define opções para transações
type TxOptions struct {
	IsoLevel   IsolationLevel
	ReadOnly   bool
	Deferrable bool
}

// IsolationLevel define os níveis de isolamento para transações
type IsolationLevel string

const (
	ReadCommitted  IsolationLevel = "READ COMMITTED"
	RepeatableRead IsolationLevel = "REPEATABLE READ"
	Serializable   IsolationLevel = "SERIALIZABLE"
)

// IBatch define a interface para operações em lote
type IBatch interface {
	Queue(query string, arguments ...any)
	GetBatch() interface{}
}

// IBatchResults define a interface para resultados de operações em lote
type IBatchResults interface {
	QueryOne(dst interface{}) error
	QueryAll(dst interface{}) error
	Exec() error
	Close() error
}

// IRow define a interface para uma única linha de resultado
type IRow interface {
	Scan(dest ...any) error
}

// IRows define a interface para múltiplas linhas de resultado
type IRows interface {
	Scan(dest ...any) error
	Close() error
	Next() bool
	Err() error
}

// PoolStats contém estatísticas do pool de conexões
type PoolStats struct {
	AcquiredConns    int
	TotalConns       int
	IdleConns        int
	MaxConns         int
	MinConns         int
	ConstructedConns int
}
