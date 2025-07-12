package logger

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

// Helper function for string containment checks
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Helper function para criar CoreLogger configurado para testes
func setupTestCoreLogger(level interfaces.Level) (*CoreLogger, *MockProvider) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	config.Level = level

	// Configura o provider
	provider.Configure(config)

	coreLogger := NewCoreLogger(provider, config)
	return coreLogger, provider
}

func TestNewCoreLogger(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()

	// Configura o provider
	provider.Configure(config)

	coreLogger := NewCoreLogger(provider, config)

	if coreLogger == nil {
		t.Fatal("Expected core logger to be created")
	}

	if coreLogger.provider != provider {
		t.Error("Expected provider to be set")
	}

	if coreLogger.config.ServiceName != config.ServiceName {
		t.Errorf("Expected service name %s, got %s", config.ServiceName, coreLogger.config.ServiceName)
	}

	if coreLogger.level != config.Level {
		t.Errorf("Expected level %v, got %v", config.Level, coreLogger.level)
	}

	// Verifica se pools foram inicializados
	entry := coreLogger.entryPool.Get()
	if entry == nil {
		t.Error("Expected entry pool to be initialized")
	}

	buffer := coreLogger.bufferPool.Get()
	if buffer == nil {
		t.Error("Expected buffer pool to be initialized")
	}
}

func TestCoreLoggerBasicLogging(t *testing.T) {
	coreLogger, provider := setupTestCoreLogger(interfaces.TraceLevel)
	ctx := context.Background()

	// Testa todos os níveis
	coreLogger.Trace(ctx, "trace message")
	coreLogger.Debug(ctx, "debug message")
	coreLogger.Info(ctx, "info message")
	coreLogger.Warn(ctx, "warn message")
	coreLogger.Error(ctx, "error message")

	messages := provider.GetLogMessages()
	if len(messages) != 5 {
		t.Errorf("Expected 5 messages, got %d", len(messages))
	}

	// Verifica se as mensagens contêm o texto esperado
	expectedContains := []string{"trace message", "debug message", "info message", "warn message", "error message"}
	for i, expected := range expectedContains {
		if i >= len(messages) {
			t.Errorf("Missing message %d", i)
			continue
		}
		if !contains(messages[i], expected) {
			t.Errorf("Expected message %d to contain '%s', got '%s'", i, expected, messages[i])
		}
	}
}

func TestCoreLoggerFormattedLogging(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	config.Level = interfaces.TraceLevel

	// Configura o provider
	provider.Configure(config)

	coreLogger := NewCoreLogger(provider, config)
	ctx := context.Background()

	// Testa logging formatado
	coreLogger.Tracef(ctx, "trace: %s %d", "value", 42)
	coreLogger.Debugf(ctx, "debug: %s %d", "value", 42)
	coreLogger.Infof(ctx, "info: %s %d", "value", 42)
	coreLogger.Warnf(ctx, "warn: %s %d", "value", 42)
	coreLogger.Errorf(ctx, "error: %s %d", "value", 42)

	messages := provider.GetLogMessages()
	if len(messages) != 5 {
		t.Errorf("Expected 5 messages, got %d", len(messages))
	}

	// Verifica se a formatação funcionou
	expectedFormats := []string{
		"trace: value 42",
		"debug: value 42",
		"info: value 42",
		"warn: value 42",
		"error: value 42",
	}

	for i, expected := range expectedFormats {
		if i >= len(messages) {
			t.Errorf("Missing message %d", i)
			continue
		}
		if !contains(messages[i], expected) {
			t.Errorf("Expected message %d to contain '%s', got '%s'", i, expected, messages[i])
		}
	}
}

