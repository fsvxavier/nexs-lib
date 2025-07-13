package fibertrace

import (
	"math"
	"os"
	"testing"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockFiberCtx struct {
	mock.Mock
}

func (m *mockFiberCtx) Route() fiber.Route {
	args := m.Called()
	return args.Get(0).(fiber.Route)
}

type mockRoute struct {
	mock.Mock
}

func (m *mockRoute) Method() string {
	return m.Called().String(0)
}

func (m *mockRoute) Path() string {
	return m.Called().String(0)
}

func TestDefaults(t *testing.T) {
	// Backup the APP_NAME and restore it later
	oldAppName := os.Getenv("APP_NAME")
	os.Setenv("APP_NAME", "test-app")
	defer os.Setenv("APP_NAME", oldAppName)

	cfg := &config{}
	defaults(cfg)

	assert.Equal(t, "test-app", cfg.serviceName)
	assert.Equal(t, defaultSpanName, cfg.spanName)
	assert.Equal(t, defaultEnvironment, cfg.environment)
	assert.NotEmpty(t, cfg.uuid)
	assert.True(t, math.IsNaN(cfg.analyticsRate))

	// Test function assignments
	assert.Equal(t, true, cfg.isStatusError(500))
	assert.Equal(t, false, cfg.isStatusError(404))
	assert.Equal(t, false, cfg.ignoreRequest(nil))
}

func TestOptionApplications(t *testing.T) {
	tests := []struct {
		name     string
		option   OptionFn
		validate func(*testing.T, *config)
	}{
		{
			name:   "WithEnvironment",
			option: WithEnvironment("prod"),
			validate: func(t *testing.T, cfg *config) {
				assert.Equal(t, "prod", cfg.environment)
			},
		},
		{
			name:   "WithTraceID",
			option: WithTraceID("test-trace-id"),
			validate: func(t *testing.T, cfg *config) {
				assert.Equal(t, "test-trace-id", cfg.uuid)
			},
		},
		{
			name:   "WithService",
			option: WithService("test-service"),
			validate: func(t *testing.T, cfg *config) {
				assert.Equal(t, "test-service", cfg.serviceName)
			},
		},
		{
			name:   "WithSpanOptions",
			option: WithSpanOptions(tracer.Tag("key", "value")),
			validate: func(t *testing.T, cfg *config) {
				assert.Len(t, cfg.spanOpts, 1)
			},
		},
		{
			name:   "WithAnalytics true",
			option: WithAnalytics(true),
			validate: func(t *testing.T, cfg *config) {
				assert.Equal(t, 1.0, cfg.analyticsRate)
			},
		},
		{
			name:   "WithAnalytics false",
			option: WithAnalytics(false),
			validate: func(t *testing.T, cfg *config) {
				assert.True(t, math.IsNaN(cfg.analyticsRate))
			},
		},
		{
			name:   "WithAnalyticsRate valid",
			option: WithAnalyticsRate(0.5),
			validate: func(t *testing.T, cfg *config) {
				assert.Equal(t, 0.5, cfg.analyticsRate)
			},
		},
		{
			name:   "WithAnalyticsRate invalid",
			option: WithAnalyticsRate(2.0),
			validate: func(t *testing.T, cfg *config) {
				assert.True(t, math.IsNaN(cfg.analyticsRate))
			},
		},
		{
			name: "WithStatusCheck",
			option: WithStatusCheck(func(code int) bool {
				return code >= 400
			}),
			validate: func(t *testing.T, cfg *config) {
				assert.True(t, cfg.isStatusError(400))
				assert.True(t, cfg.isStatusError(500))
				assert.False(t, cfg.isStatusError(200))
			},
		},
		{
			name: "WithResourceNamer",
			option: WithResourceNamer(func(c *fiber.Ctx) string {
				return "custom-resource"
			}),
			validate: func(t *testing.T, cfg *config) {
				assert.Equal(t, "custom-resource", cfg.resourceNamer(nil))
			},
		},
		{
			name: "WithIgnoreRequest",
			option: WithIgnoreRequest(func(c *fiber.Ctx) bool {
				return true
			}),
			validate: func(t *testing.T, cfg *config) {
				assert.True(t, cfg.ignoreRequest(nil))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{}
			tt.option.apply(cfg)
			tt.validate(t, cfg)
		})
	}
}

func TestIsServerError(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{200, false},
		{400, false},
		{499, false},
		{500, true},
		{501, true},
		{599, true},
		{600, false},
	}

	for _, tt := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			assert.Equal(t, tt.expected, isServerError(tt.statusCode))
		})
	}
}
