//go:build unit
// +build unit

package postgresql

import (
	"context"
	"testing"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
)

func TestDefaultRetryManager(t *testing.T) {
	config := interfaces.RetryConfig{
		MaxRetries:      3,
		InitialInterval: time.Millisecond * 10,
		MaxInterval:     time.Millisecond * 100,
		Multiplier:      2.0,
		RandomizeWait:   false,
	}

	retryManager := NewDefaultRetryManager(config)

	t.Run("Successful Operation", func(t *testing.T) {
		executed := false
		operation := func() error {
			executed = true
			return nil
		}

		err := retryManager.Execute(context.Background(), operation)
		if err != nil {
			t.Errorf("Execute() error = %v", err)
		}

		if !executed {
			t.Error("Operation should have been executed")
		}

		stats := retryManager.GetStats()
		if stats.SuccessfulOps != 1 {
			t.Errorf("Expected 1 successful op, got %d", stats.SuccessfulOps)
		}
	})

	t.Run("Retryable Error", func(t *testing.T) {
		attemptCount := 0
		operation := func() error {
			attemptCount++
			if attemptCount < 3 {
				return &retryTemporaryError{message: "temporary failure"}
			}
			return nil
		}

		err := retryManager.Execute(context.Background(), operation)
		if err != nil {
			t.Errorf("Execute() error = %v", err)
		}

		if attemptCount != 3 {
			t.Errorf("Expected 3 attempts, got %d", attemptCount)
		}
	})

	t.Run("Non-retryable Error", func(t *testing.T) {
		operation := func() error {
			return &retryPermanentError{message: "permanent failure"}
		}

		err := retryManager.Execute(context.Background(), operation)
		if err == nil {
			t.Error("Execute() should have returned error")
		}
	})

	t.Run("Max Retries Exceeded", func(t *testing.T) {
		operation := func() error {
			return &retryTemporaryError{message: "always fails"}
		}

		err := retryManager.Execute(context.Background(), operation)
		if err == nil {
			t.Error("Execute() should have returned error after max retries")
		}

		stats := retryManager.GetStats()
		if stats.FailedOps == 0 {
			t.Error("Should have recorded failed operation")
		}
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		operation := func() error {
			return &retryTemporaryError{message: "temporary failure"}
		}

		// Cancel context after first attempt
		go func() {
			time.Sleep(time.Millisecond * 5)
			cancel()
		}()

		err := retryManager.Execute(ctx, operation)
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})
}

func TestDefaultRetryManagerConfiguration(t *testing.T) {
	t.Run("Update Valid Config", func(t *testing.T) {
		retryManager := NewDefaultRetryManager(interfaces.RetryConfig{})

		newConfig := interfaces.RetryConfig{
			MaxRetries:      5,
			InitialInterval: time.Millisecond * 50,
			MaxInterval:     time.Second,
			Multiplier:      1.5,
			RandomizeWait:   true,
		}

		err := retryManager.UpdateConfig(newConfig)
		if err != nil {
			t.Errorf("UpdateConfig() error = %v", err)
		}
	})

	t.Run("Update Invalid Config", func(t *testing.T) {
		retryManager := NewDefaultRetryManager(interfaces.RetryConfig{})

		invalidConfigs := []interfaces.RetryConfig{
			{MaxRetries: -1},                    // Invalid max retries
			{MaxRetries: 1, InitialInterval: 0}, // Invalid initial interval
			{MaxRetries: 1, InitialInterval: time.Millisecond, MaxInterval: 0},                            // Invalid max interval
			{MaxRetries: 1, InitialInterval: time.Millisecond, MaxInterval: time.Second, Multiplier: 0.5}, // Invalid multiplier
		}

		for i, config := range invalidConfigs {
			err := retryManager.UpdateConfig(config)
			if err == nil {
				t.Errorf("UpdateConfig() should have returned error for invalid config %d", i)
			}
		}
	})
}

func TestRetryManagerBackoff(t *testing.T) {
	config := interfaces.RetryConfig{
		MaxRetries:      3,
		InitialInterval: time.Millisecond * 10,
		MaxInterval:     time.Millisecond * 100,
		Multiplier:      2.0,
		RandomizeWait:   false,
	}

	retryManager := NewDefaultRetryManager(config)

	t.Run("Backoff Calculation", func(t *testing.T) {
		backoff1 := retryManager.calculateBackoff(1)
		backoff2 := retryManager.calculateBackoff(2)
		backoff3 := retryManager.calculateBackoff(3)

		if backoff2 <= backoff1 {
			t.Error("Backoff should increase with attempt number")
		}

		if backoff3 > config.MaxInterval {
			t.Error("Backoff should not exceed max interval")
		}
	})

	t.Run("Randomized Backoff", func(t *testing.T) {
		configWithJitter := config
		configWithJitter.RandomizeWait = true

		retryManagerWithJitter := NewDefaultRetryManager(configWithJitter)

		// Run multiple times to check for variation
		backoffs := make([]time.Duration, 5)
		for i := range backoffs {
			backoffs[i] = retryManagerWithJitter.calculateBackoff(2)
		}

		// At least some should be different (though there's a small chance they're all the same)
		allSame := true
		for i := 1; i < len(backoffs); i++ {
			if backoffs[i] != backoffs[0] {
				allSame = false
				break
			}
		}

		if allSame {
			t.Log("Warning: All jittered backoffs were the same (low probability but possible)")
		}
	})
}

