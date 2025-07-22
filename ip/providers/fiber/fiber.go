// Package fiber provides HTTP framework adapter for the Fiber framework.
package fiber

import (
	"errors"

	"github.com/fsvxavier/nexs-lib/ip/interfaces"
)

// Adapter implements RequestAdapter for the Fiber framework
type Adapter struct {
	ctx Context
}

// Context defines the interface for Fiber context to avoid import cycle
type Context interface {
	Get(key string, defaultValue ...string) string
	GetReqHeaders() map[string]string
	IP() string
	Method() string
	Path() string
}

// NewAdapter creates a new adapter for Fiber requests
func NewAdapter(ctx Context) *Adapter {
	return &Adapter{ctx: ctx}
}

// GetHeader returns the value of the specified header
func (a *Adapter) GetHeader(key string) string {
	if a.ctx == nil {
		return ""
	}
	return a.ctx.Get(key)
}

// GetAllHeaders returns all headers as a map
func (a *Adapter) GetAllHeaders() map[string]string {
	if a.ctx == nil {
		return make(map[string]string)
	}
	return a.ctx.GetReqHeaders()
}

// GetRemoteAddr returns the remote address of the connection
func (a *Adapter) GetRemoteAddr() string {
	if a.ctx == nil {
		return ""
	}
	return a.ctx.IP()
}

// GetMethod returns the HTTP method
func (a *Adapter) GetMethod() string {
	if a.ctx == nil {
		return ""
	}
	return a.ctx.Method()
}

// GetPath returns the request path
func (a *Adapter) GetPath() string {
	if a.ctx == nil {
		return ""
	}
	return a.ctx.Path()
}

// ProviderFactory implements ProviderFactory for Fiber
type ProviderFactory struct{}

// NewProviderFactory creates a new factory for Fiber providers
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{}
}

// CreateAdapter creates a new request adapter for Fiber requests
func (f *ProviderFactory) CreateAdapter(request interface{}) (interfaces.RequestAdapter, error) {
	ctx, ok := request.(Context)
	if !ok {
		return nil, errors.New("request must implement fiber.Context interface")
	}
	return NewAdapter(ctx), nil
}

// GetProviderName returns the name of the provider
func (f *ProviderFactory) GetProviderName() string {
	return "fiber"
}

// SupportsType checks if the provider supports the given request type
func (f *ProviderFactory) SupportsType(request interface{}) bool {
	_, ok := request.(Context)
	return ok
}
