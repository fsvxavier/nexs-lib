package middleware

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentTypeMiddleware(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(ContentTypeMiddleware("POST", "application/json", "application/xml"))

	app.Post("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	tests := []struct {
		name           string
		method         string
		contentType    string
		expectedBody   string
		expectedStatus int
	}{
		{
			name:           "Valid Content-Type",
			method:         "POST",
			contentType:    "application/json",
			expectedStatus: fiber.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "Different Method",
			method:         "GET",
			contentType:    "application/json",
			expectedStatus: fiber.StatusMethodNotAllowed,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)
			req.Header.Set("Content-Type", tt.contentType)
			req.Header.Set("Trace-Id", "test-trace-id")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedBody != "" {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tt.expectedBody, string(body))
			}
		})
	}
}
func TestContentTypeMiddleware_UnsupportedContentType(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(ContentTypeMiddleware("POST", "application/json", "application/xml"))

	app.Post("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Trace-Id", "trace-123")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnsupportedMediaType, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"code":"415"`)
	assert.Contains(t, string(body), `"Unsupported Media Type"`)
}

func TestContentTypeMiddleware_MissingContentType(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(ContentTypeMiddleware("POST", "application/json"))

	app.Post("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Trace-Id", "trace-456")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnsupportedMediaType, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"code":"415"`)
	assert.Contains(t, string(body), `"Unsupported Media Type"`)
}

func TestContentTypeMiddleware_AllowedXmlContentType(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(ContentTypeMiddleware("POST", "application/json", "application/xml"))

	app.Post("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Trace-Id", "trace-xml")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
