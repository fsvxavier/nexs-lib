package pgx

import (
	"github.com/fsvxavier/nexs-lib/db/postgresql"
	"github.com/jackc/pgx/v5"
)

// Row implements postgresql.IRow
type Row struct {
	row pgx.Row
}

// Scan scans the row into destinations
func (r *Row) Scan(dest ...interface{}) error {
	return r.row.Scan(dest...)
}

// Rows implements postgresql.IRows
type Rows struct {
	rows pgx.Rows
}

// Scan scans the current row into destinations
func (r *Rows) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

// Close closes the rows
func (r *Rows) Close() {
	r.rows.Close()
}

// Next advances to the next row
func (r *Rows) Next() bool {
	return r.rows.Next()
}

// Err returns any error that occurred during iteration
func (r *Rows) Err() error {
	return r.rows.Err()
}

// RawValues returns the raw bytes of the current row's values
func (r *Rows) RawValues() [][]byte {
	return r.rows.RawValues()
}

// Batch implements postgresql.IBatch
type Batch struct {
	batch *pgx.Batch
}

// NewBatch creates a new batch
func NewBatch() postgresql.IBatch {
	return &Batch{
		batch: &pgx.Batch{},
	}
}

// Queue adds a query to the batch
func (b *Batch) Queue(query string, arguments ...interface{}) {
	b.batch.Queue(query, arguments...)
}

// Len returns the number of queued queries
func (b *Batch) Len() int {
	return b.batch.Len()
}

// Clear removes all queued queries
func (b *Batch) Clear() {
	// Create a new batch to clear
	b.batch = &pgx.Batch{}
}

// BatchResults implements postgresql.IBatchResults
type BatchResults struct {
	results pgx.BatchResults
}

// QueryOne scans a single row from the current query result
func (br *BatchResults) QueryOne(dst interface{}) error {
	rows, err := br.results.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		if rows.Err() != nil {
			return rows.Err()
		}
		return pgx.ErrNoRows
	}

	return rows.Scan(dst)
}

// QueryAll scans all rows from the current query result
func (br *BatchResults) QueryAll(dst interface{}) error {
	rows, err := br.results.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	// TODO: Implement proper scanning to dst slice/array
	// For now, this is a placeholder implementation
	return nil
}

// Exec executes the current query and advances to the next
func (br *BatchResults) Exec() error {
	_, err := br.results.Exec()
	return err
}

// Close closes the batch results
func (br *BatchResults) Close() {
	br.results.Close()
}

// Err returns the error, if any, that was encountered during iteration
func (br *BatchResults) Err() error {
	// pgx.BatchResults doesn't have an Err() method like Rows
	// Error handling is done through individual operation results
	return nil
}
