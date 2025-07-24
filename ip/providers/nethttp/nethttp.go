// Package nethttp provides HTTP framework adapter for the standard net/http package.
package nethttp

import (
	"errors"
	"net/http"

	"github.com/fsvxavier/nexs-lib/ip/interfaces"
)

// Adapter implements RequestAdapter for the standard net/http package
type Adapter struct {
	request *http.Request
}

// NewAdapter creates a new adapter for net/http requests
func NewAdapter(r *http.Request) *Adapter {
	return &Adapter{request: r}
}

// GetHeader returns the value of the specified header
func (a *Adapter) GetHeader(key string) string {
	if a.request == nil {
		return ""
	}
	return a.request.Header.Get(key)
}

// GetAllHeaders returns all headers as a map
func (a *Adapter) GetAllHeaders() map[string]string {
	headers := make(map[string]string)
	if a.request == nil {
		return headers
	}

	for key, values := range a.request.Header {
		if len(values) > 0 {
			headers[key] = values[0] // Take the first value
		}
	}
	return headers
}

// GetRemoteAddr returns the remote address of the connection
func (a *Adapter) GetRemoteAddr() string {
	if a.request == nil {
		return ""
	}
	return a.request.RemoteAddr
}

// GetMethod returns the HTTP method
func (a *Adapter) GetMethod() string {
	if a.request == nil {
		return ""
	}
	return a.request.Method
}

// GetPath returns the request path
func (a *Adapter) GetPath() string {
	if a.request == nil {
		return ""
	}
	return a.request.URL.Path
}

// ProviderFactory implements ProviderFactory for net/http
type ProviderFactory struct{}

// NewProviderFactory creates a new factory for net/http providers
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{}
}

// CreateAdapter creates a new request adapter for net/http requests
func (f *ProviderFactory) CreateAdapter(request interface{}) (interfaces.RequestAdapter, error) {
	httpReq, ok := request.(*http.Request)
	if !ok {
		return nil, errors.New("request must be of type *http.Request")
	}
	return NewAdapter(httpReq), nil
}

// GetProviderName returns the name of the provider
func (f *ProviderFactory) GetProviderName() string {
	return "net/http"
}

// SupportsType checks if the provider supports the given request type
func (f *ProviderFactory) SupportsType(request interface{}) bool {
	_, ok := request.(*http.Request)
	return ok
}
