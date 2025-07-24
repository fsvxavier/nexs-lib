package logrus

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fsvxavier/nexs-lib/observability/logger/interfaces"
)

func TestNewProvider(t *testing.T) {
	t.Run("should create provider with default configuration", func(t *testing.T) {
		provider := NewProvider()

		assert.NotNil(t, provider)
		assert.NotNil(t, provider.logger)
		assert.Equal(t, interfaces.InfoLevel, provider.GetLevel())
		assert.NotNil(t, provider.fields)
	})
}

func TestNewProviderWithLogger(t *testing.T) {
	t.Run("should create provider with existing logrus logger", func(t *testing.T) {
		logrusLogger := logrus.New()
		logrusLogger.SetLevel(logrus.DebugLevel)

		provider := NewProviderWithLogger(logrusLogger)

		assert.NotNil(t, provider)
		assert.Equal(t, logrusLogger, provider.logger)
		assert.Equal(t, interfaces.DebugLevel, provider.GetLevel())
	})
}

func TestProviderConfigure(t *testing.T) {
	t.Run("should configure with nil config", func(t *testing.T) {
		provider := NewProvider()
		err := provider.Configure(nil)

		assert.NoError(t, err)
	})

	t.Run("should configure level", func(t *testing.T) {
		provider := NewProvider()
		config := &interfaces.Config{
			Level: interfaces.DebugLevel,
		}

		err := provider.Configure(config)

		assert.NoError(t, err)
		assert.Equal(t, interfaces.DebugLevel, provider.GetLevel())
	})

	t.Run("should configure format JSON", func(t *testing.T) {
		provider := NewProvider()
		config := &interfaces.Config{
			Format: interfaces.JSONFormat,
		}

		err := provider.Configure(config)

		assert.NoError(t, err)
		assert.IsType(t, &logrus.JSONFormatter{}, provider.logger.Formatter)
	})

	t.Run("should configure format Text", func(t *testing.T) {
		provider := NewProvider()
		config := &interfaces.Config{
			Format: interfaces.TextFormat,
		}

		err := provider.Configure(config)

		assert.NoError(t, err)
		assert.IsType(t, &logrus.TextFormatter{}, provider.logger.Formatter)
	})

	t.Run("should configure global fields", func(t *testing.T) {
		provider := NewProvider()
		config := &interfaces.Config{
			Fields: map[string]any{
				"app": "test",
				"env": "testing",
			},
		}

		err := provider.Configure(config)

		assert.NoError(t, err)
		assert.Equal(t, "test", provider.fields["app"])
		assert.Equal(t, "testing", provider.fields["env"])
	})

	t.Run("should configure service fields", func(t *testing.T) {
		provider := NewProvider()
		config := &interfaces.Config{
			ServiceName:    "test-service",
			ServiceVersion: "1.0.0",
			Environment:    "production",
		}

		err := provider.Configure(config)

		assert.NoError(t, err)
		assert.Equal(t, "test-service", provider.fields["service"])
		assert.Equal(t, "1.0.0", provider.fields["version"])
		assert.Equal(t, "production", provider.fields["env"])
	})
}

func TestProviderLogging(t *testing.T) {
	var buf bytes.Buffer
	provider := NewProvider()
	provider.writer = &buf
	provider.logger.SetOutput(&buf)
	provider.logger.SetFormatter(&logrus.JSONFormatter{})

	ctx := context.Background()

	t.Run("should log debug message", func(t *testing.T) {
		buf.Reset()
		provider.SetLevel(interfaces.DebugLevel)

		provider.Debug(ctx, "debug message",
			interfaces.Field{Key: "key1", Value: "value1"},
			interfaces.Field{Key: "key2", Value: 42},
		)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)

		assert.NoError(t, err)
		assert.Equal(t, "debug", logEntry["level"])
		assert.Equal(t, "debug message", logEntry["msg"])
		assert.Equal(t, "value1", logEntry["key1"])
		assert.Equal(t, float64(42), logEntry["key2"])
	})

	t.Run("should log info message", func(t *testing.T) {
		buf.Reset()

		provider.Info(ctx, "info message",
			interfaces.Field{Key: "user_id", Value: "123"},
		)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)

		assert.NoError(t, err)
		assert.Equal(t, "info", logEntry["level"])
		assert.Equal(t, "info message", logEntry["msg"])
		assert.Equal(t, "123", logEntry["user_id"])
	})

	t.Run("should log warn message", func(t *testing.T) {
		buf.Reset()

		provider.Warn(ctx, "warning message")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)

		assert.NoError(t, err)
		assert.Equal(t, "warning", logEntry["level"])
		assert.Equal(t, "warning message", logEntry["msg"])
	})

	t.Run("should log error message", func(t *testing.T) {
		buf.Reset()

		provider.Error(ctx, "error message",
			interfaces.Field{Key: "error_code", Value: "E001"},
		)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)

		assert.NoError(t, err)
		assert.Equal(t, "error", logEntry["level"])
		assert.Equal(t, "error message", logEntry["msg"])
		assert.Equal(t, "E001", logEntry["error_code"])
	})
}

