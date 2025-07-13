package gpgx

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/ext"
	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	jsoniter "github.com/json-iterator/go"
)

const (
	spanOperation  = "pgx.query"
	queryTypeQuery = "Query"
	tagComponent   = "jackc/pgxpool"
)

type tracedBatchQuery struct {
	span *tracer.Span
	data pgx.TraceBatchQueryData
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// TracerConfig contains configs to tracing and would implement QueryTracer, BatchTracer,
// CopyFromTracer, PrepareTracer and ConnectTracer from pgx.
type TracerConfig struct {
	prevBatchQuery       *tracedBatchQuery
	spanMap              map[string]*tracer.Span
	ObservabilityEnabled bool
	QueryTracerEnabled   bool
	Service              string
}

type spanMapKeyType string

var (
	spanMapKey spanMapKeyType = "spanMapKey"
	mtx        sync.Mutex
)

type Option func(*TracerConfig)

func (tb *tracedBatchQuery) finish() {
	tb.span.Finish(tracer.WithError(tb.data.Err))
}

func (cfg *TracerConfig) TraceAcquireStart(ctx context.Context, pool *pgxpool.Pool, data pgxpool.TraceAcquireStartData) context.Context {
	if cfg.ObservabilityEnabled {
		poolStats := pool.Stat()
		opts := []tracer.StartSpanOption{
			tracer.SpanType(ext.SpanTypeSQL),
			tracer.Tag(ext.DBSystem, ext.DBSystemPostgreSQL),
			tracer.Tag(ext.Component, tagComponent),
			tracer.Tag(ext.SpanKind, ext.SpanKindClient),
			tracer.Tag(ext.DBName, pool.Config().ConnConfig.Database),
			tracer.Tag(ext.DBUser, pool.Config().ConnConfig.User),
			tracer.Tag("db.host", pool.Config().ConnConfig.Host),
			tracer.Tag("db.operation", "Acquire"),
			tracer.Tag("db.pool.snapshot.total_connections", poolStats.TotalConns()),
			tracer.Tag("db.pool.snapshot.constructing_connections", poolStats.ConstructingConns()),
			tracer.Tag("db.pool.snapshot.acquired_connections", poolStats.AcquiredConns()),
			tracer.Tag("db.pool.snapshot.max_connections", poolStats.MaxConns()),
			tracer.Tag("db.pool.snapshot.idle_connections", poolStats.IdleConns()),
			tracer.Tag("db.pool.snapshot.used_connections", poolStats.TotalConns()-poolStats.IdleConns()),
		}

		_, ddCtx := tracer.StartSpanFromContext(ctx, "Pgx.Pool.AcquireTracer", opts...)
		return ddCtx
	}

	return ctx
}

func (cfg *TracerConfig) TraceAcquireEnd(ctx context.Context, _ *pgxpool.Pool, data pgxpool.TraceAcquireEndData) {
	span, ok := tracer.SpanFromContext(ctx)
	if !ok {
		return
	}

	span.Finish(tracer.WithError(data.Err))
}

// TraceQueryStart is called at the beginning of Query, QueryRow, and Exec calls. The returned context is used for the
// rest of the call and will be passed to TraceQueryEnd.
func (cfg *TracerConfig) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if cfg.QueryTracerEnabled {
		dataArgs, err := json.Marshal(data.Args)
		if err != nil {
			fmt.Println("Error marshal query args")
		}
		fmt.Printf("\n\nQUERY: %s - ARGS: - %s\n\n", data.SQL, string(dataArgs))
	}

	if cfg.ObservabilityEnabled {
		opts := []tracer.StartSpanOption{
			tracer.ServiceName(cfg.Service),
			tracer.SpanType(ext.SpanTypeSQL),
			tracer.StartTime(time.Now()),
			tracer.Tag(ext.SpanKind, ext.SpanKindClient),
			tracer.Tag(ext.DBSystem, ext.DBSystemPostgreSQL),
			tracer.Tag(ext.DBUser, conn.Config().User),
			tracer.Tag(ext.Component, tagComponent),
			tracer.Tag(ext.ResourceName, data.SQL),
			tracer.Tag(ext.DBStatement, data.SQL),
			tracer.Tag(ext.DBName, conn.Config().Database),
			tracer.Tag("db.host", conn.Config().Host),
			tracer.Tag("sql.query_type", queryTypeQuery),
			tracer.Tag("dd.env", os.Getenv("DD_ENV")),
			tracer.Tag("dd.version", os.Getenv("DD_VERSION")),
		}

		sn := os.Getenv("DD_SERVICE_DB")
		if sn == "" {
			sn = os.Getenv("DD_SERVICE_DB")
		}
		if sn != "" {
			opts = append(opts, tracer.ServiceName(sn))
		} else {
			opts = append(opts, tracer.ServiceName(os.Getenv("DD_SERVICE")+".db"))
		}

		span, _ := tracer.StartSpanFromContext(ctx, spanOperation, opts...)
		mtx.Lock()
		defer mtx.Unlock()
		if cfg.spanMap == nil {
			cfg.spanMap = make(map[string]*tracer.Span)
		}

		uuidKey := uuid.New().String()
		cfg.spanMap[uuidKey] = span

		return context.WithValue(ctx, spanMapKey, uuidKey)
	}

	return ctx
}