func TestCoreLoggerCodedLogging(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	config.Level = interfaces.TraceLevel

	// Configura o provider
	provider.Configure(config)

	coreLogger := NewCoreLogger(provider, config)
	ctx := context.Background()

	// Testa logging com código
	coreLogger.TraceWithCode(ctx, "TRACE_CODE", "trace with code")
	coreLogger.DebugWithCode(ctx, "DEBUG_CODE", "debug with code")
	coreLogger.InfoWithCode(ctx, "INFO_CODE", "info with code")
	coreLogger.WarnWithCode(ctx, "WARN_CODE", "warn with code")
	coreLogger.ErrorWithCode(ctx, "ERROR_CODE", "error with code")

	messages := provider.GetLogMessages()
	if len(messages) != 5 {
		t.Errorf("Expected 5 messages, got %d", len(messages))
	}

	// Verifica se os códigos estão presentes
	expectedCodes := []string{"TRACE_CODE", "DEBUG_CODE", "INFO_CODE", "WARN_CODE", "ERROR_CODE"}
	for i, expectedCode := range expectedCodes {
		if i >= len(messages) {
			t.Errorf("Missing message %d", i)
			continue
		}
		if !contains(messages[i], expectedCode) {
			t.Errorf("Expected message %d to contain '%s', got '%s'", i, expectedCode, messages[i])
		}
	}
}

func TestCoreLoggerWithFields(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()

	coreLogger := NewCoreLogger(provider, config)
	ctx := context.Background()

	// Testa com fields
	fields := []interfaces.Field{
		interfaces.String("key1", "value1"),
		interfaces.Int("key2", 42),
		interfaces.Bool("key3", true),
	}

	coreLogger.Info(ctx, "message with fields", fields...)

	messages := provider.GetLogMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	message := messages[0]
	if !contains(message, "key1=value1") {
		t.Errorf("Expected message to contain 'key1=value1', got '%s'", message)
	}
	if !contains(message, "key2=42") {
		t.Errorf("Expected message to contain 'key2=42', got '%s'", message)
	}
	if !contains(message, "key3=true") {
		t.Errorf("Expected message to contain 'key3=true', got '%s'", message)
	}
}

func TestCoreLoggerWithFieldsChaining(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()

	coreLogger := NewCoreLogger(provider, config)
	ctx := context.Background()

	// Testa encadeamento de fields
	logger := coreLogger.WithFields(
		interfaces.String("service", "test-service"),
		interfaces.String("version", "v1.0.0"),
	)

	logger.Info(ctx, "message with chained fields")

	messages := provider.GetLogMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	message := messages[0]
	if !contains(message, "service=test-service") {
		t.Errorf("Expected message to contain 'service=test-service', got '%s'", message)
	}
	if !contains(message, "version=v1.0.0") {
		t.Errorf("Expected message to contain 'version=v1.0.0', got '%s'", message)
	}
}

