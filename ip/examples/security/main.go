package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fsvxavier/nexs-lib/ip"
)

func main() {
	fmt.Println("=== IP Security Validation Examples ===")

	// Example 1: Basic security validation
	fmt.Println("1. Basic Security Validation:")
	demonstrateBasicSecurity()

	// Example 2: IP spoofing detection
	fmt.Println("\n2. IP Spoofing Detection:")
	demonstrateSpoofingDetection()

	// Example 3: Enhanced IPv6 validation
	fmt.Println("\n3. Enhanced IPv6 Validation:")
	demonstrateIPv6Validation()

	// Example 4: Input sanitization
	fmt.Println("\n4. Input Sanitization:")
	demonstrateInputSanitization()

	// Example 5: Custom security configuration
	fmt.Println("\n5. Custom Security Configuration:")
	demonstrateCustomConfig()

	// Example 6: Trust score calculation
	fmt.Println("\n6. Trust Score Calculation:")
	demonstrateTrustScore()
}

func demonstrateBasicSecurity() {
	// Create a clean request
	req := createMockRequest(map[string][]string{
		"X-Forwarded-For": {"203.0.113.1"},
		"User-Agent":      {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"},
	})

	// Validate security
	report := ip.ValidateSecurity(req)

	fmt.Printf("  Request Status: Valid=%v, Suspicious=%v\n", report.IsValid, report.IsSuspicious)
	fmt.Printf("  Threat Level: %s\n", report.ThreatLevel.String())
	fmt.Printf("  Trust Score: %.2f\n", report.TrustScore)
	fmt.Printf("  Violations: %d\n", len(report.Violations))

	// Check if request is secure
	isSecure := ip.IsRequestSecure(req)
	fmt.Printf("  Is Secure: %v\n", isSecure)
}

func demonstrateSpoofingDetection() {
	// Create a request with inconsistent headers (potential spoofing)
	req := createMockRequest(map[string][]string{
		"CF-Connecting-IP": {"203.0.113.1"},  // Trusted header
		"X-Forwarded-For":  {"198.51.100.1"}, // Different IP in standard header
		"X-Real-IP":        {"192.168.1.1"},  // Private IP in another header
	})

	report := ip.ValidateSecurity(req)

	fmt.Printf("  Spoofing Detection Results:\n")
	fmt.Printf("  - Suspicious: %v\n", report.IsSuspicious)
	fmt.Printf("  - Threat Level: %s\n", report.ThreatLevel.String())
	fmt.Printf("  - Trust Score: %.2f\n", report.TrustScore)

	fmt.Printf("  Violations Found:\n")
	for i, violation := range report.Violations {
		fmt.Printf("    %d. Type: %s\n", i+1, violation.Type.String())
		fmt.Printf("       Severity: %s\n", violation.Severity.String())
		fmt.Printf("       Description: %s\n", violation.Description)
		fmt.Printf("       Header: %s\n", violation.Header)
		fmt.Printf("       Suggestion: %s\n", violation.Suggestion)
		fmt.Println()
	}
}

func demonstrateIPv6Validation() {
	// Create requests with various IPv6 formats
	testCases := []struct {
		name string
		ipv6 string
	}{
		{"Valid IPv6", "2001:db8::1"},
		{"Valid IPv6 loopback", "::1"},
		{"Invalid format", "2001:db8::invalid::1"},
		{"IPv4-mapped with private IP", "::ffff:192.168.1.1"},
		{"Malformed", "not:an:ipv6:address"},
	}

	for _, tc := range testCases {
		req := createMockRequest(map[string][]string{
			"X-Forwarded-For": {tc.ipv6},
		})

		report := ip.ValidateSecurity(req)

		fmt.Printf("  %s (%s):\n", tc.name, tc.ipv6)
		fmt.Printf("    Valid: %v, Suspicious: %v\n", report.IsValid, report.IsSuspicious)
		fmt.Printf("    Violations: %d\n", len(report.Violations))

		for _, violation := range report.Violations {
			if violation.Type.String() == "invalid_ipv6" {
				fmt.Printf("    IPv6 Issue: %s\n", violation.Description)
			}
		}
		fmt.Println()
	}
}

func demonstrateInputSanitization() {
	// Create request with malicious input
	req := createMockRequest(map[string][]string{
		"X-Forwarded-For": {"203.0.113.1\x00\x01<script>alert('xss')</script>"},
		"User-Agent":      {"Mozilla/5.0' UNION SELECT * FROM users--"},
		"X-Real-IP":       {"192.168.1.1\r\n\r\nHTTP/1.1 200 OK"},
	})

	report := ip.ValidateSecurity(req)

	fmt.Printf("  Input Sanitization Results:\n")
	fmt.Printf("  - Violations found: %d\n", len(report.Violations))
	fmt.Printf("  - Headers sanitized: %d\n", len(report.SanitizedHeaders))

	fmt.Printf("  Original vs Sanitized Headers:\n")
	for header, sanitized := range report.SanitizedHeaders {
		original := req.Header.Get(header)
		fmt.Printf("    %s:\n", header)
		fmt.Printf("      Original: %q\n", original)
		fmt.Printf("      Sanitized: %q\n", sanitized)
	}

	fmt.Printf("  Security Violations:\n")
	for _, violation := range report.Violations {
		fmt.Printf("    - %s: %s\n", violation.Type.String(), violation.Description)
	}
}

func demonstrateCustomConfig() {
	// Save original config
	originalConfig := ip.GetSecurityConfig()

	// Create custom configuration
	customConfig := &ip.SecurityConfig{
		EnableSpoofingDetection:  true,
		EnableEnhancedValidation: true,
		EnableInputSanitization:  true,
		MaxHeaderLength:          100, // Very strict
		MaxIPChainLength:         3,   // Limit chain length
		StrictIPv6Validation:     true,
	}

	// Apply custom config
	ip.SetSecurityConfig(customConfig)

	// Test with long header
	req := createMockRequest(map[string][]string{
		"X-Forwarded-For": {"203.0.113.1, 198.51.100.1, 192.0.2.1, 203.0.113.2, 198.51.100.2"}, // Long chain
		"User-Agent":      {generateLongString(200)},                                           // Long header
	})

	report := ip.ValidateSecurity(req)

	fmt.Printf("  Custom Configuration Results:\n")
	fmt.Printf("  - Max Header Length: %d\n", customConfig.MaxHeaderLength)
	fmt.Printf("  - Max IP Chain Length: %d\n", customConfig.MaxIPChainLength)
	fmt.Printf("  - Violations: %d\n", len(report.Violations))

	for _, violation := range report.Violations {
		fmt.Printf("    - %s: %s\n", violation.Type.String(), violation.Description)
	}

	// Restore original config
	ip.SetSecurityConfig(originalConfig)
}

func demonstrateTrustScore() {
	// Test different scenarios and their trust scores
	scenarios := []struct {
		name    string
		headers map[string][]string
	}{
		{
			"Clean Request",
			map[string][]string{
				"X-Forwarded-For": {"203.0.113.1"},
				"User-Agent":      {"Mozilla/5.0"},
			},
		},
		{
			"Suspicious Headers",
			map[string][]string{
				"CF-Connecting-IP": {"192.168.1.1"}, // Private in public header
				"X-Real-IP":        {"203.0.113.1"},
			},
		},
		{
			"Malicious Input",
			map[string][]string{
				"X-Forwarded-For": {"203.0.113.1' OR 1=1--"},
				"User-Agent":      {"<script>alert('xss')</script>"},
			},
		},
		{
			"Multiple Issues",
			map[string][]string{
				"CF-Connecting-IP": {"192.168.1.1"},
				"X-Forwarded-For":  {"198.51.100.1' UNION SELECT"},
				"User-Agent":       {generateLongString(2000)},
			},
		},
	}

	fmt.Printf("  Trust Score Analysis:\n")
	for _, scenario := range scenarios {
		req := createMockRequest(scenario.headers)
		trustScore := ip.GetTrustScore(req)
		isSecure := ip.IsRequestSecure(req)

		fmt.Printf("    %s:\n", scenario.name)
		fmt.Printf("      Trust Score: %.2f\n", trustScore)
		fmt.Printf("      Is Secure: %v\n", isSecure)

		// Get detailed report
		report := ip.ValidateSecurity(req)
		fmt.Printf("      Threat Level: %s\n", report.ThreatLevel.String())
		fmt.Printf("      Violations: %d\n", len(report.Violations))
		fmt.Println()
	}
}

// Helper functions

func createMockRequest(headers map[string][]string) *http.Request {
	req := &http.Request{
		Method: "GET",
		Header: make(http.Header),
	}

	for key, values := range headers {
		req.Header[key] = values
	}

	return req
}

func generateLongString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}

func init() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
