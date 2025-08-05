// Package interfaces defines the core interfaces for the i18n translation system.
// It provides contracts for translation providers, observers, hooks, and middlewares.
package interfaces

import (
	"context"
	"time"
)

// I18n defines the main interface for translation providers.
// All translation providers must implement this interface to be compatible
// with the i18n system.
type I18n interface {
	// Translate translates a key to the target language with optional parameters
	Translate(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error)

	// LoadTranslations loads translation data from the configured source
	LoadTranslations(ctx context.Context) error

	// GetSupportedLanguages returns a list of supported language codes
	GetSupportedLanguages() []string

	// HasTranslation checks if a translation exists for the given key and language
	HasTranslation(key string, lang string) bool

	// GetDefaultLanguage returns the default language code
	GetDefaultLanguage() string

	// SetDefaultLanguage sets the default language code
	SetDefaultLanguage(lang string)

	// Start initializes the translation provider
	Start(ctx context.Context) error

	// Stop gracefully shuts down the translation provider
	Stop(ctx context.Context) error

	// Health returns the health status of the translation provider
	Health(ctx context.Context) error
}

// I18nObserver defines the interface for observing i18n events.
// Components that need to react to translation provider lifecycle events
// should implement this interface.
type I18nObserver interface {
	// OnStart is called when a translation provider starts
	OnStart(ctx context.Context, providerName string) error

	// OnStop is called when a translation provider stops
	OnStop(ctx context.Context, providerName string) error

	// OnError is called when a translation provider encounters an error
	OnError(ctx context.Context, providerName string, err error) error

	// OnTranslate is called when a translation is performed
	OnTranslate(ctx context.Context, providerName string, key string, lang string, result string) error
}

// Hook defines the interface for hooks that can be executed at specific
// points in the translation provider lifecycle.
type Hook interface {
	I18nObserver

	// Name returns the unique name of the hook
	Name() string

	// Priority returns the execution priority (lower numbers execute first)
	Priority() int
}

// Middleware defines the interface for middlewares that can wrap
// translation operations to add additional functionality.
type Middleware interface {
	I18nObserver

	// Name returns the unique name of the middleware
	Name() string

	// WrapTranslate wraps a translation operation
	WrapTranslate(next TranslateFunc) TranslateFunc
}

// TranslateFunc is a function type for translation operations.
// Used by middlewares to wrap translation calls.
type TranslateFunc func(ctx context.Context, key string, lang string, params map[string]interface{}) (string, error)

// ProviderFactory defines the interface for creating translation providers.
// Each provider type must implement this factory interface.
type ProviderFactory interface {
	// Create creates a new instance of the translation provider
	Create(config interface{}) (I18n, error)

	// Name returns the unique name of the provider type
	Name() string

	// ValidateConfig validates the provider-specific configuration
	ValidateConfig(config interface{}) error
}

// Registry defines the interface for managing translation providers.
// It handles registration, creation, and lifecycle management of providers.
type Registry interface {
	// RegisterProvider registers a new provider factory
	RegisterProvider(factory ProviderFactory) error

	// CreateProvider creates a provider instance by name with configuration
	CreateProvider(name string, config interface{}) (I18n, error)

	// GetProviderNames returns all registered provider names
	GetProviderNames() []string

	// HasProvider checks if a provider is registered
	HasProvider(name string) bool

	// AddHook adds a hook to be executed for all providers
	AddHook(hook Hook) error

	// RemoveHook removes a hook by name
	RemoveHook(name string) error

	// AddMiddleware adds a middleware to be applied to all providers
	AddMiddleware(middleware Middleware) error

	// RemoveMiddleware removes a middleware by name
	RemoveMiddleware(name string) error

	// GetHooks returns all registered hooks
	GetHooks() []Hook

	// GetMiddlewares returns all registered middlewares
	GetMiddlewares() []Middleware
}

// CurrencyFormatter defines the interface for currency formatting operations.
// Provides methods to format monetary values according to different locales.
type CurrencyFormatter interface {
	// FormatCurrency formats a monetary value according to the specified locale and currency
	FormatCurrency(amount float64, currency string, locale string) (string, error)

	// GetSupportedCurrencies returns a list of supported currency codes
	GetSupportedCurrencies() []string

	// GetSupportedLocales returns a list of supported locale codes for currency formatting
	GetSupportedLocales() []string

	// ParseCurrency parses a formatted currency string back to a numeric value
	ParseCurrency(formatted string, currency string, locale string) (float64, error)
}

// DateTimeFormatter defines the interface for date and time formatting operations.
// Provides methods to format dates and times according to different locales.
type DateTimeFormatter interface {
	// FormatDateTime formats a time value according to the specified locale and format
	FormatDateTime(t time.Time, locale string, format string) (string, error)

	// FormatDate formats a date according to the specified locale
	FormatDate(t time.Time, locale string) (string, error)

	// FormatTime formats a time according to the specified locale
	FormatTime(t time.Time, locale string) (string, error)

	// ParseDateTime parses a formatted datetime string back to a time value
	ParseDateTime(formatted string, locale string, format string) (time.Time, error)

	// GetSupportedFormats returns a list of supported date/time formats for a locale
	GetSupportedFormats(locale string) []string
}

// LocaleProvider defines the interface for locale-specific operations.
// Combines translation, currency formatting, and datetime formatting capabilities.
type LocaleProvider interface {
	I18n
	CurrencyFormatter
	DateTimeFormatter

	// GetLocaleInfo returns detailed information about a specific locale
	GetLocaleInfo(locale string) (*LocaleInfo, error)

	// SetLocale sets the current locale for the provider
	SetLocale(locale string) error

	// GetCurrentLocale returns the currently active locale
	GetCurrentLocale() string
}

// LocaleInfo contains detailed information about a specific locale.
type LocaleInfo struct {
	// Code is the locale code (e.g., "en-US", "pt-BR")
	Code string `json:"code"`

	// Name is the human-readable name of the locale
	Name string `json:"name"`

	// Language is the primary language code (e.g., "en", "pt")
	Language string `json:"language"`

	// Country is the country code (e.g., "US", "BR")
	Country string `json:"country"`

	// CurrencyCode is the default currency for this locale
	CurrencyCode string `json:"currency_code"`

	// DateFormat is the default date format for this locale
	DateFormat string `json:"date_format"`

	// TimeFormat is the default time format for this locale
	TimeFormat string `json:"time_format"`

	// DecimalSeparator is the character used for decimal separation
	DecimalSeparator string `json:"decimal_separator"`

	// ThousandsSeparator is the character used for thousands separation
	ThousandsSeparator string `json:"thousands_separator"`

	// RTL indicates if this locale uses right-to-left text direction
	RTL bool `json:"rtl"`
}
