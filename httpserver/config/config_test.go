package config

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

func TestNewBaseConfig(t *testing.T) {
	config := NewBaseConfig()

	if config.GetAddr() != "0.0.0.0" {
		t.Errorf("Expected addr '0.0.0.0', got '%s'", config.GetAddr())
	}

	if config.GetPort() != 8080 {
		t.Errorf("Expected port 8080, got %d", config.GetPort())
	}

	if config.GetReadTimeout() != 30*time.Second {
		t.Errorf("Expected read timeout 30s, got %v", config.GetReadTimeout())
	}

	if config.GetWriteTimeout() != 30*time.Second {
		t.Errorf("Expected write timeout 30s, got %v", config.GetWriteTimeout())
	}

	if config.GetIdleTimeout() != 60*time.Second {
		t.Errorf("Expected idle timeout 60s, got %v", config.GetIdleTimeout())
	}

	if !config.IsGracefulShutdown() {
		t.Error("Expected graceful shutdown to be enabled")
	}

	if config.GetShutdownTimeout() != 30*time.Second {
		t.Errorf("Expected shutdown timeout 30s, got %v", config.GetShutdownTimeout())
	}

	if config.GetFullAddr() != "0.0.0.0:8080" {
		t.Errorf("Expected full addr '0.0.0.0:8080', got '%s'", config.GetFullAddr())
	}
}

func TestBaseConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *BaseConfig
		wantError bool
	}{
		{
			name:      "valid config",
			config:    NewBaseConfig(),
			wantError: false,
		},
		{
			name: "invalid port - zero",
			config: &BaseConfig{
				port:            0,
				readTimeout:     30 * time.Second,
				writeTimeout:    30 * time.Second,
				idleTimeout:     60 * time.Second,
				shutdownTimeout: 30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "invalid port - negative",
			config: &BaseConfig{
				port:            -1,
				readTimeout:     30 * time.Second,
				writeTimeout:    30 * time.Second,
				idleTimeout:     60 * time.Second,
				shutdownTimeout: 30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "invalid port - too high",
			config: &BaseConfig{
				port:            65536,
				readTimeout:     30 * time.Second,
				writeTimeout:    30 * time.Second,
				idleTimeout:     60 * time.Second,
				shutdownTimeout: 30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "invalid read timeout",
			config: &BaseConfig{
				port:            8080,
				readTimeout:     0,
				writeTimeout:    30 * time.Second,
				idleTimeout:     60 * time.Second,
				shutdownTimeout: 30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "invalid write timeout",
			config: &BaseConfig{
				port:            8080,
				readTimeout:     30 * time.Second,
				writeTimeout:    0,
				idleTimeout:     60 * time.Second,
				shutdownTimeout: 30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "invalid idle timeout",
			config: &BaseConfig{
				port:            8080,
				readTimeout:     30 * time.Second,
				writeTimeout:    30 * time.Second,
				idleTimeout:     0,
				shutdownTimeout: 30 * time.Second,
			},
			wantError: true,
		},
		{
			name: "invalid shutdown timeout",
			config: &BaseConfig{
				port:            8080,
				readTimeout:     30 * time.Second,
				writeTimeout:    30 * time.Second,
				idleTimeout:     60 * time.Second,
				shutdownTimeout: 0,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("BaseConfig.Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestBaseConfigClone(t *testing.T) {
	original := NewBaseConfig()
	original.SetCustom("test", "value")

	clone := original.Clone()

	// Test that basic values are copied
	if clone.GetAddr() != original.GetAddr() {
		t.Error("Clone addr does not match original")
	}

	if clone.GetPort() != original.GetPort() {
		t.Error("Clone port does not match original")
	}

	// Test that custom values are copied
	value, exists := clone.GetCustom("test")
	if !exists || value != "value" {
		t.Error("Clone custom value does not match original")
	}

	// Test that modifications to clone don't affect original
	clone.SetCustom("test", "new_value")
	originalValue, _ := original.GetCustom("test")
	if originalValue != "value" {
		t.Error("Original custom value was modified by clone")
	}
}

func TestWithOptions(t *testing.T) {
	tests := []struct {
		name      string
		option    Option
		wantError bool
		validator func(*BaseConfig) bool
	}{
		{
			name:      "WithAddr valid",
			option:    WithAddr("127.0.0.1"),
			wantError: false,
			validator: func(c *BaseConfig) bool { return c.GetAddr() == "127.0.0.1" },
		},
		{
			name:      "WithAddr empty",
			option:    WithAddr(""),
			wantError: true,
			validator: nil,
		},
		{
			name:      "WithPort valid",
			option:    WithPort(9000),
			wantError: false,
			validator: func(c *BaseConfig) bool { return c.GetPort() == 9000 },
		},
		{
			name:      "WithPort invalid - zero",
			option:    WithPort(0),
			wantError: true,
			validator: nil,
		},
		{
			name:      "WithPort invalid - negative",
			option:    WithPort(-1),
			wantError: true,
			validator: nil,
		},
		{
			name:      "WithPort invalid - too high",
			option:    WithPort(65536),
			wantError: true,
			validator: nil,
		},
		{
			name:      "WithReadTimeout valid",
			option:    WithReadTimeout(60 * time.Second),
			wantError: false,
			validator: func(c *BaseConfig) bool { return c.GetReadTimeout() == 60*time.Second },
		},
		{
			name:      "WithReadTimeout invalid",
			option:    WithReadTimeout(0),
			wantError: true,
			validator: nil,
		},
		{
			name:      "WithWriteTimeout valid",
			option:    WithWriteTimeout(60 * time.Second),
			wantError: false,
			validator: func(c *BaseConfig) bool { return c.GetWriteTimeout() == 60*time.Second },
		},
		{
			name:      "WithWriteTimeout invalid",
			option:    WithWriteTimeout(0),
			wantError: true,
			validator: nil,
		},
		{
			name:      "WithIdleTimeout valid",
			option:    WithIdleTimeout(120 * time.Second),
			wantError: false,
			validator: func(c *BaseConfig) bool { return c.GetIdleTimeout() == 120*time.Second },
		},
		{
			name:      "WithIdleTimeout invalid",
			option:    WithIdleTimeout(0),
			wantError: true,
			validator: nil,
		},
		{
			name:      "WithGracefulShutdown true",
			option:    WithGracefulShutdown(true),
			wantError: false,
			validator: func(c *BaseConfig) bool { return c.IsGracefulShutdown() },
		},
		{
			name:      "WithGracefulShutdown false",
			option:    WithGracefulShutdown(false),
			wantError: false,
			validator: func(c *BaseConfig) bool { return !c.IsGracefulShutdown() },
		},
		{
			name:      "WithShutdownTimeout valid",
			option:    WithShutdownTimeout(60 * time.Second),
			wantError: false,
			validator: func(c *BaseConfig) bool { return c.GetShutdownTimeout() == 60*time.Second },
		},
		{
			name:      "WithShutdownTimeout invalid",
			option:    WithShutdownTimeout(0),
			wantError: true,
			validator: nil,
		},
		{
			name:      "WithProvider valid",
			option:    WithProvider("gin"),
			wantError: false,
			validator: func(c *BaseConfig) bool { return c.GetProvider() == "gin" },
		},
		{
			name:      "WithProvider empty",
			option:    WithProvider(""),
			wantError: true,
			validator: nil,
		},
		{
			name:      "WithCustom valid",
			option:    WithCustom("key", "value"),
			wantError: false,
			validator: func(c *BaseConfig) bool {
				value, exists := c.GetCustom("key")
				return exists && value == "value"
			},
		},
		{
			name:      "WithCustom empty key",
			option:    WithCustom("", "value"),
			wantError: true,
			validator: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewBaseConfig()
			err := tt.option(config)

			if (err != nil) != tt.wantError {
				t.Errorf("Option error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validator != nil && !tt.validator(config) {
				t.Error("Option validation failed")
			}
		})
	}
}

func TestWithMiddleware(t *testing.T) {
	config := NewBaseConfig()
	middleware := func(next http.HandlerFunc) http.HandlerFunc {
		return next
	}

	// Test valid middleware
	err := WithMiddleware(middleware)(config)
	if err != nil {
		t.Errorf("WithMiddleware error = %v, want nil", err)
	}

	if len(config.GetMiddlewares()) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(config.GetMiddlewares()))
	}

	// Test nil middleware
	err = WithMiddleware(nil)(config)
	if err == nil {
		t.Error("WithMiddleware with nil should return error")
	}
}

func TestWithObserver(t *testing.T) {
	config := NewBaseConfig()
	observer := &mockObserver{}

	// Test valid observer
	err := WithObserver(observer)(config)
	if err != nil {
		t.Errorf("WithObserver error = %v, want nil", err)
	}

	if len(config.GetObservers()) != 1 {
		t.Errorf("Expected 1 observer, got %d", len(config.GetObservers()))
	}

	// Test nil observer
	err = WithObserver(nil)(config)
	if err == nil {
		t.Error("WithObserver with nil should return error")
	}
}

func TestWithHook(t *testing.T) {
	config := NewBaseConfig()
	hook := func(ctx context.Context, data interface{}) error {
		return nil
	}

	// Test valid hook
	err := WithHook(interfaces.EventStart, hook)(config)
	if err != nil {
		t.Errorf("WithHook error = %v, want nil", err)
	}

	hooks := config.GetHooks(interfaces.EventStart)
	if len(hooks) != 1 {
		t.Errorf("Expected 1 hook, got %d", len(hooks))
	}

	// Test nil hook
	err = WithHook(interfaces.EventStart, nil)(config)
	if err == nil {
		t.Error("WithHook with nil should return error")
	}
}

func TestBuilder(t *testing.T) {
	// Test successful build
	config, err := NewBuilder().
		Apply(WithAddr("127.0.0.1")).
		Apply(WithPort(9000)).
		Apply(WithProvider("gin")).
		Build()

	if err != nil {
		t.Errorf("Builder.Build() error = %v, want nil", err)
	}

	if config.GetAddr() != "127.0.0.1" {
		t.Errorf("Expected addr '127.0.0.1', got '%s'", config.GetAddr())
	}

	if config.GetPort() != 9000 {
		t.Errorf("Expected port 9000, got %d", config.GetPort())
	}

	if config.GetProvider() != "gin" {
		t.Errorf("Expected provider 'gin', got '%s'", config.GetProvider())
	}
}

func TestBuilderWithErrors(t *testing.T) {
	// Test build with option errors
	_, err := NewBuilder().
		Apply(WithAddr("")).
		Apply(WithPort(0)).
		Build()

	if err == nil {
		t.Error("Builder.Build() should return error for invalid options")
	}
}

func TestBuilderMustBuild(t *testing.T) {
	// Test successful MustBuild
	config := NewBuilder().
		Apply(WithAddr("127.0.0.1")).
		Apply(WithPort(9000)).
		MustBuild()

	if config.GetAddr() != "127.0.0.1" {
		t.Errorf("Expected addr '127.0.0.1', got '%s'", config.GetAddr())
	}

	// Test panic on error
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustBuild should panic on error")
		}
	}()

	NewBuilder().
		Apply(WithAddr("")).
		MustBuild()
}

// Mock observer for testing
type mockObserver struct{}

func (m *mockObserver) OnStart(ctx context.Context, addr string) error {
	return nil
}

func (m *mockObserver) OnStop(ctx context.Context) error {
	return nil
}

func (m *mockObserver) OnError(ctx context.Context, err error) error {
	return nil
}

func (m *mockObserver) OnRequest(ctx context.Context, req interface{}) error {
	return nil
}

func (m *mockObserver) OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error {
	return nil
}

func (m *mockObserver) OnRouteEnter(ctx context.Context, method, path string, req interface{}) error {
	return nil
}

func (m *mockObserver) OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error {
	return nil
}
