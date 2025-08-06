// Package main demonstrates middleware usage with i18n library.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/i18n"
	"github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
	"github.com/fsvxavier/nexs-lib/i18n/providers/json"
)

// CustomCachingMiddleware demonstrates a custom middleware implementation
type CustomCachingMiddleware struct {
	name  string
	cache *MemoryCache
	stats CacheStats
	mu    sync.RWMutex
}

type CacheStats struct {
	Hits    int64
	Misses  int64
	Sets    int64
	Deletes int64
	Errors  int64
}

type MemoryCache struct {
	data map[string]cacheItem
	mu   sync.RWMutex
}

type cacheItem struct {
	value     string
	expiresAt time.Time
}

// CustomRateLimitingMiddleware demonstrates rate limiting
type CustomRateLimitingMiddleware struct {
	name        string
	requests    map[string][]time.Time
	maxRequests int
	window      time.Duration
	mu          sync.RWMutex
}

func main() {
	// Create temporary directory for translation files
	tempDir, err := os.MkdirTemp("", "i18n_middleware_example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create translation files
	if err := createTranslationFiles(tempDir); err != nil {
		log.Fatal("Failed to create translation files:", err)
	}

	fmt.Println("=== I18n Middleware Demonstration ===")
	fmt.Println()

	// Demo 1: Basic provider without middleware
	fmt.Println("🔸 Demo 1: Basic Provider (No Middleware)")
	basicProvider, err := setupBasicProvider(tempDir)
	if err != nil {
		log.Fatal("Failed to setup basic provider:", err)
	}
	defer basicProvider.Stop(context.Background())

	demoBasicTranslations(basicProvider)
	fmt.Println()

	// Demo 2: Provider with caching middleware
	fmt.Println("🔸 Demo 2: Provider with Caching Middleware")
	cachedProvider, cacheMiddleware, err := setupCachedProvider(tempDir)
	if err != nil {
		log.Fatal("Failed to setup cached provider:", err)
	}
	defer cachedProvider.Stop(context.Background())

	demoCachedTranslations(cachedProvider, cacheMiddleware)
	fmt.Println()

	// Demo 3: Provider with rate limiting middleware
	fmt.Println("🔸 Demo 3: Provider with Rate Limiting Middleware")
	rateLimitedProvider, rateLimitMiddleware, err := setupRateLimitedProvider(tempDir)
	if err != nil {
		log.Fatal("Failed to setup rate limited provider:", err)
	}
	defer rateLimitedProvider.Stop(context.Background())

	demoRateLimitedTranslations(rateLimitedProvider, rateLimitMiddleware)
	fmt.Println()

	// Demo 4: Provider with multiple middlewares
	fmt.Println("🔸 Demo 4: Provider with Multiple Middlewares")
	multiMiddlewareProvider, cache, rateLimit, err := setupMultiMiddlewareProvider(tempDir)
	if err != nil {
		log.Fatal("Failed to setup multi-middleware provider:", err)
	}
	defer multiMiddlewareProvider.Stop(context.Background())

	demoMultiMiddlewareTranslations(multiMiddlewareProvider, cache, rateLimit)
	fmt.Println()

	fmt.Println("✅ All middleware demonstrations completed successfully!")
}

func setupBasicProvider(translationDir string) (interfaces.I18n, error) {
	// Configure i18n
	cfg, err := config.NewConfigBuilder().
		WithSupportedLanguages("en", "pt", "es").
		WithDefaultLanguage("en").
		WithFallbackToDefault(true).
		WithStrictMode(false).
		WithCache(false, 0). // Disable built-in cache
		WithLoadTimeout(10 * time.Second).
		WithProviderConfig(&config.JSONProviderConfig{
			FilePath:     translationDir,
			FilePattern:  "{lang}.json",
			Encoding:     "utf-8",
			NestedKeys:   true,
			ValidateJSON: true,
		}).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create configuration: %w", err)
	}

	// Create registry and register provider
	registry := i18n.NewRegistry()
	jsonFactory := &json.Factory{}
	if err := registry.RegisterProvider(jsonFactory); err != nil {
		return nil, fmt.Errorf("failed to register provider: %w", err)
	}

	// Create provider
	provider, err := registry.CreateProvider("json", cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Start provider
	ctx := context.Background()
	if err := provider.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start provider: %w", err)
	}

	return provider, nil
}

func setupCachedProvider(translationDir string) (interfaces.I18n, *CustomCachingMiddleware, error) {
	// Create basic provider
	provider, err := setupBasicProvider(translationDir)
	if err != nil {
		return nil, nil, err
	}

	// Create caching middleware
	cache := &MemoryCache{data: make(map[string]cacheItem)}
	middleware := &CustomCachingMiddleware{
		name:  "custom-cache",
		cache: cache,
		stats: CacheStats{},
	}

	// Wrap the provider with middleware
	wrappedProvider := &MiddlewareWrappedProvider{
		provider:    provider,
		middlewares: []interfaces.Middleware{middleware},
	}

	return wrappedProvider, middleware, nil
}

