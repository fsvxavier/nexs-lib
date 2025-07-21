package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCustomMiddleware(t *testing.T) {
	called := false
	middleware := NewCustomMiddleware(
		"test-middleware",
		100,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.Header().Set("X-Test", "true")
				next.ServeHTTP(w, r)
			})
		},
	)

	// Test basic properties
	if middleware.Name() != "test-middleware" {
		t.Errorf("Expected name 'test-middleware', got '%s'", middleware.Name())
	}

	if middleware.Priority() != 100 {
		t.Errorf("Expected priority 100, got %d", middleware.Priority())
	}

	if !middleware.IsEnabled() {
		t.Error("Expected middleware to be enabled by default")
	}

	// Test wrapping
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if !called {
		t.Error("Expected middleware to be called")
	}

	if rec.Header().Get("X-Test") != "true" {
		t.Error("Expected X-Test header to be set")
	}
}

func TestCustomMiddlewareWithFilters(t *testing.T) {
	called := false
	middleware := NewCustomMiddleware(
		"filtered-middleware",
		100,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				next.ServeHTTP(w, r)
			})
		},
	)

	// Set filters
	middleware.SetPathFilter(func(path string) bool {
		return path == "/allowed"
	})

	middleware.SetMethodFilter(func(method string) bool {
		return method == "POST"
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	// Test with matching path and method
	req := httptest.NewRequest("POST", "/allowed", nil)
	rec := httptest.NewRecorder()

	called = false
	wrapped.ServeHTTP(rec, req)

	if !called {
		t.Error("Expected middleware to be called for matching request")
	}

	// Test with non-matching path
	req = httptest.NewRequest("POST", "/denied", nil)
	rec = httptest.NewRecorder()

	called = false
	wrapped.ServeHTTP(rec, req)

	if called {
		t.Error("Expected middleware not to be called for non-matching path")
	}

	// Test with non-matching method
	req = httptest.NewRequest("GET", "/allowed", nil)
	rec = httptest.NewRecorder()

	called = false
	wrapped.ServeHTTP(rec, req)

	if called {
		t.Error("Expected middleware not to be called for non-matching method")
	}
}

func TestCustomMiddlewareWithSkipPaths(t *testing.T) {
	called := false
	middleware := NewCustomMiddleware(
		"skip-middleware",
		100,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				next.ServeHTTP(w, r)
			})
		},
	)

	middleware.SetSkipPaths([]string{"/health", "/metrics"})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	// Test skipped path
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	called = false
	wrapped.ServeHTTP(rec, req)

	if called {
		t.Error("Expected middleware to be skipped for /health")
	}

	// Test non-skipped path
	req = httptest.NewRequest("GET", "/api/test", nil)
	rec = httptest.NewRecorder()

	called = false
	wrapped.ServeHTTP(rec, req)

	if !called {
		t.Error("Expected middleware to be called for /api/test")
	}
}

func TestCustomMiddlewareWithSkipFunc(t *testing.T) {
	called := false
	middleware := NewCustomMiddleware(
		"skip-func-middleware",
		100,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				next.ServeHTTP(w, r)
			})
		},
	)

	middleware.SetSkipFunc(func(path string) bool {
		return path == "/skip-me"
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware.Wrap(handler)

	// Test skipped path
	req := httptest.NewRequest("GET", "/skip-me", nil)
	rec := httptest.NewRecorder()

	called = false
	wrapped.ServeHTTP(rec, req)

	if called {
		t.Error("Expected middleware to be skipped for /skip-me")
	}

	// Test non-skipped path
	req = httptest.NewRequest("GET", "/process-me", nil)
	rec = httptest.NewRecorder()

	called = false
	wrapped.ServeHTTP(rec, req)

	if !called {
		t.Error("Expected middleware to be called for /process-me")
	}
}

func TestCustomMiddlewareWithBeforeAfter(t *testing.T) {
	beforeCalled := false
	afterCalled := false
	var capturedStatusCode int
	var capturedDuration time.Duration

	middleware := NewCustomMiddleware(
		"before-after-middleware",
		100,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		},
	)

	middleware.SetBeforeFunc(func(w http.ResponseWriter, r *http.Request) {
		beforeCalled = true
		w.Header().Set("X-Before", "true")
	})

	middleware.SetAfterFunc(func(w http.ResponseWriter, r *http.Request, statusCode int, duration time.Duration) {
		afterCalled = true
		capturedStatusCode = statusCode
		capturedDuration = duration
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if !beforeCalled {
		t.Error("Expected before function to be called")
	}

	if !afterCalled {
		t.Error("Expected after function to be called")
	}

	if rec.Header().Get("X-Before") != "true" {
		t.Error("Expected X-Before header to be set")
	}

	if capturedStatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, capturedStatusCode)
	}

	if capturedDuration <= 0 {
		t.Error("Expected positive duration")
	}
}

