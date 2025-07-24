package zerolog

import (
	"github.com/fsvxavier/nexs-lib/observability/logger"
)

const ProviderName = "zerolog"

func init() {
	logger.RegisterProvider(ProviderName, NewProvider())
}
