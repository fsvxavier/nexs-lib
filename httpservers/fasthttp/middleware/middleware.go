package middleware

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"github.com/fsvxavier/nexs-lib/httpservers/common"
)

// Logger middleware logs HTTP requests
func Logger(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Start timer
		start := time.Now()

		// Get request ID from header
		requestID := string(ctx.Request.Header.Peek("X-Request-ID"))
		if requestID == "" {
			requestID = uuid.New().String()
			ctx.Response.Header.Set("X-Request-ID", requestID)
		}

		// Process request
		next(ctx)

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
	}
}

// Recover is a middleware that recovers from panics
func Recover(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}

				apiErr := common.NewAPIError(500, "INTERNAL_SERVER_ERROR", "Internal Server Error")
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				ctx.SetContentType("application/json")
				ctx.Write(apiErr.Bytes())

				// Log the error
				fmt.Printf("PANIC RECOVERED: %v\n", err)
			}
		}()

		next(ctx)
	}
}

// Tracing adds tracing capability
func Tracing(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Generate trace ID if not present
		traceID := string(ctx.Request.Header.Peek("X-Trace-ID"))
		if traceID == "" {
			traceID = uuid.New().String()
			ctx.Response.Header.Set("X-Trace-ID", traceID)
		}

		// Set span ID
		spanID := uuid.New().String()
		ctx.Response.Header.Set("X-Span-ID", spanID)

		next(ctx)
	}
}
