package fiber

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/skip"
	"github.com/gofiber/swagger"

	"github.com/dock-tech/isis-golang-lib/httpserver/fiber/middleware"
	log "github.com/dock-tech/isis-golang-lib/observability/logger/zap"
	"github.com/dock-tech/isis-golang-lib/observability/tracer/datadog/fibertrace"
)

const (
	defaultReadBufferSize = 4096
)

type FiberEngine struct {
	app  *fiber.App
	port string
}

var healthcheckPath = func(c *fiber.Ctx) bool { return c.Path() == "/health" || c.Path() == "/readyz" }

func (engine *FiberEngine) NewWebserver() {
	api := fiber.New(fiber.Config{
		ErrorHandler:          middleware.ApplicationErrorHandler,
		ReadBufferSize:        getReadBufferSizeConfiguration(os.Getenv("HTTP_READ_BUFFER_SIZE")),
		DisableStartupMessage: os.Getenv("HTTP_DISABLE_START_MSG") == "true",
		Prefork:               os.Getenv("HTTP_PREFORK") == "true",
	})

	if os.Getenv("PPROF_ENABLED") == "true" {
		api.Use(pprof.New())
		api.Get("/metrics", monitor.New())
	}

	api.Use(skip.New(fibertrace.Middleware(fibertrace.WithEnvironment(os.Getenv("DD_ENV"))), healthcheckPath))

	api.Use(recover.New(recover.Config{
		EnableStackTrace: os.Getenv("SHOW_STACK_TRACE") == "true",
	}))

	api.Use(skip.New(middleware.LoggerMiddleware(os.Stdout), healthcheckPath))
	api.Use(skip.New(middleware.TraceIdMiddleware, healthcheckPath))
	api.Use(skip.New(middleware.TenantIdMiddleware, healthcheckPath))
	api.Use(middleware.ContentTypeMiddleware("POST", fiber.MIMEApplicationJSON))

	engine.app = api
	engine.port = os.Getenv("HTTP_PORT")
}

func (engine *FiberEngine) GetApp() *fiber.App {
	return engine.app
}

func (engine *FiberEngine) GetPort() string {
	return engine.port
}

func (engine *FiberEngine) Run() error {
	log.Debugln(fmt.Sprintf("Listening on port %s", engine.port))
	return engine.app.Listen(("0.0.0.0:" + engine.port))
}

func (engine *FiberEngine) Router(app *fiber.App) {
	app.Route("/docs/*", func(r fiber.Router) {
		r.Get("", swagger.New(swagger.Config{
			DocExpansion: "none",
		}))
	})

	app.All("/*", func(ctx *fiber.Ctx) error {
		ctx.Status(http.StatusForbidden)
		return ctx.JSON(fiber.Map{"message": "Forbidden"})
	})

	engine.app = app
}

func (engine *FiberEngine) Shutdown(ctx context.Context) error {
	return engine.app.ShutdownWithContext(ctx)
}

func getReadBufferSizeConfiguration(bufferSizeConfigValue string) int {
	if bufferSizeConfigValue == "" {
		return defaultReadBufferSize
	}

	bufferSize, err := strconv.Atoi(bufferSizeConfigValue)
	if err != nil {
		return defaultReadBufferSize
	}

	if bufferSize < defaultReadBufferSize {
		return defaultReadBufferSize
	}

	return bufferSize
}