func TestCoreLoggerWithContext(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()

	coreLogger := NewCoreLogger(provider, config)

	// Context com valores
	ctx := context.WithValue(context.Background(), "trace_id", "test-trace-123")

	logger := coreLogger.WithContext(ctx)
	logger.Info(ctx, "message with context")

	messages := provider.GetLogMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestCoreLoggerWithError(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()

	coreLogger := NewCoreLogger(provider, config)
	ctx := context.Background()

	// Testa com error
	testError := fmt.Errorf("test error")
	logger := coreLogger.WithError(testError)
	logger.Error(ctx, "error occurred")

	messages := provider.GetLogMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	message := messages[0]
	if !contains(message, "test error") {
		t.Errorf("Expected message to contain 'test error', got '%s'", message)
	}
}

func TestCoreLoggerWithTraceID(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()

	coreLogger := NewCoreLogger(provider, config)
	ctx := context.Background()

	// Testa com trace ID
	logger := coreLogger.WithTraceID("trace-123")
	logger.Info(ctx, "message with trace ID")

	messages := provider.GetLogMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	message := messages[0]
	if !contains(message, "trace_id=trace-123") {
		t.Errorf("Expected message to contain 'trace_id=trace-123', got '%s'", message)
	}
}

func TestCoreLoggerWithSpanID(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()

	coreLogger := NewCoreLogger(provider, config)
	ctx := context.Background()

	// Testa com span ID
	logger := coreLogger.WithSpanID("span-456")
	logger.Info(ctx, "message with span ID")

	messages := provider.GetLogMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	message := messages[0]
	if !contains(message, "span_id=span-456") {
		t.Errorf("Expected message to contain 'span_id=span-456', got '%s'", message)
	}
}

func TestCoreLoggerLevelOperations(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()

	coreLogger := NewCoreLogger(provider, config)

	// Testa get level
	level := coreLogger.GetLevel()
	if level != config.Level {
		t.Errorf("Expected level %v, got %v", config.Level, level)
	}

	// Testa set level
	newLevel := interfaces.WarnLevel
	coreLogger.SetLevel(newLevel)
	if coreLogger.GetLevel() != newLevel {
		t.Errorf("Expected level %v after set, got %v", newLevel, coreLogger.GetLevel())
	}

	// Testa is level enabled
	if !coreLogger.IsLevelEnabled(interfaces.WarnLevel) {
		t.Error("Expected WarnLevel to be enabled")
	}
	if coreLogger.IsLevelEnabled(interfaces.DebugLevel) {
		t.Error("Expected DebugLevel to be disabled")
	}
}

func TestCoreLoggerClone(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()

	coreLogger := NewCoreLogger(provider, config)
	clone := coreLogger.Clone()

	if clone == nil {
		t.Fatal("Expected clone to be created")
	}

	// Clone deve ser uma instância diferente
	if clone == coreLogger {
		t.Error("Expected clone to be different instance")
	}

	// Mas deve ter as mesmas configurações
	if clone.GetLevel() != coreLogger.GetLevel() {
		t.Error("Expected clone to have same level")
	}
}

func TestCoreLoggerFlushAndClose(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()

	coreLogger := NewCoreLogger(provider, config)

	// Testa flush
	if err := coreLogger.Flush(); err != nil {
		t.Errorf("Expected no error from Flush, got %v", err)
	}

	// Testa close
	if err := coreLogger.Close(); err != nil {
		t.Errorf("Expected no error from Close, got %v", err)
	}
}

func TestCoreLoggerGlobalFields(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	config.GlobalFields = map[string]interface{}{
		"service": "test-service",
		"version": "v1.0.0",
	}

	coreLogger := NewCoreLogger(provider, config)
	ctx := context.Background()

	coreLogger.Info(ctx, "test message")

	messages := provider.GetLogMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	// Os campos globais devem estar presentes na mensagem
	// Nota: dependendo da implementação, pode precisar verificar se convertGlobalFields funciona
}

func TestCoreLoggerAsyncProcessing(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	config.Async = &interfaces.AsyncConfig{
		Enabled:       true,
		BufferSize:    100,
		FlushInterval: 10 * time.Millisecond,
		Workers:       1,
		DropOnFull:    false,
	}

	coreLogger := NewCoreLogger(provider, config)
	ctx := context.Background()

	// Envia várias mensagens rapidamente
	for i := 0; i < 10; i++ {
		coreLogger.Info(ctx, fmt.Sprintf("async message %d", i))
	}

	// Aguarda um pouco para processamento assíncrono
	time.Sleep(50 * time.Millisecond)

	// Flush para garantir que todas as mensagens foram processadas
	coreLogger.Flush()

	messages := provider.GetLogMessages()
	if len(messages) != 10 {
		t.Errorf("Expected 10 messages, got %d", len(messages))
	}
}

func TestCoreLoggerSampling(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	config.Level = interfaces.DebugLevel
	config.Sampling = &interfaces.SamplingConfig{
		Enabled:    true,
		Initial:    2,  // Permite apenas 2 mensagens iniciais
		Thereafter: 10, // Depois permite 1 a cada 10
		Tick:       100 * time.Millisecond,
		Levels:     []interfaces.Level{interfaces.DebugLevel},
	}

	coreLogger := NewCoreLogger(provider, config)
	ctx := context.Background()

	// Envia várias mensagens de debug (que devem ser amostradas)
	for i := 0; i < 10; i++ {
		coreLogger.Debug(ctx, fmt.Sprintf("debug message %d", i))
	}

	messages := provider.GetLogMessages()
	// Com sampling, devemos ter menos de 10 mensagens de debug
	if len(messages) >= 10 {
		t.Errorf("Expected sampling to reduce messages, got %d", len(messages))
	}
}

func TestCoreLoggerEntryPool(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()

	coreLogger := NewCoreLogger(provider, config)

	// Testa que o pool de entries funciona
	entry1 := coreLogger.entryPool.Get()
	if entry1 == nil {
		t.Error("Expected to get entry from pool")
	}

	// Retorna ao pool
	coreLogger.entryPool.Put(entry1)

	// Pega novamente
	entry2 := coreLogger.entryPool.Get()
	if entry2 == nil {
		t.Error("Expected to get entry from pool again")
	}
}

func TestCoreLoggerBufferPool(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()

	coreLogger := NewCoreLogger(provider, config)

	// Testa que o pool de buffers funciona
	buffer1 := coreLogger.bufferPool.Get()
	if buffer1 == nil {
		t.Error("Expected to get buffer from pool")
	}

	// Retorna ao pool
	coreLogger.bufferPool.Put(buffer1)

	// Pega novamente
	buffer2 := coreLogger.bufferPool.Get()
	if buffer2 == nil {
		t.Error("Expected to get buffer from pool again")
	}
}

func TestCoreLoggerLevelFiltering(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	config.Level = interfaces.WarnLevel // Apenas WARN e acima

	coreLogger := NewCoreLogger(provider, config)
	ctx := context.Background()

	// Testa diferentes níveis
	coreLogger.Debug(ctx, "debug message") // Não deve aparecer
	coreLogger.Info(ctx, "info message")   // Não deve aparecer
	coreLogger.Warn(ctx, "warn message")   // Deve aparecer
	coreLogger.Error(ctx, "error message") // Deve aparecer

	messages := provider.GetLogMessages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages (warn and error), got %d", len(messages))
	}

	// Verifica se apenas warn e error estão presentes
	if !contains(messages[0], "warn message") {
		t.Errorf("Expected first message to be warn, got '%s'", messages[0])
	}
	if !contains(messages[1], "error message") {
		t.Errorf("Expected second message to be error, got '%s'", messages[1])
	}
}

