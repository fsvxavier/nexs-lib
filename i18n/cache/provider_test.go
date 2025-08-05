package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/text/language"
)

type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Translate(key string, data map[string]interface{}) (string, error) {
	args := m.Called(key, data)
	return args.String(0), args.Error(1)
}

func (m *MockProvider) TranslatePlural(key string, count interface{}, data map[string]interface{}) (string, error) {
	args := m.Called(key, count, data)
	return args.String(0), args.Error(1)
}

func (m *MockProvider) LoadTranslations(path string, format string) error {
	args := m.Called(path, format)
	return args.Error(0)
}

func (m *MockProvider) GetLanguages() []language.Tag {
	args := m.Called()
	return args.Get(0).([]language.Tag)
}

func (m *MockProvider) SetLanguages(languages ...string) error {
	args := m.Called(languages)
	return args.Error(0)
}

func TestCachedProvider_Translate(t *testing.T) {
	mockProvider := &MockProvider{}
	cache := NewCachedProvider(mockProvider, 1*time.Second, 100)

	// Set up mock
	mockProvider.On("GetLanguages").Return([]language.Tag{language.English}).Maybe()
	mockProvider.On("SetLanguages", []string{"en"}).Return(nil).Maybe()
	mockProvider.On("Translate", "test_key", map[string]interface{}{"name": "John"}).
		Return("Hello John", nil).Once()

	// First call - should hit provider
	result, err := cache.Translate("test_key", map[string]interface{}{"name": "John"})
	assert.NoError(t, err)
	assert.Equal(t, "Hello John", result)

	// Second call - should hit cache
	result, err = cache.Translate("test_key", map[string]interface{}{"name": "John"})
	assert.NoError(t, err)
	assert.Equal(t, "Hello John", result)

	// Wait for TTL
	time.Sleep(2 * time.Second)

	// Set up mock for third call
	mockProvider.On("Translate", "test_key", map[string]interface{}{"name": "John"}).
		Return("Hello John", nil).Once()

	// Third call - should hit provider again
	result, err = cache.Translate("test_key", map[string]interface{}{"name": "John"})
	assert.NoError(t, err)
	assert.Equal(t, "Hello John", result)

	mockProvider.AssertExpectations(t)
}

func TestCachedProvider_TranslatePlural(t *testing.T) {
	mockProvider := &MockProvider{}
	cache := NewCachedProvider(mockProvider, 1*time.Second, 100)

	count := 2

	// Set up mock
	mockProvider.On("GetLanguages").Return([]language.Tag{language.English}).Maybe()
	mockProvider.On("SetLanguages", []string{"en"}).Return(nil).Maybe()
	mockProvider.On("TranslatePlural", "items", count, map[string]interface{}{"count": 2}).
		Return("2 items", nil).Once()

	// First call - should hit provider
	result, err := cache.TranslatePlural("items", count, map[string]interface{}{"count": 2})
	assert.NoError(t, err)
	assert.Equal(t, "2 items", result)

	// Second call - should hit cache
	result, err = cache.TranslatePlural("items", count, map[string]interface{}{"count": 2})
	assert.NoError(t, err)
	assert.Equal(t, "2 items", result)

	// Wait for TTL
	time.Sleep(2 * time.Second)

	// Set up mock for third call
	mockProvider.On("TranslatePlural", "items", count, map[string]interface{}{"count": 2}).
		Return("2 items", nil).Once()

	// Third call - should hit provider again
	result, err = cache.TranslatePlural("items", count, map[string]interface{}{"count": 2})
	assert.NoError(t, err)
	assert.Equal(t, "2 items", result)

	mockProvider.AssertExpectations(t)
}

func TestCachedProvider_MaxEntries(t *testing.T) {
	mockProvider := &MockProvider{}
	cache := NewCachedProvider(mockProvider, 1*time.Hour, 2)

	data := make(map[string]interface{})

	// Set up mocks
	mockProvider.On("GetLanguages").Return([]language.Tag{language.English}).Maybe()
	mockProvider.On("SetLanguages", []string{"en"}).Return(nil).Maybe()
	mockProvider.On("Translate", "key1", data).Return("value1", nil).Once()
	mockProvider.On("Translate", "key2", data).Return("value2", nil).Once()
	mockProvider.On("Translate", "key3", data).Return("value3", nil).Once()

	// Fill cache
	_, _ = cache.Translate("key1", data)
	_, _ = cache.Translate("key2", data)

	// Add one more - should evict oldest
	_, _ = cache.Translate("key3", data)

	// Set up mock for evicted key
	mockProvider.On("Translate", "key1", data).Return("value1", nil).Once()

	// Try to get evicted key
	result, err := cache.Translate("key1", data)
	assert.NoError(t, err)
	assert.Equal(t, "value1", result)

	mockProvider.AssertExpectations(t)
}

func TestCachedProvider_ClearCache(t *testing.T) {
	mockProvider := &MockProvider{}
	cache := NewCachedProvider(mockProvider, 1*time.Hour, 100)

	data := make(map[string]interface{})

	// Set up mock
	mockProvider.On("GetLanguages").Return([]language.Tag{language.English}).Maybe()
	mockProvider.On("SetLanguages", []string{"en"}).Return(nil).Maybe()
	mockProvider.On("Translate", "test_key", data).Return("test_value", nil).Twice()

	// First call - should hit provider
	_, _ = cache.Translate("test_key", data)

	// Clear cache
	cache.ClearCache()

	// Second call - should hit provider again
	result, err := cache.Translate("test_key", data)
	assert.NoError(t, err)
	assert.Equal(t, "test_value", result)

	mockProvider.AssertExpectations(t)
}
