package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/tracer"
	"github.com/fsvxavier/nexs-lib/v2/observability/tracer/providers/datadog"
)

func main() {
	// Configure Datadog provider
	config := &datadog.Config{
		ServiceName:        "example-datadog-service",
		ServiceVersion:     "1.0.0",
		Environment:        "development",
		AgentHost:          "localhost",
		AgentPort:          8126,
		SampleRate:         1.0, // 100% sampling for development
		EnableProfiling:    true,
		RuntimeMetrics:     true,
		AnalyticsEnabled:   true,
		PrioritySampling:   true,
		Debug:              true,
		MaxTracesPerSecond: 1000,
		Tags: map[string]string{
			"team":    "backend",
			"service": "api",
			"region":  "us-east-1",
		},
		ObfuscationEnabled: true,
		ObfuscatedTags:     []string{"password", "token", "api_key"},
	}

	// Create provider
	provider, err := datadog.NewProvider(config)
	if err != nil {
		log.Fatalf("Failed to create Datadog provider: %v", err)
	}

	// Ensure proper shutdown
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := provider.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down provider: %v", err)
		}
	}()

	// Create tracer
	tr, err := provider.CreateTracer("http-server",
		tracer.WithServiceName("example-service"),
		tracer.WithEnvironment("development"),
	)
	if err != nil {
		log.Fatalf("Failed to create tracer: %v", err)
	}

	// Setup HTTP server with tracing
	http.HandleFunc("/", handleRoot(tr))
	http.HandleFunc("/users", handleUsers(tr))
	http.HandleFunc("/health", handleHealth(tr))

	fmt.Println("Starting server on :8080...")
	fmt.Println("Try:")
	fmt.Println("  curl http://localhost:8080/")
	fmt.Println("  curl http://localhost:8080/users")
	fmt.Println("  curl http://localhost:8080/health")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleRoot demonstrates basic span creation and attributes
func handleRoot(tr tracer.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Start root span
		ctx, span := tr.StartSpan(r.Context(), "http.request",
			tracer.WithSpanKind(tracer.SpanKindServer),
			tracer.WithSpanAttributes(map[string]interface{}{
				"http.method":      r.Method,
				"http.url":         r.URL.String(),
				"http.user_agent":  r.UserAgent(),
				"http.remote_addr": r.RemoteAddr,
			}),
		)
		defer span.End()

		// Add event for request start
		span.AddEvent("request.started", map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"path":      r.URL.Path,
		})

		// Simulate some processing
		processRequest(ctx, tr)

		// Set response attributes
		span.SetAttribute("http.status_code", 200)
		span.SetAttribute("response.size", 13)

		// Add event for request completion
		span.AddEvent("request.completed", map[string]interface{}{
			"timestamp":   time.Now().Unix(),
			"status_code": 200,
		})

		// Set successful status
		span.SetStatus(tracer.StatusCodeOk, "Request processed successfully")

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Hello, World!")
	}
}

// handleUsers demonstrates database simulation and error handling
func handleUsers(tr tracer.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Start root span
		ctx, span := tr.StartSpan(r.Context(), "get_users",
			tracer.WithSpanKind(tracer.SpanKindServer),
			tracer.WithSpanAttributes(map[string]interface{}{
				"http.method": r.Method,
				"http.route":  "/users",
			}),
		)
		defer span.End()

		// Simulate authentication check
		if !authenticateUser(ctx, tr, r) {
			span.SetStatus(tracer.StatusCodeError, "Authentication failed")
			span.SetAttribute("error", true)
			span.SetAttribute("error.type", "authentication_error")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Simulate database query
		users, err := getUsersFromDB(ctx, tr)
		if err != nil {
			span.SetStatus(tracer.StatusCodeError, fmt.Sprintf("Database error: %v", err))
			span.SetAttribute("error", true)
			span.SetAttribute("error.type", "database_error")
			span.SetAttribute("error.message", err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set success attributes
		span.SetAttribute("users.count", len(users))
		span.SetAttribute("http.status_code", 200)
		span.SetStatus(tracer.StatusCodeOk, "Users retrieved successfully")

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"users": %d, "data": %v}`, len(users), users)
	}
}

// handleHealth demonstrates health check tracing
func handleHealth(tr tracer.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tr.StartSpan(r.Context(), "health_check",
			tracer.WithSpanKind(tracer.SpanKindServer),
		)
		defer span.End()

		// Check various components
		healthy := true
		components := map[string]bool{
			"database": checkDatabase(ctx, tr),
			"cache":    checkCache(ctx, tr),
			"external": checkExternalService(ctx, tr),
		}

		for component, status := range components {
			span.SetAttribute(fmt.Sprintf("health.%s", component), status)
			if !status {
				healthy = false
			}
		}

		if healthy {
			span.SetStatus(tracer.StatusCodeOk, "All systems healthy")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"status": "healthy", "components": {"database": true, "cache": true, "external": true}}`)
		} else {
			span.SetStatus(tracer.StatusCodeError, "Some components unhealthy")
			span.SetAttribute("error", true)
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, `{"status": "unhealthy", "components": {"database": false, "cache": false, "external": false}}`)
		}
	}
}

