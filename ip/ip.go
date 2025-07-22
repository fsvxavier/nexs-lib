// Package ip provides robust IP address identification and manipulation utilities.
// It specializes in extracting real client IPs from HTTP requests even when
// they pass through proxies, relays, VPNs, and other network intermediaries.
//
// This package now supports multiple HTTP frameworks through a pluggable adapter system:
// - net/http (standard library)
// - Fiber
// - Gin
// - Echo
// - FastHTTP
// - Atreugo
package ip

import (
	"net"
	"net/http"
	"strings"

	"github.com/fsvxavier/nexs-lib/ip/interfaces"
	"github.com/fsvxavier/nexs-lib/ip/providers"
)

// Re-export types from interfaces package for backward compatibility
type (
	IPType = interfaces.IPType
	IPInfo = interfaces.IPInfo
)

// Re-export constants from interfaces package
const (
	IPTypeUnknown   = interfaces.IPTypeUnknown
	IPTypePublic    = interfaces.IPTypePublic
	IPTypePrivate   = interfaces.IPTypePrivate
	IPTypeLoopback  = interfaces.IPTypeLoopback
	IPTypeMulticast = interfaces.IPTypeMulticast
	IPTypeLinkLocal = interfaces.IPTypeLinkLocal
	IPTypeBroadcast = interfaces.IPTypeBroadcast
)

// Extractor implements the IPExtractor interface and provides IP extraction functionality
type Extractor struct{}

// NewExtractor creates a new IP extractor instance
func NewExtractor() *Extractor {
	return &Extractor{}
}

// Headers that commonly contain the real client IP
var realIPHeaders = []string{
	"CF-Connecting-IP",         // Cloudflare
	"True-Client-IP",           // Cloudflare Enterprise
	"X-Real-IP",                // Nginx
	"X-Forwarded-For",          // Standard proxy header
	"X-Client-IP",              // Apache
	"X-Cluster-Client-IP",      // Cluster environments
	"X-Forwarded",              // RFC 7239
	"Forwarded-For",            // RFC 7239
	"Forwarded",                // RFC 7239
	"X-Original-Forwarded-For", // Some CDNs
	"X-Azure-ClientIP",         // Azure
	"X-Google-Real-IP",         // Google Cloud
}

// Private IP ranges for IPv4 and IPv6
var (
	privateIPv4Ranges = []*net.IPNet{
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},     // 10.0.0.0/8
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},  // 172.16.0.0/12
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)}, // 192.168.0.0/16
	}

	linkLocalIPv4Range = &net.IPNet{
		IP:   net.IPv4(169, 254, 0, 0),
		Mask: net.CIDRMask(16, 32), // 169.254.0.0/16
	}

	privateIPv6Ranges = []*net.IPNet{
		{IP: net.ParseIP("fc00::"), Mask: net.CIDRMask(7, 128)},  // fc00::/7
		{IP: net.ParseIP("fe80::"), Mask: net.CIDRMask(10, 128)}, // fe80::/10 (link-local)
	}
)

// Default extractor instance
var defaultExtractor = NewExtractor()

// GetRealIP extracts the real client IP from any supported HTTP framework request.
// It automatically detects the framework and uses the appropriate adapter.
// For net/http compatibility, it also accepts *http.Request directly.
func GetRealIP(request interface{}) string {
	// Handle net/http.Request directly for backward compatibility
	if httpReq, ok := request.(*http.Request); ok {
		adapter, err := providers.CreateAdapter(httpReq)
		if err != nil {
			return ""
		}
		return defaultExtractor.GetRealIP(adapter)
	}

	// Handle other frameworks through adapter
	adapter, err := providers.CreateAdapter(request)
	if err != nil {
		return ""
	}
	return defaultExtractor.GetRealIP(adapter)
}

// GetRealIPInfo extracts detailed IP information from any supported HTTP framework request.
// It automatically detects the framework and uses the appropriate adapter.
func GetRealIPInfo(request interface{}) *IPInfo {
	// Handle net/http.Request directly for backward compatibility
	if httpReq, ok := request.(*http.Request); ok {
		adapter, err := providers.CreateAdapter(httpReq)
		if err != nil {
			return nil
		}
		return defaultExtractor.GetRealIPInfo(adapter)
	}

	// Handle other frameworks through adapter
	adapter, err := providers.CreateAdapter(request)
	if err != nil {
		return nil
	}
	return defaultExtractor.GetRealIPInfo(adapter)
}