func setupRateLimitedProvider(translationDir string) (interfaces.I18n, *CustomRateLimitingMiddleware, error) {
	// Create basic provider
	provider, err := setupBasicProvider(translationDir)
	if err != nil {
		return nil, nil, err
	}

	// Create rate limiting middleware
	middleware := &CustomRateLimitingMiddleware{
		name:        "rate-limiter",
		requests:    make(map[string][]time.Time),
		maxRequests: 5, // 5 requests per window
		window:      time.Second,
	}

	// Wrap the provider with middleware
	wrappedProvider := &MiddlewareWrappedProvider{
		provider:    provider,
		middlewares: []interfaces.Middleware{middleware},
	}

	return wrappedProvider, middleware, nil
}

func setupMultiMiddlewareProvider(translationDir string) (interfaces.I18n, *CustomCachingMiddleware, *CustomRateLimitingMiddleware, error) {
	// Create basic provider
	provider, err := setupBasicProvider(translationDir)
	if err != nil {
		return nil, nil, nil, err
	}

	// Create middlewares
	cache := &MemoryCache{data: make(map[string]cacheItem)}
	cacheMiddleware := &CustomCachingMiddleware{
		name:  "multi-cache",
		cache: cache,
		stats: CacheStats{},
	}

	rateLimitMiddleware := &CustomRateLimitingMiddleware{
		name:        "multi-rate-limiter",
		requests:    make(map[string][]time.Time),
		maxRequests: 10, // More permissive for demo
		window:      time.Second,
	}

	// Wrap the provider with multiple middlewares
	wrappedProvider := &MiddlewareWrappedProvider{
		provider:    provider,
		middlewares: []interfaces.Middleware{rateLimitMiddleware, cacheMiddleware}, // Order matters!
	}

	return wrappedProvider, cacheMiddleware, rateLimitMiddleware, nil
}

// Demonstration functions
func demoBasicTranslations(provider interfaces.I18n) {
	ctx := context.Background()
	keys := []string{"hello", "greeting", "user.profile.title"}
	languages := []string{"en", "pt", "es"}

	start := time.Now()
	for _, key := range keys {
		for _, lang := range languages {
			result, err := provider.Translate(ctx, key, lang, map[string]interface{}{
				"name": "João",
				"age":  25,
			})
			if err != nil {
				fmt.Printf("  ❌ %s [%s]: %v\n", key, lang, err)
			} else {
				fmt.Printf("  ✅ %s [%s]: %s\n", key, lang, result)
			}
		}
	}
	duration := time.Since(start)
	fmt.Printf("  ⏱️  Total time: %v\n", duration)
}

func demoCachedTranslations(provider interfaces.I18n, middleware *CustomCachingMiddleware) {
	ctx := context.Background()
	key := "greeting"
	params := map[string]interface{}{"name": "Maria", "age": 30}

	fmt.Println("  First request (cache miss):")
	start := time.Now()
	result1, _ := provider.Translate(ctx, key, "en", params)
	duration1 := time.Since(start)
	fmt.Printf("  ✅ %s [en]: %s (took %v)\n", key, result1, duration1)

	fmt.Println("  Second request (cache hit):")
	start = time.Now()
	result2, _ := provider.Translate(ctx, key, "en", params)
	duration2 := time.Since(start)
	fmt.Printf("  ✅ %s [en]: %s (took %v)\n", key, result2, duration2)

	fmt.Println("  Cache statistics:")
	stats := middleware.GetStats()
	fmt.Printf("    Hits: %d, Misses: %d, Sets: %d\n", stats.Hits, stats.Misses, stats.Sets)
	fmt.Printf("    Cache size: %d entries\n", middleware.cache.Size())
}

func demoRateLimitedTranslations(provider interfaces.I18n, middleware *CustomRateLimitingMiddleware) {
	ctx := context.Background()
	key := "hello"

	fmt.Println("  Making requests rapidly (rate limit: 5/second):")
	for i := 1; i <= 8; i++ {
		result, err := provider.Translate(ctx, key, "en", nil)
		if err != nil {
			fmt.Printf("  %d. ❌ Rate limited: %v\n", i, err)
		} else {
			fmt.Printf("  %d. ✅ %s\n", i, result)
		}
		time.Sleep(100 * time.Millisecond) // Small delay
	}

	fmt.Println("  Waiting for rate limit window to reset...")
	time.Sleep(time.Second)

	fmt.Println("  Making request after reset:")
	result, err := provider.Translate(ctx, key, "en", nil)
	if err != nil {
		fmt.Printf("  ❌ %v\n", err)
	} else {
		fmt.Printf("  ✅ %s\n", result)
	}
}

