package tracer

import (
	"context"
	"testing"
)

func TestNoopProvider_StartSpanFromContext(t *testing.T) {
	provider := &NoopProvider{}
	ctx := context.Background()
	spanName := "test-span"

	newCtx, span := provider.StartSpanFromContext(ctx, spanName)
	if newCtx != ctx {
		t.Errorf("expected context to be unchanged, got %v", newCtx)
	}
	if span == nil {
		t.Errorf("expected span to be non-nil")
	}
}

func TestNoopProvider_SpanFromContext(t *testing.T) {
	provider := &NoopProvider{}
	ctx := context.Background()

	span, ok := provider.SpanFromContext(ctx)
	if !ok {
		t.Errorf("expected ok to be true, got false")
	}
	if span == nil {
		t.Errorf("expected span to be non-nil")
	}
}

func TestNoopSpan_Finish(t *testing.T) {
	span := &NoopSpan{}
	span.Finish()
	// NoopSpan.Finish() does nothing, so just ensure it doesn't panic
}

func TestSetProvider(t *testing.T) {
	provider := &NoopProvider{}
	SetProvider(provider)
	if _provider != provider {
		t.Errorf("expected provider to be set")
	}
}

func TestStartSpanFromContext(t *testing.T) {
	provider := &NoopProvider{}
	SetProvider(provider)
	ctx := context.Background()
	spanName := "test-span"

	newCtx, span := StartSpanFromContext(ctx, spanName)
	if newCtx != ctx {
		t.Errorf("expected context to be unchanged, got %v", newCtx)
	}
	if span == nil {
		t.Errorf("expected span to be non-nil")
	}
}

func TestSpanFromContext(t *testing.T) {
	provider := &NoopProvider{}
	SetProvider(provider)
	ctx := context.Background()

	span, ok := SpanFromContext(ctx)
	if !ok {
		t.Errorf("expected ok to be true, got false")
	}
	if span == nil {
		t.Errorf("expected span to be non-nil")
	}
}
