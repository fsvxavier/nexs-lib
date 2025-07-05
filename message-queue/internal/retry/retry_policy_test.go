package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

func TestNewRetryer(t *testing.T) {
	tests := []struct {
		name   string
		policy *interfaces.RetryPolicy
		want   bool // true if should succeed
	}{
		{
			name:   "nil policy uses default",
			policy: nil,
			want:   true,
		},
		{
			name: "valid policy",
			policy: &interfaces.RetryPolicy{
				MaxAttempts:       3,
				InitialDelay:      100 * time.Millisecond,
				BackoffMultiplier: 2.0,
				MaxDelay:          5 * time.Second,
				Jitter:            true,
			},
			want: true,
		},
		{
			name: "zero max attempts",
			policy: &interfaces.RetryPolicy{
				MaxAttempts:       0,
				InitialDelay:      100 * time.Millisecond,
				BackoffMultiplier: 2.0,
				MaxDelay:          5 * time.Second,
				Jitter:            false,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retryer := NewRetryer(tt.policy)

			if tt.want {
				if retryer == nil {
					t.Error("NewRetryer() returned nil for valid inputs")
					return
				}

				policy := retryer.GetPolicy()
				if policy == nil {
					t.Error("NewRetryer() returned retryer with nil policy")
					return
				}
			} else {
				if retryer != nil {
					t.Error("NewRetryer() returned non-nil retryer for invalid inputs")
				}
			}
		})
	}
}

func TestDefaultRetryPolicy(t *testing.T) {
	policy := DefaultRetryPolicy()

	if policy == nil {
		t.Error("DefaultRetryPolicy() returned nil")
		return
	}

	if policy.MaxAttempts <= 0 {
		t.Errorf("DefaultRetryPolicy() returned policy with invalid MaxAttempts: %d", policy.MaxAttempts)
	}

	if policy.InitialDelay <= 0 {
		t.Errorf("DefaultRetryPolicy() returned policy with invalid InitialDelay: %v", policy.InitialDelay)
	}

	if policy.MaxDelay <= policy.InitialDelay {
		t.Errorf("DefaultRetryPolicy() returned policy with MaxDelay (%v) <= InitialDelay (%v)", policy.MaxDelay, policy.InitialDelay)
	}

	if policy.BackoffMultiplier <= 1.0 {
		t.Errorf("DefaultRetryPolicy() returned policy with invalid BackoffMultiplier: %f", policy.BackoffMultiplier)
	}
}

func TestExponentialBackoffPolicy(t *testing.T) {
	maxAttempts := 5
	initialDelay := 100 * time.Millisecond
	maxDelay := 2 * time.Second

	policy := ExponentialBackoffPolicy(maxAttempts, initialDelay, maxDelay)

	if policy == nil {
		t.Error("ExponentialBackoffPolicy() returned nil")
		return
	}

	if policy.MaxAttempts != maxAttempts {
		t.Errorf("ExponentialBackoffPolicy() MaxAttempts = %d, want %d", policy.MaxAttempts, maxAttempts)
	}

	if policy.InitialDelay != initialDelay {
		t.Errorf("ExponentialBackoffPolicy() InitialDelay = %v, want %v", policy.InitialDelay, initialDelay)
	}

	if policy.MaxDelay != maxDelay {
		t.Errorf("ExponentialBackoffPolicy() MaxDelay = %v, want %v", policy.MaxDelay, maxDelay)
	}

	if policy.BackoffMultiplier != 2.0 {
		t.Errorf("ExponentialBackoffPolicy() BackoffMultiplier = %f, want 2.0", policy.BackoffMultiplier)
	}

	if !policy.Jitter {
		t.Error("ExponentialBackoffPolicy() Jitter = false, want true")
	}
}

