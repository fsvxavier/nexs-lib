package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

func main() {
	fmt.Println("=== Domain Errors - Specific Error Types Examples ===")
	fmt.Println()

	// Example 1: Validation Errors
	validationErrorExample()

	// Example 2: Business Errors
	businessErrorExample()

	// Example 3: Database Errors
	databaseErrorExample()

	// Example 4: External Service Errors
	externalServiceErrorExample()

	// Example 5: Authentication/Authorization Errors
	authErrorExample()

	// Example 6: Timeout and Circuit Breaker Errors
	timeoutCircuitBreakerExample()

	// Example 7: Rate Limiting Errors
	rateLimitingExample()

	// Example 8: Infrastructure Errors
	infrastructureErrorExample()

	// Example 9: Workflow Errors
	workflowErrorExample()

	// Example 10: Cache and Configuration Errors
	cacheConfigurationExample()
}

func validationErrorExample() {
	fmt.Println("1. Validation Errors:")

	// Basic validation error
	err := domainerrors.NewValidationError("USER_VALIDATION_FAILED", "User data validation failed", nil)
	err.WithField("email", "invalid email format")
	err.WithField("email", "email already exists")
	err.WithField("age", "must be 18 or older")
	err.WithField("phone", "invalid format")

	fmt.Printf("   Error: %s\n", err.Error())
	fmt.Printf("   Code: %s\n", err.Code)
	fmt.Printf("   HTTP Status: %d\n", err.HTTPStatus())
	fmt.Printf("   Validation Fields:\n")
	for field, messages := range err.Fields {
		fmt.Printf("     %s: %v\n", field, messages)
	}

	// Validation error with multiple fields
	fmt.Println("\n   Complex Validation:")
	complexErr := domainerrors.NewValidationError("PRODUCT_VALIDATION_FAILED", "Product validation failed", nil)
	complexErr.WithField("name", "required")
	complexErr.WithField("price", "must be positive")
	complexErr.WithField("price", "must be less than 10000")
	complexErr.WithField("category", "invalid category")
	complexErr.WithField("tags", "minimum 2 tags required")

	fmt.Printf("   Product Validation Error: %s\n", complexErr.Error())
	fmt.Printf("   Total Fields with Errors: %d\n", len(complexErr.Fields))

	fmt.Println()
}

func businessErrorExample() {
	fmt.Println("2. Business Errors:")

	// Business rule violation
	err := domainerrors.NewBusinessError("INSUFFICIENT_BALANCE", "Account balance insufficient")
	err.WithRule("Minimum balance of $10 required")
	err.WithRule("Account must be active")
	err.WithRule("Daily withdrawal limit not exceeded")

	fmt.Printf("   Error: %s\n", err.Error())
	fmt.Printf("   Business Code: %s\n", err.BusinessCode)
	fmt.Printf("   HTTP Status: %d\n", err.HTTPStatus())
	fmt.Printf("   Business Rules Violated:\n")
	for i, rule := range err.Rules {
		fmt.Printf("     %d. %s\n", i+1, rule)
	}

	// Another business error example
	fmt.Println("\n   Order Processing Business Error:")
	orderErr := domainerrors.NewBusinessError("ORDER_PROCESSING_FAILED", "Order cannot be processed")
	orderErr.WithRule("Product must be in stock")
	orderErr.WithRule("Customer must have valid payment method")
	orderErr.WithRule("Shipping address must be valid")

	fmt.Printf("   Order Error: %s\n", orderErr.Error())
	fmt.Printf("   Rules Count: %d\n", len(orderErr.Rules))

	fmt.Println()
}

