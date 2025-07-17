// Package pgxprovider contém as interfaces internas e definições compartilhadas
// do provider PGX para PostgreSQL
package pgxprovider

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Erros personalizados do provider PGX
// Estes erros são usados em todo o provider para manter consistência
var (
	// ErrPoolClosed é retornado quando uma operação é tentada em um pool fechado
	ErrPoolClosed = errors.New("pool is closed")

	// ErrUnhealthyState é retornado quando o pool está em um estado não saudável
	ErrUnhealthyState = errors.New("unhealthy state detected")

	// ErrConnClosed é retornado quando uma operação é tentada em uma conexão fechada
	ErrConnClosed = errors.New("connection is closed")
)

// pgxConnInterface representa a interface comum para *pgx.Conn e *pgxpool.Conn
// Esta é uma interface interna do provider PGX para abstrair as diferenças
// entre conexões diretas e conexões de pool do pgx.
//
// Esta interface inclui apenas os métodos que são comuns a ambos os tipos,
// permitindo que o código seja reutilizado independentemente do tipo de conexão.
type pgxConnInterface interface {
	// Operações de consulta
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)

	// Operações de execução
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)

	// Operações de batch
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults

	// Operações de transação
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)

	// Operações de conectividade
	Ping(ctx context.Context) error

	// Operações de cópia (PostgreSQL específico)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}
