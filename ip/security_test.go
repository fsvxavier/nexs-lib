package ip

import (
	"net/http"
	"strings"
	"testing"

	"github.com/fsvxavier/nexs-lib/ip/interfaces"
)

// Mock request adapter for testing
type mockRequestAdapter struct {
	headers map[string]string
}

func (m *mockRequestAdapter) GetHeader(key string) string {
	return m.headers[key]
}

func (m *mockRequestAdapter) GetAllHeaders() map[string]string {
	return m.headers
}

func (m *mockRequestAdapter) GetRemoteAddr() string {
	return "192.168.1.100:8080"
}

func (m *mockRequestAdapter) GetMethod() string {
	return "GET"
}

func (m *mockRequestAdapter) GetPath() string {
	return "/"
}

func newMockAdapter(headers map[string]string) *mockRequestAdapter {
	return &mockRequestAdapter{headers: headers}
}

func TestSecurityValidator_NewSecurityValidator(t *testing.T) {
	tests := []struct {
		name   string
		config *SecurityConfig
	}{
		{
			name:   "default config",
			config: nil,
		},
		{
			name: "custom config",
			config: &SecurityConfig{
				EnableSpoofingDetection:  false,
				EnableEnhancedValidation: true,
				EnableInputSanitization:  false,
				MaxHeaderLength:          512,
				MaxIPChainLength:         5,
				StrictIPv6Validation:     false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewSecurityValidator(tt.config)
			if validator == nil {
				t.Error("NewSecurityValidator() returned nil")
			}
			if validator.config == nil {
				t.Error("validator config is nil")
			}
		})
	}
}

func TestSecurityValidator_ValidateHeaders(t *testing.T) {
	validator := NewSecurityValidator(DefaultSecurityConfig())

	tests := []struct {
		name               string
		headers            map[string]string
		expectedViolations int
		expectedSuspicious bool
	}{
		{
			name: "clean headers",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1",
				"User-Agent":      "Mozilla/5.0",
			},
			expectedViolations: 0,
			expectedSuspicious: false,
		},
		{
			name: "header too long",
			headers: map[string]string{
				"X-Forwarded-For": strings.Repeat("a", 2000),
			},
			expectedViolations: 1,
			expectedSuspicious: true,
		},
		{
			name: "suspicious characters",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1<script>alert(1)</script>",
			},
			expectedViolations: 2, // Both suspicious chars and script pattern
			expectedSuspicious: true,
		},
		{
			name: "SQL injection attempt",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.1' UNION SELECT * FROM users--",
			},
			expectedViolations: 2, // Both suspicious chars and SQL pattern
			expectedSuspicious: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &SecurityReport{
				IsValid:          true,
				IsSuspicious:     false,
				ThreatLevel:      ThreatLevelNone,
				Violations:       make([]SecurityViolation, 0),
				SanitizedHeaders: make(map[string]string),
				TrustScore:       1.0,
			}

			validator.validateHeaders(tt.headers, report)

			if len(report.Violations) != tt.expectedViolations {
				t.Errorf("validateHeaders() violations = %d, want %d", len(report.Violations), tt.expectedViolations)
			}

			if report.IsSuspicious != tt.expectedSuspicious {
				t.Errorf("validateHeaders() suspicious = %v, want %v", report.IsSuspicious, tt.expectedSuspicious)
			}
		})
	}
}

func TestSecurityValidator_DetectSpoofing(t *testing.T) {
	validator := NewSecurityValidator(DefaultSecurityConfig())

	tests := []struct {
		name               string
		headers            map[string]string
		ipChain            []string
		expectedViolations int
		expectedSuspicious bool
	}{
		{
			name: "consistent headers",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.1",
				"X-Forwarded-For":  "203.0.113.1",
			},
			ipChain:            []string{"203.0.113.1"},
			expectedViolations: 0,
			expectedSuspicious: false,
		},
		{
			name: "inconsistent headers",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.1",
				"X-Forwarded-For":  "198.51.100.1",
			},
			ipChain:            []string{"203.0.113.1", "198.51.100.1"},
			expectedViolations: 1,
			expectedSuspicious: true,
		},
		{
			name: "suspicious chain pattern",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 203.0.113.1, 10.0.0.1",
			},
			ipChain:            []string{"192.168.1.1", "203.0.113.1", "10.0.0.1"},
			expectedViolations: 1,
			expectedSuspicious: true,
		},
		{
			name: "multiple loopback IPs",
			headers: map[string]string{
				"X-Real-IP": "127.0.0.1",
			},
			ipChain:            []string{"127.0.0.1", "127.0.0.2"},
			expectedViolations: 1,
			expectedSuspicious: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &SecurityReport{
				IsValid:          true,
				IsSuspicious:     false,
				ThreatLevel:      ThreatLevelNone,
				Violations:       make([]SecurityViolation, 0),
				SanitizedHeaders: make(map[string]string),
				TrustScore:       1.0,
			}

			validator.detectSpoofing(tt.headers, tt.ipChain, report)

			if len(report.Violations) != tt.expectedViolations {
				t.Errorf("detectSpoofing() violations = %d, want %d", len(report.Violations), tt.expectedViolations)
			}

			if report.IsSuspicious != tt.expectedSuspicious {
				t.Errorf("detectSpoofing() suspicious = %v, want %v", report.IsSuspicious, tt.expectedSuspicious)
			}
		})
	}
}

