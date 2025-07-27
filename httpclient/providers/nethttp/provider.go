// Package nethttp provides HTTP client implementation using the standard net/http library.
package nethttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
	jsoniter "github.com/json-iterator/go"
)

var (
	jsonCodec = jsoniter.ConfigCompatibleWithStandardLibrary
)

// Provider implements the HTTP provider using the standard net/http library.
type Provider struct {
	client  *http.Client
	config  *interfaces.Config
	metrics *interfaces.ProviderMetrics
	mu      sync.RWMutex
}

// NewProvider creates a new net/http provider instance.
func NewProvider(config *interfaces.Config) (interfaces.Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	p := &Provider{
		config: config,
		metrics: &interfaces.ProviderMetrics{
			TotalRequests:      0,
			SuccessfulRequests: 0,
			FailedRequests:     0,
			AverageLatency:     0,
			LastRequestTime:    time.Time{},
		},
	}

	if err := p.Configure(config); err != nil {
		return nil, fmt.Errorf("failed to configure provider: %w", err)
	}

	return p, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "nethttp"
}

// Version returns the provider version.
func (p *Provider) Version() string {
	return "1.0.0"
}

// Configure sets up the HTTP client with the provided configuration.
func (p *Provider) Configure(config *interfaces.Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	p.config = config

	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		IdleConnTimeout:     config.IdleConnTimeout,
		TLSHandshakeTimeout: config.TLSHandshakeTimeout,
		DisableKeepAlives:   config.DisableKeepAlives,
		DisableCompression:  config.DisableCompression,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify,
		},
		DialContext: (&net.Dialer{
			Timeout:   config.Timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	p.client = &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return nil
}

// SetDefaults sets default configuration values.
func (p *Provider) SetDefaults() {
	// Default values are handled by the config package
}

// IsHealthy checks if the provider is healthy and ready to handle requests.
func (p *Provider) IsHealthy() bool {
	return p.client != nil
}

// GetMetrics returns the current provider metrics.
func (p *Provider) GetMetrics() *interfaces.ProviderMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Return a copy to avoid race conditions
	return &interfaces.ProviderMetrics{
		TotalRequests:      p.metrics.TotalRequests,
		SuccessfulRequests: p.metrics.SuccessfulRequests,
		FailedRequests:     p.metrics.FailedRequests,
		AverageLatency:     p.metrics.AverageLatency,
		LastRequestTime:    p.metrics.LastRequestTime,
	}
}

// DoRequest performs an HTTP request using the net/http client.
func (p *Provider) DoRequest(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	start := time.Now()

	// Update metrics
	p.updateMetricsStart()

	// Build URL
	url := req.URL
	if p.config.BaseURL != "" && !strings.HasPrefix(req.URL, "http") {
		url = strings.TrimSuffix(p.config.BaseURL, "/") + "/" + strings.TrimPrefix(req.URL, "/")
	}

	// Prepare request body
	var body io.Reader
	if req.Body != nil {
		bodyBytes, err := p.marshalBody(req.Body)
		if err != nil {
			p.updateMetricsEnd(start, false)
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, body)
	if err != nil {
		p.updateMetricsEnd(start, false)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	p.setHeaders(httpReq, req.Headers)

	// Add tracing if enabled
	if p.config.TracingEnabled {
		p.addTracing(ctx, httpReq)
	}

	// Use request-specific timeout if provided
	client := p.client
	if req.Timeout > 0 {
		client = &http.Client{
			Transport: p.client.Transport,
			Timeout:   req.Timeout,
		}
	}

	// Perform request
	httpResp, err := client.Do(httpReq)
	if err != nil {
		p.updateMetricsEnd(start, false)
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		p.updateMetricsEnd(start, false)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	latency := time.Since(start)
	isSuccess := httpResp.StatusCode >= 200 && httpResp.StatusCode < 300

	p.updateMetricsEnd(start, isSuccess)

	// Build response
	response := &interfaces.Response{
		StatusCode: httpResp.StatusCode,
		Body:       respBody,
		Headers:    make(map[string]string),
		IsError:    !isSuccess,
		Latency:    latency,
		Request:    req,
	}

	// Copy response headers
	for k, v := range httpResp.Header {
		if len(v) > 0 {
			response.Headers[k] = strings.Join(v, ", ")
		}
	}

	return response, nil
}

// marshalBody marshals the request body to JSON bytes.
func (p *Provider) marshalBody(body interface{}) ([]byte, error) {
	switch v := body.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case io.Reader:
		return io.ReadAll(v)
	default:
		return jsonCodec.Marshal(body)
	}
}

// setHeaders sets the headers on the HTTP request.
func (p *Provider) setHeaders(req *http.Request, headers map[string]string) {
	// Set default headers from config
	for k, v := range p.config.Headers {
		req.Header.Set(k, v)
	}

	// Set request-specific headers (override defaults)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Set default Content-Type if not provided and body exists
	if req.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
}

// addTracing adds distributed tracing to the HTTP request.
func (p *Provider) addTracing(ctx context.Context, req *http.Request) {
	// Simplified tracing implementation
	// Can be extended with actual tracing library integration
	req.Header.Set("X-Trace-ID", fmt.Sprintf("trace-%d", time.Now().UnixNano()))
	req.Header.Set("X-Component", "httpclient.nethttp")
}

// updateMetricsStart updates metrics at the start of a request.
func (p *Provider) updateMetricsStart() {
	if !p.config.MetricsEnabled {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.metrics.TotalRequests++
	p.metrics.LastRequestTime = time.Now()
}

// updateMetricsEnd updates metrics at the end of a request.
func (p *Provider) updateMetricsEnd(start time.Time, success bool) {
	if !p.config.MetricsEnabled {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	latency := time.Since(start)

	if success {
		p.metrics.SuccessfulRequests++
	} else {
		p.metrics.FailedRequests++
	}

	// Update average latency using exponential moving average
	if p.metrics.AverageLatency == 0 {
		p.metrics.AverageLatency = latency
	} else {
		// EMA with alpha = 0.1
		p.metrics.AverageLatency = time.Duration(
			0.9*float64(p.metrics.AverageLatency) + 0.1*float64(latency),
		)
	}
}

// TraceableTransport wraps http.Transport to provide detailed request tracing.
type TraceableTransport struct {
	transport *http.Transport
}

// NewTraceableTransport creates a new traceable transport.
func NewTraceableTransport(transport *http.Transport) *TraceableTransport {
	return &TraceableTransport{
		transport: transport,
	}
}

// RoundTrip implements http.RoundTripper interface with detailed tracing.
func (t *TraceableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	trace := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			// Connection acquisition started
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			// Connection acquired
		},
		DNSStart: func(dnsInfo httptrace.DNSStartInfo) {
			// DNS lookup started
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			// DNS lookup completed
		},
		ConnectStart: func(network, addr string) {
			// TCP connection started
		},
		ConnectDone: func(network, addr string, err error) {
			// TCP connection completed
		},
		TLSHandshakeStart: func() {
			// TLS handshake started
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			// TLS handshake completed
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			// Request written to connection
		},
		GotFirstResponseByte: func() {
			// First response byte received
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	return t.transport.RoundTrip(req)
}
