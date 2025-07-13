package logger

import (
	"context"
	"testing"
)

type mockLogger struct {
	lastLevel string
	lastCode  string
	lastMsg   string
	lastArgs  []interface{}
}

func (m *mockLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	m.lastLevel = "debug"
	m.lastMsg = format
	m.lastArgs = args
}

func (m *mockLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	m.lastLevel = "info"
	m.lastMsg = format
	m.lastArgs = args
}

func (m *mockLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	m.lastLevel = "warn"
	m.lastMsg = format
	m.lastArgs = args
}

func (m *mockLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	m.lastLevel = "error"
	m.lastMsg = format
	m.lastArgs = args
}

func (m *mockLogger) Panicf(ctx context.Context, format string, args ...interface{}) {
	m.lastLevel = "panic"
	m.lastMsg = format
	m.lastArgs = args
}

func (m *mockLogger) DebugCode(ctx context.Context, code, format string, args ...interface{}) {
	m.lastLevel = "debug"
	m.lastCode = code
	m.lastMsg = format
	m.lastArgs = args
}

func (m *mockLogger) InfoCode(ctx context.Context, code, format string, args ...interface{}) {
	m.lastLevel = "info"
	m.lastCode = code
	m.lastMsg = format
	m.lastArgs = args
}

func (m *mockLogger) WarnCode(ctx context.Context, code, format string, args ...interface{}) {
	m.lastLevel = "warn"
	m.lastCode = code
	m.lastMsg = format
	m.lastArgs = args
}

func (m *mockLogger) ErrorCode(ctx context.Context, code, format string, args ...interface{}) {
	m.lastLevel = "error"
	m.lastCode = code
	m.lastMsg = format
	m.lastArgs = args
}

func (m *mockLogger) PanicCode(ctx context.Context, code, format string, args ...interface{}) {
	m.lastLevel = "panic"
	m.lastCode = code
	m.lastMsg = format
	m.lastArgs = args
}

func TestLogger(t *testing.T) {
	mock := &mockLogger{}
	SetLoggerProvider(mock)
	ctx := context.Background()

	tests := []struct {
		name    string
		fn      func()
		level   string
		code    string
		message string
		args    []interface{}
	}{
		{
			name:    "Debugf",
			fn:      func() { Debugf(ctx, "test %s", "debug") },
			level:   "debug",
			message: "test %s",
			args:    []interface{}{"debug"},
		},
		{
			name:    "DebugCode",
			fn:      func() { DebugCode(ctx, "CODE1", "test %s", "debug") },
			level:   "debug",
			code:    "CODE1",
			message: "test %s",
			args:    []interface{}{"debug"},
		},
		{
			name:    "Infof",
			fn:      func() { Infof(ctx, "test %s", "info") },
			level:   "info",
			message: "test %s",
			args:    []interface{}{"info"},
		},
		{
			name:    "InfoCode",
			fn:      func() { InfoCode(ctx, "CODE2", "test %s", "info") },
			level:   "info",
			code:    "CODE2",
			message: "test %s",
			args:    []interface{}{"info"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.lastLevel = ""
			mock.lastCode = ""
			mock.lastMsg = ""
			mock.lastArgs = nil

			tt.fn()

			if mock.lastLevel != tt.level {
				t.Errorf("expected level %s, got %s", tt.level, mock.lastLevel)
			}
			if tt.code != "" && mock.lastCode != tt.code {
				t.Errorf("expected code %s, got %s", tt.code, mock.lastCode)
			}
			if mock.lastMsg != tt.message {
				t.Errorf("expected message %s, got %s", tt.message, mock.lastMsg)
			}
			if len(mock.lastArgs) != len(tt.args) {
				t.Errorf("expected %d args, got %d", len(tt.args), len(mock.lastArgs))
			}
		})
	}
}
func TestWarnErrorPanicFunctions(t *testing.T) {
	mock := &mockLogger{}
	SetLoggerProvider(mock)
	ctx := context.Background()

	tests := []struct {
		name    string
		fn      func()
		level   string
		message string
		args    []interface{}
	}{
		{
			name:    "Warnf",
			fn:      func() { Warnf(ctx, "test %s", "warn") },
			level:   "warn",
			message: "test %s",
			args:    []interface{}{"warn"},
		},
		{
			name:    "Errorf",
			fn:      func() { Errorf(ctx, "test %s", "error") },
			level:   "error",
			message: "test %s",
			args:    []interface{}{"error"},
		},
		{
			name:    "Panicf",
			fn:      func() { Panicf(ctx, "test %s", "panic") },
			level:   "panic",
			message: "test %s",
			args:    []interface{}{"panic"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.lastLevel = ""
			mock.lastMsg = ""
			mock.lastArgs = nil

			tt.fn()

			if mock.lastLevel != tt.level {
				t.Errorf("expected level %s, got %s", tt.level, mock.lastLevel)
			}
			if mock.lastMsg != tt.message {
				t.Errorf("expected message %s, got %s", tt.message, mock.lastMsg)
			}
			if len(mock.lastArgs) != len(tt.args) {
				t.Errorf("expected %d args, got %d", len(tt.args), len(mock.lastArgs))
			}
		})
	}
}
func TestWarnErrorPanicCodeFunctions(t *testing.T) {
	mock := &mockLogger{}
	SetLoggerProvider(mock)
	ctx := context.Background()

	tests := []struct {
		name    string
		fn      func()
		level   string
		code    string
		message string
		args    []interface{}
	}{
		{
			name:    "WarnCode",
			fn:      func() { WarnCode(ctx, "CODE3", "test %s", "warn") },
			level:   "warn",
			code:    "CODE3",
			message: "test %s",
			args:    []interface{}{"warn"},
		},
		{
			name:    "ErrorCode",
			fn:      func() { ErrorCode(ctx, "CODE4", "test %s", "error") },
			level:   "error",
			code:    "CODE4",
			message: "test %s",
			args:    []interface{}{"error"},
		},
		{
			name:    "PanicCode",
			fn:      func() { PanicCode(ctx, "CODE5", "test %s", "panic") },
			level:   "panic",
			code:    "CODE5",
			message: "test %s",
			args:    []interface{}{"panic"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.lastLevel = ""
			mock.lastCode = ""
			mock.lastMsg = ""
			mock.lastArgs = nil

			tt.fn()

			if mock.lastLevel != tt.level {
				t.Errorf("expected level %s, got %s", tt.level, mock.lastLevel)
			}
			if mock.lastCode != tt.code {
				t.Errorf("expected code %s, got %s", tt.code, mock.lastCode)
			}
			if mock.lastMsg != tt.message {
				t.Errorf("expected message %s, got %s", tt.message, mock.lastMsg)
			}
			if len(mock.lastArgs) != len(tt.args) {
				t.Errorf("expected %d args, got %d", len(tt.args), len(mock.lastArgs))
			}
		})
	}
}