// processRequest simulates request processing with child spans
func processRequest(ctx context.Context, tr tracer.Tracer) {
	_, span := tr.StartSpan(ctx, "process_request",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Simulate validation
	validateInput(ctx, tr)

	// Simulate business logic
	executeBusinessLogic(ctx, tr)

	span.SetStatus(tracer.StatusCodeOk, "Request processed")
}

func validateInput(ctx context.Context, tr tracer.Tracer) {
	_, span := tr.StartSpan(ctx, "validate_input",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Simulate validation time
	time.Sleep(10 * time.Millisecond)

	span.SetAttribute("validation.rules", 3)
	span.SetAttribute("validation.passed", true)
	span.SetStatus(tracer.StatusCodeOk, "Input valid")
}

func executeBusinessLogic(ctx context.Context, tr tracer.Tracer) {
	_, span := tr.StartSpan(ctx, "business_logic",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Simulate processing time
	time.Sleep(50 * time.Millisecond)

	span.SetAttribute("logic.complexity", "medium")
	span.SetAttribute("logic.duration_ms", 50)
	span.SetStatus(tracer.StatusCodeOk, "Logic executed")
}

func authenticateUser(ctx context.Context, tr tracer.Tracer, r *http.Request) bool {
	_, span := tr.StartSpan(ctx, "authenticate_user",
		tracer.WithSpanKind(tracer.SpanKindInternal),
	)
	defer span.End()

	// Simulate auth check
	time.Sleep(20 * time.Millisecond)

	// Check for auth header (simplified)
	authHeader := r.Header.Get("Authorization")
	authenticated := authHeader != ""

	span.SetAttribute("auth.method", "bearer_token")
	span.SetAttribute("auth.user_id", "12345")
	span.SetAttribute("auth.authenticated", authenticated)

	if authenticated {
		span.SetStatus(tracer.StatusCodeOk, "User authenticated")
	} else {
		span.SetStatus(tracer.StatusCodeError, "Authentication failed")
	}

	return authenticated
}

func getUsersFromDB(ctx context.Context, tr tracer.Tracer) ([]string, error) {
	_, span := tr.StartSpan(ctx, "db.query",
		tracer.WithSpanKind(tracer.SpanKindClient),
		tracer.WithSpanAttributes(map[string]interface{}{
			"db.system":    "postgresql",
			"db.name":      "users_db",
			"db.operation": "SELECT",
			"db.table":     "users",
		}),
	)
	defer span.End()

	// Simulate database query time
	time.Sleep(100 * time.Millisecond)

	// Simulate occasional errors
	if time.Now().UnixNano()%10 == 0 {
		err := fmt.Errorf("database connection timeout")
		span.SetStatus(tracer.StatusCodeError, err.Error())
		span.SetAttribute("error", true)
		span.SetAttribute("error.type", "timeout")
		return nil, err
	}

	users := []string{"alice", "bob", "charlie", "diana"}
	span.SetAttribute("db.rows_affected", len(users))
	span.SetAttribute("db.query_duration_ms", 100)
	span.SetStatus(tracer.StatusCodeOk, "Query successful")

	return users, nil
}

func checkDatabase(ctx context.Context, tr tracer.Tracer) bool {
	_, span := tr.StartSpan(ctx, "health.database",
		tracer.WithSpanKind(tracer.SpanKindClient),
	)
	defer span.End()

	time.Sleep(10 * time.Millisecond)
	healthy := true // Simulate always healthy for demo

	span.SetAttribute("db.ping_duration_ms", 10)
	span.SetStatus(tracer.StatusCodeOk, "Database healthy")
	return healthy
}

func checkCache(ctx context.Context, tr tracer.Tracer) bool {
	_, span := tr.StartSpan(ctx, "health.cache",
		tracer.WithSpanKind(tracer.SpanKindClient),
	)
	defer span.End()

	time.Sleep(5 * time.Millisecond)
	healthy := true

	span.SetAttribute("cache.ping_duration_ms", 5)
	span.SetStatus(tracer.StatusCodeOk, "Cache healthy")
	return healthy
}

func checkExternalService(ctx context.Context, tr tracer.Tracer) bool {
	_, span := tr.StartSpan(ctx, "health.external_service",
		tracer.WithSpanKind(tracer.SpanKindClient),
	)
	defer span.End()

	time.Sleep(30 * time.Millisecond)
	healthy := true

	span.SetAttribute("external.service", "payment_api")
	span.SetAttribute("external.ping_duration_ms", 30)
	span.SetStatus(tracer.StatusCodeOk, "External service healthy")
	return healthy
}
