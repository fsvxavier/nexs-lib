package httpservers

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/httpservers/atreugo"
	"github.com/fsvxavier/nexs-lib/httpservers/common"
	"github.com/fsvxavier/nexs-lib/httpservers/echo"
	"github.com/fsvxavier/nexs-lib/httpservers/fasthttp"
	"github.com/fsvxavier/nexs-lib/httpservers/fiber"
	"github.com/fsvxavier/nexs-lib/httpservers/gin"
	"github.com/fsvxavier/nexs-lib/httpservers/nethttp"
)

// ServerType represents the type of HTTP server to use
type ServerType string

const (
	// ServerTypeFiber uses the Fiber framework
	ServerTypeFiber ServerType = "fiber"

	// ServerTypeFastHTTP uses the FastHTTP framework
	ServerTypeFastHTTP ServerType = "fasthttp"

	// ServerTypeNetHTTP uses the standard net/http package
	ServerTypeNetHTTP ServerType = "nethttp"

	// ServerTypeGin uses the Gin framework
	ServerTypeGin ServerType = "gin"

	// ServerTypeEcho uses the Echo framework
	ServerTypeEcho ServerType = "echo"

	// ServerTypeAtreugo uses the Atreugo framework
	ServerTypeAtreugo ServerType = "atreugo"
)

// NewServer creates a new HTTP server of the specified type
func NewServer(serverType ServerType, options ...common.ServerOption) (common.Server, error) {
	switch serverType {
	case ServerTypeFiber:
		return fiber.NewServer(options...), nil
	case ServerTypeFastHTTP:
		return fasthttp.NewServer(options...), nil
	case ServerTypeNetHTTP:
		return nethttp.NewServer(options...), nil
	case ServerTypeGin:
		return gin.NewServer(options...), nil
	case ServerTypeEcho:
		return echo.NewServer(options...), nil
	case ServerTypeAtreugo:
		return atreugo.NewServer(options...), nil
	default:
		return nil, fmt.Errorf("unknown server type: %s", serverType)
	}
}

// Example usage:
//
// import (
//     "fmt"
//     "os"
//     "time"
//
//     "github.com/fsvxavier/nexs-lib/httpservers"
//     "github.com/fsvxavier/nexs-lib/httpservers/common"
// )
//
// func main() {
//     // Create a Fiber server with custom options
//     server, err := httpservers.NewServer(
//         httpservers.ServerTypeFiber,
//         common.WithPort("8080"),
//         common.WithReadTimeout(10*time.Second),
//         common.WithPprof(true),
//         common.WithSwagger(true),
//     )
//     if err != nil {
//         panic(err)
//     }
//
//     // Start the server (this will block until interrupted)
//     if err := server.Start(); err != nil {
//         fmt.Printf("Error starting server: %v\n", err)
//         os.Exit(1)
//     }
// }
