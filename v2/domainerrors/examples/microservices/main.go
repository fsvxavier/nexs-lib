package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/factory"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

func main() {
	fmt.Println("🏗️ Domain Errors v2 - Microservices Examples")
	fmt.Println("=============================================")

	serviceRegistryExample()
	circuitBreakerExample()
	distributedTracingExample()
	errorPropagationExample()
	serviceMeshExample()
	timeoutHandlingExample()
	bulkheadExample()
	correlationExample()
}

// serviceRegistryExample demonstrates service registry pattern with error handling
func serviceRegistryExample() {
	fmt.Println("\n📋 Service Registry Example:")

	registry := NewServiceRegistry()

	// Register services
	services := []*ServiceInfo{
		{
			Name:           "user-service",
			Version:        "v1.2.3",
			Host:           "user-service.default.svc.cluster.local",
			Port:           8080,
			HealthEndpoint: "/health",
			Tags:           []string{"user", "authentication"},
		},
		{
			Name:           "payment-service",
			Version:        "v2.1.0",
			Host:           "payment-service.default.svc.cluster.local",
			Port:           8081,
			HealthEndpoint: "/health",
			Tags:           []string{"payment", "external"},
		},
		{
			Name:           "notification-service",
			Version:        "v1.0.1",
			Host:           "notification-service.default.svc.cluster.local",
			Port:           8082,
			HealthEndpoint: "/health",
			Tags:           []string{"notification", "async"},
		},
	}

	for _, service := range services {
		registry.Register(service)
	}

	// Simulate service discovery and calls
	fmt.Println("  Service Discovery Results:")

	testCases := []struct {
		serviceName string
		operation   string
		shouldFail  bool
	}{
		{"user-service", "get_user", false},
		{"payment-service", "process_payment", false},
		{"notification-service", "send_email", false},
		{"unknown-service", "some_operation", true},
		{"payment-service", "refund", true}, // Simulate service failure
	}

	for _, tc := range testCases {
		service, err := registry.Discover(tc.serviceName)
		if err != nil {
			fmt.Printf("    ❌ %s: %s\n", tc.serviceName, err.Error())
			fmt.Printf("       Code: %s, Type: %s\n", err.Code(), err.Type())
			continue
		}

		// Simulate service call
		result, callErr := registry.CallService(service, tc.operation, tc.shouldFail)
		if callErr != nil {
			fmt.Printf("    ❌ %s.%s: %s\n", tc.serviceName, tc.operation, callErr.Error())
			fmt.Printf("       Code: %s, Service: %s\n", callErr.Code(), callErr.Details()["service"])
		} else {
			fmt.Printf("    ✅ %s.%s: %s\n", tc.serviceName, tc.operation, result)
		}
	}
}

