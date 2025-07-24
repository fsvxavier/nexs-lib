package zap

import (
	"bytes"
	"context"
	"io"
	"os"
	"sync"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestInfo(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		msg    string
		fields []zap.Field
	}{
		{
			name:   "should log info message without fields",
			ctx:    context.Background(),
			msg:    "test message",
			fields: []zap.Field{},
		},
		{
			name: "should log info message with fields",
			ctx:  context.Background(),
			msg:  "test message with fields",
			fields: []zap.Field{
				zap.String("key", "value"),
				zap.Int("count", 1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test execution should not panic
			Info(tt.ctx, tt.msg, tt.fields...)
		})
	}
}
func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	if logger == nil {
		t.Error("NewLogger() returned nil")
	}
}

func TestDebug(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		msg    string
		fields []zap.Field
	}{
		{
			name:   "should log debug message without fields",
			ctx:    context.Background(),
			msg:    "test debug message",
			fields: []zap.Field{},
		},
		{
			name: "should log debug message with fields",
			ctx:  context.Background(),
			msg:  "test debug message with fields",
			fields: []zap.Field{
				zap.String("key", "value"),
				zap.Int("count", 1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Debug(tt.ctx, tt.msg, tt.fields...)
		})
	}
}

func TestWarn(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		msg    string
		fields []zap.Field
	}{
		{
			name:   "should log warn message without fields",
			ctx:    context.Background(),
			msg:    "test warn message",
			fields: []zap.Field{},
		},
		{
			name: "should log warn message with fields",
			ctx:  context.Background(),
			msg:  "test warn message with fields",
			fields: []zap.Field{
				zap.String("key", "value"),
				zap.Int("count", 1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Warn(tt.ctx, tt.msg, tt.fields...)
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		msg    string
		fields []zap.Field
	}{
		{
			name:   "should log error message without fields",
			ctx:    context.Background(),
			msg:    "test error message",
			fields: []zap.Field{},
		},
		{
			name: "should log error message with fields",
			ctx:  context.Background(),
			msg:  "test error message with fields",
			fields: []zap.Field{
				zap.String("key", "value"),
				zap.Int("count", 1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Error(tt.ctx, tt.msg, tt.fields...)
		})
	}
}

func TestWithLevel(t *testing.T) {
	logger := NewLogger()

	tests := []struct {
		name     string
		level    string
		expected zapcore.Level
	}{
		{
			name:     "should set debug level",
			level:    "debug",
			expected: zapcore.DebugLevel,
		},
		{
			name:     "should set info level",
			level:    "info",
			expected: zapcore.InfoLevel,
		},
		{
			name:     "should set warn level",
			level:    "warn",
			expected: zapcore.WarnLevel,
		},
		{
			name:     "should set error level",
			level:    "error",
			expected: zapcore.ErrorLevel,
		},
		{
			name:     "should set info level when invalid level provided",
			level:    "invalid",
			expected: zapcore.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.WithLevel(tt.level)
			if logger.Level.Level() != tt.expected {
				t.Errorf("WithLevel() got = %v, want %v", logger.Level.Level(), tt.expected)
			}
		})
	}
}
func TestDebugCode(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		code   string
		msg    string
		fields []interface{}
	}{
		{
			name:   "should log debug code message without fields",
			ctx:    context.Background(),
			code:   "TEST001",
			msg:    "test debug message",
			fields: []interface{}{},
		},
		{
			name:   "should log debug code message with fields",
			ctx:    context.Background(),
			code:   "TEST002",
			msg:    "test debug message with fields",
			fields: []interface{}{"key", "value", "count", 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DebugCode(tt.ctx, tt.code, tt.msg, tt.fields...)
		})
	}
}

func TestInfoCode(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		code   string
		msg    string
		fields []interface{}
	}{
		{
			name:   "should log info code message without fields",
			ctx:    context.Background(),
			code:   "TEST001",
			msg:    "test info message",
			fields: []interface{}{},
		},
		{
			name:   "should log info code message with fields",
			ctx:    context.Background(),
			code:   "TEST002",
			msg:    "test info message with fields",
			fields: []interface{}{"key", "value", "count", 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InfoCode(tt.ctx, tt.code, tt.msg, tt.fields...)
		})
	}
}

func TestWarnCode(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		code   string
		msg    string
		fields []interface{}
	}{
		{
			name:   "should log warn code message without fields",
			ctx:    context.Background(),
			code:   "TEST001",
			msg:    "test warn message",
			fields: []interface{}{},
		},
		{
			name:   "should log warn code message with fields",
			ctx:    context.Background(),
			code:   "TEST002",
			msg:    "test warn message with fields",
			fields: []interface{}{"key", "value", "count", 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WarnCode(tt.ctx, tt.code, tt.msg, tt.fields...)
		})
	}
}

