package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/fsvxavier/nexs-lib/httpservers/common"
)

// ErrorHandler is the default error handler for Fiber
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Default status code is 500
	code := fiber.StatusInternalServerError

	// Check if it's a Fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Check if it's our custom APIError
	if apiErr, ok := err.(*common.APIError); ok {
		code = apiErr.StatusCode
		return c.Status(code).JSON(apiErr)
	}

	// Return default error
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}

// Logger middleware logs HTTP requests
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Start timer
		start := time.Now()

		// Store the request ID in context
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Set("X-Request-ID", requestID)
		}

		// Process request
		err := c.Next()

		// Measure execution time
		execTime := time.Since(start).String()

		// Log request details
		c.Set("X-Response-Time", execTime)

		// Log format: method path status response-time request-id
		fmt.Printf("[%s] %s %s - %d - %s - %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			execTime,
			requestID,
		)

		return err
	}
}

// Tracing adds tracing capability
func Tracing() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Generate trace ID if not present
		traceID := c.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
			c.Set("X-Trace-ID", traceID)
		}

		// Set span ID
		spanID := uuid.New().String()
		c.Set("X-Span-ID", spanID)

		return c.Next()
	}
}
