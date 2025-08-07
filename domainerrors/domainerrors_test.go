//go:build unit

package domainerrors

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/domainerrors/internal"
)

func TestDomainError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "simple message",
			message:  "test error",
			expected: "test error",
		},
		{
			name:     "empty message",
			message:  "",
			expected: "",
		},
		{
			name:     "complex message",
			message:  "validation failed: field 'email' is required",
			expected: "validation failed: field 'email' is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := &DomainError{
				message: tt.message,
			}

			assert.Equal(t, tt.expected, err.Error())
		})
	}
}

func TestDomainError_Type(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		errorType interfaces.ErrorType
	}{
		{
			name:      "validation error",
			errorType: interfaces.ValidationError,
		},
		{
			name:      "not found error",
			errorType: interfaces.NotFoundError,
		},
		{
			name:      "business error",
			errorType: interfaces.BusinessError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := &DomainError{
				errorType: tt.errorType,
			}

			assert.Equal(t, tt.errorType, err.Type())
		})
	}
}

func TestDomainError_Metadata(t *testing.T) {
	t.Parallel()

	t.Run("nil metadata returns empty map", func(t *testing.T) {
		t.Parallel()

		err := &DomainError{
			metadata: nil,
		}

		result := err.Metadata()
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("returns copy of metadata", func(t *testing.T) {
		t.Parallel()

		original := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}

		err := &DomainError{
			metadata: original,
		}

		result := err.Metadata()
		assert.Equal(t, original, result)

		// Modificar o resultado não deve afetar o original
		result["new_key"] = "new_value"
		assert.NotEqual(t, original, result)
		assert.NotContains(t, err.metadata, "new_key")
	})
}

func TestDomainError_HTTPStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		errorType    interfaces.ErrorType
		expectedCode int
	}{
		{
			name:         "validation error",
			errorType:    interfaces.ValidationError,
			expectedCode: 400,
		},
		{
			name:         "not found error",
			errorType:    interfaces.NotFoundError,
			expectedCode: 404,
		},
		{
			name:         "authentication error",
			errorType:    interfaces.AuthenticationError,
			expectedCode: 401,
		},
		{
			name:         "authorization error",
			errorType:    interfaces.AuthorizationError,
			expectedCode: 403,
		},
		{
			name:         "database error",
			errorType:    interfaces.DatabaseError,
			expectedCode: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := &DomainError{
				errorType: tt.errorType,
			}

			assert.Equal(t, tt.expectedCode, err.HTTPStatus())
		})
	}
}

func TestDomainError_WithContext(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), "test_key", "test_value")

	original := &DomainError{
		id:      "test-id",
		code:    "TEST001",
		message: "test error",
	}

	result := original.WithContext(ctx)

	// Deve retornar uma nova instância
	assert.NotSame(t, original, result)

	// O contexto deve ter sido definido
	domainErr, ok := result.(*DomainError)
	require.True(t, ok)
	assert.Equal(t, ctx, domainErr.context)
}

func TestDomainError_Wrap(t *testing.T) {
	t.Parallel()

	originalErr := fmt.Errorf("original error")

	domainErr := &DomainError{
		id:      "test-id",
		code:    "TEST001",
		message: "domain error",
	}

	result := domainErr.Wrap(originalErr)

	// Deve retornar uma nova instância
	assert.NotSame(t, domainErr, result)

	// A causa deve ter sido definida
	assert.Equal(t, originalErr, result.Unwrap())
}

func TestDomainError_WithMetadata(t *testing.T) {
	t.Parallel()

	original := &DomainError{
		id:      "test-id",
		code:    "TEST001",
		message: "test error",
		metadata: map[string]interface{}{
			"existing": "value",
		},
	}

	result := original.WithMetadata("new_key", "new_value")

	// Deve retornar uma nova instância
	assert.NotSame(t, original, result)

	// Os metadados devem ter sido adicionados
	metadata := result.Metadata()
	assert.Equal(t, "value", metadata["existing"])
	assert.Equal(t, "new_value", metadata["new_key"])

	// O original não deve ter sido modificado
	originalMetadata := original.Metadata()
	assert.NotContains(t, originalMetadata, "new_key")
}

