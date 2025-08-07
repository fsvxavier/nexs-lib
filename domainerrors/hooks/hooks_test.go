//go:build unit

package hooks

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

func TestStartHookManager(t *testing.T) {
	t.Parallel()

	t.Run("new manager starts empty", func(t *testing.T) {
		t.Parallel()

		manager := NewStartHookManager()
		assert.Equal(t, 0, manager.Count())
	})

	t.Run("register hook increases count", func(t *testing.T) {
		t.Parallel()

		manager := NewStartHookManager()

		hook := func(ctx context.Context) error { return nil }
		manager.Register(hook)

		assert.Equal(t, 1, manager.Count())
	})

	t.Run("register nil hook is ignored", func(t *testing.T) {
		t.Parallel()

		manager := NewStartHookManager()
		manager.Register(nil)

		assert.Equal(t, 0, manager.Count())
	})

	t.Run("execute calls all hooks", func(t *testing.T) {
		t.Parallel()

		manager := NewStartHookManager()

		called1 := false
		called2 := false

		hook1 := func(ctx context.Context) error {
			called1 = true
			return nil
		}

		hook2 := func(ctx context.Context) error {
			called2 = true
			return nil
		}

		manager.Register(hook1)
		manager.Register(hook2)

		err := manager.Execute(context.Background())

		assert.NoError(t, err)
		assert.True(t, called1)
		assert.True(t, called2)
	})

	t.Run("execute stops on first error", func(t *testing.T) {
		t.Parallel()

		manager := NewStartHookManager()

		called1 := false
		called2 := false

		hook1 := func(ctx context.Context) error {
			called1 = true
			return assert.AnError
		}

		hook2 := func(ctx context.Context) error {
			called2 = true
			return nil
		}

		manager.Register(hook1)
		manager.Register(hook2)

		err := manager.Execute(context.Background())

		assert.Error(t, err)
		assert.True(t, called1)
		assert.False(t, called2)
	})

	t.Run("clear removes all hooks", func(t *testing.T) {
		t.Parallel()

		manager := NewStartHookManager()

		hook := func(ctx context.Context) error { return nil }
		manager.Register(hook)
		manager.Register(hook)

		assert.Equal(t, 2, manager.Count())

		manager.Clear()
		assert.Equal(t, 0, manager.Count())
	})
}

func TestStopHookManager(t *testing.T) {
	t.Parallel()

	t.Run("new manager starts empty", func(t *testing.T) {
		t.Parallel()

		manager := NewStopHookManager()
		assert.Equal(t, 0, manager.Count())
	})

	t.Run("execute calls all hooks", func(t *testing.T) {
		t.Parallel()

		manager := NewStopHookManager()

		called1 := false
		called2 := false

		hook1 := func(ctx context.Context) error {
			called1 = true
			return nil
		}

		hook2 := func(ctx context.Context) error {
			called2 = true
			return nil
		}

		manager.Register(hook1)
		manager.Register(hook2)

		err := manager.Execute(context.Background())

		assert.NoError(t, err)
		assert.True(t, called1)
		assert.True(t, called2)
	})
}

func TestErrorHookManager(t *testing.T) {
	t.Parallel()

	t.Run("execute with nil error returns nil", func(t *testing.T) {
		t.Parallel()

		manager := NewErrorHookManager()
		err := manager.Execute(context.Background(), nil)

		assert.NoError(t, err)
	})

	t.Run("execute calls all hooks with error", func(t *testing.T) {
		t.Parallel()

		manager := NewErrorHookManager()

		var receivedErr1, receivedErr2 interfaces.DomainErrorInterface

		hook1 := func(ctx context.Context, err interfaces.DomainErrorInterface) error {
			receivedErr1 = err
			return nil
		}

		hook2 := func(ctx context.Context, err interfaces.DomainErrorInterface) error {
			receivedErr2 = err
			return nil
		}

		manager.Register(hook1)
		manager.Register(hook2)

		testErr := &mockDomainError{}
		err := manager.Execute(context.Background(), testErr)

		assert.NoError(t, err)
		assert.Equal(t, testErr, receivedErr1)
		assert.Equal(t, testErr, receivedErr2)
	})
}

