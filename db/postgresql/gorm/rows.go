package gorm

import (
	"database/sql"
)

// Rows implementa a interface common.IRows usando GORM
type Rows struct {
	rows *sql.Rows
}

// Scan escaneia os valores da linha atual para os destinos fornecidos
func (r *Rows) Scan(dest ...any) error {
	err := r.rows.Scan(dest...)
	if err != nil {
		return WrapError(err, "falha ao escanear linhas")
	}
	return nil
}

// Close fecha o resultado da consulta
func (r *Rows) Close() error {
	err := r.rows.Close()
	if err != nil {
		return WrapError(err, "falha ao fechar linhas")
	}
	return nil
}

// Next avança para a próxima linha
func (r *Rows) Next() bool {
	return r.rows.Next()
}

// Err retorna qualquer erro que tenha ocorrido durante a iteração
func (r *Rows) Err() error {
	if err := r.rows.Err(); err != nil {
		return WrapError(err, "erro ao iterar sobre linhas")
	}
	return nil
}
