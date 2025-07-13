package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
	"github.com/fsvxavier/nexs-lib/v2/observability/tracer/providers/datadog"
)

// Service represents a microservice with tracing capabilities
type Service struct {
	name    string
	port    int
	tracer  tracer.Tracer
	server  *http.Server
	clients map[string]*http.Client
}

// TraceMiddleware injects tracing into HTTP handlers
func (s *Service) traceMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := s.tracer.StartSpan(r.Context(), fmt.Sprintf("%s %s", r.Method, r.URL.Path),
			tracer.WithSpanKind(tracer.SpanKindServer),
			tracer.WithSpanAttributes(map[string]interface{}{
				"service.name":     s.name,
				"http.method":      r.Method,
				"http.url":         r.URL.String(),
				"http.user_agent":  r.UserAgent(),
				"http.remote_addr": r.RemoteAddr,
			}),
		)
		defer span.End()

		// Inject span context into request
		r = r.WithContext(ctx)

		// Add request start event
		span.AddEvent("request.started", map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"service":   s.name,
		})

		// Call next handler
		next(w, r)

		// Add completion event
		span.AddEvent("request.completed", map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"service":   s.name,
		})

		span.SetStatus(tracer.StatusCodeOk, "Request processed successfully")
	}
}

// httpClient makes traced HTTP requests to other services
func (s *Service) httpClient(ctx context.Context, method, url string) (*http.Response, error) {
	_, span := s.tracer.StartSpan(ctx, fmt.Sprintf("http.client %s", method),
		tracer.WithSpanKind(tracer.SpanKindClient),
		tracer.WithSpanAttributes(map[string]interface{}{
			"http.method": method,
			"http.url":    url,
			"client.name": s.name,
		}),
	)
	defer span.End()

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Failed to create request: %v", err))
		span.RecordError(err, map[string]interface{}{
			"error.type": "request_creation_error",
		})
		return nil, err
	}

	// Add tracing headers (simplified - in real implementation you'd use proper propagation)
	spanCtx := span.Context()
	req.Header.Set("X-Trace-ID", spanCtx.TraceID)
	req.Header.Set("X-Span-ID", spanCtx.SpanID)

	// Make request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Request failed: %v", err))
		span.RecordError(err, map[string]interface{}{
			"error.type": "network_error",
		})
		return nil, err
	}

	span.SetAttribute("http.status_code", resp.StatusCode)
	if resp.StatusCode >= 400 {
		span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("HTTP error: %d", resp.StatusCode))
	} else {
		span.SetStatus(tracer.StatusCodeOk, "Request successful")
	}

	return resp, nil
}

// Gateway Service - Entry point for all requests
func createGatewayService(tr tracer.Tracer) *Service {
	s := &Service{
		name:   "gateway",
		port:   8081,
		tracer: tr,
	}

	mux := http.NewServeMux()

	// User routes
	mux.HandleFunc("/api/user/", s.traceMiddleware(func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Path[len("/api/user/"):]

		ctx, span := s.tracer.StartSpan(r.Context(), "gateway.get_user",
			tracer.WithSpanKind(tracer.SpanKindInternal),
			tracer.WithSpanAttributes(map[string]interface{}{
				"user.id": userID,
			}),
		)
		defer span.End()

		// Authenticate user
		if !s.authenticateUser(ctx, r) {
			span.SetStatus(tracer.StatusCodeError, "Authentication failed")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Forward to user service
		resp, err := s.httpClient(ctx, "GET", fmt.Sprintf("http://localhost:8083/user/%s", userID))
		if err != nil {
			span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("User service error: %v", err))
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		w.WriteHeader(resp.StatusCode)
		fmt.Fprintf(w, `{"gateway": "success", "user_id": "%s", "status": %d}`, userID, resp.StatusCode)
	}))

	// Order routes
	mux.HandleFunc("/api/orders/", s.traceMiddleware(func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Path[len("/api/orders/"):]

		ctx, span := s.tracer.StartSpan(r.Context(), "gateway.get_order",
			tracer.WithSpanKind(tracer.SpanKindInternal),
			tracer.WithSpanAttributes(map[string]interface{}{
				"order.id": orderID,
			}),
		)
		defer span.End()

		// Forward to order service
		resp, err := s.httpClient(ctx, "GET", fmt.Sprintf("http://localhost:8084/order/%s", orderID))
		if err != nil {
			span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Order service error: %v", err))
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		w.WriteHeader(resp.StatusCode)
		fmt.Fprintf(w, `{"gateway": "success", "order_id": "%s", "status": %d}`, orderID, resp.StatusCode)
	}))

	// Health check
	mux.HandleFunc("/health", s.traceMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status": "healthy", "service": "gateway"}`)
	}))

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	return s
}

