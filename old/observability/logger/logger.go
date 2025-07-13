package logger

import "context"

var loggerProvider Logger = &NoopLog{}

type Logger interface {
	Debugf(ctx context.Context, format string, args ...interface{})
	Infof(ctx context.Context, format string, args ...interface{})
	Warnf(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
	Panicf(ctx context.Context, format string, args ...interface{})
	DebugCode(ctx context.Context, code, format string, args ...interface{})
	InfoCode(ctx context.Context, code, format string, args ...interface{})
	WarnCode(ctx context.Context, code, format string, args ...interface{})
	ErrorCode(ctx context.Context, code, format string, args ...interface{})
	PanicCode(ctx context.Context, code, format string, args ...interface{})
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	loggerProvider.Debugf(ctx, format, args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	loggerProvider.Infof(ctx, format, args...)
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	loggerProvider.Warnf(ctx, format, args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	loggerProvider.Errorf(ctx, format, args...)
}

func Panicf(ctx context.Context, format string, args ...interface{}) {
	loggerProvider.Panicf(ctx, format, args...)
}

func DebugCode(ctx context.Context, code, format string, args ...interface{}) {
	loggerProvider.DebugCode(ctx, code, format, args...)
}

func InfoCode(ctx context.Context, code, format string, args ...interface{}) {
	loggerProvider.InfoCode(ctx, code, format, args...)
}

func WarnCode(ctx context.Context, code, format string, args ...interface{}) {
	loggerProvider.WarnCode(ctx, code, format, args...)
}

func ErrorCode(ctx context.Context, code, format string, args ...interface{}) {
	loggerProvider.ErrorCode(ctx, code, format, args...)
}

func PanicCode(ctx context.Context, code, format string, args ...interface{}) {
	loggerProvider.PanicCode(ctx, code, format, args...)
}

func SetLoggerProvider(provider Logger) {
	loggerProvider = provider
}

type NoopLog struct{}

func (*NoopLog) Debugf(ctx context.Context, format string, args ...interface{})          {}
func (*NoopLog) Infof(ctx context.Context, format string, args ...interface{})           {}
func (*NoopLog) Warnf(ctx context.Context, format string, args ...interface{})           {}
func (*NoopLog) Errorf(ctx context.Context, format string, args ...interface{})          {}
func (*NoopLog) Panicf(ctx context.Context, format string, args ...interface{})          {}
func (*NoopLog) DebugCode(ctx context.Context, code, format string, args ...interface{}) {}
func (*NoopLog) InfoCode(ctx context.Context, code, format string, args ...interface{})  {}
func (*NoopLog) WarnCode(ctx context.Context, code, format string, args ...interface{})  {}
func (*NoopLog) ErrorCode(ctx context.Context, code, format string, args ...interface{}) {}
func (*NoopLog) PanicCode(ctx context.Context, code, format string, args ...interface{}) {}