func TestCustomMiddlewareBuilder(t *testing.T) {
	builder := NewCustomMiddlewareBuilder()

	middleware, err := builder.
		WithName("builder-middleware").
		WithPriority(200).
		WithSkipPaths("/health").
		WithSkipFunc(func(path string) bool {
			return path == "/skip"
		}).
		WithBeforeFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Before", "true")
		}).
		WithAfterFunc(func(w http.ResponseWriter, r *http.Request, statusCode int, duration time.Duration) {
			w.Header().Set("X-Duration", duration.String())
		}).
		WithWrapFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Wrapped", "true")
				next.ServeHTTP(w, r)
			})
		}).
		Build()

	if err != nil {
		t.Errorf("Expected no error building middleware, got %v", err)
	}

	if middleware.Name() != "builder-middleware" {
		t.Errorf("Expected name 'builder-middleware', got '%s'", middleware.Name())
	}

	if middleware.Priority() != 200 {
		t.Errorf("Expected priority 200, got %d", middleware.Priority())
	}
}

func TestCustomMiddlewareBuilderValidation(t *testing.T) {
	// Test missing name
	builder1 := NewCustomMiddlewareBuilder()
	_, err := builder1.
		WithWrapFunc(func(next http.Handler) http.Handler { return next }).
		Build()

	if err == nil {
		t.Error("Expected error for missing name")
	}

	// Test missing wrap function
	builder2 := NewCustomMiddlewareBuilder()
	_, err = builder2.
		WithName("test").
		Build()

	if err == nil {
		t.Error("Expected error for missing wrap function")
	}
}

func TestCustomMiddlewareFactory(t *testing.T) {
	factory := NewCustomMiddlewareFactory()

	// Test simple middleware creation
	middleware := factory.NewSimpleMiddleware(
		"simple",
		100,
		func(next http.Handler) http.Handler { return next },
	)

	if middleware.Name() != "simple" {
		t.Errorf("Expected name 'simple', got '%s'", middleware.Name())
	}

	// Test conditional middleware creation
	conditionalMiddleware := factory.NewConditionalMiddleware(
		"conditional",
		100,
		func(path string) bool { return path != "/skip" },
		func(next http.Handler) http.Handler { return next },
	)

	if conditionalMiddleware.Name() != "conditional" {
		t.Errorf("Expected name 'conditional', got '%s'", conditionalMiddleware.Name())
	}

	// Test timing middleware
	timingMiddleware := factory.NewTimingMiddleware(
		"timing",
		100,
		func(duration time.Duration, path string) {
			// timing handler
		},
	)

	if timingMiddleware.Name() != "timing" {
		t.Errorf("Expected name 'timing', got '%s'", timingMiddleware.Name())
	}

	// Test logging middleware
	loggingMiddleware := factory.NewLoggingMiddleware(
		"logging",
		100,
		func(method, path string, statusCode int, duration time.Duration) {
			// logging handler
		},
	)

	if loggingMiddleware.Name() != "logging" {
		t.Errorf("Expected name 'logging', got '%s'", loggingMiddleware.Name())
	}
}

func TestCustomMiddlewareDisabling(t *testing.T) {
	called := false
	middleware := NewCustomMiddleware(
		"test-middleware",
		100,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				next.ServeHTTP(w, r)
			})
		},
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Disable the middleware
	middleware.SetEnabled(false)

	wrapped := middleware.Wrap(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	// Middleware should not be called when disabled
	if called {
		t.Error("Expected disabled middleware not to be called")
	}

	// Re-enable the middleware and create a new wrapper
	middleware.SetEnabled(true)
	wrapped = middleware.Wrap(handler)

	called = false
	wrapped.ServeHTTP(rec, req)

	if !called {
		t.Error("Expected enabled middleware to be called")
	}
}

func TestResponseWriterCapture(t *testing.T) {
	rw := &responseWriter{
		ResponseWriter: httptest.NewRecorder(),
		statusCode:     http.StatusOK,
	}

	// Test default status code
	if rw.statusCode != http.StatusOK {
		t.Errorf("Expected default status code %d, got %d", http.StatusOK, rw.statusCode)
	}

	// Test WriteHeader
	rw.WriteHeader(http.StatusNotFound)

	if rw.statusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rw.statusCode)
	}
}
