package logger

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Validações básicas
	if config.Level != interfaces.InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", config.Level)
	}

	if config.Format != interfaces.JSONFormat {
		t.Errorf("Expected JSONFormat, got %v", config.Format)
	}

	if config.Output != os.Stdout {
		t.Errorf("Expected os.Stdout, got %v", config.Output)
	}

	if config.TimeFormat != time.RFC3339Nano {
		t.Errorf("Expected RFC3339Nano, got %s", config.TimeFormat)
	}

	if config.ServiceName != "unknown" {
		t.Errorf("Expected 'unknown', got %s", config.ServiceName)
	}

	if config.ServiceVersion != "unknown" {
		t.Errorf("Expected 'unknown', got %s", config.ServiceVersion)
	}

	if config.Environment != "development" {
		t.Errorf("Expected 'development', got %s", config.Environment)
	}

	if config.AddSource != false {
		t.Errorf("Expected false for AddSource, got %v", config.AddSource)
	}

	if config.AddStacktrace != false {
		t.Errorf("Expected false for AddStacktrace, got %v", config.AddStacktrace)
	}

	if config.AddCaller != true {
		t.Errorf("Expected true for AddCaller, got %v", config.AddCaller)
	}

	if config.GlobalFields == nil {
		t.Error("Expected GlobalFields to be initialized")
	}

	if config.BufferSize != 1024 {
		t.Errorf("Expected BufferSize 1024, got %d", config.BufferSize)
	}

	if config.FlushInterval != 100*time.Millisecond {
		t.Errorf("Expected FlushInterval 100ms, got %v", config.FlushInterval)
	}

	if config.EnableMetrics != false {
		t.Errorf("Expected EnableMetrics false, got %v", config.EnableMetrics)
	}

	if config.MetricsPrefix != "logger" {
		t.Errorf("Expected MetricsPrefix 'logger', got %s", config.MetricsPrefix)
	}
}

func TestProductionConfig(t *testing.T) {
	config := ProductionConfig()

	// Validações específicas para produção
	if config.Level != interfaces.InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", config.Level)
	}

	if config.Format != interfaces.JSONFormat {
		t.Errorf("Expected JSONFormat, got %v", config.Format)
	}

	if config.Environment != "production" {
		t.Errorf("Expected 'production', got %s", config.Environment)
	}

	if config.AddSource != false {
		t.Errorf("Expected false for AddSource, got %v", config.AddSource)
	}

	if config.AddStacktrace != true {
		t.Errorf("Expected true for AddStacktrace, got %v", config.AddStacktrace)
	}

	if config.AddCaller != false {
		t.Errorf("Expected false for AddCaller, got %v", config.AddCaller)
	}

	if config.EnableMetrics != true {
		t.Errorf("Expected EnableMetrics true, got %v", config.EnableMetrics)
	}

	// Validações de configuração assíncrona
	if config.Async == nil {
		t.Fatal("Expected Async config to be set")
	}

	if !config.Async.Enabled {
		t.Error("Expected Async to be enabled")
	}

	if config.Async.BufferSize != 4096 {
		t.Errorf("Expected Async BufferSize 4096, got %d", config.Async.BufferSize)
	}

	if config.Async.FlushInterval != 50*time.Millisecond {
		t.Errorf("Expected Async FlushInterval 50ms, got %v", config.Async.FlushInterval)
	}

	if config.Async.Workers != 2 {
		t.Errorf("Expected Async Workers 2, got %d", config.Async.Workers)
	}

	if config.Async.DropOnFull != false {
		t.Errorf("Expected Async DropOnFull false, got %v", config.Async.DropOnFull)
	}

	// Validações de configuração de sampling
	if config.Sampling == nil {
		t.Fatal("Expected Sampling config to be set")
	}

	if !config.Sampling.Enabled {
		t.Error("Expected Sampling to be enabled")
	}

	if config.Sampling.Initial != 100 {
		t.Errorf("Expected Sampling Initial 100, got %d", config.Sampling.Initial)
	}

	if config.Sampling.Thereafter != 1000 {
		t.Errorf("Expected Sampling Thereafter 1000, got %d", config.Sampling.Thereafter)
	}

	if config.Sampling.Tick != 1*time.Second {
		t.Errorf("Expected Sampling Tick 1s, got %v", config.Sampling.Tick)
	}

	expectedLevels := []interfaces.Level{interfaces.DebugLevel, interfaces.TraceLevel}
	if len(config.Sampling.Levels) != len(expectedLevels) {
		t.Errorf("Expected %d sampling levels, got %d", len(expectedLevels), len(config.Sampling.Levels))
	}

	for i, level := range expectedLevels {
		if i >= len(config.Sampling.Levels) || config.Sampling.Levels[i] != level {
			t.Errorf("Expected sampling level %v at index %d, got %v", level, i, config.Sampling.Levels[i])
		}
	}
}