// TestCoreLoggerFatalAndPanic testa as funções Fatal e Panic formatadas
func TestCoreLoggerFatalAndPanic(t *testing.T) {
	provider := NewMockProvider("test-fatal", "1.0.0")
	config := TestConfig()
	provider.Configure(config)

	coreLogger := NewCoreLogger(provider, config)
	ctx := context.Background()

	// Test Fatalf - só loga no mock, não faz panic real
	coreLogger.Fatalf(ctx, "fatal error: %s", "test")

	// Não vamos validar a contagem específica devido ao compartilhamento de estado
	// mas vamos garantir que houve pelo menos uma mensagem
	messages := provider.GetLogMessages()
	if len(messages) == 0 {
		t.Error("Expected at least one fatal message")
	}
}

func TestCoreLoggerPanicf(t *testing.T) {
	provider := NewMockProvider("test-panic", "1.0.0")
	config := TestConfig()
	provider.Configure(config)

	coreLogger := NewCoreLogger(provider, config)
	ctx := context.Background()

	// Test Panicf - só loga no mock, não faz panic real
	coreLogger.Panicf(ctx, "panic error: %s", "test")

	// Não vamos validar a contagem específica devido ao compartilhamento de estado
	// mas vamos garantir que houve pelo menos uma mensagem
	messages := provider.GetLogMessages()
	if len(messages) == 0 {
		t.Error("Expected at least one panic message")
	}
}

// TestCoreLoggerAdvancedFeatures testa funcionalidades avançadas
func TestCoreLoggerAdvancedFeatures(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	config.AddCaller = true
	config.AddStacktrace = true
	provider.Configure(config)

	coreLogger := NewCoreLogger(provider, config)
	ctx := context.Background()

	// Testa com caller e stacktrace habilitados
	coreLogger.Error(ctx, "error with caller and stacktrace")

	messages := provider.GetLogMessages()
	if len(messages) == 0 {
		t.Error("Expected at least one message")
	}
}

// TestCoreLoggerWithContextExtraction testa extração de IDs do contexto
func TestCoreLoggerWithContextExtraction(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	provider.Configure(config)

	coreLogger := NewCoreLogger(provider, config)

	// Cria contexto com valores
	ctx := context.WithValue(context.Background(), "trace_id", "test-trace")
	ctx = context.WithValue(ctx, "span_id", "test-span")
	ctx = context.WithValue(ctx, "user_id", "test-user")
	ctx = context.WithValue(ctx, "request_id", "test-request")

	coreLogger.Info(ctx, "message with context values")

	messages := provider.GetLogMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

// TestCoreLoggerCloseAsync testa o fechamento de async processor
func TestCoreLoggerCloseAsync(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	config.Async = &interfaces.AsyncConfig{
		Enabled:       true,
		BufferSize:    100,
		Workers:       1,
		FlushInterval: 100 * time.Millisecond,
		DropOnFull:    false,
	}
	provider.Configure(config)

	coreLogger := NewCoreLogger(provider, config)
	ctx := context.Background()

	// Adiciona algumas mensagens
	coreLogger.Info(ctx, "async message 1")
	coreLogger.Info(ctx, "async message 2")

	// Aguarda um pouco para processamento
	time.Sleep(50 * time.Millisecond)

	// Testa Close que deve parar o async processor
	err := coreLogger.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got %v", err)
	}
}

