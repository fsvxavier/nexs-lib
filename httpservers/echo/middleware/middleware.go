package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/fsvxavier/nexs-lib/httpservers/common"
)

// ErrorHandler is a custom error handler for Echo
func ErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError

	// Check if it's an Echo HTTPError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	// Check if it's our custom APIError
	if apiErr, ok := err.(*common.APIError); ok {
		code = apiErr.StatusCode
		c.JSON(code, apiErr)
		return
	}

	// Default error
	c.JSON(code, map[string]interface{}{
		"error": err.Error(),
	})
}

// Logger is a middleware that logs HTTP requests
func Logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Start timer
		start := time.Now()

		// Store the request ID in context
		requestID := c.Request().Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Response().Header().Set("X-Request-ID", requestID)
		}

		// Process request
		err := next(c)

		// Measure execution time
		execTime := time.Since(start).String()

		// Set response header
		c.Response().Header().Set("X-Response-Time", execTime)

		// Log format: method path status response-time request-id
		fmt.Printf("[%s] %s %s - %d - %s - %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			c.Request().Method,
			c.Request().URL.Path,
			c.Response().Status,
			execTime,
			requestID,
		)

		return err
	}
}

// Tracing adds tracing capability
func Tracing(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Generate trace ID if not present
		traceID := c.Request().Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
			c.Response().Header().Set("X-Trace-ID", traceID)
		}

		// Set span ID
		spanID := uuid.New().String()
		c.Response().Header().Set("X-Span-ID", spanID)

		return next(c)
	}
}

// CORS middleware for handling CORS requests
func CORS(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Response().Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request().Method == "OPTIONS" {
			return c.NoContent(http.StatusNoContent)
		}

		return next(c)
	}
}

// RegisterPprof registers pprof routes on the Echo instance
func RegisterPprof(e *echo.Echo) {
	// Implementation depends on project dependencies
	// Example:
	// pprof.Register(e)
	e.GET("/debug/pprof", func(c echo.Context) error {
		return c.String(http.StatusOK, "Pprof enabled")
	})
}

// RegisterMetrics registers metric routes on the Echo instance
func RegisterMetrics(e *echo.Echo) {
	// Implementation depends on project dependencies
	// Example:
	// e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	e.GET("/metrics", func(c echo.Context) error {
		return c.String(http.StatusOK, "Metrics endpoint enabled")
	})
}

// RegisterSwagger registers swagger routes on the Echo instance
func RegisterSwagger(e *echo.Echo) {
	// Implementation depends on project dependencies
	// Example:
	// e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/swagger/*", func(c echo.Context) error {
		return c.String(http.StatusOK, "Swagger enabled")
	})
}
