package commons

import (
	"context"
	"testing"
	"time"
)

func TestNewMemoryIdempotencyManager(t *testing.T) {
	tests := []struct {
		name string
		ttl  time.Duration
		want bool // true if should succeed
	}{
		{
			name: "zero TTL",
			ttl:  0,
			want: true, // Should use default TTL
		},
		{
			name: "valid TTL",
			ttl:  1 * time.Hour,
			want: true,
		},
		{
			name: "short TTL",
			ttl:  100 * time.Millisecond,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMemoryIdempotencyManager(tt.ttl)

			if tt.want {
				if manager == nil {
					t.Error("NewMemoryIdempotencyManager() returned nil manager")
					return
				}
			}
		})
	}
}

func TestIdempotencyManager_IsProcessed(t *testing.T) {
	manager := NewMemoryIdempotencyManager(100 * time.Millisecond)
	ctx := context.Background()

	// Test first occurrence - should not be processed
	messageID := "test-msg-1"

	isProcessed, err := manager.IsProcessed(ctx, messageID)
	if err != nil {
		t.Errorf("IsProcessed() error = %v, want nil", err)
	}
	if isProcessed {
		t.Error("IsProcessed() = true, want false for first occurrence")
	}

	// Mark as processed
	err = manager.MarkAsProcessed(ctx, messageID)
	if err != nil {
		t.Errorf("MarkAsProcessed() error = %v, want nil", err)
	}

	// Test second occurrence - should be processed
	isProcessed, err = manager.IsProcessed(ctx, messageID)
	if err != nil {
		t.Errorf("IsProcessed() error = %v, want nil", err)
	}
	if !isProcessed {
		t.Error("IsProcessed() = false, want true after marking as processed")
	}

	// Test different message - should not be processed
	messageID2 := "test-msg-2"

	isProcessed, err = manager.IsProcessed(ctx, messageID2)
	if err != nil {
		t.Errorf("IsProcessed() error = %v, want nil", err)
	}
	if isProcessed {
		t.Error("IsProcessed() = true, want false for different message")
	}
}

func TestIdempotencyManager_TTL(t *testing.T) {
	manager := NewMemoryIdempotencyManager(50 * time.Millisecond) // Very short TTL for testing
	ctx := context.Background()

	messageID := "test-msg-ttl"

	// Mark as processed
	err := manager.MarkAsProcessed(ctx, messageID)
	if err != nil {
		t.Errorf("MarkAsProcessed() error = %v, want nil", err)
	}

	// Immediate check - should be processed
	isProcessed, err := manager.IsProcessed(ctx, messageID)
	if err != nil {
		t.Errorf("IsProcessed() error = %v, want nil", err)
	}
	if !isProcessed {
		t.Error("IsProcessed() = false, want true immediately after marking")
	}

	// Wait for TTL to expire
	time.Sleep(60 * time.Millisecond)

	// Check after TTL - should not be processed anymore
	isProcessed, err = manager.IsProcessed(ctx, messageID)
	if err != nil {
		t.Errorf("IsProcessed() error = %v, want nil", err)
	}
	if isProcessed {
		t.Error("IsProcessed() = true, want false after TTL expiration")
	}
}

func TestIdempotencyManager_MarkAsProcessedWithTTL(t *testing.T) {
	manager := NewMemoryIdempotencyManager(1 * time.Hour) // Long default TTL
	ctx := context.Background()

	messageID := "test-msg-custom-ttl"

	// Mark as processed with custom short TTL
	customTTL := 50 * time.Millisecond
	err := manager.MarkAsProcessedWithTTL(ctx, messageID, customTTL)
	if err != nil {
		t.Errorf("MarkAsProcessedWithTTL() error = %v, want nil", err)
	}

	// Should be processed immediately
	isProcessed, err := manager.IsProcessed(ctx, messageID)
	if err != nil {
		t.Errorf("IsProcessed() error = %v, want nil", err)
	}
	if !isProcessed {
		t.Error("IsProcessed() = false, want true immediately after marking with custom TTL")
	}

	// Wait for custom TTL to expire
	time.Sleep(60 * time.Millisecond)

	// Should not be processed after custom TTL expires
	isProcessed, err = manager.IsProcessed(ctx, messageID)
	if err != nil {
		t.Errorf("IsProcessed() error = %v, want nil", err)
	}
	if isProcessed {
		t.Error("IsProcessed() = true, want false after custom TTL expiration")
	}
}