func demoMultiMiddlewareTranslations(provider interfaces.I18n, cache *CustomCachingMiddleware, rateLimit *CustomRateLimitingMiddleware) {
	ctx := context.Background()
	key := "user.profile.title"

	fmt.Println("  Demonstrating middleware chain (Rate Limit → Cache → Provider):")

	// Make several requests to show caching and rate limiting working together
	for i := 1; i <= 3; i++ {
		fmt.Printf("  Request %d:\n", i)
		start := time.Now()
		result, err := provider.Translate(ctx, key, "pt", nil)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("    ❌ Error: %v\n", err)
		} else {
			fmt.Printf("    ✅ Result: %s (took %v)\n", result, duration)
		}

		// Show middleware stats
		cacheStats := cache.GetStats()
		fmt.Printf("    📊 Cache - Hits: %d, Misses: %d\n", cacheStats.Hits, cacheStats.Misses)

		time.Sleep(200 * time.Millisecond)
	}
}

// Middleware implementations

// CustomCachingMiddleware implementation
func (m *CustomCachingMiddleware) Name() string {
	return m.name
}

func (m *CustomCachingMiddleware) WrapTranslate(next interfaces.TranslateFunc) interfaces.TranslateFunc {
	return func(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
		// Create cache key
		cacheKey := fmt.Sprintf("%s:%s:%v", key, lang, params)

		// Try to get from cache
		if value, found := m.cache.Get(cacheKey); found {
			m.mu.Lock()
			m.stats.Hits++
			m.mu.Unlock()
			return value, nil
		}

		// Cache miss - call next middleware/provider
		m.mu.Lock()
		m.stats.Misses++
		m.mu.Unlock()

		result, err := next(ctx, key, lang, params)
		if err != nil {
			m.mu.Lock()
			m.stats.Errors++
			m.mu.Unlock()
			return "", err
		}

		// Store in cache
		if err := m.cache.Set(cacheKey, result, 5*time.Minute); err == nil {
			m.mu.Lock()
			m.stats.Sets++
			m.mu.Unlock()
		}

		return result, nil
	}
}

func (m *CustomCachingMiddleware) GetStats() CacheStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// Observer methods (required by interfaces.Middleware)
func (m *CustomCachingMiddleware) OnStart(ctx context.Context, providerName string) error {
	fmt.Printf("  🔧 Cache middleware '%s' started for provider '%s'\n", m.name, providerName)
	return nil
}

func (m *CustomCachingMiddleware) OnStop(ctx context.Context, providerName string) error {
	fmt.Printf("  🔧 Cache middleware '%s' stopped for provider '%s'\n", m.name, providerName)
	return nil
}

func (m *CustomCachingMiddleware) OnError(ctx context.Context, providerName string, err error) error {
	fmt.Printf("  🔧 Cache middleware '%s' detected error in provider '%s': %v\n", m.name, providerName, err)
	return nil
}

func (m *CustomCachingMiddleware) OnTranslate(ctx context.Context, providerName string, key string, lang string, result string) error {
	// Optional logging
	return nil
}

// CustomRateLimitingMiddleware implementation
func (m *CustomRateLimitingMiddleware) Name() string {
	return m.name
}

func (m *CustomRateLimitingMiddleware) WrapTranslate(next interfaces.TranslateFunc) interfaces.TranslateFunc {
	return func(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
		clientID := "default" // In real app, extract from context/request

		m.mu.Lock()
		now := time.Now()

		// Clean old requests outside the window
		requests := m.requests[clientID]
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) < m.window {
				validRequests = append(validRequests, reqTime)
			}
		}

		// Check if limit exceeded
		if len(validRequests) >= m.maxRequests {
			m.mu.Unlock()
			return "", fmt.Errorf("rate limit exceeded: %d requests per %v", m.maxRequests, m.window)
		}

		// Add current request
		validRequests = append(validRequests, now)
		m.requests[clientID] = validRequests
		m.mu.Unlock()

		// Call next middleware/provider
		return next(ctx, key, lang, params)
	}
}

// Observer methods (required by interfaces.Middleware)
func (m *CustomRateLimitingMiddleware) OnStart(ctx context.Context, providerName string) error {
	fmt.Printf("  🔧 Rate limiting middleware '%s' started for provider '%s'\n", m.name, providerName)
	return nil
}

func (m *CustomRateLimitingMiddleware) OnStop(ctx context.Context, providerName string) error {
	fmt.Printf("  🔧 Rate limiting middleware '%s' stopped for provider '%s'\n", m.name, providerName)
	return nil
}