func TestLinearBackoffPolicy(t *testing.T) {
	maxAttempts := 3
	delay := 500 * time.Millisecond

	policy := LinearBackoffPolicy(maxAttempts, delay)

	if policy == nil {
		t.Error("LinearBackoffPolicy() returned nil")
		return
	}

	if policy.MaxAttempts != maxAttempts {
		t.Errorf("LinearBackoffPolicy() MaxAttempts = %d, want %d", policy.MaxAttempts, maxAttempts)
	}

	if policy.InitialDelay != delay {
		t.Errorf("LinearBackoffPolicy() InitialDelay = %v, want %v", policy.InitialDelay, delay)
	}

	if policy.BackoffMultiplier != 1.0 {
		t.Errorf("LinearBackoffPolicy() BackoffMultiplier = %f, want 1.0", policy.BackoffMultiplier)
	}

	if policy.Jitter {
		t.Error("LinearBackoffPolicy() Jitter = true, want false")
	}
}

func TestNoRetryPolicy(t *testing.T) {
	policy := NoRetryPolicy()

	if policy == nil {
		t.Error("NoRetryPolicy() returned nil")
		return
	}

	if policy.MaxAttempts != 1 {
		t.Errorf("NoRetryPolicy() MaxAttempts = %d, want 1", policy.MaxAttempts)
	}

	if policy.InitialDelay != 0 {
		t.Errorf("NoRetryPolicy() InitialDelay = %v, want 0", policy.InitialDelay)
	}

	if policy.BackoffMultiplier != 1.0 {
		t.Errorf("NoRetryPolicy() BackoffMultiplier = %f, want 1.0", policy.BackoffMultiplier)
	}

	if policy.Jitter {
		t.Error("NoRetryPolicy() Jitter = true, want false")
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("generic error"),
			want: false, // Generic errors are not retryable by default
		},
		{
			name: "connection error",
			err:  errors.New("connection failed"),
			want: true,
		},
		{
			name: "timeout error keyword",
			err:  errors.New("operation timeout occurred"),
			want: true,
		},
		{
			name: "network error",
			err:  errors.New("network unavailable"),
			want: true,
		},
		{
			name: "validation error",
			err:  domainerrors.NewValidationError("invalid input", map[string][]string{"field": {"error"}}),
			want: false,
		},
		{
			name: "business error",
			err:  domainerrors.NewBusinessError("business rule violation", "RULE"),
			want: false,
		},
		{
			name: "infrastructure error",
			err:  domainerrors.NewInfrastructureError("database connection failed", "DB", errors.New("connection refused")),
			want: true,
		},
		{
			name: "timeout domain error",
			err:  domainerrors.NewTimeoutError("operation timed out", "OP"),
			want: true, // Domain timeout errors should be retryable based on ErrorType
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetryableError(tt.err)
			if got != tt.want {
				t.Errorf("IsRetryableError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestRetryer_Execute(t *testing.T) {
	policy := &interfaces.RetryPolicy{
		MaxAttempts:       3,
		InitialDelay:      10 * time.Millisecond,
		BackoffMultiplier: 2.0,
		MaxDelay:          100 * time.Millisecond,
		Jitter:            false,
	}
	retryer := NewRetryer(policy)
	ctx := context.Background()

	t.Run("success on first attempt", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			return nil
		}

		err := retryer.Execute(ctx, operation)
		if err != nil {
			t.Errorf("Execute() error = %v, want nil", err)
		}

		if callCount != 1 {
			t.Errorf("Expected operation to be called 1 time, got %d", callCount)
		}
	})

	t.Run("success after retries", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			if callCount < 2 {
				return errors.New("temporary failure")
			}
			return nil
		}

		err := retryer.Execute(ctx, operation)
		if err != nil {
			t.Errorf("Execute() error = %v, want nil", err)
		}

		if callCount != 2 {
			t.Errorf("Expected operation to be called 2 times, got %d", callCount)
		}
	})

	t.Run("failure after max retries", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			return errors.New("persistent failure")
		}

		err := retryer.Execute(ctx, operation)
		if err == nil {
			t.Error("Execute() error = nil, want error")
		}

		expectedCalls := policy.MaxAttempts
		if callCount != expectedCalls {
			t.Errorf("Expected operation to be called %d times, got %d", expectedCalls, callCount)
		}
	})

	t.Run("non-retryable error still retries", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			return domainerrors.NewValidationError("invalid input", map[string][]string{"field": {"error"}})
		}

		err := retryer.Execute(ctx, operation)
		if err == nil {
			t.Error("Execute() error = nil, want error")
		}

		// Even non-retryable errors go through all attempts in this implementation
		if callCount != policy.MaxAttempts {
			t.Errorf("Expected operation to be called %d times, got %d", policy.MaxAttempts, callCount)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		callCount := 0
		operation := func() error {
			callCount++
			return errors.New("some error")
		}

		err := retryer.Execute(cancelCtx, operation)
		if err == nil {
			t.Error("Execute() error = nil, want context cancelled error")
		}

		// Should not retry when context is cancelled
		if callCount > 1 {
			t.Errorf("Expected operation to be called at most 1 time with cancelled context, got %d", callCount)
		}
	})
}

