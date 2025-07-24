package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestTraceIdMiddleware(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(TraceIdMiddleware)

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("OK")
	})

	tests := []struct {
		name           string
		traceIdHeader  string
		expectedStatus int
	}{
		{
			name:           "No Trace-Id header",
			traceIdHeader:  "",
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "With Trace-Id header",
			traceIdHeader:  "existing-trace-id",
			expectedStatus: fiber.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.traceIdHeader != "" {
				req.Header.Set("Trace-Id", tt.traceIdHeader)
			}

			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			traceId := resp.Header.Get("Trace-Id")
			if tt.traceIdHeader == "" {
				assert.NotEmpty(t, traceId)
			} else {
				assert.Equal(t, tt.traceIdHeader, traceId)
			}
		})
	}
}
