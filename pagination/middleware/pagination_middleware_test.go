package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fsvxavier/nexs-lib/pagination"
	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginationMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedPage  int
		expectedLimit int
		expectedField string
		expectedOrder string
		shouldError   bool
	}{
		{
			name:          "Default pagination parameters",
			url:           "/users",
			expectedPage:  1,
			expectedLimit: 50, // Default limit from config
		},
		{
			name:          "Custom pagination parameters",
			url:           "/users?page=2&limit=20&sort=name&order=desc",
			expectedPage:  2,
			expectedLimit: 20,
			expectedField: "name",
			expectedOrder: "desc",
		},
		{
			name:        "Invalid page parameter",
			url:         "/users?page=invalid",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware configuration
			cfg := DefaultPaginationConfig()
			cfg.SortableFields["/users"] = []string{"id", "name", "created_at"}

			// Create test handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				params := GetPaginationParams(r)
				if params != nil {
					assert.Equal(t, tt.expectedPage, params.Page)
					assert.Equal(t, tt.expectedLimit, params.Limit)
					if tt.expectedField != "" {
						assert.Equal(t, tt.expectedField, params.SortField)
					}
					if tt.expectedOrder != "" {
						assert.Equal(t, tt.expectedOrder, params.SortOrder)
					}
				}
				w.WriteHeader(http.StatusOK)
			})

			// Create middleware
			middleware := PaginationMiddleware(cfg)
			middlewareHandler := middleware(handler)

			// Create request
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			// Execute request
			middlewareHandler.ServeHTTP(w, req)

			if tt.shouldError {
				assert.Equal(t, http.StatusBadRequest, w.Code)
			} else {
				assert.Equal(t, http.StatusOK, w.Code)
			}
		})
	}
}

func TestPaginatedResponseMiddleware(t *testing.T) {
	service := pagination.NewPaginationService(nil)

	// Create test handler that returns data
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate returning some data
		data := []map[string]interface{}{
			{"id": 1, "name": "User 1"},
			{"id": 2, "name": "User 2"},
		}

		// Set total count header
		w.Header().Set("X-Total-Count", "10")

		json.NewEncoder(w).Encode(data)
	})

	// Create pagination middleware first
	paginationCfg := DefaultPaginationConfig()
	paginationMiddleware := PaginationMiddleware(paginationCfg)

	// Create response middleware
	responseMiddleware := PaginatedResponseMiddleware(service)

	// Chain middlewares
	chainedHandler := paginationMiddleware(responseMiddleware(handler))

	// Create request with pagination params
	req := httptest.NewRequest("GET", "/users?page=1&limit=2", nil)
	w := httptest.NewRecorder()

	// Execute request
	chainedHandler.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response
	var response interfaces.PaginatedResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	// Verify structure
	assert.NotNil(t, response.Content)
	assert.NotNil(t, response.Metadata)
	assert.Equal(t, 1, response.Metadata.CurrentPage)
	assert.Equal(t, 2, response.Metadata.RecordsPerPage)
	assert.Equal(t, 10, response.Metadata.TotalRecords)
	assert.Equal(t, 5, response.Metadata.TotalPages)
}

func TestHookConfiguration(t *testing.T) {
	cfg := DefaultPaginationConfig()

	// Test pre-validation hook
	var hookExecuted bool

	// Create a simple hook implementation
	hook := &testHook{
		executor: func(ctx context.Context, data interface{}) error {
			hookExecuted = true
			return nil
		},
	}

	// Configure hooks using fluent interface
	hookCfg := cfg.WithHooks().
		PreValidation(hook).
		Done()

	// Verify configuration
	assert.NotNil(t, hookCfg)
	assert.Equal(t, cfg, hookCfg)

	// Create middleware and test
	middleware := PaginationMiddleware(cfg)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware(handler).ServeHTTP(w, req)

	assert.True(t, hookExecuted)
	assert.Equal(t, http.StatusOK, w.Code)
}

// testHook is a simple implementation of the Hook interface for testing
type testHook struct {
	executor func(ctx context.Context, data interface{}) error
}

func (h *testHook) Execute(ctx context.Context, data interface{}) error {
	if h.executor != nil {
		return h.executor(ctx, data)
	}
	return nil
}

func TestSkipPaths(t *testing.T) {
	cfg := DefaultPaginationConfig()
	cfg.AddSkipPath("/health")
	cfg.AddSkipPath("/metrics")

	// Test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should not have pagination params for skipped paths
		params := GetPaginationParams(r)
		assert.Nil(t, params)
		w.WriteHeader(http.StatusOK)
	})

	middleware := PaginationMiddleware(cfg)
	middlewareHandler := middleware(handler)

	// Test skipped paths
	skipPaths := []string{"/health", "/metrics", "/health/check"}
	for _, path := range skipPaths {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()

		middlewareHandler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}