func TestProviderFormattedLogging(t *testing.T) {
	var buf bytes.Buffer
	provider := NewProvider()
	provider.writer = &buf
	provider.logger.SetOutput(&buf)
	provider.logger.SetFormatter(&logrus.JSONFormatter{})

	ctx := context.Background()

	t.Run("should log formatted debug message", func(t *testing.T) {
		buf.Reset()
		provider.SetLevel(interfaces.DebugLevel)

		provider.Debugf(ctx, "debug message with %s and %d", "string", 42)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)

		assert.NoError(t, err)
		assert.Equal(t, "debug", logEntry["level"])
		assert.Equal(t, "debug message with string and 42", logEntry["msg"])
	})

	t.Run("should log formatted info message", func(t *testing.T) {
		buf.Reset()

		provider.Infof(ctx, "user %s logged in", "john")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)

		assert.NoError(t, err)
		assert.Equal(t, "info", logEntry["level"])
		assert.Equal(t, "user john logged in", logEntry["msg"])
	})
}

func TestProviderWithCode(t *testing.T) {
	var buf bytes.Buffer
	provider := NewProvider()
	provider.writer = &buf
	provider.logger.SetOutput(&buf)
	provider.logger.SetFormatter(&logrus.JSONFormatter{})

	ctx := context.Background()

	t.Run("should log debug message with code", func(t *testing.T) {
		buf.Reset()
		provider.SetLevel(interfaces.DebugLevel)

		provider.DebugWithCode(ctx, "D001", "debug with code",
			interfaces.Field{Key: "extra", Value: "info"},
		)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)

		assert.NoError(t, err)
		assert.Equal(t, "debug", logEntry["level"])
		assert.Equal(t, "debug with code", logEntry["msg"])
		assert.Equal(t, "D001", logEntry["code"])
		assert.Equal(t, "info", logEntry["extra"])
	})

	t.Run("should log error message with code", func(t *testing.T) {
		buf.Reset()

		provider.ErrorWithCode(ctx, "E001", "error with code")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)

		assert.NoError(t, err)
		assert.Equal(t, "error", logEntry["level"])
		assert.Equal(t, "error with code", logEntry["msg"])
		assert.Equal(t, "E001", logEntry["code"])
	})
}

func TestProviderWithFields(t *testing.T) {
	t.Run("should create logger with additional fields", func(t *testing.T) {
		provider := NewProvider()
		provider.fields["global"] = "value"

		newLogger := provider.WithFields(
			interfaces.Field{Key: "field1", Value: "value1"},
			interfaces.Field{Key: "field2", Value: 42},
		)

		newProvider := newLogger.(*Provider)
		assert.Equal(t, "value", newProvider.fields["global"])
		assert.Equal(t, "value1", newProvider.fields["field1"])
		assert.Equal(t, 42, newProvider.fields["field2"])

		// Provider original não deve ser modificado
		assert.NotContains(t, provider.fields, "field1")
		assert.NotContains(t, provider.fields, "field2")
	})
}

func TestProviderWithContext(t *testing.T) {
	t.Run("should return same provider with context", func(t *testing.T) {
		provider := NewProvider()
		ctx := context.WithValue(context.Background(), "key", "value")

		newLogger := provider.WithContext(ctx)

		// Logrus não suporta contexto nativamente desta forma,
		// então deve retornar a mesma instância
		assert.Equal(t, provider, newLogger)
	})
}

func TestProviderSetGetLevel(t *testing.T) {
	provider := NewProvider()

	testCases := []struct {
		name  string
		level interfaces.Level
	}{
		{"Debug", interfaces.DebugLevel},
		{"Info", interfaces.InfoLevel},
		{"Warn", interfaces.WarnLevel},
		{"Error", interfaces.ErrorLevel},
		{"Fatal", interfaces.FatalLevel},
		{"Panic", interfaces.PanicLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider.SetLevel(tc.level)
			assert.Equal(t, tc.level, provider.GetLevel())
		})
	}
}

func TestProviderClone(t *testing.T) {
	t.Run("should create independent clone", func(t *testing.T) {
		provider := NewProvider()
		provider.fields["original"] = "value"
		provider.SetLevel(interfaces.DebugLevel)

		clone := provider.Clone().(*Provider)

		// Deve ter os mesmos valores iniciais
		assert.Equal(t, provider.GetLevel(), clone.GetLevel())
		assert.Equal(t, "value", clone.fields["original"])

		// Modificações no clone não devem afetar o original
		clone.fields["clone_field"] = "clone_value"
		clone.SetLevel(interfaces.ErrorLevel)

		assert.NotContains(t, provider.fields, "clone_field")
		assert.Equal(t, interfaces.DebugLevel, provider.GetLevel())
		assert.Equal(t, interfaces.ErrorLevel, clone.GetLevel())
	})
}