// circuitBreakerExample demonstrates circuit breaker pattern
func circuitBreakerExample() {
	fmt.Println("\n🔌 Circuit Breaker Example:")

	cb := NewCircuitBreaker("payment-service", CircuitBreakerConfig{
		FailureThreshold: 3,
		TimeoutDuration:  5 * time.Second,
		RecoveryTimeout:  10 * time.Second,
	})

	// Simulate multiple calls to demonstrate circuit breaker states
	fmt.Println("  Circuit Breaker State Transitions:")

	for i := 0; i < 10; i++ {
		// Simulate failing calls first, then successful ones
		shouldFail := i < 5

		result, err := cb.Call(func() (interface{}, interfaces.DomainErrorInterface) {
			if shouldFail {
				return nil, factory.GetDefaultFactory().Builder().
					WithCode("SERVICE_ERROR").
					WithMessage("Service temporarily unavailable").
					WithType(string(types.ErrorTypeExternalService)).
					WithSeverity(interfaces.Severity(types.SeverityHigh)).
					WithDetail("service", "payment-service").
					WithTag("circuit_breaker").
					Build()
			}
			return "Payment processed successfully", nil
		})

		state := cb.GetState()
		fmt.Printf("    Call %d - State: %s", i+1, state)

		if err != nil {
			fmt.Printf(" - ❌ Error: %s\n", err.Error())
			if err.Type() == string(types.ErrorTypeCircuitBreaker) {
				fmt.Printf("       Circuit breaker is open, calls are being rejected\n")
			}
		} else {
			fmt.Printf(" - ✅ Result: %s\n", result)
		}

		// Small delay to simulate real-world timing
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("  Final Circuit Breaker Stats:\n")
	stats := cb.GetStats()
	fmt.Printf("    Total Calls: %d\n", stats.TotalCalls)
	fmt.Printf("    Success Rate: %.2f%%\n", stats.SuccessRate*100)
	fmt.Printf("    Current State: %s\n", cb.GetState())
}

// distributedTracingExample demonstrates distributed tracing with error context
func distributedTracingExample() {
	fmt.Println("\n🔍 Distributed Tracing Example:")

	tracer := NewDistributedTracer()

	// Simulate a request flow through multiple services
	ctx := context.Background()
	traceID := tracer.StartTrace(ctx, "order_processing")

	fmt.Printf("  Trace ID: %s\n", traceID)
	fmt.Println("  Request Flow:")

	// Service 1: User Service
	ctx, span1 := tracer.StartSpan(ctx, "user-service", "validate_user")
	userResult, userErr := simulateServiceCall("user-service", "validate_user", false)
	span1.Finish(userErr)

	if userErr != nil {
		fmt.Printf("    ❌ user-service.validate_user: %s\n", userErr.Error())
		tracer.RecordError(ctx, userErr)
		return
	}
	fmt.Printf("    ✅ user-service.validate_user: %s\n", userResult)

	// Service 2: Payment Service
	ctx, span2 := tracer.StartSpan(ctx, "payment-service", "process_payment")
	paymentResult, paymentErr := simulateServiceCall("payment-service", "process_payment", false)
	span2.Finish(paymentErr)

	if paymentErr != nil {
		fmt.Printf("    ❌ payment-service.process_payment: %s\n", paymentErr.Error())
		tracer.RecordError(ctx, paymentErr)
		return
	}
	fmt.Printf("    ✅ payment-service.process_payment: %s\n", paymentResult)

	// Service 3: Inventory Service (with error)
	ctx, span3 := tracer.StartSpan(ctx, "inventory-service", "reserve_items")
	_, inventoryErr := simulateServiceCall("inventory-service", "reserve_items", true)
	span3.Finish(inventoryErr)

	if inventoryErr != nil {
		fmt.Printf("    ❌ inventory-service.reserve_items: %s\n", inventoryErr.Error())
		tracer.RecordError(ctx, inventoryErr)

		// Show distributed error context
		fmt.Printf("  Distributed Error Context:\n")
		fmt.Printf("    Trace ID: %s\n", inventoryErr.Details()["trace_id"])
		fmt.Printf("    Span ID: %s\n", inventoryErr.Details()["span_id"])
		fmt.Printf("    Service: %s\n", inventoryErr.Details()["service"])
		fmt.Printf("    Operation: %s\n", inventoryErr.Details()["operation"])
		fmt.Printf("    Parent Spans: %v\n", inventoryErr.Details()["parent_spans"])
		return
	}

	tracer.FinishTrace(ctx)
	fmt.Println("  ✅ Order processing completed successfully")
}

// errorPropagationExample demonstrates error propagation across services
func errorPropagationExample() {
	fmt.Println("\n🔄 Error Propagation Example:")

	propagator := NewErrorPropagator()

	// Simulate error chain across services
	fmt.Println("  Error Propagation Chain:")

	// Original error from database
	dbErr := factory.GetDefaultFactory().Builder().
		WithCode("DB_CONNECTION_ERROR").
		WithMessage("Database connection failed").
		WithType(string(types.ErrorTypeDatabase)).
		WithSeverity(interfaces.Severity(types.SeverityCritical)).
		WithDetail("database", "postgresql").
		WithDetail("host", "db.cluster.local").
		WithTag("database").
		Build()

	// Propagate through user service
	userServiceErr := propagator.PropagateError(dbErr, PropagationContext{
		Service:   "user-service",
		Operation: "get_user_profile",
		Version:   "v1.2.3",
	})

	// Propagate through API gateway
	gatewayErr := propagator.PropagateError(userServiceErr, PropagationContext{
		Service:   "api-gateway",
		Operation: "proxy_request",
		Version:   "v2.1.0",
	})

	// Show propagation chain
	fmt.Printf("    Original Error:\n")
	fmt.Printf("      Service: database\n")
	fmt.Printf("      Code: %s\n", dbErr.Code())
	fmt.Printf("      Message: %s\n", dbErr.Error())

	fmt.Printf("    User Service Error:\n")
	fmt.Printf("      Service: %s\n", userServiceErr.Details()["service"])
	fmt.Printf("      Code: %s\n", userServiceErr.Code())
	fmt.Printf("      Message: %s\n", userServiceErr.Error())
	fmt.Printf("      Root Cause: %s\n", userServiceErr.Details()["root_cause_code"])

	fmt.Printf("    API Gateway Error:\n")
	fmt.Printf("      Service: %s\n", gatewayErr.Details()["service"])
	fmt.Printf("      Code: %s\n", gatewayErr.Code())
	fmt.Printf("      Message: %s\n", gatewayErr.Error())
	fmt.Printf("      Error Chain Length: %v\n", gatewayErr.Details()["chain_length"])
	fmt.Printf("      Services Involved: %v\n", gatewayErr.Details()["services_involved"])
}

// serviceMeshExample demonstrates service mesh error handling
func serviceMeshExample() {
	fmt.Println("\n🕸️ Service Mesh Example:")

	mesh := NewServiceMesh()

	// Configure services in the mesh
	services := []string{"user-service", "payment-service", "order-service", "notification-service"}
	for _, service := range services {
		mesh.RegisterService(service)
	}

	// Configure traffic policies
	mesh.SetRetryPolicy("payment-service", RetryPolicy{
		MaxRetries:    3,
		BackoffFactor: 2,
		InitialDelay:  100 * time.Millisecond,
	})

	mesh.SetTimeoutPolicy("order-service", TimeoutPolicy{
		RequestTimeout:    5 * time.Second,
		ConnectionTimeout: 2 * time.Second,
	})

	// Simulate service mesh routing with errors
	fmt.Println("  Service Mesh Routing:")

	testCases := []struct {
		from       string
		to         string
		operation  string
		shouldFail bool
	}{
		{"api-gateway", "user-service", "authenticate", false},
		{"order-service", "payment-service", "charge_card", true}, // Will retry
		{"order-service", "notification-service", "send_confirmation", false},
		{"user-service", "unknown-service", "some_operation", true}, // No retry for unknown service
	}

	for _, tc := range testCases {
		fmt.Printf("    %s -> %s.%s:\n", tc.from, tc.to, tc.operation)

		result, err := mesh.RouteRequest(tc.from, tc.to, tc.operation, tc.shouldFail)

		if err != nil {
			fmt.Printf("      ❌ Error: %s\n", err.Error())
			fmt.Printf("      Code: %s\n", err.Code())
			fmt.Printf("      Retry Attempts: %v\n", err.Details()["retry_attempts"])
			fmt.Printf("      Total Duration: %v\n", err.Details()["total_duration"])
		} else {
			fmt.Printf("      ✅ Success: %s\n", result)
		}
	}

	// Show mesh statistics
	fmt.Println("\n  Service Mesh Statistics:")
	stats := mesh.GetStatistics()
	for service, stat := range stats {
		fmt.Printf("    %s:\n", service)
		fmt.Printf("      Total Requests: %d\n", stat.TotalRequests)
		fmt.Printf("      Success Rate: %.2f%%\n", stat.SuccessRate*100)
		fmt.Printf("      Average Latency: %v\n", stat.AverageLatency)
		fmt.Printf("      Error Rate: %.2f%%\n", (1-stat.SuccessRate)*100)
	}
}

// timeoutHandlingExample demonstrates timeout handling in distributed systems
func timeoutHandlingExample() {
	fmt.Println("\n⏰ Timeout Handling Example:")

	timeoutHandler := NewTimeoutHandler()

	testCases := []struct {
		service   string
		operation string
		duration  time.Duration
		timeout   time.Duration
	}{
		{"fast-service", "quick_operation", 100 * time.Millisecond, 1 * time.Second},        // Success
		{"slow-service", "slow_operation", 2 * time.Second, 1 * time.Second},                // Timeout
		{"variable-service", "variable_operation", 500 * time.Millisecond, 1 * time.Second}, // Success
		{"stuck-service", "stuck_operation", 10 * time.Second, 2 * time.Second},             // Timeout
	}

	fmt.Println("  Timeout Handling Results:")

	for _, tc := range testCases {
		fmt.Printf("    %s.%s (timeout: %v):\n", tc.service, tc.operation, tc.timeout)

		start := time.Now()
		result, err := timeoutHandler.CallWithTimeout(tc.service, tc.operation, tc.duration, tc.timeout)
		actualDuration := time.Since(start)

		if err != nil {
			fmt.Printf("      ❌ Error: %s\n", err.Error())
			fmt.Printf("      Code: %s\n", err.Code())
			fmt.Printf("      Expected Duration: %v\n", tc.duration)
			fmt.Printf("      Actual Duration: %v\n", actualDuration)
			fmt.Printf("      Timeout Exceeded: %v\n", actualDuration >= tc.timeout)
		} else {
			fmt.Printf("      ✅ Success: %s\n", result)
			fmt.Printf("      Duration: %v\n", actualDuration)
		}
	}
}

// bulkheadExample demonstrates bulkhead pattern for fault isolation
func bulkheadExample() {
	fmt.Println("\n🚢 Bulkhead Pattern Example:")

	bulkhead := NewBulkhead()

	// Configure resource pools
	bulkhead.CreatePool("critical", BulkheadConfig{
		MaxConcurrency: 5,
		QueueSize:      10,
		Timeout:        2 * time.Second,
	})

	bulkhead.CreatePool("normal", BulkheadConfig{
		MaxConcurrency: 3,
		QueueSize:      5,
		Timeout:        1 * time.Second,
	})

	bulkhead.CreatePool("background", BulkheadConfig{
		MaxConcurrency: 2,
		QueueSize:      20,
		Timeout:        5 * time.Second,
	})

	// Simulate concurrent requests
	fmt.Println("  Bulkhead Resource Isolation:")

	var wg sync.WaitGroup
	results := make(chan string, 20)
	errors := make(chan interfaces.DomainErrorInterface, 20)

	// Critical requests
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			result, err := bulkhead.Execute("critical", fmt.Sprintf("critical-task-%d", id), func() (interface{}, interfaces.DomainErrorInterface) {
				time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
				return fmt.Sprintf("Critical task %d completed", id), nil
			})

			if err != nil {
				errors <- err
			} else {
				results <- result.(string)
			}
		}(i)
	}

	// Normal requests
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			result, err := bulkhead.Execute("normal", fmt.Sprintf("normal-task-%d", id), func() (interface{}, interfaces.DomainErrorInterface) {
				time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
				return fmt.Sprintf("Normal task %d completed", id), nil
			})

			if err != nil {
				errors <- err
			} else {
				results <- result.(string)
			}
		}(i)
	}

	// Background requests
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			result, err := bulkhead.Execute("background", fmt.Sprintf("background-task-%d", id), func() (interface{}, interfaces.DomainErrorInterface) {
				time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
				return fmt.Sprintf("Background task %d completed", id), nil
			})

			if err != nil {
				errors <- err
			} else {
				results <- result.(string)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)
	close(errors)

	// Print results
	fmt.Println("  Successful Tasks:")
	for result := range results {
		fmt.Printf("    ✅ %s\n", result)
	}

	fmt.Println("  Failed Tasks:")
	for err := range errors {
		fmt.Printf("    ❌ %s (Pool: %s)\n", err.Error(), err.Details()["pool"])
	}

	// Show pool statistics
	fmt.Println("\n  Bulkhead Pool Statistics:")
	stats := bulkhead.GetPoolStatistics()
	for pool, stat := range stats {
		fmt.Printf("    %s Pool:\n", pool)
		fmt.Printf("      Max Concurrency: %d\n", stat.MaxConcurrency)
		fmt.Printf("      Current Active: %d\n", stat.CurrentActive)
		fmt.Printf("      Total Executed: %d\n", stat.TotalExecuted)
		fmt.Printf("      Total Rejected: %d\n", stat.TotalRejected)
		fmt.Printf("      Success Rate: %.2f%%\n", stat.SuccessRate*100)
	}
}

