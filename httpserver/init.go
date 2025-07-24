package httpserver

import (
	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
	"github.com/fsvxavier/nexs-lib/httpserver/providers/atreugo"
	"github.com/fsvxavier/nexs-lib/httpserver/providers/echo"
	"github.com/fsvxavier/nexs-lib/httpserver/providers/fasthttp"
	"github.com/fsvxavier/nexs-lib/httpserver/providers/fiber"
	"github.com/fsvxavier/nexs-lib/httpserver/providers/gin"
	"github.com/fsvxavier/nexs-lib/httpserver/providers/nethttp"
)

func init() {
	// Auto-register all providers
	registerProviders()
}

// registerProviders registers all available HTTP server providers.
func registerProviders() {
	providers := map[string]interfaces.ServerFactory{
		"atreugo":  atreugo.Factory,
		"echo":     echo.Factory,
		"fasthttp": fasthttp.Factory,
		"fiber":    fiber.Factory,
		"gin":      gin.Factory,
		"nethttp":  nethttp.Factory,
	}

	for name, factory := range providers {
		if err := defaultRegistry.Register(name, factory); err != nil {
			// Log error but don't panic - some providers might be optional
			continue
		}
	}
}
