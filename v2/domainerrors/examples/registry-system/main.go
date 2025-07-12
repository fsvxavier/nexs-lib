package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/factory"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

func main() {
	fmt.Println("🗂️ Domain Errors v2 - Registry System Examples")
	fmt.Println("===============================================")

	basicRegistryExample()
	domainRegistryExample()
	registryWithMiddlewareExample()
	errorMappingExample()
	registryPerformanceExample()
	distributedRegistryExample()
	registryObservabilityExample()
}

// basicRegistryExample demonstrates basic registry functionality
func basicRegistryExample() {
	fmt.Println("\n📋 Basic Registry Example:")

	registry := NewErrorRegistry()

	// Register different error types
	registry.Register("USER_NOT_FOUND", ErrorRegistration{
		Code:       "USER_NOT_FOUND",
		Message:    "User not found",
		Type:       string(types.ErrorTypeNotFound),
		Severity:   types.SeverityMedium,
		StatusCode: 404,
		Category:   "user",
		Retryable:  false,
		Tags:       []string{"user", "not_found"},
	})

	registry.Register("PAYMENT_FAILED", ErrorRegistration{
		Code:       "PAYMENT_FAILED",
		Message:    "Payment processing failed",
		Type:       string(types.ErrorTypeExternalService),
		Severity:   types.SeverityHigh,
		StatusCode: 402,
		Category:   "payment",
		Retryable:  true,
		Tags:       []string{"payment", "external", "retryable"},
	})

	registry.Register("RATE_LIMIT_EXCEEDED", ErrorRegistration{
		Code:       "RATE_LIMIT_EXCEEDED",
		Message:    "Rate limit exceeded",
		Type:       string(types.ErrorTypeRateLimit),
		Severity:   types.SeverityMedium,
		StatusCode: 429,
		Category:   "security",
		Retryable:  true,
		Tags:       []string{"rate_limit", "security", "retryable"},
	})

	// Create errors from registry
	userErr := registry.CreateError("USER_NOT_FOUND", map[string]interface{}{
		"user_id":   "12345",
		"operation": "fetch_profile",
	})

	paymentErr := registry.CreateError("PAYMENT_FAILED", map[string]interface{}{
		"payment_id": "pay_67890",
		"amount":     99.99,
		"currency":   "USD",
	})

	rateLimitErr := registry.CreateError("RATE_LIMIT_EXCEEDED", map[string]interface{}{
		"client_id": "client_123",
		"limit":     1000,
		"window":    "1h",
	})

	errors := []interfaces.DomainErrorInterface{userErr, paymentErr, rateLimitErr}

	fmt.Println("  Registered Errors:")
	for _, err := range errors {
		reg, found := registry.GetRegistration(err.Code())
		if !found {
			continue
		}
		fmt.Printf("    %s: %s\n", err.Code(), err.Error())
		fmt.Printf("      Category: %s, Retryable: %v\n", reg.Category, reg.Retryable)
		fmt.Printf("      Tags: %v\n", err.Tags())
	}
}

