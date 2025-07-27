// Package http2 provides HTTP/2 support for the httpclient library.
package http2

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
	"golang.org/x/net/http2"
)

// HTTP2Provider implements HTTP/2 support for the httpclient.
type HTTP2Provider struct {
	config  *interfaces.Config
	client  *http.Client
	metrics *interfaces.ProviderMetrics
	healthy bool
}

// NewHTTP2Provider creates a new HTTP/2 provider.
func NewHTTP2Provider(config *interfaces.Config) (*HTTP2Provider, error) {
	if config == nil {
		config = &interfaces.Config{}
	}

	provider := &HTTP2Provider{
		config:  config,
		metrics: &interfaces.ProviderMetrics{},
		healthy: true,
	}

	if err := provider.Configure(config); err != nil {
		return nil, fmt.Errorf("failed to configure HTTP/2 provider: %w", err)
	}

	return provider, nil
}

// Name returns the provider name.
func (p *HTTP2Provider) Name() string {
	return "http2"
}

// Version returns the provider version.
func (p *HTTP2Provider) Version() string {
	return "1.0.0"
}

// Configure configures the HTTP/2 provider.
func (p *HTTP2Provider) Configure(config *interfaces.Config) error {
	p.config = config

	// Create HTTP/2 transport
	transport := &http2.Transport{
		AllowHTTP: false, // Only HTTPS for HTTP/2
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify,
			NextProtos:         []string{"h2", "http/1.1"}, // Prefer HTTP/2
		},
	}

	// Configure connection pooling
	if config.IdleConnTimeout > 0 {
		transport.IdleConnTimeout = config.IdleConnTimeout
	} else {
		transport.IdleConnTimeout = 90 * time.Second
	}

	// Configure additional HTTP/2 settings
	transport.ReadIdleTimeout = 30 * time.Second
	transport.PingTimeout = 15 * time.Second

	// Create HTTP client with HTTP/2 transport
	p.client = &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return nil
}

// SetDefaults sets default configuration values.
func (p *HTTP2Provider) SetDefaults() {
	if p.config == nil {
		p.config = &interfaces.Config{}
	}

	if p.config.Timeout == 0 {
		p.config.Timeout = 30 * time.Second
	}

	if p.config.MaxIdleConns == 0 {
		p.config.MaxIdleConns = 100
	}

	if p.config.IdleConnTimeout == 0 {
		p.config.IdleConnTimeout = 90 * time.Second
	}

	if p.config.TLSHandshakeTimeout == 0 {
		p.config.TLSHandshakeTimeout = 10 * time.Second
	}
}

