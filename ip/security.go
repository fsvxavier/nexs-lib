// Package ip security provides advanced security validations for IP extraction.
// This file contains IP spoofing detection, enhanced validation, and input sanitization
// to protect against malicious attempts and ensure data integrity.
package ip

import (
	"net"
	"regexp"
	"strings"
	"sync"

	"github.com/fsvxavier/nexs-lib/ip/interfaces"
	"github.com/fsvxavier/nexs-lib/ip/providers"
)

// SecurityConfig holds configuration for security validations
type SecurityConfig struct {
	// EnableSpoofingDetection enables IP spoofing detection
	EnableSpoofingDetection bool
	// EnableEnhancedValidation enables advanced IPv6 and header validation
	EnableEnhancedValidation bool
	// EnableInputSanitization enables automatic input sanitization
	EnableInputSanitization bool
	// MaxHeaderLength maximum allowed header length (default: 1024)
	MaxHeaderLength int
	// MaxIPChainLength maximum allowed IP chain length (default: 10)
	MaxIPChainLength int
	// StrictIPv6Validation enables strict IPv6 format validation
	StrictIPv6Validation bool
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		EnableSpoofingDetection:  true,
		EnableEnhancedValidation: true,
		EnableInputSanitization:  true,
		MaxHeaderLength:          1024,
		MaxIPChainLength:         10,
		StrictIPv6Validation:     true,
	}
}

// SecurityReport contains results of security validation
type SecurityReport struct {
	IsValid          bool
	IsSuspicious     bool
	ThreatLevel      ThreatLevel
	Violations       []SecurityViolation
	SanitizedHeaders map[string]string
	TrustScore       float64 // 0.0 (untrusted) to 1.0 (trusted)
}

// ThreatLevel represents the severity of security threats
type ThreatLevel int

const (
	ThreatLevelNone ThreatLevel = iota
	ThreatLevelLow
	ThreatLevelMedium
	ThreatLevelHigh
	ThreatLevelCritical
)

