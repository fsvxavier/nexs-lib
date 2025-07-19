package logger

import (
	"context"
	"testing"

	"github.com/fsvxavier/nexs-lib/observability/logger/interfaces"
)

func BenchmarkObservableLogger(b *testing.B) {
	// Cria logger observável
	mockProvider := &mockProvider{logs: make([]logEntry, 0, b.N)}
	logger := NewObservableLogger(mockProvider)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		logger.Info(ctx, "benchmark test message",
			String("iteration", string(rune(i))),
			Int("number", i))
	}
}

func BenchmarkMetricsCollector(b *testing.B) {
	collector := NewMetricsCollector()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		collector.RecordLog(interfaces.InfoLevel, 0)
	}
}

func BenchmarkHookExecution(b *testing.B) {
	manager := NewHookManager()

	// Registra alguns hooks simples
	hook1 := NewTransformHook(func(entry *interfaces.LogEntry) error {
		entry.Message = entry.Message + "."
		return nil
	})
	hook2 := NewValidationHook(func(entry *interfaces.LogEntry) error {
		return nil // validação trivial
	})

	manager.RegisterHook(interfaces.BeforeHook, hook1)
	manager.RegisterHook(interfaces.BeforeHook, hook2)

	entry := &interfaces.LogEntry{
		Level:   interfaces.InfoLevel,
		Message: "benchmark message",
		Fields:  make(map[string]any),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		manager.ExecuteBeforeHooks(context.Background(), entry)
	}
}

// Helpers para benchmarks
type logEntry struct {
	level   interfaces.Level
	message string
	fields  []interfaces.Field
}

type mockProvider struct {
	logs []logEntry
}

func (m *mockProvider) Debug(ctx context.Context, msg string, fields ...interfaces.Field) {
	m.logs = append(m.logs, logEntry{interfaces.DebugLevel, msg, fields})
}

func (m *mockProvider) Info(ctx context.Context, msg string, fields ...interfaces.Field) {
	m.logs = append(m.logs, logEntry{interfaces.InfoLevel, msg, fields})
}

func (m *mockProvider) Warn(ctx context.Context, msg string, fields ...interfaces.Field) {
	m.logs = append(m.logs, logEntry{interfaces.WarnLevel, msg, fields})
}

func (m *mockProvider) Error(ctx context.Context, msg string, fields ...interfaces.Field) {
	m.logs = append(m.logs, logEntry{interfaces.ErrorLevel, msg, fields})
}

func (m *mockProvider) Fatal(ctx context.Context, msg string, fields ...interfaces.Field) {
	m.logs = append(m.logs, logEntry{interfaces.FatalLevel, msg, fields})
}

func (m *mockProvider) Panic(ctx context.Context, msg string, fields ...interfaces.Field) {
	m.logs = append(m.logs, logEntry{interfaces.PanicLevel, msg, fields})
}

func (m *mockProvider) Debugf(ctx context.Context, format string, args ...any) {}
func (m *mockProvider) Infof(ctx context.Context, format string, args ...any)  {}
func (m *mockProvider) Warnf(ctx context.Context, format string, args ...any)  {}
func (m *mockProvider) Errorf(ctx context.Context, format string, args ...any) {}
func (m *mockProvider) Fatalf(ctx context.Context, format string, args ...any) {}
func (m *mockProvider) Panicf(ctx context.Context, format string, args ...any) {}

func (m *mockProvider) DebugWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
}
func (m *mockProvider) InfoWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
}
func (m *mockProvider) WarnWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
}
func (m *mockProvider) ErrorWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
}

func (m *mockProvider) WithFields(fields ...interfaces.Field) interfaces.Logger { return m }
func (m *mockProvider) WithContext(ctx context.Context) interfaces.Logger       { return m }
func (m *mockProvider) SetLevel(level interfaces.Level)                         {}
func (m *mockProvider) GetLevel() interfaces.Level                              { return interfaces.InfoLevel }
func (m *mockProvider) Clone() interfaces.Logger                                { return m }
func (m *mockProvider) Close() error                                            { return nil }

func (m *mockProvider) Configure(config *interfaces.Config) error { return nil }
func (m *mockProvider) GetBuffer() interfaces.Buffer              { return nil }
func (m *mockProvider) SetBuffer(buffer interfaces.Buffer) error  { return nil }
func (m *mockProvider) FlushBuffer() error                        { return nil }
func (m *mockProvider) GetBufferStats() interfaces.BufferStats    { return interfaces.BufferStats{} }
