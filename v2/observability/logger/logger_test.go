package logger

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/observability/logger/interfaces"
)

// Helper function para configurar logger global para testes
func setupTestGlobalLogger(level interfaces.Level) *MockProvider {
	provider := NewMockProvider("global", "1.0.0")
	config := TestConfig()
	config.Level = level

	// Configura o provider
	provider.Configure(config)

	coreLogger := NewCoreLogger(provider, config)
	SetCurrentLogger(coreLogger)

	return provider
}

func TestGlobalLoggerFunctions(t *testing.T) {
	provider := setupTestGlobalLogger(interfaces.TraceLevel)
	ctx := context.Background()

	// Testa funções globais básicas
	Trace(ctx, "global trace message")
	Debug(ctx, "global debug message")
	Info(ctx, "global info message")
	Warn(ctx, "global warn message")
	Error(ctx, "global error message")

	messages := provider.GetLogMessages()
	if len(messages) != 5 {
		t.Errorf("Expected 5 messages, got %d", len(messages))
	}
}

func TestGlobalFormattedLoggerFunctions(t *testing.T) {
	provider := setupTestGlobalLogger(interfaces.TraceLevel)
	ctx := context.Background()

	// Testa funções globais formatadas
	Tracef(ctx, "global trace: %s %d", "value", 42)
	Debugf(ctx, "global debug: %s %d", "value", 42)
	Infof(ctx, "global info: %s %d", "value", 42)
	Warnf(ctx, "global warn: %s %d", "value", 42)
	Errorf(ctx, "global error: %s %d", "value", 42)

	messages := provider.GetLogMessages()
	if len(messages) != 5 {
		t.Errorf("Expected 5 messages, got %d", len(messages))
	}

	// Verifica formatação
	for _, message := range messages {
		if !contains(message, "value 42") {
			t.Errorf("Expected message to contain 'value 42', got '%s'", message)
		}
	}
}

func TestGlobalCodedLoggerFunctions(t *testing.T) {
	provider := setupTestGlobalLogger(interfaces.TraceLevel)
	ctx := context.Background()

	// Testa funções globais com código
	TraceWithCode(ctx, "TRACE_CODE", "global trace with code")
	DebugWithCode(ctx, "DEBUG_CODE", "global debug with code")
	InfoWithCode(ctx, "INFO_CODE", "global info with code")
	WarnWithCode(ctx, "WARN_CODE", "global warn with code")
	ErrorWithCode(ctx, "ERROR_CODE", "global error with code")

	messages := provider.GetLogMessages()
	if len(messages) != 5 {
		t.Errorf("Expected 5 messages, got %d", len(messages))
	}

	// Verifica códigos
	expectedCodes := []string{"TRACE_CODE", "DEBUG_CODE", "INFO_CODE", "WARN_CODE", "ERROR_CODE"}
	for i, expectedCode := range expectedCodes {
		if i >= len(messages) {
			continue
		}
		if !contains(messages[i], expectedCode) {
			t.Errorf("Expected message %d to contain '%s', got '%s'", i, expectedCode, messages[i])
		}
	}
}

func TestGlobalLoggerWithFields(t *testing.T) {
	provider := setupTestGlobalLogger(interfaces.InfoLevel)
	ctx := context.Background()

	// Testa com fields
	fields := []interfaces.Field{
		interfaces.String("key1", "value1"),
		interfaces.Int("key2", 42),
	}

	WithFields(fields...).Info(ctx, "message with fields")

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
}

