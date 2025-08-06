// Package i18n - Performance Optimization Benchmarks
// This file contains benchmarks specifically for the performance optimizations implemented.
package i18n

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"

	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
)

// Helper function to create a mock provider for benchmarking optimizations
func createMockProvider() interfaces.I18n {
	return &mockProvider{
		translations: map[string]map[string]string{
			"en": {
				"hello.world":   "Hello World",
				"goodbye.world": "Goodbye World",
			},
			"es": {
				"hello.world":   "Hola Mundo",
				"goodbye.world": "Adiós Mundo",
			},
			"pt": {
				"hello.world":   "Olá Mundo",
				"goodbye.world": "Tchau Mundo",
			},
		},
		defaultLang:    "en",
		supportedLangs: []string{"en", "es", "pt"},
	}
}

// mockProvider implements interfaces.I18n for benchmarking purposes
type mockProvider struct {
	translations   map[string]map[string]string
	defaultLang    string
	supportedLangs []string
}

func (m *mockProvider) Translate(_ context.Context, key, lang string, _ map[string]interface{}) (string, error) {
	if langTranslations, exists := m.translations[lang]; exists {
		if translation, exists := langTranslations[key]; exists {
			return translation, nil
		}
	}
	return "", fmt.Errorf("translation not found for key '%s' in language '%s'", key, lang)
}

func (m *mockProvider) LoadTranslations(_ context.Context) error { return nil }
func (m *mockProvider) GetSupportedLanguages() []string          { return m.supportedLangs }
func (m *mockProvider) HasTranslation(key, lang string) bool {
	if langTranslations, exists := m.translations[lang]; exists {
		_, exists := langTranslations[key]
		return exists
	}
	return false
}
func (m *mockProvider) GetDefaultLanguage() string     { return m.defaultLang }
func (m *mockProvider) SetDefaultLanguage(lang string) { m.defaultLang = lang }
func (m *mockProvider) Start(_ context.Context) error  { return nil }
func (m *mockProvider) Stop(_ context.Context) error   { return nil }
func (m *mockProvider) Health(_ context.Context) error { return nil }
func (m *mockProvider) GetTranslationCount() int {
	count := 0
	for _, langTranslations := range m.translations {
		count += len(langTranslations)
	}
	return count
}
func (m *mockProvider) GetTranslationCountByLanguage(lang string) int {
	if langTranslations, exists := m.translations[lang]; exists {
		return len(langTranslations)
	}
	return 0
}
func (m *mockProvider) GetLoadedLanguages() []string { return m.supportedLangs }

// BenchmarkStringInterner benchmarks string interning performance.
func BenchmarkStringInterner(b *testing.B) {
	interner := NewStringInterner()

	// Pre-populate with some strings
	for i := 0; i < 100; i++ {
		interner.Intern(fmt.Sprintf("key.%d", i))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key.%d", rand.Intn(200)) // 50% chance of existing key
		result := interner.Intern(key)
		_ = result
	}
}

// BenchmarkStringInterner_Concurrent benchmarks concurrent string interning.
func BenchmarkStringInterner_Concurrent(b *testing.B) {
	interner := NewStringInterner()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := fmt.Sprintf("key.%d", rand.Intn(100))
			result := interner.Intern(key)
			_ = result
		}
	})
}

// BenchmarkStringPool benchmarks string pool performance.
func BenchmarkStringPool(b *testing.B) {
	pool := NewStringPool()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		slice := pool.Get()
		slice = append(slice, "test1", "test2", "test3")
		pool.Put(slice)
	}
}

// BenchmarkStringPool_Concurrent benchmarks concurrent string pool usage.
func BenchmarkStringPool_Concurrent(b *testing.B) {
	pool := NewStringPool()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			slice := pool.Get()
			slice = append(slice, "test")
			pool.Put(slice)
		}
	})
}