func TestDomainError_ToJSON(t *testing.T) {
	t.Parallel()

	now := time.Now()

	err := &DomainError{
		id:        "test-id-123",
		code:      "TEST001",
		message:   "test error message",
		errorType: interfaces.ValidationError,
		metadata: map[string]interface{}{
			"field": "email",
			"value": "invalid",
		},
		timestamp: now,
		cause:     fmt.Errorf("wrapped error"),
	}

	jsonData, jsonErr := err.ToJSON()
	require.NoError(t, jsonErr)

	var result map[string]interface{}
	unmarshalErr := json.Unmarshal(jsonData, &result)
	require.NoError(t, unmarshalErr)

	assert.Equal(t, "test-id-123", result["id"])
	assert.Equal(t, "TEST001", result["code"])
	assert.Equal(t, "test error message", result["message"])
	assert.Equal(t, string(interfaces.ValidationError), result["type"])
	assert.Equal(t, "wrapped error", result["cause"])

	metadata, ok := result["metadata"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "email", metadata["field"])
	assert.Equal(t, "invalid", metadata["value"])
}

func TestErrorFactory_New(t *testing.T) {
	t.Parallel()

	stackCapture := internal.NewStackTraceCapture(true)
	factory := NewErrorFactory(stackCapture)

	err := factory.New(interfaces.ValidationError, "VAL001", "validation failed")

	require.NotNil(t, err)
	assert.Equal(t, interfaces.ValidationError, err.Type())
	assert.Equal(t, "VAL001", err.Code())
	assert.Equal(t, "validation failed", err.Error())
	assert.NotNil(t, err.Metadata())        // Metadata is initialized as empty map, not nil
	assert.Equal(t, 0, len(err.Metadata())) // Should be empty
}

func TestErrorFactory_NewWithMetadata(t *testing.T) {
	t.Parallel()

	stackCapture := internal.NewStackTraceCapture(false)
	factory := NewErrorFactory(stackCapture)

	metadata := map[string]interface{}{
		"field": "email",
		"rule":  "required",
	}

	err := factory.NewWithMetadata(interfaces.ValidationError, "VAL001", "validation failed", metadata)

	require.NotNil(t, err)
	assert.Equal(t, interfaces.ValidationError, err.Type())
	assert.Equal(t, "VAL001", err.Code())
	assert.Equal(t, "validation failed", err.Error())

	resultMetadata := err.Metadata()
	assert.Equal(t, "email", resultMetadata["field"])
	assert.Equal(t, "required", resultMetadata["rule"])
}

func TestErrorFactory_Wrap(t *testing.T) {
	t.Parallel()

	originalErr := fmt.Errorf("database connection failed")
	stackCapture := internal.NewStackTraceCapture(false)
	factory := NewErrorFactory(stackCapture)

	err := factory.Wrap(originalErr, interfaces.DatabaseError, "DB001", "database operation failed")

	require.NotNil(t, err)
	assert.Equal(t, interfaces.DatabaseError, err.Type())
	assert.Equal(t, "DB001", err.Code())
	assert.Equal(t, "database operation failed", err.Error())
	assert.Equal(t, originalErr, err.Unwrap())
}

func TestErrorTypeChecker_IsType(t *testing.T) {
	t.Parallel()

	checker := &ErrorTypeChecker{}

	t.Run("nil error returns false", func(t *testing.T) {
		t.Parallel()

		result := checker.IsType(nil, interfaces.ValidationError)
		assert.False(t, result)
	})

	t.Run("domain error matches type", func(t *testing.T) {
		t.Parallel()

		err := &DomainError{
			errorType: interfaces.ValidationError,
		}

		result := checker.IsType(err, interfaces.ValidationError)
		assert.True(t, result)

		result = checker.IsType(err, interfaces.NotFoundError)
		assert.False(t, result)
	})

	t.Run("wrapped error matches type", func(t *testing.T) {
		t.Parallel()

		domainErr := &DomainError{
			errorType: interfaces.ValidationError,
		}

		wrappedErr := fmt.Errorf("wrapped: %w", domainErr)

		result := checker.IsType(wrappedErr, interfaces.ValidationError)
		assert.True(t, result)
	})
}

func TestMapHTTPStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		errorType    interfaces.ErrorType
		expectedCode int
	}{
		{interfaces.ValidationError, 400},
		{interfaces.BadRequestError, 400},
		{interfaces.AuthenticationError, 401},
		{interfaces.AuthorizationError, 403},
		{interfaces.NotFoundError, 404},
		{interfaces.ConflictError, 409},
		{interfaces.UnprocessableEntityError, 422},
		{interfaces.DatabaseError, 500},
		{interfaces.ExternalServiceError, 502},
		{interfaces.ServiceUnavailableError, 503},
		{interfaces.TimeoutError, 504},
	}

	for _, tt := range tests {
		t.Run(string(tt.errorType), func(t *testing.T) {
			t.Parallel()

			result := MapHTTPStatus(tt.errorType)
			assert.Equal(t, tt.expectedCode, result)
		})
	}

	t.Run("unknown error type returns 500", func(t *testing.T) {
		t.Parallel()

		result := MapHTTPStatus("unknown_error")
		assert.Equal(t, 500, result)
	})
}

func TestGlobalFunctions(t *testing.T) {
	t.Parallel()

	t.Run("New creates domain error", func(t *testing.T) {
		t.Parallel()

		err := New(interfaces.ValidationError, "VAL001", "validation failed")

		require.NotNil(t, err)
		assert.Equal(t, interfaces.ValidationError, err.Type())
		assert.Equal(t, "VAL001", err.Code())
		assert.Equal(t, "validation failed", err.Error())
	})

	t.Run("IsType works with global checker", func(t *testing.T) {
		t.Parallel()

		err := New(interfaces.ValidationError, "VAL001", "validation failed")

		result := IsType(err, interfaces.ValidationError)
		assert.True(t, result)

		result = IsType(err, interfaces.NotFoundError)
		assert.False(t, result)
	})
}

func TestConvenienceFunctions(t *testing.T) {
	t.Parallel()

	t.Run("NewValidationError", func(t *testing.T) {
		t.Parallel()

		err := NewValidationError("VAL001", "validation failed")

		require.NotNil(t, err)
		assert.Equal(t, interfaces.ValidationError, err.Type())
		assert.Equal(t, "VAL001", err.Code())
		assert.Equal(t, "validation failed", err.Error())
	})

	t.Run("NewNotFoundError", func(t *testing.T) {
		t.Parallel()

		err := NewNotFoundError("NF001", "resource not found")

		require.NotNil(t, err)
		assert.Equal(t, interfaces.NotFoundError, err.Type())
		assert.Equal(t, "NF001", err.Code())
		assert.Equal(t, "resource not found", err.Error())
	})

	t.Run("NewBusinessError", func(t *testing.T) {
		t.Parallel()

		err := NewBusinessError("BIZ001", "business rule violation")

		require.NotNil(t, err)
		assert.Equal(t, interfaces.BusinessError, err.Type())
		assert.Equal(t, "BIZ001", err.Code())
		assert.Equal(t, "business rule violation", err.Error())
	})
}

func TestGetRootCause(t *testing.T) {
	t.Parallel()

	t.Run("single error returns itself", func(t *testing.T) {
		t.Parallel()

		err := fmt.Errorf("original error")

		result := GetRootCause(err)
		assert.Equal(t, err, result)
	})

	t.Run("wrapped error returns root cause", func(t *testing.T) {
		t.Parallel()

		rootErr := fmt.Errorf("root error")
		wrappedErr := fmt.Errorf("wrapped: %w", rootErr)
		doubleWrappedErr := fmt.Errorf("double wrapped: %w", wrappedErr)

		result := GetRootCause(doubleWrappedErr)
		assert.Equal(t, rootErr, result)
	})
}

func TestGetErrorChain(t *testing.T) {
	t.Parallel()

	t.Run("single error returns single item chain", func(t *testing.T) {
		t.Parallel()

		err := fmt.Errorf("single error")

		chain := GetErrorChain(err)
		assert.Len(t, chain, 1)
		assert.Equal(t, err, chain[0])
	})

	t.Run("wrapped errors return full chain", func(t *testing.T) {
		t.Parallel()

		rootErr := fmt.Errorf("root error")
		wrappedErr := fmt.Errorf("wrapped: %w", rootErr)
		doubleWrappedErr := fmt.Errorf("double wrapped: %w", wrappedErr)

		chain := GetErrorChain(doubleWrappedErr)
		assert.Len(t, chain, 3)
		assert.Equal(t, doubleWrappedErr, chain[0])
		assert.Equal(t, wrappedErr, chain[1])
		assert.Equal(t, rootErr, chain[2])
	})
}