func TestSecurityValidator_ValidateIPv6Addresses(t *testing.T) {
	validator := NewSecurityValidator(DefaultSecurityConfig())

	tests := []struct {
		name               string
		ipChain            []string
		expectedViolations int
		expectedSuspicious bool
	}{
		{
			name:               "valid IPv6",
			ipChain:            []string{"2001:db8::1", "fe80::1"},
			expectedViolations: 0,
			expectedSuspicious: false,
		},
		{
			name:               "invalid IPv6 format",
			ipChain:            []string{"2001:db8::invalid::1"},
			expectedViolations: 1,
			expectedSuspicious: true,
		},
		{
			name:               "malformed IPv6",
			ipChain:            []string{"2001:db8:invalid:format"},
			expectedViolations: 1,
			expectedSuspicious: true,
		},
		{
			name:               "IPv4-mapped IPv6 with private IP",
			ipChain:            []string{"::ffff:192.168.1.1"},
			expectedViolations: 1,
			expectedSuspicious: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &SecurityReport{
				IsValid:          true,
				IsSuspicious:     false,
				ThreatLevel:      ThreatLevelNone,
				Violations:       make([]SecurityViolation, 0),
				SanitizedHeaders: make(map[string]string),
				TrustScore:       1.0,
			}

			validator.validateIPv6Addresses(tt.ipChain, report)

			if len(report.Violations) != tt.expectedViolations {
				t.Errorf("validateIPv6Addresses() violations = %d, want %d", len(report.Violations), tt.expectedViolations)
			}

			if report.IsSuspicious != tt.expectedSuspicious {
				t.Errorf("validateIPv6Addresses() suspicious = %v, want %v", report.IsSuspicious, tt.expectedSuspicious)
			}
		})
	}
}

func TestSecurityValidator_DetectPrivateInPublic(t *testing.T) {
	validator := NewSecurityValidator(DefaultSecurityConfig())

	tests := []struct {
		name               string
		headers            map[string]string
		expectedViolations int
		expectedSuspicious bool
	}{
		{
			name: "public IP in public header",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.1",
			},
			expectedViolations: 0,
			expectedSuspicious: false,
		},
		{
			name: "private IP in public header",
			headers: map[string]string{
				"CF-Connecting-IP": "192.168.1.1",
			},
			expectedViolations: 1,
			expectedSuspicious: true,
		},
		{
			name: "private IP in standard header (allowed)",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1",
			},
			expectedViolations: 0,
			expectedSuspicious: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &SecurityReport{
				IsValid:          true,
				IsSuspicious:     false,
				ThreatLevel:      ThreatLevelNone,
				Violations:       make([]SecurityViolation, 0),
				SanitizedHeaders: make(map[string]string),
				TrustScore:       1.0,
			}

			validator.detectPrivateInPublic(tt.headers, report)

			if len(report.Violations) != tt.expectedViolations {
				t.Errorf("detectPrivateInPublic() violations = %d, want %d", len(report.Violations), tt.expectedViolations)
			}

			if report.IsSuspicious != tt.expectedSuspicious {
				t.Errorf("detectPrivateInPublic() suspicious = %v, want %v", report.IsSuspicious, tt.expectedSuspicious)
			}
		})
	}
}