func TestProviderClose(t *testing.T) {
	t.Run("should close without error when no buffer", func(t *testing.T) {
		provider := NewProvider()
		err := provider.Close()
		assert.NoError(t, err)
	})

	t.Run("should flush buffer on close", func(t *testing.T) {
		provider := NewProvider()
		err := provider.Close()
		assert.NoError(t, err)
	})
}

func TestProviderBuffer(t *testing.T) {
	provider := NewProvider()

	t.Run("should get nil buffer initially", func(t *testing.T) {
		buffer := provider.GetBuffer()
		assert.Nil(t, buffer)
	})

	t.Run("should get empty buffer stats when no buffer", func(t *testing.T) {
		stats := provider.GetBufferStats()
		assert.Equal(t, interfaces.BufferStats{}, stats)
	})

	t.Run("should flush buffer successfully when no buffer", func(t *testing.T) {
		err := provider.FlushBuffer()
		assert.NoError(t, err)
	})
}

func TestProviderLogrusSpecific(t *testing.T) {
	t.Run("should return underlying logrus logger", func(t *testing.T) {
		provider := NewProvider()
		logrusLogger := provider.GetLogrusLogger()

		assert.NotNil(t, logrusLogger)
		assert.Equal(t, provider.logger, logrusLogger)
	})

	t.Run("should add logrus hook", func(t *testing.T) {
		provider := NewProvider()

		// Hook simples para teste
		hook := &testHook{}
		provider.AddHook(hook)

		// Verifica se o hook foi adicionado
		hooks := provider.logger.Hooks
		assert.Contains(t, hooks[logrus.InfoLevel], hook)
	})
}

func TestLevelConversion(t *testing.T) {
	provider := NewProvider()

	testCases := []struct {
		interfaceLevel interfaces.Level
		logrusLevel    logrus.Level
	}{
		{interfaces.DebugLevel, logrus.DebugLevel},
		{interfaces.InfoLevel, logrus.InfoLevel},
		{interfaces.WarnLevel, logrus.WarnLevel},
		{interfaces.ErrorLevel, logrus.ErrorLevel},
		{interfaces.FatalLevel, logrus.FatalLevel},
		{interfaces.PanicLevel, logrus.PanicLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.interfaceLevel.String(), func(t *testing.T) {
			// Teste conversão interface para logrus
			logrusLevel := provider.levelToLogrus(tc.interfaceLevel)
			assert.Equal(t, tc.logrusLevel, logrusLevel)

			// Teste conversão logrus para interface
			interfaceLevel := provider.logrusToLevel(tc.logrusLevel)
			assert.Equal(t, tc.interfaceLevel, interfaceLevel)
		})
	}
}

func TestNewWithConfig(t *testing.T) {
	t.Run("should create provider with config", func(t *testing.T) {
		config := &interfaces.Config{
			Level:  interfaces.DebugLevel,
			Format: interfaces.JSONFormat,
		}

		provider, err := NewWithConfig(config)

		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, interfaces.DebugLevel, provider.GetLevel())
	})

	t.Run("should create provider with nil config", func(t *testing.T) {
		provider, err := NewWithConfig(nil)

		require.NoError(t, err)
		assert.NotNil(t, provider)
	})
}

func TestNewWithWriter(t *testing.T) {
	t.Run("should create provider with custom writer", func(t *testing.T) {
		var buf bytes.Buffer
		provider := NewWithWriter(&buf)

		assert.NotNil(t, provider)
		assert.Equal(t, &buf, provider.writer)
	})
}

func TestFactoryMethods(t *testing.T) {
	t.Run("should create text provider", func(t *testing.T) {
		provider := NewTextProvider()

		assert.NotNil(t, provider)
		assert.IsType(t, &logrus.TextFormatter{}, provider.logger.Formatter)
	})

	t.Run("should create JSON provider", func(t *testing.T) {
		provider := NewJSONProvider()

		assert.NotNil(t, provider)
		assert.IsType(t, &logrus.JSONFormatter{}, provider.logger.Formatter)
	})

	t.Run("should create console provider", func(t *testing.T) {
		provider := NewConsoleProvider()

		assert.NotNil(t, provider)
	})

	t.Run("should create buffered provider", func(t *testing.T) {
		provider := NewBufferedProvider(1000, 5*time.Second)

		assert.NotNil(t, provider)
	})
}

// testHook implementa logrus.Hook para testes
type testHook struct{}

func (h *testHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *testHook) Fire(entry *logrus.Entry) error {
	return nil
}
