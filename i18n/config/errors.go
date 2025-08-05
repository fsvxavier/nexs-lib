package config

import "errors"

var (
	// ErrInvalidDefaultLanguage is returned when the default language is invalid
	ErrInvalidDefaultLanguage = errors.New("invalid default language")

	// ErrNoFallbackLanguages is returned when no fallback languages are specified
	ErrNoFallbackLanguages = errors.New("no fallback languages specified")

	// ErrInvalidTranslationsPath is returned when the translations path is invalid
	ErrInvalidTranslationsPath = errors.New("invalid translations path")

	// ErrInvalidTranslationsFormat is returned when the translations format is invalid
	ErrInvalidTranslationsFormat = errors.New("invalid translations format: must be json or yaml")
)
