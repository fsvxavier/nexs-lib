package middleware

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/savsgio/atreugo/v11"

	"github.com/fsvxavier/nexs-lib/httpservers/common"
)

// Logger logs HTTP requests
func Logger(ctx *atreugo.RequestCtx) error {
	// Start timer
	start := time.Now()

	// Store the request ID in context
	requestID := string(ctx.Request.Header.Peek("X-Request-ID"))
	if requestID == "" {
		requestID = uuid.New().String()
		ctx.Response.Header.Set("X-Request-ID", requestID)
	}

	// Call next handler
	err := ctx.Next()

	// Measure execution time
	execTime := time.Since(start).String()

	// Set response header
	ctx.Response.Header.Set("X-Response-Time", execTime)

	// Log format: method path status response-time request-id
	fmt.Printf("[%s] %s %s - %d - %s - %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		string(ctx.Method()),
		string(ctx.Path()),
		ctx.Response.StatusCode(),
		execTime,
		requestID,
	)

	return err
}

// Recover recovers from panics
func Recover(ctx *atreugo.RequestCtx) error {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}

			apiErr := common.NewAPIError(500, "INTERNAL_SERVER_ERROR", "Internal Server Error")
			ctx.SetStatusCode(500)
			ctx.SetContentType("application/json")
			ctx.Write(apiErr.Bytes())

			// Log the error
			fmt.Printf("PANIC RECOVERED: %v\n", err)
		}
	}()

	return ctx.Next()
}

// Tracing adds tracing capability
func Tracing(ctx *atreugo.RequestCtx) error {
	// Generate trace ID if not present
	traceID := string(ctx.Request.Header.Peek("X-Trace-ID"))
	if traceID == "" {
		traceID = uuid.New().String()
		ctx.Response.Header.Set("X-Trace-ID", traceID)
	}

	// Set span ID
	spanID := uuid.New().String()
	ctx.Response.Header.Set("X-Span-ID", spanID)

	return ctx.Next()
}

// CORS handles CORS headers
func CORS(ctx *atreugo.RequestCtx) error {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
	ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")

	// Handle preflight requests
	if string(ctx.Method()) == "OPTIONS" {
		ctx.SetStatusCode(204)
		return nil
	}

	return ctx.Next()
}
