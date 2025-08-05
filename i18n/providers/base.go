package providers

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/text/language"
)

// extractLanguageFromPath extracts language code from file path
func extractLanguageFromPath(path string) string {
	parts := strings.Split(filepath.Base(path), ".")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// BaseProvider implements the common functionality for all providers
type BaseProvider struct {
	sync.RWMutex
	translations map[language.Tag]map[string]interface{}
	languages    []language.Tag
	fallbacks    []language.Tag
}

// NewBaseProvider creates a new base provider
func NewBaseProvider() *BaseProvider {
	return &BaseProvider{
		translations: make(map[language.Tag]map[string]interface{}),
		languages:    make([]language.Tag, 0),
		fallbacks:    make([]language.Tag, 0),
	}
}

// SetLanguages implements Provider.SetLanguages
func (p *BaseProvider) SetLanguages(langs ...string) error {
	p.Lock()
	defer p.Unlock()

	tags := make([]language.Tag, 0, len(langs))
	for _, lang := range langs {
		tag, err := language.Parse(lang)
		if err != nil {
			return fmt.Errorf("invalid language code %q: %w", lang, err)
		}
		tags = append(tags, tag)
	}

	p.languages = tags
	return nil
}

// GetLanguages implements Provider.GetLanguages
func (p *BaseProvider) GetLanguages() []language.Tag {
	p.RLock()
	defer p.RUnlock()

	result := make([]language.Tag, len(p.languages))
	copy(result, p.languages)
	return result
}

// SetFallbacks sets the fallback languages
func (p *BaseProvider) SetFallbacks(langs ...language.Tag) {
	p.Lock()
	defer p.Unlock()

	p.fallbacks = make([]language.Tag, len(langs))
	copy(p.fallbacks, langs)
}

// getTranslation returns a translation for a key trying all configured languages
func (p *BaseProvider) getTranslation(key string) (interface{}, bool) {
	p.RLock()
	defer p.RUnlock()

	keys := strings.Split(key, ".")

	// Try configured languages
	for _, lang := range p.languages {
		if translations, ok := p.translations[lang]; ok {
			current := translations
			found := true

			// Navigate through nested keys
			for i, k := range keys[:len(keys)-1] {
				if next, ok := current[k].(map[string]interface{}); ok {
					current = next
				} else {
					// If we can't navigate further, try the full remaining key
					remainingKey := strings.Join(keys[i:], ".")
					if val, ok := translations[remainingKey]; ok {
						return val, true
					}
					found = false
					break
				}
			}

			if found {
				// Check the final key
				if val, ok := current[keys[len(keys)-1]]; ok {
					return val, true
				}
			}
		}
	}

	// Try fallback languages with the same nested key logic
	for _, lang := range p.fallbacks {
		if translations, ok := p.translations[lang]; ok {
			current := translations
			found := true

			// Navigate through nested keys
			for i, k := range keys[:len(keys)-1] {
				if next, ok := current[k].(map[string]interface{}); ok {
					current = next
				} else {
					// If we can't navigate further, try the full remaining key
					remainingKey := strings.Join(keys[i:], ".")
					if val, ok := translations[remainingKey]; ok {
						return val, true
					}
					found = false
					break
				}
			}

			if found {
				// Check the final key
				if val, ok := current[keys[len(keys)-1]]; ok {
					return val, true
				}
			}
		}
	}

	return nil, false
}

// addTranslation adds a single translation to the provider
func (p *BaseProvider) addTranslation(lang language.Tag, key string, value interface{}) {
	p.Lock()
	defer p.Unlock()

	if p.translations == nil {
		p.translations = make(map[language.Tag]map[string]interface{})
	}

	if p.translations[lang] == nil {
		p.translations[lang] = make(map[string]interface{})
	}

	// Handle nested keys (e.g., "nested.greeting")
	keys := strings.Split(key, ".")
	current := p.translations[lang]

	for i, k := range keys[:len(keys)-1] {
		if _, ok := current[k]; !ok {
			current[k] = make(map[string]interface{})
		}
		if nextMap, ok := current[k].(map[string]interface{}); ok {
			current = nextMap
		} else {
			// If we can't navigate further, store the remaining key parts joined
			remainingKey := strings.Join(keys[i:], ".")
			current[remainingKey] = value
			return
		}
	}

	current[keys[len(keys)-1]] = value
}
