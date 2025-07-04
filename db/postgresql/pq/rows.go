package pq

import (
	"context"
	"database/sql"
	"errors"

	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
)

// Row implementa a interface common.IRow usando database/sql
type Row struct {
	row *sql.Row
}

// Scan digitaliza os valores da linha nos destinos fornecidos
func (r *Row) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}

// Rows implementa a interface common.IRows usando database/sql
type Rows struct {
	rows *sql.Rows
}

// Scan digitaliza os valores da linha atual nos destinos fornecidos
func (r *Rows) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

// Close fecha o resultado da consulta
func (r *Rows) Close() error {
	return r.rows.Close()
}

// Next avança para a próxima linha
func (r *Rows) Next() bool {
	return r.rows.Next()
}

// Err retorna qualquer erro que ocorreu durante a iteração
func (r *Rows) Err() error {
	return r.rows.Err()
}

// Batch implementa a interface common.IBatch usando database/sql
type Batch struct {
	queries []query
}

type query struct {
	sql  string
	args []interface{}
}

// NewBatch cria um novo lote de consultas
func NewBatch() common.IBatch {
	return &Batch{
		queries: make([]query, 0),
	}
}

// Queue adiciona uma consulta ao lote
func (b *Batch) Queue(sql string, args ...any) {
	b.queries = append(b.queries, query{sql: sql, args: args})
}

// GetBatch retorna o objeto de lote subjacente
func (b *Batch) GetBatch() interface{} {
	return b.queries
}

// BatchResults implementa a interface common.IBatchResults usando database/sql
type BatchResults struct {
	tx    *sql.Tx
	ctx   context.Context
	batch *Batch
	idx   int
}

// QueryOne executa uma consulta do lote e digitaliza uma única linha no destino
func (br *BatchResults) QueryOne(dst interface{}) error {
	if br.idx >= len(br.batch.queries) {
		return errors.New("índice de consulta fora dos limites")
	}

	q := br.batch.queries[br.idx]
	br.idx++

	rows, err := br.tx.QueryContext(br.ctx, q.sql, q.args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanOne(rows, dst)
}

// QueryAll executa uma consulta do lote e digitaliza todas as linhas no destino
func (br *BatchResults) QueryAll(dst interface{}) error {
	if br.idx >= len(br.batch.queries) {
		return errors.New("índice de consulta fora dos limites")
	}

	q := br.batch.queries[br.idx]
	br.idx++

	rows, err := br.tx.QueryContext(br.ctx, q.sql, q.args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanAll(rows, dst)
}

// Exec executa um comando do lote (sem retorno de linhas)
func (br *BatchResults) Exec() error {
	if br.idx >= len(br.batch.queries) {
		return errors.New("índice de consulta fora dos limites")
	}

	q := br.batch.queries[br.idx]
	br.idx++

	_, err := br.tx.ExecContext(br.ctx, q.sql, q.args...)
	return err
}

// Close fecha os resultados do lote e confirma ou reverte a transação
func (br *BatchResults) Close() error {
	// Se todas as consultas foram executadas com sucesso, confirmamos a transação
	if br.idx >= len(br.batch.queries) {
		return br.tx.Commit()
	}

	// Caso contrário, revertemos a transação
	return br.tx.Rollback()
}