// domainRegistryExample shows domain-specific registries
func domainRegistryExample() {
	fmt.Println("\n🏢 Domain Registry Example:")

	// User domain registry
	userRegistry := NewDomainRegistry("user")
	userRegistry.RegisterDomainErrors(map[string]ErrorRegistration{
		"USER_NOT_FOUND": {
			Code:       "USER_NOT_FOUND",
			Message:    "User not found",
			Type:       string(types.ErrorTypeNotFound),
			Severity:   types.SeverityMedium,
			StatusCode: 404,
			Category:   "user",
			Retryable:  false,
		},
		"USER_ALREADY_EXISTS": {
			Code:       "USER_ALREADY_EXISTS",
			Message:    "User already exists",
			Type:       string(types.ErrorTypeConflict),
			Severity:   types.SeverityMedium,
			StatusCode: 409,
			Category:   "user",
			Retryable:  false,
		},
		"INVALID_USER_DATA": {
			Code:       "INVALID_USER_DATA",
			Message:    "Invalid user data provided",
			Type:       string(types.ErrorTypeValidation),
			Severity:   types.SeverityLow,
			StatusCode: 400,
			Category:   "user",
			Retryable:  false,
		},
	})

	// Payment domain registry
	paymentRegistry := NewDomainRegistry("payment")
	paymentRegistry.RegisterDomainErrors(map[string]ErrorRegistration{
		"PAYMENT_FAILED": {
			Code:       "PAYMENT_FAILED",
			Message:    "Payment processing failed",
			Type:       string(types.ErrorTypeExternalService),
			Severity:   types.SeverityHigh,
			StatusCode: 402,
			Category:   "payment",
			Retryable:  true,
		},
		"INSUFFICIENT_FUNDS": {
			Code:       "INSUFFICIENT_FUNDS",
			Message:    "Insufficient funds for transaction",
			Type:       string(types.ErrorTypeBusinessRule),
			Severity:   types.SeverityMedium,
			StatusCode: 402,
			Category:   "payment",
			Retryable:  false,
		},
		"CARD_EXPIRED": {
			Code:       "CARD_EXPIRED",
			Message:    "Payment card has expired",
			Type:       string(types.ErrorTypeBusinessRule),
			Severity:   types.SeverityMedium,
			StatusCode: 402,
			Category:   "payment",
			Retryable:  false,
		},
	})

	// Create global registry manager
	manager := NewRegistryManager()
	manager.RegisterDomain("user", userRegistry)
	manager.RegisterDomain("payment", paymentRegistry)

	// Create errors through manager
	userNotFound := manager.CreateDomainError("user", "USER_NOT_FOUND", map[string]interface{}{
		"user_id": "user_123",
	})

	paymentFailed := manager.CreateDomainError("payment", "PAYMENT_FAILED", map[string]interface{}{
		"payment_id": "pay_456",
		"gateway":    "stripe",
	})

	fmt.Println("  Domain Errors:")
	fmt.Printf("    User: %s (Domain: %s)\n", userNotFound.Error(), userNotFound.Details()["domain"])
	fmt.Printf("    Payment: %s (Domain: %s)\n", paymentFailed.Error(), paymentFailed.Details()["domain"])

	// Show registry statistics
	fmt.Println("\n  Registry Statistics:")
	for domain, registry := range manager.GetDomains() {
		fmt.Printf("    %s: %d error types registered\n", domain, registry.Count())
	}
}

// registryWithMiddlewareExample demonstrates middleware pattern
func registryWithMiddlewareExample() {
	fmt.Println("\n🔧 Registry with Middleware Example:")

	registry := NewErrorRegistry()

	// Add middleware
	registry.AddMiddleware(NewLoggingMiddleware())
	registry.AddMiddleware(NewMetricsMiddleware())
	registry.AddMiddleware(NewEnrichmentMiddleware())

	// Register error
	registry.Register("API_ERROR", ErrorRegistration{
		Code:       "API_ERROR",
		Message:    "API request failed",
		Type:       string(types.ErrorTypeExternalService),
		Severity:   types.SeverityHigh,
		StatusCode: 500,
		Category:   "api",
		Retryable:  true,
	})

	// Create error (will go through middleware chain)
	apiErr := registry.CreateError("API_ERROR", map[string]interface{}{
		"endpoint":      "/api/v1/users",
		"method":        "GET",
		"response_time": 5000,
	})

	fmt.Printf("  Middleware-processed error: %s\n", apiErr.Error())
	fmt.Printf("  Request ID: %s\n", apiErr.Details()["request_id"])
	fmt.Printf("  Processing Time: %v\n", apiErr.Details()["processing_time"])
	fmt.Printf("  Middleware Count: %v\n", apiErr.Details()["middleware_count"])
}

// errorMappingExample shows error code mapping
func errorMappingExample() {
	fmt.Println("\n🗺️ Error Mapping Example:")

	mapper := NewErrorMapper()

	// Map internal codes to external codes
	mapper.AddMapping("USER_NOT_FOUND", "USR_404", "User resource not found")
	mapper.AddMapping("PAYMENT_FAILED", "PAY_001", "Payment processing error")
	mapper.AddMapping("RATE_LIMIT_EXCEEDED", "SEC_429", "Request rate limit exceeded")

	// Map error types to HTTP status codes
	mapper.AddStatusMapping(string(types.ErrorTypeNotFound), 404)
	mapper.AddStatusMapping(string(types.ErrorTypeValidation), 400)
	mapper.AddStatusMapping(string(types.ErrorTypeAuthentication), 401)
	mapper.AddStatusMapping(string(types.ErrorTypeAuthorization), 403)

	registry := NewErrorRegistry()
	registry.SetMapper(mapper)

	// Register errors
	registry.Register("USER_NOT_FOUND", ErrorRegistration{
		Code:     "USER_NOT_FOUND",
		Message:  "User not found",
		Type:     string(types.ErrorTypeNotFound),
		Severity: types.SeverityMedium,
		Category: "user",
	})

	// Create error and map it
	internalErr := registry.CreateError("USER_NOT_FOUND", map[string]interface{}{
		"user_id": "123",
	})

	externalErr := mapper.MapError(internalErr)

	fmt.Println("  Error Mapping:")
	fmt.Printf("    Internal Code: %s\n", internalErr.Code())
	fmt.Printf("    External Code: %s\n", externalErr.Code())
	fmt.Printf("    Internal Message: %s\n", internalErr.Error())
	fmt.Printf("    External Message: %s\n", externalErr.Error())
	fmt.Printf("    Status Code: %d\n", externalErr.StatusCode())
}

