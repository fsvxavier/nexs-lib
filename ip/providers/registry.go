// Package providers contains all HTTP framework providers and the main factory.
package providers

import (
	"errors"
	"fmt"

	"github.com/fsvxavier/nexs-lib/ip/interfaces"
	"github.com/fsvxavier/nexs-lib/ip/providers/atreugo"
	"github.com/fsvxavier/nexs-lib/ip/providers/echo"
	"github.com/fsvxavier/nexs-lib/ip/providers/fasthttp"
	"github.com/fsvxavier/nexs-lib/ip/providers/fiber"
	"github.com/fsvxavier/nexs-lib/ip/providers/gin"
	"github.com/fsvxavier/nexs-lib/ip/providers/nethttp"
)

// ProviderRegistry manages all available HTTP framework providers
type ProviderRegistry struct {
	providers []interfaces.ProviderFactory
}

// NewProviderRegistry creates a new provider registry with all available providers
func NewProviderRegistry() *ProviderRegistry {
	registry := &ProviderRegistry{
		providers: make([]interfaces.ProviderFactory, 0),
	}

	// Register all available providers
	registry.RegisterProvider(nethttp.NewProviderFactory())
	registry.RegisterProvider(fiber.NewProviderFactory())
	registry.RegisterProvider(gin.NewProviderFactory())
	registry.RegisterProvider(echo.NewProviderFactory())
	registry.RegisterProvider(fasthttp.NewProviderFactory())
	registry.RegisterProvider(atreugo.NewProviderFactory())

	return registry
}

// RegisterProvider registers a new provider factory
func (r *ProviderRegistry) RegisterProvider(provider interfaces.ProviderFactory) {
	r.providers = append(r.providers, provider)
}

// CreateAdapter creates an adapter for the given request object
// It tries all registered providers until one supports the request type
func (r *ProviderRegistry) CreateAdapter(request interface{}) (interfaces.RequestAdapter, error) {
	if request == nil {
		return nil, errors.New("request cannot be nil")
	}

	for _, provider := range r.providers {
		if provider.SupportsType(request) {
			return provider.CreateAdapter(request)
		}
	}

	return nil, fmt.Errorf("no provider found for request type: %T", request)
}

// GetSupportedProviders returns a list of all registered provider names
func (r *ProviderRegistry) GetSupportedProviders() []string {
	providers := make([]string, len(r.providers))
	for i, provider := range r.providers {
		providers[i] = provider.GetProviderName()
	}
	return providers
}

// GetProviderByName returns a specific provider by name
func (r *ProviderRegistry) GetProviderByName(name string) (interfaces.ProviderFactory, error) {
	for _, provider := range r.providers {
		if provider.GetProviderName() == name {
			return provider, nil
		}
	}
	return nil, fmt.Errorf("provider not found: %s", name)
}

// Default registry instance
var defaultRegistry = NewProviderRegistry()

// CreateAdapter creates an adapter using the default registry
func CreateAdapter(request interface{}) (interfaces.RequestAdapter, error) {
	return defaultRegistry.CreateAdapter(request)
}

// GetSupportedProviders returns supported providers from the default registry
func GetSupportedProviders() []string {
	return defaultRegistry.GetSupportedProviders()
}

// RegisterProvider registers a provider in the default registry
func RegisterProvider(provider interfaces.ProviderFactory) {
	defaultRegistry.RegisterProvider(provider)
}
