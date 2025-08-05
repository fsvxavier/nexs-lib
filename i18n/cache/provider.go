package cache

import (
	"time"

	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
	"golang.org/x/text/language"
)

type cacheEntry struct {
	value      string
	expiresAt  time.Time
	isPlural   bool
	pluralData map[string]string
}

type CachedProvider struct {
	provider   interfaces.Provider
	cache      *LRUCache
	ttl        time.Duration
	maxEntries int
	metrics    *MetricsCollector
}

// NewCachedProvider creates a new cached provider decorator
func NewCachedProvider(provider interfaces.Provider, ttl time.Duration, maxEntries int) *CachedProvider {
	return &CachedProvider{
		provider:   provider,
		cache:      NewLRUCache(maxEntries),
		ttl:        ttl,
		maxEntries: maxEntries,
		metrics:    NewMetricsCollector(),
	}
}

func (c *CachedProvider) getCacheKey(key string, data map[string]interface{}) string {
	if data == nil {
		return key
	}
	return key + ":" + hashMap(data)
}

func (c *CachedProvider) Translate(key string, data map[string]interface{}) (string, error) {
	cacheKey := c.getCacheKey(key, data)

	// Try to get from cache
	if entry, ok := c.cache.Get(cacheKey); ok {
		e := entry.(cacheEntry)
		if !e.isPlural && time.Now().Before(e.expiresAt) {
			c.metrics.RecordHit()
			return e.value, nil
		}
		c.cache.Remove(cacheKey)
	}

	// Get from provider
	c.metrics.RecordMiss()
	result, err := c.provider.Translate(key, data)
	if err != nil {
		return "", err
	}

	// Store in cache
	c.cache.Set(cacheKey, cacheEntry{
		value:     result,
		expiresAt: time.Now().Add(c.ttl),
	})

	return result, nil
}

func (c *CachedProvider) TranslatePlural(key string, count interface{}, data map[string]interface{}) (string, error) {
	cacheKey := c.getCacheKey(key, data)
	countInt, _ := count.(int)

	// Try to get from cache
	if entry, ok := c.cache.Get(cacheKey); ok {
		e := entry.(cacheEntry)
		if e.isPlural && time.Now().Before(e.expiresAt) {
			if result, ok := e.pluralData[getPluralKey(countInt)]; ok {
				c.metrics.RecordHit()
				return result, nil
			}
		}
		c.cache.Remove(cacheKey)
	}

	// Get from provider
	c.metrics.RecordMiss()
	result, err := c.provider.TranslatePlural(key, count, data)
	if err != nil {
		return "", err
	}

	// Store in cache
	c.cache.Set(cacheKey, cacheEntry{
		value:      result,
		expiresAt:  time.Now().Add(c.ttl),
		isPlural:   true,
		pluralData: map[string]string{getPluralKey(countInt): result},
	})

	return result, nil
}

func (c *CachedProvider) LoadTranslations(path string, format string) error {
	return c.provider.LoadTranslations(path, format)
}

func (c *CachedProvider) ClearCache() {
	c.cache.Clear()
	c.metrics.ResetStats()
}

func (c *CachedProvider) GetLanguages() []language.Tag {
	return c.provider.GetLanguages()
}

func (c *CachedProvider) SetLanguages(languages ...string) error {
	return c.provider.SetLanguages(languages...)
}
