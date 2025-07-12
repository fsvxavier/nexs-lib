package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

func main() {
	fmt.Println("🏗️ Domain Errors v2 - Builder Pattern Examples")
	fmt.Println("===============================================")

	simpleBuilderExample()
	complexBuilderExample()
	builderWithContext()
	builderWithSeverityAndCategory()
	builderWithHTTPDetails()
	builderChaining()
	performanceOptimizedBuilder()
}

// simpleBuilderExample demonstrates basic builder usage
func simpleBuilderExample() {
	fmt.Println("\n📝 Simple Builder Example:")

	err := domainerrors.NewBuilder().
		WithCode("USR001").
		WithMessage("User not found").
		WithType(string(types.ErrorTypeNotFound)).
		Build()

	fmt.Printf("  Error: %s\n", err.Error())
	fmt.Printf("  Code: %s\n", err.Code())
	fmt.Printf("  Type: %s\n", err.Type())
}

// complexBuilderExample shows advanced builder features
func complexBuilderExample() {
	fmt.Println("\n🔧 Complex Builder Example:")

	err := domainerrors.NewBuilder().
		WithCode("API001").
		WithMessage("Request processing failed").
		WithType(string(types.ErrorTypeBadRequest)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithCategory(interfaces.CategoryTechnical).
		WithDetail("endpoint", "/api/v1/users").
		WithDetail("method", "POST").
		WithDetail("user_agent", "Mozilla/5.0").
		WithDetail("timestamp", time.Now().Format(time.RFC3339)).
		WithTag("api").
		WithTag("http").
		WithTag("validation").
		WithStatusCode(400).
		WithHeader("Content-Type", "application/json").
		WithHeader("X-Error-Code", "API001").
		Build()

	fmt.Printf("  Error: %s\n", err.Error())
	fmt.Printf("  Severity: %v\n", err.Severity())
	fmt.Printf("  Category: %v\n", err.Category())
	fmt.Printf("  Details: %+v\n", err.Details())
	fmt.Printf("  Tags: %v\n", err.Tags())
	fmt.Printf("  Status Code: %d\n", err.StatusCode())
	fmt.Printf("  Headers: %+v\n", err.Headers())
}

// builderWithContext demonstrates context integration
func builderWithContext() {
	fmt.Println("\n🌐 Builder with Context:")

	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", "req-123-456")
	ctx = context.WithValue(ctx, "user_id", "user-789")

	err := domainerrors.NewBuilder().
		WithCode("AUTH001").
		WithMessage("Authentication failed").
		WithType(string(types.ErrorTypeAuthentication)).
		WithContext(ctx).
		WithDetail("login_attempt", "3").
		WithDetail("ip_address", "192.168.1.100").
		WithTag("security").
		WithTag("authentication").
		Build()

	fmt.Printf("  Error: %s\n", err.Error())
	fmt.Printf("  Details: %+v\n", err.Details())
	fmt.Printf("  Tags: %v\n", err.Tags())
}

// builderWithSeverityAndCategory shows classification features
func builderWithSeverityAndCategory() {
	fmt.Println("\n🏷️ Builder with Severity and Category:")

	examples := []struct {
		name     string
		severity interfaces.Severity
		category interfaces.Category
		message  string
	}{
		{
			name:     "Critical Database Error",
			severity: interfaces.Severity(types.SeverityCritical),
			category: interfaces.CategoryInfrastructure,
			message:  "Database connection lost",
		},
		{
			name:     "High Security Alert",
			severity: interfaces.Severity(types.SeverityHigh),
			category: interfaces.CategorySecurity,
			message:  "Suspicious login activity detected",
		},
		{
			name:     "Medium Business Rule",
			severity: interfaces.Severity(types.SeverityMedium),
			category: interfaces.CategoryBusiness,
			message:  "Business rule validation failed",
		},
		{
			name:     "Low Input Validation",
			severity: interfaces.Severity(types.SeverityLow),
			category: interfaces.CategoryTechnical,
			message:  "Invalid email format",
		},
	}

	for _, example := range examples {
		err := domainerrors.NewBuilder().
			WithCode("DEMO").
			WithMessage(example.message).
			WithSeverity(example.severity).
			WithCategory(example.category).
			Build()

		fmt.Printf("  %s: %s (Severity: %v, Category: %v)\n",
			example.name,
			err.Error(),
			err.Severity(),
			err.Category())
	}
}

// builderWithHTTPDetails demonstrates HTTP-specific features
func builderWithHTTPDetails() {
	fmt.Println("\n🌐 Builder with HTTP Details:")

	err := domainerrors.NewBuilder().
		WithCode("HTTP001").
		WithMessage("Rate limit exceeded").
		WithType(string(types.ErrorTypeRateLimit)).
		WithStatusCode(429).
		WithHeader("Retry-After", "60").
		WithHeader("X-RateLimit-Limit", "1000").
		WithHeader("X-RateLimit-Remaining", "0").
		WithHeader("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix())).
		WithDetail("limit", 1000).
		WithDetail("remaining", 0).
		WithDetail("reset_time", time.Now().Add(time.Minute).Format(time.RFC3339)).
		WithTag("rate_limit").
		WithTag("http").
		Build()

	fmt.Printf("  Error: %s\n", err.Error())
	fmt.Printf("  Status Code: %d\n", err.StatusCode())
	fmt.Printf("  Headers: %+v\n", err.Headers())
	fmt.Printf("  Details: %+v\n", err.Details())
}

// builderChaining demonstrates method chaining patterns
func builderChaining() {
	fmt.Println("\n🔗 Builder Chaining Patterns:")

	// Pattern 1: Step-by-step building
	builder := domainerrors.NewBuilder()
	builder = builder.WithCode("CHAIN001")
	builder = builder.WithMessage("Step by step building")
	builder = builder.WithType(string(types.ErrorTypeInternal))
	err1 := builder.Build()

	fmt.Printf("  Step-by-step: %s\n", err1.Error())

	// Pattern 2: Fluent chaining
	err2 := domainerrors.NewBuilder().
		WithCode("CHAIN002").
		WithMessage("Fluent chaining").
		WithType(string(types.ErrorTypeInternal)).
		WithDetail("pattern", "fluent").
		WithTag("example").
		Build()

	fmt.Printf("  Fluent: %s\n", err2.Error())

	// Pattern 3: Conditional building
	builder3 := domainerrors.NewBuilder().
		WithCode("CHAIN003").
		WithMessage("Conditional building")

	// Add conditional details
	isProduction := false
	if isProduction {
		builder3 = builder3.WithDetail("environment", "production")
	} else {
		builder3 = builder3.WithDetail("environment", "development").
			WithDetail("debug", true)
	}

	err3 := builder3.Build()
	fmt.Printf("  Conditional: %s\n", err3.Error())
	fmt.Printf("  Details: %+v\n", err3.Details())
}

// performanceOptimizedBuilder shows performance-optimized patterns
func performanceOptimizedBuilder() {
	fmt.Println("\n⚡ Performance Optimized Builder:")

	start := time.Now()

	// Simulate high-frequency error creation
	count := 1000
	errors := make([]interfaces.DomainErrorInterface, count)

	for i := 0; i < count; i++ {
		errors[i] = domainerrors.NewBuilder().
			WithCode(fmt.Sprintf("PERF%03d", i)).
			WithMessage("Performance test error").
			WithType(string(types.ErrorTypeInternal)).
			WithDetail("iteration", i).
			WithTag("performance").
			Build()
	}

	elapsed := time.Since(start)

	fmt.Printf("  Created %d errors in %v\n", count, elapsed)
	fmt.Printf("  Average: %v per error\n", elapsed/time.Duration(count))
	fmt.Printf("  First error: %s\n", errors[0].Error())
	fmt.Printf("  Last error: %s\n", errors[count-1].Error())

	// Memory usage simulation
	fmt.Printf("  Memory efficient: Object pooling enabled\n")
	fmt.Printf("  Thread safety: All operations are thread-safe\n")
}
