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

// Default optimized extractor instance (uses zero-allocation optimizations by default)
var defaultOptimizedExtractor = NewOptimizedExtractor()

// GetRealIP extracts the real client IP from any supported HTTP framework request.
// It automatically detects the framework and uses the appropriate adapter.
// This function uses zero-allocation optimizations with caching and buffer pooling.
// For net/http compatibility, it also accepts *http.Request directly.
func GetRealIP(request interface{}) string {
	// Handle net/http.Request directly for backward compatibility
	if httpReq, ok := request.(*http.Request); ok {
		adapter, err := providers.CreateAdapter(httpReq)
		if err != nil {
			return ""
		}
		return defaultOptimizedExtractor.GetRealIPOptimized(adapter)
	}

	// Handle other frameworks through adapter
	adapter, err := providers.CreateAdapter(request)
	if err != nil {
		return ""
	}
	return defaultOptimizedExtractor.GetRealIPOptimized(adapter)
}

// GetRealIPInfo extracts detailed IP information from any supported HTTP framework request.
// It automatically detects the framework and uses the appropriate adapter.
// This function uses zero-allocation optimizations with caching and buffer pooling.
func GetRealIPInfo(request interface{}) *IPInfo {
	// Handle net/http.Request directly for backward compatibility
	if httpReq, ok := request.(*http.Request); ok {
		adapter, err := providers.CreateAdapter(httpReq)
		if err != nil {
			return nil
		}
		return defaultOptimizedExtractor.GetRealIPInfoOptimized(adapter)
	}

	// Handle other frameworks through adapter
	adapter, err := providers.CreateAdapter(request)
	if err != nil {
		return nil
	}
	return defaultOptimizedExtractor.GetRealIPInfoOptimized(adapter)
}

// GetIPChain extracts the complete chain of IPs from any supported HTTP framework request.
// This function uses zero-allocation optimizations with caching and buffer pooling.
func GetIPChain(request interface{}) []string {
	// Handle net/http.Request directly for backward compatibility
	if httpReq, ok := request.(*http.Request); ok {
		adapter, err := providers.CreateAdapter(httpReq)
		if err != nil {
			return nil
		}
		return defaultOptimizedExtractor.GetIPChainOptimized(adapter)
	}

	// Handle other frameworks through adapter
	adapter, err := providers.CreateAdapter(request)
	if err != nil {
		return nil
	}
	return defaultOptimizedExtractor.GetIPChainOptimized(adapter)
}

// GetRealIP extracts the real client IP from a request adapter
// This method now uses optimized implementations internally for better performance
func (e *Extractor) GetRealIP(adapter interfaces.RequestAdapter) string {
	return defaultOptimizedExtractor.GetRealIPOptimized(adapter)
}

// GetRealIPInfo extracts detailed IP information from a request adapter
// This method now uses optimized implementations internally for better performance
func (e *Extractor) GetRealIPInfo(adapter interfaces.RequestAdapter) *IPInfo {
	return defaultOptimizedExtractor.GetRealIPInfoOptimized(adapter)
}

// GetIPChain extracts the complete chain of IPs from a request adapter
// This method now uses optimized implementations internally for better performance
func (e *Extractor) GetIPChain(adapter interfaces.RequestAdapter) []string {
	return defaultOptimizedExtractor.GetIPChainOptimized(adapter)
}

// ParseIP parses an IP string and returns detailed information about it
// This function now uses optimized parsing with caching for better performance
func ParseIP(ipStr string) *IPInfo {
	return parseIPOptimized(ipStr)
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
