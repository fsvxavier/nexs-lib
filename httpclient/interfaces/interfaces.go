// Package interfaces defines contracts for HTTP client functionality.
package interfaces

import (
	"context"
	"time"
)

// Client represents the main HTTP client interface with request capabilities.
type Client interface {
	// Core HTTP methods
	Get(ctx context.Context, endpoint string) (*Response, error)
	Post(ctx context.Context, endpoint string, body interface{}) (*Response, error)
	Put(ctx context.Context, endpoint string, body interface{}) (*Response, error)
	Delete(ctx context.Context, endpoint string) (*Response, error)
	Patch(ctx context.Context, endpoint string, body interface{}) (*Response, error)
	Head(ctx context.Context, endpoint string) (*Response, error)
	Options(ctx context.Context, endpoint string) (*Response, error)

	// Generic execute method
	Execute(ctx context.Context, method, endpoint string, body interface{}) (*Response, error)

	// Configuration methods
	SetHeaders(headers map[string]string) Client
	SetTimeout(timeout time.Duration) Client
	SetErrorHandler(handler ErrorHandler) Client
	SetRetryConfig(config *RetryConfig) Client

	// Result unmarshaling
	Unmarshal(v interface{}) Client
	UnmarshalResponse(resp *Response, v interface{}) error

	// Middleware and hooks
	AddMiddleware(middleware Middleware) Client
	RemoveMiddleware(middleware Middleware) Client
	AddHook(hook Hook) Client
	RemoveHook(hook Hook) Client

	// Batch operations
	Batch() BatchRequestBuilder

	// Streaming support
	Stream(ctx context.Context, method, endpoint string, handler StreamHandler) error

	// Provider access
	GetProvider() Provider
	GetConfig() *Config
	GetID() string

	// Health and metrics
	IsHealthy() bool
	GetMetrics() *ProviderMetrics
}

// Provider defines the interface that all HTTP providers must implement.
type Provider interface {
	// Core provider identification
	Name() string
	Version() string

	// HTTP operations
	DoRequest(ctx context.Context, req *Request) (*Response, error)

	// Configuration
	Configure(config *Config) error
	SetDefaults()

	// Health and status
	IsHealthy() bool
	GetMetrics() *ProviderMetrics
}

// Request represents an HTTP request structure.
type Request struct {
	Method               string
	URL                  string
	Headers              map[string]string
	Body                 interface{}
	Timeout              time.Duration
	Context              context.Context
	ContentType          string
	AcceptEncoding       string
	DisableAutoUnmarshal bool
}

// Response represents an HTTP response structure.
type Response struct {
	StatusCode    int
	Body          []byte
	Headers       map[string]string
	IsError       bool
	Latency       time.Duration
	Request       *Request
	ContentType   string
	ContentLength int64
	IsCompressed  bool
}

// ErrorHandler defines the signature for custom error handling functions.
type ErrorHandler func(*Response) error

// RetryConfig defines retry configuration for failed requests.
type RetryConfig struct {
	MaxRetries      int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	RetryCondition  func(*Response, error) bool
}

// Config represents provider configuration options.
type Config struct {
	BaseURL             string
	Timeout             time.Duration
	MaxIdleConns        int
	IdleConnTimeout     time.Duration
	TLSHandshakeTimeout time.Duration
	DisableKeepAlives   bool
	DisableCompression  bool
	InsecureSkipVerify  bool
	Headers             map[string]string
	RetryConfig         *RetryConfig
	TracingEnabled      bool
	MetricsEnabled      bool

	// New advanced features
	EnableHTTP2       bool
	CompressionTypes  []CompressionType
	CompressionLevel  int // 1-9 for gzip/deflate
	AutoUnmarshal     bool
	UnmarshalStrategy UnmarshalStrategy
	StreamingEnabled  bool
	BatchingEnabled   bool
	MaxBatchSize      int
	BatchTimeout      time.Duration
}

// ProviderMetrics holds provider performance metrics.
type ProviderMetrics struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	AverageLatency     time.Duration
	LastRequestTime    time.Time
}

// Factory defines the interface for creating HTTP clients.
type Factory interface {
	CreateClient(providerType ProviderType, config *Config) (Client, error)
	RegisterProvider(providerType ProviderType, constructor ProviderConstructor) error
	GetAvailableProviders() []ProviderType
}

// ClientManager defines the interface for managing HTTP clients with connection reuse.
type ClientManager interface {
	GetOrCreateClient(name string, providerType ProviderType, cfg *Config) (Client, error)
	GetClient(name string) (Client, bool)
	RemoveClient(name string)
	ListClients() []string
	Shutdown() error
}

// ProviderType represents the type of HTTP provider.
type ProviderType string

const (
	ProviderNetHTTP  ProviderType = "nethttp"
	ProviderFiber    ProviderType = "fiber"
	ProviderFastHTTP ProviderType = "fasthttp"
)

// ProviderConstructor defines the signature for provider constructors.
type ProviderConstructor func(config *Config) (Provider, error)

// Middleware defines the interface for request/response middleware.
type Middleware interface {
	Process(ctx context.Context, req *Request, next func(context.Context, *Request) (*Response, error)) (*Response, error)
}

// Hook defines lifecycle hooks for HTTP operations.
type Hook interface {
	BeforeRequest(ctx context.Context, req *Request) error
	AfterResponse(ctx context.Context, req *Request, resp *Response) error
	OnError(ctx context.Context, req *Request, err error) error
}

// BatchRequestBuilder defines the interface for batch operations.
type BatchRequestBuilder interface {
	Add(method, endpoint string, body interface{}) BatchRequestBuilder
	AddRequest(req *Request) BatchRequestBuilder
	Execute(ctx context.Context) ([]*Response, error)
	ExecuteParallel(ctx context.Context, maxConcurrency int) ([]*Response, error)
}

// StreamHandler defines the handler for streaming responses.
type StreamHandler interface {
	OnData(data []byte) error
	OnError(err error)
	OnComplete()
}

// StreamHandlerFunc is a function adapter for StreamHandler.
type StreamHandlerFunc func(data []byte) error

func (f StreamHandlerFunc) OnData(data []byte) error { return f(data) }
func (f StreamHandlerFunc) OnError(err error)        {}
func (f StreamHandlerFunc) OnComplete()              {}

// CompressionType defines supported compression algorithms.
type CompressionType string

const (
	CompressionGzip    CompressionType = "gzip"
	CompressionDeflate CompressionType = "deflate"
	CompressionBrotli  CompressionType = "br"
)

// UnmarshalStrategy defines how responses should be unmarshaled.
type UnmarshalStrategy string

const (
	UnmarshalAuto UnmarshalStrategy = "auto"
	UnmarshalJSON UnmarshalStrategy = "json"
	UnmarshalXML  UnmarshalStrategy = "xml"
	UnmarshalNone UnmarshalStrategy = "none"
)