// Auth Service - Authentication and authorization
func createAuthService(tr tracer.Tracer) *Service {
	s := &Service{
		name:   "auth",
		port:   8082,
		tracer: tr,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/validate", s.traceMiddleware(func(w http.ResponseWriter, r *http.Request) {
		_, span := s.tracer.StartSpan(r.Context(), "auth.validate_token",
			tracer.WithSpanKind(tracer.SpanKindInternal),
		)
		defer span.End()

		token := r.Header.Get("Authorization")
		span.SetAttribute("auth.has_token", token != "")

		// Simulate token validation
		time.Sleep(50 * time.Millisecond) // Simulate auth processing

		if token == "" {
			span.SetStatus(tracer.StatusCodeError, "Missing authorization token")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		span.SetStatus(tracer.StatusCodeOk, "Token validated")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"valid": true, "user_id": "123"}`)
	}))

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	return s
}

// User Service - User profile management
func createUserService(tr tracer.Tracer) *Service {
	s := &Service{
		name:   "user",
		port:   8083,
		tracer: tr,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/user/", s.traceMiddleware(func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Path[len("/user/"):]

		ctx, span := s.tracer.StartSpan(r.Context(), "user.get_profile",
			tracer.WithSpanKind(tracer.SpanKindInternal),
			tracer.WithSpanAttributes(map[string]interface{}{
				"user.id": userID,
			}),
		)
		defer span.End()

		// Simulate database query
		s.simulateDBQuery(ctx, "users", fmt.Sprintf("SELECT * FROM users WHERE id = %s", userID))

		span.SetStatus(tracer.StatusCodeOk, "User profile retrieved")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"user_id": "%s", "name": "John Doe", "email": "john@example.com"}`, userID)
	}))

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	return s
}

// Order Service - Order processing
func createOrderService(tr tracer.Tracer) *Service {
	s := &Service{
		name:   "order",
		port:   8084,
		tracer: tr,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/order/", s.traceMiddleware(func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Path[len("/order/"):]

		ctx, span := s.tracer.StartSpan(r.Context(), "order.get_details",
			tracer.WithSpanKind(tracer.SpanKindInternal),
			tracer.WithSpanAttributes(map[string]interface{}{
				"order.id": orderID,
			}),
		)
		defer span.End()

		// Simulate database query
		s.simulateDBQuery(ctx, "orders", fmt.Sprintf("SELECT * FROM orders WHERE id = %s", orderID))

		// Simulate payment service call
		_, err := s.httpClient(ctx, "GET", fmt.Sprintf("http://localhost:8085/payment/order/%s", orderID))
		if err != nil {
			log.Printf("Payment service unavailable: %v", err)
		}

		span.SetStatus(tracer.StatusCodeOk, "Order details retrieved")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"order_id": "%s", "status": "completed", "amount": 99.99}`, orderID)
	}))

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	return s
}

// Payment Service - Payment processing
func createPaymentService(tr tracer.Tracer) *Service {
	s := &Service{
		name:   "payment",
		port:   8085,
		tracer: tr,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/payment/order/", s.traceMiddleware(func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Path[len("/payment/order/"):]

		ctx, span := s.tracer.StartSpan(r.Context(), "payment.get_status",
			tracer.WithSpanKind(tracer.SpanKindInternal),
			tracer.WithSpanAttributes(map[string]interface{}{
				"order.id": orderID,
			}),
		)
		defer span.End()

		// Simulate external payment gateway call
		s.simulateExternalCall(ctx, "payment-gateway", "GET", "/api/transaction/status")

		span.SetStatus(tracer.StatusCodeOk, "Payment status retrieved")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"order_id": "%s", "payment_status": "paid", "transaction_id": "txn_123"}`, orderID)
	}))

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	return s
}

