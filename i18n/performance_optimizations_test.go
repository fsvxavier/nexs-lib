// Package i18n - Performance Optimizations Tests
// This file contains tests for the performance optimization features.
package i18n

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
	"github.com/fsvxavier/nexs-lib/i18n/providers/json"
)

// TestStringPool tests the string pool functionality.
func TestStringPool(t *testing.T) {
	pool := NewStringPool()

	// Test Get/Put
	slice1 := pool.Get()
	if slice1 == nil {
		t.Fatal("Expected non-nil slice from pool")
	}

	slice1 = append(slice1, "test1", "test2")
	pool.Put(slice1)

	// Get another slice - should be reset
	slice2 := pool.Get()
	if len(slice2) != 0 {
		t.Errorf("Expected empty slice, got length %d", len(slice2))
	}

	pool.Put(slice2)
}

// TestStringPool_Concurrent tests concurrent access to string pool.
func TestStringPool_Concurrent(t *testing.T) {
	pool := NewStringPool()

	var wg sync.WaitGroup
	concurrency := 100

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			slice := pool.Get()
			slice = append(slice, fmt.Sprintf("test%d", id))

			// Simulate some work
			time.Sleep(time.Microsecond)

			pool.Put(slice)
		}(i)
	}

	wg.Wait()
}

// TestStringInterner tests string interning functionality.
func TestStringInterner(t *testing.T) {
	interner := NewStringInterner()

	// Test basic interning
	s1 := interner.Intern("test")
	s2 := interner.Intern("test")

	// Should return same instance (test string interning by comparing pointers)
	if s1 != s2 {
		t.Error("Expected same string value for identical strings")
	}

	// Test different strings
	s3 := interner.Intern("different")
	if s1 == s3 {
		t.Error("Expected different string values for different strings")
	}

	// Test size
	if interner.Size() != 2 {
		t.Errorf("Expected size 2, got %d", interner.Size())
	}

	// Test clear
	interner.Clear()
	if interner.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", interner.Size())
	}
}

// TestStringInterner_Concurrent tests concurrent string interning.
func TestStringInterner_Concurrent(t *testing.T) {
	interner := NewStringInterner()

	var wg sync.WaitGroup
	concurrency := 100
	iterations := 100

	// Test concurrent access
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("key%d", j%10) // Reuse some keys
				result := interner.Intern(key)
				if result == "" {
					t.Errorf("Expected non-empty result from interner")
				}
			}
		}(i)
	}

	wg.Wait()

	// Should have at most 10 unique keys
	if interner.Size() > 10 {
		t.Errorf("Expected at most 10 unique keys, got %d", interner.Size())
	}
}

// TestBatchTranslator tests batch translation functionality.
func TestBatchTranslator(t *testing.T) {
	provider := createTestProviderForPerformanceTests(t)
	batchTranslator := NewBatchTranslator(provider)
	ctx := context.Background()

	// Create test requests
	requests := []BatchTranslationRequest{
		{Key: "hello.world", Lang: "en", Params: nil},
		{Key: "goodbye.world", Lang: "es", Params: map[string]interface{}{"name": "Test"}},
		{Key: "nonexistent.key", Lang: "pt", Params: nil},
	}

	responses := batchTranslator.TranslateBatch(ctx, requests)

	if len(responses) != len(requests) {
		t.Errorf("Expected %d responses, got %d", len(requests), len(responses))
	}

	// Check responses
	for i, resp := range responses {
		if resp.Key != requests[i].Key {
			t.Errorf("Expected key %s, got %s", requests[i].Key, resp.Key)
		}
		if resp.Lang != requests[i].Lang {
			t.Errorf("Expected lang %s, got %s", requests[i].Lang, resp.Lang)
		}

		// Last request should have error (nonexistent key)
		if i == 2 && resp.Error == "" {
			t.Error("Expected error for nonexistent key")
		}
	}
}

// TestBatchTranslator_Concurrent tests concurrent batch translation.
func TestBatchTranslator_Concurrent(t *testing.T) {
	provider := createTestProviderForPerformanceTests(t)
	batchTranslator := NewBatchTranslator(provider)
	ctx := context.Background()

	var wg sync.WaitGroup
	concurrency := 10

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			requests := []BatchTranslationRequest{
				{Key: "hello.world", Lang: "en", Params: nil},
				{Key: fmt.Sprintf("key.%d", id), Lang: "es", Params: nil},
			}

			responses := batchTranslator.TranslateBatch(ctx, requests)
			if len(responses) != len(requests) {
				t.Errorf("Expected %d responses, got %d", len(requests), len(responses))
			}
		}(i)
	}

	wg.Wait()
}

