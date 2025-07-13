package datadog

import (
	"context"
	"os"
	"testing"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/mocktracer"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/DataDog/dd-trace-go/v2/profiler"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestStartTracing(t *testing.T) {
	os.Setenv("DD_ENABLED", "127.0.0.1")
	StartTracing(tracer.WithAgentAddr("127.0.0.1"))
	assert.True(t, Enabled())
	StopTracing()
	assert.False(t, Enabled())
}

func TestStartTracingMock(t *testing.T) {
	os.Setenv("DD_ENABLED", "")
	StartTracing(tracer.WithAgentAddr("127.0.0.1"))
	assert.True(t, Enabled())
	StopTracing()
	assert.False(t, Enabled())
}

func TestStartProfiling(t *testing.T) {
	os.Setenv("DD_ENABLED", "127.0.0.1")
	StartTracing(tracer.WithAgentAddr("127.0.0.1"))
	err := StartProfiling(profiler.WithHostname("127.0.0.1"))
	assert.NoError(t, err)
	StopProfiling()
}

func TestStartProfilingMock(t *testing.T) {
	os.Setenv("DD_ENABLED", "")
	StartTracing(tracer.WithAgentAddr("127.0.0.1"))
	err := StartProfiling(profiler.WithHostname("127.0.0.1"))
	assert.NoError(t, err)
	StopProfiling()
}

func TestExtractContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "x-datadog-trace-id", uint64(12345))
	ctx = context.WithValue(ctx, "x-datadog-span-id", uint64(67890))
	sc, err := ExtractContext(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, sc)
}

func TestExtractContextFromParent(t *testing.T) {
	sc, err := ExtractContextFromParent(12345, 67890)
	assert.NoError(t, err)
	assert.NotNil(t, sc)
}

func TestStartMainReqSpan(t *testing.T) {
	req := &fasthttp.Request{}
	req.Header.SetMethod("GET")
	req.SetRequestURI("http://example.com/test")
	span := StartMainReqSpan(req)
	assert.NotNil(t, span)
	span.Finish()
}

func TestStartMainQueueSpan(t *testing.T) {
	span := StartMainQueueSpan("test-queue")
	assert.NotNil(t, span)
	span.Finish()
}

func TestSpanChildFromParentTraceId(t *testing.T) {
	span, err := SpanChildFromParentTraceId(12345, "test-resource", "test-span")
	assert.NoError(t, err)
	assert.NotNil(t, span)
	span.Finish()
}

func TestSpanChildFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "x-datadog-trace-id", uint64(12345))
	ctx = context.WithValue(ctx, "x-datadog-span-id", uint64(67890))
	span, err := SpanChildFromContext(ctx, "test-resource", "test-span")
	assert.NoError(t, err)
	assert.NotNil(t, span)
	span.Finish()
}

func TestStartSpanFromContext(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	span, _ := StartSpanFromContext(context.Background(), "test-resource", "test-span")
	assert.NotNil(t, span)
}

func TestSpanFromContext(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	span := tracer.StartSpan("test")
	ctx := tracer.ContextWithSpan(context.Background(), span)
	retrievedSpan, err := SpanFromContext(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedSpan)
	span.Finish()
}

func TestSpanChild(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	span := tracer.StartSpan("test")
	childSpan := spanChild(span.Context(), span.Context().TraceIDLower(), "child-resource", "child-span")
	assert.NotNil(t, childSpan)
	childSpan.Finish()
}

func TestSpanFinish(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	span := tracer.StartSpan("test")
	wrappedSpan := &Span{
		originalSpan: span,
		spanType:     SpanTypeReq,
		traceId:      span.Context().TraceIDLower(),
	}
	wrappedSpan.Finish()
}
func TestSpanContext(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	span := tracer.StartSpan("test")
	wrappedSpan := &Span{
		originalSpan: span,
		spanType:     SpanTypeReq,
		traceId:      span.Context().TraceIDLower(),
	}

	ctx := context.Background()
	spanContext := wrappedSpan.Context(ctx)
	assert.NotNil(t, spanContext)
	assert.Equal(t, span.Context(), spanContext)
	span.Finish()
}
func TestSpanTraceID(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	span := tracer.StartSpan("test")
	wrappedSpan := &Span{
		originalSpan: span,
		spanType:     SpanTypeReq,
		traceId:      span.Context().TraceIDLower(),
	}

	traceID := wrappedSpan.TraceID()
	assert.Equal(t, span.Context().TraceIDLower(), traceID)
	span.Finish()
}
func TestSpan_SpanChild(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	parentSpan := tracer.StartSpan("parent")
	wrappedParentSpan := &Span{
		originalSpan: parentSpan,
		spanType:     SpanTypeReq,
		traceId:      parentSpan.Context().TraceIDLower(),
	}

	childSpan := wrappedParentSpan.SpanChild("child-resource", "child-span")
	assert.NotNil(t, childSpan)
	assert.Equal(t, "child-span", childSpan.spanType)
	childSpan.Finish()
	parentSpan.Finish()
}
func TestStartProfilingNotStarted(t *testing.T) {
	// Reset state
	started = false

	// Test when tracing not started
	err := StartProfiling()
	assert.NoError(t, err)
}

func TestStartProfilingWithOptions(t *testing.T) {
	// Setup environment
	os.Setenv("DD_ENABLED", "true")
	os.Setenv("DD_ENV", "test")
	os.Setenv("DD_SERVICE", "test-service")
	os.Setenv("DD_VERSION", "1.0")

	// Start tracing first
	StartTracing()

	// Test profiling with options
	opts := []profiler.Option{
		profiler.WithHostname("localhost"),
		profiler.WithEnv(os.Getenv("DD_ENV")),
		profiler.WithService(os.Getenv("DD_SERVICE")),
		profiler.WithProfileTypes(profiler.CPUProfile, profiler.MutexProfile, profiler.BlockProfile),
		profiler.WithLogStartup(false),
	}

	err := StartProfiling(opts...)
	assert.NoError(t, err)

	// Cleanup
	StopProfiling()
	StopTracing()

	os.Unsetenv("DD_ENABLED")
	os.Unsetenv("DD_ENV")
	os.Unsetenv("DD_SERVICE")
	os.Unsetenv("DD_VERSION")
}
