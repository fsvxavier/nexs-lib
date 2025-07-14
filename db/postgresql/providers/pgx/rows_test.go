//go:build unit

package pgx

import (
	"testing"

	"github.com/fsvxavier/nexs-lib/db/postgresql"
)

func TestRow_Scan(t *testing.T) {
	// Create a mock row - we can't easily test this without a real connection
	// But we can test the interface compliance
	var _ postgresql.IRow = &Row{}
}

func TestRows_Interface(t *testing.T) {
	// Test interface compliance
	var _ postgresql.IRows = &Rows{}
}

func TestBatch_Queue(t *testing.T) {
	batch := NewBatch()

	// Test initial state
	if batch.Len() != 0 {
		t.Errorf("Expected new batch to have length 0, got %d", batch.Len())
	}

	// Queue some operations
	batch.Queue("INSERT INTO users (name) VALUES ($1)", "Alice")
	if batch.Len() != 1 {
		t.Errorf("Expected batch length to be 1, got %d", batch.Len())
	}

	batch.Queue("INSERT INTO users (name) VALUES ($1)", "Bob")
	if batch.Len() != 2 {
		t.Errorf("Expected batch length to be 2, got %d", batch.Len())
	}

	batch.Queue("SELECT COUNT(*) FROM users")
	if batch.Len() != 3 {
		t.Errorf("Expected batch length to be 3, got %d", batch.Len())
	}
}

func TestBatch_Clear(t *testing.T) {
	batch := NewBatch().(*Batch)

	// Add some operations
	batch.Queue("INSERT INTO users (name) VALUES ($1)", "Alice")
	batch.Queue("INSERT INTO users (name) VALUES ($1)", "Bob")

	if batch.Len() != 2 {
		t.Errorf("Expected batch length to be 2, got %d", batch.Len())
	}

	// Clear the batch
	batch.Clear()

	if batch.Len() != 0 {
		t.Errorf("Expected batch length to be 0 after clear, got %d", batch.Len())
	}
}

func TestBatch_Interface(t *testing.T) {
	// Test interface compliance
	var _ postgresql.IBatch = &Batch{}
	var _ postgresql.IBatch = NewBatch()
}

func TestBatchResults_Interface(t *testing.T) {
	// Test interface compliance
	var _ postgresql.IBatchResults = &BatchResults{}
}

func TestBatchResults_Err(t *testing.T) {
	br := &BatchResults{}

	// Should not panic and should return nil
	err := br.Err()
	if err != nil {
		t.Errorf("Expected Err() to return nil for empty BatchResults, got %v", err)
	}
}

func TestNewBatch(t *testing.T) {
	batch := NewBatch()

	if batch == nil {
		t.Error("Expected NewBatch() to return non-nil batch")
	}

	if batch.Len() != 0 {
		t.Errorf("Expected new batch to have length 0, got %d", batch.Len())
	}

	// Test that it's the correct type
	pgxBatch, ok := batch.(*Batch)
	if !ok {
		t.Error("Expected NewBatch() to return *Batch type")
	}

	if pgxBatch.batch == nil {
		t.Error("Expected internal pgx.Batch to be initialized")
	}
}

// Test edge cases and error handling

func TestBatch_EdgeCases(t *testing.T) {
	batch := NewBatch().(*Batch)

	// Test queuing with nil arguments
	batch.Queue("SELECT 1")
	if batch.Len() != 1 {
		t.Errorf("Expected batch length to be 1, got %d", batch.Len())
	}

	// Test queuing with empty query
	batch.Queue("")
	if batch.Len() != 2 {
		t.Errorf("Expected batch length to be 2, got %d", batch.Len())
	}

	// Test queuing with multiple arguments
	batch.Queue("INSERT INTO users (name, email, age) VALUES ($1, $2, $3)", "John", "john@example.com", 30)
	if batch.Len() != 3 {
		t.Errorf("Expected batch length to be 3, got %d", batch.Len())
	}
}

func TestBatch_MultipleClears(t *testing.T) {
	batch := NewBatch().(*Batch)

	// Add operations
	batch.Queue("SELECT 1")
	batch.Queue("SELECT 2")

	// Clear multiple times
	batch.Clear()
	if batch.Len() != 0 {
		t.Errorf("Expected batch length to be 0 after first clear, got %d", batch.Len())
	}

	batch.Clear()
	if batch.Len() != 0 {
		t.Errorf("Expected batch length to be 0 after second clear, got %d", batch.Len())
	}

	// Add operation after clear
	batch.Queue("SELECT 3")
	if batch.Len() != 1 {
		t.Errorf("Expected batch length to be 1 after adding operation post-clear, got %d", batch.Len())
	}
}