// correlationExample demonstrates correlation ID propagation
func correlationExample() {
	fmt.Println("\n🔗 Correlation ID Example:")

	correlator := NewCorrelationManager()

	// Simulate request with correlation ID
	correlationID := correlator.GenerateCorrelationID()
	fmt.Printf("  Generated Correlation ID: %s\n", correlationID)

	// Propagate through service calls
	fmt.Println("  Service Call Chain:")

	services := []string{"api-gateway", "user-service", "payment-service", "notification-service"}

	for i, service := range services {
		fmt.Printf("    %d. %s:\n", i+1, service)

		// Simulate processing time
		time.Sleep(50 * time.Millisecond)

		// Create service-specific error with correlation
		if service == "payment-service" {
			err := correlator.CreateCorrelatedError(correlationID, service, "PAYMENT_GATEWAY_ERROR", "Payment gateway is temporarily unavailable")

			fmt.Printf("       ❌ Error: %s\n", err.Error())
			fmt.Printf("       Correlation ID: %s\n", err.Details()["correlation_id"])
			fmt.Printf("       Service: %s\n", err.Details()["service"])
			fmt.Printf("       Timestamp: %s\n", err.Details()["timestamp"])
			fmt.Printf("       Request Path: %v\n", err.Details()["request_path"])

			// Show error correlation across services
			fmt.Println("\n  Error Correlation Analysis:")
			analysis := correlator.AnalyzeErrorCorrelation(correlationID)
			fmt.Printf("    Correlation ID: %s\n", analysis.CorrelationID)
			fmt.Printf("    Services Involved: %v\n", analysis.ServicesInvolved)
			fmt.Printf("    Error Service: %s\n", analysis.ErrorService)
			fmt.Printf("    Total Request Duration: %v\n", analysis.TotalDuration)
			fmt.Printf("    Request Success Rate: %.2f%%\n", analysis.SuccessRate*100)

			return
		} else {
			result := correlator.ProcessRequest(correlationID, service, fmt.Sprintf("%s_operation", service))
			fmt.Printf("       ✅ %s\n", result)
		}
	}
}

