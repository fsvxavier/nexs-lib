// Package httpserver provides automatic provider registration.
package httpserver

import (
	"github.com/fsvxavier/nexs-lib/httpserver/config"
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/fsvxavier/nexs-lib/httpserver/providers/fiber"
	"github.com/fsvxavier/nexs-lib/httpserver/providers/nethttp"
)

// init automatically registers available providers
func init() {
	// Register Fiber as default provider (highest priority)
	if err := RegisterProvider("fiber", fiber.NewFactory()); err != nil {
		panic("Failed to register fiber provider: " + err.Error())
	}

	// Register net/http provider as fallback
	if err := RegisterProvider("nethttp", nethttp.NewFactory()); err != nil {
		panic("Failed to register nethttp provider: " + err.Error())
	}
}

// CreateDefaultServer creates a server using the default provider (Fiber).
func CreateDefaultServer(options ...config.Option) (interfaces.HTTPServer, error) {
	return CreateServer("fiber", options...)
}

// CreateDefaultServerWithConfig creates a server using the default provider with pre-built config.
func CreateDefaultServerWithConfig(cfg *config.BaseConfig) (interfaces.HTTPServer, error) {
	return CreateServerWithConfig("fiber", cfg)
}