func TestDevelopmentConfig(t *testing.T) {
	config := DevelopmentConfig()

	// Validações específicas para desenvolvimento
	if config.Level != interfaces.DebugLevel {
		t.Errorf("Expected DebugLevel, got %v", config.Level)
	}

	if config.Format != interfaces.ConsoleFormat {
		t.Errorf("Expected ConsoleFormat, got %v", config.Format)
	}

	if config.Environment != "development" {
		t.Errorf("Expected 'development', got %s", config.Environment)
	}

	if config.AddSource != true {
		t.Errorf("Expected true for AddSource, got %v", config.AddSource)
	}

	if config.AddStacktrace != false {
		t.Errorf("Expected false for AddStacktrace, got %v", config.AddStacktrace)
	}

	if config.AddCaller != true {
		t.Errorf("Expected true for AddCaller, got %v", config.AddCaller)
	}

	if config.EnableMetrics != false {
		t.Errorf("Expected EnableMetrics false, got %v", config.EnableMetrics)
	}

	// Configuração assíncrona deve estar desabilitada
	if config.Async != nil && config.Async.Enabled {
		t.Error("Expected Async to be disabled in development")
	}
}

func TestTestConfig(t *testing.T) {
	config := TestConfig()

	// Validações específicas para teste
	if config.Level != interfaces.DebugLevel {
		t.Errorf("Expected DebugLevel, got %v", config.Level)
	}

	if config.Format != interfaces.JSONFormat {
		t.Errorf("Expected JSONFormat, got %v", config.Format)
	}

	if config.Environment != "test" {
		t.Errorf("Expected 'test', got %s", config.Environment)
	}

	if config.AddSource != false {
		t.Errorf("Expected false for AddSource, got %v", config.AddSource)
	}

	if config.AddStacktrace != false {
		t.Errorf("Expected false for AddStacktrace, got %v", config.AddStacktrace)
	}

	if config.AddCaller != false {
		t.Errorf("Expected false for AddCaller, got %v", config.AddCaller)
	}

	if config.EnableMetrics != false {
		t.Errorf("Expected EnableMetrics false, got %v", config.EnableMetrics)
	}

	// Output deve ser descartado
	if config.Output == nil {
		t.Error("Expected Output to be set")
	}
}