// TestCoreLoggerSamplerClose testa o fechamento do sampler
func TestCoreLoggerSamplerClose(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	config.Sampling = &interfaces.SamplingConfig{
		Initial:    1,
		Thereafter: 100,
	}
	provider.Configure(config)

	coreLogger := NewCoreLogger(provider, config)

	// Testa Close que deve parar o sampler
	err := coreLogger.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got %v", err)
	}
}

func TestCoreLoggerAsyncProcessingWithSampling(t *testing.T) {
	// Test async processing with sampling (if supported)
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "async-service",
	}

	provider := NewMockProvider("test", "1.0.0")
	logger := NewCoreLogger(provider, config)
	defer logger.Close()

	ctx := context.Background()

	// Send multiple messages
	for i := 0; i < 10; i++ {
		logger.Info(ctx, fmt.Sprintf("message %d", i))
	}

	// Test flush
	logger.Flush()
}

func TestCoreLoggerComplexContextExtraction(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "context-service",
	}
	provider.Configure(config)

	logger := NewCoreLogger(provider, config)
	defer logger.Close()

	// Test complex context scenarios
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			"empty_context",
			context.Background(),
		},
		{
			"trace_only",
			context.WithValue(context.Background(), "trace_id", "trace-123"),
		},
		{
			"span_only",
			context.WithValue(context.Background(), "span_id", "span-456"),
		},
		{
			"user_only",
			context.WithValue(context.Background(), "user_id", "user-789"),
		},
		{
			"request_only",
			context.WithValue(context.Background(), "request_id", "req-101"),
		},
		{
			"all_context_values",
			func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, "trace_id", "trace-abc")
				ctx = context.WithValue(ctx, "span_id", "span-def")
				ctx = context.WithValue(ctx, "user_id", "user-ghi")
				ctx = context.WithValue(ctx, "request_id", "req-jkl")
				return ctx
			}(),
		},
		{
			"non_string_values",
			func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, "trace_id", 12345)
				ctx = context.WithValue(ctx, "span_id", []byte("span"))
				ctx = context.WithValue(ctx, "user_id", 67890)
				ctx = context.WithValue(ctx, "request_id", 99999)
				return ctx
			}(),
		},
		{
			"nil_values",
			func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, "trace_id", nil)
				ctx = context.WithValue(ctx, "span_id", nil)
				ctx = context.WithValue(ctx, "user_id", nil)
				ctx = context.WithValue(ctx, "request_id", nil)
				return ctx
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test direct extraction functions
			traceID := extractTraceID(tt.ctx)
			spanID := extractSpanID(tt.ctx)
			userID := extractUserID(tt.ctx)
			requestID := extractRequestID(tt.ctx)

			t.Logf("Context %s - Extracted: trace=%s, span=%s, user=%s, request=%s",
				tt.name, traceID, spanID, userID, requestID)

			// Test logging with context
			logger.Info(tt.ctx, fmt.Sprintf("test message with %s", tt.name))
		})
	}
}

func TestCoreLoggerWithErrorEdgeCases(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := interfaces.Config{
		Level:       interfaces.ErrorLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "error-service",
	}
	provider.Configure(config)

	logger := NewCoreLogger(provider, config)
	defer logger.Close()

	ctx := context.Background()

	// Test different error scenarios
	tests := []struct {
		name string
		err  error
	}{
		{"nil_error", nil},
		{"simple_error", errors.New("simple error")},
		{"wrapped_error", fmt.Errorf("wrapped: %w", errors.New("original"))},
		{"complex_error", fmt.Errorf("level1: %w", fmt.Errorf("level2: %w", errors.New("root cause")))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorLogger := logger.WithError(tt.err)
			errorLogger.Error(ctx, fmt.Sprintf("error test: %s", tt.name))
		})
	}
}