func TestIdempotencyManager_Remove(t *testing.T) {
	manager := NewMemoryIdempotencyManager(1 * time.Hour)
	ctx := context.Background()

	messageID := "test-msg-remove"

	// Mark as processed
	err := manager.MarkAsProcessed(ctx, messageID)
	if err != nil {
		t.Errorf("MarkAsProcessed() error = %v, want nil", err)
	}

	// Should be processed
	isProcessed, err := manager.IsProcessed(ctx, messageID)
	if err != nil {
		t.Errorf("IsProcessed() error = %v, want nil", err)
	}
	if !isProcessed {
		t.Error("IsProcessed() = false, want true after marking")
	}

	// Remove the entry
	err = manager.Remove(ctx, messageID)
	if err != nil {
		t.Errorf("Remove() error = %v, want nil", err)
	}

	// Should not be processed after removal
	isProcessed, err = manager.IsProcessed(ctx, messageID)
	if err != nil {
		t.Errorf("IsProcessed() error = %v, want nil", err)
	}
	if isProcessed {
		t.Error("IsProcessed() = true, want false after removal")
	}
}

func TestIdempotencyManager_Clear(t *testing.T) {
	manager := NewMemoryIdempotencyManager(1 * time.Hour)
	ctx := context.Background()

	messageIDs := []string{"test-msg-1", "test-msg-2", "test-msg-3"}

	// Mark all as processed
	for _, messageID := range messageIDs {
		err := manager.MarkAsProcessed(ctx, messageID)
		if err != nil {
			t.Errorf("MarkAsProcessed() error = %v, want nil", err)
		}
	}

	// Verify all are processed
	for _, messageID := range messageIDs {
		isProcessed, err := manager.IsProcessed(ctx, messageID)
		if err != nil {
			t.Errorf("IsProcessed() error = %v, want nil", err)
		}
		if !isProcessed {
			t.Errorf("IsProcessed() = false, want true for messageID %s", messageID)
		}
	}

	// Clear all entries
	err := manager.Clear(ctx)
	if err != nil {
		t.Errorf("Clear() error = %v, want nil", err)
	}

	// Verify all are not processed after clear
	for _, messageID := range messageIDs {
		isProcessed, err := manager.IsProcessed(ctx, messageID)
		if err != nil {
			t.Errorf("IsProcessed() error = %v, want nil", err)
		}
		if isProcessed {
			t.Errorf("IsProcessed() = true, want false after clear for messageID %s", messageID)
		}
	}
}

func TestIdempotencyManager_GetStats(t *testing.T) {
	manager := NewMemoryIdempotencyManager(1 * time.Hour)
	ctx := context.Background()

	messageID := "test-msg-stats"

	// Initial stats
	stats := manager.GetStats()
	if stats == nil {
		t.Error("GetStats() returned nil")
		return
	}

	initialChecks := stats.TotalChecks

	// Check a message (miss)
	_, err := manager.IsProcessed(ctx, messageID)
	if err != nil {
		t.Errorf("IsProcessed() error = %v, want nil", err)
	}

	// Mark as processed
	err = manager.MarkAsProcessed(ctx, messageID)
	if err != nil {
		t.Errorf("MarkAsProcessed() error = %v, want nil", err)
	}

	// Check the same message again (hit)
	_, err = manager.IsProcessed(ctx, messageID)
	if err != nil {
		t.Errorf("IsProcessed() error = %v, want nil", err)
	}

	// Verify stats
	stats = manager.GetStats()
	if stats.TotalChecks <= initialChecks {
		t.Errorf("Expected TotalChecks to increase, got %d", stats.TotalChecks)
	}

	if stats.Hits == 0 {
		t.Error("Expected at least one hit, got 0")
	}

	if stats.Misses == 0 {
		t.Error("Expected at least one miss, got 0")
	}

	if stats.HitRate < 0 || stats.HitRate > 1 {
		t.Errorf("Expected HitRate between 0 and 1, got %f", stats.HitRate)
	}
}

func TestIdempotencyManager_ConcurrentAccess(t *testing.T) {
	manager := NewMemoryIdempotencyManager(1 * time.Second)
	ctx := context.Background()
	messageID := "test-msg-concurrent"

	// Run multiple goroutines concurrently
	done := make(chan bool, 10)
	processedCount := make(chan int, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Try to mark as processed
			err := manager.MarkAsProcessed(ctx, messageID)
			if err != nil {
				t.Errorf("Concurrent MarkAsProcessed() error = %v", err)
				processedCount <- 0
				return
			}

			// Check if processed
			isProcessed, err := manager.IsProcessed(ctx, messageID)
			if err != nil {
				t.Errorf("Concurrent IsProcessed() error = %v", err)
				processedCount <- 0
				return
			}

			if isProcessed {
				processedCount <- 1
			} else {
				processedCount <- 0
			}
		}()
	}

	// Wait for all goroutines to complete
	totalProcessed := 0
	for i := 0; i < 10; i++ {
		<-done
		totalProcessed += <-processedCount
	}

	// All should see it as processed
	if totalProcessed != 10 {
		t.Errorf("Expected all 10 to see as processed, got %d", totalProcessed)
	}
}
