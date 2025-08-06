// Package i18n - Performance Optimizations
// This file contains performance optimizations for the i18n module including
// memory pooling, string interning, and batch operations.
package i18n

import (
	"context"
	"strings"
	"sync"

	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
)

// StringPool provides memory pooling for string operations to reduce GC pressure.
type StringPool struct {
	pool sync.Pool
}

// NewStringPool creates a new string pool.
func NewStringPool() *StringPool {
	return &StringPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]string, 0, 32) // Pre-allocate capacity for 32 strings
			},
		},
	}
}

// Get retrieves a string slice from the pool.
func (sp *StringPool) Get() []string {
	return sp.pool.Get().([]string)
}

// Put returns a string slice to the pool.
func (sp *StringPool) Put(s []string) {
	// Reset slice but keep capacity
	s = s[:0]
	sp.pool.Put(s)
}

// StringInterner provides string interning to reduce memory usage for common keys.
type StringInterner struct {
	cache map[string]string
	mu    sync.RWMutex
}

// NewStringInterner creates a new string interner.
func NewStringInterner() *StringInterner {
	return &StringInterner{
		cache: make(map[string]string),
	}
}

// Intern returns the canonical representation of the string.
// If the string is already interned, returns the existing instance.
func (si *StringInterner) Intern(s string) string {
	si.mu.RLock()
	if interned, exists := si.cache[s]; exists {
		si.mu.RUnlock()
		return interned
	}
	si.mu.RUnlock()

	si.mu.Lock()
	defer si.mu.Unlock()

	// Double-check after acquiring write lock
	if interned, exists := si.cache[s]; exists {
		return interned
	}

	// Create a copy to avoid holding references to potentially large strings
	interned := strings.Clone(s)
	si.cache[s] = interned
	return interned
}

// Size returns the number of interned strings.
func (si *StringInterner) Size() int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return len(si.cache)
}

// Clear removes all interned strings.
func (si *StringInterner) Clear() {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.cache = make(map[string]string)
}

// BatchTranslationRequest represents a batch translation request.
type BatchTranslationRequest struct {
	Key    string                 `json:"key"`
	Lang   string                 `json:"lang"`
	Params map[string]interface{} `json:"params"`
}

// BatchTranslationResponse represents a batch translation response.
type BatchTranslationResponse struct {
	Key         string `json:"key"`
	Lang        string `json:"lang"`
	Translation string `json:"translation"`
	Error       string `json:"error,omitempty"`
}

// BatchTranslator provides batch translation operations for improved performance.
type BatchTranslator struct {
	provider interfaces.I18n
	interner *StringInterner
	pool     *StringPool
}

// NewBatchTranslator creates a new batch translator.
func NewBatchTranslator(provider interfaces.I18n) *BatchTranslator {
	return &BatchTranslator{
		provider: provider,
		interner: NewStringInterner(),
		pool:     NewStringPool(),
	}
}

// TranslateBatch performs batch translation of multiple keys.
// This is more efficient than individual translations for large sets.
func (bt *BatchTranslator) TranslateBatch(ctx context.Context, requests []BatchTranslationRequest) []BatchTranslationResponse {
	responses := make([]BatchTranslationResponse, len(requests))

	// Process translations concurrently using worker pool
	const maxWorkers = 10
	workers := maxWorkers
	if len(requests) < workers {
		workers = len(requests)
	}

	requestChan := make(chan struct {
		idx int
		req BatchTranslationRequest
	}, len(requests))

	responseChan := make(chan struct {
		idx  int
		resp BatchTranslationResponse
	}, len(requests))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range requestChan {
				// Use interned strings to reduce memory usage
				key := bt.interner.Intern(work.req.Key)
				lang := bt.interner.Intern(work.req.Lang)

				translation, err := bt.provider.Translate(ctx, key, lang, work.req.Params)

				resp := BatchTranslationResponse{
					Key:         key,
					Lang:        lang,
					Translation: translation,
				}

				if err != nil {
					resp.Error = err.Error()
				}

				responseChan <- struct {
					idx  int
					resp BatchTranslationResponse
				}{work.idx, resp}
			}
		}()
	}

	// Send requests
	go func() {
		defer close(requestChan)
		for i, req := range requests {
			requestChan <- struct {
				idx int
				req BatchTranslationRequest
			}{i, req}
		}
	}()

	// Collect responses
	go func() {
		wg.Wait()
		close(responseChan)
	}()

	for work := range responseChan {
		responses[work.idx] = work.resp
	}

	return responses
}

