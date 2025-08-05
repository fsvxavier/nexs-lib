// Package json provides a JSON-based translation provider for the i18n system.
// It implements the I18n interface and supports loading translations from JSON files.
package json

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	i18nconfig "github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
)

// Provider is a JSON-based translation provider.
// It loads translations from JSON files and implements the I18n interface.
type Provider struct {
	config             *i18nconfig.Config
	providerConfig     *i18nconfig.JSONProviderConfig
	translations       map[string]map[string]interface{}
	supportedLanguages []string
	defaultLanguage    string
	started            bool
	mu                 sync.RWMutex
}

// Factory is the factory for creating JSON translation providers.
type Factory struct{}

// NewFactory creates a new JSON provider factory.
func NewFactory() *Factory {
	return &Factory{}
}

// Create creates a new JSON translation provider with the given configuration.
func (f *Factory) Create(cfg interface{}) (interfaces.I18n, error) {
	config, ok := cfg.(*i18nconfig.Config)
	if !ok {
		return nil, fmt.Errorf("invalid configuration type, expected *i18nconfig.Config")
	}

	var providerConfig *i18nconfig.JSONProviderConfig
	if config.ProviderConfig != nil {
		if pc, ok := config.ProviderConfig.(*i18nconfig.JSONProviderConfig); ok {
			providerConfig = pc
		} else {
			return nil, fmt.Errorf("invalid provider configuration type, expected *i18nconfig.JSONProviderConfig")
		}
	} else {
		providerConfig = i18nconfig.DefaultJSONProviderConfig()
	}

	provider := &Provider{
		config:             config,
		providerConfig:     providerConfig,
		translations:       make(map[string]map[string]interface{}),
		supportedLanguages: config.SupportedLanguages,
		defaultLanguage:    config.DefaultLanguage,
		started:            false,
	}

	return provider, nil
}

// Name returns the name of the provider factory.
func (f *Factory) Name() string {
	return "json"
}

// ValidateConfig validates the configuration for the JSON provider.
func (f *Factory) ValidateConfig(config interface{}) error {
	cfg, ok := config.(*i18nconfig.Config)
	if !ok {
		return fmt.Errorf("invalid configuration type, expected *i18nconfig.Config")
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid base configuration: %w", err)
	}

	if cfg.ProviderConfig != nil {
		if pc, ok := cfg.ProviderConfig.(*i18nconfig.JSONProviderConfig); ok {
			if err := pc.Validate(); err != nil {
				return fmt.Errorf("invalid JSON provider configuration: %w", err)
			}
		} else {
			return fmt.Errorf("invalid provider configuration type, expected *i18nconfig.JSONProviderConfig")
		}
	}

	return nil
}

// Translate translates a key to the target language with optional parameters.
func (p *Provider) Translate(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.started {
		return "", fmt.Errorf("provider not started")
	}

	if key == "" {
		return "", fmt.Errorf("translation key cannot be empty")
	}

	if lang == "" {
		return "", fmt.Errorf("language cannot be empty")
	}

	// Try to get translation for the requested language
	if translation, found := p.getTranslation(key, lang); found {
		return p.interpolateParams(translation, params), nil
	}

	// Fallback to default language if enabled and different from requested
	if p.config.FallbackToDefault && lang != p.defaultLanguage {
		if translation, found := p.getTranslation(key, p.defaultLanguage); found {
			return p.interpolateParams(translation, params), nil
		}
	}

	// Return error if strict mode is enabled
	if p.config.StrictMode {
		return "", fmt.Errorf("translation not found for key '%s' in language '%s'", key, lang)
	}

	// Return the key itself as fallback
	return key, nil
}

// LoadTranslations loads translation data from JSON files.
func (p *Provider) LoadTranslations(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.providerConfig.FilePath == "" {
		return fmt.Errorf("file path not configured")
	}

	// Check if the directory exists
	if _, err := os.Stat(p.providerConfig.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("translation directory does not exist: %s", p.providerConfig.FilePath)
	}

	// Load translations for each supported language
	newTranslations := make(map[string]map[string]interface{})

	for _, lang := range p.supportedLanguages {
		filename := strings.ReplaceAll(p.providerConfig.FilePattern, "{lang}", lang)
		filePath := filepath.Join(p.providerConfig.FilePath, filename)

		translations, err := p.loadLanguageFile(filePath)
		if err != nil {
			// If file doesn't exist, create empty translations for this language
			if os.IsNotExist(err) {
				newTranslations[lang] = make(map[string]interface{})
				continue
			}
			return fmt.Errorf("failed to load translations for language '%s': %w", lang, err)
		}

		newTranslations[lang] = translations
	}

	p.translations = newTranslations
	return nil
}

// GetSupportedLanguages returns a list of supported language codes.
func (p *Provider) GetSupportedLanguages() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	languages := make([]string, len(p.supportedLanguages))
	copy(languages, p.supportedLanguages)
	return languages
}

// HasTranslation checks if a translation exists for the given key and language.
func (p *Provider) HasTranslation(key string, lang string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if lang == "" {
		lang = p.defaultLanguage
	}

	_, found := p.getTranslation(key, lang)
	return found
}

// GetDefaultLanguage returns the default language code.
func (p *Provider) GetDefaultLanguage() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.defaultLanguage
}

// SetDefaultLanguage sets the default language code.
func (p *Provider) SetDefaultLanguage(lang string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.defaultLanguage = lang
}

// Start initializes the translation provider.
func (p *Provider) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return nil // Already started
	}

	// Load translations with timeout
	loadCtx, cancel := context.WithTimeout(ctx, p.config.LoadTimeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- p.loadTranslationsUnsafe()
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to load translations: %w", err)
		}
	case <-loadCtx.Done():
		return fmt.Errorf("timeout loading translations after %v", p.config.LoadTimeout)
	}

	p.started = true
	return nil
}

