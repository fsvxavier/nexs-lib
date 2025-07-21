// Package hooks provides observer implementations for HTTP server lifecycle events.
package hooks

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// TracingObserver implements ServerObserver for distributed tracing.
type TracingObserver struct {
	serviceName string
}

// NewTracingObserver creates a new tracing observer.
func NewTracingObserver(serviceName string) *TracingObserver {
	if serviceName == "" {
		serviceName = "httpserver"
	}
	return &TracingObserver{
		serviceName: serviceName,
	}
}

// OnStart traces when the server starts.
func (t *TracingObserver) OnStart(name string) {
	fmt.Printf("TRACE: Server %s started\n", name)
}

// OnStop traces when the server stops.
func (t *TracingObserver) OnStop(name string) {
	fmt.Printf("TRACE: Server %s stopped\n", name)
}

// OnRequest traces each HTTP request.
func (t *TracingObserver) OnRequest(name string, req *http.Request, status int, duration time.Duration) {
	fmt.Printf("TRACE: Server %s - %s %s - Status: %d - Duration: %v\n",
		name, req.Method, req.URL.Path, status, duration)
}

// OnBeforeRequest traces before processing a request.
func (t *TracingObserver) OnBeforeRequest(name string, req *http.Request) {
	fmt.Printf("TRACE: Server %s - Starting %s %s\n", name, req.Method, req.URL.Path)
}

// OnAfterRequest traces after processing a request.
func (t *TracingObserver) OnAfterRequest(name string, req *http.Request, status int, duration time.Duration) {
	fmt.Printf("TRACE: Server %s - Finished %s %s - Status: %d - Duration: %v\n",
		name, req.Method, req.URL.Path, status, duration)
}

// HookObserver implementation for generic hooks

// OnHookStart is called when a hook starts executing.
func (t *TracingObserver) OnHookStart(name string, event interfaces.HookEvent, ctx *interfaces.HookContext) {
	fmt.Printf("TRACE: Hook %s started for event %s on server %s\n", name, event, ctx.ServerName)
}

// OnHookEnd is called when a hook finishes executing.
func (t *TracingObserver) OnHookEnd(name string, event interfaces.HookEvent, ctx *interfaces.HookContext, err error, duration time.Duration) {
	status := "success"
	if err != nil {
		status = "error"
	}
	fmt.Printf("TRACE: Hook %s finished for event %s on server %s - Status: %s - Duration: %v\n",
		name, event, ctx.ServerName, status, duration)
}

// OnHookError is called when a hook encounters an error.
func (t *TracingObserver) OnHookError(name string, event interfaces.HookEvent, ctx *interfaces.HookContext, err error) {
	fmt.Printf("TRACE: Hook %s error for event %s on server %s - Error: %v\n",
		name, event, ctx.ServerName, err)
}

// OnHookSkip is called when a hook is skipped.
func (t *TracingObserver) OnHookSkip(name string, event interfaces.HookEvent, ctx *interfaces.HookContext, reason string) {
	fmt.Printf("TRACE: Hook %s skipped for event %s on server %s - Reason: %s\n",
		name, event, ctx.ServerName, reason)
}

// Verify TracingObserver implements both ServerObserver and HookObserver.
var _ interfaces.ServerObserver = (*TracingObserver)(nil)
var _ interfaces.HookObserver = (*TracingObserver)(nil)
