package middlewares

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

// TestLogger is a test logger that captures log messages for testing.
type TestLogger struct {
	InfoMessages  []string
	ErrorMessages []string
	DebugMessages []string
	WarnMessages  []string
}

func (l *TestLogger) Info(msg string, args ...interface{}) {
	l.InfoMessages = append(l.InfoMessages, msg)
}

func (l *TestLogger) Error(msg string, args ...interface{}) {
	l.ErrorMessages = append(l.ErrorMessages, msg)
}

func (l *TestLogger) Debug(msg string, args ...interface{}) {
	l.DebugMessages = append(l.DebugMessages, msg)
}

func (l *TestLogger) Warn(msg string, args ...interface{}) {
	l.WarnMessages = append(l.WarnMessages, msg)
}

// TestMiddleware is a simple middleware for testing.
type TestMiddleware struct {
	*BaseMiddleware
	processCount int
}

func NewTestMiddleware(name string, priority int) *TestMiddleware {
	return &TestMiddleware{
		BaseMiddleware: NewBaseMiddleware(name, priority),
	}
}

func (tm *TestMiddleware) Process(ctx context.Context, req interface{}, next MiddlewareNext) (interface{}, error) {
	tm.processCount++
	tm.GetLogger().Debug("TestMiddleware %s: Processing request (count: %d)", tm.Name(), tm.processCount)
	return next(ctx, req)
}

func (tm *TestMiddleware) GetProcessCount() int {
	return tm.processCount
}

// ErrorMiddleware is a middleware that returns an error for testing.
type ErrorMiddleware struct {
	*BaseMiddleware
	shouldError bool
}

func NewErrorMiddleware(name string, priority int) *ErrorMiddleware {
	return &ErrorMiddleware{
		BaseMiddleware: NewBaseMiddleware(name, priority),
		shouldError:    true,
	}
}

func (em *ErrorMiddleware) Process(ctx context.Context, req interface{}, next MiddlewareNext) (interface{}, error) {
	if em.shouldError {
		return nil, errors.New("middleware error")
	}
	return next(ctx, req)
}

func (em *ErrorMiddleware) SetShouldError(shouldError bool) {
	em.shouldError = shouldError
}

func TestNewMiddlewareManager(t *testing.T) {
	manager := NewMiddlewareManager()

	if manager == nil {
		t.Fatal("Expected middleware manager to be created")
	}

	if manager.middlewares == nil {
		t.Error("Expected middlewares slice to be initialized")
	}

	if manager.observerManager == nil {
		t.Error("Expected observer manager to be initialized")
	}

	if manager.logger == nil {
		t.Error("Expected logger to be initialized")
	}
}

func TestMiddlewareManager_AddMiddleware(t *testing.T) {
	manager := NewMiddlewareManager()
	middleware := NewTestMiddleware("test-middleware", 1)

	// Test successful addition
	err := manager.AddMiddleware(middleware)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if manager.GetMiddlewareCount() != 1 {
		t.Errorf("Expected 1 middleware, got %d", manager.GetMiddlewareCount())
	}

	// Test nil middleware addition
	err = manager.AddMiddleware(nil)
	if err == nil {
		t.Error("Expected error for nil middleware")
	}
}

func TestMiddlewareManager_RemoveMiddleware(t *testing.T) {
	manager := NewMiddlewareManager()
	middleware := NewTestMiddleware("test-middleware", 1)

	// Test removing non-existent middleware
	err := manager.RemoveMiddleware("non-existent")
	if err == nil {
		t.Error("Expected error for removing non-existent middleware")
	}

	// Add and then remove middleware
	manager.AddMiddleware(middleware)
	err = manager.RemoveMiddleware("test-middleware")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if manager.GetMiddlewareCount() != 0 {
		t.Errorf("Expected 0 middlewares after removal, got %d", manager.GetMiddlewareCount())
	}
}

func TestMiddlewareManager_GetMiddleware(t *testing.T) {
	manager := NewMiddlewareManager()
	middleware := NewTestMiddleware("test-middleware", 1)

	// Test getting non-existent middleware
	_, err := manager.GetMiddleware("non-existent")
	if err == nil {
		t.Error("Expected error for getting non-existent middleware")
	}

	// Add and get middleware
	manager.AddMiddleware(middleware)
	retrieved, err := manager.GetMiddleware("test-middleware")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if retrieved != middleware {
		t.Error("Expected to get the same middleware instance")
	}
}

