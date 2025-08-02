package hooks

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/httpserver/interfaces"
)

func TestNewObserverManager(t *testing.T) {
	om := NewObserverManager()

	if om == nil {
		t.Fatal("NewObserverManager returned nil")
	}

	if om.GetObserverCount() != 0 {
		t.Errorf("Expected 0 observers, got %d", om.GetObserverCount())
	}

	if om.GetHookCount(interfaces.EventStart) != 0 {
		t.Errorf("Expected 0 hooks, got %d", om.GetHookCount(interfaces.EventStart))
	}
}

func TestAttachObserver(t *testing.T) {
	om := NewObserverManager()
	observer := &mockObserver{}

	// Test successful attach
	err := om.AttachObserver(observer)
	if err != nil {
		t.Errorf("AttachObserver error = %v, want nil", err)
	}

	if om.GetObserverCount() != 1 {
		t.Errorf("Expected 1 observer, got %d", om.GetObserverCount())
	}

	// Test attach nil observer
	err = om.AttachObserver(nil)
	if err == nil {
		t.Error("AttachObserver with nil should return error")
	}

	// Test attach duplicate observer
	err = om.AttachObserver(observer)
	if err == nil {
		t.Error("AttachObserver with duplicate should return error")
	}

	if om.GetObserverCount() != 1 {
		t.Errorf("Expected 1 observer after duplicate, got %d", om.GetObserverCount())
	}
}

func TestDetachObserver(t *testing.T) {
	om := NewObserverManager()
	observer1 := &mockObserver{}
	observer2 := &mockObserver{}

	// Attach observers
	om.AttachObserver(observer1)
	om.AttachObserver(observer2)

	if om.GetObserverCount() != 2 {
		t.Errorf("Expected 2 observers, got %d", om.GetObserverCount())
	}

	// Test successful detach
	err := om.DetachObserver(observer1)
	if err != nil {
		t.Errorf("DetachObserver error = %v, want nil", err)
	}

	if om.GetObserverCount() != 1 {
		t.Errorf("Expected 1 observer after detach, got %d", om.GetObserverCount())
	}

	// Test detach nil observer
	err = om.DetachObserver(nil)
	if err == nil {
		t.Error("DetachObserver with nil should return error")
	}

	// Test detach non-existent observer
	detachedObserver := &mockObserver{}
	err = om.DetachObserver(detachedObserver)
	if err == nil {
		t.Error("DetachObserver with non-existent should return error")
	}
}

func TestAttachHook(t *testing.T) {
	om := NewObserverManager()
	hook := func(ctx context.Context, data interface{}) error {
		return nil
	}

	// Test successful attach
	err := om.AttachHook(interfaces.EventStart, hook)
	if err != nil {
		t.Errorf("AttachHook error = %v, want nil", err)
	}

	if om.GetHookCount(interfaces.EventStart) != 1 {
		t.Errorf("Expected 1 hook, got %d", om.GetHookCount(interfaces.EventStart))
	}

	// Test attach nil hook
	err = om.AttachHook(interfaces.EventStart, nil)
	if err == nil {
		t.Error("AttachHook with nil should return error")
	}

	// Test attach multiple hooks for same event
	hook2 := func(ctx context.Context, data interface{}) error {
		return nil
	}
	err = om.AttachHook(interfaces.EventStart, hook2)
	if err != nil {
		t.Errorf("AttachHook second hook error = %v, want nil", err)
	}

	if om.GetHookCount(interfaces.EventStart) != 2 {
		t.Errorf("Expected 2 hooks, got %d", om.GetHookCount(interfaces.EventStart))
	}
}

func TestDetachHook(t *testing.T) {
	om := NewObserverManager()
	hook := func(ctx context.Context, data interface{}) error {
		return nil
	}

	// Attach hook
	om.AttachHook(interfaces.EventStart, hook)
	if om.GetHookCount(interfaces.EventStart) != 1 {
		t.Errorf("Expected 1 hook, got %d", om.GetHookCount(interfaces.EventStart))
	}

	// Test successful detach
	err := om.DetachHook(interfaces.EventStart)
	if err != nil {
		t.Errorf("DetachHook error = %v, want nil", err)
	}

	if om.GetHookCount(interfaces.EventStart) != 0 {
		t.Errorf("Expected 0 hooks after detach, got %d", om.GetHookCount(interfaces.EventStart))
	}
}

func TestNotifyObservers(t *testing.T) {
	om := NewObserverManager()
	observer := &mockObserver{}
	var hookCalled bool

	hook := func(ctx context.Context, data interface{}) error {
		hookCalled = true
		return nil
	}

	om.AttachObserver(observer)
	om.AttachHook(interfaces.EventStart, hook)

	// Test start event notification
	ctx := context.Background()
	err := om.NotifyObservers(interfaces.EventStart, ctx, "127.0.0.1:8080")
	if err != nil {
		t.Errorf("NotifyObservers error = %v, want nil", err)
	}

	if !hookCalled {
		t.Error("Hook was not called")
	}

	if !observer.startCalled {
		t.Error("Observer OnStart was not called")
	}
}

