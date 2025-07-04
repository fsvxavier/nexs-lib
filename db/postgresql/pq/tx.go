package pq

import (
	"context"
	"database/sql"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
)

// Transaction implementa a interface common.ITransaction usando database/sql
type Transaction struct {
	tx                 *sql.Tx
	config             *common.Config
	multiTenantEnabled bool
}

// QueryOne executa uma consulta e digitaliza uma única linha no destino
func (t *Transaction) QueryOne(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanOne(rows, dst)
}

// QueryAll executa uma consulta e digitaliza todas as linhas no destino
func (t *Transaction) QueryAll(ctx context.Context, dst interface{}, query string, args ...interface{}) error {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanAll(rows, dst)
}

// QueryCount executa uma consulta e retorna uma contagem
func (t *Transaction) QueryCount(ctx context.Context, query string, args ...interface{}) (*int, error) {
	var count int

	row := t.tx.QueryRowContext(ctx, query, args...)
	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

// Query executa uma consulta e retorna as linhas resultantes
func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (common.IRows, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows: rows}, nil
}

// QueryRow executa uma consulta e retorna uma única linha
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) (common.IRow, error) {
	row := t.tx.QueryRowContext(ctx, query, args...)
	return &Row{row: row}, nil
}

// Exec executa um comando (sem retorno de linhas)
func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := t.tx.ExecContext(ctx, query, args...)
	return err
}

// SendBatch envia um lote de comandos para execução
func (t *Transaction) SendBatch(ctx context.Context, batch common.IBatch) (common.IBatchResults, error) {
	pqBatch, ok := batch.(*Batch)
	if !ok {
		return nil, common.ErrInvalidOperation
	}

	return &BatchResults{
		tx:    t.tx,
		ctx:   ctx,
		batch: pqBatch,
		idx:   0,
	}, nil
}

// Ping verifica a conectividade com o banco de dados
func (t *Transaction) Ping(ctx context.Context) error {
	// No database/sql, não há um método Ping específico para transações
	// Podemos executar uma consulta simples para verificar a conectividade
	_, err := t.tx.ExecContext(ctx, "SELECT 1")
	return err
}

// Close não é aplicável a transações e retorna um erro
func (t *Transaction) Close(ctx context.Context) error {
	return common.ErrInvalidOperation
}

// BeginTransaction retorna um erro, pois não é possível iniciar uma transação dentro de outra
func (t *Transaction) BeginTransaction(ctx context.Context) (common.ITransaction, error) {
	return nil, common.ErrInvalidNestedTransaction
}

// BeginTransactionWithOptions retorna um erro, pois não é possível iniciar uma transação dentro de outra
func (t *Transaction) BeginTransactionWithOptions(ctx context.Context, opts *common.TxOptions) (common.ITransaction, error) {
	return nil, common.ErrInvalidNestedTransaction
}

// Commit confirma a transação
func (t *Transaction) Commit(ctx context.Context) error {
	return t.tx.Commit()
}

// Rollback reverte a transação
func (t *Transaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback()
}