func TestMiddlewareManager_ListMiddlewares(t *testing.T) {
	manager := NewMiddlewareManager()

	// Test empty list
	middlewares := manager.ListMiddlewares()
	if len(middlewares) != 0 {
		t.Errorf("Expected empty middleware list, got %d middlewares", len(middlewares))
	}

	// Add middlewares and test list
	middleware1 := NewTestMiddleware("middleware1", 1)
	middleware2 := NewTestMiddleware("middleware2", 2)

	manager.AddMiddleware(middleware1)
	manager.AddMiddleware(middleware2)

	middlewares = manager.ListMiddlewares()
	if len(middlewares) != 2 {
		t.Errorf("Expected 2 middlewares, got %d", len(middlewares))
	}

	// Check if both middlewares are in the list
	middlewareMap := make(map[string]bool)
	for _, name := range middlewares {
		middlewareMap[name] = true
	}

	if !middlewareMap["middleware1"] || !middlewareMap["middleware2"] {
		t.Error("Expected both 'middleware1' and 'middleware2' to be in the list")
	}
}

func TestMiddlewareManager_ProcessRequest(t *testing.T) {
	manager := NewMiddlewareManager()
	testLogger := &TestLogger{}
	manager.SetLogger(testLogger)

	middleware1 := NewTestMiddleware("middleware1", 1)
	middleware2 := NewTestMiddleware("middleware2", 2)
	middleware1.SetLogger(testLogger)
	middleware2.SetLogger(testLogger)

	manager.AddMiddleware(middleware1)
	manager.AddMiddleware(middleware2)

	ctx := context.Background()
	req := "test request"

	// Process request
	resp, err := manager.ProcessRequest(ctx, req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if resp != req {
		t.Errorf("Expected response to be %v, got %v", req, resp)
	}

	// Check that both middlewares processed the request
	if middleware1.GetProcessCount() != 1 {
		t.Errorf("Expected middleware1 process count to be 1, got %d", middleware1.GetProcessCount())
	}

	if middleware2.GetProcessCount() != 1 {
		t.Errorf("Expected middleware2 process count to be 1, got %d", middleware2.GetProcessCount())
	}
}

func TestMiddlewareManager_ProcessRequestWithError(t *testing.T) {
	manager := NewMiddlewareManager()
	testLogger := &TestLogger{}
	manager.SetLogger(testLogger)

	errorMiddleware := NewErrorMiddleware("error-middleware", 1)
	errorMiddleware.SetLogger(testLogger)

	manager.AddMiddleware(errorMiddleware)

	ctx := context.Background()
	req := "test request"

	// Process request
	_, err := manager.ProcessRequest(ctx, req)
	if err == nil {
		t.Error("Expected error from error middleware")
	}

	if err.Error() != "middleware error" {
		t.Errorf("Expected 'middleware error', got %v", err)
	}
}

func TestMiddlewareManager_MiddlewarePriority(t *testing.T) {
	manager := NewMiddlewareManager()

	// Add middlewares in reverse priority order
	middleware3 := NewTestMiddleware("middleware3", 3)
	middleware1 := NewTestMiddleware("middleware1", 1)
	middleware2 := NewTestMiddleware("middleware2", 2)

	manager.AddMiddleware(middleware3)
	manager.AddMiddleware(middleware1)
	manager.AddMiddleware(middleware2)

	// Check that they are sorted by priority
	middlewares := manager.ListMiddlewares()
	expectedOrder := []string{"middleware1", "middleware2", "middleware3"}

	for i, expected := range expectedOrder {
		if middlewares[i] != expected {
			t.Errorf("Expected middleware at position %d to be %s, got %s", i, expected, middlewares[i])
		}
	}
}

func TestMiddlewareManager_DisabledMiddleware(t *testing.T) {
	manager := NewMiddlewareManager()

	middleware1 := NewTestMiddleware("middleware1", 1)
	middleware2 := NewTestMiddleware("middleware2", 2)

	// Disable middleware2
	middleware2.SetEnabled(false)

	manager.AddMiddleware(middleware1)
	manager.AddMiddleware(middleware2)

	ctx := context.Background()
	req := "test request"

	// Process request
	_, err := manager.ProcessRequest(ctx, req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check that only enabled middleware processed the request
	if middleware1.GetProcessCount() != 1 {
		t.Errorf("Expected middleware1 process count to be 1, got %d", middleware1.GetProcessCount())
	}

	if middleware2.GetProcessCount() != 0 {
		t.Errorf("Expected middleware2 process count to be 0 (disabled), got %d", middleware2.GetProcessCount())
	}

	// Check enabled middleware count
	if manager.GetEnabledMiddlewareCount() != 1 {
		t.Errorf("Expected 1 enabled middleware, got %d", manager.GetEnabledMiddlewareCount())
	}
}

func TestMiddlewareManager_Clear(t *testing.T) {
	manager := NewMiddlewareManager()

	middleware1 := NewTestMiddleware("middleware1", 1)
	middleware2 := NewTestMiddleware("middleware2", 2)

	manager.AddMiddleware(middleware1)
	manager.AddMiddleware(middleware2)

	if manager.GetMiddlewareCount() != 2 {
		t.Errorf("Expected 2 middlewares before clear, got %d", manager.GetMiddlewareCount())
	}

	manager.Clear()

	if manager.GetMiddlewareCount() != 0 {
		t.Errorf("Expected 0 middlewares after clear, got %d", manager.GetMiddlewareCount())
	}

	middlewares := manager.ListMiddlewares()
	if len(middlewares) != 0 {
		t.Errorf("Expected empty middleware list after clear, got %d middlewares", len(middlewares))
	}
}

func TestNewBaseMiddleware(t *testing.T) {
	name := "test-base"
	priority := 5
	middleware := NewBaseMiddleware(name, priority)

	if middleware.Name() != name {
		t.Errorf("Expected name to be %s, got %s", name, middleware.Name())
	}

	if middleware.Priority() != priority {
		t.Errorf("Expected priority to be %d, got %d", priority, middleware.Priority())
	}

	if !middleware.IsEnabled() {
		t.Error("Expected middleware to be enabled by default")
	}

	if middleware.logger == nil {
		t.Error("Expected logger to be initialized")
	}
}

func TestBaseMiddleware_SetEnabled(t *testing.T) {
	middleware := NewBaseMiddleware("test", 1)

	middleware.SetEnabled(false)
	if middleware.IsEnabled() {
		t.Error("Expected middleware to be disabled")
	}

	middleware.SetEnabled(true)
	if !middleware.IsEnabled() {
		t.Error("Expected middleware to be enabled")
	}
}

func TestBaseMiddleware_SetPriority(t *testing.T) {
	middleware := NewBaseMiddleware("test", 1)

	newPriority := 10
	middleware.SetPriority(newPriority)

	if middleware.Priority() != newPriority {
		t.Errorf("Expected priority to be %d, got %d", newPriority, middleware.Priority())
	}
}

func TestBaseMiddleware_SetLogger(t *testing.T) {
	middleware := NewBaseMiddleware("test", 1)
	testLogger := &TestLogger{}

	middleware.SetLogger(testLogger)

	if middleware.GetLogger() != testLogger {
		t.Error("Expected logger to be set to test logger")
	}

	// Test that nil logger is ignored
	middleware.SetLogger(nil)
	if middleware.GetLogger() == nil {
		t.Error("Expected logger to not be nil after setting nil logger")
	}
}

func TestBaseMiddleware_Process(t *testing.T) {
	middleware := NewBaseMiddleware("test", 1)
	testLogger := &TestLogger{}
	middleware.SetLogger(testLogger)

	ctx := context.Background()
	req := "test request"

	nextCalled := false
	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		nextCalled = true
		return req, nil
	}

	resp, err := middleware.Process(ctx, req, next)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if resp != req {
		t.Errorf("Expected response to be %v, got %v", req, resp)
	}

	if !nextCalled {
		t.Error("Expected next function to be called")
	}

	if len(testLogger.DebugMessages) != 1 {
		t.Errorf("Expected 1 debug message, got %d", len(testLogger.DebugMessages))
	}
}

func TestObserverManager_AttachObserver(t *testing.T) {
	om := NewObserverManager()

	// Test nil observer
	err := om.AttachObserver(nil)
	if err == nil {
		t.Error("Expected error for nil observer")
	}

	// Test valid observer
	observer := &TestObserver{}
	err = om.AttachObserver(observer)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(om.observers) != 1 {
		t.Errorf("Expected 1 observer, got %d", len(om.observers))
	}
}

// TestObserver is a simple observer for testing.
type TestObserver struct {
	onRequestCalled  bool
	onResponseCalled bool
	onErrorCalled    bool
}

func (to *TestObserver) OnStart(ctx context.Context, addr string) error {
	return nil
}

func (to *TestObserver) OnStop(ctx context.Context) error {
	return nil
}

func (to *TestObserver) OnError(ctx context.Context, err error) error {
	to.onErrorCalled = true
	return nil
}

func (to *TestObserver) OnRequest(ctx context.Context, req interface{}) error {
	to.onRequestCalled = true
	return nil
}

func (to *TestObserver) OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error {
	to.onResponseCalled = true
	return nil
}

func (to *TestObserver) OnRouteEnter(ctx context.Context, method, path string, req interface{}) error {
	return nil
}

func (to *TestObserver) OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error {
	return nil
}

func TestObserverManager_NotifyObservers(t *testing.T) {
	om := NewObserverManager()
	observer := &TestObserver{}
	om.AttachObserver(observer)

	ctx := context.Background()

	// Test request event
	err := om.NotifyObservers(interfaces.EventRequest, ctx, "test request")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !observer.onRequestCalled {
		t.Error("Expected OnRequest to be called")
	}

	// Test response event
	respData := ResponseEventData{
		Request:  "test request",
		Response: "test response",
		Duration: time.Millisecond * 100,
	}
	err = om.NotifyObservers(interfaces.EventResponse, ctx, respData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !observer.onResponseCalled {
		t.Error("Expected OnResponse to be called")
	}

	// Test error event
	testErr := errors.New("test error")
	err = om.NotifyObservers(interfaces.EventError, ctx, testErr)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !observer.onErrorCalled {
		t.Error("Expected OnError to be called")
	}
}