func TestConfigWithCustomValues(t *testing.T) {
	// Testa criação de config com valores customizados
	buffer := &bytes.Buffer{}

	config := DefaultConfig()
	config.ServiceName = "test-service"
	config.ServiceVersion = "v1.0.0"
	config.Environment = "staging"
	config.Output = buffer
	config.Level = interfaces.WarnLevel
	config.Format = interfaces.TextFormat
	config.AddSource = true
	config.AddStacktrace = true
	config.AddCaller = false
	config.BufferSize = 2048
	config.FlushInterval = 200 * time.Millisecond
	config.EnableMetrics = true
	config.MetricsPrefix = "custom"

	// Valida as alterações
	if config.ServiceName != "test-service" {
		t.Errorf("Expected 'test-service', got %s", config.ServiceName)
	}

	if config.ServiceVersion != "v1.0.0" {
		t.Errorf("Expected 'v1.0.0', got %s", config.ServiceVersion)
	}

	if config.Environment != "staging" {
		t.Errorf("Expected 'staging', got %s", config.Environment)
	}

	if config.Output != buffer {
		t.Error("Expected custom buffer as output")
	}

	if config.Level != interfaces.WarnLevel {
		t.Errorf("Expected WarnLevel, got %v", config.Level)
	}

	if config.Format != interfaces.TextFormat {
		t.Errorf("Expected TextFormat, got %v", config.Format)
	}

	if config.AddSource != true {
		t.Errorf("Expected true for AddSource, got %v", config.AddSource)
	}

	if config.AddStacktrace != true {
		t.Errorf("Expected true for AddStacktrace, got %v", config.AddStacktrace)
	}

	if config.AddCaller != false {
		t.Errorf("Expected false for AddCaller, got %v", config.AddCaller)
	}

	if config.BufferSize != 2048 {
		t.Errorf("Expected BufferSize 2048, got %d", config.BufferSize)
	}

	if config.FlushInterval != 200*time.Millisecond {
		t.Errorf("Expected FlushInterval 200ms, got %v", config.FlushInterval)
	}

	if config.EnableMetrics != true {
		t.Errorf("Expected EnableMetrics true, got %v", config.EnableMetrics)
	}

	if config.MetricsPrefix != "custom" {
		t.Errorf("Expected MetricsPrefix 'custom', got %s", config.MetricsPrefix)
	}
}

func TestConfigWithAsyncSettings(t *testing.T) {
	config := DefaultConfig()

	// Configuração assíncrona customizada
	config.Async = &interfaces.AsyncConfig{
		Enabled:       true,
		BufferSize:    8192,
		FlushInterval: 25 * time.Millisecond,
		Workers:       4,
		DropOnFull:    true,
	}

	if !config.Async.Enabled {
		t.Error("Expected Async to be enabled")
	}

	if config.Async.BufferSize != 8192 {
		t.Errorf("Expected Async BufferSize 8192, got %d", config.Async.BufferSize)
	}

	if config.Async.FlushInterval != 25*time.Millisecond {
		t.Errorf("Expected Async FlushInterval 25ms, got %v", config.Async.FlushInterval)
	}

	if config.Async.Workers != 4 {
		t.Errorf("Expected Async Workers 4, got %d", config.Async.Workers)
	}

	if config.Async.DropOnFull != true {
		t.Errorf("Expected Async DropOnFull true, got %v", config.Async.DropOnFull)
	}
}

func TestConfigWithSamplingSettings(t *testing.T) {
	config := DefaultConfig()

	// Configuração de sampling customizada
	config.Sampling = &interfaces.SamplingConfig{
		Enabled:    true,
		Initial:    50,
		Thereafter: 500,
		Tick:       2 * time.Second,
		Levels:     []interfaces.Level{interfaces.DebugLevel, interfaces.TraceLevel, interfaces.InfoLevel},
	}

	if !config.Sampling.Enabled {
		t.Error("Expected Sampling to be enabled")
	}

	if config.Sampling.Initial != 50 {
		t.Errorf("Expected Sampling Initial 50, got %d", config.Sampling.Initial)
	}

	if config.Sampling.Thereafter != 500 {
		t.Errorf("Expected Sampling Thereafter 500, got %d", config.Sampling.Thereafter)
	}

	if config.Sampling.Tick != 2*time.Second {
		t.Errorf("Expected Sampling Tick 2s, got %v", config.Sampling.Tick)
	}

	expectedLevels := []interfaces.Level{interfaces.DebugLevel, interfaces.TraceLevel, interfaces.InfoLevel}
	if len(config.Sampling.Levels) != len(expectedLevels) {
		t.Errorf("Expected %d sampling levels, got %d", len(expectedLevels), len(config.Sampling.Levels))
	}

	for i, level := range expectedLevels {
		if config.Sampling.Levels[i] != level {
			t.Errorf("Expected sampling level %v at index %d, got %v", level, i, config.Sampling.Levels[i])
		}
	}
}