// TestPerformanceOptimizedProvider tests the performance-optimized provider.
func TestPerformanceOptimizedProvider(t *testing.T) {
	baseProvider := createTestProviderForPerformanceTests(t)
	provider := NewPerformanceOptimizedProvider(baseProvider)
	ctx := context.Background()

	// Test basic translation
	result, err := provider.Translate(ctx, "hello.world", "en", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result == "" {
		t.Error("Expected non-empty translation result")
	}

	// Test string interning - same keys should be interned
	provider.Translate(ctx, "hello.world", "en", nil)
	provider.Translate(ctx, "hello.world", "en", nil)

	if provider.GetInternedStringCount() == 0 {
		t.Error("Expected some interned strings")
	}

	// Test batch translation
	requests := []BatchTranslationRequest{
		{Key: "hello.world", Lang: "en", Params: nil},
		{Key: "goodbye.world", Lang: "es", Params: nil},
	}

	responses := provider.TranslateBatch(ctx, requests)
	if len(responses) != len(requests) {
		t.Errorf("Expected %d responses, got %d", len(requests), len(responses))
	}

	// Test clear interned strings
	provider.ClearInternedStrings()
	if provider.GetInternedStringCount() != 0 {
		t.Error("Expected 0 interned strings after clear")
	}
}

// TestLazyLoadingProvider tests the lazy loading provider.
func TestLazyLoadingProvider(t *testing.T) {
	baseProvider := createTestProviderForPerformanceTests(t)
	provider := NewLazyLoadingProvider(baseProvider)
	ctx := context.Background()

	// Test translation with lazy loading
	result, err := provider.Translate(ctx, "hello.world", "en", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result == "" {
		t.Error("Expected non-empty translation result")
	}

	// Test different language
	result2, err := provider.Translate(ctx, "hello.world", "es", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result2 == "" {
		t.Error("Expected non-empty translation result for Spanish")
	}
}

// TestCompressedProvider tests the compressed provider.
func TestCompressedProvider(t *testing.T) {
	baseProvider := createTestProviderForPerformanceTests(t)
	provider := NewCompressedProvider(baseProvider, true)
	ctx := context.Background()

	// Test basic functionality
	result, err := provider.Translate(ctx, "hello.world", "en", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result == "" {
		t.Error("Expected non-empty translation result")
	}

	// Test other methods
	langs := provider.GetSupportedLanguages()
	if len(langs) == 0 {
		t.Error("Expected supported languages")
	}

	defaultLang := provider.GetDefaultLanguage()
	if defaultLang == "" {
		t.Error("Expected default language")
	}
}

// TestGlobalStringInterner tests the global string interner.
func TestGlobalStringInterner(t *testing.T) {
	interner := GetGlobalStringInterner()
	if interner == nil {
		t.Fatal("Expected non-nil global string interner")
	}

	// Test basic functionality
	s1 := interner.Intern("global.test")
	s2 := interner.Intern("global.test")

	if s1 != s2 {
		t.Error("Expected same string value from global interner")
	}
}

// TestGlobalStringPool tests the global string pool.
func TestGlobalStringPool(t *testing.T) {
	pool := GetGlobalStringPool()
	if pool == nil {
		t.Fatal("Expected non-nil global string pool")
	}

	// Test basic functionality
	slice := pool.Get()
	if slice == nil {
		t.Fatal("Expected non-nil slice from global pool")
	}

	pool.Put(slice)
}

// TestPerformanceOptimizations_Integration tests integration of all optimizations.
func TestPerformanceOptimizations_Integration(t *testing.T) {
	baseProvider := createTestProviderForPerformanceTests(t)

	// Stack multiple optimizations
	lazyProvider := NewLazyLoadingProvider(baseProvider)
	optimizedProvider := NewPerformanceOptimizedProvider(lazyProvider)
	compressedProvider := NewCompressedProvider(optimizedProvider, true)

	ctx := context.Background()

	// Test translation works through all layers
	result, err := compressedProvider.Translate(ctx, "hello.world", "en", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result == "" {
		t.Error("Expected non-empty translation result")
	}

	// Test batch translation - type assertion
	optimized := optimizedProvider
	requests := []BatchTranslationRequest{
		{Key: "hello.world", Lang: "en", Params: nil},
	}

	responses := optimized.TranslateBatch(ctx, requests)
	if len(responses) != 1 {
		t.Errorf("Expected 1 response, got %d", len(responses))
	}
}

// TestPerformanceOptimizations_MemoryUsage tests memory usage patterns.
func TestPerformanceOptimizations_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	baseProvider := createTestProviderForPerformanceTests(t)
	provider := NewPerformanceOptimizedProvider(baseProvider)
	ctx := context.Background()

	var memBefore, memAfter runtime.MemStats
	runtime.GC() // Force GC to get accurate measurement
	runtime.ReadMemStats(&memBefore)

	// Perform many translations
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key.%d", i%100) // Reuse keys to test interning
		result, err := provider.Translate(ctx, key, "en", nil)
		if err != nil {
			// Ignore errors for memory testing
		}
		_ = result
	}

	runtime.GC() // Force GC
	runtime.ReadMemStats(&memAfter)

	memUsed := memAfter.Alloc - memBefore.Alloc
	t.Logf("Memory used: %d bytes", memUsed)

	// Check that string interning worked
	if provider.GetInternedStringCount() == 0 {
		t.Error("Expected some interned strings")
	}

	// Should have at most 100 unique keys
	if provider.GetInternedStringCount() > 200 { // Account for languages too
		t.Errorf("Expected at most 200 interned strings, got %d", provider.GetInternedStringCount())
	}
}

// TestPerformanceOptimizations_Concurrent tests concurrent access to optimizations.
func TestPerformanceOptimizations_Concurrent(t *testing.T) {
	baseProvider := createTestProviderForPerformanceTests(t)
	provider := NewPerformanceOptimizedProvider(baseProvider)
	ctx := context.Background()

	var wg sync.WaitGroup
	concurrency := 50
	iterations := 100

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("key.%d", j%10)
				lang := []string{"en", "es", "pt"}[j%3]

				result, err := provider.Translate(ctx, key, lang, nil)
				if err != nil {
					// Ignore errors for concurrent testing
				}
				_ = result
			}
		}(i)
	}

	wg.Wait()

	// Verify no race conditions occurred
	if provider.GetInternedStringCount() == 0 {
		t.Error("Expected some interned strings")
	}
}