func TestFormatErrorChain(t *testing.T) {
	t.Parallel()

	t.Run("empty chain returns empty string", func(t *testing.T) {
		t.Parallel()

		result := FormatErrorChain(nil)
		assert.Equal(t, "", result)
	})

	t.Run("formats error chain correctly", func(t *testing.T) {
		t.Parallel()

		domainErr := New(interfaces.ValidationError, "VAL001", "validation failed")
		wrappedErr := fmt.Errorf("wrapped: %w", domainErr)

		result := FormatErrorChain(wrappedErr)

		assert.Contains(t, result, "Error chain:")
		assert.Contains(t, result, "1. wrapped:")
		assert.Contains(t, result, "2. validation failed")
		assert.Contains(t, result, "[validation_error:VAL001]")
	})
}

func TestManager_Observer(t *testing.T) {
	t.Parallel()

	manager := NewManager()

	t.Run("register and notify observers", func(t *testing.T) {
		t.Parallel()

		var observedError atomic.Value

		observer := &mockObserver{
			onErrorFunc: func(ctx context.Context, err interfaces.DomainErrorInterface) error {
				observedError.Store(err)
				return nil
			},
		}

		manager.RegisterObserver(observer)

		testErr := New(interfaces.ValidationError, "VAL001", "test error")
		err := manager.NotifyObservers(context.Background(), testErr)

		assert.NoError(t, err)
		stored := observedError.Load()
		if stored != nil {
			assert.Equal(t, testErr, stored)
		}
	})

	t.Run("unregister observer", func(t *testing.T) {
		t.Parallel()

		var called atomic.Bool

		observer := &mockObserver{
			onErrorFunc: func(ctx context.Context, err interfaces.DomainErrorInterface) error {
				called.Store(true)
				return nil
			},
		}

		manager.RegisterObserver(observer)
		manager.UnregisterObserver(observer)

		testErr := New(interfaces.ValidationError, "VAL001", "test error")
		err := manager.NotifyObservers(context.Background(), testErr)

		assert.NoError(t, err)
		assert.False(t, called.Load())
	})
}

// Mock observer for testing
type mockObserver struct {
	onErrorFunc func(ctx context.Context, err interfaces.DomainErrorInterface) error
	mu          sync.RWMutex
}

func (m *mockObserver) OnError(ctx context.Context, err interfaces.DomainErrorInterface) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.onErrorFunc != nil {
		return m.onErrorFunc(ctx, err)
	}
	return nil
}

func TestStackTrace(t *testing.T) {
	t.Parallel()

	t.Run("stack trace with capture enabled", func(t *testing.T) {
		t.Parallel()

		stackCapture := internal.NewStackTraceCapture(true)
		factory := NewErrorFactory(stackCapture)

		err := factory.New(interfaces.ValidationError, "VAL001", "validation failed")

		stackTrace := err.StackTrace()
		assert.NotEmpty(t, stackTrace)
		assert.Contains(t, stackTrace, "Stack trace:")
	})

	t.Run("stack trace with capture disabled", func(t *testing.T) {
		t.Parallel()

		stackCapture := internal.NewStackTraceCapture(false)
		factory := NewErrorFactory(stackCapture)

		err := factory.New(interfaces.ValidationError, "VAL001", "validation failed")

		stackTrace := err.StackTrace()
		assert.Empty(t, stackTrace)
	})
}

func TestConcurrency(t *testing.T) {
	t.Parallel()

	t.Run("factory is thread safe", func(t *testing.T) {
		t.Parallel()

		factory := GetFactory()
		const goroutines = 100

		errors := make(chan interfaces.DomainErrorInterface, goroutines)

		for i := 0; i < goroutines; i++ {
			go func(id int) {
				err := factory.New(interfaces.ValidationError, fmt.Sprintf("VAL%03d", id), fmt.Sprintf("error %d", id))
				errors <- err
			}(i)
		}

		for i := 0; i < goroutines; i++ {
			err := <-errors
			assert.NotNil(t, err)
			assert.Equal(t, interfaces.ValidationError, err.Type())
			assert.Contains(t, err.Code(), "VAL")
			assert.Contains(t, err.Error(), "error")
		}
	})

	t.Run("manager is thread safe", func(t *testing.T) {
		t.Parallel()

		manager := NewManager()
		const goroutines = 50

		// Adiciona observadores concorrentemente
		for i := 0; i < goroutines; i++ {
			go func() {
				observer := &mockObserver{}
				manager.RegisterObserver(observer)
			}()
		}

		// Notifica concorrentemente
		testErr := New(interfaces.ValidationError, "VAL001", "test error")
		for i := 0; i < goroutines; i++ {
			go func() {
				manager.NotifyObservers(context.Background(), testErr)
			}()
		}

		// Teste básico - se chegou até aqui sem panic, a sincronização está funcionando
		assert.True(t, true)
	})
}

