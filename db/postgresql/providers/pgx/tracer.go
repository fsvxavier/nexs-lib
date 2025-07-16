package pgx

import (
	"context"
	"fmt"
	"time"

	interfaces "github.com/fsvxavier/nexs-lib/db/postgresql/interface"
	"github.com/jackc/pgx/v5"
)

// PGXTracer implements pgx tracing for detailed observability
type PGXTracer struct {
	hookManager interfaces.HookManager
	enabled     bool
}

// NewPGXTracer creates a new PGX tracer
func NewPGXTracer(hookManager interfaces.HookManager) *PGXTracer {
	return &PGXTracer{
		hookManager: hookManager,
		enabled:     true,
	}
}

// TraceQueryStart is called when a query starts
func (t *PGXTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if !t.enabled {
		return ctx
	}

	execCtx := &interfaces.ExecutionContext{
		Context:   ctx,
		Operation: "query",
		Query:     data.SQL,
		Args:      data.Args,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	execCtx.Metadata["trace_query_start"] = true
	execCtx.Metadata["connection_id"] = getConnectionID(conn)

	// Execute before query hooks
	if t.hookManager != nil {
		_ = t.hookManager.ExecuteHooks(interfaces.BeforeQueryHook, execCtx)
	}

	// Store execution context in the request context
	return context.WithValue(ctx, "execution_context", execCtx)
}

// TraceQueryEnd is called when a query ends
func (t *PGXTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	if !t.enabled {
		return
	}

	// Retrieve execution context from request context
	execCtxVal := ctx.Value("execution_context")
	if execCtxVal == nil {
		return
	}

	execCtx, ok := execCtxVal.(*interfaces.ExecutionContext)
	if !ok {
		return
	}

	// Update execution context with results
	execCtx.Duration = time.Since(execCtx.StartTime)
	execCtx.Error = data.Err

	if data.CommandTag.String() != "" {
		execCtx.RowsAffected = data.CommandTag.RowsAffected()
		execCtx.Metadata["command_tag"] = data.CommandTag.String()
	}

	// Execute appropriate hooks
	if t.hookManager != nil {
		if data.Err != nil {
			execCtx.Metadata["trace_query_error"] = true
			_ = t.hookManager.ExecuteHooks(interfaces.OnErrorHook, execCtx)
		} else {
			execCtx.Metadata["trace_query_success"] = true
			_ = t.hookManager.ExecuteHooks(interfaces.AfterQueryHook, execCtx)
		}
	}
}

// TraceBatchStart is called when a batch starts
func (t *PGXTracer) TraceBatchStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchStartData) context.Context {
	if !t.enabled {
		return ctx
	}

	execCtx := &interfaces.ExecutionContext{
		Context:   ctx,
		Operation: "batch",
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	execCtx.Metadata["trace_batch_start"] = true
	execCtx.Metadata["connection_id"] = getConnectionID(conn)
	execCtx.Metadata["batch_size"] = data.Batch.Len()

	// Execute before batch hooks
	if t.hookManager != nil {
		_ = t.hookManager.ExecuteHooks(interfaces.BeforeBatchHook, execCtx)
	}

	return context.WithValue(ctx, "batch_execution_context", execCtx)
}

// TraceBatchQuery is called for each query in a batch
func (t *PGXTracer) TraceBatchQuery(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchQueryData) {
	if !t.enabled {
		return
	}

	// Individual query tracing within batch can be implemented here
	// For now, we'll just track it in metadata
	execCtxVal := ctx.Value("batch_execution_context")
	if execCtxVal != nil {
		if execCtx, ok := execCtxVal.(*interfaces.ExecutionContext); ok {
			if execCtx.Metadata["batch_queries"] == nil {
				execCtx.Metadata["batch_queries"] = make([]string, 0)
			}
			queries := execCtx.Metadata["batch_queries"].([]string)
			execCtx.Metadata["batch_queries"] = append(queries, data.SQL)
		}
	}
}

// TraceBatchEnd is called when a batch ends
func (t *PGXTracer) TraceBatchEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchEndData) {
	if !t.enabled {
		return
	}

	// Retrieve execution context from request context
	execCtxVal := ctx.Value("batch_execution_context")
	if execCtxVal == nil {
		return
	}

	execCtx, ok := execCtxVal.(*interfaces.ExecutionContext)
	if !ok {
		return
	}

	// Update execution context with results
	execCtx.Duration = time.Since(execCtx.StartTime)
	execCtx.Error = data.Err

	// Execute appropriate hooks
	if t.hookManager != nil {
		if data.Err != nil {
			execCtx.Metadata["trace_batch_error"] = true
			_ = t.hookManager.ExecuteHooks(interfaces.OnErrorHook, execCtx)
		} else {
			execCtx.Metadata["trace_batch_success"] = true
			_ = t.hookManager.ExecuteHooks(interfaces.AfterBatchHook, execCtx)
		}
	}
}

