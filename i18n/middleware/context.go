package middleware

import (
	"context"

	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
)

// contextKey is a custom type for context keys
type contextKey string

const (
	// LocalizerKey is the key used to store the localizer in the context
	LocalizerKey = contextKey("localizer")
)

// GetProvider retrieves the i18n provider from the request context
func GetProvider(ctx context.Context) (interfaces.Provider, bool) {
	v := ctx.Value(LocalizerKey)
	if v == nil {
		return nil, false
	}
	provider, ok := v.(interfaces.Provider)
	return provider, ok
}

// MustGetProvider retrieves the i18n provider from the request context
// and panics if not found
func MustGetProvider(ctx context.Context) interfaces.Provider {
	provider, ok := GetProvider(ctx)
	if !ok {
		panic("i18n provider not found in context")
	}
	return provider
}