// Helper methods for services
func (s *Service) authenticateUser(ctx context.Context, r *http.Request) bool {
	_, span := s.tracer.StartSpan(ctx, "auth.check",
		tracer.WithSpanKind(tracer.SpanKindClient),
	)
	defer span.End()

	// Simulate auth service call
	time.Sleep(25 * time.Millisecond)
	span.SetStatus(tracer.StatusCodeOk, "Authentication successful")
	return true
}

func (s *Service) simulateDBQuery(ctx context.Context, table, query string) {
	_, span := s.tracer.StartSpan(ctx, "db.query",
		tracer.WithSpanKind(tracer.SpanKindClient),
		tracer.WithSpanAttributes(map[string]interface{}{
			"db.system":    "postgresql",
			"db.table":     table,
			"db.operation": "SELECT",
			"db.query":     query,
		}),
	)
	defer span.End()

	// Simulate database latency
	time.Sleep(100 * time.Millisecond)
	span.SetStatus(tracer.StatusCodeOk, "Query executed successfully")
}

func (s *Service) simulateExternalCall(ctx context.Context, service, method, endpoint string) {
	_, span := s.tracer.StartSpan(ctx, fmt.Sprintf("external.%s", service),
		tracer.WithSpanKind(tracer.SpanKindClient),
		tracer.WithSpanAttributes(map[string]interface{}{
			"external.service": service,
			"http.method":      method,
			"http.endpoint":    endpoint,
		}),
	)
	defer span.End()

	// Simulate external service latency
	time.Sleep(200 * time.Millisecond)
	span.SetStatus(tracer.StatusCodeOk, "External call successful")
}

func (s *Service) start() error {
	log.Printf("Starting %s service on port %d", s.name, s.port)
	return s.server.ListenAndServe()
}

func (s *Service) shutdown(ctx context.Context) error {
	log.Printf("Shutting down %s service", s.name)
	return s.server.Shutdown(ctx)
}

func main() {
	// Configure Datadog provider
	config := &datadog.Config{
		ServiceName:        "microservices-example",
		ServiceVersion:     "1.0.0",
		Environment:        "development",
		AgentHost:          "localhost",
		AgentPort:          8126,
		SampleRate:         1.0,
		EnableProfiling:    true,
		RuntimeMetrics:     true,
		AnalyticsEnabled:   true,
		Debug:              true,
		MaxTracesPerSecond: 1000,
		Tags: map[string]string{
			"example": "microservices",
			"version": "v2",
		},
	}

	// Create provider
	provider, err := datadog.NewProvider(config)
	if err != nil {
		log.Fatalf("Failed to create Datadog provider: %v", err)
	}

	// Create tracer
	tr, err := provider.CreateTracer("microservices",
		tracer.WithServiceName("microservices-example"),
		tracer.WithEnvironment("development"),
	)
	if err != nil {
		log.Fatalf("Failed to create tracer: %v", err)
	}

	// Create all services
	services := []*Service{
		createGatewayService(tr),
		createAuthService(tr),
		createUserService(tr),
		createOrderService(tr),
		createPaymentService(tr),
	}

	// Start all services
	var wg sync.WaitGroup
	for _, service := range services {
		wg.Add(1)
		go func(s *Service) {
			defer wg.Done()
			if err := s.start(); err != nil && err != http.ErrServerClosed {
				log.Printf("Service %s error: %v", s.name, err)
			}
		}(service)
	}

	// Give services time to start
	time.Sleep(2 * time.Second)

	fmt.Println("Microservices example started!")
	fmt.Println("Gateway running on :8081")
	fmt.Println("")
	fmt.Println("Try these commands:")
	fmt.Println("  curl http://localhost:8081/api/user/123")
	fmt.Println("  curl http://localhost:8081/api/orders/456")
	fmt.Println("  curl http://localhost:8081/health")
	fmt.Println("")
	fmt.Println("Press Ctrl+C to stop...")

	// Wait for shutdown signal
	// In a real application, you'd handle graceful shutdown here
	wg.Wait()

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, service := range services {
		if err := service.shutdown(ctx); err != nil {
			log.Printf("Error shutting down %s: %v", service.name, err)
		}
	}

	if err := provider.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down provider: %v", err)
	}
}
