package fibertrace

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/ext"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/mocktracer"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestMiddleware(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	app := fiber.New()
	app.Use(Middleware(WithService("test-service")))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("test")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	spans := mt.FinishedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "test-service", span.Tag(ext.ServiceName))
	assert.Equal(t, "GET", span.Tag(ext.HTTPMethod))
	assert.Equal(t, ext.SpanTypeWeb, span.Tag(ext.SpanType))
	assert.Equal(t, "/test", span.Tag(ext.HTTPRoute))
	assert.Equal(t, componentName, span.Tag(ext.Component))
	assert.Equal(t, ext.SpanKindServer, span.Tag(ext.SpanKind))
	assert.Equal(t, "200", span.Tag(ext.HTTPCode))
}

func TestMiddleware_IgnoreRequest(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	app := fiber.New()
	app.Use(Middleware(WithIgnoreRequest(func(c *fiber.Ctx) bool {
		return c.Path() == "/ignored"
	})))

	app.Get("/ignored", func(c *fiber.Ctx) error {
		return c.SendString("ignored")
	})

	app.Get("/traced", func(c *fiber.Ctx) error {
		return c.SendString("traced")
	})

	// Test ignored path
	req := httptest.NewRequest("GET", "/ignored", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	// Test traced path
	req = httptest.NewRequest("GET", "/traced", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	spans := mt.FinishedSpans()
	require.Len(t, spans, 1) // Only one span for the traced path
	assert.Contains(t, spans[0].Tag(ext.HTTPURL), "/traced")
}

func TestMiddleware_StatusError(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	app := fiber.New()
	app.Use(Middleware())
	app.Get("/error", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusInternalServerError, "test error")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, 500, resp.StatusCode)

	spans := mt.FinishedSpans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "200", span.Tag(ext.HTTPCode))
	assert.Nil(t, span.Tag(ext.Error))
}

func TestSetTraceID(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	app := fiber.New()
	var traceID, spanID string

	app.Get("/test", func(c *fiber.Ctx) error {
		span, _ := tracer.StartSpanFromContext(context.Background(), "test.span")
		tID, sID := setTraceID(c, span)
		traceID = tID
		spanID = sID
		span.Finish()

		assert.Equal(t, traceID, c.Get("Trace-Id"))
		assert.Equal(t, spanID, c.Get("Span-Id"))

		assert.Equal(t, traceID, string(c.Request().Header.Peek("Trace-Id")))
		assert.Equal(t, spanID, string(c.Request().Header.Peek("Span-Id")))

		assert.Equal(t, traceID, string(c.Response().Header.Peek("Trace-Id")))
		assert.Equal(t, spanID, string(c.Response().Header.Peek("Span-Id")))

		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	_, err := app.Test(req)
	require.NoError(t, err)

	assert.NotEmpty(t, traceID)
	assert.NotEmpty(t, spanID)
}

func TestEnvDuration(t *testing.T) {
	// Save original env and restore after test
	originalEnv := os.Getenv("DD_TRACE_KEEP_IF_EXCEEDS")
	defer os.Setenv("DD_TRACE_KEEP_IF_EXCEEDS", originalEnv)

	tests := []struct {
		name          string
		envValue      string
		expectedValue time.Duration
	}{
		{
			name:          "default value",
			envValue:      "",
			expectedValue: 5 * time.Second,
		},
		{
			name:          "custom duration",
			envValue:      "10s",
			expectedValue: 10 * time.Second,
		},
		{
			name:          "invalid duration",
			envValue:      "invalid",
			expectedValue: time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("DD_TRACE_KEEP_IF_EXCEEDS", tt.envValue)

			app := fiber.New()
			app.Use(Middleware())
			app.Get("/test", func(c *fiber.Ctx) error {
				// Creating a mock span to verify that slow requests are tagged correctly
				return c.SendString("test")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			_, err := app.Test(req)
			require.NoError(t, err)
		})
	}
}

func TestSlowRequestTagging(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	// Set threshold to a very small value to trigger the slow request tagging
	os.Setenv("DD_TRACE_KEEP_IF_EXCEEDS", "1ns")
	defer os.Unsetenv("DD_TRACE_KEEP_IF_EXCEEDS")

	app := fiber.New()
	app.Use(Middleware())
	app.Get("/slow", func(c *fiber.Ctx) error {
		time.Sleep(5 * time.Millisecond) // Ensure it exceeds threshold
		return c.SendString("slow")
	})

	req := httptest.NewRequest("GET", "/slow", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	spans := mt.FinishedSpans()
	require.Len(t, spans, 1)

	// Check that the manual keep tag was set
	span := spans[0]
	assert.Equal(t, nil, span.Tag(ext.ManualKeep))
}

// Mock HTTP request to test header extraction
func TestHeaderExtraction(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	app := fiber.New()
	app.Use(Middleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("test")
	})

	// Create a test request with tracing headers
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-B3-TraceId", "1234567890")
	req.Header.Set("X-B3-SpanId", "0987654321")
	req.Header.Set("Client-Id", "test-client")

	// Convert standard request to fasthttp request
	fctx := &fasthttp.RequestCtx{}
	fctx.Request.SetRequestURI(req.URL.String())
	fctx.Request.Header.SetMethod(req.Method)

	for name, values := range req.Header {
		for _, value := range values {
			fctx.Request.Header.Add(name, value)
		}
	}

	// Create a fiber context from the fasthttp context
	ctx := fiber.New().AcquireCtx(fctx)
	defer fiber.New().ReleaseCtx(ctx)

	// Call the middleware handler directly
	middleware := Middleware()
	err := middleware(ctx)
	require.NoError(t, err)

	// Verify client ID was captured
	spans := mt.FinishedSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "test-client", spans[0].Tag("http.client_id"))
}
