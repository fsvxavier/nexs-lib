package middleware

// contextKey type for middleware context keys
type contextKey string

const (
	// LocalizerKey is the context key for the i18n localizer
	LocalizerKey contextKey = "i18n_localizer"
)
