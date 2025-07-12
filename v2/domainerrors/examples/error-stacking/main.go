package main

import (
	"errors"
	"fmt"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

func main() {
	fmt.Println("🔗 Domain Errors v2 - Error Stacking Examples")
	fmt.Println("============================================")

	basicWrapping()
	errorChaining()
	rootCauseAnalysis()
	stackTraceExample()
	complexErrorHierarchy()
}

// basicWrapping demonstrates simple error wrapping
func basicWrapping() {
	fmt.Println("\n📦 Basic Error Wrapping:")

	// Original error
	originalErr := errors.New("database connection timeout")

	// Wrap with domain error
	wrappedErr := domainerrors.NewBuilder().
		WithCode("DB001").
		WithMessage("Database operation failed").
		WithType(string(types.ErrorTypeDatabase)).
		WithCause(originalErr).
		Build()

	fmt.Printf("  Original: %s\n", originalErr.Error())
	fmt.Printf("  Wrapped: %s\n", wrappedErr.Error())
	fmt.Printf("  Unwrapped: %s\n", wrappedErr.Unwrap().Error())
}

// errorChaining demonstrates error chaining
func errorChaining() {
	fmt.Println("\n⛓️  Error Chaining:")

	// Base error
	dbErr := domainerrors.NewBuilder().
		WithCode("DB001").
		WithMessage("Connection failed").
		WithType(string(types.ErrorTypeDatabase)).
		Build()

	// Chain another error
	serviceErr := domainerrors.NewBuilder().
		WithCode("SVC001").
		WithMessage("Service unavailable").
		WithType(string(types.ErrorTypeExternalService)).
		Build()

	chainedErr := dbErr.Chain(serviceErr)

	fmt.Printf("  DB Error: %s\n", dbErr.Error())
	fmt.Printf("  Service Error: %s\n", serviceErr.Error())
	fmt.Printf("  Chained: %s\n", chainedErr.Error())
}

// rootCauseAnalysis demonstrates finding root causes
func rootCauseAnalysis() {
	fmt.Println("\n🔍 Root Cause Analysis:")

	// Create a chain of errors
	level1 := errors.New("network timeout")

	level2 := domainerrors.NewBuilder().
		WithCode("NET001").
		WithMessage("Network error").
		WithType(string(types.ErrorTypeNetwork)).
		WithCause(level1).
		Build()

	level3 := domainerrors.NewBuilder().
		WithCode("API001").
		WithMessage("API call failed").
		WithType(string(types.ErrorTypeExternalService)).
		WithCause(level2).
		Build()

	level4 := domainerrors.NewBuilder().
		WithCode("USR001").
		WithMessage("User operation failed").
		WithType(string(types.ErrorTypeBusinessRule)).
		WithCause(level3).
		Build()

	fmt.Printf("  Current Error: %s\n", level4.Error())
	fmt.Printf("  Root Cause: %s\n", level4.RootCause().Error())

	// Demonstrate error chain walking
	fmt.Printf("  Error Chain:\n")
	current := level4
	depth := 0
	for current != nil {
		fmt.Printf("    Level %d: %s\n", depth, current.Error())
		if unwrapped := current.Unwrap(); unwrapped != nil {
			if domainErr, ok := unwrapped.(interfaces.DomainErrorInterface); ok {
				current = domainErr
			} else {
				fmt.Printf("    Level %d: %s (stdlib error)\n", depth+1, unwrapped.Error())
				break
			}
		} else {
			break
		}
		depth++
	}
}

// stackTraceExample demonstrates stack trace functionality
func stackTraceExample() {
	fmt.Println("\n📊 Stack Trace Example:")

	err := createNestedError()

	fmt.Printf("  Error: %s\n", err.Error())
	fmt.Printf("  Stack Trace:\n%s\n", err.FormatStackTrace())
}

// createNestedError simulates a nested function call that creates an error
func createNestedError() interfaces.DomainErrorInterface {
	return simulateServiceCall()
}

func simulateServiceCall() interfaces.DomainErrorInterface {
	return simulateDatabaseCall()
}

func simulateDatabaseCall() interfaces.DomainErrorInterface {
	originalErr := errors.New("connection refused")

	return domainerrors.NewBuilder().
		WithCode("DB002").
		WithMessage("Database query failed").
		WithType(string(types.ErrorTypeDatabase)).
		WithCause(originalErr).
		WithDetail("table", "users").
		WithDetail("operation", "SELECT").
		Build()
}

// complexErrorHierarchy demonstrates complex error hierarchies
func complexErrorHierarchy() {
	fmt.Println("\n🏗️ Complex Error Hierarchy:")

	// Simulate a complex business operation with multiple failure points
	authErr := domainerrors.NewBuilder().
		WithCode("AUTH001").
		WithMessage("Invalid credentials").
		WithType(string(types.ErrorTypeAuthentication)).
		Build()

	validationErr := domainerrors.NewBuilder().
		WithCode("VAL001").
		WithMessage("Data validation failed").
		WithType(string(types.ErrorTypeValidation)).
		WithDetail("field", "email").
		Build()

	businessErr := domainerrors.NewBuilder().
		WithCode("BIZ001").
		WithMessage("Business rule violation").
		WithType(string(types.ErrorTypeBusinessRule)).
		Build()

	// Create a complex hierarchy
	combinedErr := authErr.
		Chain(validationErr).
		Chain(businessErr)

	// Wrap everything in a higher-level error
	finalErr := domainerrors.NewBuilder().
		WithCode("OP001").
		WithMessage("User registration failed").
		WithType(string(types.ErrorTypeBusinessRule)).
		WithCause(combinedErr).
		WithDetail("operation", "user_registration").
		WithDetail("user_id", "user-123").
		Build()

	fmt.Printf("  Final Error: %s\n", finalErr.Error())
	fmt.Printf("  Root Cause: %s\n", finalErr.RootCause().Error())
	fmt.Printf("  Error Details: %+v\n", finalErr.Details())

	// Demonstrate error type checking
	fmt.Printf("\n  Error Type Analysis:\n")
	if errors.Is(finalErr, authErr) {
		fmt.Printf("    ✓ Contains authentication error\n")
	}
	if errors.As(finalErr, &validationErr) {
		fmt.Printf("    ✓ Contains validation error\n")
	}
}