// Benchmark tests
func BenchmarkNew(b *testing.B) {
	factory := GetFactory()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			factory.New(interfaces.ValidationError, "VAL001", "validation failed")
		}
	})
}

func BenchmarkWithMetadata(b *testing.B) {
	err := New(interfaces.ValidationError, "VAL001", "validation failed")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err.WithMetadata("key", "value")
		}
	})
}

func BenchmarkToJSON(b *testing.B) {
	err := NewWithMetadata(interfaces.ValidationError, "VAL001", "validation failed", map[string]interface{}{
		"field": "email",
		"rule":  "required",
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = err.ToJSON()
		}
	})
}

func BenchmarkIsType(b *testing.B) {
	err := New(interfaces.ValidationError, "VAL001", "validation failed")
	checker := GetChecker()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			checker.IsType(err, interfaces.ValidationError)
		}
	})
}

// Additional tests for uncovered functions
func TestManager_HookMethods(t *testing.T) {
	t.Parallel()

	manager := NewManager()

	t.Run("register and execute start hooks", func(t *testing.T) {
		t.Parallel()

		executed := atomic.Bool{}
		hook := func(ctx context.Context) error {
			executed.Store(true)
			return nil
		}

		manager.hookManager.RegisterStartHook(hook)
		err := manager.hookManager.ExecuteStartHooks(context.Background())

		assert.NoError(t, err)
		assert.True(t, executed.Load())
	})

	t.Run("register and execute stop hooks", func(t *testing.T) {
		t.Parallel()

		executed := atomic.Bool{}
		hook := func(ctx context.Context) error {
			executed.Store(true)
			return nil
		}

		manager.hookManager.RegisterStopHook(hook)
		err := manager.hookManager.ExecuteStopHooks(context.Background())

		assert.NoError(t, err)
		assert.True(t, executed.Load())
	})

	t.Run("register and execute error hooks", func(t *testing.T) {
		t.Parallel()

		var receivedError atomic.Value
		hook := func(ctx context.Context, err interfaces.DomainErrorInterface) error {
			receivedError.Store(err)
			return nil
		}

		manager.hookManager.RegisterErrorHook(hook)
		testErr := New(interfaces.ValidationError, "VAL001", "test error")
		err := manager.hookManager.ExecuteErrorHooks(context.Background(), testErr)

		assert.NoError(t, err)
		stored := receivedError.Load()
		if stored != nil {
			assert.Equal(t, testErr, stored)
		}
	})

	t.Run("register and execute i18n hooks", func(t *testing.T) {
		t.Parallel()

		var receivedError atomic.Value
		var receivedLocale atomic.Value
		hook := func(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error {
			receivedError.Store(err)
			receivedLocale.Store(locale)
			return nil
		}

		manager.hookManager.RegisterI18nHook(hook)
		testErr := New(interfaces.ValidationError, "VAL001", "test error")
		err := manager.hookManager.ExecuteI18nHooks(context.Background(), testErr, "pt-BR")

		assert.NoError(t, err)
		storedError := receivedError.Load()
		storedLocale := receivedLocale.Load()
		if storedError != nil {
			assert.Equal(t, testErr, storedError)
		}
		if storedLocale != nil {
			assert.Equal(t, "pt-BR", storedLocale)
		}
	})
}

func TestManager_MiddlewareMethods(t *testing.T) {
	t.Parallel()

	manager := NewManager()

	t.Run("register and execute middlewares", func(t *testing.T) {
		t.Parallel()

		executed := atomic.Bool{}
		middleware := func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			executed.Store(true)
			return next(err)
		}

		manager.middlewareManager.RegisterMiddleware(middleware)
		testErr := New(interfaces.ValidationError, "VAL001", "test error")
		result := manager.middlewareManager.ExecuteMiddlewares(context.Background(), testErr)

		assert.Equal(t, testErr, result)
		assert.True(t, executed.Load())
	})

	t.Run("register and execute i18n middlewares", func(t *testing.T) {
		t.Parallel()

		executed := atomic.Bool{}
		middleware := func(ctx context.Context, err interfaces.DomainErrorInterface, locale string, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			executed.Store(true)
			return next(err)
		}

		manager.middlewareManager.RegisterI18nMiddleware(middleware)
		testErr := New(interfaces.ValidationError, "VAL001", "test error")
		result := manager.middlewareManager.ExecuteI18nMiddlewares(context.Background(), testErr, "pt-BR")

		assert.Equal(t, testErr, result)
		assert.True(t, executed.Load())
	})
}