// Service Registry Implementation
type ServiceRegistry struct {
	services map[string]*ServiceInfo
	factory  interfaces.ErrorFactory
	mu       sync.RWMutex
}

type ServiceInfo struct {
	Name           string
	Version        string
	Host           string
	Port           int
	HealthEndpoint string
	Tags           []string
	Status         string
	LastHeartbeat  time.Time
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]*ServiceInfo),
		factory:  factory.GetDefaultFactory(),
	}
}

func (sr *ServiceRegistry) Register(service *ServiceInfo) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	service.Status = "healthy"
	service.LastHeartbeat = time.Now()
	sr.services[service.Name] = service
}

func (sr *ServiceRegistry) Discover(serviceName string) (*ServiceInfo, interfaces.DomainErrorInterface) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	service, exists := sr.services[serviceName]
	if !exists {
		return nil, sr.factory.Builder().
			WithCode("SERVICE_NOT_FOUND").
			WithMessage(fmt.Sprintf("Service '%s' not found in registry", serviceName)).
			WithType(string(types.ErrorTypeNotFound)).
			WithSeverity(interfaces.Severity(types.SeverityHigh)).
			WithDetail("service_name", serviceName).
			WithDetail("available_services", sr.getServiceNames()).
			WithTag("service_discovery").
			Build()
	}

	if service.Status != "healthy" {
		return nil, sr.factory.Builder().
			WithCode("SERVICE_UNHEALTHY").
			WithMessage(fmt.Sprintf("Service '%s' is not healthy", serviceName)).
			WithType(string(types.ErrorTypeExternalService)).
			WithSeverity(interfaces.Severity(types.SeverityHigh)).
			WithDetail("service_name", serviceName).
			WithDetail("service_status", service.Status).
			WithDetail("last_heartbeat", service.LastHeartbeat).
			WithTag("service_discovery").
			Build()
	}

	return service, nil
}

func (sr *ServiceRegistry) CallService(service *ServiceInfo, operation string, shouldFail bool) (string, interfaces.DomainErrorInterface) {
	if shouldFail {
		return "", sr.factory.Builder().
			WithCode("SERVICE_CALL_FAILED").
			WithMessage(fmt.Sprintf("Call to %s.%s failed", service.Name, operation)).
			WithType(string(types.ErrorTypeExternalService)).
			WithSeverity(interfaces.Severity(types.SeverityHigh)).
			WithDetail("service", service.Name).
			WithDetail("operation", operation).
			WithDetail("host", service.Host).
			WithDetail("port", service.Port).
			WithTag("service_call").
			Build()
	}

	return fmt.Sprintf("%s.%s executed successfully", service.Name, operation), nil
}

func (sr *ServiceRegistry) getServiceNames() []string {
	names := make([]string, 0, len(sr.services))
	for name := range sr.services {
		names = append(names, name)
	}
	return names
}

