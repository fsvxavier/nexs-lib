package interfaces

import (
	"testing"
	"time"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{FatalLevel, "FATAL"},
		{PanicLevel, "PANIC"},
		{Level(99), "UNKNOWN"}, // Valor inválido
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("Level(%d).String() = %s, want %s", tt.level, result, tt.expected)
			}
		})
	}
}

func TestFormat_Constants(t *testing.T) {
	tests := []struct {
		format   Format
		expected string
	}{
		{JSONFormat, "json"},
		{ConsoleFormat, "console"},
		{TextFormat, "text"},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			if string(tt.format) != tt.expected {
				t.Errorf("Format constant %s = %s, want %s", tt.format, string(tt.format), tt.expected)
			}
		})
	}
}

func TestContextKey_Constants(t *testing.T) {
	tests := []struct {
		key      ContextKey
		expected string
	}{
		{TraceIDKey, "trace_id"},
		{SpanIDKey, "span_id"},
		{UserIDKey, "user_id"},
		{RequestIDKey, "request_id"},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			if string(tt.key) != tt.expected {
				t.Errorf("ContextKey constant %s = %s, want %s", tt.key, string(tt.key), tt.expected)
			}
		})
	}
}

func TestField_Structure(t *testing.T) {
	field := Field{
		Key:   "test_key",
		Value: "test_value",
	}

	if field.Key != "test_key" {
		t.Errorf("Field.Key = %s, want test_key", field.Key)
	}

	if field.Value != "test_value" {
		t.Errorf("Field.Value = %v, want test_value", field.Value)
	}
}

func TestConfig_Structure(t *testing.T) {
	config := &Config{
		Level:          InfoLevel,
		Format:         JSONFormat,
		TimeFormat:     time.RFC3339,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		AddSource:      true,
		AddStacktrace:  false,
		Fields:         map[string]any{"key": "value"},
	}

	// Verifica campos básicos
	if config.Level != InfoLevel {
		t.Errorf("Config.Level = %v, want %v", config.Level, InfoLevel)
	}

	if config.Format != JSONFormat {
		t.Errorf("Config.Format = %v, want %v", config.Format, JSONFormat)
	}

	if config.ServiceName != "test-service" {
		t.Errorf("Config.ServiceName = %s, want test-service", config.ServiceName)
	}

	if config.ServiceVersion != "1.0.0" {
		t.Errorf("Config.ServiceVersion = %s, want 1.0.0", config.ServiceVersion)
	}

	if config.Environment != "test" {
		t.Errorf("Config.Environment = %s, want test", config.Environment)
	}

	if !config.AddSource {
		t.Error("Config.AddSource = false, want true")
	}

	if config.AddStacktrace {
		t.Error("Config.AddStacktrace = true, want false")
	}

	if config.Fields["key"] != "value" {
		t.Errorf("Config.Fields[key] = %v, want value", config.Fields["key"])
	}
}

func TestSamplingConfig_Structure(t *testing.T) {
	sampling := &SamplingConfig{
		Initial:    100,
		Thereafter: 10,
		Tick:       time.Second,
	}

	if sampling.Initial != 100 {
		t.Errorf("SamplingConfig.Initial = %d, want 100", sampling.Initial)
	}

	if sampling.Thereafter != 10 {
		t.Errorf("SamplingConfig.Thereafter = %d, want 10", sampling.Thereafter)
	}

	if sampling.Tick != time.Second {
		t.Errorf("SamplingConfig.Tick = %v, want %v", sampling.Tick, time.Second)
	}
}

func TestBufferConfig_Structure(t *testing.T) {
	buffer := &BufferConfig{
		Enabled:      true,
		Size:         1000,
		BatchSize:    100,
		FlushTimeout: 5 * time.Second,
		MemoryLimit:  1024 * 1024,
		AutoFlush:    true,
		ForceSync:    false,
	}

	if !buffer.Enabled {
		t.Error("BufferConfig.Enabled = false, want true")
	}

	if buffer.Size != 1000 {
		t.Errorf("BufferConfig.Size = %d, want 1000", buffer.Size)
	}

	if buffer.BatchSize != 100 {
		t.Errorf("BufferConfig.BatchSize = %d, want 100", buffer.BatchSize)
	}

	if buffer.FlushTimeout != 5*time.Second {
		t.Errorf("BufferConfig.FlushTimeout = %v, want %v", buffer.FlushTimeout, 5*time.Second)
	}

	if buffer.MemoryLimit != 1024*1024 {
		t.Errorf("BufferConfig.MemoryLimit = %d, want %d", buffer.MemoryLimit, 1024*1024)
	}

	if !buffer.AutoFlush {
		t.Error("BufferConfig.AutoFlush = false, want true")
	}

	if buffer.ForceSync {
		t.Error("BufferConfig.ForceSync = true, want false")
	}
}