func TestGlobalLoggerWithContext(t *testing.T) {
	provider := setupTestGlobalLogger(interfaces.InfoLevel)
	ctx := context.WithValue(context.Background(), "trace_id", "test-trace")

	WithContext(ctx).Info(ctx, "message with context")

	messages := provider.GetLogMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestGlobalLoggerWithError(t *testing.T) {
	provider := setupTestGlobalLogger(interfaces.ErrorLevel)
	ctx := context.Background()
	testError := fmt.Errorf("test error")

	WithError(testError).Error(ctx, "error occurred")

	messages := provider.GetLogMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	message := messages[0]
	if !contains(message, "test error") {
		t.Errorf("Expected message to contain 'test error', got '%s'", message)
	}
}

func TestGlobalLoggerWithTraceID(t *testing.T) {
	provider := setupTestGlobalLogger(interfaces.InfoLevel)
	ctx := context.Background()

	WithTraceID("trace-123").Info(ctx, "message with trace ID")

	messages := provider.GetLogMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	message := messages[0]
	if !contains(message, "trace_id=trace-123") {
		t.Errorf("Expected message to contain 'trace_id=trace-123', got '%s'", message)
	}
}

func TestGlobalLoggerWithSpanID(t *testing.T) {
	provider := setupTestGlobalLogger(interfaces.InfoLevel)
	ctx := context.Background()

	WithSpanID("span-456").Info(ctx, "message with span ID")

	messages := provider.GetLogMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	message := messages[0]
	if !contains(message, "span_id=span-456") {
		t.Errorf("Expected message to contain 'span_id=span-456', got '%s'", message)
	}
}

func TestGlobalLoggerLevelOperations(t *testing.T) {
	provider := NewMockProvider("global", "1.0.0")
	config := TestConfig()

	coreLogger := NewCoreLogger(provider, config)
	SetCurrentLogger(coreLogger)

	// Testa get level
	level := GetLevel()
	if level != config.Level {
		t.Errorf("Expected level %v, got %v", config.Level, level)
	}

	// Testa set level
	newLevel := interfaces.WarnLevel
	SetLevel(newLevel)
	if GetLevel() != newLevel {
		t.Errorf("Expected level %v after set, got %v", newLevel, GetLevel())
	}

	// Testa is level enabled
	if !IsLevelEnabled(interfaces.WarnLevel) {
		t.Error("Expected WarnLevel to be enabled")
	}
	if IsLevelEnabled(interfaces.DebugLevel) {
		t.Error("Expected DebugLevel to be disabled")
	}
}

func TestGlobalLoggerClone(t *testing.T) {
	setupTestGlobalLogger(interfaces.InfoLevel)

	clone := Clone()
	if clone == nil {
		t.Fatal("Expected clone to be created")
	}

	// Testa se clone é funcional
	ctx := context.Background()
	clone.Info(ctx, "test clone message")
}

func TestGlobalLoggerFlushAndClose(t *testing.T) {
	setupTestGlobalLogger(interfaces.InfoLevel)

	// Testa flush
	if err := Flush(); err != nil {
		t.Errorf("Expected no error from Flush, got %v", err)
	}

	// Testa close
	if err := Close(); err != nil {
		t.Errorf("Expected no error from Close, got %v", err)
	}
}

func TestGlobalLoggerNoopBehavior(t *testing.T) {
	// Limpa o logger global
	SetCurrentLogger(nil)

	ctx := context.Background()

	// Todas essas chamadas devem funcionar sem erro (noop behavior)
	Debug(ctx, "debug message")
	Info(ctx, "info message")
	Warn(ctx, "warn message")
	Error(ctx, "error message")

	Debugf(ctx, "debug: %s", "value")
	Infof(ctx, "info: %s", "value")
	Warnf(ctx, "warn: %s", "value")
	Errorf(ctx, "error: %s", "value")

	DebugWithCode(ctx, "DEBUG", "debug with code")
	InfoWithCode(ctx, "INFO", "info with code")
	WarnWithCode(ctx, "WARN", "warn with code")
	ErrorWithCode(ctx, "ERROR", "error with code")

	// Operações que retornam valores também devem funcionar
	logger := WithFields(interfaces.String("key", "value"))
	if logger == nil {
		t.Error("Expected noop logger to be returned")
	}

	logger = WithContext(ctx)
	if logger == nil {
		t.Error("Expected noop logger to be returned")
	}

	logger = WithError(fmt.Errorf("test error"))
	if logger == nil {
		t.Error("Expected noop logger to be returned")
	}

	logger = WithTraceID("trace-123")
	if logger == nil {
		t.Error("Expected noop logger to be returned")
	}

	logger = WithSpanID("span-456")
	if logger == nil {
		t.Error("Expected noop logger to be returned")
	}

	logger = Clone()
	if logger == nil {
		t.Error("Expected noop logger to be returned")
	}

	// Level operations
	level := GetLevel()
	if level != interfaces.InfoLevel { // noop logger deve retornar InfoLevel como padrão
		t.Errorf("Expected InfoLevel from noop logger, got %v", level)
	}

	SetLevel(interfaces.WarnLevel) // Não deve causar panic

	if IsLevelEnabled(interfaces.ErrorLevel) { // noop logger deve retornar false para IsLevelEnabled
		t.Error("Expected noop logger to return false for IsLevelEnabled")
	}

	// Flush e Close não devem causar erro
	if err := Flush(); err != nil {
		t.Errorf("Expected no error from noop Flush, got %v", err)
	}

	if err := Close(); err != nil {
		t.Errorf("Expected no error from noop Close, got %v", err)
	}
}

func TestSetAndGetCurrentLogger(t *testing.T) {
	// Salva o logger atual
	originalLogger := GetCurrentLogger()

	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	testLogger := NewCoreLogger(provider, config)

	// Define novo logger
	SetCurrentLogger(testLogger)

	// Verifica se foi definido
	currentLogger := GetCurrentLogger()
	if currentLogger != testLogger {
		t.Error("Expected current logger to be the test logger")
	}

	// Restaura o logger original
	SetCurrentLogger(originalLogger)
}

func TestGlobalLoggerConcurrency(t *testing.T) {
	provider := NewMockProvider("concurrent", "1.0.0")
	config := TestConfig()
	config.Level = interfaces.DebugLevel

	coreLogger := NewCoreLogger(provider, config)
	SetCurrentLogger(coreLogger)

	ctx := context.Background()
	const numGoroutines = 10
	const messagesPerGoroutine = 100

	// Executa logging concorrente
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < messagesPerGoroutine; j++ {
				Info(ctx, fmt.Sprintf("concurrent message from goroutine %d, message %d", id, j))
			}
		}(i)
	}

	// Aguarda todas as goroutines terminarem
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Aguarda um pouco para o processamento assíncrono
	time.Sleep(100 * time.Millisecond)

	messages := provider.GetLogMessages()
	expectedMessages := numGoroutines * messagesPerGoroutine

	if len(messages) != expectedMessages {
		t.Errorf("Expected %d messages, got %d", expectedMessages, len(messages))
	}
}

