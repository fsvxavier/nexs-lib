package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestRequestLoggerMiddleware(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(RequestLoggerMiddleware)

	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	tests := []struct {
		body           interface{}
		headers        map[string]string
		name           string
		expectedStatus int
	}{
		{
			name:           "Valid request with body",
			body:           map[string]string{"key": "value"},
			headers:        map[string]string{"Uuid": "12345", "Client-Id": "client123"},
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "Invalid JSON body",
			body:           "invalid json",
			headers:        map[string]string{"Uuid": "12345", "Client-Id": "client123"},
			expectedStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			var request *http.Request
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					bodyBytes = []byte(str)
				} else {
					bodyBytes, _ = json.Marshal(tt.body)
				}
				reqBody := bytes.NewReader(bodyBytes)
				request = httptest.NewRequest(fiber.MethodPost, "/test", reqBody)
			} else {
				request = httptest.NewRequest(fiber.MethodPost, "/test", nil)
			}

			for key, value := range tt.headers {
				request.Header.Set(key, value)
			}

			resp, err := app.Test(request)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}
func TestRemoveSensitiveData(t *testing.T) {
	// Set sensitiveData for test
	originalSensitiveData := sensitiveData
	sensitiveData = []string{"password", "token"}
	defer func() { sensitiveData = originalSensitiveData }()

	tests := []struct {
		name           string
		input          map[string]interface{}
		expected       map[string]interface{}
		additionalData map[string]interface{}
	}{
		{
			name: "Removes sensitive keys from root and additional_data",
			input: map[string]interface{}{
				"username": "user1",
				"password": "secret",
				"token":    "abc123",
				ADDITIONAL_DATA_KEY: map[string]interface{}{
					"token":    "shouldRemove",
					"password": "shouldRemove",
					"foo":      "bar",
				},
			},
			expected: map[string]interface{}{
				"username": "user1",
				ADDITIONAL_DATA_KEY: map[string]interface{}{
					"foo": "bar",
				},
			},
		},
		{
			name: "No sensitive keys present",
			input: map[string]interface{}{
				"username": "user2",
				ADDITIONAL_DATA_KEY: map[string]interface{}{
					"foo": "bar",
				},
			},
			expected: map[string]interface{}{
				"username": "user2",
				ADDITIONAL_DATA_KEY: map[string]interface{}{
					"foo": "bar",
				},
			},
		},
		{
			name: "additional_data not a map",
			input: map[string]interface{}{
				"password":          "secret",
				ADDITIONAL_DATA_KEY: "not_a_map",
			},
			expected: map[string]interface{}{
				ADDITIONAL_DATA_KEY: "not_a_map",
			},
		},
		{
			name: "additional_data missing",
			input: map[string]interface{}{
				"token": "abc123",
			},
			expected: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Deep copy input to avoid mutation issues
			inputCopy, _ := json.Marshal(tt.input)
			var data map[string]interface{}
			_ = json.Unmarshal(inputCopy, &data)

			removeSensitiveData(data)

			assert.Equal(t, tt.expected, data)
		})
	}
}
func TestRequestLoggerMiddleware_ValidJSONBody(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Use(RequestLoggerMiddleware)
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	body := map[string]interface{}{"foo": "bar"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Uuid", "uuid-123")
	req.Header.Set("Client-Id", "client-xyz")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRequestLoggerMiddleware_InvalidJSONBody(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Use(RequestLoggerMiddleware)
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Uuid", "uuid-123")
	req.Header.Set("Client-Id", "client-xyz")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestRequestLoggerMiddleware_NoBody(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Use(RequestLoggerMiddleware)
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Uuid", "uuid-123")
	req.Header.Set("Client-Id", "client-xyz")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRequestLoggerMiddleware_WithParams(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Use(RequestLoggerMiddleware)
	app.Post("/test/:id", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test/42", nil)
	req.Header.Set("Uuid", "uuid-123")
	req.Header.Set("Client-Id", "client-xyz")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