// TraceConnectStart is called when a connection starts
func (t *PGXTracer) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context {
	if !t.enabled {
		return ctx
	}

	execCtx := &interfaces.ExecutionContext{
		Context:   ctx,
		Operation: "connect",
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	execCtx.Metadata["trace_connect_start"] = true
	execCtx.Metadata["connect_config"] = data.ConnConfig

	// Execute before connection hooks
	if t.hookManager != nil {
		_ = t.hookManager.ExecuteHooks(interfaces.BeforeConnectionHook, execCtx)
	}

	return context.WithValue(ctx, "connect_execution_context", execCtx)
}

// TraceConnectEnd is called when a connection ends
func (t *PGXTracer) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
	if !t.enabled {
		return
	}

	// Retrieve execution context from request context
	execCtxVal := ctx.Value("connect_execution_context")
	if execCtxVal == nil {
		return
	}

	execCtx, ok := execCtxVal.(*interfaces.ExecutionContext)
	if !ok {
		return
	}

	// Update execution context with results
	execCtx.Duration = time.Since(execCtx.StartTime)
	execCtx.Error = data.Err

	if data.Conn != nil {
		execCtx.Metadata["connection_id"] = getConnectionID(data.Conn)
		execCtx.Metadata["connection_pid"] = data.Conn.PgConn().PID()
	}

	// Execute appropriate hooks
	if t.hookManager != nil {
		if data.Err != nil {
			execCtx.Metadata["trace_connect_error"] = true
			_ = t.hookManager.ExecuteHooks(interfaces.OnErrorHook, execCtx)
		} else {
			execCtx.Metadata["trace_connect_success"] = true
			_ = t.hookManager.ExecuteHooks(interfaces.AfterConnectionHook, execCtx)
		}
	}
}

// TracePrepareStart is called when statement preparation starts
func (t *PGXTracer) TracePrepareStart(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareStartData) context.Context {
	if !t.enabled {
		return ctx
	}

	execCtx := &interfaces.ExecutionContext{
		Context:   ctx,
		Operation: "prepare",
		Query:     data.SQL,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	execCtx.Metadata["trace_prepare_start"] = true
	execCtx.Metadata["connection_id"] = getConnectionID(conn)
	execCtx.Metadata["statement_name"] = data.Name

	return context.WithValue(ctx, "prepare_execution_context", execCtx)
}

// TracePrepareEnd is called when statement preparation ends
func (t *PGXTracer) TracePrepareEnd(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareEndData) {
	if !t.enabled {
		return
	}

	// Retrieve execution context from request context
	execCtxVal := ctx.Value("prepare_execution_context")
	if execCtxVal == nil {
		return
	}

	execCtx, ok := execCtxVal.(*interfaces.ExecutionContext)
	if !ok {
		return
	}

	// Update execution context with results
	execCtx.Duration = time.Since(execCtx.StartTime)
	execCtx.Error = data.Err

	execCtx.Metadata["trace_prepare_end"] = true
	if data.AlreadyPrepared {
		execCtx.Metadata["already_prepared"] = true
	}
}

// SetEnabled enables or disables tracing
func (t *PGXTracer) SetEnabled(enabled bool) {
	t.enabled = enabled
}

// IsEnabled returns whether tracing is enabled
func (t *PGXTracer) IsEnabled() bool {
	return t.enabled
}

// Helper function to get connection ID
func getConnectionID(conn *pgx.Conn) string {
	if conn == nil || conn.PgConn() == nil {
		return "unknown"
	}
	// Use PID as connection identifier since RemoteAddr is not available
	return fmt.Sprintf("conn_%d", conn.PgConn().PID())
}
