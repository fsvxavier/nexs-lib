package pgx

import (
	"github.com/fsvxavier/nexs-lib/db/postgresql/common"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

// Row implementa a interface common.IRow usando pgx
type Row struct {
	row pgx.Row
}

// Scan digitaliza os valores da linha nos destinos fornecidos
func (r *Row) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}

// Rows implementa a interface common.IRows usando pgx
type Rows struct {
	rows pgx.Rows
}

// Scan digitaliza os valores da linha atual nos destinos fornecidos
func (r *Rows) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

// Close fecha o resultado da consulta
func (r *Rows) Close() error {
	r.rows.Close()
	return nil
}

// Next avança para a próxima linha
func (r *Rows) Next() bool {
	return r.rows.Next()
}

// Err retorna qualquer erro que ocorreu durante a iteração
func (r *Rows) Err() error {
	return r.rows.Err()
}

// Batch implementa a interface common.IBatch usando pgx
type Batch struct {
	batch *pgx.Batch
}

// NewBatch cria um novo lote de consultas
func NewBatch() common.IBatch {
	return &Batch{
		batch: &pgx.Batch{},
	}
}

// Queue adiciona uma consulta ao lote
func (b *Batch) Queue(query string, arguments ...any) {
	b.batch.Queue(query, arguments...)
}

// GetBatch retorna o objeto de lote subjacente
func (b *Batch) GetBatch() interface{} {
	return b.batch
}

// BatchResults implementa a interface common.IBatchResults usando pgx
type BatchResults struct {
	br pgx.BatchResults
}

// QueryOne executa uma consulta do lote e digitaliza uma única linha no destino
func (br *BatchResults) QueryOne(dst interface{}) error {
	rows, err := br.br.Query()
	if err != nil {
		return err
	}
	defer rows.Close()
	return pgxscan.ScanOne(dst, rows)
}

// QueryAll executa uma consulta do lote e digitaliza todas as linhas no destino
func (br *BatchResults) QueryAll(dst interface{}) error {
	rows, err := br.br.Query()
	if err != nil {
		return err
	}
	defer rows.Close()
	return pgxscan.ScanAll(dst, rows)
}

// Exec executa um comando do lote (sem retorno de linhas)
func (br *BatchResults) Exec() error {
	_, err := br.br.Exec()
	return err
}

// Close fecha os resultados do lote
func (br *BatchResults) Close() error {
	br.br.Close()
	return nil
}
