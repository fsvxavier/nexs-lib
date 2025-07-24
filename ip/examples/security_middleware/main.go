package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/ip"
)

// SecurityMiddleware wraps HTTP handlers with IP security validation
func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate request security
		report := ip.ValidateSecurity(r)

		// Log security information
		logSecurityReport(r, report)

		// Check if request should be blocked
		if shouldBlockRequest(report) {
			handleSecurityViolation(w, r, report)
			return
		}

		// Add security headers to response
		addSecurityHeaders(w, report)

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// shouldBlockRequest determines if a request should be blocked based on security report
func shouldBlockRequest(report *ip.SecurityReport) bool {
	// Block critical threats
	if report.ThreatLevel >= ip.ThreatLevelCritical {
		return true
	}

	// Block requests with very low trust scores
	if report.TrustScore < 0.3 {
		return true
	}

	// Block specific violation types
	for _, violation := range report.Violations {
		switch violation.Type {
		case ip.ViolationIPSpoofing:
			if violation.Severity >= ip.ThreatLevelHigh {
				return true
			}
		case ip.ViolationMalformedHeader:
			if violation.Severity >= ip.ThreatLevelCritical {
				return true
			}
		}
	}

	return false
}

// handleSecurityViolation responds to blocked requests
func handleSecurityViolation(w http.ResponseWriter, r *http.Request, report *ip.SecurityReport) {
	w.Header().Set("Content-Type", "application/json")

	// Set appropriate status code based on threat level
	statusCode := http.StatusBadRequest
	switch report.ThreatLevel {
	case ip.ThreatLevelHigh:
		statusCode = http.StatusForbidden
	case ip.ThreatLevelCritical:
		statusCode = http.StatusTeapot // 418 for obvious attacks
	}

	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":        "Security validation failed",
		"threat_level": report.ThreatLevel.String(),
		"trust_score":  report.TrustScore,
		"request_id":   generateRequestID(),
		"violations":   len(report.Violations),
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// addSecurityHeaders adds security-related headers to the response
func addSecurityHeaders(w http.ResponseWriter, report *ip.SecurityReport) {
	// Add trust score header (for monitoring)
	w.Header().Set("X-IP-Trust-Score", fmt.Sprintf("%.2f", report.TrustScore))

	// Add threat level header
	w.Header().Set("X-Threat-Level", report.ThreatLevel.String())

	// Add security validation header
	if report.IsSuspicious {
		w.Header().Set("X-Security-Status", "suspicious")
	} else {
		w.Header().Set("X-Security-Status", "clean")
	}
}

// logSecurityReport logs security validation results
func logSecurityReport(r *http.Request, report *ip.SecurityReport) {
	clientIP := ip.GetRealIP(r)
	userAgent := r.Header.Get("User-Agent")

	logEntry := map[string]interface{}{
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"method":        r.Method,
		"path":          r.URL.Path,
		"client_ip":     clientIP,
		"user_agent":    userAgent,
		"threat_level":  report.ThreatLevel.String(),
		"trust_score":   report.TrustScore,
		"is_suspicious": report.IsSuspicious,
		"violations":    len(report.Violations),
	}

	// Log violations details if any
	if len(report.Violations) > 0 {
		violations := make([]map[string]interface{}, len(report.Violations))
		for i, v := range report.Violations {
			violations[i] = map[string]interface{}{
				"type":        v.Type.String(),
				"severity":    v.Severity.String(),
				"description": v.Description,
				"header":      v.Header,
			}
		}
		logEntry["violation_details"] = violations
	}

	// Log as JSON for structured logging
	logData, _ := json.Marshal(logEntry)

	if report.IsSuspicious {
		log.Printf("SECURITY ALERT: %s", string(logData))
	} else {
		log.Printf("SECURITY OK: %s", string(logData))
	}
}

// generateRequestID generates a unique request ID for tracking
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// Main handler for demonstration
func mainHandler(w http.ResponseWriter, r *http.Request) {
	// Get IP information
	realIP := ip.GetRealIP(r)
	ipInfo := ip.GetRealIPInfo(r)
	ipChain := ip.GetIPChain(r)
	trustScore := ip.GetTrustScore(r)

	response := map[string]interface{}{
		"message":     "Request processed successfully",
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"client_ip":   realIP,
		"ip_type":     ipInfo.Type.String(),
		"ip_chain":    ipChain,
		"trust_score": trustScore,
		"request_id":  generateRequestID(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Status handler to show security configuration
func statusHandler(w http.ResponseWriter, r *http.Request) {
	config := ip.GetSecurityConfig()

	response := map[string]interface{}{
		"security_config": map[string]interface{}{
			"spoofing_detection":     config.EnableSpoofingDetection,
			"enhanced_validation":    config.EnableEnhancedValidation,
			"input_sanitization":     config.EnableInputSanitization,
			"max_header_length":      config.MaxHeaderLength,
			"max_ip_chain_length":    config.MaxIPChainLength,
			"strict_ipv6_validation": config.StrictIPv6Validation,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Test endpoints for different security scenarios
func createTestHandlers() {
	// Endpoint that simulates clean requests
	http.HandleFunc("/api/clean", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"message":     "Clean endpoint accessed",
			"ip":          ip.GetRealIP(r),
			"trust_score": ip.GetTrustScore(r),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Endpoint that would trigger security alerts with suspicious headers
	http.HandleFunc("/api/test-security", func(w http.ResponseWriter, r *http.Request) {
		report := ip.ValidateSecurity(r)

		response := map[string]interface{}{
			"message":       "Security test endpoint",
			"threat_level":  report.ThreatLevel.String(),
			"trust_score":   report.TrustScore,
			"is_suspicious": report.IsSuspicious,
			"violations":    len(report.Violations),
		}

		if len(report.Violations) > 0 {
			violations := make([]string, len(report.Violations))
			for i, v := range report.Violations {
				violations[i] = v.Type.String()
			}
			response["violation_types"] = violations
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
}

func main() {
	fmt.Println("=== IP Security Middleware Demo ===")
	fmt.Println("Starting HTTP server with IP security validation...")

	// Configure custom security settings for demonstration
	customConfig := &ip.SecurityConfig{
		EnableSpoofingDetection:  true,
		EnableEnhancedValidation: true,
		EnableInputSanitization:  true,
		MaxHeaderLength:          1024,
		MaxIPChainLength:         10,
		StrictIPv6Validation:     true,
	}
	ip.SetSecurityConfig(customConfig)

	// Create router with security middleware
	mux := http.NewServeMux()

	// Add main handler with security middleware
	mux.Handle("/", SecurityMiddleware(http.HandlerFunc(mainHandler)))
	mux.Handle("/status", SecurityMiddleware(http.HandlerFunc(statusHandler)))

	// Create test endpoints
	createTestHandlers()
	mux.Handle("/api/", SecurityMiddleware(http.DefaultServeMux))

	// Start server
	port := ":8080"
	fmt.Printf("Server listening on %s\n", port)
	fmt.Println("\nEndpoints:")
	fmt.Println("  GET  /           - Main endpoint with IP analysis")
	fmt.Println("  GET  /status     - Security configuration status")
	fmt.Println("  GET  /api/clean  - Test clean requests")
	fmt.Println("  ANY  /api/test-security - Test security validation")

	fmt.Println("\nTest examples:")
	fmt.Println("  # Clean request:")
	fmt.Println("  curl http://localhost:8080/")

	fmt.Println("\n  # Request with suspicious headers:")
	fmt.Println("  curl -H 'X-Forwarded-For: 192.168.1.1' -H 'CF-Connecting-IP: 203.0.113.1' http://localhost:8080/api/test-security")

	fmt.Println("\n  # Request with malicious payload:")
	fmt.Println("  curl -H \"User-Agent: Mozilla' UNION SELECT * FROM users--\" http://localhost:8080/api/test-security")

	fmt.Println("\n  # Request with long header:")
	fmt.Printf("  curl -H 'X-Custom: %s' http://localhost:8080/api/test-security\n", generateLongString(2000)[:50]+"...")

	fmt.Println("\nMonitor the logs for security validation results...")

	// Auto-stop server after 3 seconds for demonstration
	fmt.Println("\n🚀 Server will stop automatically in 3 seconds for demonstration...")

	server := &http.Server{Addr: port, Handler: mux}
	go func() {
		time.Sleep(3 * time.Second)
		fmt.Println("✅ Security middleware example completed successfully - Server stopped automatically")
		server.Shutdown(context.Background())
	}()

	server.ListenAndServe()
}

func generateLongString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}