// Circuit Breaker Implementation
type CircuitBreaker struct {
	serviceName string
	config      CircuitBreakerConfig
	state       CircuitBreakerState
	stats       CircuitBreakerStats
	factory     interfaces.ErrorFactory
	mu          sync.RWMutex
}

type CircuitBreakerConfig struct {
	FailureThreshold int
	TimeoutDuration  time.Duration
	RecoveryTimeout  time.Duration
}

type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitBreakerClosed:
		return "CLOSED"
	case CircuitBreakerOpen:
		return "OPEN"
	case CircuitBreakerHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

type CircuitBreakerStats struct {
	TotalCalls   int
	SuccessCalls int
	FailureCalls int
	SuccessRate  float64
	LastFailure  time.Time
	StateChanged time.Time
}

func NewCircuitBreaker(serviceName string, config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		serviceName: serviceName,
		config:      config,
		state:       CircuitBreakerClosed,
		factory:     factory.GetDefaultFactory(),
		stats: CircuitBreakerStats{
			StateChanged: time.Now(),
		},
	}
}

func (cb *CircuitBreaker) Call(fn func() (interface{}, interfaces.DomainErrorInterface)) (interface{}, interfaces.DomainErrorInterface) {
	cb.mu.Lock()

	// Check if circuit breaker is open
	if cb.state == CircuitBreakerOpen {
		if time.Since(cb.stats.StateChanged) > cb.config.RecoveryTimeout {
			cb.state = CircuitBreakerHalfOpen
			cb.stats.StateChanged = time.Now()
		} else {
			cb.mu.Unlock()
			return nil, cb.factory.Builder().
				WithCode("CIRCUIT_BREAKER_OPEN").
				WithMessage(fmt.Sprintf("Circuit breaker is open for service %s", cb.serviceName)).
				WithType(string(types.ErrorTypeCircuitBreaker)).
				WithSeverity(interfaces.Severity(types.SeverityHigh)).
				WithDetail("service", cb.serviceName).
				WithDetail("state", cb.state.String()).
				WithDetail("time_until_retry", cb.config.RecoveryTimeout-time.Since(cb.stats.StateChanged)).
				WithTag("circuit_breaker").
				Build()
		}
	}

	cb.stats.TotalCalls++
	cb.mu.Unlock()

	// Execute the function
	result, err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.stats.FailureCalls++
		cb.stats.LastFailure = time.Now()

		// Check if we should open the circuit
		if cb.state == CircuitBreakerClosed && cb.stats.FailureCalls >= cb.config.FailureThreshold {
			cb.state = CircuitBreakerOpen
			cb.stats.StateChanged = time.Now()
		} else if cb.state == CircuitBreakerHalfOpen {
			// If half-open and we get a failure, go back to open
			cb.state = CircuitBreakerOpen
			cb.stats.StateChanged = time.Now()
		}
	} else {
		cb.stats.SuccessCalls++

		// If half-open and we get a success, close the circuit
		if cb.state == CircuitBreakerHalfOpen {
			cb.state = CircuitBreakerClosed
			cb.stats.StateChanged = time.Now()
			// Reset failure count
			cb.stats.FailureCalls = 0
		}
	}

	// Update success rate
	if cb.stats.TotalCalls > 0 {
		cb.stats.SuccessRate = float64(cb.stats.SuccessCalls) / float64(cb.stats.TotalCalls)
	}

	return result, err
}

func (cb *CircuitBreaker) GetState() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state.String()
}

func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.stats
}

// Distributed Tracing Implementation
type DistributedTracer struct {
	factory interfaces.ErrorFactory
	traces  map[string]*TraceInfo
	mu      sync.RWMutex
}

type TraceInfo struct {
	TraceID   string
	Spans     []*SpanInfo
	StartTime time.Time
	EndTime   time.Time
}

type SpanInfo struct {
	SpanID    string
	Service   string
	Operation string
	StartTime time.Time
	EndTime   time.Time
	Error     interfaces.DomainErrorInterface
	Parent    *SpanInfo
}

func NewDistributedTracer() *DistributedTracer {
	return &DistributedTracer{
		factory: factory.GetDefaultFactory(),
		traces:  make(map[string]*TraceInfo),
	}
}

func (dt *DistributedTracer) StartTrace(ctx context.Context, operation string) string {
	traceID := fmt.Sprintf("trace_%d", time.Now().UnixNano())

	dt.mu.Lock()
	dt.traces[traceID] = &TraceInfo{
		TraceID:   traceID,
		Spans:     make([]*SpanInfo, 0),
		StartTime: time.Now(),
	}
	dt.mu.Unlock()

	return traceID
}

func (dt *DistributedTracer) StartSpan(ctx context.Context, service, operation string) (context.Context, *SpanInfo) {
	span := &SpanInfo{
		SpanID:    fmt.Sprintf("span_%d", time.Now().UnixNano()),
		Service:   service,
		Operation: operation,
		StartTime: time.Now(),
	}

	// Add span to trace (simplified - in real implementation would extract from context)
	dt.mu.Lock()
	for _, trace := range dt.traces {
		trace.Spans = append(trace.Spans, span)
		break // Just add to first trace for demo
	}
	dt.mu.Unlock()

	return ctx, span
}

func (span *SpanInfo) Finish(err interfaces.DomainErrorInterface) {
	span.EndTime = time.Now()
	span.Error = err
}