func databaseErrorExample() {
	fmt.Println("3. Database Errors:")

	// Database connection error
	connErr := errors.New("connection pool exhausted")
	err := domainerrors.NewDatabaseError("DB_CONNECTION_FAILED", "Database connection failed", connErr)
	err.WithOperation("SELECT", "users")
	err.WithQuery("SELECT * FROM users WHERE email = ? AND status = ?")

	fmt.Printf("   Error: %s\n", err.Error())
	fmt.Printf("   Operation: %s on table %s\n", err.Operation, err.Table)
	fmt.Printf("   Query: %s\n", err.Query)
	fmt.Printf("   Root Cause: %v\n", err.Unwrap())

	// Database constraint violation
	fmt.Println("\n   Database Constraint Violation:")
	constraintErr := domainerrors.NewDatabaseError("DB_CONSTRAINT_VIOLATION", "Unique constraint violation",
		errors.New("UNIQUE constraint failed: users.email"))
	constraintErr.WithOperation("INSERT", "users")
	constraintErr.WithQuery("INSERT INTO users (email, name) VALUES (?, ?)")

	fmt.Printf("   Constraint Error: %s\n", constraintErr.Error())
	fmt.Printf("   Operation: %s\n", constraintErr.Operation)

	fmt.Println()
}

func externalServiceErrorExample() {
	fmt.Println("4. External Service Errors:")

	// Payment service error
	paymentErr := domainerrors.NewExternalServiceError("PAYMENT_API_ERROR", "payment-gateway",
		"Payment processing failed", errors.New("connection timeout"))
	paymentErr.WithEndpoint("/api/v1/payments/charge")
	paymentErr.WithResponse(503, "Service temporarily unavailable")

	fmt.Printf("   Error: %s\n", paymentErr.Error())
	fmt.Printf("   Service: %s\n", paymentErr.Service)
	fmt.Printf("   Endpoint: %s\n", paymentErr.Endpoint)
	fmt.Printf("   Response Status: %d\n", paymentErr.StatusCode)
	fmt.Printf("   Response Body: %s\n", paymentErr.Response)

	// Email service error
	fmt.Println("\n   Email Service Error:")
	emailErr := domainerrors.NewExternalServiceError("EMAIL_SERVICE_ERROR", "email-service",
		"Failed to send email", errors.New("SMTP connection failed"))
	emailErr.WithEndpoint("/api/v1/emails/send")
	emailErr.WithResponse(500, "Internal server error")

	fmt.Printf("   Email Error: %s\n", emailErr.Error())
	fmt.Printf("   Service: %s\n", emailErr.Service)

	fmt.Println()
}

func authErrorExample() {
	fmt.Println("5. Authentication/Authorization Errors:")

	// Authentication error
	authErr := domainerrors.NewAuthenticationError("INVALID_TOKEN", "Invalid authentication token",
		errors.New("token expired"))
	authErr.WithScheme("Bearer")

	fmt.Printf("   Auth Error: %s\n", authErr.Error())
	fmt.Printf("   Scheme: %s\n", authErr.Scheme)

	// Authorization error
	fmt.Println("\n   Authorization Error:")
	authzErr := domainerrors.NewAuthorizationError("INSUFFICIENT_PERMISSIONS", "Insufficient permissions", nil)
	authzErr.WithPermission("admin:write", "users")

	fmt.Printf("   Authz Error: %s\n", authzErr.Error())
	fmt.Printf("   Permission: %s\n", authzErr.Permission)
	fmt.Printf("   Resource: %s\n", authzErr.Resource)

	// Security error
	fmt.Println("\n   Security Error:")
	secErr := domainerrors.NewSecurityError("SUSPICIOUS_ACTIVITY", "Suspicious activity detected")
	secErr.WithThreat("brute_force", "high")
	secErr.WithClientIP("192.168.1.100")

	fmt.Printf("   Security Error: %s\n", secErr.Error())
	fmt.Printf("   Threat Type: %s\n", secErr.ThreatType)
	fmt.Printf("   Severity: %s\n", secErr.Severity)
	fmt.Printf("   Client IP: %s\n", secErr.ClientIP)

	fmt.Println()
}

