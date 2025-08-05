// Package yaml provides a YAML-based translation provider for the i18n system.
// It supports loading translations from YAML files with features like nested keys,
// parameter interpolation, caching, and file watching.
package yaml

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	i18nconfig "github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
	"gopkg.in/yaml.v3"
)

// Provider implements the I18n interface for YAML-based translations.
type Provider struct {
	config             *i18nconfig.Config
	providerConfig     *i18nconfig.YAMLProviderConfig
	translations       map[string]map[string]interface{}
	supportedLanguages []string
	defaultLanguage    string
	started            bool
	mu                 sync.RWMutex
}

// Factory implements the ProviderFactory interface for YAML providers.
type Factory struct{}

// Name returns the name of this provider.
func (f *Factory) Name() string {
	return "yaml"
}

// Create creates a new YAML translation provider with the given configuration.
func (f *Factory) Create(cfg interface{}) (interfaces.I18n, error) {
	config, ok := cfg.(*i18nconfig.Config)
	if !ok {
		return nil, fmt.Errorf("invalid configuration type, expected *i18nconfig.Config")
	}

	var providerConfig *i18nconfig.YAMLProviderConfig
	if config.ProviderConfig != nil {
		if pc, ok := config.ProviderConfig.(*i18nconfig.YAMLProviderConfig); ok {
			providerConfig = pc
		} else {
			return nil, fmt.Errorf("invalid provider configuration type, expected *i18nconfig.YAMLProviderConfig")
		}
	} else {
		providerConfig = i18nconfig.DefaultYAMLProviderConfig()
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

// ValidateConfig validates the configuration for the YAML provider.
func (f *Factory) ValidateConfig(config interface{}) error {
	cfg, ok := config.(*i18nconfig.Config)
	if !ok {
		return fmt.Errorf("invalid configuration type, expected *i18nconfig.Config")
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid base configuration: %w", err)
	}

	if cfg.ProviderConfig != nil {
		if pc, ok := cfg.ProviderConfig.(*i18nconfig.YAMLProviderConfig); ok {
			if err := pc.Validate(); err != nil {
				return fmt.Errorf("invalid YAML provider configuration: %w", err)
			}
		} else {
			return fmt.Errorf("invalid provider configuration type, expected *i18nconfig.YAMLProviderConfig")
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

// LoadTranslations loads translation data from YAML files.
func (p *Provider) LoadTranslations(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.loadTranslationsUnsafe()
}

// GetSupportedLanguages returns the list of supported languages.
func (p *Provider) GetSupportedLanguages() []string {
	return p.supportedLanguages
}

// GetDefaultLanguage returns the default language.
func (p *Provider) GetDefaultLanguage() string {
	return p.defaultLanguage
}

// SetDefaultLanguage sets the default language.
func (p *Provider) SetDefaultLanguage(lang string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.defaultLanguage = lang
}

// HasTranslation checks if a translation exists for the given key and language.
func (p *Provider) HasTranslation(key string, lang string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	_, found := p.getTranslation(key, lang)
	return found
}

// Health returns the health status of the translation provider.
func (p *Provider) Health(ctx context.Context) error {
	return p.IsHealthy(ctx)
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

// Stop shuts down the translation provider.
func (p *Provider) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.started = false
	return nil
}

// IsHealthy checks if the provider is healthy and operational.
func (p *Provider) IsHealthy(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.started {
		return fmt.Errorf("provider not started")
	}

	// Check if we have translations loaded
	if len(p.translations) == 0 {
		return fmt.Errorf("no translations loaded")
	}

	return nil
}

// Reload reloads translation data from sources.
func (p *Provider) Reload(ctx context.Context) error {
	return p.LoadTranslations(ctx)
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

	// Replace translations atomically
	p.translations = newTranslations
	return nil
}

// loadLanguageFile loads a single YAML language file.
func (p *Provider) loadLanguageFile(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var translations map[string]interface{}
	if err := yaml.Unmarshal(data, &translations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML file %s: %w", filePath, err)
	}

	return translations, nil
}

// getTranslation retrieves a translation for a given key and language.
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

		// Navigate to next level
		if value, exists := current[part]; exists {
			if nextLevel, ok := value.(map[string]interface{}); ok {
				current = nextLevel
			} else {
				return "", false
			}
		} else {
			return "", false
		}
	}

	return "", false
}

// interpolateParams replaces placeholders in the translation with parameter values.
func (p *Provider) interpolateParams(translation string, params map[string]interface{}) string {
	if len(params) == 0 {
		return translation
	}

	result := translation
	for key, value := range params {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	return result
}

// GetTranslationCount returns the total number of translations across all languages.
func (p *Provider) GetTranslationCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := 0
	for _, langTranslations := range p.translations {
		total += p.countTranslations(langTranslations)
	}
	return total
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
