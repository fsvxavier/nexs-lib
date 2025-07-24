package datadog

import (
	"context"
	"testing"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/mocktracer"
	"github.com/stretchr/testify/assert"
)

func initializeProvider() *Provider {
	return &Provider{
		service: "test-service",
		env:     "test-env",
		version: "1.0.0",
	}
}

func initializeContext() context.Context {
	return context.Background()
}

func TestProvider_StartSpanFromContext(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()
	provider := initializeProvider()
	ctx := initializeContext()
	ctxs, span := provider.StartSpanFromContext(ctx, "test-span")
	defer span.Finish()

	assert.NotNil(t, ctxs)
	assert.NotNil(t, span)

	span.Finish()
}

// func TestProvider_SpanFromContext(t *testing.T) {
// 	mt := mocktracer.Start()
// 	defer mt.Stop()
// 	provider := initializeProvider()
// 	ctx := initializeContext()

// 	retrievedSpan, found := provider.SpanFromContext(ctx)
// 	assert.True(t, found)
// 	assert.NotNil(t, retrievedSpan)
// }

func TestSetProvider(t *testing.T) {
	service := "test-service"
	env := "test-env"
	version := "1.0.0"

	setProvider(service, env, version)
}
