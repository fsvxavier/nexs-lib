package gorm

import (
	"database/sql"
)

// Row implementa a interface common.IRow usando GORM
type Row struct {
	row *sql.Row
}

// Scan escaneia os valores da linha para os destinos fornecidos
func (r *Row) Scan(dest ...any) error {
	err := r.row.Scan(dest...)
	if err != nil {
		return WrapError(err, "falha ao escanear linha")
	}
	return nil
}