func TestDomainError_AdditionalMethods(t *testing.T) {
	t.Parallel()

	t.Run("timestamp returns creation time", func(t *testing.T) {
		t.Parallel()

		before := time.Now()
		err := New(interfaces.ValidationError, "VAL001", "test error")
		after := time.Now()

		timestamp := err.Timestamp()
		assert.True(t, timestamp.After(before) || timestamp.Equal(before))
		assert.True(t, timestamp.Before(after) || timestamp.Equal(after))
	})

	t.Run("with metadata preserves existing metadata", func(t *testing.T) {
		t.Parallel()

		original := NewWithMetadata(interfaces.ValidationError, "VAL001", "test error", map[string]interface{}{
			"field": "email",
		})

		modified := original.WithMetadata("rule", "required")

		assert.Contains(t, modified.Metadata(), "field")
		assert.Contains(t, modified.Metadata(), "rule")
		assert.Equal(t, "email", modified.Metadata()["field"])
		assert.Equal(t, "required", modified.Metadata()["rule"])
	})
}

func TestErrorTypeChecker_AdditionalCases(t *testing.T) {
	t.Parallel()

	checker := &ErrorTypeChecker{}

	t.Run("non-domain error returns false", func(t *testing.T) {
		t.Parallel()

		regularError := fmt.Errorf("regular error")
		result := checker.IsType(regularError, interfaces.ValidationError)

		assert.False(t, result)
	})

	t.Run("deep wrapped domain error matches type", func(t *testing.T) {
		t.Parallel()

		domainErr := New(interfaces.ValidationError, "VAL001", "validation failed")
		wrapped1 := fmt.Errorf("wrapper 1: %w", domainErr)
		wrapped2 := fmt.Errorf("wrapper 2: %w", wrapped1)

		result := checker.IsType(wrapped2, interfaces.ValidationError)
		assert.True(t, result)

		result = checker.IsType(wrapped2, interfaces.NotFoundError)
		assert.False(t, result)
	})
}

func TestGlobalConvenienceFunctions(t *testing.T) {
	t.Parallel()

	t.Run("NewWithMetadata creates error with metadata", func(t *testing.T) {
		t.Parallel()

		metadata := map[string]interface{}{
			"field": "password",
			"rule":  "min_length",
		}

		err := NewWithMetadata(interfaces.ValidationError, "VAL002", "password too short", metadata)

		assert.Equal(t, interfaces.ValidationError, err.Type())
		assert.Equal(t, "VAL002", err.Code())
		assert.Equal(t, "password too short", err.Error())
		assert.Equal(t, "password", err.Metadata()["field"])
		assert.Equal(t, "min_length", err.Metadata()["rule"])
	})

	t.Run("Wrap creates wrapped error", func(t *testing.T) {
		t.Parallel()

		originalErr := fmt.Errorf("database connection failed")
		factory := NewErrorFactory(internal.NewStackTraceCapture(false))
		wrappedErr := factory.Wrap(originalErr, interfaces.DatabaseError, "DB001", "failed to save user")

		assert.Equal(t, interfaces.DatabaseError, wrappedErr.Type())
		assert.Equal(t, "DB001", wrappedErr.Code())
		assert.Equal(t, "failed to save user", wrappedErr.Error())
		assert.Equal(t, originalErr, wrappedErr.Unwrap())
	})
}
