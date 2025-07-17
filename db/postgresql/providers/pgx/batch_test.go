//go:build unit

package pgx

import (
	"testing"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
	"github.com/stretchr/testify/assert"
)

func TestPGXBatch(t *testing.T) {
	t.Run("NewBatch", func(t *testing.T) {
		batch := NewBatch()
		assert.NotNil(t, batch)
		assert.Equal(t, 0, batch.Len())
	})

	t.Run("Queue", func(t *testing.T) {
		batch := NewBatch()
		batch.Queue("SELECT $1", 1)
		assert.Equal(t, 1, batch.Len())
	})

	t.Run("QueueFunc", func(t *testing.T) {
		batch := NewBatch()
		called := false
		callback := func(br interfaces.IBatchResults) error {
			called = true
			return nil
		}

		batch.QueueFunc("SELECT $1", []any{1}, callback)
		assert.Equal(t, 1, batch.Len())
		assert.False(t, called) // Callback should not be called until execution
	})

	t.Run("Clear", func(t *testing.T) {
		batch := NewBatch()
		batch.Queue("SELECT 1")
		batch.Queue("SELECT 2")
		assert.Equal(t, 2, batch.Len())

		batch.Clear()
		assert.Equal(t, 0, batch.Len())
	})

	t.Run("Reset", func(t *testing.T) {
		batch := NewBatch()
		batch.Queue("SELECT 1")
		batch.Queue("SELECT 2")
		assert.Equal(t, 2, batch.Len())

		batch.Reset()
		assert.Equal(t, 0, batch.Len())
	})

	t.Run("Multiple operations", func(t *testing.T) {
		batch := NewBatch()

		// Test multiple queue operations
		batch.Queue("INSERT INTO test (id) VALUES ($1)", 1)
		batch.Queue("UPDATE test SET name = $1 WHERE id = $2", "test", 1)
		batch.Queue("DELETE FROM test WHERE id = $1", 1)

		assert.Equal(t, 3, batch.Len())

		// Test clear and reuse
		batch.Clear()
		assert.Equal(t, 0, batch.Len())

		batch.Queue("SELECT 1")
		assert.Equal(t, 1, batch.Len())
	})

	t.Run("Empty query handling", func(t *testing.T) {
		batch := NewBatch()

		// Should still add empty queries (let database handle errors)
		batch.Queue("")
		assert.Equal(t, 1, batch.Len())

		batch.Queue("  ")
		assert.Equal(t, 2, batch.Len())
	})

	t.Run("Nil arguments", func(t *testing.T) {
		batch := NewBatch()

		// Should handle nil arguments gracefully
		batch.Queue("SELECT 1", nil)
		assert.Equal(t, 1, batch.Len())

		batch.QueueFunc("SELECT 1", nil, func(br interfaces.IBatchResults) error { return nil })
		assert.Equal(t, 2, batch.Len())
	})

	t.Run("Complex arguments", func(t *testing.T) {
		batch := NewBatch()

		now := time.Now()
		complexArgs := []any{
			1,
			"string",
			true,
			now,
			[]byte("binary"),
			nil,
		}

		batch.Queue("INSERT INTO test VALUES ($1, $2, $3, $4, $5, $6)", complexArgs...)
		assert.Equal(t, 1, batch.Len())
	})

	t.Run("Thread safety", func(t *testing.T) {
		batch := NewBatch()

		// Note: PGX batch is not thread-safe by design
		// This test ensures our wrapper doesn't add thread safety
		// Users should manage concurrency themselves

		// Sequential operations should work
		for i := 0; i < 10; i++ {
			batch.Queue("SELECT $1", i)
		}
		assert.Equal(t, 10, batch.Len())
	})
}

func TestPGXBatchResults(t *testing.T) {
	// Note: PGXBatchResults tests would require actual database connection
	// These are placeholder tests for the interface compliance

	t.Run("Interface compliance", func(t *testing.T) {
		// Verify that PGXBatchResults implements IBatchResults
		var _ interfaces.IBatchResults = (*PGXBatchResults)(nil)
	})

	t.Run("Close should not panic", func(t *testing.T) {
		results := &PGXBatchResults{}

		// Close should be safe to call multiple times
		assert.NotPanics(t, func() {
			results.Close()
			results.Close()
		})
	})

	t.Run("Err should not panic", func(t *testing.T) {
		results := &PGXBatchResults{}

		// Err should not panic when called on empty results
		assert.NotPanics(t, func() {
			err := results.Err()
			// err might be nil or an actual error, both are valid
			_ = err
		})
	})
}

// Benchmark tests
func BenchmarkPGXBatch_Queue(b *testing.B) {
	batch := NewBatch()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch.Queue("SELECT $1", i)
	}
}

func BenchmarkPGXBatch_QueueFunc(b *testing.B) {
	batch := NewBatch()
	callback := func(br interfaces.IBatchResults) error { return nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch.QueueFunc("SELECT $1", []any{i}, callback)
	}
}

func BenchmarkPGXBatch_Clear(b *testing.B) {
	batch := NewBatch()

	// Pre-populate batch
	for i := 0; i < 100; i++ {
		batch.Queue("SELECT $1", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch.Clear()
		// Re-populate for next iteration
		if i < b.N-1 {
			for j := 0; j < 100; j++ {
				batch.Queue("SELECT $1", j)
			}
		}
	}
}
