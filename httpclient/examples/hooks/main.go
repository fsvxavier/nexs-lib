package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient"
	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
)

// MetricsHook tracks request metrics and performance
type MetricsHook struct {
	requests     int
	errors       int
	totalTime    time.Duration
	slowRequests int
	startTimes   map[string]time.Time
}

func NewMetricsHook() *MetricsHook {
	return &MetricsHook{
		startTimes: make(map[string]time.Time),
	}
}

func (h *MetricsHook) BeforeRequest(ctx context.Context, req *interfaces.Request) error {
	h.requests++
	requestID := req.Headers["X-Request-ID"]
	if requestID == "" {
		requestID = fmt.Sprintf("req-%d", time.Now().UnixNano())
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["X-Request-ID"] = requestID
	}

	h.startTimes[requestID] = time.Now()
	fmt.Printf("📊 [METRICS] Starting request %s: %s %s\n", requestID, req.Method, req.URL)
	return nil
}

func (h *MetricsHook) AfterResponse(ctx context.Context, req *interfaces.Request, resp *interfaces.Response) error {
	requestID := req.Headers["X-Request-ID"]
	if startTime, exists := h.startTimes[requestID]; exists {
		duration := time.Since(startTime)
		h.totalTime += duration

		if duration > 2*time.Second {
			h.slowRequests++
		}

		delete(h.startTimes, requestID)

		if resp != nil {
			fmt.Printf("📊 [METRICS] Completed request %s: %d (took %v)\n", requestID, resp.StatusCode, duration)
		}
	}
	return nil
}

func (h *MetricsHook) OnError(ctx context.Context, req *interfaces.Request, err error) error {
	h.errors++
	requestID := req.Headers["X-Request-ID"]
	if startTime, exists := h.startTimes[requestID]; exists {
		duration := time.Since(startTime)
		delete(h.startTimes, requestID)
		fmt.Printf("📊 [METRICS] Error in request %s: %v (took %v)\n", requestID, err, duration)
	}
	return nil
}

func (h *MetricsHook) PrintStats() {
	avgTime := time.Duration(0)
	if h.requests > 0 {
		avgTime = h.totalTime / time.Duration(h.requests)
	}

	fmt.Printf("\n📈 Request Metrics Summary:\n")
	fmt.Printf("  Total Requests: %d\n", h.requests)
	fmt.Printf("  Errors: %d\n", h.errors)
	fmt.Printf("  Success Rate: %.1f%%\n", float64(h.requests-h.errors)/float64(h.requests)*100)
	fmt.Printf("  Average Response Time: %v\n", avgTime)
	fmt.Printf("  Slow Requests (>2s): %d\n", h.slowRequests)
}

// SecurityHook adds security headers and validates responses
type SecurityHook struct {
	allowedOrigins []string
}

func NewSecurityHook(allowedOrigins []string) *SecurityHook {
	return &SecurityHook{allowedOrigins: allowedOrigins}
}

func (h *SecurityHook) BeforeRequest(ctx context.Context, req *interfaces.Request) error {
	// Add security headers
	if req.Headers == nil {
		req.Headers = make(map[string]string)
	}

	req.Headers["X-Content-Type-Options"] = "nosniff"
	req.Headers["X-Frame-Options"] = "DENY"
	req.Headers["X-XSS-Protection"] = "1; mode=block"
	req.Headers["User-Agent"] = "SecureHTTPClient/1.0"

	fmt.Printf("🔒 [SECURITY] Added security headers to request\n")
	return nil
}

func (h *SecurityHook) AfterResponse(ctx context.Context, req *interfaces.Request, resp *interfaces.Response) error {
	if resp == nil {
		return nil
	}

	// Check for security headers in response
	securityHeaders := []string{
		"X-Frame-Options",
		"X-Content-Type-Options",
		"Strict-Transport-Security",
	}

	missingHeaders := []string{}
	for _, header := range securityHeaders {
		if _, exists := resp.Headers[header]; !exists {
			missingHeaders = append(missingHeaders, header)
		}
	}

	if len(missingHeaders) > 0 {
		fmt.Printf("⚠️  [SECURITY] Missing security headers: %v\n", missingHeaders)
	} else {
		fmt.Printf("🔒 [SECURITY] All security headers present\n")
	}

	return nil
}

func (h *SecurityHook) OnError(ctx context.Context, req *interfaces.Request, err error) error {
	fmt.Printf("🔒 [SECURITY] Request failed securely: %v\n", err)
	return nil
}

// AuditHook logs all requests for compliance
type AuditHook struct {
	logger *log.Logger
}

func NewAuditHook(logger *log.Logger) *AuditHook {
	return &AuditHook{logger: logger}
}

