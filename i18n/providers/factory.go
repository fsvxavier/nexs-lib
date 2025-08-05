package providers

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/i18n/config"
	"github.com/fsvxavier/nexs-lib/i18n/interfaces"
)

// CreateProvider creates a new provider based on the format
func CreateProvider(format string) (interfaces.Provider, error) {
	switch format {
	case config.FormatJSON:
		return NewJSONProvider(), nil
	case "yaml":
		return NewYAMLProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