// BenchmarkBatchTranslation benchmarks batch translation performance.
func BenchmarkBatchTranslation(b *testing.B) {
	provider := createMockProvider()
	batchTranslator := NewBatchTranslator(provider)
	ctx := context.Background()

	// Create batch requests
	requests := make([]BatchTranslationRequest, 100)
	for i := 0; i < 100; i++ {
		requests[i] = BatchTranslationRequest{
			Key:  "hello.world",
			Lang: []string{"en", "es", "pt"}[i%3],
			Params: map[string]interface{}{
				"index": i,
			},
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		responses := batchTranslator.TranslateBatch(ctx, requests)
		_ = responses
	}
}

// BenchmarkBatchTranslation_SmallBatch benchmarks small batch translation.
func BenchmarkBatchTranslation_SmallBatch(b *testing.B) {
	provider := createMockProvider()
	batchTranslator := NewBatchTranslator(provider)
	ctx := context.Background()

	requests := []BatchTranslationRequest{
		{Key: "hello.world", Lang: "en", Params: nil},
		{Key: "goodbye.world", Lang: "es", Params: nil},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		responses := batchTranslator.TranslateBatch(ctx, requests)
		_ = responses
	}
}

// BenchmarkBatchTranslation_LargeBatch benchmarks large batch translation.
func BenchmarkBatchTranslation_LargeBatch(b *testing.B) {
	provider := createMockProvider()
	batchTranslator := NewBatchTranslator(provider)
	ctx := context.Background()

	// Create large batch
	requests := make([]BatchTranslationRequest, 1000)
	for i := 0; i < 1000; i++ {
		requests[i] = BatchTranslationRequest{
			Key:  "hello.world",
			Lang: []string{"en", "es", "pt"}[i%3],
			Params: map[string]interface{}{
				"index": i,
			},
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		responses := batchTranslator.TranslateBatch(ctx, requests)
		_ = responses
	}
}

// BenchmarkPerformanceOptimizedProvider benchmarks the performance-optimized provider.
func BenchmarkPerformanceOptimizedProvider(b *testing.B) {
	baseProvider := createMockProvider()
	provider := NewPerformanceOptimizedProvider(baseProvider)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := provider.Translate(ctx, "hello.world", "en", nil)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

// BenchmarkPerformanceOptimizedProvider_Concurrent tests concurrent access.
func BenchmarkPerformanceOptimizedProvider_Concurrent(b *testing.B) {
	baseProvider := createMockProvider()
	provider := NewPerformanceOptimizedProvider(baseProvider)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result, err := provider.Translate(ctx, "hello.world", "en", nil)
			if err != nil {
				b.Error(err)
				return
			}
			_ = result
		}
	})
}

// BenchmarkLazyLoadingProvider benchmarks the lazy loading provider.
func BenchmarkLazyLoadingProvider(b *testing.B) {
	baseProvider := createMockProvider()
	provider := NewLazyLoadingProvider(baseProvider)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := provider.Translate(ctx, "hello.world", "en", nil)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

// BenchmarkCompressedProvider benchmarks the compressed provider.
func BenchmarkCompressedProvider(b *testing.B) {
	baseProvider := createMockProvider()
	provider := NewCompressedProvider(baseProvider, true)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := provider.Translate(ctx, "hello.world", "en", nil)
		if err != nil {
			b.Fatal(err)
		}
		_ = result
	}
}

