// Package gin provides HTTP framework adapter for the Gin framework.
package gin

import (
	"errors"

	"github.com/fsvxavier/nexs-lib/ip/interfaces"
)

// Adapter implements RequestAdapter for the Gin framework
type Adapter struct {
	ctx Context
}

// Context defines the interface for Gin context to avoid import cycle
type Context interface {
	GetHeader(key string) string
	Request() Request
	ClientIP() string
}

// Request defines the interface for Gin request to avoid import cycle
type Request interface {
	Header() map[string][]string
	Method() string
	URL() URL
	RemoteAddr() string
}

// URL defines the interface for Gin URL to avoid import cycle
type URL interface {
	Path() string
}

// NewAdapter creates a new adapter for Gin requests
func NewAdapter(ctx Context) *Adapter {
	return &Adapter{ctx: ctx}
}

// GetHeader returns the value of the specified header
func (a *Adapter) GetHeader(key string) string {
	if a.ctx == nil {
		return ""
	}
	return a.ctx.GetHeader(key)
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

	headerMap := req.Header()
	for key, values := range headerMap {
		if len(values) > 0 {
			headers[key] = values[0] // Take the first value
		}
	}
	return headers
}

// GetRemoteAddr returns the remote address of the connection
func (a *Adapter) GetRemoteAddr() string {
	if a.ctx == nil {
		return ""
	}

	// First try Gin's ClientIP which handles X-Forwarded-For, etc.
	if ip := a.ctx.ClientIP(); ip != "" {
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

// ProviderFactory implements ProviderFactory for Gin
type ProviderFactory struct{}

// NewProviderFactory creates a new factory for Gin providers
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{}
}

// CreateAdapter creates a new request adapter for Gin requests
func (f *ProviderFactory) CreateAdapter(request interface{}) (interfaces.RequestAdapter, error) {
	ctx, ok := request.(Context)
	if !ok {
		return nil, errors.New("request must implement gin.Context interface")
	}
	return NewAdapter(ctx), nil
}

// GetProviderName returns the name of the provider
func (f *ProviderFactory) GetProviderName() string {
	return "gin"
}

// SupportsType checks if the provider supports the given request type
func (f *ProviderFactory) SupportsType(request interface{}) bool {
	_, ok := request.(Context)
	return ok
}
