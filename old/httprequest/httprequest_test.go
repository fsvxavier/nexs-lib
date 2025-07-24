package httprequest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockResponse struct {
	Message string `json:"message"`
}

type mockErrorResponse struct {
	Error string `json:"error"`
}

func setupServer(t *testing.T, port string) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ReadBufferSize:        4096,
	})

	// Setup test routes
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(mockResponse{Message: "get success"})
	})

	app.Post("/test", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return err
		}
		return c.JSON(mockResponse{Message: "post success"})
	})

	app.Put("/test", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return err
		}
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

func TestHTTPMethods(t *testing.T) {
	app := setupServer(t, "3000")
	defer app.Shutdown()

	tests := []struct {
		name           string
		method         string
		endpoint       string
		body           interface{}
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
			body:           map[string]string{"test": "data"},
			expectedStatus: 200,
			expectedMsg:    "post success",
		},
		{
			name:           "PUT Request",
			method:         "PUT",
			endpoint:       "/test",
			body:           map[string]string{"test": "data"},
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

	baseURL := "http://localhost:3000"
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := New(baseURL)
			req.SetContext(ctx)

			var resp *Response
			var err error

			switch tt.method {
			case "GET":
				resp, err = req.Get(tt.endpoint)
			case "POST":
				resp, err = req.Post(tt.endpoint, tt.body)
			case "PUT":
				resp, err = req.Put(tt.endpoint, tt.body)
			case "DELETE":
				resp, err = req.Delete(tt.endpoint)
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var result mockResponse
			err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(resp.Body, &result)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, result.Message)
		})
	}
}

func TestErrorHandling(t *testing.T) {
	app := setupServer(t, "3001")
	defer app.Shutdown()

	req := New("http://localhost:3001")
	req.SetContext(context.Background())

	customErrorHandler := func(resp *Response) error {
		var errResp mockErrorResponse
		if err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(resp.Body, &errResp); err != nil {
			return err
		}
		return fmt.Errorf(errResp.Error)
	}

	req.SetErrorHandler(customErrorHandler)
	resp, err := req.Get("/error")

	assert.Error(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.Contains(t, err.Error(), "internal server error")
}

func TestHeaders(t *testing.T) {
	app := setupServer(t, "3002")
	defer app.Shutdown()

	req := New("http://localhost:3002")
	req.SetContext(context.Background())

	headers := map[string]string{
		"Authorization": "Bearer token123",
		"Custom-Header": "test-value",
	}
	req.SetHeaders(headers)

	resp, err := req.Get("/test")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestUnmarshal(t *testing.T) {
	app := setupServer(t, "3003")
	defer app.Shutdown()

	req := New("http://localhost:3003")
	req.SetContext(context.Background())

	var result mockResponse
	req.Unmarshal(&result)

	resp, err := req.Get("/test")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "get success", result.Message)
}

func TestRawExecute(t *testing.T) {
	app := setupServer(t, "3004")
	defer app.Shutdown()

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	query := map[string]string{
		"param": "value",
	}

	statusCode, body, respHeaders, err := RawExecute(
		"GET",
		"http://localhost:3004/test",
		"",
		headers,
		query,
	)

	require.NoError(t, err)
	assert.Equal(t, 200, statusCode)
	assert.Contains(t, body, "get success")
	assert.NotEmpty(t, respHeaders)
}
