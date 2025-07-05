package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

func BenchmarkRetryer_SuccessFirstAttempt(b *testing.B) {
	policy := ExponentialBackoffPolicy(3, 100*time.Millisecond, 2.0)
	retryer := NewRetryer(policy)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		retryer.Execute(ctx, func() error {
			return nil // Sucesso imediato
		})
	}
}

func BenchmarkRetryer_SuccessAfterRetries(b *testing.B) {
	policy := ExponentialBackoffPolicy(3, 1*time.Millisecond, 2.0) // Delay mínimo para benchmark
	retryer := NewRetryer(policy)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		attempts := 0
		retryer.Execute(ctx, func() error {
			attempts++
			if attempts < 3 {
				return domainerrors.New("NETWORK_ERROR", "temporary network issue").
					WithType(domainerrors.ErrorTypeInfrastructure)
			}
			return nil
		})
	}
}

func BenchmarkRetryer_MaxRetriesReached(b *testing.B) {
	policy := ExponentialBackoffPolicy(2, 1*time.Millisecond, 2.0) // Delay mínimo
	retryer := NewRetryer(policy)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		retryer.Execute(ctx, func() error {
			return domainerrors.New("TIMEOUT_ERROR", "persistent timeout").
				WithType(domainerrors.ErrorTypeInfrastructure)
		})
	}
}

func BenchmarkRetryer_NonRetryableError(b *testing.B) {
	policy := ExponentialBackoffPolicy(3, 1*time.Millisecond, 2.0)
	retryer := NewRetryer(policy)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		retryer.Execute(ctx, func() error {
			return domainerrors.New("VALIDATION_ERROR", "invalid input").
				WithType(domainerrors.ErrorTypeValidation)
		})
	}
}

func BenchmarkRetryer_WithCallback(b *testing.B) {
	policy := ExponentialBackoffPolicy(3, 1*time.Millisecond, 2.0)
	retryer := NewRetryer(policy)
	ctx := context.Background()

	callbackCount := 0
	callback := func(attempt int, err error) {
		callbackCount++
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		attempts := 0
		retryer.ExecuteWithCallback(ctx, func() error {
			attempts++
			if attempts < 2 {
				return errors.New("temporary error")
			}
			return nil
		}, callback)
	}
}

func BenchmarkPolicyCreation(b *testing.B) {
	b.Run("DefaultPolicy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			DefaultRetryPolicy()
		}
	})

	b.Run("ExponentialPolicy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ExponentialBackoffPolicy(5, 100*time.Millisecond, 2.0)
		}
	})

	b.Run("LinearPolicy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			LinearBackoffPolicy(5, 100*time.Millisecond)
		}
	})

	b.Run("NoRetryPolicy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NoRetryPolicy()
		}
	})
}

func BenchmarkIsRetryableError(b *testing.B) {
	errors := []error{
		nil,
		domainerrors.New("NETWORK_ERROR", "network issue").WithType(domainerrors.ErrorTypeInfrastructure),
		domainerrors.New("TIMEOUT_ERROR", "timeout").WithType(domainerrors.ErrorTypeInfrastructure),
		domainerrors.New("VALIDATION_ERROR", "validation").WithType(domainerrors.ErrorTypeValidation),
		domainerrors.New("BUSINESS_ERROR", "business").WithType(domainerrors.ErrorTypeRepository),
		errors.New("generic error"),
		errors.New("connection refused"),
		errors.New("timeout occurred"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := errors[i%len(errors)]
		IsRetryableError(err)
	}
}
