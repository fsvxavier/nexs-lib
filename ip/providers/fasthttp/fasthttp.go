// Package fasthttp provides HTTP framework adapter for the FastHTTP framework.
package fasthttp

import (
	"errors"

	"github.com/fsvxavier/nexs-lib/ip/interfaces"
)

// Adapter implements RequestAdapter for the FastHTTP framework
type Adapter struct {
	ctx RequestContext
}

// RequestContext defines the interface for FastHTTP context to avoid import cycle
type RequestContext interface {
	Request() Request
	Response() Response
	RemoteIP() []byte
}

// Request defines the interface for FastHTTP request to avoid import cycle
type Request interface {
	Header() RequestHeader
	URI() URI
}

// Response defines the interface for FastHTTP response to avoid import cycle
type Response interface {
	Header() ResponseHeader
}

// RequestHeader defines the interface for FastHTTP request header to avoid import cycle
type RequestHeader interface {
	Peek(key string) []byte
	VisitAll(f func(key, value []byte))
}

// ResponseHeader defines the interface for FastHTTP response header to avoid import cycle
type ResponseHeader interface {
	Peek(key string) []byte
}

// URI defines the interface for FastHTTP URI to avoid import cycle
type URI interface {
	Path() []byte
}

// NewAdapter creates a new adapter for FastHTTP requests
func NewAdapter(ctx RequestContext) *Adapter {
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

	value := header.Peek(key)
	return string(value)
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

	header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})

	return headers
}

// GetRemoteAddr returns the remote address of the connection
func (a *Adapter) GetRemoteAddr() string {
	if a.ctx == nil {
		return ""
	}

	// FastHTTP provides RemoteIP() method
	if ip := a.ctx.RemoteIP(); len(ip) > 0 {
		return string(ip)
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

	// FastHTTP doesn't expose Method() directly in our interface
	// We'll try to get it from a common header or return empty
	header := req.Header()
	if header != nil {
		if method := header.Peek("Method"); len(method) > 0 {
			return string(method)
		}
	}

	return ""
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

	uri := req.URI()
	if uri == nil {
		return ""
	}

	path := uri.Path()
	return string(path)
}

// ProviderFactory implements ProviderFactory for FastHTTP
type ProviderFactory struct{}

// NewProviderFactory creates a new factory for FastHTTP providers
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{}
}

// CreateAdapter creates a new request adapter for FastHTTP requests
func (f *ProviderFactory) CreateAdapter(request interface{}) (interfaces.RequestAdapter, error) {
	ctx, ok := request.(RequestContext)
	if !ok {
		return nil, errors.New("request must implement fasthttp.RequestContext interface")
	}
	return NewAdapter(ctx), nil
}

// GetProviderName returns the name of the provider
func (f *ProviderFactory) GetProviderName() string {
	return "fasthttp"
}

// SupportsType checks if the provider supports the given request type
func (f *ProviderFactory) SupportsType(request interface{}) bool {
	_, ok := request.(RequestContext)
	return ok
}