// String returns string representation of ThreatLevel
func (t ThreatLevel) String() string {
	switch t {
	case ThreatLevelNone:
		return "none"
	case ThreatLevelLow:
		return "low"
	case ThreatLevelMedium:
		return "medium"
	case ThreatLevelHigh:
		return "high"
	case ThreatLevelCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// SecurityViolation represents a security rule violation
type SecurityViolation struct {
	Type        ViolationType
	Severity    ThreatLevel
	Description string
	Header      string
	Value       string
	Suggestion  string
}

// ViolationType represents types of security violations
type ViolationType int

const (
	ViolationIPSpoofing ViolationType = iota
	ViolationPrivateInPublic
	ViolationMalformedHeader
	ViolationInvalidIPv6
	ViolationSuspiciousChain
	ViolationHeaderTooLong
	ViolationChainTooLong
	ViolationInconsistentHeaders
)

// String returns string representation of ViolationType
func (v ViolationType) String() string {
	switch v {
	case ViolationIPSpoofing:
		return "ip_spoofing"
	case ViolationPrivateInPublic:
		return "private_in_public"
	case ViolationMalformedHeader:
		return "malformed_header"
	case ViolationInvalidIPv6:
		return "invalid_ipv6"
	case ViolationSuspiciousChain:
		return "suspicious_chain"
	case ViolationHeaderTooLong:
		return "header_too_long"
	case ViolationChainTooLong:
		return "chain_too_long"
	case ViolationInconsistentHeaders:
		return "inconsistent_headers"
	default:
		return "unknown"
	}
}

// SecurityValidator provides advanced security validations
type SecurityValidator struct {
	config         *SecurityConfig
	ipv6Regex      *regexp.Regexp
	headerPatterns map[string]*regexp.Regexp
	mu             sync.RWMutex
}

// Global security validator with default config
var globalSecurityValidator = NewSecurityValidator(DefaultSecurityConfig())

// NewSecurityValidator creates a new security validator with given config
func NewSecurityValidator(config *SecurityConfig) *SecurityValidator {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	// Enhanced IPv6 regex with strict validation
	ipv6Pattern := `^(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$|^::$|^::1$|^::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])$|^::(ffff(:0{1,4}){0,1}:){0,1}[0-9a-fA-F]{1,4}:[0-9a-fA-F]{1,4}$|^(([0-9a-fA-F]{1,4}:){1,7}|:)::$|^::([0-9a-fA-F]{1,4}:){1,7}$|^([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}$|^([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}$|^([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}$|^([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}$|^([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}$|^[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})$`

	headerPatterns := map[string]*regexp.Regexp{
		"ip":            regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$|^[0-9a-fA-F:]+$`),
		"forwarded":     regexp.MustCompile(`^for=(?:"?[^";,]+?"?);?.*$`),
		"port":          regexp.MustCompile(`^[0-9]+$`),
		"suspicious":    regexp.MustCompile(`[<>'"&\x00-\x1f\x7f-\x9f]`),
		"sql_injection": regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|exec|script)`),
	}

	return &SecurityValidator{
		config:         config,
		ipv6Regex:      regexp.MustCompile(ipv6Pattern),
		headerPatterns: headerPatterns,
	}
}

// ValidateSecurityOptimized performs comprehensive security validation on request adapter
func (sv *SecurityValidator) ValidateSecurityOptimized(adapter interfaces.RequestAdapter) *SecurityReport {
	report := &SecurityReport{
		IsValid:          true,
		IsSuspicious:     false,
		ThreatLevel:      ThreatLevelNone,
		Violations:       make([]SecurityViolation, 0),
		SanitizedHeaders: make(map[string]string),
		TrustScore:       1.0,
	}

	if adapter == nil {
		return report
	}

	// Get all headers for comprehensive analysis
	headers := adapter.GetAllHeaders()

	// 1. Validate header lengths and detect malformed headers
	sv.validateHeaders(headers, report)

	// 2. Extract and validate IP chain
	ipChain := sv.extractAndValidateIPChain(adapter, report)

	// 3. Detect IP spoofing attempts
	if sv.config.EnableSpoofingDetection {
		sv.detectSpoofing(headers, ipChain, report)
	}

	// 4. Enhanced IPv6 validation
	if sv.config.StrictIPv6Validation {
		sv.validateIPv6Addresses(ipChain, report)
	}

	// 5. Detect private IPs in public headers
	sv.detectPrivateInPublic(headers, report)

	// 6. Sanitize input if enabled
	if sv.config.EnableInputSanitization {
		sv.sanitizeHeaders(headers, report)
	}

	// 7. Calculate final trust score
	sv.calculateTrustScore(report)

	return report
}

// validateHeaders checks for malformed headers and excessive lengths
func (sv *SecurityValidator) validateHeaders(headers map[string]string, report *SecurityReport) {
	for header, value := range headers {
		// Check header length
		if len(value) > sv.config.MaxHeaderLength {
			report.Violations = append(report.Violations, SecurityViolation{
				Type:        ViolationHeaderTooLong,
				Severity:    ThreatLevelMedium,
				Description: "Header value exceeds maximum allowed length",
				Header:      header,
				Value:       value[:50] + "...", // Truncate for logging
				Suggestion:  "Verify if this is a legitimate long header or potential attack",
			})
			report.IsSuspicious = true
		}

		// Check for malicious patterns
		if sv.headerPatterns["suspicious"].MatchString(value) {
			report.Violations = append(report.Violations, SecurityViolation{
				Type:        ViolationMalformedHeader,
				Severity:    ThreatLevelHigh,
				Description: "Header contains suspicious characters",
				Header:      header,
				Value:       value,
				Suggestion:  "Header may contain malicious payload - sanitize input",
			})
			report.IsSuspicious = true
		}

		// Check for SQL injection patterns
		if sv.headerPatterns["sql_injection"].MatchString(value) {
			report.Violations = append(report.Violations, SecurityViolation{
				Type:        ViolationMalformedHeader,
				Severity:    ThreatLevelCritical,
				Description: "Header contains potential SQL injection patterns",
				Header:      header,
				Value:       value,
				Suggestion:  "Potential SQL injection attempt - block request",
			})
			report.IsSuspicious = true
		}
	}
}

// extractAndValidateIPChain extracts IP chain and validates length
func (sv *SecurityValidator) extractAndValidateIPChain(adapter interfaces.RequestAdapter, report *SecurityReport) []string {
	ipChain := defaultOptimizedExtractor.GetIPChainOptimized(adapter)

	// Check chain length
	if len(ipChain) > sv.config.MaxIPChainLength {
		report.Violations = append(report.Violations, SecurityViolation{
			Type:        ViolationChainTooLong,
			Severity:    ThreatLevelMedium,
			Description: "IP chain exceeds maximum allowed length",
			Header:      "IP-Chain",
			Value:       strings.Join(ipChain, ", "),
			Suggestion:  "Verify if this is legitimate proxy chain or potential attack",
		})
		report.IsSuspicious = true
	}

	return ipChain
}

// detectSpoofing detects IP spoofing attempts through header consistency analysis
func (sv *SecurityValidator) detectSpoofing(headers map[string]string, ipChain []string, report *SecurityReport) {
	trustedHeaders := []string{"CF-Connecting-IP", "True-Client-IP"}
	standardHeaders := []string{"X-Forwarded-For", "X-Real-IP"}

	trustedIPs := make([]string, 0)
	standardIPs := make([]string, 0)

	// Collect IPs from trusted vs standard headers
	for _, header := range trustedHeaders {
		if value, exists := headers[header]; exists && value != "" {
			ips := getIPsFromHeaderOptimized(header, value)
			trustedIPs = append(trustedIPs, ips...)
			returnStringSlice(ips)
		}
	}

	for _, header := range standardHeaders {
		if value, exists := headers[header]; exists && value != "" {
			ips := getIPsFromHeaderOptimized(header, value)
			standardIPs = append(standardIPs, ips...)
			returnStringSlice(ips)
		}
	}

	// Detect inconsistencies
	if len(trustedIPs) > 0 && len(standardIPs) > 0 {
		if !sv.hasCommonIP(trustedIPs, standardIPs) {
			report.Violations = append(report.Violations, SecurityViolation{
				Type:        ViolationInconsistentHeaders,
				Severity:    ThreatLevelHigh,
				Description: "Inconsistency detected between trusted and standard headers",
				Header:      "Multiple",
				Value:       "Trusted: " + strings.Join(trustedIPs, ",") + " | Standard: " + strings.Join(standardIPs, ","),
				Suggestion:  "Potential spoofing attempt - verify request authenticity",
			})
			report.IsSuspicious = true
		}
	}

	// Detect suspicious patterns in IP chain
	sv.detectSuspiciousChainPatterns(ipChain, report)
}

// hasCommonIP checks if two IP slices have at least one common IP
func (sv *SecurityValidator) hasCommonIP(ips1, ips2 []string) bool {
	ipSet := make(map[string]bool)
	for _, ip := range ips1 {
		ipSet[ip] = true
	}
	for _, ip := range ips2 {
		if ipSet[ip] {
			return true
		}
	}
	return false
}

// detectSuspiciousChainPatterns detects suspicious patterns in IP chains
func (sv *SecurityValidator) detectSuspiciousChainPatterns(ipChain []string, report *SecurityReport) {
	if len(ipChain) < 2 {
		return
	}

	privateCount := 0
	publicCount := 0
	loopbackCount := 0

	for _, ipStr := range ipChain {
		if ipInfo := parseIPOptimized(ipStr); ipInfo != nil && ipInfo.IP != nil {
			switch ipInfo.Type {
			case interfaces.IPTypePrivate:
				privateCount++
			case interfaces.IPTypePublic:
				publicCount++
			case interfaces.IPTypeLoopback:
				loopbackCount++
			}
		}
	}

	// Suspicious pattern: private -> public -> private (unusual routing)
	if privateCount > 1 && publicCount > 0 {
		report.Violations = append(report.Violations, SecurityViolation{
			Type:        ViolationSuspiciousChain,
			Severity:    ThreatLevelMedium,
			Description: "Suspicious IP chain pattern detected (private->public->private)",
			Header:      "IP-Chain",
			Value:       strings.Join(ipChain, " -> "),
			Suggestion:  "Review network topology - unusual routing pattern",
		})
		report.IsSuspicious = true
	}

	// Suspicious pattern: multiple loopback IPs
	if loopbackCount > 1 {
		report.Violations = append(report.Violations, SecurityViolation{
			Type:        ViolationSuspiciousChain,
			Severity:    ThreatLevelHigh,
			Description: "Multiple loopback IPs in chain",
			Header:      "IP-Chain",
			Value:       strings.Join(ipChain, " -> "),
			Suggestion:  "Potential spoofing - multiple loopback IPs unusual",
		})
		report.IsSuspicious = true
	}
}

// validateIPv6Addresses performs enhanced IPv6 validation
func (sv *SecurityValidator) validateIPv6Addresses(ipChain []string, report *SecurityReport) {
	for _, ipStr := range ipChain {
		if strings.Contains(ipStr, ":") { // Likely IPv6
			// Parse with net.ParseIP first
			if ip := net.ParseIP(ipStr); ip != nil && ip.To4() == nil {
				// Additional strict validation with regex
				if !sv.ipv6Regex.MatchString(ipStr) {
					report.Violations = append(report.Violations, SecurityViolation{
						Type:        ViolationInvalidIPv6,
						Severity:    ThreatLevelMedium,
						Description: "IPv6 address format validation failed",
						Header:      "IPv6-Address",
						Value:       ipStr,
						Suggestion:  "IPv6 format may be non-standard or malformed",
					})
					report.IsSuspicious = true
				}

				// Check for IPv6 embedding attacks
				sv.checkIPv6Embedding(ipStr, report)
			} else if strings.Contains(ipStr, ":") {
				// Contains colon but not valid IPv6
				report.Violations = append(report.Violations, SecurityViolation{
					Type:        ViolationInvalidIPv6,
					Severity:    ThreatLevelHigh,
					Description: "Invalid IPv6 format detected",
					Header:      "IPv6-Address",
					Value:       ipStr,
					Suggestion:  "Malformed IPv6 address - potential attack vector",
				})
				report.IsSuspicious = true
			}
		}
	}
}

// checkIPv6Embedding checks for IPv6 embedding attacks
func (sv *SecurityValidator) checkIPv6Embedding(ipv6 string, report *SecurityReport) {
	// Check for IPv4-mapped IPv6 addresses in suspicious contexts
	if strings.Contains(strings.ToLower(ipv6), "::ffff:") {
		// Extract embedded IPv4
		parts := strings.Split(ipv6, "::ffff:")
		if len(parts) == 2 {
			embeddedIP := parts[1]
			if ipInfo := parseIPOptimized(embeddedIP); ipInfo != nil {
				if ipInfo.Type == interfaces.IPTypePrivate {
					report.Violations = append(report.Violations, SecurityViolation{
						Type:        ViolationPrivateInPublic,
						Severity:    ThreatLevelMedium,
						Description: "IPv6-embedded private IPv4 address detected",
						Header:      "IPv6-Embedded",
						Value:       ipv6,
						Suggestion:  "Verify if private IP embedding is legitimate",
					})
					report.IsSuspicious = true
				}
			}
		}
	}
}

// detectPrivateInPublic detects private IPs in headers that should contain public IPs
func (sv *SecurityValidator) detectPrivateInPublic(headers map[string]string, report *SecurityReport) {
	publicHeaders := []string{"CF-Connecting-IP", "True-Client-IP", "X-Google-Real-IP", "X-Azure-ClientIP"}

	for _, header := range publicHeaders {
		if value, exists := headers[header]; exists && value != "" {
			ips := getIPsFromHeaderOptimized(header, value)
			for _, ipStr := range ips {
				if ipInfo := parseIPOptimized(ipStr); ipInfo != nil && ipInfo.IP != nil {
					if ipInfo.Type == interfaces.IPTypePrivate {
						report.Violations = append(report.Violations, SecurityViolation{
							Type:        ViolationPrivateInPublic,
							Severity:    ThreatLevelHigh,
							Description: "Private IP found in header that should contain public IP",
							Header:      header,
							Value:       ipStr,
							Suggestion:  "Potential spoofing - public headers should not contain private IPs",
						})
						report.IsSuspicious = true
					}
				}
			}
			returnStringSlice(ips)
		}
	}
}

// sanitizeHeaders sanitizes headers by removing dangerous characters
func (sv *SecurityValidator) sanitizeHeaders(headers map[string]string, report *SecurityReport) {
	for header, value := range headers {
		original := value

		// Remove control characters and suspicious sequences
		sanitized := strings.Map(func(r rune) rune {
			// Allow printable ASCII and basic extended characters
			if r >= 32 && r <= 126 || r >= 160 && r <= 255 {
				return r
			}
			// Replace control characters with space
			return ' '
		}, value)

		// Remove multiple spaces
		sanitized = regexp.MustCompile(`\s+`).ReplaceAllString(sanitized, " ")
		sanitized = strings.TrimSpace(sanitized)

		// Remove SQL injection patterns
		sanitized = sv.headerPatterns["sql_injection"].ReplaceAllString(sanitized, "")

		if sanitized != original {
			report.SanitizedHeaders[header] = sanitized
		}
	}
}

// calculateTrustScore calculates overall trust score based on violations
func (sv *SecurityValidator) calculateTrustScore(report *SecurityReport) {
	if len(report.Violations) == 0 {
		report.TrustScore = 1.0
		report.ThreatLevel = ThreatLevelNone
		return
	}

	score := 1.0
	maxThreat := ThreatLevelNone

	for _, violation := range report.Violations {
		// Deduct score based on severity
		switch violation.Severity {
		case ThreatLevelLow:
			score -= 0.1
		case ThreatLevelMedium:
			score -= 0.25
		case ThreatLevelHigh:
			score -= 0.4
		case ThreatLevelCritical:
			score -= 0.7
		}

		// Track highest threat level
		if violation.Severity > maxThreat {
			maxThreat = violation.Severity
		}
	}

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	report.TrustScore = score
	report.ThreatLevel = maxThreat

	// Mark as suspicious if trust score is low
	if score < 0.7 {
		report.IsSuspicious = true
	}
}

// Public API functions for security validation

// ValidateSecurity performs security validation on a request using global validator
func ValidateSecurity(request interface{}) *SecurityReport {
	adapter, err := providers.CreateAdapter(request)
	if err != nil {
		return &SecurityReport{
			IsValid:     false,
			ThreatLevel: ThreatLevelHigh,
			Violations: []SecurityViolation{{
				Type:        ViolationMalformedHeader,
				Severity:    ThreatLevelHigh,
				Description: "Failed to create request adapter",
				Suggestion:  "Invalid request type or malformed request",
			}},
		}
	}

	return globalSecurityValidator.ValidateSecurityOptimized(adapter)
}

// GetSecurityConfig returns current global security configuration
func GetSecurityConfig() *SecurityConfig {
	return globalSecurityValidator.config
}

// SetSecurityConfig updates global security configuration
func SetSecurityConfig(config *SecurityConfig) {
	globalSecurityValidator = NewSecurityValidator(config)
}

// IsRequestSecure checks if a request passes basic security validation
func IsRequestSecure(request interface{}) bool {
	report := ValidateSecurity(request)
	return report.IsValid && !report.IsSuspicious && report.ThreatLevel <= ThreatLevelLow
}

// GetTrustScore returns trust score for a request (0.0 - 1.0)
func GetTrustScore(request interface{}) float64 {
	report := ValidateSecurity(request)
	return report.TrustScore
}