// registryPerformanceExample demonstrates performance optimizations
func registryPerformanceExample() {
	fmt.Println("\n⚡ Registry Performance Example:")

	registry := NewErrorRegistry()

	// Pre-register common errors
	commonErrors := map[string]ErrorRegistration{
		"VALIDATION_ERROR": {
			Code:     "VALIDATION_ERROR",
			Message:  "Validation failed",
			Type:     string(types.ErrorTypeValidation),
			Severity: types.SeverityLow,
			Category: "validation",
		},
		"NOT_FOUND": {
			Code:     "NOT_FOUND",
			Message:  "Resource not found",
			Type:     string(types.ErrorTypeNotFound),
			Severity: types.SeverityMedium,
			Category: "resource",
		},
		"INTERNAL_ERROR": {
			Code:     "INTERNAL_ERROR",
			Message:  "Internal server error",
			Type:     string(types.ErrorTypeInternal),
			Severity: types.SeverityCritical,
			Category: "system",
		},
	}

	for code, reg := range commonErrors {
		registry.Register(code, reg)
	}

	// Performance test
	start := time.Now()
	iterations := 10000

	for i := 0; i < iterations; i++ {
		_ = registry.CreateError("VALIDATION_ERROR", map[string]interface{}{
			"field":     "email",
			"iteration": i,
		})
	}

	duration := time.Since(start)
	fmt.Printf("  Performance Test:\n")
	fmt.Printf("    Created %d errors in %v\n", iterations, duration)
	fmt.Printf("    Average: %v per error\n", duration/time.Duration(iterations))
	fmt.Printf("    Rate: %.0f errors/second\n", float64(iterations)/duration.Seconds())

	// Cache hit test
	fmt.Printf("    Registry Cache Size: %d\n", registry.CacheSize())
	fmt.Printf("    Cache Hit Rate: %.2f%%\n", registry.CacheHitRate()*100)
}

// distributedRegistryExample shows distributed registry pattern
func distributedRegistryExample() {
	fmt.Println("\n🌐 Distributed Registry Example:")

	// Simulate multiple service registries
	services := map[string]*ErrorRegistry{
		"user-service":    NewErrorRegistry(),
		"payment-service": NewErrorRegistry(),
		"order-service":   NewErrorRegistry(),
	}

	// Register service-specific errors
	services["user-service"].Register("USER_SERVICE_ERROR", ErrorRegistration{
		Code:     "USER_SERVICE_ERROR",
		Message:  "User service unavailable",
		Category: "service",
		Severity: types.SeverityHigh,
	})

	services["payment-service"].Register("PAYMENT_SERVICE_ERROR", ErrorRegistration{
		Code:     "PAYMENT_SERVICE_ERROR",
		Message:  "Payment service unavailable",
		Category: "service",
		Severity: types.SeverityHigh,
	})

	services["order-service"].Register("ORDER_SERVICE_ERROR", ErrorRegistration{
		Code:     "ORDER_SERVICE_ERROR",
		Message:  "Order service unavailable",
		Category: "service",
		Severity: types.SeverityHigh,
	})

	// Create distributed registry
	distributedRegistry := NewDistributedRegistry()
	for service, registry := range services {
		distributedRegistry.RegisterService(service, registry)
	}

	// Create errors from different services
	userServiceErr := distributedRegistry.CreateServiceError("user-service", "USER_SERVICE_ERROR", map[string]interface{}{
		"operation": "get_user_profile",
	})

	paymentServiceErr := distributedRegistry.CreateServiceError("payment-service", "PAYMENT_SERVICE_ERROR", map[string]interface{}{
		"operation": "process_payment",
	})

	fmt.Println("  Distributed Errors:")
	fmt.Printf("    User Service: %s (Service: %s)\n", userServiceErr.Error(), userServiceErr.Details()["service"])
	fmt.Printf("    Payment Service: %s (Service: %s)\n", paymentServiceErr.Error(), paymentServiceErr.Details()["service"])

	// Show distributed statistics
	fmt.Println("\n  Distributed Statistics:")
	for service, registry := range distributedRegistry.GetServices() {
		fmt.Printf("    %s: %d error types\n", service, registry.Count())
	}
}