// Stop gracefully shuts down the translation provider.
func (p *Provider) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.started = false

	// Clear translations to free memory
	p.translations = make(map[string]map[string]interface{})

	return nil
}

// Health returns the health status of the translation provider.
func (p *Provider) Health(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.started {
		return fmt.Errorf("provider not started")
	}

	// Check if we have translations for the default language
	if _, exists := p.translations[p.defaultLanguage]; !exists {
		return fmt.Errorf("no translations available for default language '%s'", p.defaultLanguage)
	}

	// Check if translation files are still accessible
	if p.providerConfig.FilePath != "" {
		if _, err := os.Stat(p.providerConfig.FilePath); os.IsNotExist(err) {
			return fmt.Errorf("translation directory no longer exists: %s", p.providerConfig.FilePath)
		}
	}

	return nil
}

// loadLanguageFile loads translations from a JSON file.
func (p *Provider) loadLanguageFile(filePath string) (map[string]interface{}, error) {
	// Check file size if limit is set
	if p.providerConfig.MaxFileSize > 0 {
		if info, err := os.Stat(filePath); err == nil {
			if info.Size() > p.providerConfig.MaxFileSize {
				return nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", info.Size(), p.providerConfig.MaxFileSize)
			}
		}
	}

	// Read file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var translations map[string]interface{}
	if err := json.Unmarshal(data, &translations); err != nil {
		if p.providerConfig.ValidateJSON {
			return nil, fmt.Errorf("invalid JSON in file %s: %w", filePath, err)
		}

		// If validation is disabled, return empty translations
		translations = make(map[string]interface{})
	}

	return translations, nil
}

// loadTranslationsUnsafe loads translations without locking (internal use only).
func (p *Provider) loadTranslationsUnsafe() error {
	if p.providerConfig.FilePath == "" {
		return fmt.Errorf("file path not configured")
	}

	// Check if the directory exists
	if _, err := os.Stat(p.providerConfig.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("translation directory does not exist: %s", p.providerConfig.FilePath)
	}

	// Load translations for each supported language
	newTranslations := make(map[string]map[string]interface{})

	for _, lang := range p.supportedLanguages {
		filename := strings.ReplaceAll(p.providerConfig.FilePattern, "{lang}", lang)
		filePath := filepath.Join(p.providerConfig.FilePath, filename)

		translations, err := p.loadLanguageFile(filePath)
		if err != nil {
			// If file doesn't exist, create empty translations for this language
			if os.IsNotExist(err) {
				newTranslations[lang] = make(map[string]interface{})
				continue
			}
			return fmt.Errorf("failed to load translations for language '%s': %w", lang, err)
		}

		newTranslations[lang] = translations
	}

	p.translations = newTranslations
	return nil
}

// getTranslation retrieves a translation for the given key and language.
func (p *Provider) getTranslation(key string, lang string) (string, bool) {
	langTranslations, exists := p.translations[lang]
	if !exists {
		return "", false
	}

	// Support nested keys if enabled
	if p.providerConfig.NestedKeys && strings.Contains(key, ".") {
		return p.getNestedTranslation(langTranslations, key)
	}

	// Direct key lookup
	if value, exists := langTranslations[key]; exists {
		if strValue, ok := value.(string); ok {
			return strValue, true
		}
	}

	return "", false
}

// getNestedTranslation retrieves a nested translation using dot notation.
func (p *Provider) getNestedTranslation(translations map[string]interface{}, key string) (string, bool) {
	parts := strings.Split(key, ".")
	current := translations

	// Navigate through nested structure
	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - should be the actual translation
			if value, exists := current[part]; exists {
				if strValue, ok := value.(string); ok {
					return strValue, true
				}
			}
			return "", false
		}

		// Intermediate part - should be a map
		if value, exists := current[part]; exists {
			if mapValue, ok := value.(map[string]interface{}); ok {
				current = mapValue
			} else {
				return "", false
			}
		} else {
			return "", false
		}
	}

	return "", false
}

// interpolateParams replaces parameters in the translation string.
func (p *Provider) interpolateParams(translation string, params map[string]interface{}) string {
	if params == nil || len(params) == 0 {
		return translation
	}

	result := translation
	for key, value := range params {
		placeholder := fmt.Sprintf("{{%s}}", key)
		replacement := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}

	return result
}

// GetTranslationCount returns the total number of loaded translations.
func (p *Provider) GetTranslationCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	count := 0
	for _, langTranslations := range p.translations {
		count += p.countTranslations(langTranslations)
	}

	return count
}

// countTranslations recursively counts translations in a map.
func (p *Provider) countTranslations(translations map[string]interface{}) int {
	count := 0
	for _, value := range translations {
		switch v := value.(type) {
		case string:
			count++
		case map[string]interface{}:
			count += p.countTranslations(v)
		}
	}
	return count
}

// GetTranslationCountByLanguage returns the number of translations for a specific language.
func (p *Provider) GetTranslationCountByLanguage(lang string) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if langTranslations, exists := p.translations[lang]; exists {
		return p.countTranslations(langTranslations)
	}

	return 0
}

// ReloadTranslations reloads translations from files.
func (p *Provider) ReloadTranslations(ctx context.Context) error {
	return p.LoadTranslations(ctx)
}

// GetLoadedLanguages returns the list of languages that have been loaded.
func (p *Provider) GetLoadedLanguages() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	languages := make([]string, 0, len(p.translations))
	for lang := range p.translations {
		languages = append(languages, lang)
	}
	return languages
}
