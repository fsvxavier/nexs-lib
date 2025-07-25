// Package url provides URL parsing functionality with enhanced features.
package url

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/fsvxavier/nexs-lib/parsers/interfaces"
)

// ParsedURL represents a parsed URL with additional metadata.
type ParsedURL struct {
	*url.URL
	Port       int
	Subdomain  string
	Domain     string
	TLD        string
	IsSecure   bool
	IsLocalIP  bool
	Parameters map[string][]string
}

// Parser implements URL parsing with validation and error handling.
type Parser struct {
	config       *interfaces.ParserConfig
	allowedHosts []string
	blockedHosts []string
}

// NewParser creates a new URL parser with default configuration.
func NewParser() *Parser {
	return &Parser{
		config: interfaces.DefaultConfig(),
	}
}

// NewParserWithConfig creates a new URL parser with custom configuration.
func NewParserWithConfig(config *interfaces.ParserConfig) *Parser {
	return &Parser{
		config: config,
	}
}

// WithAllowedHosts sets allowed hosts for URL validation.
func (p *Parser) WithAllowedHosts(hosts []string) *Parser {
	p.allowedHosts = make([]string, len(hosts))
	copy(p.allowedHosts, hosts)
	return p
}

// WithBlockedHosts sets blocked hosts for URL validation.
func (p *Parser) WithBlockedHosts(hosts []string) *Parser {
	p.blockedHosts = make([]string, len(hosts))
	copy(p.blockedHosts, hosts)
	return p
}

// Parse implements interfaces.Parser.
func (p *Parser) Parse(ctx context.Context, data []byte) (*ParsedURL, error) {
	return p.ParseString(ctx, string(data))
}

// ParseString parses a URL string.
func (p *Parser) ParseString(ctx context.Context, input string) (*ParsedURL, error) {
	if err := p.validateInput(input); err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(input)
	if err != nil {
		return nil, &interfaces.ParseError{
			Type:    interfaces.ErrorTypeSyntax,
			Message: "invalid URL format",
			Cause:   err,
		}
	}

	result := &ParsedURL{
		URL:        parsedURL,
		Parameters: make(map[string][]string),
	}

	// Extract additional information
	if err := p.enrichURL(result); err != nil {
		return nil, err
	}

	// Validate parsed URL
	if err := p.Validate(ctx, result); err != nil {
		return nil, err
	}

	return result, nil
}

// ParseReader is not applicable for URLs, returns error.
func (p *Parser) ParseReader(ctx context.Context, reader interface{}) (*ParsedURL, error) {
	return nil, &interfaces.ParseError{
		Type:    interfaces.ErrorTypeValidation,
		Message: "ParseReader not supported for URL parser",
	}
}

// Validate validates the parsed URL.
func (p *Parser) Validate(ctx context.Context, result *ParsedURL) error {
	if result == nil {
		return &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "result cannot be nil",
		}
	}

	if result.URL == nil {
		return &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "URL cannot be nil",
		}
	}

	// Check allowed hosts
	if len(p.allowedHosts) > 0 {
		allowed := false
		for _, host := range p.allowedHosts {
			if result.Host == host || strings.HasSuffix(result.Host, "."+host) {
				allowed = true
				break
			}
		}
		if !allowed {
			return &interfaces.ParseError{
				Type:    interfaces.ErrorTypeValidation,
				Message: fmt.Sprintf("host %s is not allowed", result.Host),
			}
		}
	}

	// Check blocked hosts
	if len(p.blockedHosts) > 0 {
		for _, host := range p.blockedHosts {
			if result.Host == host || strings.HasSuffix(result.Host, "."+host) {
				return &interfaces.ParseError{
					Type:    interfaces.ErrorTypeValidation,
					Message: fmt.Sprintf("host %s is blocked", result.Host),
				}
			}
		}
	}

	return nil
}

