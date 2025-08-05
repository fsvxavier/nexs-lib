package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/text/language"
)

func TestCachedProvider_ConcurrentAccess(t *testing.T) {
	mockProvider := &MockProvider{}
	cache := NewCachedProvider(mockProvider, 1*time.Second, 100)

	data := make(map[string]interface{})
	wg := sync.WaitGroup{}
	concurrentAccess := 100

	// Set up mock
	mockProvider.On("GetLanguages").Return([]language.Tag{language.English}).Maybe()
	mockProvider.On("SetLanguages", []string{"en"}).Return(nil).Maybe()
	mockProvider.On("Translate", "test_key", data).Return("test_value", nil).Maybe()

	// Test concurrent access
	for i := 0; i < concurrentAccess; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := cache.Translate("test_key", data)
			assert.NoError(t, err)
			assert.Equal(t, "test_value", result)
		}()
	}

	wg.Wait()
	mockProvider.AssertExpectations(t)
}

func TestCachedProvider_ConcurrentPluralAccess(t *testing.T) {
	mockProvider := &MockProvider{}
	cache := NewCachedProvider(mockProvider, 1*time.Second, 100)

	data := make(map[string]interface{})
	wg := sync.WaitGroup{}
	concurrentAccess := 100

	// Set up mock
	mockProvider.On("GetLanguages").Return([]language.Tag{language.English}).Maybe()
	mockProvider.On("SetLanguages", []string{"en"}).Return(nil).Maybe()
	mockProvider.On("TranslatePlural", "items", 2, data).Return("2 items", nil).Maybe()

	// Test concurrent access
	for i := 0; i < concurrentAccess; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := cache.TranslatePlural("items", 2, data)
			assert.NoError(t, err)
			assert.Equal(t, "2 items", result)
		}()
	}

	wg.Wait()
	mockProvider.AssertExpectations(t)
}

func TestCachedProvider_RaceConditions(t *testing.T) {
	mockProvider := &MockProvider{}
	cache := NewCachedProvider(mockProvider, 1*time.Millisecond, 100)

	data := make(map[string]interface{})
	wg := sync.WaitGroup{}

	mockProvider.On("Translate", mock.Anything, mock.Anything).Return("value", nil).Maybe()

	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			cache.Translate("key", data)
		}()
		go func() {
			defer wg.Done()
			cache.ClearCache()
		}()
	}

	wg.Wait()
	mockProvider.AssertExpectations(t)
}

func TestCachedProvider_TranslateError(t *testing.T) {
	mockProvider := &MockProvider{}
	cache := NewCachedProvider(mockProvider, 1*time.Second, 100)

	data := make(map[string]interface{})
	expectedError := fmt.Errorf("translation error")

	mockProvider.On("Translate", "error_key", data).Return("", expectedError).Once()

	result, err := cache.Translate("error_key", data)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result)

	mockProvider.AssertExpectations(t)
}

func BenchmarkCachedProvider_MixedOperations(b *testing.B) {
	mockProvider := &MockProvider{}
	cache := NewCachedProvider(mockProvider, 1*time.Hour, 1000)
	data := make(map[string]interface{})

	mockProvider.On("Translate", mock.Anything, mock.Anything).Return("value", nil).Maybe()
	mockProvider.On("TranslatePlural", mock.Anything, mock.Anything, mock.Anything).Return("values", nil).Maybe()

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			switch counter % 3 {
			case 0:
				cache.Translate("key", data)
			case 1:
				cache.TranslatePlural("key", 2, data)
			case 2:
				cache.ClearCache()
			}
			counter++
		}
	})
}
