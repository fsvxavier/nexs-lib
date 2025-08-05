package cache

import (
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
)

type cacheEntry struct {
	value      string
	expiresAt  time.Time
	isPlural   bool
	pluralData map[string]string
}

type CachedProvider struct {
	provider    interfaces.Provider
	cache       sync.Map
	ttl         time.Duration
	maxEntries  int
	entriesLock sync.RWMutex
	entries     int
}

// NewCachedProvider creates a new cached provider decorator
func NewCachedProvider(provider interfaces.Provider, ttl time.Duration, maxEntries int) *CachedProvider {
	return &CachedProvider{
		provider:   provider,
		ttl:        ttl,
		maxEntries: maxEntries,
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
	if entry, ok := c.cache.Load(cacheKey); ok {
		if e := entry.(cacheEntry); !e.isPlural && time.Now().Before(e.expiresAt) {
			return e.value, nil
		}
		c.cache.Delete(cacheKey)
		c.decrementEntries()
	}

	// Get from provider
	result, err := c.provider.Translate(key, data)
	if err != nil {
		return "", err
	}

	// Store in cache
	c.storeInCache(cacheKey, cacheEntry{
		value:     result,
		expiresAt: time.Now().Add(c.ttl),
	})

	return result, nil
}

func (c *CachedProvider) TranslatePlural(key string, count int, data map[string]interface{}) (string, error) {
	cacheKey := c.getCacheKey(key, data)

	// Try to get from cache
	if entry, ok := c.cache.Load(cacheKey); ok {
		if e := entry.(cacheEntry); e.isPlural && time.Now().Before(e.expiresAt) {
			if result, ok := e.pluralData[getPluralKey(count)]; ok {
				return result, nil
			}
		}
		c.cache.Delete(cacheKey)
		c.decrementEntries()
	}

	// Get from provider
	result, err := c.provider.TranslatePlural(key, count, data)
	if err != nil {
		return "", err
	}

	// Store in cache
	c.storeInCache(cacheKey, cacheEntry{
		value:      result,
		expiresAt:  time.Now().Add(c.ttl),
		isPlural:   true,
		pluralData: map[string]string{getPluralKey(count): result},
	})

	return result, nil
}

func (c *CachedProvider) LoadTranslations(path string, format string) error {
	return c.provider.LoadTranslations(path, format)
}

func (c *CachedProvider) ClearCache() {
	c.cache.Range(func(key, _ interface{}) bool {
		c.cache.Delete(key)
		return true
	})
	c.entriesLock.Lock()
	c.entries = 0
	c.entriesLock.Unlock()
}

func (c *CachedProvider) storeInCache(key string, entry cacheEntry) {
	// Check if we need to evict entries
	if c.maxEntries > 0 {
		c.entriesLock.Lock()
		if c.entries >= c.maxEntries {
			c.evictOldest()
		}
		c.entries++
		c.entriesLock.Unlock()
	}

	c.cache.Store(key, entry)
}

func (c *CachedProvider) decrementEntries() {
	c.entriesLock.Lock()
	c.entries--
	if c.entries < 0 {
		c.entries = 0
	}
	c.entriesLock.Unlock()
}

func (c *CachedProvider) evictOldest() {
	var oldestKey interface{}
	var oldestTime time.Time
	first := true

	c.cache.Range(func(key, value interface{}) bool {
		entry := value.(cacheEntry)
		if first || entry.expiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.expiresAt
			first = false
		}
		return true
	})

	if oldestKey != nil {
		c.cache.Delete(oldestKey)
		c.entries--
	}
}