func TestBufferStats_Structure(t *testing.T) {
	now := time.Now()
	stats := &BufferStats{
		TotalEntries:   1000,
		DroppedEntries: 10,
		FlushCount:     50,
		BufferSize:     100,
		UsedSize:       75,
		LastFlush:      now,
		MemoryUsage:    1024,
		FlushDuration:  100 * time.Millisecond,
	}

	if stats.TotalEntries != 1000 {
		t.Errorf("BufferStats.TotalEntries = %d, want 1000", stats.TotalEntries)
	}

	if stats.DroppedEntries != 10 {
		t.Errorf("BufferStats.DroppedEntries = %d, want 10", stats.DroppedEntries)
	}

	if stats.FlushCount != 50 {
		t.Errorf("BufferStats.FlushCount = %d, want 50", stats.FlushCount)
	}

	if stats.BufferSize != 100 {
		t.Errorf("BufferStats.BufferSize = %d, want 100", stats.BufferSize)
	}

	if stats.UsedSize != 75 {
		t.Errorf("BufferStats.UsedSize = %d, want 75", stats.UsedSize)
	}

	if !stats.LastFlush.Equal(now) {
		t.Errorf("BufferStats.LastFlush = %v, want %v", stats.LastFlush, now)
	}

	if stats.MemoryUsage != 1024 {
		t.Errorf("BufferStats.MemoryUsage = %d, want 1024", stats.MemoryUsage)
	}

	if stats.FlushDuration != 100*time.Millisecond {
		t.Errorf("BufferStats.FlushDuration = %v, want %v", stats.FlushDuration, 100*time.Millisecond)
	}
}

func TestLogEntry_Structure(t *testing.T) {
	now := time.Now()
	entry := &LogEntry{
		Timestamp: now,
		Level:     InfoLevel,
		Message:   "test message",
		Fields:    map[string]any{"key": "value"},
		Code:      "TEST_001",
		Source:    "test.go:123",
		Stack:     "stack trace here",
		Size:      256,
	}

	if !entry.Timestamp.Equal(now) {
		t.Errorf("LogEntry.Timestamp = %v, want %v", entry.Timestamp, now)
	}

	if entry.Level != InfoLevel {
		t.Errorf("LogEntry.Level = %v, want %v", entry.Level, InfoLevel)
	}

	if entry.Message != "test message" {
		t.Errorf("LogEntry.Message = %s, want test message", entry.Message)
	}

	if entry.Fields["key"] != "value" {
		t.Errorf("LogEntry.Fields[key] = %v, want value", entry.Fields["key"])
	}

	if entry.Code != "TEST_001" {
		t.Errorf("LogEntry.Code = %s, want TEST_001", entry.Code)
	}

	if entry.Source != "test.go:123" {
		t.Errorf("LogEntry.Source = %s, want test.go:123", entry.Source)
	}

	if entry.Stack != "stack trace here" {
		t.Errorf("LogEntry.Stack = %s, want stack trace here", entry.Stack)
	}

	if entry.Size != 256 {
		t.Errorf("LogEntry.Size = %d, want 256", entry.Size)
	}
}

func TestLevel_Values(t *testing.T) {
	// Verifica se os valores dos níveis estão corretos
	if DebugLevel != -1 {
		t.Errorf("DebugLevel = %d, want -1", DebugLevel)
	}

	if InfoLevel != 0 {
		t.Errorf("InfoLevel = %d, want 0", InfoLevel)
	}

	if WarnLevel != 1 {
		t.Errorf("WarnLevel = %d, want 1", WarnLevel)
	}

	if ErrorLevel != 2 {
		t.Errorf("ErrorLevel = %d, want 2", ErrorLevel)
	}

	if FatalLevel != 3 {
		t.Errorf("FatalLevel = %d, want 3", FatalLevel)
	}

	if PanicLevel != 4 {
		t.Errorf("PanicLevel = %d, want 4", PanicLevel)
	}
}

func TestLevel_Order(t *testing.T) {
	// Verifica se a ordem dos níveis está correta (menor = mais verboso)
	levels := []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel, PanicLevel}

	for i := 1; i < len(levels); i++ {
		if levels[i-1] >= levels[i] {
			t.Errorf("Level order incorrect: %s (%d) should be less than %s (%d)",
				levels[i-1].String(), levels[i-1],
				levels[i].String(), levels[i])
		}
	}
}
