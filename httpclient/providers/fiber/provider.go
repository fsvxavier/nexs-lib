// Package fiber provides HTTP client implementation using the Fiber framework's client.
package fiber

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/httpclient/interfaces"
	"github.com/gofiber/fiber/v2"
	jsoniter "github.com/json-iterator/go"
)

var (
	jsonCodec = jsoniter.ConfigCompatibleWithStandardLibrary
)

// Provider implements the HTTP provider using Fiber's client.
type Provider struct {
	client  *fiber.Client
	config  *interfaces.Config
	metrics *interfaces.ProviderMetrics
	mu      sync.RWMutex
}

// NewProvider creates a new Fiber provider instance.
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
	return "fiber"
}

// Version returns the provider version.
func (p *Provider) Version() string {
	return "1.0.0"
}

// Configure sets up the Fiber client with the provided configuration.
func (p *Provider) Configure(config *interfaces.Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	p.config = config

	// Create Fiber client
	p.client = &fiber.Client{
		JSONEncoder: jsonCodec.Marshal,
		JSONDecoder: jsonCodec.Unmarshal,
	}

	// Configure client settings based on config
	agent := p.client.Get("")
	if config.Timeout > 0 {
		agent.Timeout(config.Timeout)
	}

	if config.InsecureSkipVerify {
		agent.TLSConfig(nil) // Fiber handles this differently
	}

	// Set default headers
	for k, v := range config.Headers {
		agent.Set(k, v)
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

// DoRequest performs an HTTP request using the Fiber client.
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

	// Create Fiber agent
	agent := p.client.Get(url)

	// Set method
	switch strings.ToUpper(req.Method) {
	case "GET":
		agent = p.client.Get(url)
	case "POST":
		agent = p.client.Post(url)
	case "PUT":
		agent = p.client.Put(url)
	case "DELETE":
		agent = p.client.Delete(url)
	case "PATCH":
		agent = p.client.Patch(url)
	case "HEAD":
		agent = p.client.Head(url)
	default:
		p.updateMetricsEnd(start, false)
		return nil, fmt.Errorf("unsupported HTTP method: %s", req.Method)
	}

	// Set headers
	p.setHeaders(agent, req.Headers)

	// Set request body
	if req.Body != nil {
		bodyBytes, err := p.marshalBody(req.Body)
		if err != nil {
			p.updateMetricsEnd(start, false)
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		agent.Body(bodyBytes)
	}

	// Add tracing if enabled
	if p.config.TracingEnabled {
		p.addTracing(ctx, agent)
	}

	// Set timeout
	if req.Timeout > 0 {
		agent.Timeout(req.Timeout)
	} else if p.config.Timeout > 0 {
		agent.Timeout(p.config.Timeout)
	}

	// Perform request
	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		p.updateMetricsEnd(start, false)
		return nil, fmt.Errorf("Fiber request failed: %v", errs)
	}

	latency := time.Since(start)
	isSuccess := statusCode >= 200 && statusCode < 300

	p.updateMetricsEnd(start, isSuccess)

	// Build response
	response := &interfaces.Response{
		StatusCode: statusCode,
		Body:       body,
		Headers:    make(map[string]string), // Fiber doesn't easily expose response headers
		IsError:    !isSuccess,
		Latency:    latency,
		Request:    req,
	}

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

// setHeaders sets the headers on the Fiber agent.
func (p *Provider) setHeaders(agent *fiber.Agent, headers map[string]string) {
	// Set default headers from config (already set during Configure)

	// Set request-specific headers (override defaults)
	for k, v := range headers {
		agent.Set(k, v)
	}

	// Set default Content-Type if not provided and this is not a GET/HEAD request
	if agent.Request().Header.ContentType() == nil {
		agent.Set("Content-Type", "application/json")
	}
}

// addTracing adds distributed tracing to the Fiber agent.
func (p *Provider) addTracing(ctx context.Context, agent *fiber.Agent) {
	// Simplified tracing implementation
	// Can be extended with actual tracing library integration
	agent.Set("X-Trace-ID", fmt.Sprintf("trace-%d", time.Now().UnixNano()))
	agent.Set("X-Component", "httpclient.fiber")
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