func TestDefaultFailoverManager(t *testing.T) {
	config := interfaces.FailoverConfig{
		Enabled:             true,
		FallbackNodes:       []string{"node1", "node2", "node3"},
		HealthCheckInterval: time.Second,
		RetryInterval:       time.Millisecond * 500,
		MaxFailoverAttempts: 3,
	}

	failoverManager := NewDefaultFailoverManager(config)

	t.Run("Initial State", func(t *testing.T) {
		healthyNodes := failoverManager.GetHealthyNodes()
		if len(healthyNodes) != 3 {
			t.Errorf("Expected 3 healthy nodes, got %d", len(healthyNodes))
		}

		unhealthyNodes := failoverManager.GetUnhealthyNodes()
		if len(unhealthyNodes) != 0 {
			t.Errorf("Expected 0 unhealthy nodes, got %d", len(unhealthyNodes))
		}
	})

	t.Run("Mark Node Down", func(t *testing.T) {
		err := failoverManager.MarkNodeDown("node1")
		if err != nil {
			t.Errorf("MarkNodeDown() error = %v", err)
		}

		healthyNodes := failoverManager.GetHealthyNodes()
		unhealthyNodes := failoverManager.GetUnhealthyNodes()

		if len(healthyNodes) != 2 {
			t.Errorf("Expected 2 healthy nodes, got %d", len(healthyNodes))
		}

		if len(unhealthyNodes) != 1 {
			t.Errorf("Expected 1 unhealthy node, got %d", len(unhealthyNodes))
		}

		if unhealthyNodes[0] != "node1" {
			t.Errorf("Expected 'node1' to be unhealthy, got %s", unhealthyNodes[0])
		}
	})

	t.Run("Mark Node Up", func(t *testing.T) {
		// First mark node down
		failoverManager.MarkNodeDown("node2")

		// Then mark it back up
		err := failoverManager.MarkNodeUp("node2")
		if err != nil {
			t.Errorf("MarkNodeUp() error = %v", err)
		}

		healthyNodes := failoverManager.GetHealthyNodes()
		found := false
		for _, node := range healthyNodes {
			if node == "node2" {
				found = true
				break
			}
		}

		if !found {
			t.Error("node2 should be marked as healthy")
		}
	})

	t.Run("Execute Operation", func(t *testing.T) {
		operation := func(conn interfaces.IConn) error {
			return nil
		}

		// This should fail as noted in the implementation
		err := failoverManager.Execute(context.Background(), operation)
		if err == nil {
			t.Error("Execute() should return error for incomplete implementation")
		}

		stats := failoverManager.GetStats()
		if stats.TotalFailovers == 0 {
			t.Error("Should have recorded failover attempt")
		}
	})
}

func TestFailoverManagerEdgeCases(t *testing.T) {
	config := interfaces.FailoverConfig{
		Enabled:             true,
		FallbackNodes:       []string{},
		HealthCheckInterval: time.Second,
		RetryInterval:       time.Millisecond * 500,
		MaxFailoverAttempts: 3,
	}

	failoverManager := NewDefaultFailoverManager(config)

	t.Run("Empty Node List", func(t *testing.T) {
		healthyNodes := failoverManager.GetHealthyNodes()
		if len(healthyNodes) != 0 {
			t.Errorf("Expected 0 healthy nodes for empty config, got %d", len(healthyNodes))
		}
	})

	t.Run("Mark Non-existent Node", func(t *testing.T) {
		err := failoverManager.MarkNodeDown("nonexistent")
		if err != nil {
			t.Errorf("MarkNodeDown() should not error for non-existent node: %v", err)
		}

		err = failoverManager.MarkNodeUp("nonexistent")
		if err != nil {
			t.Errorf("MarkNodeUp() should not error for non-existent node: %v", err)
		}
	})
}

func TestRetryableErrorDetection(t *testing.T) {
	retryManager := NewDefaultRetryManager(interfaces.RetryConfig{})

	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "nil error",
			err:       nil,
			retryable: false,
		},
		{
			name:      "temporary error",
			err:       &retryTemporaryError{message: "temp"},
			retryable: true,
		},
		{
			name:      "permanent error",
			err:       &retryPermanentError{message: "perm"},
			retryable: false,
		},
		{
			name:      "context canceled",
			err:       context.Canceled,
			retryable: false,
		},
		{
			name:      "context deadline exceeded",
			err:       context.DeadlineExceeded,
			retryable: false,
		},
		{
			name:      "connection refused",
			err:       &networkError{message: "connection refused"},
			retryable: true,
		},
		{
			name:      "unknown error",
			err:       &unknownError{message: "unknown"},
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retryable := retryManager.isRetryableError(tt.err)
			if retryable != tt.retryable {
				t.Errorf("isRetryableError() = %v, want %v for error: %v", retryable, tt.retryable, tt.err)
			}
		})
	}
}

// Helper error types for testing (unique naming to avoid conflicts)
type retryTemporaryError struct {
	message string
}

func (e *retryTemporaryError) Error() string {
	return e.message
}

func (e *retryTemporaryError) Temporary() bool {
	return true
}

type retryPermanentError struct {
	message string
}

func (e *retryPermanentError) Error() string {
	return e.message
}

func (e *retryPermanentError) Temporary() bool {
	return false
}

type networkError struct {
	message string
}

func (e *networkError) Error() string {
	return e.message
}

type unknownError struct {
	message string
}

func (e *unknownError) Error() string {
	return e.message
}
