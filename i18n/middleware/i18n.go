package middleware

import (
	"context"
	"net/http"

	"golang.org/x/text/language"

	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
)

const (
	// DefaultHeaderName is the default header name for language preference
	DefaultHeaderName = "Accept-Language"
) // Config represents the configuration for the i18n middleware
type Config struct {
	// Provider is the i18n provider to use
	Provider interfaces.Provider

	// HeaderName is the name of the header containing language preferences
	HeaderName string

	// QueryParam is the name of the query parameter containing language preference
	QueryParam string

	// DefaultLanguage is the default language to use when no language is specified
	DefaultLanguage language.Tag
}

// New creates a new i18n middleware
func New(cfg Config) func(next http.Handler) http.Handler {
	if cfg.HeaderName == "" {
		cfg.HeaderName = DefaultHeaderName
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to get language from query parameter
			lang := r.URL.Query().Get(cfg.QueryParam)

			// If not in query, try header
			if lang == "" {
				lang = r.Header.Get(cfg.HeaderName)
			}

			// Parse language preferences
			tags, _, err := language.ParseAcceptLanguage(lang)
			if err != nil || len(tags) == 0 {
				tags = []language.Tag{cfg.DefaultLanguage}
			}

			// Set languages in provider
			langCodes := make([]string, len(tags))
			for i, tag := range tags {
				langCodes[i] = tag.String()
			}
			if err := cfg.Provider.SetLanguages(langCodes...); err != nil {
				http.Error(w, "Invalid language preference", http.StatusBadRequest)
				return
			}

			// Store provider in context
			ctx := context.WithValue(r.Context(), LocalizerKey, cfg.Provider)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