func TestNotifyObserversHookError(t *testing.T) {
	om := NewObserverManager()
	hookError := errors.New("hook error")

	hook := func(ctx context.Context, data interface{}) error {
		return hookError
	}

	om.AttachHook(interfaces.EventStart, hook)

	// Test hook error propagation
	ctx := context.Background()
	err := om.NotifyObservers(interfaces.EventStart, ctx, "127.0.0.1:8080")
	if err == nil {
		t.Error("NotifyObservers should return error when hook fails")
	}
}

func TestNotifyObserversEventTypes(t *testing.T) {
	om := NewObserverManager()
	observer := &mockObserver{}
	om.AttachObserver(observer)
	ctx := context.Background()

	tests := []struct {
		name      string
		eventType interfaces.EventType
		data      interface{}
		wantError bool
	}{
		{
			name:      "start event",
			eventType: interfaces.EventStart,
			data:      "127.0.0.1:8080",
			wantError: false,
		},
		{
			name:      "start event invalid data",
			eventType: interfaces.EventStart,
			data:      123,
			wantError: true,
		},
		{
			name:      "stop event",
			eventType: interfaces.EventStop,
			data:      nil,
			wantError: false,
		},
		{
			name:      "error event",
			eventType: interfaces.EventError,
			data:      errors.New("test error"),
			wantError: false,
		},
		{
			name:      "error event invalid data",
			eventType: interfaces.EventError,
			data:      "not an error",
			wantError: true,
		},
		{
			name:      "request event",
			eventType: interfaces.EventRequest,
			data:      &http.Request{},
			wantError: false,
		},
		{
			name:      "request event with generic data",
			eventType: interfaces.EventRequest,
			data:      "any request data",
			wantError: false,
		},
		{
			name:      "response event",
			eventType: interfaces.EventResponse,
			data: ResponseEventData{
				Request:  &http.Request{},
				Response: &http.Response{},
				Duration: time.Millisecond,
			},
			wantError: false,
		},
		{
			name:      "response event invalid data",
			eventType: interfaces.EventResponse,
			data:      "not response data",
			wantError: true,
		},
		{
			name:      "route enter event",
			eventType: interfaces.EventRouteEnter,
			data: RouteEventData{
				Method:  "GET",
				Path:    "/test",
				Request: &http.Request{},
			},
			wantError: false,
		},
		{
			name:      "route enter event invalid data",
			eventType: interfaces.EventRouteEnter,
			data:      "not route data",
			wantError: true,
		},
		{
			name:      "route exit event",
			eventType: interfaces.EventRouteExit,
			data: RouteExitEventData{
				Method:   "GET",
				Path:     "/test",
				Request:  &http.Request{},
				Duration: time.Millisecond,
			},
			wantError: false,
		},
		{
			name:      "route exit event invalid data",
			eventType: interfaces.EventRouteExit,
			data:      "not route exit data",
			wantError: true,
		},
		{
			name:      "unknown event type",
			eventType: interfaces.EventType("unknown"),
			data:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observer.reset()
			err := om.NotifyObservers(tt.eventType, ctx, tt.data)
			if (err != nil) != tt.wantError {
				t.Errorf("NotifyObservers() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestClear(t *testing.T) {
	om := NewObserverManager()
	observer := &mockObserver{}
	hook := func(ctx context.Context, data interface{}) error {
		return nil
	}

	// Add observer and hook
	om.AttachObserver(observer)
	om.AttachHook(interfaces.EventStart, hook)

	if om.GetObserverCount() != 1 {
		t.Errorf("Expected 1 observer, got %d", om.GetObserverCount())
	}

	if om.GetHookCount(interfaces.EventStart) != 1 {
		t.Errorf("Expected 1 hook, got %d", om.GetHookCount(interfaces.EventStart))
	}

	// Clear all
	om.Clear()

	if om.GetObserverCount() != 0 {
		t.Errorf("Expected 0 observers after clear, got %d", om.GetObserverCount())
	}

	if om.GetHookCount(interfaces.EventStart) != 0 {
		t.Errorf("Expected 0 hooks after clear, got %d", om.GetHookCount(interfaces.EventStart))
	}
}

// Mock observer for testing
type mockObserver struct {
	startCalled      bool
	stopCalled       bool
	errorCalled      bool
	requestCalled    bool
	responseCalled   bool
	routeEnterCalled bool
	routeExitCalled  bool
}

func (m *mockObserver) OnStart(ctx context.Context, addr string) error {
	m.startCalled = true
	return nil
}

func (m *mockObserver) OnStop(ctx context.Context) error {
	m.stopCalled = true
	return nil
}

func (m *mockObserver) OnError(ctx context.Context, err error) error {
	m.errorCalled = true
	return nil
}

func (m *mockObserver) OnRequest(ctx context.Context, req interface{}) error {
	m.requestCalled = true
	return nil
}

func (m *mockObserver) OnResponse(ctx context.Context, req interface{}, resp interface{}, duration time.Duration) error {
	m.responseCalled = true
	return nil
}

func (m *mockObserver) OnRouteEnter(ctx context.Context, method, path string, req interface{}) error {
	m.routeEnterCalled = true
	return nil
}

func (m *mockObserver) OnRouteExit(ctx context.Context, method, path string, req interface{}, duration time.Duration) error {
	m.routeExitCalled = true
	return nil
}

func (m *mockObserver) reset() {
	m.startCalled = false
	m.stopCalled = false
	m.errorCalled = false
	m.requestCalled = false
	m.responseCalled = false
	m.routeEnterCalled = false
	m.routeExitCalled = false
}