func (dt *DistributedTracer) RecordError(ctx context.Context, err interfaces.DomainErrorInterface) {
	// In a real implementation, this would extract trace/span info from context
	// and add distributed tracing metadata to the error
}

func (dt *DistributedTracer) FinishTrace(ctx context.Context) {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	for _, trace := range dt.traces {
		trace.EndTime = time.Now()
		break // Just finish first trace for demo
	}
}

func simulateServiceCall(service, operation string, shouldFail bool) (string, interfaces.DomainErrorInterface) {
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

	if shouldFail {
		return "", factory.GetDefaultFactory().Builder().
			WithCode("SERVICE_ERROR").
			WithMessage(fmt.Sprintf("%s.%s failed", service, operation)).
			WithType(string(types.ErrorTypeExternalService)).
			WithSeverity(interfaces.Severity(types.SeverityHigh)).
			WithDetail("service", service).
			WithDetail("operation", operation).
			WithDetail("trace_id", "trace_123456").
			WithDetail("span_id", "span_789012").
			WithDetail("parent_spans", []string{"span_123", "span_456"}).
			WithTag("distributed").
			Build()
	}

	return fmt.Sprintf("%s.%s completed successfully", service, operation), nil
}

// Error Propagation Implementation
type ErrorPropagator struct {
	factory interfaces.ErrorFactory
}

type PropagationContext struct {
	Service   string
	Operation string
	Version   string
}

func NewErrorPropagator() *ErrorPropagator {
	return &ErrorPropagator{
		factory: factory.GetDefaultFactory(),
	}
}

func (ep *ErrorPropagator) PropagateError(originalErr interfaces.DomainErrorInterface, ctx PropagationContext) interfaces.DomainErrorInterface {
	// Get existing chain information
	chainLength := 1
	servicesInvolved := []string{ctx.Service}
	rootCauseCode := originalErr.Code()

	if existing, ok := originalErr.Details()["chain_length"].(int); ok {
		chainLength = existing + 1
	}

	if existing, ok := originalErr.Details()["services_involved"].([]string); ok {
		servicesInvolved = append(existing, ctx.Service)
	}

	if existing, ok := originalErr.Details()["root_cause_code"].(string); ok {
		rootCauseCode = existing
	}

	return ep.factory.Builder().
		WithCode(fmt.Sprintf("%s_PROPAGATED", ctx.Service)).
		WithMessage(fmt.Sprintf("Error propagated through %s: %s", ctx.Service, originalErr.Error())).
		WithType(originalErr.Type()).
		WithSeverity(originalErr.Severity()).
		WithDetail("service", ctx.Service).
		WithDetail("operation", ctx.Operation).
		WithDetail("version", ctx.Version).
		WithDetail("root_cause_code", rootCauseCode).
		WithDetail("chain_length", chainLength).
		WithDetail("services_involved", servicesInvolved).
		WithDetail("propagation_time", time.Now()).
		WithTag("propagated").
		Build()
}

// Service Mesh Implementation
type ServiceMesh struct {
	services        map[string]*MeshServiceInfo
	retryPolicies   map[string]RetryPolicy
	timeoutPolicies map[string]TimeoutPolicy
	statistics      map[string]*MeshStatistics
	factory         interfaces.ErrorFactory
	mu              sync.RWMutex
}

type MeshServiceInfo struct {
	Name   string
	Status string
}

type RetryPolicy struct {
	MaxRetries    int
	BackoffFactor float64
	InitialDelay  time.Duration
}

type TimeoutPolicy struct {
	RequestTimeout    time.Duration
	ConnectionTimeout time.Duration
}

type MeshStatistics struct {
	TotalRequests  int
	SuccessRate    float64
	AverageLatency time.Duration
	totalLatency   time.Duration
	successCount   int
}

func NewServiceMesh() *ServiceMesh {
	return &ServiceMesh{
		services:        make(map[string]*MeshServiceInfo),
		retryPolicies:   make(map[string]RetryPolicy),
		timeoutPolicies: make(map[string]TimeoutPolicy),
		statistics:      make(map[string]*MeshStatistics),
		factory:         factory.GetDefaultFactory(),
	}
}

func (sm *ServiceMesh) RegisterService(serviceName string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.services[serviceName] = &MeshServiceInfo{
		Name:   serviceName,
		Status: "healthy",
	}

	sm.statistics[serviceName] = &MeshStatistics{}
}

func (sm *ServiceMesh) SetRetryPolicy(serviceName string, policy RetryPolicy) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.retryPolicies[serviceName] = policy
}

func (sm *ServiceMesh) SetTimeoutPolicy(serviceName string, policy TimeoutPolicy) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.timeoutPolicies[serviceName] = policy
}