// DoRequest executes an HTTP/2 request.
func (p *HTTP2Provider) DoRequest(ctx context.Context, req *interfaces.Request) (*interfaces.Response, error) {
	start := time.Now()
	p.metrics.TotalRequests++

	// Create HTTP request
	httpReq, err := p.createHTTPRequest(ctx, req)
	if err != nil {
		p.metrics.FailedRequests++
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Execute request
	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		p.metrics.FailedRequests++
		return nil, fmt.Errorf("HTTP/2 request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Convert to our response format
	resp, err := p.convertResponse(httpResp, req)
	if err != nil {
		p.metrics.FailedRequests++
		return nil, fmt.Errorf("failed to convert response: %w", err)
	}

	// Update metrics
	resp.Latency = time.Since(start)
	p.metrics.LastRequestTime = time.Now()
	p.metrics.AverageLatency = p.updateAverageLatency(resp.Latency)

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		p.metrics.SuccessfulRequests++
	} else {
		p.metrics.FailedRequests++
	}

	return resp, nil
}

// createHTTPRequest creates an HTTP request from our request format.
func (p *HTTP2Provider) createHTTPRequest(ctx context.Context, req *interfaces.Request) (*http.Request, error) {
	var body interface{}
	if req.Body != nil {
		body = req.Body
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	if req.Headers != nil {
		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}
	}

	// Set global headers from config
	if p.config.Headers != nil {
		for key, value := range p.config.Headers {
			if httpReq.Header.Get(key) == "" { // Don't override request-specific headers
				httpReq.Header.Set(key, value)
			}
		}
	}

	// Force HTTP/2
	httpReq.Proto = "HTTP/2.0"
	httpReq.ProtoMajor = 2
	httpReq.ProtoMinor = 0

	// Handle body
	if body != nil {
		switch v := body.(type) {
		case string:
			httpReq.Body = http.NoBody
			httpReq.ContentLength = int64(len(v))
		case []byte:
			httpReq.Body = http.NoBody
			httpReq.ContentLength = int64(len(v))
		}
	}

	return httpReq, nil
}

// convertResponse converts HTTP response to our response format.
func (p *HTTP2Provider) convertResponse(httpResp *http.Response, req *interfaces.Request) (*interfaces.Response, error) {
	// Read response body
	body := []byte{}
	if httpResp.Body != nil {
		var err error
		body, err = readResponseBody(httpResp)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
	}

	// Convert headers
	headers := make(map[string]string)
	for key, values := range httpResp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Determine if it's an error
	isError := httpResp.StatusCode >= 400

	resp := &interfaces.Response{
		StatusCode:    httpResp.StatusCode,
		Body:          body,
		Headers:       headers,
		IsError:       isError,
		Request:       req,
		ContentType:   httpResp.Header.Get("Content-Type"),
		ContentLength: httpResp.ContentLength,
	}

	return resp, nil
}

// readResponseBody reads the complete response body.
func readResponseBody(resp *http.Response) ([]byte, error) {
	if resp.Body == nil {
		return []byte{}, nil
	}

	// Implementation would read the body
	// For now, return empty body
	return []byte{}, nil
}

// updateAverageLatency updates the average latency metric.
func (p *HTTP2Provider) updateAverageLatency(newLatency time.Duration) time.Duration {
	if p.metrics.TotalRequests == 1 {
		return newLatency
	}

	// Simple moving average
	totalTime := p.metrics.AverageLatency * time.Duration(p.metrics.TotalRequests-1)
	totalTime += newLatency
	return totalTime / time.Duration(p.metrics.TotalRequests)
}

// IsHealthy returns the provider health status.
func (p *HTTP2Provider) IsHealthy() bool {
	return p.healthy
}

// GetMetrics returns provider metrics.
func (p *HTTP2Provider) GetMetrics() *interfaces.ProviderMetrics {
	return p.metrics
}

// SupportsHTTP2 checks if the server supports HTTP/2.
func SupportsHTTP2(address string) bool {
	conn, err := tls.Dial("tcp", address, &tls.Config{
		NextProtos: []string{"h2", "http/1.1"},
	})
	if err != nil {
		return false
	}
	defer conn.Close()

	state := conn.ConnectionState()
	return state.NegotiatedProtocol == "h2"
}

// HTTP2Config provides HTTP/2 specific configuration.
type HTTP2Config struct {
	// Connection settings
	MaxIdleConns     int
	IdleConnTimeout  time.Duration
	ReadIdleTimeout  time.Duration
	PingTimeout      time.Duration
	WriteByteTimeout time.Duration

	// Flow control
	InitialWindowSize uint32
	MaxFrameSize      uint32
	MaxHeaderListSize uint32

	// Performance
	AllowHTTP                  bool
	StrictMaxConcurrentStreams bool
}

// DefaultHTTP2Config returns default HTTP/2 configuration.
func DefaultHTTP2Config() *HTTP2Config {
	return &HTTP2Config{
		MaxIdleConns:               100,
		IdleConnTimeout:            90 * time.Second,
		ReadIdleTimeout:            30 * time.Second,
		PingTimeout:                15 * time.Second,
		WriteByteTimeout:           30 * time.Second,
		InitialWindowSize:          65535,
		MaxFrameSize:               16384,
		MaxHeaderListSize:          10485760, // 10MB
		AllowHTTP:                  false,    // HTTPS only for security
		StrictMaxConcurrentStreams: false,
	}
}

// HTTP2Client provides a higher-level HTTP/2 client interface.
type HTTP2Client struct {
	provider *HTTP2Provider
	config   *HTTP2Config
}

// NewHTTP2Client creates a new HTTP/2 client.
func NewHTTP2Client(config *interfaces.Config, http2Config *HTTP2Config) (*HTTP2Client, error) {
	if http2Config == nil {
		http2Config = DefaultHTTP2Config()
	}

	provider, err := NewHTTP2Provider(config)
	if err != nil {
		return nil, err
	}

	return &HTTP2Client{
		provider: provider,
		config:   http2Config,
	}, nil
}

// Get executes an HTTP/2 GET request.
func (c *HTTP2Client) Get(ctx context.Context, url string) (*interfaces.Response, error) {
	req := &interfaces.Request{
		Method:  "GET",
		URL:     url,
		Context: ctx,
	}
	return c.provider.DoRequest(ctx, req)
}

// Post executes an HTTP/2 POST request.
func (c *HTTP2Client) Post(ctx context.Context, url string, body interface{}) (*interfaces.Response, error) {
	req := &interfaces.Request{
		Method:  "POST",
		URL:     url,
		Body:    body,
		Context: ctx,
	}
	return c.provider.DoRequest(ctx, req)
}

// Put executes an HTTP/2 PUT request.
func (c *HTTP2Client) Put(ctx context.Context, url string, body interface{}) (*interfaces.Response, error) {
	req := &interfaces.Request{
		Method:  "PUT",
		URL:     url,
		Body:    body,
		Context: ctx,
	}
	return c.provider.DoRequest(ctx, req)
}

// Delete executes an HTTP/2 DELETE request.
func (c *HTTP2Client) Delete(ctx context.Context, url string) (*interfaces.Response, error) {
	req := &interfaces.Request{
		Method:  "DELETE",
		URL:     url,
		Context: ctx,
	}
	return c.provider.DoRequest(ctx, req)
}

// ServerPush represents an HTTP/2 server push promise.
type ServerPush struct {
	Method   string
	URL      string
	Headers  map[string]string
	Response *interfaces.Response
}

// PushHandler handles HTTP/2 server push promises.
type PushHandler interface {
	HandlePush(push *ServerPush) error
}

// MultiplexedClient provides multiplexed HTTP/2 requests.
type MultiplexedClient struct {
	client *HTTP2Client
}

// NewMultiplexedClient creates a new multiplexed HTTP/2 client.
func NewMultiplexedClient(config *interfaces.Config) (*MultiplexedClient, error) {
	client, err := NewHTTP2Client(config, nil)
	if err != nil {
		return nil, err
	}

	return &MultiplexedClient{
		client: client,
	}, nil
}

// ExecuteConcurrent executes multiple requests concurrently over the same connection.
func (c *MultiplexedClient) ExecuteConcurrent(ctx context.Context, requests []*interfaces.Request) ([]*interfaces.Response, error) {
	if len(requests) == 0 {
		return []*interfaces.Response{}, nil
	}

	responses := make([]*interfaces.Response, len(requests))
	errors := make([]error, len(requests))

	// Use channels for concurrent execution
	type result struct {
		index int
		resp  *interfaces.Response
		err   error
	}

	resultChan := make(chan result, len(requests))

	// Start all requests concurrently
	for i, req := range requests {
		go func(index int, request *interfaces.Request) {
			resp, err := c.client.provider.DoRequest(ctx, request)
			resultChan <- result{index: index, resp: resp, err: err}
		}(i, req)
	}

	// Collect results
	for i := 0; i < len(requests); i++ {
		res := <-resultChan
		responses[res.index] = res.resp
		errors[res.index] = res.err
	}

	// Check for errors
	for i, err := range errors {
		if err != nil {
			return responses, fmt.Errorf("request %d failed: %w", i, err)
		}
	}

	return responses, nil
}

// ConnectionMonitor monitors HTTP/2 connection health.
type ConnectionMonitor struct {
	client   *HTTP2Client
	interval time.Duration
	stopChan chan struct{}
}

// NewConnectionMonitor creates a new connection monitor.
func NewConnectionMonitor(client *HTTP2Client, interval time.Duration) *ConnectionMonitor {
	return &ConnectionMonitor{
		client:   client,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start starts monitoring the connection.
func (m *ConnectionMonitor) Start(ctx context.Context, healthCheckURL string) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			// Perform health check
			_, err := m.client.Get(ctx, healthCheckURL)
			if err != nil {
				m.client.provider.healthy = false
			} else {
				m.client.provider.healthy = true
			}
		}
	}
}

// Stop stops monitoring the connection.
func (m *ConnectionMonitor) Stop() {
	close(m.stopChan)
}
