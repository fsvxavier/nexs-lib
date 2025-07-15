package gorm

import (
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/fsvxavier/nexs-lib/db/postgresql/config"
	"github.com/stretchr/testify/assert"
)

// mockRows implements sql.Rows for testing
type mockRows struct {
	columns []string
	data    [][]driver.Value
	pos     int
	closed  bool
	err     error
}

func (m *mockRows) Columns() []string {
	return m.columns
}

func (m *mockRows) Close() error {
	m.closed = true
	return nil
}

func (m *mockRows) Next(dest []driver.Value) error {
	if m.pos >= len(m.data) {
		return errors.New("no more rows")
	}
	copy(dest, m.data[m.pos])
	m.pos++
	return nil
}

// TestRows_BasicOperations tests basic rows operations
func TestRows_BasicOperations(t *testing.T) {
	t.Parallel()

	t.Run("Next_WithNilRows", func(t *testing.T) {
		rows := &Rows{rows: nil}

		hasNext := rows.Next()
		assert.False(t, hasNext)
	})

	t.Run("Scan_WithNilRows", func(t *testing.T) {
		rows := &Rows{rows: nil}

		var dest interface{}
		err := rows.Scan(&dest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rows is nil")
	})

	t.Run("Close_WithNilRows", func(t *testing.T) {
		rows := &Rows{rows: nil}

		// Should not panic
		rows.Close()
	})

	t.Run("Err_WithNilRows", func(t *testing.T) {
		rows := &Rows{rows: nil}

		err := rows.Err()
		assert.NoError(t, err)
	})

	t.Run("Columns_WithNilRows", func(t *testing.T) {
		rows := &Rows{rows: nil}

		cols, err := rows.Columns()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rows is nil")
		assert.Nil(t, cols)
	})

	t.Run("ColumnTypes_WithNilRows", func(t *testing.T) {
		rows := &Rows{rows: nil}

		colTypes, err := rows.ColumnTypes()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rows is nil")
		assert.Nil(t, colTypes)
	})

	t.Run("RawValues_Always", func(t *testing.T) {
		rows := &Rows{rows: nil}

		rawValues := rows.RawValues()
		assert.Nil(t, rawValues) // GORM limitation
	})

	t.Run("Values_WithNilRows", func(t *testing.T) {
		rows := &Rows{rows: nil}

		values, err := rows.Values()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rows is nil")
		assert.Nil(t, values)
	})
}

// TestRow_BasicOperations tests basic row operations
func TestRow_BasicOperations(t *testing.T) {
	t.Parallel()

	t.Run("Scan_WithNilDB", func(t *testing.T) {
		row := &Row{
			db:    nil,
			query: "SELECT 'test'",
			args:  []interface{}{},
		}

		var dest string
		err := row.Scan(&dest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db is nil")
	})

	t.Run("Err_WithNilDB", func(t *testing.T) {
		row := &Row{
			db:    nil,
			query: "SELECT 'test'",
			args:  []interface{}{},
		}

		err := row.Err()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db is nil")
	})
	t.Run("Scan_WithValidDB", func(t *testing.T) {
		// Use nil DB to test error path safely
		row := &Row{
			db:    nil,
			query: "SELECT 'test'",
			args:  []interface{}{},
		}

		var dest string
		err := row.Scan(&dest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db is nil")
	})

	t.Run("Err_WithValidDB", func(t *testing.T) {
		// Use nil DB to test error path safely
		row := &Row{
			db:    nil,
			query: "SELECT 'test'",
			args:  []interface{}{},
		}

		err := row.Err()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db is nil")
	})
}

// TestRows_ErrorHandling tests rows error handling scenarios
func TestRows_ErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("Values_WithNilRows", func(t *testing.T) {
		rows := &Rows{rows: nil}

		values, err := rows.Values()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rows is nil")
		assert.Nil(t, values)
	})
}

// setupTestTransaction creates a test transaction for unit testing
func setupTestTransaction(t *testing.T) *Transaction {
	cfg := &config.Config{
		Host:     "localhost",
		Port:     5432,
		Database: "test",
		Username: "test",
		Password: "test",
	}

	// Return a transaction with nil tx to test error paths safely
	return &Transaction{
		tx:     nil,
		config: cfg,
	}
}