func TestErrorCode(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		code   string
		msg    string
		fields []interface{}
	}{
		{
			name:   "should log error code message without fields",
			ctx:    context.Background(),
			code:   "TEST001",
			msg:    "test error message",
			fields: []interface{}{},
		},
		{
			name:   "should log error code message with fields",
			ctx:    context.Background(),
			code:   "TEST002",
			msg:    "test error message with fields",
			fields: []interface{}{"key", "value", "count", 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ErrorCode(tt.ctx, tt.code, tt.msg, tt.fields...)
		})
	}
}

func TestInfoOutCtx(t *testing.T) {
	tests := []struct {
		name   string
		msg    string
		fields []zap.Field
	}{
		{
			name:   "should log info message without context and fields",
			msg:    "test message",
			fields: []zap.Field{},
		},
		{
			name: "should log info message without context with fields",
			msg:  "test message with fields",
			fields: []zap.Field{
				zap.String("key", "value"),
				zap.Int("count", 1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InfoOutCtx(tt.msg, tt.fields...)
		})
	}
}

func TestErrorOutCtx(t *testing.T) {
	tests := []struct {
		name   string
		msg    string
		fields []zap.Field
	}{
		{
			name:   "should log error message without context and fields",
			msg:    "test message",
			fields: []zap.Field{},
		},
		{
			name: "should log error message without context with fields",
			msg:  "test message with fields",
			fields: []zap.Field{
				zap.String("key", "value"),
				zap.Int("count", 1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ErrorOutCtx(tt.msg, tt.fields...)
		})
	}
}

func TestInfof(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		format string
		args   []interface{}
	}{
		{
			name:   "should log formatted info message without args",
			ctx:    context.Background(),
			format: "test message",
			args:   []interface{}{},
		},
		{
			name:   "should log formatted info message with args",
			ctx:    context.Background(),
			format: "test message %s %d",
			args:   []interface{}{"value", 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Infof(tt.ctx, tt.format, tt.args...)
		})
	}
}

func TestErrorf(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		format string
		args   []interface{}
	}{
		{
			name:   "should log formatted error message without args",
			ctx:    context.Background(),
			format: "test message",
			args:   []interface{}{},
		},
		{
			name:   "should log formatted error message with args",
			ctx:    context.Background(),
			format: "test message %s %d",
			args:   []interface{}{"value", 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Errorf(tt.ctx, tt.format, tt.args...)
		})
	}
}

func TestInfoln(t *testing.T) {
	tests := []struct {
		name string
		args []interface{}
	}{
		{
			name: "should log info message without args",
			args: []interface{}{},
		},
		{
			name: "should log info message with args",
			args: []interface{}{"test message", "value", 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Infoln(tt.args...)
		})
	}
}

func TestErrorln(t *testing.T) {
	tests := []struct {
		name string
		args []interface{}
	}{
		{
			name: "should log error message without args",
			args: []interface{}{},
		},
		{
			name: "should log error message with args",
			args: []interface{}{"test message", "value", 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Errorln(tt.args...)
		})
	}
}
func TestSetOutput(t *testing.T) {
	tests := []struct {
		name   string
		writer io.Writer
	}{
		{
			name:   "should set output to buffer writer",
			writer: &bytes.Buffer{},
		},
		{
			name:   "should set output to os.Stderr",
			writer: os.Stderr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger()
			result := logger.SetOutput(tt.writer)

			if result == nil {
				t.Error("SetOutput() returned nil logger")
			}

			if result.Zlg == nil {
				t.Error("SetOutput() resulted in nil zap logger")
			}
		})
	}
}
func TestDebugln(t *testing.T) {
	tests := []struct {
		name string
		args []interface{}
	}{
		{
			name: "should log debug message without args",
			args: []interface{}{},
		},
		{
			name: "should log debug message with args",
			args: []interface{}{"test message", "value", 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Debugln(tt.args...)
		})
	}
}

func TestWarnln(t *testing.T) {
	tests := []struct {
		name string
		args []interface{}
	}{
		{
			name: "should log warn message without args",
			args: []interface{}{},
		},
		{
			name: "should log warn message with args",
			args: []interface{}{"test message", "value", 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Warnln(tt.args...)
		})
	}
}

