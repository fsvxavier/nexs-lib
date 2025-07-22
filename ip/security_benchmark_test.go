package ip

import (
	"net/http"
	"strings"
	"testing"
)

// Benchmark security validation functions
func BenchmarkValidateSecurity_Clean(b *testing.B) {
	req := &http.Request{
		Header: http.Header{
			"X-Forwarded-For": []string{"203.0.113.1"},
			"User-Agent":      []string{"Mozilla/5.0"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateSecurity(req)
	}
}

func BenchmarkValidateSecurity_Suspicious(b *testing.B) {
	req := &http.Request{
		Header: http.Header{
			"CF-Connecting-IP": []string{"203.0.113.1"},
			"X-Forwarded-For":  []string{"198.51.100.1"},
			"X-Real-IP":        []string{"192.168.1.1"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateSecurity(req)
	}
}

func BenchmarkValidateSecurity_Malicious(b *testing.B) {
	req := &http.Request{
		Header: http.Header{
			"X-Forwarded-For": []string{"203.0.113.1' UNION SELECT * FROM users--"},
			"User-Agent":      []string{"<script>alert('xss')</script>Mozilla/5.0"},
			"X-Real-IP":       []string{"192.168.1.1\x00\x01\x02"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateSecurity(req)
	}
}

func BenchmarkIsRequestSecure(b *testing.B) {
	req := &http.Request{
		Header: http.Header{
			"X-Forwarded-For": []string{"203.0.113.1"},
			"User-Agent":      []string{"Mozilla/5.0"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsRequestSecure(req)
	}
}

func BenchmarkGetTrustScore(b *testing.B) {
	req := &http.Request{
		Header: http.Header{
			"X-Forwarded-For": []string{"203.0.113.1"},
			"User-Agent":      []string{"Mozilla/5.0"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetTrustScore(req)
	}
}

func BenchmarkSecurityValidator_IPv6Validation(b *testing.B) {
	validator := NewSecurityValidator(DefaultSecurityConfig())
	ipChain := []string{
		"2001:db8::1",
		"::ffff:192.168.1.1",
		"2001:db8::invalid::format",
		"fe80::1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		report := &SecurityReport{
			IsValid:          true,
			IsSuspicious:     false,
			ThreatLevel:      ThreatLevelNone,
			Violations:       make([]SecurityViolation, 0),
			SanitizedHeaders: make(map[string]string),
			TrustScore:       1.0,
		}
		validator.validateIPv6Addresses(ipChain, report)
	}
}

func BenchmarkSecurityValidator_HeaderValidation(b *testing.B) {
	validator := NewSecurityValidator(DefaultSecurityConfig())
	headers := map[string]string{
		"X-Forwarded-For":  "203.0.113.1, 198.51.100.1",
		"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		"X-Real-IP":        "203.0.113.1",
		"CF-Connecting-IP": "203.0.113.1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		report := &SecurityReport{
			IsValid:          true,
			IsSuspicious:     false,
			ThreatLevel:      ThreatLevelNone,
			Violations:       make([]SecurityViolation, 0),
			SanitizedHeaders: make(map[string]string),
			TrustScore:       1.0,
		}
		validator.validateHeaders(headers, report)
	}
}

func BenchmarkSecurityValidator_SpoofingDetection(b *testing.B) {
	validator := NewSecurityValidator(DefaultSecurityConfig())
	headers := map[string]string{
		"CF-Connecting-IP": "203.0.113.1",
		"X-Forwarded-For":  "198.51.100.1",
		"X-Real-IP":        "192.168.1.1",
	}
	ipChain := []string{"203.0.113.1", "198.51.100.1", "192.168.1.1"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		report := &SecurityReport{
			IsValid:          true,
			IsSuspicious:     false,
			ThreatLevel:      ThreatLevelNone,
			Violations:       make([]SecurityViolation, 0),
			SanitizedHeaders: make(map[string]string),
			TrustScore:       1.0,
		}
		validator.detectSpoofing(headers, ipChain, report)
	}
}

func BenchmarkSecurityValidator_InputSanitization(b *testing.B) {
	validator := NewSecurityValidator(DefaultSecurityConfig())
	headers := map[string]string{
		"X-Forwarded-For": "203.0.113.1\x00\x01<script>alert(1)</script>",
		"User-Agent":      "Mozilla/5.0' UNION SELECT * FROM users--",
		"X-Real-IP":       "192.168.1.1\r\n\r\nHTTP/1.1 200 OK",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		report := &SecurityReport{
			IsValid:          true,
			IsSuspicious:     false,
			ThreatLevel:      ThreatLevelNone,
			Violations:       make([]SecurityViolation, 0),
			SanitizedHeaders: make(map[string]string),
			TrustScore:       1.0,
		}
		validator.sanitizeHeaders(headers, report)
	}
}

func BenchmarkSecurityValidator_Full(b *testing.B) {
	validator := NewSecurityValidator(DefaultSecurityConfig())
	adapter := &mockRequestAdapter{
		headers: map[string]string{
			"X-Forwarded-For":  "203.0.113.1, 198.51.100.1",
			"CF-Connecting-IP": "203.0.113.1",
			"User-Agent":       "Mozilla/5.0",
			"X-Real-IP":        "203.0.113.1",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateSecurityOptimized(adapter)
	}
}

// Benchmark comparison: Basic IP extraction vs Security validation
func BenchmarkGetRealIP_vs_ValidateSecurity(b *testing.B) {
	req := &http.Request{
		Header: http.Header{
			"X-Forwarded-For": []string{"203.0.113.1"},
			"User-Agent":      []string{"Mozilla/5.0"},
		},
	}

	b.Run("GetRealIP", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = GetRealIP(req)
		}
	})

	b.Run("ValidateSecurity", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ValidateSecurity(req)
		}
	})
}

// Benchmark different configuration scenarios
func BenchmarkSecurityConfig_Impact(b *testing.B) {
	req := &http.Request{
		Header: http.Header{
			"X-Forwarded-For": []string{"203.0.113.1"},
			"User-Agent":      []string{"Mozilla/5.0"},
		},
	}

	// Minimal config
	minimalConfig := &SecurityConfig{
		EnableSpoofingDetection:  false,
		EnableEnhancedValidation: false,
		EnableInputSanitization:  false,
		MaxHeaderLength:          1024,
		MaxIPChainLength:         10,
		StrictIPv6Validation:     false,
	}

	// Full config
	fullConfig := DefaultSecurityConfig()

	b.Run("MinimalSecurity", func(b *testing.B) {
		SetSecurityConfig(minimalConfig)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ValidateSecurity(req)
		}
	})

	b.Run("FullSecurity", func(b *testing.B) {
		SetSecurityConfig(fullConfig)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ValidateSecurity(req)
		}
	})

	// Restore default
	SetSecurityConfig(fullConfig)
}

// Benchmark with different request sizes
func BenchmarkSecurity_RequestSizes(b *testing.B) {
	// Small request
	smallReq := &http.Request{
		Header: http.Header{
			"X-Forwarded-For": []string{"203.0.113.1"},
		},
	}

	// Medium request
	mediumReq := &http.Request{
		Header: http.Header{
			"X-Forwarded-For": []string{"203.0.113.1, 198.51.100.1, 192.168.1.1"},
			"User-Agent":      []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"},
			"X-Real-IP":       []string{"203.0.113.1"},
		},
	}

	// Large request (many headers)
	largeHeaders := make(http.Header)
	largeHeaders.Set("X-Forwarded-For", strings.Repeat("203.0.113.1, ", 20)+"203.0.113.1")
	largeHeaders.Set("User-Agent", strings.Repeat("a", 500))
	for i := 0; i < 10; i++ {
		largeHeaders.Set(string(rune('A'+i))+"-Custom-Header", strings.Repeat("value", 10))
	}
	largeReq := &http.Request{Header: largeHeaders}

	b.Run("SmallRequest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ValidateSecurity(smallReq)
		}
	})

	b.Run("MediumRequest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ValidateSecurity(mediumReq)
		}
	})

	b.Run("LargeRequest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ValidateSecurity(largeReq)
		}
	})
}
