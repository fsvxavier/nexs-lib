package zap

import (
	"github.com/fsvxavier/nexs-lib/observability/logger"
)

func init() {
	logger.RegisterProvider("zap", NewProvider())
}