// validateInput validates input URL string.
func (p *Parser) validateInput(input string) error {
	if len(input) == 0 {
		return &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "input URL is empty",
		}
	}

	if p.config.MaxSize > 0 && int64(len(input)) > p.config.MaxSize {
		return &interfaces.ParseError{
			Type:    interfaces.ErrorTypeSize,
			Message: fmt.Sprintf("URL length %d exceeds maximum %d", len(input), p.config.MaxSize),
		}
	}

	// Basic URL validation - must have a scheme or be a valid relative URL
	if !strings.Contains(input, "://") && !strings.HasPrefix(input, "/") && !strings.Contains(input, ".") {
		return &interfaces.ParseError{
			Type:    interfaces.ErrorTypeSyntax,
			Message: "invalid URL format: missing scheme, path, or domain",
		}
	}

	return nil
}

// enrichURL extracts additional information from the parsed URL.
func (p *Parser) enrichURL(result *ParsedURL) error {
	// Extract port
	if result.URL.Port() != "" {
		port, err := strconv.Atoi(result.URL.Port())
		if err != nil {
			return &interfaces.ParseError{
				Type:    interfaces.ErrorTypeSyntax,
				Message: "invalid port number",
				Cause:   err,
			}
		}
		result.Port = port
	} else {
		// Default ports
		switch result.Scheme {
		case "http":
			result.Port = 80
		case "https":
			result.Port = 443
		case "ftp":
			result.Port = 21
		case "ssh":
			result.Port = 22
		}
	}

	// Check if secure
	result.IsSecure = result.Scheme == "https" || result.Scheme == "ftps"

	// Parse hostname components
	p.parseHostname(result)

	// Parse query parameters
	result.Parameters = result.Query()

	// Check if local IP
	result.IsLocalIP = p.isLocalIP(result.Hostname())

	return nil
}

// parseHostname extracts subdomain, domain, and TLD from hostname.
func (p *Parser) parseHostname(result *ParsedURL) {
	hostname := result.Hostname()
	if hostname == "" {
		return
	}

	// Skip IP addresses
	if p.isIPAddress(hostname) {
		result.Domain = hostname
		return
	}

	parts := strings.Split(hostname, ".")
	if len(parts) < 2 {
		result.Domain = hostname
		return
	}

	// Extract TLD (last part)
	result.TLD = parts[len(parts)-1]

	// Extract domain (second to last part)
	if len(parts) >= 2 {
		result.Domain = parts[len(parts)-2] + "." + result.TLD
	}

	// Extract subdomain (everything before domain)
	if len(parts) > 2 {
		subdomainParts := parts[:len(parts)-2]
		result.Subdomain = strings.Join(subdomainParts, ".")
	}
}

// isIPAddress checks if hostname is an IP address.
func (p *Parser) isIPAddress(hostname string) bool {
	// Simple check for IPv4 and IPv6
	return strings.Contains(hostname, ":") || // IPv6
		(strings.Count(hostname, ".") == 3 && // IPv4
			!strings.Contains(hostname, " "))
}

// isLocalIP checks if hostname represents a local IP address.
func (p *Parser) isLocalIP(hostname string) bool {
	localIPs := []string{
		"localhost",
		"127.",
		"192.168.",
		"10.",
		"172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.",
		"172.24.", "172.25.", "172.26.", "172.27.",
		"172.28.", "172.29.", "172.30.", "172.31.",
		"::1",
		"fc00:",
		"fd00:",
	}

	for _, localIP := range localIPs {
		if strings.HasPrefix(hostname, localIP) {
			return true
		}
	}

	return false
}

// Formatter implements URL formatting and building.
type Formatter struct{}

// NewFormatter creates a new URL formatter.
func NewFormatter() *Formatter {
	return &Formatter{}
}

// Format implements interfaces.Formatter.
func (f *Formatter) Format(ctx context.Context, data *ParsedURL) ([]byte, error) {
	if data == nil || data.URL == nil {
		return nil, &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "data cannot be nil",
		}
	}

	return []byte(data.String()), nil
}

// FormatString implements interfaces.Formatter.
func (f *Formatter) FormatString(ctx context.Context, data *ParsedURL) (string, error) {
	if data == nil || data.URL == nil {
		return "", &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "data cannot be nil",
		}
	}

	return data.String(), nil
}