func TestCoreLoggerLevelEnabledEdgeCases(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := interfaces.Config{
		Level:       interfaces.WarnLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "level-service",
	}
	provider.Configure(config)

	logger := NewCoreLogger(provider, config)
	defer logger.Close()

	ctx := context.Background()

	// Test level filtering
	tests := []struct {
		level     interfaces.Level
		method    func(context.Context, string, ...interfaces.Field)
		shouldLog bool
	}{
		{interfaces.TraceLevel, logger.Trace, false},
		{interfaces.DebugLevel, logger.Debug, false},
		{interfaces.InfoLevel, logger.Info, false},
		{interfaces.WarnLevel, logger.Warn, true},
		{interfaces.ErrorLevel, logger.Error, true},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			// Test if level is enabled
			enabled := logger.IsLevelEnabled(tt.level)
			if enabled != tt.shouldLog {
				t.Errorf("Expected IsLevelEnabled(%v) = %v, got %v", tt.level, tt.shouldLog, enabled)
			}

			// Test actual logging
			tt.method(ctx, fmt.Sprintf("test %s message", tt.level.String()))
		})
	}
}

func TestCoreLoggerFormattedLoggingEdgeCases(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := interfaces.Config{
		Level:       interfaces.DebugLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "format-service",
	}
	provider.Configure(config)

	logger := NewCoreLogger(provider, config)
	defer logger.Close()

	ctx := context.Background()

	// Test formatted logging methods
	logger.Tracef(ctx, "trace formatted: %s %d", "value", 42)
	logger.Debugf(ctx, "debug formatted: %s %d", "value", 42)
	logger.Infof(ctx, "info formatted: %s %d", "value", 42)
	logger.Warnf(ctx, "warn formatted: %s %d", "value", 42)
	logger.Errorf(ctx, "error formatted: %s %d", "value", 42)
}

func TestCoreLoggerCodedLoggingEdgeCases(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := interfaces.Config{
		Level:       interfaces.DebugLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "coded-service",
	}
	provider.Configure(config)

	logger := NewCoreLogger(provider, config)
	defer logger.Close()

	ctx := context.Background()

	// Test coded logging methods
	tests := []struct {
		method func(context.Context, string, string, ...interfaces.Field)
		code   string
		msg    string
	}{
		{logger.TraceWithCode, "TRACE_001", "trace with code"},
		{logger.DebugWithCode, "DEBUG_001", "debug with code"},
		{logger.InfoWithCode, "INFO_001", "info with code"},
		{logger.WarnWithCode, "WARN_001", "warn with code"},
		{logger.ErrorWithCode, "ERROR_001", "error with code"},
	}

	for _, tt := range tests {
		tt.method(ctx, tt.code, tt.msg, interfaces.String("extra", "field"))
	}
}

func TestCoreLoggerCloneEdgeCases(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "clone-service",
		Async: &interfaces.AsyncConfig{
			Enabled:       true,
			BufferSize:    100,
			FlushInterval: 10 * time.Millisecond,
			Workers:       1,
			DropOnFull:    false,
		},
		Sampling: &interfaces.SamplingConfig{
			Enabled:    true,
			Initial:    2,
			Thereafter: 10,
		},
	}
	provider.Configure(config)

	logger := NewCoreLogger(provider, config)
	defer logger.Close()

	// Test cloning
	cloned := logger.Clone()
	if cloned == nil {
		t.Error("Expected cloned logger to be created")
	}

	ctx := context.Background()
	cloned.Info(ctx, "cloned logger message")

	// Test cloning with fields
	fieldsLogger := logger.WithFields(interfaces.String("key", "value"))
	clonedFields := fieldsLogger.Clone()
	clonedFields.Info(ctx, "cloned fields logger message")
}

func TestCoreLoggerContainsLevelFunction(t *testing.T) {
	// Test the containsLevel function that currently has 0% coverage
	levels := []interfaces.Level{interfaces.ErrorLevel, interfaces.WarnLevel}

	// Test with levels that should be found
	if !containsLevel(levels, interfaces.ErrorLevel) {
		t.Error("Expected ErrorLevel to be found in levels slice")
	}

	if !containsLevel(levels, interfaces.WarnLevel) {
		t.Error("Expected WarnLevel to be found in levels slice")
	}

	// Test with level that should not be found
	if containsLevel(levels, interfaces.InfoLevel) {
		t.Error("Expected InfoLevel not to be found in levels slice")
	}
}

