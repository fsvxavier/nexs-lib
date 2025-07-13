package fiber

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type mockResponse struct {
	Message string `json:"message"`
}

func setupTestServer() *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(mockResponse{Message: "get success"})
	})

	app.Post("/test", func(c *fiber.Ctx) error {
		return c.JSON(mockResponse{Message: "post success"})
	})

	app.Put("/test", func(c *fiber.Ctx) error {
		return c.JSON(mockResponse{Message: "put success"})
	})

	app.Delete("/test", func(c *fiber.Ctx) error {
		return c.JSON(mockResponse{Message: "delete success"})
	})

	app.Get("/error", func(c *fiber.Ctx) error {
		return c.Status(500).JSON(fiber.Map{
			"error": "internal server error",
		})
	})

	return app
}

func TestNew(t *testing.T) {
	baseURL := "http://localhost:3000"
	req := New(baseURL)

	assert.NotNil(t, req)
	assert.Equal(t, baseURL, req.baseURL)
	assert.NotNil(t, req.client)
}

func TestSetHeaders(t *testing.T) {
	req := New("http://localhost:3000")
	headers := map[string]string{
		"Authorization": "Bearer token",
		"Content-Type":  "application/json",
	}

	req.SetHeaders(headers)
	assert.Equal(t, headers, req.headers)
}

func TestSetBaseURL(t *testing.T) {
	req := New("http://localhost:3000")
	newBaseURL := "http://localhost:3001"

	req.SetBaseURL(newBaseURL)
	assert.Equal(t, newBaseURL, req.baseURL)
}

func TestHTTPMethods(t *testing.T) {
	app := setupTestServer()
	go app.Listen(":3000")
	defer app.Shutdown()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name     string
		method   string
		endpoint string
		body     []byte
		expected string
	}{
		{
			name:     "GET request",
			method:   "GET",
			endpoint: "/test",
			body:     nil,
			expected: "get success",
		},
		{
			name:     "POST request",
			method:   "POST",
			endpoint: "/test",
			body:     []byte(`{"test":"data"}`),
			expected: "post success",
		},
		{
			name:     "PUT request",
			method:   "PUT",
			endpoint: "/test",
			body:     []byte(`{"test":"data"}`),
			expected: "put success",
		},
		{
			name:     "DELETE request",
			method:   "DELETE",
			endpoint: "/test",
			body:     nil,
			expected: "delete success",
		},
	}

	ctx := context.Background()
	req := New("http://localhost:3000")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *Response
			var err error

			switch tt.method {
			case "GET":
				resp, err = req.Get(ctx, tt.endpoint)
			case "POST":
				resp, err = req.Post(ctx, tt.endpoint, tt.body)
			case "PUT":
				resp, err = req.Put(ctx, tt.endpoint, tt.body)
			case "DELETE":
				resp, err = req.Delete(ctx, tt.endpoint)
			}

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, 200, resp.StatusCode)

			var result mockResponse
			err = json.Unmarshal(resp.Body, &result)
			assert.NoError(t, err)
			assert.Contains(t, result.Message, tt.expected)
		})
	}
}

func TestExecute_WithError(t *testing.T) {
	app := setupTestServer()
	go app.Listen(":3001")
	defer app.Shutdown()

	time.Sleep(100 * time.Millisecond)

	req := New("http://localhost:3001")
	resp, err := req.Get(context.Background(), "/error")

	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	var result map[string]string
	err = json.Unmarshal(resp.Body, &result)
	assert.NoError(t, err)
	assert.Equal(t, "internal server error", result["error"])
}

func TestUnmarshal(t *testing.T) {
	app := setupTestServer()
	go app.Listen(":3002")
	defer app.Shutdown()

	time.Sleep(100 * time.Millisecond)

	req := New("http://localhost:3002")
	var result mockResponse
	req.Unmarshal(&result)

	resp, err := req.Get(context.Background(), "/test")
	assert.NoError(t, err)
	assert.Equal(t, "get success", result.Message)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestApiErrorHandler(t *testing.T) {
	tests := []struct {
		name        string
		response    *Response
		expectedErr bool
	}{
		{
			name: "Success response",
			response: &Response{
				StatusCode: 200,
				IsError:    false,
				Body:       []byte(`{"message":"success"}`),
			},
			expectedErr: false,
		},
		{
			name: "Error response",
			response: &Response{
				StatusCode: 500,
				IsError:    true,
				Body:       []byte(`{"error":"internal server error"}`),
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ApiErrorHandler(tt.response)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient()

	assert.NotNil(t, client)
	assert.NotNil(t, client.JSONEncoder)
	assert.NotNil(t, client.JSONDecoder)
}
func TestSetErrorHandler(t *testing.T) {
	req := New("http://localhost:3000")
	handler := func(res *Response) error {
		return nil
	}

	req.SetErrorHandler(handler)
	assert.NotNil(t, req.errHandler)
	assert.Equal(t, fmt.Sprintf("%p", handler), fmt.Sprintf("%p", req.errHandler))
}
func TestExecute(t *testing.T) {
	app := setupTestServer()
	go app.Listen(":3003")
	defer app.Shutdown()

	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name           string
		method         string
		endpoint       string
		body           []byte
		headers        map[string]string
		expectedStatus int
		expectedBody   string
		expectError    bool
	}{
		{
			name:           "GET request with headers",
			method:         "GET",
			endpoint:       "/test",
			headers:        map[string]string{"X-Test": "test-header"},
			expectedStatus: 200,
			expectedBody:   "get success",
			expectError:    false,
		},
		{
			name:           "POST request with body",
			method:         "POST",
			endpoint:       "/test",
			body:           []byte(`{"test":"data"}`),
			expectedStatus: 200,
			expectedBody:   "post success",
			expectError:    false,
		},
		{
			name:           "Error endpoint",
			method:         "GET",
			endpoint:       "/error",
			expectedStatus: 500,
			expectedBody:   "internal server error",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := New("http://localhost:3003")
			if tt.headers != nil {
				req.SetHeaders(tt.headers)
			}

			resp, err := req.Execute(context.Background(), tt.method, tt.endpoint, tt.body)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == 200 {
				var result mockResponse
				err = json.Unmarshal(resp.Body, &result)
				assert.NoError(t, err)
				assert.Contains(t, result.Message, tt.expectedBody)
			} else {
				var result map[string]string
				err = json.Unmarshal(resp.Body, &result)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, result["error"])
			}
		})
	}
}