// TraceQueryEnd traces the end of the query, implementing pgx.QueryTracer.
func (cfg *TracerConfig) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	uuidKey := ctx.Value(spanMapKey)
	if uuidKey != nil {
		if k, ok := uuidKey.(string); ok {
			mtx.Lock()
			defer mtx.Unlock()
			span := cfg.spanMap[k]
			(*span).Finish()

			delete(cfg.spanMap, k)
		}
	}
}

func (cfg *TracerConfig) TraceBatchStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchStartData) context.Context {
	opts := []tracer.StartSpanOption{
		tracer.ServiceName(cfg.Service),
		tracer.SpanType(ext.SpanTypeSQL),
		tracer.StartTime(time.Now()),
		tracer.Tag(ext.Component, tagComponent),
		tracer.Tag(ext.ResourceName, "batch.inserts"),
		tracer.Tag(ext.SpanKind, ext.SpanKindClient),
		tracer.Tag(ext.DBSystem, ext.DBSystemPostgreSQL),
		tracer.Tag(ext.DBUser, conn.Config().User),
		tracer.Tag(ext.DBName, conn.Config().Database),
		tracer.Tag("db.host", conn.Config().Host),
		tracer.Tag("db.batch.num_queries", data.Batch.Len()),
	}
	_, ctx = tracer.StartSpanFromContext(ctx, "db.batch", opts...)

	return ctx
}

func (cfg *TracerConfig) TraceBatchQuery(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchQueryData) {
	// Finish the previous batch query span before starting the next one, since pgx doesn't provide hooks or timestamp
	// information about when the actual operation started or finished.
	// pgx.Batch* types don't support concurrency. This function doesn't support it either.
	if cfg.prevBatchQuery != nil {
		cfg.prevBatchQuery.finish()
	}

	opts := []tracer.StartSpanOption{
		tracer.ServiceName(cfg.Service),
		tracer.SpanType(ext.SpanTypeSQL),
		tracer.StartTime(time.Now()),
		tracer.Tag(ext.Component, tagComponent),
		tracer.Tag(ext.ResourceName, data.SQL),
		tracer.Tag(ext.SpanKind, ext.SpanKindClient),
		tracer.Tag(ext.DBSystem, ext.DBSystemPostgreSQL),
		tracer.Tag(ext.DBUser, conn.Config().User),
		tracer.Tag(ext.DBStatement, data.SQL),
		tracer.Tag(ext.DBName, conn.Config().Database),
		tracer.Tag("db.host", conn.Config().Host),
		tracer.Tag("db.batch.result.rows_affected", data.CommandTag.RowsAffected()),
	}

	span, _ := tracer.StartSpanFromContext(ctx, "db.batch.query", opts...)

	cfg.prevBatchQuery = &tracedBatchQuery{
		span: span,
		data: data,
	}
}

func (cfg *TracerConfig) TraceBatchEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchEndData) {
	if cfg.prevBatchQuery != nil {
		cfg.prevBatchQuery.finish()
		cfg.prevBatchQuery = nil
	}

	if span, ok := tracer.SpanFromContext(ctx); ok {
		if data.Err != nil {
			span.SetTag(ext.Error, data.Err)
		}

		span.Finish(tracer.WithError(data.Err))
	}
}
