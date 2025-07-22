// Package atreugo provides HTTP framework adapter for the Atreugo framework.
package atreugo

import (
	"errors"

	"github.com/fsvxavier/nexs-lib/ip/interfaces"
)

// Adapter implements RequestAdapter for the Atreugo framework
type Adapter struct {
	ctx RequestContext
}

// RequestContext defines the interface for Atreugo context to avoid import cycle
// Atreugo is built on top of FastHTTP, so it shares similar interfaces
type RequestContext interface {
	Request() Request
	Response() Response
	RemoteIP() []byte
	Method() string
	Path() string
}

// Request defines the interface for Atreugo request to avoid import cycle
type Request interface {
	Header() RequestHeader
	URI() URI
}

// Response defines the interface for Atreugo response to avoid import cycle
type Response interface {
	Header() ResponseHeader
}

// RequestHeader defines the interface for Atreugo request header to avoid import cycle
type RequestHeader interface {
	Peek(key string) []byte
	VisitAll(f func(key, value []byte))
}

// ResponseHeader defines the interface for Atreugo response header to avoid import cycle
type ResponseHeader interface {
	Peek(key string) []byte
}

// URI defines the interface for Atreugo URI to avoid import cycle
type URI interface {
	Path() []byte
}

// NewAdapter creates a new adapter for Atreugo requests
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

	// Atreugo provides RemoteIP() method
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

	// Atreugo provides Method() method directly
	return a.ctx.Method()
}

// GetPath returns the request path
func (a *Adapter) GetPath() string {
	if a.ctx == nil {
		return ""
	}

	// Atreugo provides Path() method directly
	return a.ctx.Path()
}

// ProviderFactory implements ProviderFactory for Atreugo
type ProviderFactory struct{}

// NewProviderFactory creates a new factory for Atreugo providers
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{}
}

// CreateAdapter creates a new request adapter for Atreugo requests
func (f *ProviderFactory) CreateAdapter(request interface{}) (interfaces.RequestAdapter, error) {
	ctx, ok := request.(RequestContext)
	if !ok {
		return nil, errors.New("request must implement atreugo.RequestContext interface")
	}
	return NewAdapter(ctx), nil
}

// GetProviderName returns the name of the provider
func (f *ProviderFactory) GetProviderName() string {
	return "atreugo"
}

// SupportsType checks if the provider supports the given request type
func (f *ProviderFactory) SupportsType(request interface{}) bool {
	_, ok := request.(RequestContext)
	return ok
}
