package main

import (
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/factory"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

func main() {
	fmt.Println("🏭 Domain Errors v2 - Factory Usage Examples")
	fmt.Println("=============================================")

	defaultFactoryExample()
	databaseFactoryExample()
	httpFactoryExample()
	customFactoryExample()
	factoryWithDependencyInjection()
	factoryChainExample()
	factoryPerformanceExample()
}

// defaultFactoryExample demonstrates the default factory
func defaultFactoryExample() {
	fmt.Println("\n📝 Default Factory Example:")

	factory := factory.GetDefaultFactory()

	// Common error types
	notFoundErr := factory.NewNotFound("User", "12345")
	unauthorizedErr := factory.NewUnauthorized("Invalid token")
	forbiddenErr := factory.NewForbidden("Access denied")
	internalErr := factory.NewInternal("System error", nil)
	badRequestErr := factory.NewBadRequest("Invalid request format")
	conflictErr := factory.NewConflict("Resource already exists")
	timeoutErr := factory.NewTimeout("Operation timeout")

	errors := []interfaces.DomainErrorInterface{
		notFoundErr, unauthorizedErr, forbiddenErr,
		internalErr, badRequestErr, conflictErr, timeoutErr,
	}

	for _, err := range errors {
		fmt.Printf("  %s: %s (Status: %d)\n",
			err.Type(), err.Error(), err.StatusCode())
	}
}

// databaseFactoryExample shows database-specific error factory
func databaseFactoryExample() {
	fmt.Println("\n🗄️ Database Factory Example:")

	dbFactory := factory.GetDatabaseFactory()

	// Database-specific errors
	connErr := dbFactory.NewConnectionError("postgresql", fmt.Errorf("connection refused"))
	queryErr := dbFactory.NewQueryError("SELECT * FROM users", fmt.Errorf("syntax error"))
	transactionErr := dbFactory.NewTransactionError("user_creation", fmt.Errorf("deadlock"))

	errors := []interfaces.DomainErrorInterface{
		connErr, queryErr, transactionErr,
	}

	fmt.Println("  Database Errors:")
	for _, err := range errors {
		fmt.Printf("    %s: %s\n", err.Code(), err.Error())
		if details := err.Details(); len(details) > 0 {
			fmt.Printf("      Details: %+v\n", details)
		}
	}
}

// httpFactoryExample demonstrates HTTP-specific factory
func httpFactoryExample() {
	fmt.Println("\n🌐 HTTP Factory Example:")

	httpFactory := factory.GetHTTPFactory()

	// HTTP-specific errors
	httpErr := httpFactory.NewHTTPError(404, "Resource not found")
	serviceUnavailableErr := httpFactory.NewServiceUnavailableError("payment-service")
	unauthorizedErr := httpFactory.NewUnauthorized("Bearer token invalid")
	internalServerErr := httpFactory.NewInternal("Internal server error", fmt.Errorf("database down"))

	errors := []interfaces.DomainErrorInterface{
		httpErr, serviceUnavailableErr, unauthorizedErr, internalServerErr,
	}

	fmt.Println("  HTTP Errors:")
	for _, err := range errors {
		fmt.Printf("    %s: %s (Status: %d)\n",
			err.Code(), err.Error(), err.StatusCode())
	}
}

// customFactoryExample shows how to create custom factories
func customFactoryExample() {
	fmt.Println("\n🎯 Custom Factory Example:")

	// Payment processing factory
	paymentFactory := NewPaymentFactory()

	paymentFailed := paymentFactory.NewPaymentFailed("pay_123", "insufficient_funds", 99.99, "USD")
	cardDeclined := paymentFactory.NewCardDeclined("card_456", "expired_card")
	fraudDetected := paymentFactory.NewFraudDetected("pay_789", "suspicious_activity")
	refundFailed := paymentFactory.NewRefundFailed("ref_321", "original_payment_not_found")

	errors := []interfaces.DomainErrorInterface{
		paymentFailed, cardDeclined, fraudDetected, refundFailed,
	}

	fmt.Println("  Payment Errors:")
	for _, err := range errors {
		fmt.Printf("    %s: %s\n", err.Code(), err.Error())
		fmt.Printf("      Severity: %v\n", err.Severity())
		fmt.Printf("      Details: %+v\n", err.Details())
		fmt.Printf("      Tags: %v\n", err.Tags())
	}
}

// factoryWithDependencyInjection shows DI pattern with factories
func factoryWithDependencyInjection() {
	fmt.Println("\n💉 Factory with Dependency Injection:")

	// Service configuration
	config := ServiceConfig{
		ServiceName:   "user-service",
		Version:       "v1.2.3",
		Environment:   "production",
		CorrelationID: "req-123-456",
	}

	// Create factory with configuration
	serviceFactory := NewServiceFactory(config)

	// Service-specific errors with context
	serviceDown := serviceFactory.NewServiceUnavailable("payment-gateway", "Health check failed")
	circuitOpen := serviceFactory.NewCircuitBreakerOpen("notification-service")
	dependencyFailed := serviceFactory.NewDependencyFailed("redis-cache", fmt.Errorf("connection timeout"))
	configError := serviceFactory.NewConfigurationError("database.url", "required configuration missing")

	errors := []interfaces.DomainErrorInterface{
		serviceDown, circuitOpen, dependencyFailed, configError,
	}

	fmt.Println("  Service Errors:")
	for _, err := range errors {
		fmt.Printf("    %s: %s\n", err.Code(), err.Error())
		fmt.Printf("      Service: %s\n", err.Details()["service_name"])
		fmt.Printf("      Environment: %s\n", err.Details()["environment"])
		fmt.Printf("      Correlation ID: %s\n", err.Details()["correlation_id"])
	}
}

// factoryChainExample demonstrates chaining factories
func factoryChainExample() {
	fmt.Println("\n⛓️ Factory Chain Example:")

	// Simulate a request processing chain
	request := ProcessingRequest{
		RequestID: "req-999",
		UserID:    "user-123",
		Action:    "create_order",
		Data:      map[string]interface{}{"amount": 100.0},
	}

	// Chain of processing with different factories
	processors := []RequestProcessor{
		&ValidationProcessor{factory: factory.GetDefaultFactory()},
		&AuthProcessor{factory: factory.GetHTTPFactory()},
		&BusinessProcessor{factory: NewPaymentFactory()},
		&PersistenceProcessor{factory: factory.GetDatabaseFactory()},
	}

	fmt.Printf("  Processing request: %s\n", request.RequestID)

	for i, processor := range processors {
		fmt.Printf("    Step %d: %s\n", i+1, processor.Name())

		if err := processor.Process(request); err != nil {
			fmt.Printf("      ❌ Error: %s\n", err.Error())
			fmt.Printf("      Code: %s\n", err.Code())
			fmt.Printf("      Type: %s\n", err.Type())
			break
		} else {
			fmt.Printf("      ✅ Success\n")
		}
	}
}

// factoryPerformanceExample shows performance optimizations
func factoryPerformanceExample() {
	fmt.Println("\n⚡ Factory Performance Example:")

	start := time.Now()

	// Test different factories under load
	factories := map[string]interfaces.ErrorFactory{
		"Default":  factory.GetDefaultFactory(),
		"Database": factory.GetDatabaseFactory(),
		"HTTP":     factory.GetHTTPFactory(),
		"Payment":  NewPaymentFactory(),
	}

	iterations := 1000

	for name, f := range factories {
		factoryStart := time.Now()

		for i := 0; i < iterations; i++ {
			_ = f.NewInternal("Performance test", nil)
		}

		factoryDuration := time.Since(factoryStart)
		fmt.Printf("  %s Factory: %d errors in %v (avg: %v/error)\n",
			name, iterations, factoryDuration, factoryDuration/time.Duration(iterations))
	}

	totalDuration := time.Since(start)
	fmt.Printf("  Total performance test: %v\n", totalDuration)
	fmt.Printf("  Memory efficient: Object pooling enabled\n")
	fmt.Printf("  Thread safe: All factories are concurrent-safe\n")
}

// PaymentFactory - Custom factory for payment-related errors
type PaymentFactory struct {
	baseFactory interfaces.ErrorFactory
}

// NewPaymentFactory creates a new payment factory
func NewPaymentFactory() *PaymentFactory {
	return &PaymentFactory{
		baseFactory: factory.GetDefaultFactory(),
	}
}

// Implement ErrorFactory interface
func (pf *PaymentFactory) New(code, message string) interfaces.DomainErrorInterface {
	return pf.baseFactory.Builder().
		WithCode(code).
		WithMessage(message).
		WithTag("payment").
		Build()
}

func (pf *PaymentFactory) NewWithCause(code, message string, cause error) interfaces.DomainErrorInterface {
	return pf.baseFactory.Builder().
		WithCode(code).
		WithMessage(message).
		WithCause(cause).
		WithTag("payment").
		Build()
}

func (pf *PaymentFactory) NewValidation(message string, fields map[string][]string) interfaces.ValidationErrorInterface {
	return pf.baseFactory.NewValidation(message, fields)
}

func (pf *PaymentFactory) NewNotFound(entity, id string) interfaces.DomainErrorInterface {
	return pf.baseFactory.Builder().
		WithCode("PAY_NOT_FOUND").
		WithMessage(fmt.Sprintf("%s not found: %s", entity, id)).
		WithType(string(types.ErrorTypeNotFound)).
		WithDetail("entity", entity).
		WithDetail("id", id).
		WithTag("payment").
		WithTag("not_found").
		Build()
}

func (pf *PaymentFactory) NewUnauthorized(message string) interfaces.DomainErrorInterface {
	return pf.baseFactory.NewUnauthorized(message)
}

func (pf *PaymentFactory) NewForbidden(message string) interfaces.DomainErrorInterface {
	return pf.baseFactory.NewForbidden(message)
}

func (pf *PaymentFactory) NewInternal(message string, cause error) interfaces.DomainErrorInterface {
	return pf.baseFactory.NewInternal(message, cause)
}

func (pf *PaymentFactory) NewBadRequest(message string) interfaces.DomainErrorInterface {
	return pf.baseFactory.NewBadRequest(message)
}

func (pf *PaymentFactory) NewConflict(message string) interfaces.DomainErrorInterface {
	return pf.baseFactory.NewConflict(message)
}

func (pf *PaymentFactory) NewTimeout(message string) interfaces.DomainErrorInterface {
	return pf.baseFactory.NewTimeout(message)
}

func (pf *PaymentFactory) NewCircuitBreaker(service string) interfaces.DomainErrorInterface {
	return pf.baseFactory.NewCircuitBreaker(service)
}

func (pf *PaymentFactory) Builder() interfaces.ErrorBuilder {
	return pf.baseFactory.Builder().WithTag("payment")
}

// Payment-specific methods
func (pf *PaymentFactory) NewPaymentFailed(paymentID, reason string, amount float64, currency string) interfaces.DomainErrorInterface {
	return pf.baseFactory.Builder().
		WithCode("PAY_FAILED").
		WithMessage(fmt.Sprintf("Payment failed: %s", reason)).
		WithType(string(types.ErrorTypeExternalService)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithDetail("payment_id", paymentID).
		WithDetail("reason", reason).
		WithDetail("amount", amount).
		WithDetail("currency", currency).
		WithTag("payment").
		WithTag("failed").
		Build()
}

func (pf *PaymentFactory) NewCardDeclined(cardID, reason string) interfaces.DomainErrorInterface {
	return pf.baseFactory.Builder().
		WithCode("CARD_DECLINED").
		WithMessage(fmt.Sprintf("Card declined: %s", reason)).
		WithType(string(types.ErrorTypeExternalService)).
		WithSeverity(interfaces.Severity(types.SeverityMedium)).
		WithDetail("card_id", cardID).
		WithDetail("decline_reason", reason).
		WithTag("payment").
		WithTag("card").
		WithTag("declined").
		Build()
}

func (pf *PaymentFactory) NewFraudDetected(paymentID, details string) interfaces.DomainErrorInterface {
	return pf.baseFactory.Builder().
		WithCode("FRAUD_DETECTED").
		WithMessage("Suspicious activity detected").
		WithType(string(types.ErrorTypeSecurity)).
		WithSeverity(interfaces.Severity(types.SeverityCritical)).
		WithDetail("payment_id", paymentID).
		WithDetail("fraud_details", details).
		WithTag("payment").
		WithTag("security").
		WithTag("fraud").
		Build()
}

func (pf *PaymentFactory) NewRefundFailed(refundID, reason string) interfaces.DomainErrorInterface {
	return pf.baseFactory.Builder().
		WithCode("REFUND_FAILED").
		WithMessage(fmt.Sprintf("Refund failed: %s", reason)).
		WithType(string(types.ErrorTypeExternalService)).
		WithSeverity(interfaces.Severity(types.SeverityMedium)).
		WithDetail("refund_id", refundID).
		WithDetail("reason", reason).
		WithTag("payment").
		WithTag("refund").
		WithTag("failed").
		Build()
}

// ServiceFactory - Factory with dependency injection
type ServiceFactory struct {
	config      ServiceConfig
	baseFactory interfaces.ErrorFactory
}

type ServiceConfig struct {
	ServiceName   string
	Version       string
	Environment   string
	CorrelationID string
}

// NewServiceFactory creates a factory with configuration
func NewServiceFactory(config ServiceConfig) *ServiceFactory {
	return &ServiceFactory{
		config:      config,
		baseFactory: factory.GetDefaultFactory(),
	}
}

func (sf *ServiceFactory) NewServiceUnavailable(serviceName, reason string) interfaces.DomainErrorInterface {
	return sf.baseFactory.Builder().
		WithCode("SVC_UNAVAILABLE").
		WithMessage(fmt.Sprintf("Service unavailable: %s", serviceName)).
		WithType(string(types.ErrorTypeExternalService)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithDetail("target_service", serviceName).
		WithDetail("reason", reason).
		WithDetail("service_name", sf.config.ServiceName).
		WithDetail("service_version", sf.config.Version).
		WithDetail("environment", sf.config.Environment).
		WithDetail("correlation_id", sf.config.CorrelationID).
		WithTag("service").
		WithTag("unavailable").
		Build()
}

func (sf *ServiceFactory) NewCircuitBreakerOpen(serviceName string) interfaces.DomainErrorInterface {
	return sf.baseFactory.Builder().
		WithCode("CIRCUIT_OPEN").
		WithMessage(fmt.Sprintf("Circuit breaker open for service: %s", serviceName)).
		WithType(string(types.ErrorTypeCircuitBreaker)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithDetail("target_service", serviceName).
		WithDetail("service_name", sf.config.ServiceName).
		WithDetail("correlation_id", sf.config.CorrelationID).
		WithTag("circuit_breaker").
		WithTag("service").
		Build()
}

func (sf *ServiceFactory) NewDependencyFailed(dependency string, cause error) interfaces.DomainErrorInterface {
	return sf.baseFactory.Builder().
		WithCode("DEPENDENCY_FAILED").
		WithMessage(fmt.Sprintf("Dependency failed: %s", dependency)).
		WithType(string(types.ErrorTypeDependency)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithCause(cause).
		WithDetail("dependency", dependency).
		WithDetail("service_name", sf.config.ServiceName).
		WithDetail("correlation_id", sf.config.CorrelationID).
		WithTag("dependency").
		WithTag("service").
		Build()
}

func (sf *ServiceFactory) NewConfigurationError(configKey, reason string) interfaces.DomainErrorInterface {
	return sf.baseFactory.Builder().
		WithCode("CONFIG_ERROR").
		WithMessage(fmt.Sprintf("Configuration error: %s", configKey)).
		WithType(string(types.ErrorTypeConfiguration)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithDetail("config_key", configKey).
		WithDetail("reason", reason).
		WithDetail("service_name", sf.config.ServiceName).
		WithDetail("environment", sf.config.Environment).
		WithTag("configuration").
		WithTag("service").
		Build()
}

// Factory chain example types
type ProcessingRequest struct {
	RequestID string
	UserID    string
	Action    string
	Data      map[string]interface{}
}

type RequestProcessor interface {
	Name() string
	Process(request ProcessingRequest) interfaces.DomainErrorInterface
}

type ValidationProcessor struct {
	factory interfaces.ErrorFactory
}

func (vp *ValidationProcessor) Name() string { return "Validation" }

func (vp *ValidationProcessor) Process(request ProcessingRequest) interfaces.DomainErrorInterface {
	if request.UserID == "" {
		return vp.factory.NewBadRequest("User ID is required")
	}
	return nil
}

type AuthProcessor struct {
	factory interfaces.ErrorFactory
}

func (ap *AuthProcessor) Name() string { return "Authentication" }

func (ap *AuthProcessor) Process(request ProcessingRequest) interfaces.DomainErrorInterface {
	if request.UserID == "invalid-user" {
		return ap.factory.NewUnauthorized("Invalid user credentials")
	}
	return nil
}

type BusinessProcessor struct {
	factory *PaymentFactory
}

func (bp *BusinessProcessor) Name() string { return "Business Logic" }

func (bp *BusinessProcessor) Process(request ProcessingRequest) interfaces.DomainErrorInterface {
	if amount, ok := request.Data["amount"].(float64); ok && amount <= 0 {
		return bp.factory.NewPaymentFailed("invalid-payment", "invalid_amount", amount, "USD")
	}
	return nil
}

type PersistenceProcessor struct {
	factory interfaces.ErrorFactory
}

func (pp *PersistenceProcessor) Name() string { return "Data Persistence" }

func (pp *PersistenceProcessor) Process(request ProcessingRequest) interfaces.DomainErrorInterface {
	// Simulate database error for demo
	if request.Action == "create_order_fail" {
		return pp.factory.NewInternal("Database connection failed", fmt.Errorf("connection timeout"))
	}
	return nil
}
