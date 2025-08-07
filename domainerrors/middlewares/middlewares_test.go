//go:build unit

package middlewares

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

func TestMiddlewareManager(t *testing.T) {
	t.Parallel()

	t.Run("new manager starts empty", func(t *testing.T) {
		t.Parallel()

		manager := NewMiddlewareManager()
		general, i18n := manager.GetCounts()

		assert.Equal(t, 0, general)
		assert.Equal(t, 0, i18n)
	})

	t.Run("register middleware increases count", func(t *testing.T) {
		t.Parallel()

		manager := NewMiddlewareManager()

		middleware := func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			return next(err)
		}

		manager.RegisterMiddleware(middleware)

		general, _ := manager.GetCounts()
		assert.Equal(t, 1, general)
	})

	t.Run("register nil middleware is ignored", func(t *testing.T) {
		t.Parallel()

		manager := NewMiddlewareManager()
		manager.RegisterMiddleware(nil)

		general, _ := manager.GetCounts()
		assert.Equal(t, 0, general)
	})

	t.Run("execute middlewares", func(t *testing.T) {
		t.Parallel()

		manager := NewMiddlewareManager()

		called1 := false
		called2 := false

		middleware1 := func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			called1 = true
			return next(err.WithMetadata("middleware1", true))
		}

		middleware2 := func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			called2 = true
			return next(err.WithMetadata("middleware2", true))
		}

		manager.RegisterMiddleware(middleware1)
		manager.RegisterMiddleware(middleware2)

		testErr := &mockDomainError{metadata: make(map[string]interface{})}
		result := manager.ExecuteMiddlewares(context.Background(), testErr)

		assert.True(t, called1)
		assert.True(t, called2)
		assert.NotNil(t, result)

		metadata := result.Metadata()
		assert.Equal(t, true, metadata["middleware1"])
		assert.Equal(t, true, metadata["middleware2"])
	})

	t.Run("execute with nil error returns nil", func(t *testing.T) {
		t.Parallel()

		manager := NewMiddlewareManager()

		middleware := func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			return next(err)
		}

		manager.RegisterMiddleware(middleware)

		result := manager.ExecuteMiddlewares(context.Background(), nil)
		assert.Nil(t, result)
	})
}

func TestI18nMiddlewareManager(t *testing.T) {
	t.Parallel()

	t.Run("execute i18n middlewares", func(t *testing.T) {
		t.Parallel()

		manager := NewMiddlewareManager()

		var receivedLocale string

		middleware := func(ctx context.Context, err interfaces.DomainErrorInterface, locale string, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			receivedLocale = locale
			return next(err.WithMetadata("i18n_processed", true))
		}

		manager.RegisterI18nMiddleware(middleware)

		testErr := &mockDomainError{metadata: make(map[string]interface{})}
		result := manager.ExecuteI18nMiddlewares(context.Background(), testErr, "pt-BR")

		assert.Equal(t, "pt-BR", receivedLocale)
		assert.NotNil(t, result)

		metadata := result.Metadata()
		assert.Equal(t, true, metadata["i18n_processed"])
	})
}

func TestLoggingMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("logging middleware adds metadata", func(t *testing.T) {
		t.Parallel()

		testErr := &mockDomainError{metadata: make(map[string]interface{})}

		result := LoggingMiddleware(context.Background(), testErr, func(e interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			return e
		})

		metadata := result.Metadata()
		assert.Equal(t, true, metadata["middleware_logged"])
		assert.Equal(t, "error", metadata["log_level"])
	})

	t.Run("logging middleware with nil error", func(t *testing.T) {
		t.Parallel()

		result := LoggingMiddleware(context.Background(), nil, func(e interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			return e
		})

		assert.Nil(t, result)
	})
}

func TestMetricsMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("metrics middleware adds metadata", func(t *testing.T) {
		t.Parallel()

		testErr := &mockDomainError{
			metadata:  make(map[string]interface{}),
			errorType: interfaces.ValidationError,
		}

		result := MetricsMiddleware(context.Background(), testErr, func(e interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			return e
		})

		metadata := result.Metadata()
		assert.Equal(t, true, metadata["metrics_collected"])
		assert.Equal(t, string(interfaces.ValidationError), metadata["metric_type"])
	})
}

func TestEnrichmentMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("enrichment middleware adds context info", func(t *testing.T) {
		t.Parallel()

		ctx := context.WithValue(context.Background(), "user_id", "user123")
		ctx = context.WithValue(ctx, "request_id", "req456")

		testErr := &mockDomainError{metadata: make(map[string]interface{})}

		result := EnrichmentMiddleware(ctx, testErr, func(e interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			return e
		})

		metadata := result.Metadata()
		assert.Equal(t, true, metadata["context_enriched"])
		assert.Equal(t, "user123", metadata["user_id"])
		assert.Equal(t, "req456", metadata["request_id"])
	})

	t.Run("enrichment middleware with nil context", func(t *testing.T) {
		t.Parallel()

		testErr := &mockDomainError{metadata: make(map[string]interface{})}

		result := EnrichmentMiddleware(nil, testErr, func(e interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			return e
		})

		metadata := result.Metadata()
		assert.Equal(t, true, metadata["context_enriched"])
		// Não deve ter user_id ou request_id sem contexto
		assert.NotContains(t, metadata, "user_id")
		assert.NotContains(t, metadata, "request_id")
	})
}