// registryObservabilityExample demonstrates observability features
func registryObservabilityExample() {
	fmt.Println("\n🔍 Registry Observability Example:")

	registry := NewErrorRegistry()

	// Enable observability
	registry.EnableMetrics()
	registry.EnableTracing()
	registry.EnableHealthChecks()

	// Register and create errors
	registry.Register("OBSERVED_ERROR", ErrorRegistration{
		Code:     "OBSERVED_ERROR",
		Message:  "This error is being observed",
		Category: "observability",
		Severity: types.SeverityMedium,
	})

	// Create multiple errors to generate metrics
	for i := 0; i < 5; i++ {
		_ = registry.CreateError("OBSERVED_ERROR", map[string]interface{}{
			"iteration": i,
			"timestamp": time.Now(),
		})
	}

	// Show observability data
	metrics := registry.GetMetrics()
	fmt.Println("  Registry Metrics:")
	fmt.Printf("    Total Errors Created: %d\n", metrics.TotalErrorsCreated)
	fmt.Printf("    Errors by Category: %+v\n", metrics.ErrorsByCategory)
	fmt.Printf("    Errors by Severity: %+v\n", metrics.ErrorsBySeverity)
	fmt.Printf("    Average Creation Time: %v\n", metrics.AverageCreationTime)

	// Health check
	health := registry.HealthCheck()
	fmt.Printf("  Registry Health: %s\n", health.Status)
	fmt.Printf("  Health Details: %s\n", health.Details)
}

// Registry Types and Implementations

type ErrorRegistration struct {
	Code       string
	Message    string
	Type       string
	Severity   types.ErrorSeverity
	StatusCode int
	Category   string
	Retryable  bool
	Tags       []string
	Metadata   map[string]interface{}
}

type ErrorRegistry struct {
	mu            sync.RWMutex
	registrations map[string]ErrorRegistration
	middleware    []RegistryMiddleware
	mapper        *ErrorMapper
	metrics       *RegistryMetrics
	factory       interfaces.ErrorFactory
}

func NewErrorRegistry() *ErrorRegistry {
	return &ErrorRegistry{
		registrations: make(map[string]ErrorRegistration),
		middleware:    make([]RegistryMiddleware, 0),
		metrics:       NewRegistryMetrics(),
		factory:       factory.GetDefaultFactory(),
	}
}

func (r *ErrorRegistry) Register(code string, registration ErrorRegistration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registrations[code] = registration
}

func (r *ErrorRegistry) GetRegistration(code string) (ErrorRegistration, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	reg, exists := r.registrations[code]
	return reg, exists
}

func (r *ErrorRegistry) CreateError(code string, details map[string]interface{}) interfaces.DomainErrorInterface {
	start := time.Now()

	r.mu.RLock()
	registration, exists := r.registrations[code]
	r.mu.RUnlock()

	if !exists {
		return r.factory.New("UNKNOWN_ERROR", "Unknown error code: "+code)
	}

	// Build error
	builder := r.factory.Builder().
		WithCode(registration.Code).
		WithMessage(registration.Message).
		WithType(registration.Type).
		WithSeverity(interfaces.Severity(registration.Severity)).
		WithTags(registration.Tags)

	if registration.StatusCode > 0 {
		builder.WithStatusCode(registration.StatusCode)
	}

	// Add details
	if details != nil {
		for key, value := range details {
			builder.WithDetail(key, value)
		}
	}

	err := builder.Build()

	// Apply middleware
	for _, middleware := range r.middleware {
		err = middleware.Process(err, registration)
	}

	// Update metrics
	r.metrics.RecordError(registration.Category, registration.Severity, time.Since(start))

	return err
}

func (r *ErrorRegistry) AddMiddleware(middleware RegistryMiddleware) {
	r.middleware = append(r.middleware, middleware)
}

func (r *ErrorRegistry) SetMapper(mapper *ErrorMapper) {
	r.mapper = mapper
}

func (r *ErrorRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.registrations)
}

func (r *ErrorRegistry) CacheSize() int {
	// Simulated cache size
	return r.Count()
}

