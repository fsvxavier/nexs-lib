package slog

import (
	"github.com/fsvxavier/nexs-lib/observability/logger"
)

func init() {
	logger.RegisterProvider("slog", NewProvider())
}
