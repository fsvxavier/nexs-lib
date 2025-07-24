package fibertrace

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/ext"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"

	"github.com/DataDog/dd-trace-go/v2/instrumentation"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const componentName = "gofiber/fiber.v2"

var instr *instrumentation.Instrumentation

func init() {
	instr = instrumentation.Load(instrumentation.PackageGoFiberV2)
	tracer.MarkIntegrationImported("github.com/gofiber/fiber/v2")
}

// Middleware returns middleware that will trace incoming requests.
func Middleware(opts ...Option) func(ctx *fiber.Ctx) error {
	cfg := new(config)
	defaults(cfg)
	for _, fn := range opts {
		fn.apply(cfg)
	}
	instr.Logger().Debug("gofiber/fiber.v2: Middleware: %#v", cfg)
	return func(ctx *fiber.Ctx) error {

		if cfg.ignoreRequest(ctx) {
			return ctx.Next()
		}

		opts := []tracer.StartSpanOption{
			tracer.SpanType(ext.SpanTypeWeb),
			tracer.ServiceName(cfg.serviceName),
			tracer.Tag(ext.HTTPMethod, ctx.Method()),
			tracer.Tag(ext.HTTPURL, string(ctx.Request().URI().PathOriginal())),
			tracer.Measured(),
		}
		if !math.IsNaN(cfg.analyticsRate) {
			opts = append(opts, tracer.Tag(ext.EventSampleRate, cfg.analyticsRate))
		}
		// Create a http.Header object so that a parent trace can be extracted. Fiber uses a non-standard header carrier
		h := http.Header{}
		for k, headers := range ctx.GetReqHeaders() {
			for _, v := range headers {
				// GetReqHeaders returns a list of headers associated with the given key.
				// http.Header.Add supports appending multiple values, so the previous
				// value will not be overwritten.
				h.Add(k, v)
			}
		}

		if spanctx, err := tracer.Extract(tracer.HTTPHeadersCarrier(h)); err == nil {
			// If there are span links as a result of context extraction, add them as a StartSpanOption
			if spanctx != nil && spanctx.SpanLinks() != nil {
				opts = append(opts, tracer.WithSpanLinks(spanctx.SpanLinks()))
			}
			opts = append(opts, tracer.ChildOf(spanctx))
		}
		opts = append(opts, cfg.spanOpts...)
		opts = append(opts,
			tracer.Tag(ext.Component, componentName),
			tracer.Tag(ext.SpanKind, ext.SpanKindServer),
		)
		span, spanCtx := tracer.StartSpanFromContext(ctx.UserContext(), cfg.spanName, opts...)

		defer span.Finish()

		// pass the span through the request UserContext
		ctx.SetUserContext(spanCtx)

		start := time.Now()

		// pass the execution down the line
		err := ctx.Next()

		elapsed := time.Since(start)

		GetEnvAsDuration := func(key string, defaultValue time.Duration) time.Duration {
			value := os.Getenv(key)
			if value == "" {
				return time.Second * 5
			}
			duration, err := time.ParseDuration(value)
			if err != nil {
				return defaultValue
			}
			return duration
		}

		threshold := GetEnvAsDuration("DD_TRACE_KEEP_IF_EXCEEDS", time.Second)
		if elapsed > threshold {
			span.SetTag(ext.ManualKeep, true)
		}

		span.SetTag(ext.ResourceName, cfg.resourceNamer(ctx))
		span.SetTag(ext.HTTPRoute, ctx.Route().Path)
		span.SetTag("http.client_id", ctx.Get("client-id", ctx.Get("Client-Id", ctx.Get("client_id"))))
		span.SetTag("http.request", string(ctx.Body()))
		span.SetTag("http.response", string(ctx.Response().Body()))

		traceID, spanID := setTraceID(ctx, span)

		span.SetTag("trace_id", ctx.Get("uuid", uuid.NewString()))
		span.SetTag("dd.trace_id", traceID)
		span.SetTag("dd.span_id", spanID)

		status := ctx.Response().StatusCode()
		// on the off chance we don't yet have a status after the rest of the things have run
		if status == 0 {
			// 0 - means we do not have a status code at this point
			// in case the response was returned by a middleware without one
			status = http.StatusOK
		}
		span.SetTag(ext.HTTPCode, strconv.Itoa(status))

		if err != nil && cfg.isStatusError(status) {
			// mark 5xx server error
			span.SetTag(ext.Error, fmt.Errorf("%d: %s", status, http.StatusText(status)))
		}
		return err
	}
}

func setTraceID(ctx *fiber.Ctx, span *tracer.Span) (traceID, spanID string) {
	traceID = fmt.Sprintf("%s", span.Context().TraceID())
	ctx.Set("Trace-Id", traceID)
	ctx.Request().Header.Set("Trace-Id", traceID)
	ctx.Response().Header.Set("Trace-Id", traceID)

	spanID = fmt.Sprintf("%d", span.Context().SpanID())
	ctx.Set("Span-Id", spanID)
	ctx.Request().Header.Set("Span-Id", spanID)
	ctx.Response().Header.Set("Span-Id", spanID)

	return
}
