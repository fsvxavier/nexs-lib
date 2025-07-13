// Package main demonstrates enhanced error handling with circuit breakers and retries
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
)

func main() {
	fmt.Println("=== Enhanced Error Handling Example ===")

	// Setup error handling components
	retryConfig := tracer.DefaultRetryConfig()
	circuitBreakerConfig := tracer.DefaultCircuitBreakerConfig()
	errorHandler := tracer.NewDefaultErrorHandler(retryConfig, circuitBreakerConfig)

	ctx := context.Background()

	// Example 1: Simulate a network error
	fmt.Println("\n--- Network Error Example ---")
	networkOperation := func() error {
		return errors.New("connection timeout")
	}

	err := tracer.RetryWithBackoff(ctx, networkOperation, errorHandler, "network_op")
	if err != nil {
		classification := errorHandler.ClassifyError(err)
		fmt.Printf("Network operation failed: %v (Classification: %s)\n", err, classification)
	}

	// Example 2: Simulate an auth error (non-retryable)
	fmt.Println("\n--- Auth Error Example ---")
	authOperation := func() error {
		return errors.New("unauthorized")
	}

	err = tracer.RetryWithBackoff(ctx, authOperation, errorHandler, "auth_op")
	if err != nil {
		classification := errorHandler.ClassifyError(err)
		fmt.Printf("Auth operation failed: %v (Classification: %s)\n", err, classification)
	}

	// Example 3: Circuit breaker demonstration
	fmt.Println("\n--- Circuit Breaker Example ---")
	cb := tracer.NewDefaultCircuitBreaker(circuitBreakerConfig)

	for i := 0; i < 10; i++ {
		operation := func() error {
			if i < 7 {
				return errors.New("service failure")
			}
			return nil
		}

		err := cb.Execute(ctx, operation)
		metrics := cb.GetMetrics()
		fmt.Printf("Attempt %d: err=%v, failures=%d, successes=%d\n",
			i+1, err, metrics.FailureCount, metrics.SuccessCount)
	}

	fmt.Println("\n🎉 Error handling examples completed!")
}