func (m *CustomRateLimitingMiddleware) OnError(ctx context.Context, providerName string, err error) error {
	return nil
}

func (m *CustomRateLimitingMiddleware) OnTranslate(ctx context.Context, providerName string, key string, lang string, result string) error {
	return nil
}

// MemoryCache implementation
func (m *MemoryCache) Get(key string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.data[key]
	if !exists || time.Now().After(item.expiresAt) {
		return "", false
	}
	return item.value, true
}

func (m *MemoryCache) Set(key string, value string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (m *MemoryCache) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}

// MiddlewareWrappedProvider wraps a provider with middlewares
type MiddlewareWrappedProvider struct {
	provider    interfaces.I18n
	middlewares []interfaces.Middleware
}

func (p *MiddlewareWrappedProvider) Translate(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
	// Create the translation function chain
	translateFunc := p.provider.Translate

	// Apply middlewares in reverse order (so they execute in correct order)
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		translateFunc = p.middlewares[i].WrapTranslate(translateFunc)
	}

	return translateFunc(ctx, key, lang, params)
}

// Delegate all other methods to the underlying provider
func (p *MiddlewareWrappedProvider) LoadTranslations(ctx context.Context) error {
	return p.provider.LoadTranslations(ctx)
}

func (p *MiddlewareWrappedProvider) GetSupportedLanguages() []string {
	return p.provider.GetSupportedLanguages()
}

func (p *MiddlewareWrappedProvider) HasTranslation(key string, lang string) bool {
	return p.provider.HasTranslation(key, lang)
}

func (p *MiddlewareWrappedProvider) GetDefaultLanguage() string {
	return p.provider.GetDefaultLanguage()
}

func (p *MiddlewareWrappedProvider) SetDefaultLanguage(lang string) {
	p.provider.SetDefaultLanguage(lang)
}

func (p *MiddlewareWrappedProvider) Start(ctx context.Context) error {
	// Start middlewares
	for _, middleware := range p.middlewares {
		if err := middleware.OnStart(ctx, "middleware-wrapped"); err != nil {
			return err
		}
	}
	return p.provider.Start(ctx)
}

func (p *MiddlewareWrappedProvider) Stop(ctx context.Context) error {
	// Stop middlewares
	for _, middleware := range p.middlewares {
		middleware.OnStop(ctx, "middleware-wrapped")
	}
	return p.provider.Stop(ctx)
}

func (p *MiddlewareWrappedProvider) Health(ctx context.Context) error {
	return p.provider.Health(ctx)
}

func (p *MiddlewareWrappedProvider) GetTranslationCount() int {
	return p.provider.GetTranslationCount()
}

func (p *MiddlewareWrappedProvider) GetTranslationCountByLanguage(lang string) int {
	return p.provider.GetTranslationCountByLanguage(lang)
}

func (p *MiddlewareWrappedProvider) GetLoadedLanguages() []string {
	return p.provider.GetLoadedLanguages()
}

func createTranslationFiles(dir string) error {
	// English translations
	enContent := `{
  "hello": "Hello!",
  "goodbye": "Goodbye!",
  "greeting": "Hello {{name}}, you are {{age}} years old!",
  "user": {
    "profile": {
      "title": "User Profile",
      "edit": "Edit Profile"
    },
    "settings": {
      "title": "Settings",
      "language": "Language"
    }
  },
  "middleware": {
    "demo": "This is a middleware demonstration"
  }
}`

	// Portuguese translations
	ptContent := `{
  "hello": "Olá!",
  "goodbye": "Tchau!",
  "greeting": "Olá {{name}}, você tem {{age}} anos!",
  "user": {
    "profile": {
      "title": "Perfil do Usuário",
      "edit": "Editar Perfil"
    },
    "settings": {
      "title": "Configurações",
      "language": "Idioma"
    }
  },
  "middleware": {
    "demo": "Esta é uma demonstração de middleware"
  }
}`

	// Spanish translations
	esContent := `{
  "hello": "¡Hola!",
  "goodbye": "¡Adiós!",
  "greeting": "¡Hola {{name}}, tienes {{age}} años!",
  "user": {
    "profile": {
      "title": "Perfil de Usuario",
      "edit": "Editar Perfil"
    },
    "settings": {
      "title": "Configuración",
      "language": "Idioma"
    }
  },
  "middleware": {
    "demo": "Esta es una demostración de middleware"
  }
}`

	files := map[string]string{
		"en.json": enContent,
		"pt.json": ptContent,
		"es.json": esContent,
	}

	for filename, content := range files {
		filePath := filepath.Join(dir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
	}

	return nil
}
