//go:build unit

package pgx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPGXBatch(t *testing.T) {
	conn := &PGXConn{}
	batch := newPGXBatch(conn)

	assert.NotNil(t, batch)
	assert.NotNil(t, batch.batch)
	assert.Equal(t, conn, batch.conn)
}

func TestNewPGXBatchResults(t *testing.T) {
	conn := &PGXConn{}

	br := newPGXBatchResults(nil, conn)

	assert.NotNil(t, br)
	assert.Equal(t, conn, br.conn)
	assert.False(t, br.closed)
}

func TestPGXBatchResults_Closed(t *testing.T) {
	br := &PGXBatchResults{
		results: nil,
		conn:    &PGXConn{},
		closed:  true,
	}

	// Test QueryRow with closed
	row := br.QueryRow()
	assert.NotNil(t, row)

	// Test Query with closed
	rows, err := br.Query()
	assert.Error(t, err)
	assert.Nil(t, rows)

	// Test Exec with closed
	tag, err := br.Exec()
	assert.Error(t, err)
	assert.NotNil(t, tag)

	// Test Err with closed
	err = br.Err()
	assert.NoError(t, err)

	// Test Close when already closed
	err = br.Close()
	assert.NoError(t, err)
}