// GetIPChain extracts the complete chain of IPs from any supported HTTP framework request.
func GetIPChain(request interface{}) []string {
	// Handle net/http.Request directly for backward compatibility
	if httpReq, ok := request.(*http.Request); ok {
		adapter, err := providers.CreateAdapter(httpReq)
		if err != nil {
			return nil
		}
		return defaultExtractor.GetIPChain(adapter)
	}

	// Handle other frameworks through adapter
	adapter, err := providers.CreateAdapter(request)
	if err != nil {
		return nil
	}
	return defaultExtractor.GetIPChain(adapter)
}

// GetRealIP extracts the real client IP from a request adapter
func (e *Extractor) GetRealIP(adapter interfaces.RequestAdapter) string {
	ipInfo := e.GetRealIPInfo(adapter)
	if ipInfo != nil && ipInfo.IP != nil {
		return ipInfo.IP.String()
	}
	return ""
}

// GetRealIPInfo extracts detailed IP information from a request adapter
func (e *Extractor) GetRealIPInfo(adapter interfaces.RequestAdapter) *IPInfo {
	if adapter == nil {
		return nil
	}

	// Try each header in order of preference
	for _, header := range realIPHeaders {
		headerValue := adapter.GetHeader(header)
		if headerValue == "" {
			continue
		}

		ips := getIPsFromHeader(header, headerValue)
		for _, ipStr := range ips {
			if ipInfo := ParseIP(ipStr); ipInfo != nil && ipInfo.IP != nil {
				// Prefer public IPs
				if IsPublicIP(ipInfo.IP) {
					ipInfo.Source = header
					return ipInfo
				}
			}
		}
	}

	// If no public IP found in headers, try to find any valid IP in headers
	for _, header := range realIPHeaders {
		headerValue := adapter.GetHeader(header)
		if headerValue == "" {
			continue
		}

		ips := getIPsFromHeader(header, headerValue)
		for _, ipStr := range ips {
			if ipInfo := ParseIP(ipStr); ipInfo != nil && ipInfo.IP != nil {
				ipInfo.Source = header
				return ipInfo
			}
		}
	}

	// Fallback to remote address
	if remoteAddr := adapter.GetRemoteAddr(); remoteAddr != "" {
		if ip := getIPFromRemoteAddr(remoteAddr); ip != "" {
			if ipInfo := ParseIP(ip); ipInfo != nil {
				ipInfo.Source = "RemoteAddr"
				return ipInfo
			}
		}
	}

	return nil
}

// GetIPChain extracts the complete chain of IPs from a request adapter
func (e *Extractor) GetIPChain(adapter interfaces.RequestAdapter) []string {
	if adapter == nil {
		return nil
	}

	var chain []string
	seen := make(map[string]bool)

	// Collect IPs from all headers
	for _, header := range realIPHeaders {
		headerValue := adapter.GetHeader(header)
		if headerValue == "" {
			continue
		}

		ips := getIPsFromHeader(header, headerValue)
		for _, ipStr := range ips {
			if ipInfo := ParseIP(ipStr); ipInfo != nil && ipInfo.IP != nil {
				ipString := ipInfo.IP.String()
				if !seen[ipString] {
					chain = append(chain, ipString)
					seen[ipString] = true
				}
			}
		}
	}

	// Add remote address if not already present
	if remoteAddr := adapter.GetRemoteAddr(); remoteAddr != "" {
		if ip := getIPFromRemoteAddr(remoteAddr); ip != "" {
			if ipInfo := ParseIP(ip); ipInfo != nil && ipInfo.IP != nil {
				ipString := ipInfo.IP.String()
				if !seen[ipString] {
					chain = append(chain, ipString)
				}
			}
		}
	}

	return chain
}

// ParseIP parses an IP string and returns detailed information about it
func ParseIP(ipStr string) *IPInfo {
	ipStr = strings.TrimSpace(ipStr)
	if ipStr == "" {
		return nil
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return &IPInfo{
			Original: ipStr,
		}
	}

	ipInfo := &IPInfo{
		IP:       ip,
		Original: ipStr,
		IsIPv4:   ip.To4() != nil,
		IsIPv6:   ip.To4() == nil,
		Type:     classifyIP(ip),
	}

	// Set convenience flags
	ipInfo.IsPublic = ipInfo.Type == IPTypePublic
	ipInfo.IsPrivate = ipInfo.Type == IPTypePrivate

	return ipInfo
}

// IsValidIP checks if a string represents a valid IP address
func IsValidIP(ipStr string) bool {
	return net.ParseIP(strings.TrimSpace(ipStr)) != nil
}

// IsPrivateIP checks if an IP address is in a private range
func IsPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}

	// Check IPv4 private ranges
	if ip.To4() != nil {
		for _, privateRange := range privateIPv4Ranges {
			if privateRange.Contains(ip) {
				return true
			}
		}
		return false
	}

	// Check IPv6 private ranges
	for _, privateRange := range privateIPv6Ranges {
		if privateRange.Contains(ip) {
			return true
		}
	}

	return false
}