func TestSecurityValidator_SanitizeHeaders(t *testing.T) {
	validator := NewSecurityValidator(DefaultSecurityConfig())

	tests := []struct {
		name              string
		headers           map[string]string
		expectedSanitized map[string]string
	}{
		{
			name: "clean headers",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1",
			},
			expectedSanitized: map[string]string{},
		},
		{
			name: "headers with control characters",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.1\x00\x01\x02",
			},
			expectedSanitized: map[string]string{
				"X-Real-IP": "203.0.113.1",
			},
		},
		{
			name: "headers with SQL injection",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0 UNION SELECT * FROM users",
			},
			expectedSanitized: map[string]string{
				"User-Agent": "Mozilla/5.0   * FROM users", // Triple space due to regex replacement + space normalization
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &SecurityReport{
				IsValid:          true,
				IsSuspicious:     false,
				ThreatLevel:      ThreatLevelNone,
				Violations:       make([]SecurityViolation, 0),
				SanitizedHeaders: make(map[string]string),
				TrustScore:       1.0,
			}

			validator.sanitizeHeaders(tt.headers, report)

			for key, expectedValue := range tt.expectedSanitized {
				if sanitizedValue, exists := report.SanitizedHeaders[key]; !exists || sanitizedValue != expectedValue {
					t.Errorf("sanitizeHeaders() key %s = %s, want %s", key, sanitizedValue, expectedValue)
				}
			}
		})
	}
}

func TestSecurityValidator_CalculateTrustScore(t *testing.T) {
	validator := NewSecurityValidator(DefaultSecurityConfig())

	tests := []struct {
		name                string
		violations          []SecurityViolation
		expectedThreatLevel ThreatLevel
		expectedSuspicious  bool
		minTrustScore       float64
		maxTrustScore       float64
	}{
		{
			name:                "no violations",
			violations:          []SecurityViolation{},
			expectedThreatLevel: ThreatLevelNone,
			expectedSuspicious:  false,
			minTrustScore:       1.0,
			maxTrustScore:       1.0,
		},
		{
			name: "low severity violation",
			violations: []SecurityViolation{
				{Severity: ThreatLevelLow},
			},
			expectedThreatLevel: ThreatLevelLow,
			expectedSuspicious:  false,
			minTrustScore:       0.8,
			maxTrustScore:       1.0,
		},
		{
			name: "high severity violation",
			violations: []SecurityViolation{
				{Severity: ThreatLevelHigh},
			},
			expectedThreatLevel: ThreatLevelHigh,
			expectedSuspicious:  true,
			minTrustScore:       0.0,
			maxTrustScore:       0.7,
		},
		{
			name: "critical violation",
			violations: []SecurityViolation{
				{Severity: ThreatLevelCritical},
			},
			expectedThreatLevel: ThreatLevelCritical,
			expectedSuspicious:  true,
			minTrustScore:       0.0,
			maxTrustScore:       0.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &SecurityReport{
				IsValid:          true,
				IsSuspicious:     false,
				ThreatLevel:      ThreatLevelNone,
				Violations:       tt.violations,
				SanitizedHeaders: make(map[string]string),
				TrustScore:       1.0,
			}

			validator.calculateTrustScore(report)

			if report.ThreatLevel != tt.expectedThreatLevel {
				t.Errorf("calculateTrustScore() threat level = %v, want %v", report.ThreatLevel, tt.expectedThreatLevel)
			}

			if report.IsSuspicious != tt.expectedSuspicious {
				t.Errorf("calculateTrustScore() suspicious = %v, want %v", report.IsSuspicious, tt.expectedSuspicious)
			}

			if report.TrustScore < tt.minTrustScore || report.TrustScore > tt.maxTrustScore {
				t.Errorf("calculateTrustScore() trust score = %f, want between %f and %f",
					report.TrustScore, tt.minTrustScore, tt.maxTrustScore)
			}
		})
	}
}

func TestSecurityValidator_ValidateSecurityOptimized(t *testing.T) {
	validator := NewSecurityValidator(DefaultSecurityConfig())

	tests := []struct {
		name               string
		adapter            interfaces.RequestAdapter
		expectedValid      bool
		expectedSuspicious bool
	}{
		{
			name: "clean request",
			adapter: newMockAdapter(map[string]string{
				"X-Forwarded-For": "203.0.113.1",
				"User-Agent":      "Mozilla/5.0",
			}),
			expectedValid:      true,
			expectedSuspicious: false,
		},
		{
			name: "suspicious request",
			adapter: newMockAdapter(map[string]string{
				"CF-Connecting-IP": "192.168.1.1", // Private IP in public header
				"X-Real-IP":        "203.0.113.1",
			}),
			expectedValid:      true,
			expectedSuspicious: true,
		},
		{
			name:               "nil adapter",
			adapter:            nil,
			expectedValid:      true,
			expectedSuspicious: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := validator.ValidateSecurityOptimized(tt.adapter)

			if report.IsValid != tt.expectedValid {
				t.Errorf("ValidateSecurityOptimized() valid = %v, want %v", report.IsValid, tt.expectedValid)
			}

			if report.IsSuspicious != tt.expectedSuspicious {
				t.Errorf("ValidateSecurityOptimized() suspicious = %v, want %v", report.IsSuspicious, tt.expectedSuspicious)
			}
		})
	}
}