// FormatWriter is not commonly used for URLs, returns error.
func (f *Formatter) FormatWriter(ctx context.Context, data *ParsedURL, writer interface{}) error {
	return &interfaces.ParseError{
		Type:    interfaces.ErrorTypeValidation,
		Message: "FormatWriter not supported for URL formatter",
	}
}

// Builder provides a fluent interface for building URLs.
type Builder struct {
	scheme   string
	host     string
	port     int
	path     string
	query    url.Values
	fragment string
}

// NewBuilder creates a new URL builder.
func NewBuilder() *Builder {
	return &Builder{
		query: make(url.Values),
	}
}

// Scheme sets the URL scheme.
func (b *Builder) Scheme(scheme string) *Builder {
	b.scheme = scheme
	return b
}

// Host sets the URL host.
func (b *Builder) Host(host string) *Builder {
	b.host = host
	return b
}

// Port sets the URL port.
func (b *Builder) Port(port int) *Builder {
	b.port = port
	return b
}

// Path sets the URL path.
func (b *Builder) Path(path string) *Builder {
	b.path = path
	return b
}

// AddParam adds a query parameter.
func (b *Builder) AddParam(key, value string) *Builder {
	b.query.Add(key, value)
	return b
}

// SetParam sets a query parameter (replaces existing).
func (b *Builder) SetParam(key, value string) *Builder {
	b.query.Set(key, value)
	return b
}

// Fragment sets the URL fragment.
func (b *Builder) Fragment(fragment string) *Builder {
	b.fragment = fragment
	return b
}

// Build constructs the final URL.
func (b *Builder) Build() (*url.URL, error) {
	if b.scheme == "" {
		return nil, &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "scheme is required",
		}
	}

	if b.host == "" {
		return nil, &interfaces.ParseError{
			Type:    interfaces.ErrorTypeValidation,
			Message: "host is required",
		}
	}

	u := &url.URL{
		Scheme:   b.scheme,
		Host:     b.host,
		Path:     b.path,
		RawQuery: b.query.Encode(),
		Fragment: b.fragment,
	}

	// Set default path to "/" if empty
	if u.Path == "" {
		u.Path = "/"
	}

	// Add port if not default
	if b.port > 0 {
		defaultPorts := map[string]int{
			"http":  80,
			"https": 443,
			"ftp":   21,
			"ssh":   22,
		}

		if defaultPort, exists := defaultPorts[b.scheme]; !exists || b.port != defaultPort {
			u.Host = fmt.Sprintf("%s:%d", b.host, b.port)
		}
	}

	return u, nil
}

// BuildString constructs the final URL as string.
func (b *Builder) BuildString() (string, error) {
	u, err := b.Build()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Utility functions

// IsValidURL checks if a string is a valid URL.
func IsValidURL(input string) bool {
	parser := NewParser()
	_, err := parser.ParseString(context.Background(), input)
	return err == nil
}

// ExtractDomain extracts domain from URL string.
func ExtractDomain(urlStr string) (string, error) {
	parser := NewParser()
	parsed, err := parser.ParseString(context.Background(), urlStr)
	if err != nil {
		return "", err
	}
	return parsed.Domain, nil
}

// NormalizeURL normalizes a URL by removing default ports, etc.
func NormalizeURL(urlStr string) (string, error) {
	parser := NewParser()
	parsed, err := parser.ParseString(context.Background(), urlStr)
	if err != nil {
		return "", err
	}

	// Remove default ports
	u := parsed.URL
	if (u.Scheme == "http" && parsed.Port == 80) ||
		(u.Scheme == "https" && parsed.Port == 443) {
		u.Host = u.Hostname()
	}

	return u.String(), nil
}

// JoinURL joins a base URL with a relative path.
func JoinURL(baseURL, relativePath string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", &interfaces.ParseError{
			Type:    interfaces.ErrorTypeSyntax,
			Message: "invalid base URL",
			Cause:   err,
		}
	}

	relative, err := url.Parse(relativePath)
	if err != nil {
		return "", &interfaces.ParseError{
			Type:    interfaces.ErrorTypeSyntax,
			Message: "invalid relative path",
			Cause:   err,
		}
	}

	result := base.ResolveReference(relative)
	return result.String(), nil
}
