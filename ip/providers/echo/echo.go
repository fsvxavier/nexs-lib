// Package echo provides HTTP framework adapter for the Echo framework.
package echo

import (
	"errors"

	"github.com/fsvxavier/nexs-lib/ip/interfaces"
)

// Adapter implements RequestAdapter for the Echo framework
type Adapter struct {
	ctx Context
}

// Context defines the interface for Echo context to avoid import cycle
type Context interface {
	Request() Request
	RealIP() string
}

// Request defines the interface for Echo request to avoid import cycle
type Request interface {
	Header() Header
	Method() string
	URL() URL
	RemoteAddr() string
}

// Header defines the interface for Echo header to avoid import cycle
type Header interface {
	Get(key string) string
	Values(key string) []string
}

// URL defines the interface for Echo URL to avoid import cycle
type URL interface {
	Path() string
}

// NewAdapter creates a new adapter for Echo requests
func NewAdapter(ctx Context) *Adapter {
	return &Adapter{ctx: ctx}
}

// GetHeader returns the value of the specified header
func (a *Adapter) GetHeader(key string) string {
	if a.ctx == nil {
		return ""
	}

	req := a.ctx.Request()
	if req == nil {
		return ""
	}

	header := req.Header()
	if header == nil {
		return ""
	}

	return header.Get(key)
}

// GetAllHeaders returns all headers as a map
func (a *Adapter) GetAllHeaders() map[string]string {
	headers := make(map[string]string)
	if a.ctx == nil {
		return headers
	}

	req := a.ctx.Request()
	if req == nil {
		return headers
	}

	header := req.Header()
	if header == nil {
		return headers
	}

	// Echo doesn't expose all headers directly, so we'll use common ones
	commonHeaders := []string{
		"X-Forwarded-For", "X-Real-IP", "CF-Connecting-IP", "True-Client-IP",
		"X-Client-IP", "X-Cluster-Client-IP", "X-Forwarded", "Forwarded-For",
		"Forwarded", "X-Original-Forwarded-For", "X-Azure-ClientIP", "X-Google-Real-IP",
		"User-Agent", "Accept", "Content-Type", "Authorization",
	}

	for _, key := range commonHeaders {
		if value := header.Get(key); value != "" {
			headers[key] = value
		}
	}

	return headers
}

// GetRemoteAddr returns the remote address of the connection
func (a *Adapter) GetRemoteAddr() string {
	if a.ctx == nil {
		return ""
	}

	// First try Echo's RealIP which handles X-Forwarded-For, etc.
	if ip := a.ctx.RealIP(); ip != "" {
		return ip
	}

	// Fallback to raw RemoteAddr
	req := a.ctx.Request()
	if req != nil {
		return req.RemoteAddr()
	}

	return ""
}

// GetMethod returns the HTTP method
func (a *Adapter) GetMethod() string {
	if a.ctx == nil {
		return ""
	}

	req := a.ctx.Request()
	if req == nil {
		return ""
	}

	return req.Method()
}

// GetPath returns the request path
func (a *Adapter) GetPath() string {
	if a.ctx == nil {
		return ""
	}

	req := a.ctx.Request()
	if req == nil {
		return ""
	}

	url := req.URL()
	if url == nil {
		return ""
	}

	return url.Path()
}

// ProviderFactory implements ProviderFactory for Echo
type ProviderFactory struct{}

// NewProviderFactory creates a new factory for Echo providers
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{}
}

// CreateAdapter creates a new request adapter for Echo requests
func (f *ProviderFactory) CreateAdapter(request interface{}) (interfaces.RequestAdapter, error) {
	ctx, ok := request.(Context)
	if !ok {
		return nil, errors.New("request must implement echo.Context interface")
	}
	return NewAdapter(ctx), nil
}

// GetProviderName returns the name of the provider
func (f *ProviderFactory) GetProviderName() string {
	return "echo"
}

// SupportsType checks if the provider supports the given request type
func (f *ProviderFactory) SupportsType(request interface{}) bool {
	_, ok := request.(Context)
	return ok
}
