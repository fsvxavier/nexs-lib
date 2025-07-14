package gorm

import (
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

// Rows represents query results for GORM
type Rows struct {
	rows *sql.Rows
}

// Next moves to the next row
func (r *Rows) Next() bool {
	if r.rows == nil {
		return false
	}
	return r.rows.Next()
}

// Scan scans the current row
func (r *Rows) Scan(dest ...interface{}) error {
	if r.rows == nil {
		return fmt.Errorf("rows is nil")
	}
	return r.rows.Scan(dest...)
}

// Close closes the rows
func (r *Rows) Close() {
	if r.rows != nil {
		r.rows.Close()
	}
}

// Err returns any error encountered during iteration
func (r *Rows) Err() error {
	if r.rows == nil {
		return nil
	}
	return r.rows.Err()
}

// Columns returns column names
func (r *Rows) Columns() ([]string, error) {
	if r.rows == nil {
		return nil, fmt.Errorf("rows is nil")
	}
	return r.rows.Columns()
}

// ColumnTypes returns column types
func (r *Rows) ColumnTypes() ([]*sql.ColumnType, error) {
	if r.rows == nil {
		return nil, fmt.Errorf("rows is nil")
	}
	return r.rows.ColumnTypes()
}

// RawValues returns raw byte values
func (r *Rows) RawValues() [][]byte {
	// GORM/sql.Rows doesn't directly support raw values
	// This is a limitation of the GORM adapter
	return nil
}

// Values returns current row values
func (r *Rows) Values() ([]interface{}, error) {
	if r.rows == nil {
		return nil, fmt.Errorf("rows is nil")
	}

	cols, err := r.rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	err = r.rows.Scan(valuePtrs...)
	if err != nil {
		return nil, err
	}

	return values, nil
}

// Row represents a single row result for GORM
type Row struct {
	db    *gorm.DB
	query string
	args  []interface{}
}

// Scan scans the row into destination
func (r *Row) Scan(dest ...interface{}) error {
	if r.db == nil {
		return fmt.Errorf("db is nil")
	}

	row := r.db.Raw(r.query, r.args...).Row()
	return row.Scan(dest...)
}

// Err returns any error from the row
func (r *Row) Err() error {
	if r.db == nil {
		return fmt.Errorf("db is nil")
	}

	// Execute the query to check for errors
	result := r.db.Raw(r.query, r.args...)
	return result.Error
}
