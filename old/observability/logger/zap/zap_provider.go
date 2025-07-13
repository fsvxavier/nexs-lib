package zap

import (
	"context"

	"github.com/dock-tech/isis-golang-lib/observability/logger"
)

type Provider struct{}

func setZapProvider() {
	logger.SetLoggerProvider(&Provider{})
}

var _ logger.Logger = (*Provider)(nil)

// Debugf implements Logger.
func (*Provider) Debugf(ctx context.Context, format string, args ...interface{}) {
	Debugf(ctx, format, args...)
}

// Infof implements Logger.
func (*Provider) Infof(ctx context.Context, format string, args ...interface{}) {
	Infof(ctx, format, args...)
}

// Warnf implements Logger.
func (*Provider) Warnf(ctx context.Context, format string, args ...interface{}) {
	Warnf(ctx, format, args...)
}

// Errorf implements Logger.
func (*Provider) Errorf(ctx context.Context, format string, args ...interface{}) {
	Errorf(ctx, format, args...)
}

// Panicf implements Logger.
func (*Provider) Panicf(ctx context.Context, format string, args ...interface{}) {
	Panicf(ctx, format, args...)
}

func (*Provider) DebugCode(ctx context.Context, code, format string, args ...interface{}) {
	DebugCode(ctx, code, format, args...)
}

func (*Provider) InfoCode(ctx context.Context, code, format string, args ...interface{}) {
	InfoCode(ctx, code, format, args...)
}

func (*Provider) WarnCode(ctx context.Context, code, format string, args ...interface{}) {
	WarnCode(ctx, code, format, args...)
}

func (*Provider) ErrorCode(ctx context.Context, code, format string, args ...interface{}) {
	ErrorCode(ctx, code, format, args...)
}

func (*Provider) PanicCode(ctx context.Context, code, format string, args ...interface{}) {
	PanicCode(ctx, code, format, args...)
}