func TestFatalAndPanicBehavior(t *testing.T) {
	provider := NewMockProvider("fatal-panic", "1.0.0")
	config := TestConfig()

	coreLogger := NewCoreLogger(provider, config)
	SetCurrentLogger(coreLogger)

	ctx := context.Background()

	// Nota: Em um ambiente real, Fatal chamaria os.Exit e Panic causaria um panic
	// Aqui apenas testamos se as funções podem ser chamadas sem erro

	// Para testar Fatal, precisamos de uma versão que não chame os.Exit
	// Para testar Panic, precisamos de uma versão que não cause panic real

	// Por enquanto, vamos testar apenas se não há erros de compilação
	defer func() {
		if r := recover(); r != nil {
			// Se houver panic, isso é esperado para Panic()
			t.Logf("Recovered from panic (expected): %v", r)
		}
	}()

	// Fatal normalmente deveria sair da aplicação, mas nossa implementação mock não faz isso
	Fatal(ctx, "fatal message")

	// Panic normalmente deveria causar panic, mas nossa implementação mock não faz isso
	Panic(ctx, "panic message")

	messages := provider.GetLogMessages()
	if len(messages) < 2 {
		t.Errorf("Expected at least 2 messages (fatal and panic), got %d", len(messages))
	}
}

func TestGlobalLoggerTypeConversions(t *testing.T) {
	provider := NewMockProvider("types", "1.0.0")
	config := TestConfig()

	coreLogger := NewCoreLogger(provider, config)
	SetCurrentLogger(coreLogger)

	ctx := context.Background()

	// Testa diferentes tipos de campos
	fields := []interfaces.Field{
		interfaces.String("string_field", "test"),
		interfaces.Int("int_field", 42),
		interfaces.Int64("int64_field", int64(42)),
		interfaces.Float64("float_field", 3.14),
		interfaces.Bool("bool_field", true),
		interfaces.Time("time_field", time.Now()),
		interfaces.Duration("duration_field", time.Second),
		interfaces.Error(fmt.Errorf("test error")),
	}

	Info(ctx, "message with various field types", fields...)

	messages := provider.GetLogMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	message := messages[0]
	// Verifica se alguns campos estão presentes
	if !contains(message, "string_field=test") {
		t.Errorf("Expected string field in message: %s", message)
	}
	if !contains(message, "int_field=42") {
		t.Errorf("Expected int field in message: %s", message)
	}
	if !contains(message, "bool_field=true") {
		t.Errorf("Expected bool field in message: %s", message)
	}
}

func TestGlobalFatalfPanicf(t *testing.T) {
	provider := setupTestGlobalLogger(interfaces.InfoLevel)
	ctx := context.Background()

	// Testa Fatalf global
	Fatalf(ctx, "test fatal: %s", "message")

	// Testa Panicf global
	Panicf(ctx, "test panic: %s", "message")

	messages := provider.GetLogMessages()
	if len(messages) < 2 {
		t.Errorf("Expected at least 2 messages, got %d", len(messages))
	}
}