// BenchmarkOptimizationsComparison compares different optimization approaches.
func BenchmarkOptimizationsComparison(b *testing.B) {
	baseProvider := createMockProvider()
	ctx := context.Background()

	b.Run("Base", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			result, err := baseProvider.Translate(ctx, "hello.world", "en", nil)
			if err != nil {
				b.Fatal(err)
			}
			_ = result
		}
	})

	b.Run("PerformanceOptimized", func(b *testing.B) {
		provider := NewPerformanceOptimizedProvider(baseProvider)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			result, err := provider.Translate(ctx, "hello.world", "en", nil)
			if err != nil {
				b.Fatal(err)
			}
			_ = result
		}
	})

	b.Run("LazyLoading", func(b *testing.B) {
		provider := NewLazyLoadingProvider(baseProvider)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			result, err := provider.Translate(ctx, "hello.world", "en", nil)
			if err != nil {
				b.Fatal(err)
			}
			_ = result
		}
	})

	b.Run("Compressed", func(b *testing.B) {
		provider := NewCompressedProvider(baseProvider, true)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			result, err := provider.Translate(ctx, "hello.world", "en", nil)
			if err != nil {
				b.Fatal(err)
			}
			_ = result
		}
	})

	b.Run("AllOptimizations", func(b *testing.B) {
		lazyProvider := NewLazyLoadingProvider(baseProvider)
		optimizedProvider := NewPerformanceOptimizedProvider(lazyProvider)
		compressedProvider := NewCompressedProvider(optimizedProvider, true)

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			result, err := compressedProvider.Translate(ctx, "hello.world", "en", nil)
			if err != nil {
				b.Fatal(err)
			}
			_ = result
		}
	})
}

// BenchmarkMemoryEfficiency benchmarks memory usage of optimizations.
func BenchmarkMemoryEfficiency(b *testing.B) {
	baseProvider := createMockProvider()
	provider := NewPerformanceOptimizedProvider(baseProvider)
	ctx := context.Background()

	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("dynamic.key.%d", i%100) // Reuse keys to test interning
		result, err := provider.Translate(ctx, key, "en", nil)
		if err != nil {
			// Ignore errors for memory benchmarking
		}
		_ = result
	}

	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	b.ReportMetric(float64(memAfter.Alloc-memBefore.Alloc)/float64(b.N), "bytes/op")
}

// BenchmarkHighConcurrency benchmarks high concurrency scenarios.
func BenchmarkHighConcurrency(b *testing.B) {
	baseProvider := createMockProvider()
	provider := NewPerformanceOptimizedProvider(baseProvider)
	ctx := context.Background()

	concurrency := runtime.NumCPU() * 4

	b.ResetTimer()
	b.ReportAllocs()

	var wg sync.WaitGroup
	requests := b.N
	requestsPerWorker := requests / concurrency
	if requestsPerWorker == 0 {
		requestsPerWorker = 1
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				result, err := provider.Translate(ctx, "hello.world", "en", nil)
				if err != nil {
					b.Error(err)
					return
				}
				_ = result
			}
		}()
	}

	wg.Wait()
}

// BenchmarkProfileCPU is designed for CPU profiling.
func BenchmarkProfileCPU(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping CPU profiling benchmark in short mode")
	}

	baseProvider := createMockProvider()
	provider := NewPerformanceOptimizedProvider(baseProvider)
	ctx := context.Background()

	// Create varied workload
	keys := make([]string, 1000)
	languages := []string{"en", "es", "pt"}

	for i := 0; i < 1000; i++ {
		keys[i] = fmt.Sprintf("benchmark.key.%d", i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := keys[rand.Intn(len(keys))]
		lang := languages[rand.Intn(len(languages))]

		result, err := provider.Translate(ctx, key, lang, nil)
		if err != nil {
			// Ignore errors for CPU profiling
		}
		_ = result
	}
}

// BenchmarkProfileMemory is designed for memory profiling.
func BenchmarkProfileMemory(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping memory profiling benchmark in short mode")
	}

	baseProvider := createMockProvider()
	provider := NewPerformanceOptimizedProvider(baseProvider)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Mix of operations to simulate real usage
		operations := []func(){
			func() {
				provider.Translate(ctx, "hello.world", "en", nil)
			},
			func() {
				provider.Translate(ctx, "goodbye.world", "es", map[string]interface{}{"name": "Test"})
			},
			func() {
				provider.HasTranslation("hello.world", "pt")
			},
			func() {
				provider.GetSupportedLanguages()
			},
		}

		// Execute random operation
		operations[rand.Intn(len(operations))]()
	}
}

// init function to ensure deterministic benchmarks
func init() {
	rand.Seed(42) // Fixed seed for reproducible benchmarks
}