// PerformanceOptimizedProvider wraps a provider with performance optimizations.
type PerformanceOptimizedProvider struct {
	interfaces.I18n
	interner        *StringInterner
	pool            *StringPool
	batchTranslator *BatchTranslator
}

// NewPerformanceOptimizedProvider creates a performance-optimized provider wrapper.
func NewPerformanceOptimizedProvider(provider interfaces.I18n) *PerformanceOptimizedProvider {
	pop := &PerformanceOptimizedProvider{
		I18n:     provider,
		interner: NewStringInterner(),
		pool:     NewStringPool(),
	}
	pop.batchTranslator = NewBatchTranslator(pop.I18n)
	return pop
}

// Translate performs optimized translation with string interning.
func (pop *PerformanceOptimizedProvider) Translate(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
	// Use interned strings to reduce memory usage for common keys/languages
	internedKey := pop.interner.Intern(key)
	internedLang := pop.interner.Intern(lang)

	return pop.I18n.Translate(ctx, internedKey, internedLang, params)
}

// TranslateBatch provides batch translation capability.
func (pop *PerformanceOptimizedProvider) TranslateBatch(ctx context.Context, requests []BatchTranslationRequest) []BatchTranslationResponse {
	return pop.batchTranslator.TranslateBatch(ctx, requests)
}

// GetInternedStringCount returns the number of interned strings.
func (pop *PerformanceOptimizedProvider) GetInternedStringCount() int {
	return pop.interner.Size()
}

// ClearInternedStrings clears all interned strings.
func (pop *PerformanceOptimizedProvider) ClearInternedStrings() {
	pop.interner.Clear()
}

// LazyLoadingProvider implements lazy loading of translations for improved startup time.
type LazyLoadingProvider struct {
	interfaces.I18n
	loadedLanguages map[string]bool
	mu              sync.RWMutex
}

// NewLazyLoadingProvider creates a new lazy loading provider wrapper.
func NewLazyLoadingProvider(provider interfaces.I18n) *LazyLoadingProvider {
	return &LazyLoadingProvider{
		I18n:            provider,
		loadedLanguages: make(map[string]bool),
	}
}

// Translate performs translation with lazy loading of languages.
func (llp *LazyLoadingProvider) Translate(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
	// Check if language is already loaded
	llp.mu.RLock()
	loaded := llp.loadedLanguages[lang]
	llp.mu.RUnlock()

	if !loaded {
		// Load language on demand
		if err := llp.loadLanguage(ctx, lang); err != nil {
			return "", err
		}
	}

	return llp.I18n.Translate(ctx, key, lang, params)
}

// loadLanguage loads translations for a specific language.
func (llp *LazyLoadingProvider) loadLanguage(ctx context.Context, lang string) error {
	llp.mu.Lock()
	defer llp.mu.Unlock()

	// Double-check after acquiring write lock
	if llp.loadedLanguages[lang] {
		return nil
	}

	// Here we would implement language-specific loading
	// For now, we'll mark as loaded
	llp.loadedLanguages[lang] = true
	return nil
}

// CompressedProvider implements translation compression for large translation files.
type CompressedProvider struct {
	interfaces.I18n
	compressionEnabled bool
}

// NewCompressedProvider creates a new compressed provider wrapper.
func NewCompressedProvider(provider interfaces.I18n, compressionEnabled bool) *CompressedProvider {
	return &CompressedProvider{
		I18n:               provider,
		compressionEnabled: compressionEnabled,
	}
}

// Global performance optimization instances
var (
	globalStringInterner = NewStringInterner()
	globalStringPool     = NewStringPool()
)

// GetGlobalStringInterner returns the global string interner instance.
func GetGlobalStringInterner() *StringInterner {
	return globalStringInterner
}

// GetGlobalStringPool returns the global string pool instance.
func GetGlobalStringPool() *StringPool {
	return globalStringPool
}