func timeoutCircuitBreakerExample() {
	fmt.Println("6. Timeout and Circuit Breaker Errors:")

	// Timeout error
	timeoutErr := domainerrors.NewTimeoutError("API_CALL_TIMEOUT", "external-api", "API call timeout",
		errors.New("context deadline exceeded"))
	timeoutErr.WithDuration(5*time.Second, 3*time.Second)

	fmt.Printf("   Timeout Error: %s\n", timeoutErr.Error())
	fmt.Printf("   Operation: %s\n", timeoutErr.Operation)
	fmt.Printf("   Duration: %v\n", timeoutErr.Duration)
	fmt.Printf("   Timeout: %v\n", timeoutErr.Timeout)

	// Circuit breaker error
	fmt.Println("\n   Circuit Breaker Error:")
	cbErr := domainerrors.NewCircuitBreakerError("CIRCUIT_BREAKER_OPEN", "payment-service",
		"Circuit breaker is open")
	cbErr.WithCircuitState("OPEN", 5)

	fmt.Printf("   Circuit Breaker Error: %s\n", cbErr.Error())
	fmt.Printf("   Circuit Name: %s\n", cbErr.CircuitName)
	fmt.Printf("   State: %s\n", cbErr.State)
	fmt.Printf("   Failures: %d\n", cbErr.Failures)

	fmt.Println()
}

