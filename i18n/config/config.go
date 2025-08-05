package config

import (
	"fmt"
	"path/filepath"

	"golang.org/x/text/language"
)

// Config represents the configuration for i18n
type Config struct {
	// DefaultLanguage is the default language to use when no language is specified
	DefaultLanguage language.Tag

	// FallbackLanguages are the languages to try in order when a translation is not found
	FallbackLanguages []language.Tag

	// TranslationsPath is the path to the translations files
	TranslationsPath string

	// TranslationsFormat is the format of the translations files (json, yaml)
	TranslationsFormat string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	ptBR := language.MustParse("pt-BR")
	return &Config{
		DefaultLanguage:    ptBR,
		FallbackLanguages:  []language.Tag{ptBR, language.English},
		TranslationsPath:   "translations",
		TranslationsFormat: "json",
	}
}

// WithDefaultLanguage sets the default language
func (c *Config) WithDefaultLanguage(lang language.Tag) *Config {
	c.DefaultLanguage = lang
	return c
}

// WithFallbackLanguages sets the fallback languages
func (c *Config) WithFallbackLanguages(langs ...language.Tag) *Config {
	c.FallbackLanguages = langs
	return c
}

// WithTranslationsPath sets the translations path
func (c *Config) WithTranslationsPath(path string) *Config {
	c.TranslationsPath = path
	return c
}

// WithTranslationsFormat sets the translations format
func (c *Config) WithTranslationsFormat(format string) *Config {
	c.TranslationsFormat = format
	return c
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.DefaultLanguage == language.Und {
		return ErrInvalidDefaultLanguage
	}

	if len(c.FallbackLanguages) == 0 {
		return ErrNoFallbackLanguages
	}

	if c.TranslationsPath == "" {
		return ErrInvalidTranslationsPath
	}

	if !isValidFormat(c.TranslationsFormat) {
		return ErrInvalidTranslationsFormat
	}

	return nil
}

// GetTranslationFilePath returns the full path for a language file
func (c *Config) GetTranslationFilePath(lang string) string {
	// For JSON/YAML files, use flatter structure
	filename := fmt.Sprintf(TranslationFilePattern, lang, c.TranslationsFormat)
	return filepath.Join(c.TranslationsPath, filename)
}

// isValidFormat checks if the format is supported
func isValidFormat(format string) bool {
	switch format {
	case "json", "yaml", "yml":
		return true
	default:
		return false
	}
}
