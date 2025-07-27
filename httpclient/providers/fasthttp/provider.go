// Package fasthttp provides HTTP client implementation using the FastHTTP library.
package fasthttp

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

var (
	jsonCodec = jsoniter.ConfigCompatibleWithStandardLibrary
)

// Provider implements the HTTP provider using FastHTTP.
type Provider struct {
	client  *fasthttp.Client
	config  *interfaces.Config
	metrics *interfaces.ProviderMetrics
	mu      sync.RWMutex
}

// NewProvider creates a new FastHTTP provider instance.
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
	return "fasthttp"
}

// Version returns the provider version.
func (p *Provider) Version() string {
	return "1.0.0"
}

// Configure sets up the FastHTTP client with the provided configuration.
func (p *Provider) Configure(config *interfaces.Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	p.config = config

	// Create FastHTTP client with configuration
	p.client = &fasthttp.Client{
		ReadTimeout:                   config.Timeout,
		WriteTimeout:                  config.Timeout,
		MaxIdleConnDuration:           config.IdleConnTimeout,
		DisableHeaderNamesNormalizing: false,
		DisablePathNormalizing:        false,
		MaxConnsPerHost:               config.MaxIdleConns,
	}

	// Configure TLS settings
	if config.InsecureSkipVerify {
		// FastHTTP uses Dial function for TLS configuration
		p.client.Dial = fasthttp.DialFunc(func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, config.Timeout)
		})
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

// DoRequest performs an HTTP request using the FastHTTP client.
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

	// Create FastHTTP request and response
	fastReq := fasthttp.AcquireRequest()
	fastResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(fastReq)
	defer fasthttp.ReleaseResponse(fastResp)

	// Set method and URL
	fastReq.SetRequestURI(url)
	fastReq.Header.SetMethod(req.Method)

	// Set headers
	p.setHeaders(fastReq, req.Headers)

	// Set request body
	if req.Body != nil {
		bodyBytes, err := p.marshalBody(req.Body)
		if err != nil {
			p.updateMetricsEnd(start, false)
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		fastReq.SetBody(bodyBytes)
	}

	// Add tracing if enabled
	if p.config.TracingEnabled {
		p.addTracing(ctx, fastReq)
	}

	// Set timeout
	timeout := p.config.Timeout
	if req.Timeout > 0 {
		timeout = req.Timeout
	}

	// Perform request
	err := p.client.DoTimeout(fastReq, fastResp, timeout)
	if err != nil {
		p.updateMetricsEnd(start, false)
		return nil, fmt.Errorf("FastHTTP request failed: %w", err)
	}

	latency := time.Since(start)
	statusCode := fastResp.StatusCode()
	isSuccess := statusCode >= 200 && statusCode < 300

	p.updateMetricsEnd(start, isSuccess)

	// Build response
	response := &interfaces.Response{
		StatusCode: statusCode,
		Body:       fastResp.Body(),
		Headers:    make(map[string]string),
		IsError:    !isSuccess,
		Latency:    latency,
		Request:    req,
	}

	// Copy response headers
	fastResp.Header.VisitAll(func(key, value []byte) {
		response.Headers[string(key)] = string(value)
	})

	return response, nil
}

// marshalBody marshals the request body to bytes.
func (p *Provider) marshalBody(body interface{}) ([]byte, error) {
	switch v := body.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return jsonCodec.Marshal(body)
	}
}

// setHeaders sets the headers on the FastHTTP request.
func (p *Provider) setHeaders(req *fasthttp.Request, headers map[string]string) {
	// Set default headers from config
	for k, v := range p.config.Headers {
		req.Header.Set(k, v)
	}

	// Set request-specific headers (override defaults)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Set default Content-Type if not provided and body exists
	if len(req.Body()) > 0 && len(req.Header.ContentType()) == 0 {
		req.Header.SetContentType("application/json")
	}
}

// addTracing adds distributed tracing to the FastHTTP request.
func (p *Provider) addTracing(ctx context.Context, req *fasthttp.Request) {
	// Simplified tracing implementation
	// Can be extended with actual tracing library integration
	req.Header.Set("X-Trace-ID", fmt.Sprintf("trace-%d", time.Now().UnixNano()))
	req.Header.Set("X-Component", "httpclient.fasthttp")
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