func (sm *ServiceMesh) RouteRequest(from, to, operation string, shouldFail bool) (string, interfaces.DomainErrorInterface) {
	start := time.Now()

	sm.mu.RLock()
	_, exists := sm.services[to]
	retryPolicy, hasRetryPolicy := sm.retryPolicies[to]
	sm.mu.RUnlock()

	if !exists {
		return "", sm.factory.Builder().
			WithCode("SERVICE_NOT_IN_MESH").
			WithMessage(fmt.Sprintf("Service '%s' is not registered in the mesh", to)).
			WithType(string(types.ErrorTypeNotFound)).
			WithSeverity(interfaces.Severity(types.SeverityHigh)).
			WithDetail("target_service", to).
			WithDetail("source_service", from).
			WithTag("service_mesh").
			Build()
	}

	var lastErr interfaces.DomainErrorInterface
	maxAttempts := 1
	if hasRetryPolicy {
		maxAttempts = retryPolicy.MaxRetries + 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			delay := time.Duration(float64(retryPolicy.InitialDelay) * float64(attempt) * retryPolicy.BackoffFactor)
			time.Sleep(delay)
		}

		// Simulate service call
		if !shouldFail || (shouldFail && attempt == maxAttempts-1) { // Succeed on last retry for demo
			shouldFail = false
		}

		result, err := simulateServiceCall(to, operation, shouldFail)

		if err == nil {
			duration := time.Since(start)
			sm.updateStatistics(to, true, duration)
			return result, nil
		}

		lastErr = err
	}

	// All retries failed
	duration := time.Since(start)
	sm.updateStatistics(to, false, duration)

	return "", sm.factory.Builder().
		WithCode("SERVICE_MESH_REQUEST_FAILED").
		WithMessage(fmt.Sprintf("Request from %s to %s.%s failed after %d attempts", from, to, operation, maxAttempts)).
		WithType(string(types.ErrorTypeExternalService)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithDetail("source_service", from).
		WithDetail("target_service", to).
		WithDetail("operation", operation).
		WithDetail("retry_attempts", maxAttempts-1).
		WithDetail("total_duration", duration).
		WithDetail("last_error", lastErr.Error()).
		WithTag("service_mesh").
		Build()
}

func (sm *ServiceMesh) updateStatistics(serviceName string, success bool, duration time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	stats := sm.statistics[serviceName]
	stats.TotalRequests++
	stats.totalLatency += duration
	stats.AverageLatency = stats.totalLatency / time.Duration(stats.TotalRequests)

	if success {
		stats.successCount++
	}

	stats.SuccessRate = float64(stats.successCount) / float64(stats.TotalRequests)
}

func (sm *ServiceMesh) GetStatistics() map[string]*MeshStatistics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string]*MeshStatistics)
	for k, v := range sm.statistics {
		result[k] = &MeshStatistics{
			TotalRequests:  v.TotalRequests,
			SuccessRate:    v.SuccessRate,
			AverageLatency: v.AverageLatency,
		}
	}

	return result
}

// Timeout Handler Implementation
type TimeoutHandler struct {
	factory interfaces.ErrorFactory
}

func NewTimeoutHandler() *TimeoutHandler {
	return &TimeoutHandler{
		factory: factory.GetDefaultFactory(),
	}
}

func (th *TimeoutHandler) CallWithTimeout(service, operation string, duration, timeout time.Duration) (string, interfaces.DomainErrorInterface) {
	done := make(chan struct{})
	var result string

	go func() {
		time.Sleep(duration)
		result = fmt.Sprintf("%s.%s completed", service, operation)
		close(done)
	}()

	select {
	case <-done:
		return result, nil
	case <-time.After(timeout):
		return "", th.factory.Builder().
			WithCode("SERVICE_TIMEOUT").
			WithMessage(fmt.Sprintf("Service %s.%s timed out after %v", service, operation, timeout)).
			WithType(string(types.ErrorTypeTimeout)).
			WithSeverity(interfaces.Severity(types.SeverityHigh)).
			WithDetail("service", service).
			WithDetail("operation", operation).
			WithDetail("timeout", timeout).
			WithDetail("expected_duration", duration).
			WithTag("timeout").
			Build()
	}
}

// Bulkhead Implementation
type Bulkhead struct {
	pools   map[string]*ResourcePool
	factory interfaces.ErrorFactory
	mu      sync.RWMutex
}

type BulkheadConfig struct {
	MaxConcurrency int
	QueueSize      int
	Timeout        time.Duration
}

type ResourcePool struct {
	config     BulkheadConfig
	semaphore  chan struct{}
	queue      chan *PoolTask
	statistics PoolStatistics
	mu         sync.RWMutex
}

type PoolTask struct {
	ID       string
	Function func() (interface{}, interfaces.DomainErrorInterface)
	Result   chan PoolResult
}

type PoolResult struct {
	Value interface{}
	Error interfaces.DomainErrorInterface
}

type PoolStatistics struct {
	MaxConcurrency int
	CurrentActive  int
	TotalExecuted  int
	TotalRejected  int
	SuccessRate    float64
	successCount   int
}

func NewBulkhead() *Bulkhead {
	return &Bulkhead{
		pools:   make(map[string]*ResourcePool),
		factory: factory.GetDefaultFactory(),
	}
}

func (b *Bulkhead) CreatePool(name string, config BulkheadConfig) {
	b.mu.Lock()
	defer b.mu.Unlock()

	pool := &ResourcePool{
		config:    config,
		semaphore: make(chan struct{}, config.MaxConcurrency),
		queue:     make(chan *PoolTask, config.QueueSize),
		statistics: PoolStatistics{
			MaxConcurrency: config.MaxConcurrency,
		},
	}

	// Start worker goroutines
	for i := 0; i < config.MaxConcurrency; i++ {
		go pool.worker()
	}

	b.pools[name] = pool
}

