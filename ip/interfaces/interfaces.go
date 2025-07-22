// Package interfaces provides the core interfaces for IP extraction adapters.
// This package defines the contracts that all HTTP framework adapters must implement.
package interfaces

import (
	"context"
	"net"
)

// RequestAdapter defines the interface that all HTTP framework adapters must implement.
// It provides a uniform way to extract headers and remote address from different
// HTTP framework request objects.
type RequestAdapter interface {
	// GetHeader returns the value of the specified header
	GetHeader(key string) string

	// GetAllHeaders returns all headers as a map
	GetAllHeaders() map[string]string

	// GetRemoteAddr returns the remote address of the connection
	GetRemoteAddr() string

	// GetMethod returns the HTTP method
	GetMethod() string

	// GetPath returns the request path
	GetPath() string
}

// IPExtractor defines the interface for IP extraction functionality.
// This interface provides the core methods for extracting and analyzing IP addresses.
type IPExtractor interface {
	// GetRealIP extracts the real client IP from the request
	GetRealIP(adapter RequestAdapter) string

	// GetRealIPInfo extracts detailed IP information from the request
	GetRealIPInfo(adapter RequestAdapter) *IPInfo

	// GetIPChain extracts the complete chain of IPs from the request
	GetIPChain(adapter RequestAdapter) []string
}

// AdvancedIPDetector defines the interface for advanced IP detection capabilities
type AdvancedIPDetector interface {
	// DetectAdvanced performs comprehensive IP detection analysis
	DetectAdvanced(ctx context.Context, ip net.IP) (*DetectionResult, error)

	// LoadVPNDatabase loads VPN database from external source
	LoadVPNDatabase(source interface{}) error

	// LoadASNDatabase loads ASN database from external source
	LoadASNDatabase(source interface{}) error

	// GetCacheStats returns cache statistics
	GetCacheStats() (size int, hitRate float64)

	// ClearCache clears the detection cache
	ClearCache()
}

// ConcurrentIPProcessor defines the interface for concurrent IP processing
type ConcurrentIPProcessor interface {
	// ProcessIPs processes multiple IPs concurrently
	ProcessIPs(ctx context.Context, ips []string) <-chan IPProcessResult

	// Close closes the processor and cleans up resources
	Close() error
}

// DetectionResult contains comprehensive IP detection results
type DetectionResult struct {
	IP              net.IP
	IsVPN           bool
	IsProxy         bool
	IsTor           bool
	IsDatacenter    bool
	IsCloudProvider bool
	VPNProvider     *VPNProvider
	ASNInfo         *ASNInfo
	TrustScore      float64
	RiskLevel       string
	DetectionTime   int64 // Duration in nanoseconds
}

// VPNProvider represents a known VPN service provider
type VPNProvider struct {
	Name        string
	Type        string // "commercial", "datacenter", "tor", "proxy"
	Reliability float64
}

// ASNInfo contains information about an Autonomous System Number
type ASNInfo struct {
	ASN             uint32
	Name            string
	Country         string
	Type            string // "isp", "hosting", "government", "enterprise"
	IsCloudProvider bool
	CloudProvider   string
}

// IPProcessResult contains the result of processing a single IP
type IPProcessResult struct {
	IP        string
	Detection *DetectionResult
	Duration  int64 // Duration in nanoseconds
	Error     error
}

// IPInfo contains detailed information about an IP address
type IPInfo struct {
	IP       net.IP
	Type     IPType
	IsIPv4   bool
	IsIPv6   bool
	IsProxy  bool
	IsRelay  bool
	IsVPN    bool
	Original string
	// IsPublic indicates if the IP is publicly routable
	IsPublic bool
	// IsPrivate indicates if the IP is in a private range
	IsPrivate bool
	// Source indicates which header or source provided this IP
	Source string
	// TrustScore indicates the trustworthiness of the IP (0.0-1.0)
	TrustScore float64
	// RiskLevel indicates the risk level ("low", "medium", "high", "critical")
	RiskLevel string
	// IsDatacenter indicates if the IP belongs to a datacenter
	IsDatacenter bool
	// IsCloudProvider indicates if the IP belongs to a cloud provider
	IsCloudProvider bool
	// CloudProvider name if known
	CloudProvider string
}

// String returns a string representation of IPInfo
func (info IPInfo) String() string {
	if info.IP == nil {
		return info.Original
	}
	return info.IP.String()
}

// IPType represents the type/classification of an IP address
type IPType int

const (
	// IPTypeUnknown represents an unknown IP type
	IPTypeUnknown IPType = iota
	// IPTypePublic represents a public IP address
	IPTypePublic
	// IPTypePrivate represents a private IP address
	IPTypePrivate
	// IPTypeLoopback represents a loopback IP address
	IPTypeLoopback
	// IPTypeMulticast represents a multicast IP address
	IPTypeMulticast
	// IPTypeLinkLocal represents a link-local IP address
	IPTypeLinkLocal
	// IPTypeBroadcast represents a broadcast IP address
	IPTypeBroadcast
)

// String returns the string representation of IPType
func (t IPType) String() string {
	switch t {
	case IPTypePublic:
		return "public"
	case IPTypePrivate:
		return "private"
	case IPTypeLoopback:
		return "loopback"
	case IPTypeMulticast:
		return "multicast"
	case IPTypeLinkLocal:
		return "link-local"
	case IPTypeBroadcast:
		return "broadcast"
	default:
		return "unknown"
	}
}

// ProviderFactory defines the interface for creating request adapters
type ProviderFactory interface {
	// CreateAdapter creates a new request adapter for the given request object
	CreateAdapter(request interface{}) (RequestAdapter, error)

	// GetProviderName returns the name of the provider
	GetProviderName() string

	// SupportsType checks if the provider supports the given request type
	SupportsType(request interface{}) bool
}