func TestValidateSecurity(t *testing.T) {
	tests := []struct {
		name          string
		request       interface{}
		expectedValid bool
		expectError   bool
	}{
		{
			name: "valid http request",
			request: &http.Request{
				Header: http.Header{
					"X-Forwarded-For": []string{"203.0.113.1"},
				},
			},
			expectedValid: true,
			expectError:   false,
		},
		{
			name:          "invalid request type",
			request:       "invalid",
			expectedValid: false,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := ValidateSecurity(tt.request)

			if tt.expectError {
				if report.IsValid {
					t.Error("ValidateSecurity() expected error but got valid result")
				}
			} else {
				if report.IsValid != tt.expectedValid {
					t.Errorf("ValidateSecurity() valid = %v, want %v", report.IsValid, tt.expectedValid)
				}
			}
		})
	}
}

func TestIsRequestSecure(t *testing.T) {
	tests := []struct {
		name     string
		request  interface{}
		expected bool
	}{
		{
			name: "secure request",
			request: &http.Request{
				Header: http.Header{
					"X-Forwarded-For": []string{"203.0.113.1"},
				},
			},
			expected: true,
		},
		{
			name:     "invalid request",
			request:  "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRequestSecure(tt.request)
			if result != tt.expected {
				t.Errorf("IsRequestSecure() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetTrustScore(t *testing.T) {
	tests := []struct {
		name     string
		request  interface{}
		minScore float64
		maxScore float64
	}{
		{
			name: "clean request",
			request: &http.Request{
				Header: http.Header{
					"X-Forwarded-For": []string{"203.0.113.1"},
				},
			},
			minScore: 0.8,
			maxScore: 1.0,
		},
		{
			name:     "invalid request",
			request:  "invalid",
			minScore: 0.0,
			maxScore: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := GetTrustScore(tt.request)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("GetTrustScore() = %f, want between %f and %f", score, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestSecurityConfig(t *testing.T) {
	config := DefaultSecurityConfig()

	if !config.EnableSpoofingDetection {
		t.Error("Default config should enable spoofing detection")
	}

	if !config.EnableEnhancedValidation {
		t.Error("Default config should enable enhanced validation")
	}

	if !config.EnableInputSanitization {
		t.Error("Default config should enable input sanitization")
	}

	if config.MaxHeaderLength != 1024 {
		t.Errorf("Default max header length = %d, want 1024", config.MaxHeaderLength)
	}

	if config.MaxIPChainLength != 10 {
		t.Errorf("Default max IP chain length = %d, want 10", config.MaxIPChainLength)
	}
}

func TestThreatLevel_String(t *testing.T) {
	tests := []struct {
		level    ThreatLevel
		expected string
	}{
		{ThreatLevelNone, "none"},
		{ThreatLevelLow, "low"},
		{ThreatLevelMedium, "medium"},
		{ThreatLevelHigh, "high"},
		{ThreatLevelCritical, "critical"},
		{ThreatLevel(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("ThreatLevel.String() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestViolationType_String(t *testing.T) {
	tests := []struct {
		violation ViolationType
		expected  string
	}{
		{ViolationIPSpoofing, "ip_spoofing"},
		{ViolationPrivateInPublic, "private_in_public"},
		{ViolationMalformedHeader, "malformed_header"},
		{ViolationInvalidIPv6, "invalid_ipv6"},
		{ViolationSuspiciousChain, "suspicious_chain"},
		{ViolationHeaderTooLong, "header_too_long"},
		{ViolationChainTooLong, "chain_too_long"},
		{ViolationInconsistentHeaders, "inconsistent_headers"},
		{ViolationType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.violation.String()
			if result != tt.expected {
				t.Errorf("ViolationType.String() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestSetSecurityConfig(t *testing.T) {
	originalConfig := GetSecurityConfig()

	newConfig := &SecurityConfig{
		EnableSpoofingDetection:  false,
		EnableEnhancedValidation: false,
		EnableInputSanitization:  false,
		MaxHeaderLength:          512,
		MaxIPChainLength:         5,
		StrictIPv6Validation:     false,
	}

	SetSecurityConfig(newConfig)

	currentConfig := GetSecurityConfig()
	if currentConfig.EnableSpoofingDetection != false {
		t.Error("SetSecurityConfig() did not update spoofing detection setting")
	}

	if currentConfig.MaxHeaderLength != 512 {
		t.Error("SetSecurityConfig() did not update max header length")
	}

	// Restore original config
	SetSecurityConfig(originalConfig)
}
