package tracer

import (
	"context"
)

var _provider Provider = &NoopProvider{}

type Provider interface {
	StartSpanFromContext(ctx context.Context, spanName string) (context.Context, SpanTrace)
	SpanFromContext(ctx context.Context) (SpanTrace, bool)
}

type SpanTrace interface {
	Finish()
}

func StartSpanFromContext(ctx context.Context, spanName string) (context.Context, SpanTrace) {
	return _provider.StartSpanFromContext(ctx, spanName)
}

func SpanFromContext(ctx context.Context) (SpanTrace, bool) {
	return _provider.SpanFromContext(ctx)
}

func SetProvider(provider Provider) {
	_provider = provider
}

type NoopProvider struct{}

// StartSpanFromContext implements Provider
func (*NoopProvider) StartSpanFromContext(ctx context.Context, spanName string) (context.Context, SpanTrace) {
	return ctx, &NoopSpan{}
}

// SpanFromContext implements Provider
func (*NoopProvider) SpanFromContext(ctx context.Context) (SpanTrace, bool) {
	return &NoopSpan{}, true
}

// Verify Interface Compliance
var _ Provider = (*NoopProvider)(nil)

type NoopSpan struct{}

// Finish implements Span
func (*NoopSpan) Finish() {}

// Verify Interface Compliance
var _ SpanTrace = (*NoopSpan)(nil)