func TestRetryer_ExecuteWithCallback(t *testing.T) {
	policy := &interfaces.RetryPolicy{
		MaxAttempts:       2,
		InitialDelay:      10 * time.Millisecond,
		BackoffMultiplier: 2.0,
		MaxDelay:          100 * time.Millisecond,
		Jitter:            false,
	}
	retryer := NewRetryer(policy)
	ctx := context.Background()

	t.Run("callback called for each attempt", func(t *testing.T) {
		callCount := 0
		callbackCalls := 0
		var callbackAttempts []int
		var callbackErrors []error

		operation := func() error {
			callCount++
			return errors.New("persistent failure")
		}

		callback := func(attempt int, err error) {
			callbackCalls++
			callbackAttempts = append(callbackAttempts, attempt)
			callbackErrors = append(callbackErrors, err)
		}

		err := retryer.ExecuteWithCallback(ctx, operation, callback)
		if err == nil {
			t.Error("ExecuteWithCallback() error = nil, want error")
		}

		if callCount != policy.MaxAttempts {
			t.Errorf("Expected operation to be called %d times, got %d", policy.MaxAttempts, callCount)
		}

		if callbackCalls != policy.MaxAttempts {
			t.Errorf("Expected callback to be called %d times, got %d", policy.MaxAttempts, callbackCalls)
		}

		// Verify callback was called with correct attempt numbers (starting from 1)
		for i, attempt := range callbackAttempts {
			expectedAttempt := i + 1
			if attempt != expectedAttempt {
				t.Errorf("Expected callback attempt %d, got %d", expectedAttempt, attempt)
			}
		}

		// Verify all callback errors are non-nil
		for i, err := range callbackErrors {
			if err == nil {
				t.Errorf("Expected callback error %d to be non-nil", i)
			}
		}
	})

	t.Run("nil callback works", func(t *testing.T) {
		callCount := 0
		operation := func() error {
			callCount++
			return nil
		}

		err := retryer.ExecuteWithCallback(ctx, operation, nil)
		if err != nil {
			t.Errorf("ExecuteWithCallback() with nil callback error = %v, want nil", err)
		}

		if callCount != 1 {
			t.Errorf("Expected operation to be called 1 time, got %d", callCount)
		}
	})
}

func TestRetryer_GetSetPolicy(t *testing.T) {
	initialPolicy := DefaultRetryPolicy()
	retryer := NewRetryer(initialPolicy)

	// Test GetPolicy
	policy := retryer.GetPolicy()
	if policy != initialPolicy {
		t.Error("GetPolicy() returned different policy than the one provided")
	}

	// Test SetPolicy
	newPolicy := &interfaces.RetryPolicy{
		MaxAttempts:       5,
		InitialDelay:      200 * time.Millisecond,
		BackoffMultiplier: 3.0,
		MaxDelay:          10 * time.Second,
		Jitter:            true,
	}

	retryer.SetPolicy(newPolicy)

	updatedPolicy := retryer.GetPolicy()
	if updatedPolicy != newPolicy {
		t.Error("SetPolicy() did not update the policy correctly")
	}

	if updatedPolicy.MaxAttempts != newPolicy.MaxAttempts {
		t.Errorf("SetPolicy() MaxAttempts = %d, want %d", updatedPolicy.MaxAttempts, newPolicy.MaxAttempts)
	}

	if updatedPolicy.InitialDelay != newPolicy.InitialDelay {
		t.Errorf("SetPolicy() InitialDelay = %v, want %v", updatedPolicy.InitialDelay, newPolicy.InitialDelay)
	}
}