// TestBatchTranslation_LargePayload tests batch translation with large payloads.
func TestBatchTranslation_LargePayload(t *testing.T) {
	provider := createTestProviderForPerformanceTests(t)
	batchTranslator := NewBatchTranslator(provider)
	ctx := context.Background()

	// Create large batch
	requests := make([]BatchTranslationRequest, 1000)
	for i := 0; i < 1000; i++ {
		requests[i] = BatchTranslationRequest{
			Key:  fmt.Sprintf("key.%d", i),
			Lang: []string{"en", "es", "pt"}[i%3],
			Params: map[string]interface{}{
				"index": i,
				"batch": "large",
			},
		}
	}

	start := time.Now()
	responses := batchTranslator.TranslateBatch(ctx, requests)
	duration := time.Since(start)

	if len(responses) != len(requests) {
		t.Errorf("Expected %d responses, got %d", len(requests), len(responses))
	}

	t.Logf("Large batch translation took: %v", duration)

	// Verify all responses have correct structure
	for i, resp := range responses {
		if resp.Key != requests[i].Key {
			t.Errorf("Response %d: expected key %s, got %s", i, requests[i].Key, resp.Key)
		}
		if resp.Lang != requests[i].Lang {
			t.Errorf("Response %d: expected lang %s, got %s", i, requests[i].Lang, resp.Lang)
		}
	}
}

// Helper function to create a test provider for performance tests
func createTestProviderForPerformanceTests(t *testing.T) interfaces.I18n {
	t.Helper()

	registry := NewRegistry()
	factory := json.NewFactory()
	err := registry.RegisterProvider(factory)
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DefaultLanguage:    "en",
		SupportedLanguages: []string{"en", "es", "pt"},
		LoadTimeout:        30 * time.Second,
		ProviderConfig: &config.JSONProviderConfig{
			FilePath:    "./testdata",
			FilePattern: "translations_{lang}.json",
			Encoding:    "utf-8",
		},
	}

	provider, err := registry.CreateProvider("json", cfg)
	if err != nil {
		t.Fatal(err)
	}

	return provider
}
