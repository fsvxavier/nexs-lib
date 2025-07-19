package logrus

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger/interfaces"
	"github.com/sirupsen/logrus"
)

func BenchmarkProviderLogInfo(b *testing.B) {
	provider := NewProvider()
	provider.writer = io.Discard
	provider.logger.SetOutput(io.Discard)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		provider.Info(ctx, "benchmark message",
			interfaces.Field{Key: "iteration", Value: i},
			interfaces.Field{Key: "benchmark", Value: true},
		)
	}
}

func BenchmarkProviderLogDebug(b *testing.B) {
	provider := NewProvider()
	provider.writer = io.Discard
	provider.logger.SetOutput(io.Discard)
	provider.SetLevel(interfaces.DebugLevel)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		provider.Debug(ctx, "debug benchmark message",
			interfaces.Field{Key: "iteration", Value: i},
		)
	}
}

func BenchmarkProviderLogError(b *testing.B) {
	provider := NewProvider()
	provider.writer = io.Discard
	provider.logger.SetOutput(io.Discard)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		provider.Error(ctx, "error benchmark message",
			interfaces.Field{Key: "error_code", Value: "E001"},
			interfaces.Field{Key: "iteration", Value: i},
		)
	}
}

func BenchmarkProviderLogWithCode(b *testing.B) {
	provider := NewProvider()
	provider.writer = io.Discard
	provider.logger.SetOutput(io.Discard)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		provider.InfoWithCode(ctx, "INFO001", "info with code benchmark",
			interfaces.Field{Key: "iteration", Value: i},
		)
	}
}

func BenchmarkProviderFormattedLog(b *testing.B) {
	provider := NewProvider()
	provider.writer = io.Discard
	provider.logger.SetOutput(io.Discard)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		provider.Infof(ctx, "formatted message %d with %s", i, "benchmark")
	}
}

func BenchmarkProviderWithFields(b *testing.B) {
	provider := NewProvider()
	provider.writer = io.Discard
	provider.logger.SetOutput(io.Discard)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		enrichedLogger := provider.WithFields(
			interfaces.Field{Key: "request_id", Value: "req-123"},
			interfaces.Field{Key: "user_id", Value: "user-456"},
		)
		enrichedLogger.Info(ctx, "benchmark with fields")
	}
}

func BenchmarkProviderClone(b *testing.B) {
	provider := NewProvider()
	provider.fields["global"] = "value"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		clone := provider.Clone()
		_ = clone
	}
}

func BenchmarkProviderLevelConversion(b *testing.B) {
	provider := NewProvider()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		level := interfaces.Level(i % 6)
		logrusLevel := provider.levelToLogrus(level)
		backLevel := provider.logrusToLevel(logrusLevel)
		_ = backLevel
	}
}

func BenchmarkBufferWriter(b *testing.B) {
	provider := NewProvider()
	provider.writer = io.Discard
	bw := &bufferWriter{provider: provider}

	data := []byte(`{"time":"2023-01-01T00:00:00Z","level":"info","msg":"test message"}`)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		bw.Write(data)
	}
}

func BenchmarkNewProvider(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		provider := NewProvider()
		_ = provider
	}
}

func BenchmarkNewProviderWithConfig(b *testing.B) {
	config := &interfaces.Config{
		Level:  interfaces.InfoLevel,
		Format: interfaces.JSONFormat,
		Fields: map[string]any{
			"service": "benchmark-test",
			"version": "1.0.0",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		provider, _ := NewWithConfig(config)
		_ = provider
	}
}

func BenchmarkProviderConfigure(b *testing.B) {
	config := &interfaces.Config{
		Level:  interfaces.DebugLevel,
		Format: interfaces.JSONFormat,
		Fields: map[string]any{
			"app": "benchmark",
			"env": "test",
		},
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		provider := NewProvider()
		provider.Configure(config)
	}
}

func BenchmarkProviderSetLevel(b *testing.B) {
	provider := NewProvider()
	levels := []interfaces.Level{
		interfaces.DebugLevel,
		interfaces.InfoLevel,
		interfaces.WarnLevel,
		interfaces.ErrorLevel,
		interfaces.FatalLevel,
		interfaces.PanicLevel,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		provider.SetLevel(levels[i%len(levels)])
	}
}

func BenchmarkProviderGetLevel(b *testing.B) {
	provider := NewProvider()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		level := provider.GetLevel()
		_ = level
	}
}

func BenchmarkProviderBufferOperations(b *testing.B) {
	provider := NewProvider()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buffer := provider.GetBuffer()
		stats := provider.GetBufferStats()
		provider.FlushBuffer()
		_ = buffer
		_ = stats
	}
}

func BenchmarkProviderClose(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		provider := NewProvider()
		provider.Close()
	}
}

func BenchmarkLogrusHookAdapter(b *testing.B) {
	beforeHook := &benchmarkHook{}
	afterHook := &benchmarkHook{}
	adapter := NewLogrusHookAdapter(beforeHook, afterHook)

	entry := &logrus.Entry{
		Time:    time.Now(),
		Level:   logrus.InfoLevel,
		Message: "benchmark message",
		Data:    logrus.Fields{"key": "value"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		adapter.Fire(entry)
	}
}

func BenchmarkStringToLevel(b *testing.B) {
	provider := NewProvider()
	bw := &bufferWriter{provider: provider}

	levels := []string{"debug", "info", "warn", "error", "fatal", "panic", "unknown"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		level := bw.stringToLevel(levels[i%len(levels)])
		_ = level
	}
}

func BenchmarkLogrusEntryConversion(b *testing.B) {
	provider := NewProvider()
	bw := &bufferWriter{provider: provider}

	logrusEntry := map[string]interface{}{
		"time":  "2023-01-01T00:00:00Z",
		"level": "info",
		"msg":   "benchmark message",
		"code":  "B001",
		"extra": "data",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		entry := bw.logrusEntryToLogEntry(logrusEntry)
		_ = entry
	}
}

// benchmarkHook implementa interfaces.Hook para benchmarks
type benchmarkHook struct{}

func (h *benchmarkHook) Execute(ctx context.Context, entry *interfaces.LogEntry) error {
	return nil
}

func (h *benchmarkHook) GetName() string {
	return "benchmark-hook"
}

func (h *benchmarkHook) IsEnabled() bool {
	return true
}

func (h *benchmarkHook) SetEnabled(enabled bool) {
	// Not implemented for benchmark
}
