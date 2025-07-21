package hooks

import (
	"net/http"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

func TestBaseHook(t *testing.T) {
	events := []interfaces.HookEvent{interfaces.HookEventRequestStart, interfaces.HookEventRequestEnd}
	hook := NewBaseHook("test", events, 10)

	if hook.Name() != "test" {
		t.Fatalf("Expected name 'test', got '%s'", hook.Name())
	}

	if hook.Priority() != 10 {
		t.Fatalf("Expected priority 10, got %d", hook.Priority())
	}

	if !hook.IsEnabled() {
		t.Fatal("Expected hook to be enabled by default")
	}

	if len(hook.Events()) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(hook.Events()))
	}

	// Test enable/disable
	hook.SetEnabled(false)
	if hook.IsEnabled() {
		t.Fatal("Expected hook to be disabled")
	}

	hook.SetEnabled(true)
	if !hook.IsEnabled() {
		t.Fatal("Expected hook to be enabled")
	}

	// Test ShouldExecute
	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test",
		Timestamp:  time.Now(),
	}

	if !hook.ShouldExecute(ctx) {
		t.Fatal("Expected hook to should execute when enabled")
	}

	hook.SetEnabled(false)
	if hook.ShouldExecute(ctx) {
		t.Fatal("Expected hook to not execute when disabled")
	}

	// Test Execute (base implementation should do nothing)
	err := hook.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error from base execute, got %v", err)
	}
}

func TestConditionalBaseHook(t *testing.T) {
	events := []interfaces.HookEvent{interfaces.HookEventRequestStart}

	// Create condition that only allows GET requests
	condition := func(ctx *interfaces.HookContext) bool {
		return ctx.Request != nil && ctx.Request.Method == "GET"
	}

	hook := NewConditionalBaseHook("conditional", events, 10, condition)

	if hook.Name() != "conditional" {
		t.Fatalf("Expected name 'conditional', got '%s'", hook.Name())
	}

	if hook.Condition() == nil {
		t.Fatal("Expected condition function to be set")
	}

	// Test with GET request
	req, _ := http.NewRequest("GET", "/test", nil)
	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test",
		Timestamp:  time.Now(),
		Request:    req,
	}

	if !hook.ShouldExecute(ctx) {
		t.Fatal("Expected hook to execute for GET request")
	}

	// Test with POST request
	req, _ = http.NewRequest("POST", "/test", nil)
	ctx.Request = req

	if hook.ShouldExecute(ctx) {
		t.Fatal("Expected hook to not execute for POST request")
	}

	// Test when hook is disabled
	hook.SetEnabled(false)
	req, _ = http.NewRequest("GET", "/test", nil)
	ctx.Request = req

	if hook.ShouldExecute(ctx) {
		t.Fatal("Expected hook to not execute when disabled")
	}
}

func TestFilteredBaseHook(t *testing.T) {
	events := []interfaces.HookEvent{interfaces.HookEventRequestStart}
	hook := NewFilteredBaseHook("filtered", events, 10)

	if hook.Name() != "filtered" {
		t.Fatalf("Expected name 'filtered', got '%s'", hook.Name())
	}

	// Test with no filters (should execute)
	req, _ := http.NewRequest("GET", "/api/users", nil)
	req.Header.Set("Content-Type", "application/json")
	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test",
		Timestamp:  time.Now(),
		Request:    req,
	}

	if !hook.ShouldExecute(ctx) {
		t.Fatal("Expected hook to execute with no filters")
	}

	// Test path filter
	pathFilter := func(path string) bool {
		return path != "/api/users"
	}
	hook.SetPathFilter(pathFilter)

	if hook.ShouldExecute(ctx) {
		t.Fatal("Expected hook to not execute with path filter")
	}

	req.URL.Path = "/api/posts"
	if !hook.ShouldExecute(ctx) {
		t.Fatal("Expected hook to execute with different path")
	}

	// Test method filter
	methodFilter := func(method string) bool {
		return method != "GET"
	}
	hook.SetMethodFilter(methodFilter)

	if hook.ShouldExecute(ctx) {
		t.Fatal("Expected hook to not execute with method filter")
	}

	req.Method = "POST"
	if !hook.ShouldExecute(ctx) {
		t.Fatal("Expected hook to execute with different method")
	}

	// Test header filter
	headerFilter := func(headers http.Header) bool {
		return headers.Get("Content-Type") != "application/json"
	}
	hook.SetHeaderFilter(headerFilter)

	if hook.ShouldExecute(ctx) {
		t.Fatal("Expected hook to not execute with header filter")
	}

	req.Header.Set("Content-Type", "application/xml")
	if !hook.ShouldExecute(ctx) {
		t.Fatal("Expected hook to execute with different header")
	}

	// Test filter accessors
	if hook.PathFilter() == nil {
		t.Fatal("Expected path filter to be set")
	}
	if hook.MethodFilter() == nil {
		t.Fatal("Expected method filter to be set")
	}
	if hook.HeaderFilter() == nil {
		t.Fatal("Expected header filter to be set")
	}
}