func TestGlobalMiddlewares(t *testing.T) {
	// Note: These tests cannot be run in parallel as they affect global state

	t.Run("global middlewares", func(t *testing.T) {
		// Clear any existing global middlewares first
		ClearGlobalMiddlewares()

		called := false
		middleware := func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			called = true
			return next(err)
		}

		RegisterGlobalMiddleware(middleware)

		testErr := &mockDomainError{}
		result := ExecuteGlobalMiddlewares(context.Background(), testErr)

		assert.True(t, called)
		assert.NotNil(t, result)

		// Clean up
		ClearGlobalMiddlewares()
	})

	t.Run("global middleware counts", func(t *testing.T) {
		// Clear any existing global middlewares first
		ClearGlobalMiddlewares()

		middleware := func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			return next(err)
		}

		RegisterGlobalMiddleware(middleware)

		general, _ := GetGlobalMiddlewareCounts()
		assert.Equal(t, 1, general)

		// Clean up
		ClearGlobalMiddlewares()
	})
}

func TestMiddlewareChaining(t *testing.T) {
	t.Parallel()

	t.Run("middlewares are executed in order", func(t *testing.T) {
		t.Parallel()

		manager := NewMiddlewareManager()

		var order []string

		middleware1 := func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			order = append(order, "middleware1")
			return next(err)
		}

		middleware2 := func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			order = append(order, "middleware2")
			return next(err)
		}

		middleware3 := func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			order = append(order, "middleware3")
			return next(err)
		}

		manager.RegisterMiddleware(middleware1)
		manager.RegisterMiddleware(middleware2)
		manager.RegisterMiddleware(middleware3)

		testErr := &mockDomainError{}
		manager.ExecuteMiddlewares(context.Background(), testErr)

		expected := []string{"middleware1", "middleware2", "middleware3"}
		assert.Equal(t, expected, order)
	})
}

// Mock domain error for testing
type mockDomainError struct {
	message   string
	errorType interfaces.ErrorType
	metadata  map[string]interface{}
	code      string
}

func (m *mockDomainError) Error() string {
	if m.message == "" {
		return "mock error"
	}
	return m.message
}

func (m *mockDomainError) Unwrap() error { return nil }

func (m *mockDomainError) Type() interfaces.ErrorType {
	if m.errorType == "" {
		return interfaces.ValidationError
	}
	return m.errorType
}

func (m *mockDomainError) Metadata() map[string]interface{} {
	if m.metadata == nil {
		return make(map[string]interface{})
	}
	// Retorna uma cópia
	result := make(map[string]interface{})
	for k, v := range m.metadata {
		result[k] = v
	}
	return result
}

func (m *mockDomainError) HTTPStatus() int                                                 { return 400 }
func (m *mockDomainError) StackTrace() string                                              { return "" }
func (m *mockDomainError) WithContext(ctx context.Context) interfaces.DomainErrorInterface { return m }
func (m *mockDomainError) Wrap(err error) interfaces.DomainErrorInterface                  { return m }

func (m *mockDomainError) WithMetadata(key string, value interface{}) interfaces.DomainErrorInterface {
	if m.metadata == nil {
		m.metadata = make(map[string]interface{})
	}
	m.metadata[key] = value
	return m
}

func (m *mockDomainError) Code() string {
	if m.code == "" {
		return "MOCK001"
	}
	return m.code
}

func (m *mockDomainError) Timestamp() time.Time    { return time.Now() }
func (m *mockDomainError) ToJSON() ([]byte, error) { return []byte(`{}`), nil }

func BenchmarkMiddlewareExecution(b *testing.B) {
	manager := NewMiddlewareManager()

	middleware := func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
		return next(err.WithMetadata("processed", true))
	}

	manager.RegisterMiddleware(middleware)

	testErr := &mockDomainError{metadata: make(map[string]interface{})}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			manager.ExecuteMiddlewares(context.Background(), testErr)
		}
	})
}

func BenchmarkI18nMiddlewareExecution(b *testing.B) {
	manager := NewMiddlewareManager()

	middleware := func(ctx context.Context, err interfaces.DomainErrorInterface, locale string, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
		return next(err.WithMetadata("i18n_processed", true))
	}

	manager.RegisterI18nMiddleware(middleware)

	testErr := &mockDomainError{metadata: make(map[string]interface{})}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			manager.ExecuteI18nMiddlewares(context.Background(), testErr, "en")
		}
	})
}