func TestNewI18nHookManager_Success(t *testing.T) {
	mockClient := &mockI18nClient{}
	manager := NewI18nHookManager(mockClient)

	require.NotNil(t, manager)
	assert.IsType(t, &I18nHookManager{}, manager)

	// Test hook registration
	hook := func(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error {
		return nil
	}
	manager.Register(hook)
	assert.Equal(t, 1, manager.Count())
}

func TestHookManager(t *testing.T) {
	t.Parallel()

	t.Run("new manager starts empty", func(t *testing.T) {
		t.Parallel()

		manager := NewHookManager()
		start, stop, errorCount, i18n := manager.GetCounts()

		assert.Equal(t, 0, start)
		assert.Equal(t, 0, stop)
		assert.Equal(t, 0, errorCount)
		assert.Equal(t, 0, i18n)
	})

	t.Run("register hooks increases respective counts", func(t *testing.T) {
		t.Parallel()

		manager := NewHookManager()

		startHook := func(ctx context.Context) error { return nil }
		stopHook := func(ctx context.Context) error { return nil }
		errorHook := func(ctx context.Context, err interfaces.DomainErrorInterface) error { return nil }
		i18nHook := func(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error { return nil }

		manager.RegisterStartHook(startHook)
		manager.RegisterStopHook(stopHook)
		manager.RegisterErrorHook(errorHook)
		manager.RegisterI18nHook(i18nHook)

		start, stop, errorCount, i18n := manager.GetCounts()

		assert.Equal(t, 1, start)
		assert.Equal(t, 1, stop)
		assert.Equal(t, 1, errorCount)
		assert.Equal(t, 1, i18n)
	})

	t.Run("execute start hooks", func(t *testing.T) {
		t.Parallel()

		manager := NewHookManager()

		called := false
		hook := func(ctx context.Context) error {
			called = true
			return nil
		}

		manager.RegisterStartHook(hook)
		err := manager.ExecuteStartHooks(context.Background())

		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("execute stop hooks", func(t *testing.T) {
		t.Parallel()

		manager := NewHookManager()

		called := false
		hook := func(ctx context.Context) error {
			called = true
			return nil
		}

		manager.RegisterStopHook(hook)
		err := manager.ExecuteStopHooks(context.Background())

		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("execute error hooks", func(t *testing.T) {
		t.Parallel()

		manager := NewHookManager()

		var receivedErr interfaces.DomainErrorInterface
		hook := func(ctx context.Context, err interfaces.DomainErrorInterface) error {
			receivedErr = err
			return nil
		}

		manager.RegisterErrorHook(hook)

		testErr := &mockDomainError{}
		err := manager.ExecuteErrorHooks(context.Background(), testErr)

		assert.NoError(t, err)
		assert.Equal(t, testErr, receivedErr)
	})

	t.Run("execute i18n hooks", func(t *testing.T) {
		t.Parallel()

		manager := NewHookManager()

		var receivedErr interfaces.DomainErrorInterface
		var receivedLocale string

		hook := func(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error {
			receivedErr = err
			receivedLocale = locale
			return nil
		}

		manager.RegisterI18nHook(hook)

		testErr := &mockDomainError{}
		err := manager.ExecuteI18nHooks(context.Background(), testErr, "es")

		assert.NoError(t, err)
		assert.Equal(t, testErr, receivedErr)
		assert.Equal(t, "es", receivedLocale)
	})

	t.Run("clear removes all hooks", func(t *testing.T) {
		t.Parallel()

		manager := NewHookManager()

		startHook := func(ctx context.Context) error { return nil }
		stopHook := func(ctx context.Context) error { return nil }
		errorHook := func(ctx context.Context, err interfaces.DomainErrorInterface) error { return nil }
		i18nHook := func(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error { return nil }

		manager.RegisterStartHook(startHook)
		manager.RegisterStopHook(stopHook)
		manager.RegisterErrorHook(errorHook)
		manager.RegisterI18nHook(i18nHook)

		start, stop, errorCount, i18n := manager.GetCounts()
		assert.Equal(t, 1, start)
		assert.Equal(t, 1, stop)
		assert.Equal(t, 1, errorCount)
		assert.Equal(t, 1, i18n)

		manager.Clear()

		start, stop, errorCount, i18n = manager.GetCounts()
		assert.Equal(t, 0, start)
		assert.Equal(t, 0, stop)
		assert.Equal(t, 0, errorCount)
		assert.Equal(t, 0, i18n)
	})
}

func TestGlobalHooks(t *testing.T) {
	t.Parallel()

	// Note: These tests cannot be run in parallel as they affect global state
	// They are kept simple to minimize side effects

	t.Run("global start hooks", func(t *testing.T) {
		// Clear any existing global hooks first
		GlobalHookManager.Clear()

		called := false
		hook := func(ctx context.Context) error {
			called = true
			return nil
		}

		RegisterGlobalStartHook(hook)
		err := ExecuteGlobalStartHooks(context.Background())

		assert.NoError(t, err)
		assert.True(t, called)

		// Clean up
		GlobalHookManager.Clear()
	})

	t.Run("global hook counts", func(t *testing.T) {
		// Clear any existing global hooks first
		GlobalHookManager.Clear()

		startHook := func(ctx context.Context) error { return nil }
		RegisterGlobalStartHook(startHook)

		start, _, _, _ := GetGlobalHookCounts()
		assert.Equal(t, 1, start)

		// Clean up
		GlobalHookManager.Clear()
	})
}

// Mock domain error for testing
type mockDomainError struct{}

func (m *mockDomainError) Error() string                                                   { return "mock error" }
func (m *mockDomainError) Unwrap() error                                                   { return nil }
func (m *mockDomainError) Type() interfaces.ErrorType                                      { return interfaces.ValidationError }
func (m *mockDomainError) Metadata() map[string]interface{}                                { return nil }
func (m *mockDomainError) HTTPStatus() int                                                 { return 400 }
func (m *mockDomainError) StackTrace() string                                              { return "" }
func (m *mockDomainError) WithContext(ctx context.Context) interfaces.DomainErrorInterface { return m }
func (m *mockDomainError) Wrap(err error) interfaces.DomainErrorInterface                  { return m }
func (m *mockDomainError) WithMetadata(key string, value interface{}) interfaces.DomainErrorInterface {
	return m
}
func (m *mockDomainError) Code() string            { return "MOCK001" }
func (m *mockDomainError) Timestamp() time.Time    { return time.Now() }
func (m *mockDomainError) ToJSON() ([]byte, error) { return []byte(`{}`), nil }

// Mock i18n client for testing
type mockI18nClient struct{}

func (m *mockI18nClient) Translate(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
	return "translated message", nil
}

func (m *mockI18nClient) LoadTranslations(ctx context.Context) error {
	return nil
}

func (m *mockI18nClient) GetSupportedLanguages() []string {
	return []string{"en", "pt", "es"}
}

func (m *mockI18nClient) HasTranslation(key string, lang string) bool {
	return true
}

func (m *mockI18nClient) GetDefaultLanguage() string {
	return "en"
}

func (m *mockI18nClient) SetDefaultLanguage(lang string) {}

func (m *mockI18nClient) Start(ctx context.Context) error {
	return nil
}

func (m *mockI18nClient) Stop(ctx context.Context) error {
	return nil
}

func (m *mockI18nClient) Health(ctx context.Context) error {
	return nil
}

func (m *mockI18nClient) GetTranslationCount() int {
	return 100
}

func (m *mockI18nClient) GetTranslationCountByLanguage(lang string) int {
	return 50
}

func (m *mockI18nClient) GetLoadedLanguages() []string {
	return []string{"en", "pt"}
}

func BenchmarkStartHookExecution(b *testing.B) {
	manager := NewStartHookManager()

	hook := func(ctx context.Context) error { return nil }
	manager.Register(hook)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			manager.Execute(context.Background())
		}
	})
}

func BenchmarkErrorHookExecution(b *testing.B) {
	manager := NewErrorHookManager()

	hook := func(ctx context.Context, err interfaces.DomainErrorInterface) error { return nil }
	manager.Register(hook)

	testErr := &mockDomainError{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			manager.Execute(context.Background(), testErr)
		}
	})
}

func BenchmarkI18nHookExecution(b *testing.B) {
	mockI18nClient := &mockI18nClient{}
	manager := NewI18nHookManager(mockI18nClient)

	hook := func(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error { return nil }
	manager.Register(hook)

	testErr := &mockDomainError{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			manager.Execute(context.Background(), testErr, "en")
		}
	})
}