func rateLimitingExample() {
	fmt.Println("7. Rate Limiting Errors:")

	// Rate limit error
	rateErr := domainerrors.NewRateLimitError("RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
	rateErr.WithRateLimit(100, 0, "2025-01-18T15:30:00Z", "3600s")

	fmt.Printf("   Rate Limit Error: %s\n", rateErr.Error())
	fmt.Printf("   Limit: %d\n", rateErr.Limit)
	fmt.Printf("   Remaining: %d\n", rateErr.Remaining)
	fmt.Printf("   Reset Time: %s\n", rateErr.ResetTime)
	fmt.Printf("   Window: %s\n", rateErr.Window)

	// Resource exhausted error
	fmt.Println("\n   Resource Exhausted Error:")
	resErr := domainerrors.NewResourceExhaustedError("MEMORY_EXHAUSTED", "memory", "Memory limit exceeded")
	resErr.WithLimits(1024, 1024, "MB")

	fmt.Printf("   Resource Error: %s\n", resErr.Error())
	fmt.Printf("   Resource: %s\n", resErr.Resource)
	fmt.Printf("   Limit: %d %s\n", resErr.Limit, resErr.Unit)
	fmt.Printf("   Used: %d %s\n", resErr.Used, resErr.Unit)

	fmt.Println()
}

func infrastructureErrorExample() {
	fmt.Println("8. Infrastructure Errors:")

	// Infrastructure error
	infraErr := domainerrors.NewInfrastructureError("REDIS_CONNECTION_FAILED", "redis",
		"Redis connection failed", errors.New("connection refused"))
	infraErr.WithAction("SET")

	fmt.Printf("   Infrastructure Error: %s\n", infraErr.Error())
	fmt.Printf("   Component: %s\n", infraErr.Component)
	fmt.Printf("   Action: %s\n", infraErr.Action)

	// Dependency error
	fmt.Println("\n   Dependency Error:")
	depErr := domainerrors.NewDependencyError("ELASTICSEARCH_UNAVAILABLE", "elasticsearch",
		"Elasticsearch service unavailable", errors.New("service not responding"))
	depErr.WithDependencyInfo("7.15.0", "unhealthy")

	fmt.Printf("   Dependency Error: %s\n", depErr.Error())
	fmt.Printf("   Dependency: %s\n", depErr.Dependency)
	fmt.Printf("   Version: %s\n", depErr.Version)
	fmt.Printf("   Status: %s\n", depErr.Status)

	fmt.Println()
}

func workflowErrorExample() {
	fmt.Println("9. Workflow Errors:")

	// Workflow error
	workflowErr := domainerrors.NewWorkflowError("ORDER_WORKFLOW_FAILED", "order-processing",
		"payment-validation", "Payment validation failed")
	workflowErr.WithStateInfo("pending_payment", "payment_validated")

	fmt.Printf("   Workflow Error: %s\n", workflowErr.Error())
	fmt.Printf("   Workflow: %s\n", workflowErr.WorkflowName)
	fmt.Printf("   Step: %s\n", workflowErr.StepName)
	fmt.Printf("   Current State: %s\n", workflowErr.CurrentState)
	fmt.Printf("   Expected State: %s\n", workflowErr.ExpectedState)

	// Service unavailable error
	fmt.Println("\n   Service Unavailable Error:")
	svcErr := domainerrors.NewServiceUnavailableError("SERVICE_MAINTENANCE", "user-service",
		"Service under maintenance", nil)
	svcErr.WithRetryInfo("3600s")
	svcErr.WithEndpoint("/api/v1/users")

	fmt.Printf("   Service Error: %s\n", svcErr.Error())
	fmt.Printf("   Service: %s\n", svcErr.ServiceName)
	fmt.Printf("   Retry After: %s\n", svcErr.RetryAfter)
	fmt.Printf("   Endpoint: %s\n", svcErr.Endpoint)

	fmt.Println()
}

func cacheConfigurationExample() {
	fmt.Println("10. Cache and Configuration Errors:")

	// Cache error
	cacheErr := domainerrors.NewCacheError("CACHE_MISS", "redis", "GET", "Cache miss", nil)

	fmt.Printf("   Cache Error: %s\n", cacheErr.Error())
	fmt.Printf("   Cache Type: %s\n", cacheErr.CacheType)
	fmt.Printf("   Operation: %s\n", cacheErr.Operation)
	fmt.Printf("   Key: %s\n", cacheErr.Key)
	fmt.Printf("   TTL: %s\n", cacheErr.TTL)

	// Configuration error
	fmt.Println("\n   Configuration Error:")
	configErr := domainerrors.NewConfigurationError("INVALID_CONFIG", "database.max_connections",
		"Invalid database configuration", nil)
	configErr.WithConfigDetails("integer between 1-100", "0")

	fmt.Printf("   Config Error: %s\n", configErr.Error())
	fmt.Printf("   Config Key: %s\n", configErr.ConfigKey)
	fmt.Printf("   Expected: %s\n", configErr.Expected)
	fmt.Printf("   Received: %s\n", configErr.Received)

	// Migration error
	fmt.Println("\n   Migration Error:")
	migErr := domainerrors.NewMigrationError("MIGRATION_FAILED", "v2.1.0", "Database migration failed",
		errors.New("foreign key constraint failed"))
	migErr.WithMigrationDetails("002_add_user_roles.sql", "rollback")

	fmt.Printf("   Migration Error: %s\n", migErr.Error())
	fmt.Printf("   Version: %s\n", migErr.Version)
	fmt.Printf("   Script: %s\n", migErr.Script)
	fmt.Printf("   Stage: %s\n", migErr.Stage)

	// Serialization error
	fmt.Println("\n   Serialization Error:")
	serErr := domainerrors.NewSerializationError("JSON_MARSHAL_ERROR", "json", "JSON serialization failed",
		errors.New("unsupported type"))
	serErr.WithTypeInfo("user.CreatedAt", "time.Time", "string")

	fmt.Printf("   Serialization Error: %s\n", serErr.Error())
	fmt.Printf("   Format: %s\n", serErr.Format)
	fmt.Printf("   Field: %s\n", serErr.Field)
	fmt.Printf("   Expected: %s\n", serErr.Expected)
	fmt.Printf("   Received: %s\n", serErr.Received)

	fmt.Println()
}

func init() {
	log.SetFlags(0) // Remove timestamp from logs for cleaner output
}
