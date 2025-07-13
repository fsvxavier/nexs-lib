package gpgx

import "github.com/jackc/pgx/v5"

type pgxRows struct {
	rows pgx.Rows
}

func NewPgxRows(rows pgx.Rows) IRows {
	return pgxRows{
		rows: rows,
	}
}

func (r pgxRows) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (r pgxRows) Close() {
	r.rows.Close()
}

func (r pgxRows) Next() bool {
	return r.rows.Next()
}

func (r pgxRows) RawValues() [][]byte {
	return r.rows.RawValues()
}

func (r pgxRows) Err() error {
	return r.rows.Err()
}
