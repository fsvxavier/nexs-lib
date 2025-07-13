package fiber

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Set default environment variables for testing
	os.Setenv("HTTP_PORT", "8080")
	os.Setenv("HTTP_READ_BUFFER_SIZE", "4096")
	os.Setenv("HTTP_DISABLE_START_MSG", "true")
	os.Setenv("HTTP_PREFORK", "false")
	os.Setenv("PPROF_ENABLED", "false")
	os.Setenv("SHOW_STACK_TRACE", "false")
	os.Setenv("DD_ENV", "test")
}

func TestFiberEngine_NewWebserver(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected int
	}{
		{
			name: "Default configuration",
			envVars: map[string]string{
				"HTTP_PORT": "8080",
			},
			expected: 4096,
		},
		{
			name: "Custom buffer size",
			envVars: map[string]string{
				"HTTP_PORT":             "8080",
				"HTTP_READ_BUFFER_SIZE": "8192",
			},
			expected: 8192,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			engine := &FiberEngine{}
			engine.NewWebserver()

			assert.NotNil(t, engine.app)
			assert.Equal(t, tt.envVars["HTTP_PORT"], engine.port)
		})
	}
}

func TestFiberEngine_Router(t *testing.T) {
	engine := &FiberEngine{}
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	engine.Router(app)

	assert.NotNil(t, engine.app)

	req, _ := http.NewRequest("GET", "/unknown", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestFiberEngine_GetApp(t *testing.T) {
	engine := &FiberEngine{
		app: fiber.New(fiber.Config{
			DisableStartupMessage: true,
		}),
	}
	assert.NotNil(t, engine.GetApp())
}

func TestFiberEngine_GetPort(t *testing.T) {
	engine := &FiberEngine{
		port: "8080",
	}
	assert.Equal(t, "8080", engine.GetPort())
}

func TestFiberEngine_Shutdown(t *testing.T) {
	engine := &FiberEngine{
		app: fiber.New(fiber.Config{
			DisableStartupMessage: true,
		}),
	}
	err := engine.Shutdown(context.Background())
	assert.NoError(t, err)
}

func Test_getReadBufferSizeConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Empty input",
			input:    "",
			expected: defaultReadBufferSize,
		},
		{
			name:     "Invalid input",
			input:    "invalid",
			expected: defaultReadBufferSize,
		},
		{
			name:     "Small buffer size",
			input:    "1024",
			expected: defaultReadBufferSize,
		},
		{
			name:     "Valid buffer size",
			input:    "8192",
			expected: 8192,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getReadBufferSizeConfiguration(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
func TestFiberEngine_NewWebserver_Configuration(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
	}{
		{
			name: "Default configuration",
			envVars: map[string]string{
				"HTTP_PORT": "8080",
			},
		},
		{
			name: "Full configuration",
			envVars: map[string]string{
				"HTTP_PORT":              "8080",
				"HTTP_READ_BUFFER_SIZE":  "8192",
				"HTTP_DISABLE_START_MSG": "true",
				"HTTP_PREFORK":           "true",
				"PPROF_ENABLED":          "true",
				"SHOW_STACK_TRACE":       "true",
				"DD_ENV":                 "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			engine := &FiberEngine{}
			engine.NewWebserver()

			assert.NotNil(t, engine.app)
			assert.Equal(t, tt.envVars["HTTP_PORT"], engine.port)

			// Test PPROF routes when enabled
			if tt.envVars["PPROF_ENABLED"] == "true" {
				req, _ := http.NewRequest("GET", "/metrics", nil)
				resp, err := engine.app.Test(req)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}
		})
	}
}
func TestFiberEngine_Run(t *testing.T) {
	engine := &FiberEngine{
		app: fiber.New(fiber.Config{
			DisableStartupMessage: true,
		}),
		port: "8080",
	}

	// Start server in goroutine since Listen blocks
	go func() {
		err := engine.Run()
		assert.NoError(t, err)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Make test request
	resp, err := http.Get("http://0.0.0.0:8080")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// Cleanup
	err = engine.Shutdown(context.Background())
	assert.NoError(t, err)
}
