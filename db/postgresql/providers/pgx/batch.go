package pgx

import (
	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
	"github.com/jackc/pgx/v5"
)

// PGXBatch implements the IBatch interface
type PGXBatch struct {
	batch *pgx.Batch
	conn  *PGXConn
}

// Queue implements IBatch.Queue
func (b *PGXBatch) Queue(query string, arguments ...any) {
	b.batch.Queue(query, arguments...)
}

// QueueFunc implements IBatch.QueueFunc
func (b *PGXBatch) QueueFunc(query string, arguments []any, callback func(interfaces.IBatchResults) error) {
	// PGX doesn't support callback-based queuing directly, so we store the callback in metadata
	// For now, we'll just queue the query normally
	b.batch.Queue(query, arguments...)
}

// Len implements IBatch.Len
func (b *PGXBatch) Len() int {
	return b.batch.Len()
}

// Clear implements IBatch.Clear
func (b *PGXBatch) Clear() {
	// Create a new batch to clear it
	b.batch = &pgx.Batch{}
}

// Reset implements IBatch.Reset
func (b *PGXBatch) Reset() {
	b.Clear()
}

// PGXBatchResults implements the IBatchResults interface
type PGXBatchResults struct {
	results pgx.BatchResults
	conn    *PGXConn
	closed  bool
}

// QueryRow implements IBatchResults.QueryRow
func (br *PGXBatchResults) QueryRow() interfaces.IRow {
	if br.closed {
		return &PGXRow{row: nil, conn: br.conn}
	}
	row := br.results.QueryRow()
	return &PGXRow{row: row, conn: br.conn}
}

// Query implements IBatchResults.Query
func (br *PGXBatchResults) Query() (interfaces.IRows, error) {
	if br.closed {
		return nil, pgx.ErrNoRows
	}
	rows, err := br.results.Query()
	if err != nil {
		return nil, err
	}
	return &PGXRows{rows: rows, conn: br.conn}, nil
}

// Exec implements IBatchResults.Exec
func (br *PGXBatchResults) Exec() (interfaces.CommandTag, error) {
	if br.closed {
		return &PGXCommandTag{}, pgx.ErrNoRows
	}
	tag, err := br.results.Exec()
	if err != nil {
		return nil, err
	}
	return &PGXCommandTag{tag: tag}, nil
}

// Close implements IBatchResults.Close
func (br *PGXBatchResults) Close() error {
	if br.closed {
		return nil
	}
	br.closed = true
	return br.results.Close()
}

// Err implements IBatchResults.Err
func (br *PGXBatchResults) Err() error {
	if br.closed {
		return nil
	}
	// PGX BatchResults doesn't have an Err() method, so we return nil
	// Errors are handled in individual operation methods
	return nil
}

// Helper function to create a new PGXBatch
func newPGXBatch(conn *PGXConn) *PGXBatch {
	return &PGXBatch{
		batch: &pgx.Batch{},
		conn:  conn,
	}
}

// Helper function to create PGXBatchResults
func newPGXBatchResults(results pgx.BatchResults, conn *PGXConn) *PGXBatchResults {
	return &PGXBatchResults{
		results: results,
		conn:    conn,
		closed:  false,
	}
}