// IsPublicIP checks if an IP address is publicly routable
func IsPublicIP(ip net.IP) bool {
	if ip == nil {
		return false
	}

	// Not public if it's private, loopback, multicast, link-local, or broadcast
	ipType := classifyIP(ip)
	return ipType == IPTypePublic
}

// ConvertIPv4ToIPv6 converts an IPv4 address to IPv6 format
func ConvertIPv4ToIPv6(ip net.IP) net.IP {
	if ip == nil {
		return nil
	}

	// If already IPv6, return as-is
	if ip.To4() == nil {
		return ip
	}

	// Convert IPv4 to IPv6
	return ip.To16()
}

// Helper functions

// getIPsFromHeader extracts IPs from a header value
func getIPsFromHeader(headerName, headerValue string) []string {
	if headerValue == "" {
		return nil
	}

	// Handle RFC 7239 Forwarded header specially
	if strings.ToLower(headerName) == "forwarded" {
		return parseForwardedHeader(headerValue)
	}

	// Handle comma-separated IP lists
	return parseCommaSeparatedIPs(headerValue)
}

// parseCommaSeparatedIPs parses comma-separated IP addresses
func parseCommaSeparatedIPs(value string) []string {
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	var ips []string

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Remove port if present
		if ip := getIPFromRemoteAddr(part); ip != "" {
			ips = append(ips, ip)
		}
	}

	return ips
}

// parseForwardedHeader parses RFC 7239 Forwarded header
func parseForwardedHeader(value string) []string {
	if value == "" {
		return nil
	}

	var ips []string

	// Split by comma to handle multiple forwarded entries
	entries := strings.Split(value, ",")

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		// Parse each parameter in the entry
		params := strings.Split(entry, ";")
		for _, param := range params {
			param = strings.TrimSpace(param)

			// Look for for= parameter
			if strings.HasPrefix(strings.ToLower(param), "for=") {
				forValue := param[4:] // Remove "for="
				forValue = strings.TrimSpace(forValue)

				// Remove quotes if present
				if strings.HasPrefix(forValue, "\"") && strings.HasSuffix(forValue, "\"") {
					forValue = forValue[1 : len(forValue)-1]
				}

				// Remove brackets for IPv6
				forValue = strings.Trim(forValue, "[]")

				// Remove port if present
				if ip := getIPFromRemoteAddr(forValue); ip != "" {
					ips = append(ips, ip)
				}
			}
		}
	}

	return ips
}

// getIPFromRemoteAddr extracts IP from remote address (removes port)
func getIPFromRemoteAddr(remoteAddr string) string {
	if remoteAddr == "" {
		return ""
	}

	// Handle IPv6 with brackets and port: [::1]:8080
	if strings.HasPrefix(remoteAddr, "[") {
		if idx := strings.Index(remoteAddr, "]:"); idx != -1 {
			return remoteAddr[1:idx]
		}
		// Just brackets without port: [::1]
		if strings.HasSuffix(remoteAddr, "]") {
			return remoteAddr[1 : len(remoteAddr)-1]
		}
	}

	// Handle IPv4 with port: 192.168.1.1:8080
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		// Make sure it's not IPv6 without brackets
		if strings.Count(remoteAddr, ":") == 1 {
			return remoteAddr[:idx]
		}
	}

	// Return as-is if no port found
	return remoteAddr
}

// classifyIP classifies an IP address by type
func classifyIP(ip net.IP) IPType {
	if ip == nil {
		return IPTypeUnknown
	}

	// Check loopback
	if ip.IsLoopback() {
		return IPTypeLoopback
	}

	// Check multicast
	if ip.IsMulticast() {
		return IPTypeMulticast
	}

	// Check link-local
	if ip.IsLinkLocalUnicast() {
		return IPTypeLinkLocal
	}

	// Check IPv4 broadcast
	if ip.To4() != nil && ip.Equal(net.IPv4bcast) {
		return IPTypeBroadcast
	}

	// Check IPv4 link-local range (169.254.0.0/16)
	if ip.To4() != nil && linkLocalIPv4Range.Contains(ip) {
		return IPTypeLinkLocal
	}

	// Check private ranges
	if IsPrivateIP(ip) {
		return IPTypePrivate
	}

	// Default to public
	return IPTypePublic
}

// GetSupportedFrameworks returns a list of supported HTTP frameworks
func GetSupportedFrameworks() []string {
	return providers.GetSupportedProviders()
}

// RegisterCustomProvider allows registering custom HTTP framework providers
func RegisterCustomProvider(provider interfaces.ProviderFactory) {
	providers.RegisterProvider(provider)
}