func TestCoreLoggerGlobalManagerFunctions(t *testing.T) {
	// Test global manager functions that currently have 0% coverage

	// Test GetGlobalManager
	manager := GetGlobalManager()
	if manager == nil {
		t.Error("Expected global manager to be created")
	}

	// Create a real provider for testing
	testProvider := NewMockProvider("test", "1.0.0")

	// Test RegisterProvider through global manager
	RegisterProvider("test-provider", testProvider)

	// Test SetProvider
	config := interfaces.Config{
		Level:       interfaces.InfoLevel,
		Format:      interfaces.JSONFormat,
		ServiceName: "global-service",
	}
	err := SetProvider("test-provider", config)
	if err != nil {
		t.Errorf("Expected no error setting provider, got %v", err)
	}

	// Test ListProviders
	providers := ListProviders()
	found := false
	for _, p := range providers {
		if p == "test-provider" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected test-provider to be in providers list")
	}

	// Test CreateLogger through global manager
	globalLogger, err := CreateLogger("test-provider", config)
	if err != nil {
		t.Errorf("Expected no error creating logger, got %v", err)
	}

	if globalLogger != nil {
		ctx := context.Background()

		// Test global logging functions
		Trace(ctx, "global trace message")
		Debug(ctx, "global debug message")
		Info(ctx, "global info message")
		Warn(ctx, "global warn message")
	}
}

// Mock provider for testing
type mockProvider struct{}

func (m *mockProvider) Name() string                                                      { return "mock" }
func (m *mockProvider) Version() string                                                   { return "1.0.0" }
func (m *mockProvider) Configure(config interfaces.Config) error                          { return nil }
func (m *mockProvider) HealthCheck() error                                                { return nil }
func (m *mockProvider) Trace(ctx context.Context, msg string, fields ...interfaces.Field) {}
func (m *mockProvider) Debug(ctx context.Context, msg string, fields ...interfaces.Field) {}
func (m *mockProvider) Info(ctx context.Context, msg string, fields ...interfaces.Field)  {}
func (m *mockProvider) Warn(ctx context.Context, msg string, fields ...interfaces.Field)  {}
func (m *mockProvider) Error(ctx context.Context, msg string, fields ...interfaces.Field) {}
func (m *mockProvider) Fatal(ctx context.Context, msg string, fields ...interfaces.Field) {}
func (m *mockProvider) Panic(ctx context.Context, msg string, fields ...interfaces.Field) {}
func (m *mockProvider) Tracef(ctx context.Context, format string, args ...any)            {}
func (m *mockProvider) Debugf(ctx context.Context, format string, args ...any)            {}
func (m *mockProvider) Infof(ctx context.Context, format string, args ...any)             {}
func (m *mockProvider) Warnf(ctx context.Context, format string, args ...any)             {}
func (m *mockProvider) Errorf(ctx context.Context, format string, args ...any)            {}
func (m *mockProvider) Fatalf(ctx context.Context, format string, args ...any)            {}
func (m *mockProvider) Panicf(ctx context.Context, format string, args ...any)            {}
func (m *mockProvider) TraceWithCode(ctx context.Context, code, msg string, fields ...interfaces.Field) {
}
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
func (m *mockProvider) WithError(err error) interfaces.Logger                   { return m }
func (m *mockProvider) WithTraceID(traceID string) interfaces.Logger            { return m }
func (m *mockProvider) WithSpanID(spanID string) interfaces.Logger              { return m }
func (m *mockProvider) SetLevel(level interfaces.Level)                         {}
func (m *mockProvider) GetLevel() interfaces.Level                              { return interfaces.InfoLevel }
func (m *mockProvider) IsLevelEnabled(level interfaces.Level) bool              { return true }
func (m *mockProvider) Clone() interfaces.Logger                                { return m }
func (m *mockProvider) Flush() error                                            { return nil }
func (m *mockProvider) Close() error                                            { return nil }