func (h *AuditHook) BeforeRequest(ctx context.Context, req *interfaces.Request) error {
	h.logger.Printf("AUDIT_REQUEST: method=%s url=%s user_agent=%s",
		req.Method, req.URL, req.Headers["User-Agent"])
	return nil
}

func (h *AuditHook) AfterResponse(ctx context.Context, req *interfaces.Request, resp *interfaces.Response) error {
	status := "unknown"
	if resp != nil {
		status = fmt.Sprintf("%d", resp.StatusCode)
	}
	h.logger.Printf("AUDIT_RESPONSE: method=%s url=%s status=%s",
		req.Method, req.URL, status)
	return nil
}

func (h *AuditHook) OnError(ctx context.Context, req *interfaces.Request, err error) error {
	h.logger.Printf("AUDIT_ERROR: method=%s url=%s error=%v",
		req.Method, req.URL, err)
	return nil
}

func main() {
	fmt.Printf("🪝 Hooks Example\n")
	fmt.Printf("================\n\n")

	// Create HTTP client
	client, err := httpclient.New(interfaces.ProviderNetHTTP, "https://httpbin.org")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create hooks
	metricsHook := NewMetricsHook()
	securityHook := NewSecurityHook([]string{"https://httpbin.org"})
	auditLogger := log.New(log.Writer(), "[AUDIT] ", log.LstdFlags)
	auditHook := NewAuditHook(auditLogger)

	// Add hooks to client
	client.
		AddHook(metricsHook).
		AddHook(securityHook).
		AddHook(auditHook)

	fmt.Println("📋 Hooks added:")
	fmt.Println("  1. Metrics Hook (tracks performance)")
	fmt.Println("  2. Security Hook (adds security headers)")
	fmt.Println("  3. Audit Hook (logs for compliance)")
	fmt.Println()

	ctx := context.Background()

	// Demonstrate different request types
	fmt.Println("1️⃣ Making GET request...")
	resp1, err := client.Get(ctx, "/get")
	if err != nil {
		log.Printf("GET request failed: %v", err)
	} else {
		fmt.Printf("   ✅ GET Status: %d\n", resp1.StatusCode)
	}
	fmt.Println()

	// POST request with data
	fmt.Println("2️⃣ Making POST request...")
	postData := map[string]interface{}{
		"message":   "Hello from hooks example!",
		"timestamp": time.Now().Unix(),
		"data": map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	resp2, err := client.Post(ctx, "/post", postData)
	if err != nil {
		log.Printf("POST request failed: %v", err)
	} else {
		fmt.Printf("   ✅ POST Status: %d\n", resp2.StatusCode)
	}
	fmt.Println()

	// Request that will be slow (delay endpoint)
	fmt.Println("3️⃣ Making slow request (3 second delay)...")
	resp3, err := client.Get(ctx, "/delay/3")
	if err != nil {
		log.Printf("Slow request failed: %v", err)
	} else {
		fmt.Printf("   ✅ Slow request Status: %d\n", resp3.StatusCode)
	}
	fmt.Println()

	// Request that will fail (404)
	fmt.Println("4️⃣ Making request that will fail...")
	resp4, err := client.Get(ctx, "/status/404")
	if err != nil {
		log.Printf("Request failed as expected: %v", err)
	} else {
		fmt.Printf("   ⚠️ Failed request Status: %d\n", resp4.StatusCode)
	}
	fmt.Println()

	// Multiple rapid requests to test metrics
	fmt.Println("5️⃣ Making multiple rapid requests...")
	for i := 1; i <= 3; i++ {
		go func(requestNum int) {
			resp, err := client.Get(ctx, fmt.Sprintf("/get?rapid=%d", requestNum))
			if err != nil {
				log.Printf("Rapid request %d failed: %v", requestNum, err)
			} else {
				fmt.Printf("   Rapid request %d: Status %d\n", requestNum, resp.StatusCode)
			}
		}(i)
	}

	// Wait for rapid requests to complete
	time.Sleep(2 * time.Second)
	fmt.Println()

	// Print metrics summary
	metricsHook.PrintStats()

	// Demonstrate hook removal
	fmt.Println("\n6️⃣ Removing security hook and making another request...")
	client.RemoveHook(securityHook)

	resp6, err := client.Get(ctx, "/get?no-security-hook=true")
	if err != nil {
		log.Printf("Request without security hook failed: %v", err)
	} else {
		fmt.Printf("   ✅ Status without security hook: %d\n", resp6.StatusCode)
	}

	fmt.Println("\n🎉 Hooks example completed!")
	fmt.Println("\n💡 Key Features Demonstrated:")
	fmt.Println("  • Request lifecycle hooks (before/after/error)")
	fmt.Println("  • Performance metrics collection")
	fmt.Println("  • Security header injection and validation")
	fmt.Println("  • Audit logging for compliance")
	fmt.Println("  • Hook removal and dynamic behavior")
	fmt.Println("  • Concurrent request tracking")
}