func TestConfigWithGlobalFields(t *testing.T) {
	config := DefaultConfig()

	// Adiciona campos globais
	config.GlobalFields = map[string]interface{}{
		"service":     "test-service",
		"version":     "v1.0.0",
		"environment": "test",
		"region":      "us-east-1",
		"instance":    "i-123456789",
		"process_id":  12345,
		"enabled":     true,
		"ratio":       0.95,
	}

	if len(config.GlobalFields) != 8 {
		t.Errorf("Expected 8 global fields, got %d", len(config.GlobalFields))
	}

	// Verifica tipos e valores
	if config.GlobalFields["service"] != "test-service" {
		t.Errorf("Expected service 'test-service', got %v", config.GlobalFields["service"])
	}

	if config.GlobalFields["version"] != "v1.0.0" {
		t.Errorf("Expected version 'v1.0.0', got %v", config.GlobalFields["version"])
	}

	if config.GlobalFields["environment"] != "test" {
		t.Errorf("Expected environment 'test', got %v", config.GlobalFields["environment"])
	}

	if config.GlobalFields["region"] != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got %v", config.GlobalFields["region"])
	}

	if config.GlobalFields["instance"] != "i-123456789" {
		t.Errorf("Expected instance 'i-123456789', got %v", config.GlobalFields["instance"])
	}

	if config.GlobalFields["process_id"] != 12345 {
		t.Errorf("Expected process_id 12345, got %v", config.GlobalFields["process_id"])
	}

	if config.GlobalFields["enabled"] != true {
		t.Errorf("Expected enabled true, got %v", config.GlobalFields["enabled"])
	}

	if config.GlobalFields["ratio"] != 0.95 {
		t.Errorf("Expected ratio 0.95, got %v", config.GlobalFields["ratio"])
	}
}

func TestConfigTimeFormats(t *testing.T) {
	testCases := []struct {
		name       string
		timeFormat string
	}{
		{"RFC3339", time.RFC3339},
		{"RFC3339Nano", time.RFC3339Nano},
		{"RFC822", time.RFC822},
		{"RFC822Z", time.RFC822Z},
		{"RFC850", time.RFC850},
		{"RFC1123", time.RFC1123},
		{"RFC1123Z", time.RFC1123Z},
		{"Kitchen", time.Kitchen},
		{"Stamp", time.Stamp},
		{"StampMilli", time.StampMilli},
		{"StampMicro", time.StampMicro},
		{"StampNano", time.StampNano},
		{"DateTime", time.DateTime},
		{"DateOnly", time.DateOnly},
		{"TimeOnly", time.TimeOnly},
		{"Custom", "2006-01-02 15:04:05.000"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultConfig()
			config.TimeFormat = tc.timeFormat

			if config.TimeFormat != tc.timeFormat {
				t.Errorf("Expected TimeFormat %s, got %s", tc.timeFormat, config.TimeFormat)
			}

			// Testa se o formato é válido fazendo parse
			now := time.Now()
			formatted := now.Format(tc.timeFormat)
			if formatted == "" {
				t.Errorf("TimeFormat %s produced empty string", tc.timeFormat)
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	// Testa configurações inválidas
	invalidConfigs := []struct {
		name   string
		config interfaces.Config
	}{
		{
			name: "Empty service name",
			config: func() interfaces.Config {
				c := DefaultConfig()
				c.ServiceName = ""
				return c
			}(),
		},
		{
			name: "Invalid buffer size",
			config: func() interfaces.Config {
				c := DefaultConfig()
				c.BufferSize = -1
				return c
			}(),
		},
		{
			name: "Invalid flush interval",
			config: func() interfaces.Config {
				c := DefaultConfig()
				c.FlushInterval = -1
				return c
			}(),
		},
	}

	for _, tc := range invalidConfigs {
		t.Run(tc.name, func(t *testing.T) {
			err := interfaces.ValidateConfig(tc.config)
			if err == nil {
				t.Errorf("Expected validation error for %s", tc.name)
			}
		})
	}
}
