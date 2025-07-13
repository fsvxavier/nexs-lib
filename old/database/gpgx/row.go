package gpgx

import "github.com/jackc/pgx/v5"

type pgxRow struct {
	row pgx.Row
}

func NewPgxRow(row pgx.Row) IRow {
	return pgxRow{
		row: row,
	}
}

func (r pgxRow) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}
