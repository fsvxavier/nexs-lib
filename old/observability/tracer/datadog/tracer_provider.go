package datadog

import (
	"context"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	commontracer "github.com/dock-tech/isis-golang-lib/observability/tracer"
)

func setProvider(service, env, version string) {
	commontracer.SetProvider(&Provider{service: service, env: env, version: version})
}

type SpanTrace struct {
	*tracer.Span
}

// Finish implements tracer.Span
func (s *SpanTrace) Finish() {
	s.Span.Finish()
}

type Provider struct {
	service string
	env     string
	version string
}

// StartSpanFromContext implements tracer.Provider
func (*Provider) StartSpanFromContext(ctx context.Context, spanName string) (context.Context, commontracer.SpanTrace) {
	span := &SpanTrace{}
	span.Span, ctx = tracer.StartSpanFromContext(ctx, spanName)

	return ctx, span
}

func (*Provider) SpanFromContext(ctx context.Context) (commontracer.SpanTrace, bool) {
	span := &SpanTrace{}
	var found bool

	span.Span, found = tracer.SpanFromContext(ctx)
	return span, found
}
