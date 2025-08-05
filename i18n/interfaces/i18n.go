package interfaces

import "golang.org/x/text/language"

// Provider defines the interface for i18n providers
type Provider interface {
	// LoadTranslations loads translations from the given path and format
	LoadTranslations(path string, format string) error

	// Translate translates a message with given key and template data
	Translate(key string, templateData map[string]interface{}) (string, error)

	// TranslatePlural translates a plural message with given key, count and template data
	TranslatePlural(key string, count interface{}, templateData map[string]interface{}) (string, error)

	// SetLanguages sets the preferred languages for translation
	SetLanguages(langs ...string) error

	// GetLanguages returns the current preferred languages
	GetLanguages() []language.Tag
}

// Observer defines the interface for i18n observers/hooks
type Observer interface {
	// OnTranslationLoaded is called when translations are loaded
	OnTranslationLoaded(provider Provider, path string)

	// OnTranslationNotFound is called when a translation key is not found
	OnTranslationNotFound(provider Provider, key string)

	// OnTranslationError is called when there's an error during translation
	OnTranslationError(provider Provider, key string, err error)
}

// Config defines the interface for i18n configuration
type Config interface {
	// GetDefaultLanguage returns the default language
	GetDefaultLanguage() language.Tag

	// GetFallbackLanguages returns the fallback languages in order of preference
	GetFallbackLanguages() []language.Tag

	// GetTranslationsPath returns the path to translations files
	GetTranslationsPath() string

	// GetTranslationsFormat returns the format of translations files (json, yaml, etc)
	GetTranslationsFormat() string
}