func (pool *ResourcePool) worker() {
	for task := range pool.queue {
		pool.mu.Lock()
		pool.statistics.CurrentActive++
		pool.mu.Unlock()

		result, err := task.Function()

		pool.mu.Lock()
		pool.statistics.CurrentActive--
		pool.statistics.TotalExecuted++
		if err == nil {
			pool.statistics.successCount++
		}
		pool.statistics.SuccessRate = float64(pool.statistics.successCount) / float64(pool.statistics.TotalExecuted)
		pool.mu.Unlock()

		task.Result <- PoolResult{Value: result, Error: err}
		close(task.Result)
	}
}

func (b *Bulkhead) Execute(poolName, taskID string, fn func() (interface{}, interfaces.DomainErrorInterface)) (interface{}, interfaces.DomainErrorInterface) {
	b.mu.RLock()
	pool, exists := b.pools[poolName]
	b.mu.RUnlock()

	if !exists {
		return nil, b.factory.Builder().
			WithCode("BULKHEAD_POOL_NOT_FOUND").
			WithMessage(fmt.Sprintf("Bulkhead pool '%s' not found", poolName)).
			WithType(string(types.ErrorTypeConfiguration)).
			WithSeverity(interfaces.Severity(types.SeverityHigh)).
			WithDetail("pool", poolName).
			WithTag("bulkhead").
			Build()
	}

	task := &PoolTask{
		ID:       taskID,
		Function: fn,
		Result:   make(chan PoolResult, 1),
	}

	select {
	case pool.queue <- task:
		// Task queued successfully
	default:
		// Queue is full
		pool.mu.Lock()
		pool.statistics.TotalRejected++
		pool.mu.Unlock()

		return nil, b.factory.Builder().
			WithCode("BULKHEAD_QUEUE_FULL").
			WithMessage(fmt.Sprintf("Bulkhead pool '%s' queue is full", poolName)).
			WithType(string(types.ErrorTypeRateLimit)).
			WithSeverity(interfaces.Severity(types.SeverityMedium)).
			WithDetail("pool", poolName).
			WithDetail("queue_size", pool.config.QueueSize).
			WithTag("bulkhead").
			Build()
	}

	select {
	case result := <-task.Result:
		return result.Value, result.Error
	case <-time.After(pool.config.Timeout):
		return nil, b.factory.Builder().
			WithCode("BULKHEAD_TASK_TIMEOUT").
			WithMessage(fmt.Sprintf("Task '%s' in pool '%s' timed out", taskID, poolName)).
			WithType(string(types.ErrorTypeTimeout)).
			WithSeverity(interfaces.Severity(types.SeverityHigh)).
			WithDetail("pool", poolName).
			WithDetail("task_id", taskID).
			WithDetail("timeout", pool.config.Timeout).
			WithTag("bulkhead").
			Build()
	}
}

func (b *Bulkhead) GetPoolStatistics() map[string]PoolStatistics {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make(map[string]PoolStatistics)
	for name, pool := range b.pools {
		pool.mu.RLock()
		result[name] = pool.statistics
		pool.mu.RUnlock()
	}

	return result
}

// Correlation Manager Implementation
type CorrelationManager struct {
	factory      interfaces.ErrorFactory
	mu           sync.RWMutex
	correlations map[string]*CorrelationInfo
}

type CorrelationInfo struct {
	CorrelationID    string
	StartTime        time.Time
	ServicesInvolved []string
	RequestPath      []string
	ErrorService     string
	TotalDuration    time.Duration
	SuccessRate      float64
}

func NewCorrelationManager() *CorrelationManager {
	return &CorrelationManager{
		factory:      factory.GetDefaultFactory(),
		correlations: make(map[string]*CorrelationInfo),
	}
}

func (cm *CorrelationManager) GenerateCorrelationID() string {
	return fmt.Sprintf("corr_%d_%s", time.Now().UnixNano(), generateRandomString(8))
}

func (cm *CorrelationManager) ProcessRequest(correlationID, service, operation string) string {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if info, exists := cm.correlations[correlationID]; exists {
		info.ServicesInvolved = append(info.ServicesInvolved, service)
		info.RequestPath = append(info.RequestPath, fmt.Sprintf("%s.%s", service, operation))
	} else {
		cm.correlations[correlationID] = &CorrelationInfo{
			CorrelationID:    correlationID,
			StartTime:        time.Now(),
			ServicesInvolved: []string{service},
			RequestPath:      []string{fmt.Sprintf("%s.%s", service, operation)},
		}
	}

	return fmt.Sprintf("%s.%s processed with correlation %s", service, operation, correlationID)
}

func (cm *CorrelationManager) CreateCorrelatedError(correlationID, service, code, message string) interfaces.DomainErrorInterface {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if info, exists := cm.correlations[correlationID]; exists {
		info.ErrorService = service
		info.TotalDuration = time.Since(info.StartTime)
		info.SuccessRate = 0.0 // Error occurred
	}

	return cm.factory.Builder().
		WithCode(code).
		WithMessage(message).
		WithType(string(types.ErrorTypeExternalService)).
		WithSeverity(interfaces.Severity(types.SeverityHigh)).
		WithDetail("correlation_id", correlationID).
		WithDetail("service", service).
		WithDetail("timestamp", time.Now()).
		WithDetail("request_path", cm.correlations[correlationID].RequestPath).
		WithTag("correlated").
		Build()
}

func (cm *CorrelationManager) AnalyzeErrorCorrelation(correlationID string) CorrelationInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if info, exists := cm.correlations[correlationID]; exists {
		return *info
	}

	return CorrelationInfo{CorrelationID: correlationID}
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