func TestAsyncBaseHook(t *testing.T) {
	events := []interfaces.HookEvent{interfaces.HookEventRequestStart}
	hook := NewAsyncBaseHook("async", events, 10, 5, 10*time.Second)

	if hook.Name() != "async" {
		t.Fatalf("Expected name 'async', got '%s'", hook.Name())
	}

	if hook.BufferSize() != 5 {
		t.Fatalf("Expected buffer size 5, got %d", hook.BufferSize())
	}

	if hook.Timeout() != 10*time.Second {
		t.Fatalf("Expected timeout 10s, got %v", hook.Timeout())
	}

	// Test default values
	hook2 := NewAsyncBaseHook("async2", events, 10, 0, 0)
	if hook2.BufferSize() != 10 {
		t.Fatalf("Expected default buffer size 10, got %d", hook2.BufferSize())
	}
	if hook2.Timeout() != 30*time.Second {
		t.Fatalf("Expected default timeout 30s, got %v", hook2.Timeout())
	}

	// Test async execution
	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test",
		Timestamp:  time.Now(),
	}

	errChan := hook.ExecuteAsync(ctx)
	if errChan == nil {
		t.Fatal("Expected error channel to be returned")
	}

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("Expected no error from async execution, got %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected async execution to complete")
	}
}

func TestHookChain(t *testing.T) {
	chain := NewHookChain()

	if chain == nil {
		t.Fatal("Expected chain to be created")
	}

	executed := []string{}

	hook1 := &testHook{
		BaseHook: NewBaseHook("hook1", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1),
		execute: func(ctx *interfaces.HookContext) error {
			executed = append(executed, "hook1")
			return nil
		},
	}

	hook2 := &testHook{
		BaseHook: NewBaseHook("hook2", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 2),
		execute: func(ctx *interfaces.HookContext) error {
			executed = append(executed, "hook2")
			return nil
		},
	}

	// Test Add method
	chain.Add(hook1).Add(hook2)

	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test",
		Timestamp:  time.Now(),
	}

	// Test Execute
	err := chain.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(executed) != 2 {
		t.Fatalf("Expected 2 executions, got %d", len(executed))
	}

	if executed[0] != "hook1" || executed[1] != "hook2" {
		t.Fatalf("Expected execution order [hook1, hook2], got %v", executed)
	}

	// Test ExecuteUntil
	executed = []string{}
	condition := func(ctx *interfaces.HookContext) bool {
		return len(executed) >= 1
	}

	err = chain.ExecuteUntil(ctx, condition)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(executed) != 1 {
		t.Fatalf("Expected 1 execution with ExecuteUntil, got %d", len(executed))
	}

	// Test ExecuteIf
	executed = []string{}
	condition = func(ctx *interfaces.HookContext) bool {
		return true
	}

	err = chain.ExecuteIf(ctx, condition)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(executed) != 2 {
		t.Fatalf("Expected 2 executions with ExecuteIf(true), got %d", len(executed))
	}

	// Test ExecuteIf with false condition
	executed = []string{}
	condition = func(ctx *interfaces.HookContext) bool {
		return false
	}

	err = chain.ExecuteIf(ctx, condition)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(executed) != 0 {
		t.Fatalf("Expected 0 executions with ExecuteIf(false), got %d", len(executed))
	}
}

func TestPathFilterBuilder(t *testing.T) {
	builder := NewPathFilterBuilder()

	// Test include paths
	filter := builder.Include("/api/users", "/api/posts").Build()

	if !filter("/api/users") {
		t.Fatal("Expected /api/users to be allowed")
	}
	if !filter("/api/posts") {
		t.Fatal("Expected /api/posts to be allowed")
	}
	if filter("/api/comments") {
		t.Fatal("Expected /api/comments to be denied")
	}

	// Test exclude paths
	builder = NewPathFilterBuilder()
	filter = builder.Exclude("/api/private").Build()

	if filter("/api/private") {
		t.Fatal("Expected /api/private to be denied")
	}
	if !filter("/api/public") {
		t.Fatal("Expected /api/public to be allowed")
	}

	// Test prefixes
	builder = NewPathFilterBuilder()
	filter = builder.WithPrefix("/admin").Build()

	if filter("/admin/users") {
		t.Fatal("Expected /admin/users to be denied")
	}
	if !filter("/api/users") {
		t.Fatal("Expected /api/users to be allowed")
	}

	// Test suffixes
	builder = NewPathFilterBuilder()
	filter = builder.WithSuffix(".html").Build()

	if filter("/page.html") {
		t.Fatal("Expected /page.html to be denied")
	}
	if !filter("/api/data") {
		t.Fatal("Expected /api/data to be allowed")
	}

	// Test combined filters
	builder = NewPathFilterBuilder()
	filter = builder.Include("/api/users").Exclude("/api/private").Build()

	if !filter("/api/users") {
		t.Fatal("Expected /api/users to be allowed")
	}
	if filter("/api/private") {
		t.Fatal("Expected /api/private to be denied")
	}
	if filter("/api/posts") {
		t.Fatal("Expected /api/posts to be denied (not in include list)")
	}
}