func (r *ErrorRegistry) CacheHitRate() float64 {
	// Simulated cache hit rate
	return 0.95
}

func (r *ErrorRegistry) EnableMetrics() {
	r.metrics.enabled = true
}

func (r *ErrorRegistry) EnableTracing() {
	// Enable tracing
}

func (r *ErrorRegistry) EnableHealthChecks() {
	// Enable health checks
}

func (r *ErrorRegistry) GetMetrics() RegistryMetricsData {
	return r.metrics.GetData()
}

func (r *ErrorRegistry) HealthCheck() HealthStatus {
	return HealthStatus{
		Status:  "healthy",
		Details: "Registry is operating normally",
	}
}

// Domain Registry
type DomainRegistry struct {
	*ErrorRegistry
	domain string
}

func NewDomainRegistry(domain string) *DomainRegistry {
	return &DomainRegistry{
		ErrorRegistry: NewErrorRegistry(),
		domain:        domain,
	}
}

func (dr *DomainRegistry) RegisterDomainErrors(errors map[string]ErrorRegistration) {
	for code, registration := range errors {
		// Add domain prefix
		registration.Code = dr.domain + "_" + registration.Code
		if registration.Metadata == nil {
			registration.Metadata = make(map[string]interface{})
		}
		registration.Metadata["domain"] = dr.domain
		dr.Register(code, registration)
	}
}

// Registry Manager
type RegistryManager struct {
	domains map[string]*DomainRegistry
	mu      sync.RWMutex
}

func NewRegistryManager() *RegistryManager {
	return &RegistryManager{
		domains: make(map[string]*DomainRegistry),
	}
}

func (rm *RegistryManager) RegisterDomain(domain string, registry *DomainRegistry) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.domains[domain] = registry
}

func (rm *RegistryManager) CreateDomainError(domain, code string, details map[string]interface{}) interfaces.DomainErrorInterface {
	rm.mu.RLock()
	registry, exists := rm.domains[domain]
	rm.mu.RUnlock()

	if !exists {
		return factory.GetDefaultFactory().New("UNKNOWN_DOMAIN", "Unknown domain: "+domain)
	}

	if details == nil {
		details = make(map[string]interface{})
	}
	details["domain"] = domain

	return registry.CreateError(code, details)
}

func (rm *RegistryManager) GetDomains() map[string]*DomainRegistry {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make(map[string]*DomainRegistry)
	for k, v := range rm.domains {
		result[k] = v
	}
	return result
}

// Middleware
type RegistryMiddleware interface {
	Process(err interfaces.DomainErrorInterface, registration ErrorRegistration) interfaces.DomainErrorInterface
}

type LoggingMiddleware struct{}

func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

func (lm *LoggingMiddleware) Process(err interfaces.DomainErrorInterface, registration ErrorRegistration) interfaces.DomainErrorInterface {
	// Simulate logging
	fmt.Printf("    [LOG] Error created: %s\n", err.Code())
	return err
}

type MetricsMiddleware struct{}

func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{}
}

func (mm *MetricsMiddleware) Process(err interfaces.DomainErrorInterface, registration ErrorRegistration) interfaces.DomainErrorInterface {
	// Add metrics metadata
	builder := factory.GetDefaultFactory().Builder().
		WithCode(err.Code()).
		WithMessage(err.Error()).
		WithType(err.Type()).
		WithSeverity(err.Severity()).
		WithDetail("metric_timestamp", time.Now()).
		WithDetail("metric_category", registration.Category)

	// Copy existing details
	for key, value := range err.Details() {
		builder.WithDetail(key, value)
	}

	return builder.Build()
}

type EnrichmentMiddleware struct{}

func NewEnrichmentMiddleware() *EnrichmentMiddleware {
	return &EnrichmentMiddleware{}
}

func (em *EnrichmentMiddleware) Process(err interfaces.DomainErrorInterface, registration ErrorRegistration) interfaces.DomainErrorInterface {
	builder := factory.GetDefaultFactory().Builder().
		WithCode(err.Code()).
		WithMessage(err.Error()).
		WithType(err.Type()).
		WithSeverity(err.Severity()).
		WithDetail("request_id", fmt.Sprintf("req_%d", time.Now().UnixNano())).
		WithDetail("processing_time", time.Now().Format(time.RFC3339)).
		WithDetail("middleware_count", 3)

	// Copy existing details
	for key, value := range err.Details() {
		builder.WithDetail(key, value)
	}

	return builder.Build()
}

