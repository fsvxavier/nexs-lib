package resty

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockResponse struct {
	Message string `json:"message"`
}

type mockErrorResponse struct {
	Error string `json:"error"`
}

func setupTestServer(t *testing.T, port string) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ReadBufferSize:        4096,
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
		return c.Status(500).JSON(mockErrorResponse{Error: "internal server error"})
	})

	app.Get("/timeout", func(c *fiber.Ctx) error {
		time.Sleep(2 * time.Second)
		return c.JSON(mockResponse{Message: "timeout"})
	})

	go func() {
		if err := app.Listen(":" + port); err != nil {
			t.Logf("Error starting server: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	return app
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	assert.NotNil(t, client)
}

func TestNewRequester(t *testing.T) {
	client := NewClient()
	req := NewRequester(client)
	assert.NotNil(t, req)
	assert.NotNil(t, req.client)
	assert.NotNil(t, req.restyReq)
}

func TestRequesterMethods(t *testing.T) {
	app := setupTestServer(t, "3000")
	defer app.Shutdown()

	tests := []struct {
		name           string
		method         string
		endpoint       string
		body           []byte
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "GET Request",
			method:         "GET",
			endpoint:       "/test",
			expectedStatus: 200,
			expectedMsg:    "get success",
		},
		{
			name:           "POST Request",
			method:         "POST",
			endpoint:       "/test",
			body:           []byte(`{"test":"data"}`),
			expectedStatus: 200,
			expectedMsg:    "post success",
		},
		{
			name:           "PUT Request",
			method:         "PUT",
			endpoint:       "/test",
			body:           []byte(`{"test":"data"}`),
			expectedStatus: 200,
			expectedMsg:    "put success",
		},
		{
			name:           "DELETE Request",
			method:         "DELETE",
			endpoint:       "/test",
			expectedStatus: 200,
			expectedMsg:    "delete success",
		},
	}

	client := NewClient()
	req := NewRequester(client)
	req.SetBaseURL("http://localhost:3000")

	ctx := context.Background()

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

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var result mockResponse
			err = json.Unmarshal(resp.Body, &result)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, result.Message)
		})
	}
}

func TestRequesterConfiguration(t *testing.T) {
	client := NewClient()
	req := NewRequester(client)

	t.Run("SetHeaders", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer token",
			"Custom-Header": "test-value",
		}
		req.SetHeaders(headers)
		assert.Equal(t, headers, req.headers)
	})

	t.Run("SetBaseURL", func(t *testing.T) {
		baseURL := "http://localhost:3000"
		req.SetBaseURL(baseURL)
		assert.Equal(t, baseURL, req.baseURL)
	})

	t.Run("SetErrorHandler", func(t *testing.T) {
		handler := func(resp *Response) error { return nil }
		req.SetErrorHandler(handler)
		assert.NotNil(t, req.errHandler)
	})
}

func TestErrorHandling(t *testing.T) {
	app := setupTestServer(t, "3001")
	defer app.Shutdown()

	client := NewClient()
	req := NewRequester(client)
	req.SetBaseURL("http://localhost:3001")

	errorHandler := func(resp *Response) error {
		var errResp mockErrorResponse
		err := json.Unmarshal(resp.Body, &errResp)
		if err != nil {
			return err
		}
		return fmt.Errorf(errResp.Error)
	}
	req.SetErrorHandler(errorHandler)

	resp, err := req.Get(context.Background(), "/error")
	assert.Error(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.Contains(t, err.Error(), "internal server error")
}

func TestTimeout(t *testing.T) {
	app := setupTestServer(t, "3002")
	defer app.Shutdown()

	client := resty.New()
	client.SetTimeout(1 * time.Second)

	req := NewRequester(client)
	req.SetBaseURL("http://localhost:3002")

	_, err := req.Get(context.Background(), "/timeout")
	assert.Error(t, err)
}

func TestContextCancellation(t *testing.T) {
	app := setupTestServer(t, "3003")
	defer app.Shutdown()

	client := NewClient()
	req := NewRequester(client)
	req.SetBaseURL("http://localhost:3003")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := req.Get(ctx, "/test")
	assert.Error(t, err)
}

func TestUnmarshal(t *testing.T) {
	app := setupTestServer(t, "3004")
	defer app.Shutdown()

	client := NewClient()
	req := NewRequester(client)
	req.SetBaseURL("http://localhost:3004")

	var result mockResponse
	req.Unmarshal(&result)

	resp, err := req.Get(context.Background(), "/test")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "get success", result.Message)
}