func TestWarningln(t *testing.T) {
	tests := []struct {
		name string
		args []interface{}
	}{
		{
			name: "should log warning message without args",
			args: []interface{}{},
		},
		{
			name: "should log warning message with args",
			args: []interface{}{"test message", "value", 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Warningln(tt.args...)
		})
	}
}
func TestLoggerDebugCode(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		code   string
		msg    string
		fields []zap.Field
	}{
		{
			name:   "should log debug code message without fields",
			ctx:    context.Background(),
			code:   "TEST001",
			msg:    "test debug message",
			fields: []zap.Field{},
		},
		{
			name: "should log debug code message with fields",
			ctx:  context.Background(),
			code: "TEST002",
			msg:  "test debug message with fields",
			fields: []zap.Field{
				zap.String("key", "value"),
				zap.Int("count", 1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger()
			logger.DebugCode(tt.ctx, tt.code, tt.msg, tt.fields...)
		})
	}
}

func TestLoggerInfoCode(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		code   string
		msg    string
		fields []zap.Field
	}{
		{
			name:   "should log info code message without fields",
			ctx:    context.Background(),
			code:   "TEST001",
			msg:    "test info message",
			fields: []zap.Field{},
		},
		{
			name: "should log info code message with fields",
			ctx:  context.Background(),
			code: "TEST002",
			msg:  "test info message with fields",
			fields: []zap.Field{
				zap.String("key", "value"),
				zap.Int("count", 1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger()
			logger.InfoCode(tt.ctx, tt.code, tt.msg, tt.fields...)
		})
	}
}

func TestLoggerWarnCode(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		code   string
		msg    string
		fields []zap.Field
	}{
		{
			name:   "should log warn code message without fields",
			ctx:    context.Background(),
			code:   "TEST001",
			msg:    "test warn message",
			fields: []zap.Field{},
		},
		{
			name: "should log warn code message with fields",
			ctx:  context.Background(),
			code: "TEST002",
			msg:  "test warn message with fields",
			fields: []zap.Field{
				zap.String("key", "value"),
				zap.Int("count", 1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger()
			logger.WarnCode(tt.ctx, tt.code, tt.msg, tt.fields...)
		})
	}
}

func TestLoggerErrorCode(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		code   string
		msg    string
		fields []zap.Field
	}{
		{
			name:   "should log error code message without fields",
			ctx:    context.Background(),
			code:   "TEST001",
			msg:    "test error message",
			fields: []zap.Field{},
		},
		{
			name: "should log error code message with fields",
			ctx:  context.Background(),
			code: "TEST002",
			msg:  "test error message with fields",
			fields: []zap.Field{
				zap.String("key", "value"),
				zap.Int("count", 1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger()
			logger.ErrorCode(tt.ctx, tt.code, tt.msg, tt.fields...)
		})
	}
}
func TestDebugOutCtx(t *testing.T) {
	tests := []struct {
		name   string
		msg    string
		fields []zap.Field
	}{
		{
			name:   "should log debug message without fields",
			msg:    "test debug message",
			fields: []zap.Field{},
		},
		{
			name: "should log debug message with fields",
			msg:  "test debug message with fields",
			fields: []zap.Field{
				zap.String("key", "value"),
				zap.Int("count", 1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DebugOutCtx(tt.msg, tt.fields...)
		})
	}
}

func TestWarnOutCtx(t *testing.T) {
	tests := []struct {
		name   string
		msg    string
		fields []zap.Field
	}{
		{
			name:   "should log warn message without fields",
			msg:    "test warn message",
			fields: []zap.Field{},
		},
		{
			name: "should log warn message with fields",
			msg:  "test warn message with fields",
			fields: []zap.Field{
				zap.String("key", "value"),
				zap.Int("count", 1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WarnOutCtx(tt.msg, tt.fields...)
		})
	}
}
func TestDebugf(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		format string
		args   []interface{}
	}{
		{
			name:   "should log formatted debug message without args",
			ctx:    context.Background(),
			format: "test message",
			args:   []interface{}{},
		},
		{
			name:   "should log formatted debug message with args",
			ctx:    context.Background(),
			format: "test message %s %d",
			args:   []interface{}{"value", 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Debugf(tt.ctx, tt.format, tt.args...)
		})
	}
}

func TestWarnf(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		format string
		args   []interface{}
	}{
		{
			name:   "should log formatted warn message without args",
			ctx:    context.Background(),
			format: "test message",
			args:   []interface{}{},
		},
		{
			name:   "should log formatted warn message with args",
			ctx:    context.Background(),
			format: "test message %s %d",
			args:   []interface{}{"value", 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Warnf(tt.ctx, tt.format, tt.args...)
		})
	}
}
func TestWithFields(t *testing.T) {
	tests := []struct {
		name   string
		fields map[string]interface{}
	}{
		{
			name:   "should set fields with empty map",
			fields: map[string]interface{}{},
		},
		{
			name: "should set fields with single field",
			fields: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name: "should set fields with multiple fields",
			fields: map[string]interface{}{
				"string": "value",
				"int":    42,
				"bool":   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger()
			result := logger.WithFields(tt.fields)

			if result == nil {
				t.Error("WithFields() returned nil logger")
			}

			if result.Zlg == nil {
				t.Error("WithFields() resulted in nil zap logger")
			}
		})
	}
}
func TestSync(t *testing.T) {
	logger := NewLogger()

	// Test should not panic
	logger.Sync()
}

func TestWithContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "should set context with background context",
			ctx:  context.Background(),
		},
		{
			name: "should set context with TODO context",
			ctx:  context.TODO(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger()
			result := logger.WithContext(tt.ctx)

			if result == nil {
				t.Error("WithContext() returned nil logger")
			}

			if result.Ctx != tt.ctx {
				t.Errorf("WithContext() did not set the expected context")
			}

			if result.Zlg == nil {
				t.Error("WithContext() resulted in nil zap logger")
			}
		})
	}
}
func TestNewLoggerWithEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name           string
		setupEnv       func()
		cleanupEnv     func()
		validateLogger func(*testing.T, *Logger)
	}{
		{
			name: "should create logger with default settings",
			setupEnv: func() {
				os.Clearenv()
			},
			cleanupEnv: func() {
				os.Clearenv()
			},
			validateLogger: func(t *testing.T, logger *Logger) {
				if logger == nil {
					t.Error("Expected logger to not be nil")
				}
				if logger.Level.Level() != zapcore.InfoLevel {
					t.Errorf("Expected default level to be InfoLevel, got %v", logger.Level.Level())
				}
				if logger.Config.Encoding != "json" {
					t.Errorf("Expected encoding to be json, got %s", logger.Config.Encoding)
				}
			},
		},
		{
			name: "should create logger with custom log level",
			setupEnv: func() {
				os.Setenv("LOG_LEVEL", "info")
			},
			cleanupEnv: func() {
				os.Unsetenv("LOG_LEVEL")
			},
			validateLogger: func(t *testing.T, logger *Logger) {
				if logger.Level.Level() != zapcore.InfoLevel {
					t.Errorf("Expected level to be InfoLevel, got %v", logger.Level.Level())
				}
			},
		},
		{
			name: "should create logger with trace enabled",
			setupEnv: func() {
				os.Setenv("DD_LOG_TRACER_ENABLED", "true")
			},
			cleanupEnv: func() {
				os.Unsetenv("DD_LOG_TRACER_ENABLED")
			},
			validateLogger: func(t *testing.T, logger *Logger) {
				if logger == nil {
					t.Error("Expected logger to not be nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			tt.setupEnv()

			// Reset the singleton instance
			instanceLgger = nil
			once = sync.Once{}

			// Create new logger
			logger := NewLogger()

			// Validate logger
			tt.validateLogger(t, logger)

			// Cleanup environment
			tt.cleanupEnv()
		})
	}
}

func TestNewLoggerSingleton(t *testing.T) {
	// Reset the singleton instance
	instanceLgger = nil
	once = sync.Once{}

	// Create first instance
	logger1 := NewLogger()
	if logger1 == nil {
		t.Error("Expected first logger instance to not be nil")
	}

	// Create second instance
	logger2 := NewLogger()
	if logger2 == nil {
		t.Error("Expected second logger instance to not be nil")
	}

	// Verify singleton behavior
	if logger1 != logger2 {
		t.Error("Expected both logger instances to be the same (singleton)")
	}
}
func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		setupEnv func()
		want     string
	}{
		{
			name: "should return ENV value when set",
			setupEnv: func() {
				os.Setenv("ENV", "prod")
			},
			want: "prod",
		},
		{
			name: "should return ENV_DEV when ENV not set",
			setupEnv: func() {
				os.Unsetenv("ENV")
			},
			want: ENV_DEV,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			tt.setupEnv()

			// Test
			got := GetEnv()

			// Verify
			if got != tt.want {
				t.Errorf("GetEnv() = %v, want %v", got, tt.want)
			}

			// Cleanup
			os.Unsetenv("ENV")
		})
	}
}