// Error Mapper
type ErrorMapper struct {
	mappings       map[string]ErrorMapping
	statusMappings map[string]int
}

type ErrorMapping struct {
	ExternalCode    string
	ExternalMessage string
}

func NewErrorMapper() *ErrorMapper {
	return &ErrorMapper{
		mappings:       make(map[string]ErrorMapping),
		statusMappings: make(map[string]int),
	}
}

func (em *ErrorMapper) AddMapping(internalCode, externalCode, externalMessage string) {
	em.mappings[internalCode] = ErrorMapping{
		ExternalCode:    externalCode,
		ExternalMessage: externalMessage,
	}
}

func (em *ErrorMapper) AddStatusMapping(errorType string, statusCode int) {
	em.statusMappings[errorType] = statusCode
}

func (em *ErrorMapper) MapError(err interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
	mapping, exists := em.mappings[err.Code()]
	if !exists {
		return err
	}

	statusCode := err.StatusCode()
	if mappedStatus, exists := em.statusMappings[err.Type()]; exists {
		statusCode = mappedStatus
	}

	builder := factory.GetDefaultFactory().Builder().
		WithCode(mapping.ExternalCode).
		WithMessage(mapping.ExternalMessage).
		WithType(err.Type()).
		WithSeverity(err.Severity()).
		WithStatusCode(statusCode)

	// Copy details
	for key, value := range err.Details() {
		builder.WithDetail(key, value)
	}

	return builder.Build()
}

// Distributed Registry
type DistributedRegistry struct {
	services map[string]*ErrorRegistry
	mu       sync.RWMutex
}

func NewDistributedRegistry() *DistributedRegistry {
	return &DistributedRegistry{
		services: make(map[string]*ErrorRegistry),
	}
}

func (dr *DistributedRegistry) RegisterService(service string, registry *ErrorRegistry) {
	dr.mu.Lock()
	defer dr.mu.Unlock()
	dr.services[service] = registry
}

func (dr *DistributedRegistry) CreateServiceError(service, code string, details map[string]interface{}) interfaces.DomainErrorInterface {
	dr.mu.RLock()
	registry, exists := dr.services[service]
	dr.mu.RUnlock()

	if !exists {
		return factory.GetDefaultFactory().New("UNKNOWN_SERVICE", "Unknown service: "+service)
	}

	if details == nil {
		details = make(map[string]interface{})
	}
	details["service"] = service

	return registry.CreateError(code, details)
}

func (dr *DistributedRegistry) GetServices() map[string]*ErrorRegistry {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	result := make(map[string]*ErrorRegistry)
	for k, v := range dr.services {
		result[k] = v
	}
	return result
}

// Metrics
type RegistryMetrics struct {
	enabled bool
	mu      sync.RWMutex
	data    RegistryMetricsData
}

type RegistryMetricsData struct {
	TotalErrorsCreated  int
	ErrorsByCategory    map[string]int
	ErrorsBySeverity    map[string]int
	AverageCreationTime time.Duration
	totalCreationTime   time.Duration
}

func NewRegistryMetrics() *RegistryMetrics {
	return &RegistryMetrics{
		data: RegistryMetricsData{
			ErrorsByCategory: make(map[string]int),
			ErrorsBySeverity: make(map[string]int),
		},
	}
}

func (rm *RegistryMetrics) RecordError(category string, severity types.ErrorSeverity, duration time.Duration) {
	if !rm.enabled {
		return
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.data.TotalErrorsCreated++
	rm.data.ErrorsByCategory[category]++
	rm.data.ErrorsBySeverity[severity.String()]++
	rm.data.totalCreationTime += duration
	rm.data.AverageCreationTime = rm.data.totalCreationTime / time.Duration(rm.data.TotalErrorsCreated)
}

func (rm *RegistryMetrics) GetData() RegistryMetricsData {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Deep copy
	result := RegistryMetricsData{
		TotalErrorsCreated:  rm.data.TotalErrorsCreated,
		AverageCreationTime: rm.data.AverageCreationTime,
		ErrorsByCategory:    make(map[string]int),
		ErrorsBySeverity:    make(map[string]int),
	}

	for k, v := range rm.data.ErrorsByCategory {
		result.ErrorsByCategory[k] = v
	}

	for k, v := range rm.data.ErrorsBySeverity {
		result.ErrorsBySeverity[k] = v
	}

	return result
}

// Health Check
type HealthStatus struct {
	Status  string
	Details string
}