func TestMethodFilterBuilder(t *testing.T) {
	builder := NewMethodFilterBuilder()

	// Test allow methods
	filter := builder.Allow("GET", "POST").Build()

	if !filter("GET") {
		t.Fatal("Expected GET to be allowed")
	}
	if !filter("get") { // Test case insensitive
		t.Fatal("Expected get to be allowed (case insensitive)")
	}
	if !filter("POST") {
		t.Fatal("Expected POST to be allowed")
	}
	if filter("DELETE") {
		t.Fatal("Expected DELETE to be denied")
	}

	// Test deny methods
	builder = NewMethodFilterBuilder()
	filter = builder.Deny("DELETE", "PUT").Build()

	if filter("DELETE") {
		t.Fatal("Expected DELETE to be denied")
	}
	if filter("delete") { // Test case insensitive
		t.Fatal("Expected delete to be denied (case insensitive)")
	}
	if !filter("GET") {
		t.Fatal("Expected GET to be allowed")
	}

	// Test combined filters
	builder = NewMethodFilterBuilder()
	filter = builder.Allow("GET", "POST").Deny("DELETE").Build()

	if !filter("GET") {
		t.Fatal("Expected GET to be allowed")
	}
	if !filter("POST") {
		t.Fatal("Expected POST to be allowed")
	}
	if filter("DELETE") {
		t.Fatal("Expected DELETE to be denied")
	}
	if filter("PUT") {
		t.Fatal("Expected PUT to be denied (not in allow list)")
	}
}

func TestFilteredBaseHook_DisabledFiltering(t *testing.T) {
	events := []interfaces.HookEvent{interfaces.HookEventRequestStart}
	hook := NewFilteredBaseHook("filtered", events, 10)

	// Disable the hook
	hook.SetEnabled(false)

	// Even with a request that would pass filters, disabled hook shouldn't execute
	req, _ := http.NewRequest("GET", "/api/users", nil)
	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test",
		Timestamp:  time.Now(),
		Request:    req,
	}

	if hook.ShouldExecute(ctx) {
		t.Fatal("Expected disabled hook to not execute regardless of filters")
	}
}

func TestConditionalBaseHook_DisabledCondition(t *testing.T) {
	events := []interfaces.HookEvent{interfaces.HookEventRequestStart}

	// Condition that always returns true
	condition := func(ctx *interfaces.HookContext) bool {
		return true
	}

	hook := NewConditionalBaseHook("conditional", events, 10, condition)

	// Disable the hook
	hook.SetEnabled(false)

	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test",
		Timestamp:  time.Now(),
	}

	if hook.ShouldExecute(ctx) {
		t.Fatal("Expected disabled hook to not execute regardless of condition")
	}
}

func TestHookChain_SkippedHooks(t *testing.T) {
	chain := NewHookChain()

	executed := []string{}

	hook1 := &testHook{
		BaseHook: NewBaseHook("hook1", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 1),
		execute: func(ctx *interfaces.HookContext) error {
			executed = append(executed, "hook1")
			return nil
		},
	}

	hook2 := &testHook{
		BaseHook: NewBaseHook("hook2", []interfaces.HookEvent{interfaces.HookEventRequestStart}, 2),
		execute: func(ctx *interfaces.HookContext) error {
			executed = append(executed, "hook2")
			return nil
		},
	}

	// Disable hook2
	hook2.SetEnabled(false)

	chain.Add(hook1).Add(hook2)

	ctx := &interfaces.HookContext{
		Event:      interfaces.HookEventRequestStart,
		ServerName: "test",
		Timestamp:  time.Now(),
	}

	err := chain.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Only hook1 should have executed
	if len(executed) != 1 {
		t.Fatalf("Expected 1 execution, got %d", len(executed))
	}

	if executed[0] != "hook1" {
		t.Fatalf("Expected hook1 to execute, got %s", executed[0])
	}
}
