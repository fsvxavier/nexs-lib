// Code generated manually. DO NOT EDIT.
package mocks

import (
	"context"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// MockDomainErrorInterface is a mock implementation of interfaces.DomainErrorInterface
type MockDomainErrorInterface struct {
	ErrorFunc        func() string
	UnwrapFunc       func() error
	TypeFunc         func() interfaces.ErrorType
	MetadataFunc     func() map[string]interface{}
	HTTPStatusFunc   func() int
	StackTraceFunc   func() string
	WithContextFunc  func(ctx context.Context) interfaces.DomainErrorInterface
	WrapFunc         func(err error) interfaces.DomainErrorInterface
	WithMetadataFunc func(key string, value interface{}) interfaces.DomainErrorInterface
	CodeFunc         func() string
	TimestampFunc    func() time.Time
	ToJSONFunc       func() ([]byte, error)
}

func (m *MockDomainErrorInterface) Error() string {
	if m.ErrorFunc != nil {
		return m.ErrorFunc()
	}
	return "mock error"
}

func (m *MockDomainErrorInterface) Unwrap() error {
	if m.UnwrapFunc != nil {
		return m.UnwrapFunc()
	}
	return nil
}

func (m *MockDomainErrorInterface) Type() interfaces.ErrorType {
	if m.TypeFunc != nil {
		return m.TypeFunc()
	}
	return interfaces.ValidationError
}

func (m *MockDomainErrorInterface) Metadata() map[string]interface{} {
	if m.MetadataFunc != nil {
		return m.MetadataFunc()
	}
	return make(map[string]interface{})
}

func (m *MockDomainErrorInterface) HTTPStatus() int {
	if m.HTTPStatusFunc != nil {
		return m.HTTPStatusFunc()
	}
	return 400
}

func (m *MockDomainErrorInterface) StackTrace() string {
	if m.StackTraceFunc != nil {
		return m.StackTraceFunc()
	}
	return ""
}

func (m *MockDomainErrorInterface) WithContext(ctx context.Context) interfaces.DomainErrorInterface {
	if m.WithContextFunc != nil {
		return m.WithContextFunc(ctx)
	}
	return m
}

func (m *MockDomainErrorInterface) Wrap(err error) interfaces.DomainErrorInterface {
	if m.WrapFunc != nil {
		return m.WrapFunc(err)
	}
	return m
}

func (m *MockDomainErrorInterface) WithMetadata(key string, value interface{}) interfaces.DomainErrorInterface {
	if m.WithMetadataFunc != nil {
		return m.WithMetadataFunc(key, value)
	}
	return m
}

func (m *MockDomainErrorInterface) Code() string {
	if m.CodeFunc != nil {
		return m.CodeFunc()
	}
	return "MOCK001"
}

func (m *MockDomainErrorInterface) Timestamp() time.Time {
	if m.TimestampFunc != nil {
		return m.TimestampFunc()
	}
	return time.Now()
}

func (m *MockDomainErrorInterface) ToJSON() ([]byte, error) {
	if m.ToJSONFunc != nil {
		return m.ToJSONFunc()
	}
	return []byte(`{"code":"MOCK001","message":"mock error"}`), nil
}

// MockErrorFactory is a mock implementation of interfaces.ErrorFactory
type MockErrorFactory struct {
	NewFunc             func(errorType interfaces.ErrorType, code, message string) interfaces.DomainErrorInterface
	NewWithMetadataFunc func(errorType interfaces.ErrorType, code, message string, metadata map[string]interface{}) interfaces.DomainErrorInterface
	WrapFunc            func(err error, errorType interfaces.ErrorType, code, message string) interfaces.DomainErrorInterface
}

func (m *MockErrorFactory) New(errorType interfaces.ErrorType, code, message string) interfaces.DomainErrorInterface {
	if m.NewFunc != nil {
		return m.NewFunc(errorType, code, message)
	}
	return &MockDomainErrorInterface{
		ErrorFunc: func() string { return message },
		TypeFunc:  func() interfaces.ErrorType { return errorType },
		CodeFunc:  func() string { return code },
	}
}

func (m *MockErrorFactory) NewWithMetadata(errorType interfaces.ErrorType, code, message string, metadata map[string]interface{}) interfaces.DomainErrorInterface {
	if m.NewWithMetadataFunc != nil {
		return m.NewWithMetadataFunc(errorType, code, message, metadata)
	}
	return &MockDomainErrorInterface{
		ErrorFunc:    func() string { return message },
		TypeFunc:     func() interfaces.ErrorType { return errorType },
		CodeFunc:     func() string { return code },
		MetadataFunc: func() map[string]interface{} { return metadata },
	}
}

func (m *MockErrorFactory) Wrap(err error, errorType interfaces.ErrorType, code, message string) interfaces.DomainErrorInterface {
	if m.WrapFunc != nil {
		return m.WrapFunc(err, errorType, code, message)
	}
	return &MockDomainErrorInterface{
		ErrorFunc:  func() string { return message },
		TypeFunc:   func() interfaces.ErrorType { return errorType },
		CodeFunc:   func() string { return code },
		UnwrapFunc: func() error { return err },
	}
}

// MockErrorTypeChecker is a mock implementation of interfaces.ErrorTypeChecker
type MockErrorTypeChecker struct {
	IsTypeFunc func(err error, errorType interfaces.ErrorType) bool
}

func (m *MockErrorTypeChecker) IsType(err error, errorType interfaces.ErrorType) bool {
	if m.IsTypeFunc != nil {
		return m.IsTypeFunc(err, errorType)
	}
	return false
}

// MockObserver is a mock implementation of interfaces.Observer
type MockObserver struct {
	OnErrorFunc func(ctx context.Context, err interfaces.DomainErrorInterface) error
}

func (m *MockObserver) OnError(ctx context.Context, err interfaces.DomainErrorInterface) error {
	if m.OnErrorFunc != nil {
		return m.OnErrorFunc(ctx, err)
	}
	return nil
}

// MockStackTraceCapture is a mock implementation of interfaces.StackTraceCapture
type MockStackTraceCapture struct {
	CaptureStackTraceFunc func(skip int) []interfaces.StackFrame
	FormatStackTraceFunc  func(frames []interfaces.StackFrame) string
}

func (m *MockStackTraceCapture) CaptureStackTrace(skip int) []interfaces.StackFrame {
	if m.CaptureStackTraceFunc != nil {
		return m.CaptureStackTraceFunc(skip)
	}
	return []interfaces.StackFrame{
		{
			Function: "test.function",
			File:     "test.go",
			Line:     42,
			Time:     time.Now().Format(time.RFC3339),
		},
	}
}

func (m *MockStackTraceCapture) FormatStackTrace(frames []interfaces.StackFrame) string {
	if m.FormatStackTraceFunc != nil {
		return m.FormatStackTraceFunc(frames)
	}
	return "Stack trace:\n  1. test.function\n     test.go:42"
}

// MockHookManager is a mock implementation of interfaces.HookManager
type MockHookManager struct {
	RegisterStartHookFunc func(hook interfaces.StartHookFunc)
	RegisterStopHookFunc  func(hook interfaces.StopHookFunc)
	RegisterErrorHookFunc func(hook interfaces.ErrorHookFunc)
	RegisterI18nHookFunc  func(hook interfaces.I18nHookFunc)
	ExecuteStartHooksFunc func(ctx context.Context) error
	ExecuteStopHooksFunc  func(ctx context.Context) error
	ExecuteErrorHooksFunc func(ctx context.Context, err interfaces.DomainErrorInterface) error
	ExecuteI18nHooksFunc  func(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error
}

func (m *MockHookManager) RegisterStartHook(hook interfaces.StartHookFunc) {
	if m.RegisterStartHookFunc != nil {
		m.RegisterStartHookFunc(hook)
	}
}

func (m *MockHookManager) RegisterStopHook(hook interfaces.StopHookFunc) {
	if m.RegisterStopHookFunc != nil {
		m.RegisterStopHookFunc(hook)
	}
}

func (m *MockHookManager) RegisterErrorHook(hook interfaces.ErrorHookFunc) {
	if m.RegisterErrorHookFunc != nil {
		m.RegisterErrorHookFunc(hook)
	}
}

func (m *MockHookManager) RegisterI18nHook(hook interfaces.I18nHookFunc) {
	if m.RegisterI18nHookFunc != nil {
		m.RegisterI18nHookFunc(hook)
	}
}

func (m *MockHookManager) ExecuteStartHooks(ctx context.Context) error {
	if m.ExecuteStartHooksFunc != nil {
		return m.ExecuteStartHooksFunc(ctx)
	}
	return nil
}

func (m *MockHookManager) ExecuteStopHooks(ctx context.Context) error {
	if m.ExecuteStopHooksFunc != nil {
		return m.ExecuteStopHooksFunc(ctx)
	}
	return nil
}

func (m *MockHookManager) ExecuteErrorHooks(ctx context.Context, err interfaces.DomainErrorInterface) error {
	if m.ExecuteErrorHooksFunc != nil {
		return m.ExecuteErrorHooksFunc(ctx, err)
	}
	return nil
}

func (m *MockHookManager) ExecuteI18nHooks(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error {
	if m.ExecuteI18nHooksFunc != nil {
		return m.ExecuteI18nHooksFunc(ctx, err, locale)
	}
	return nil
}

// MockMiddlewareManager is a mock implementation of interfaces.MiddlewareManager
type MockMiddlewareManager struct {
	RegisterMiddlewareFunc     func(middleware interfaces.MiddlewareFunc)
	RegisterI18nMiddlewareFunc func(middleware interfaces.I18nMiddlewareFunc)
	ExecuteMiddlewaresFunc     func(ctx context.Context, err interfaces.DomainErrorInterface) interfaces.DomainErrorInterface
	ExecuteI18nMiddlewaresFunc func(ctx context.Context, err interfaces.DomainErrorInterface, locale string) interfaces.DomainErrorInterface
}

func (m *MockMiddlewareManager) RegisterMiddleware(middleware interfaces.MiddlewareFunc) {
	if m.RegisterMiddlewareFunc != nil {
		m.RegisterMiddlewareFunc(middleware)
	}
}

func (m *MockMiddlewareManager) RegisterI18nMiddleware(middleware interfaces.I18nMiddlewareFunc) {
	if m.RegisterI18nMiddlewareFunc != nil {
		m.RegisterI18nMiddlewareFunc(middleware)
	}
}

func (m *MockMiddlewareManager) ExecuteMiddlewares(ctx context.Context, err interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
	if m.ExecuteMiddlewaresFunc != nil {
		return m.ExecuteMiddlewaresFunc(ctx, err)
	}
	return err
}

func (m *MockMiddlewareManager) ExecuteI18nMiddlewares(ctx context.Context, err interfaces.DomainErrorInterface, locale string) interfaces.DomainErrorInterface {
	if m.ExecuteI18nMiddlewaresFunc != nil {
		return m.ExecuteI18nMiddlewaresFunc(ctx, err, locale)
	}
	return err
}
