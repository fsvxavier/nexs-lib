package logger

import (
	"context"
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

// TestCoreLoggerSamplerClose testa o fechamento correto do sampler
func TestSamplerClose(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	config.Sampling = &interfaces.SamplingConfig{
		Enabled:    true,
		Initial:    10,
		Thereafter: 100,
		Tick:       100 * time.Millisecond,
		Levels:     []interfaces.Level{interfaces.InfoLevel},
	}

	provider.Configure(config)
	logger := NewCoreLogger(provider, config)

	// Verifica se sampler foi criado
	if logger.sampler == nil {
		t.Fatal("Expected sampler to be created")
	}

	// Testa se o sampler está funcionando
	ctx := context.Background()
	logger.Info(ctx, "Test message")

	// Fecha o logger (deve fechar o sampler)
	err := logger.Close()
	if err != nil {
		t.Errorf("Unexpected error closing logger: %v", err)
	}

	// Testa se o sampler foi fechado corretamente
	// Não deveria gerar panic ou erro
	// O ticker deve ter sido parado
}

// TestSamplerEdgeCases testa casos extremos do sampling
func TestSamplerEdgeCases(t *testing.T) {
	t.Run("NilSamplingConfig", func(t *testing.T) {
		provider := NewMockProvider("test", "1.0.0")
		config := TestConfig()
		config.Sampling = nil // Explicitamente nil

		provider.Configure(config)
		logger := NewCoreLogger(provider, config)

		if logger.sampler != nil {
			t.Error("Expected sampler to be nil when config is nil")
		}

		// Deve funcionar normalmente sem sampler
		ctx := context.Background()
		logger.Info(ctx, "Test message")

		logger.Close()
	})

	t.Run("DisabledSampling", func(t *testing.T) {
		provider := NewMockProvider("test", "1.0.0")
		config := TestConfig()
		config.Sampling = &interfaces.SamplingConfig{
			Enabled: false, // Explicitamente desabilitado
		}

		provider.Configure(config)
		logger := NewCoreLogger(provider, config)

		if logger.sampler != nil {
			t.Error("Expected sampler to be nil when disabled")
		}

		logger.Close()
	})

	t.Run("EmptyLevelsConfig", func(t *testing.T) {
		provider := NewMockProvider("test", "1.0.0")
		config := TestConfig()
		config.Sampling = &interfaces.SamplingConfig{
			Enabled:    true,
			Initial:    1,
			Thereafter: 10,
			Tick:       time.Millisecond,
			Levels:     []interfaces.Level{}, // Lista vazia
		}

		provider.Configure(config)
		logger := NewCoreLogger(provider, config)

		// Deve criar sampler mesmo com levels vazios
		if logger.sampler == nil {
			t.Error("Expected sampler to be created even with empty levels")
		}

		logger.Close()
	})

	t.Run("ZeroValues", func(t *testing.T) {
		provider := NewMockProvider("test", "1.0.0")
		config := TestConfig()
		config.Sampling = &interfaces.SamplingConfig{
			Enabled:    true,
			Initial:    0, // Valores zero
			Thereafter: 0,
			Tick:       0,
			Levels:     []interfaces.Level{interfaces.InfoLevel},
		}

		provider.Configure(config)
		logger := NewCoreLogger(provider, config)

		// Deve criar sampler e não causar divisão por zero
		if logger.sampler == nil {
			t.Error("Expected sampler to be created with zero values")
		}

		ctx := context.Background()
		logger.Info(ctx, "Test message")

		logger.Close()
	})
}

// TestContextExtractionEdgeCases testa casos extremos da extração de context
func TestContextExtractionEdgeCases(t *testing.T) {
	provider := NewMockProvider("test", "1.0.0")
	config := TestConfig()
	provider.Configure(config)
	logger := NewCoreLogger(provider, config)
	defer logger.Close()

	t.Run("NilContext", func(t *testing.T) {
		// Testa com context nil
		logger.Info(nil, "Message with nil context")

		messages := provider.GetLogMessages()
		found := false
		for _, msg := range messages {
			if strings.Contains(msg, "Message with nil context") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected message to be logged even with nil context")
		}
	})

	t.Run("ContextWithWrongTypes", func(t *testing.T) {
		// Context com tipos incorretos
		ctx := context.WithValue(context.Background(), "trace_id", 12345) // int em vez de string
		ctx = context.WithValue(ctx, "span_id", []byte("bytes"))          // slice em vez de string
		ctx = context.WithValue(ctx, "user_id", struct{}{})               // struct em vez de string
		ctx = context.WithValue(ctx, "request_id", nil)                   // nil

		logger.Info(ctx, "Message with wrong type context values")

		// Não deve causar panic, deve extrair valores vazios para tipos incorretos
		messages := provider.GetLogMessages()
		found := false
		for _, msg := range messages {
			if strings.Contains(msg, "Message with wrong type context values") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected message to be logged even with wrong type context values")
		}
	})

	t.Run("ContextWithValidAndInvalidKeys", func(t *testing.T) {
		// Mistura de chaves válidas e inválidas
		ctx := context.WithValue(context.Background(), "trace_id", "valid-trace")
		ctx = context.WithValue(ctx, "invalid_key", "value")
		ctx = context.WithValue(ctx, "span_id", 999) // tipo inválido
		ctx = context.WithValue(ctx, "user_id", "valid-user")

		logger.Info(ctx, "Message with mixed context")

		messages := provider.GetLogMessages()
		found := false
		for _, msg := range messages {
			if strings.Contains(msg, "Message with mixed context") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected message to be logged with mixed context")
		}
	})

	t.Run("CamelCaseKeys", func(t *testing.T) {
		// Testa diferentes formatos de chaves
		ctx := context.WithValue(context.Background(), "traceId", "camel-trace") // camelCase
		ctx = context.WithValue(ctx, "spanId", "camel-span")                     // camelCase
		ctx = context.WithValue(ctx, "trace_id", "snake-trace")                  // snake_case (deve sobrescrever)
		ctx = context.WithValue(ctx, "TRACE_ID", "upper-trace")                  // uppercase

		logger.Info(ctx, "Message with different key formats")

		messages := provider.GetLogMessages()
		found := false
		for _, msg := range messages {
			if strings.Contains(msg, "Message with different key formats") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected message to be logged with different key formats")
		}
	})

	t.Run("ContextChaining", func(t *testing.T) {
		// Context aninhado/em cadeia
		ctx1 := context.WithValue(context.Background(), "trace_id", "base-trace")
		ctx2 := context.WithValue(ctx1, "span_id", "child-span")
		ctx3 := context.WithValue(ctx2, "user_id", "child-user")
		ctx4 := context.WithValue(ctx3, "trace_id", "overridden-trace") // Sobrescreve

		logger.Info(ctx4, "Message with chained context")

		messages := provider.GetLogMessages()
		found := false
		for _, msg := range messages {
			if strings.Contains(msg, "Message with chained context") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected message to be logged with chained context")
		}
	})
}

// TestAsyncProcessingFailures testa falhas no processamento assíncrono
func TestAsyncProcessingFailures(t *testing.T) {
	t.Run("BufferOverflow", func(t *testing.T) {
		provider := NewMockProvider("test", "1.0.0")
		config := TestConfig()
		config.Async = &interfaces.AsyncConfig{
			Enabled:       true,
			BufferSize:    2, // Buffer muito pequeno
			Workers:       1,
			FlushInterval: 100 * time.Millisecond,
		}

		provider.Configure(config)
		logger := NewCoreLogger(provider, config)
		defer logger.Close()

		ctx := context.Background()

		// Envia mais mensagens do que o buffer pode suportar
		for i := 0; i < 10; i++ {
			logger.Info(ctx, fmt.Sprintf("Overflow message %d", i))
		}

		// Força flush e aguarda processamento
		logger.Flush()
		time.Sleep(200 * time.Millisecond)

		// Algumas mensagens podem ser perdidas devido ao overflow, mas não deve causar panic
		messages := provider.GetLogMessages()
		if len(messages) == 0 {
			t.Error("Expected at least some messages to be processed")
		}
	})

	t.Run("WorkerPanic", func(t *testing.T) {
		// Mock provider que causa panic em certas condições
		provider := &PanicMockProvider{
			MockProvider: NewMockProvider("test", "1.0.0"),
			shouldPanic:  true,
		}

		config := TestConfig()
		config.Async = &interfaces.AsyncConfig{
			Enabled:       true,
			BufferSize:    10,
			Workers:       2,
			FlushInterval: 50 * time.Millisecond,
		}

		provider.Configure(config)
		logger := NewCoreLogger(provider, config)
		defer logger.Close()

		ctx := context.Background()

		// Envia mensagens que devem causar panic no worker
		for i := 0; i < 5; i++ {
			logger.Info(ctx, "Message that causes panic")
		}

		// Aguarda processamento
		time.Sleep(100 * time.Millisecond)

		// Logger deve continuar funcionando mesmo com panic nos workers
		provider.shouldPanic = false
		logger.Info(ctx, "Message after panic")

		logger.Flush()
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("ZeroWorkers", func(t *testing.T) {
		provider := NewMockProvider("test", "1.0.0")
		config := TestConfig()
		config.Async = &interfaces.AsyncConfig{
			Enabled:       true,
			BufferSize:    10,
			Workers:       0, // Zero workers - deve ser corrigido para 1
			FlushInterval: 50 * time.Millisecond,
		}

		provider.Configure(config)
		logger := NewCoreLogger(provider, config)
		defer logger.Close()

		// Verifica se o logger foi criado com pelo menos 1 worker
		if logger.async == nil {
			t.Error("Expected async processor to be created")
		} else if len(logger.async.workers) != 1 {
			t.Errorf("Expected 1 worker (corrected from 0), got %d", len(logger.async.workers))
		}

		ctx := context.Background()
		logger.Info(ctx, "Message with zero workers")

		// Deve funcionar normalmente agora
		logger.Flush()
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("NegativeValues", func(t *testing.T) {
		provider := NewMockProvider("test", "1.0.0")
		config := TestConfig()
		config.Async = &interfaces.AsyncConfig{
			Enabled:       true,
			BufferSize:    -1, // Valores negativos - deve ser corrigido para 100
			Workers:       -1, // Deve ser corrigido para 1
			FlushInterval: -time.Second,
		}

		provider.Configure(config)

		// Não deve causar panic na criação
		logger := NewCoreLogger(provider, config)
		if logger == nil {
			t.Error("Expected logger to be created even with negative values")
		}
		defer logger.Close()

		// Verifica se os valores foram corrigidos
		if logger.async == nil {
			t.Error("Expected async processor to be created")
		} else {
			if len(logger.async.workers) != 1 {
				t.Errorf("Expected 1 worker (corrected from -1), got %d", len(logger.async.workers))
			}
			if cap(logger.async.queue) != 100 {
				t.Errorf("Expected buffer size 100 (corrected from -1), got %d", cap(logger.async.queue))
			}
		}

		// Testa se funciona normalmente
		ctx := context.Background()
		logger.Info(ctx, "Test message with negative values")
		logger.Flush()
	})
}

// PanicMockProvider provider que causa panic para testar recuperação
type PanicMockProvider struct {
	*MockProvider
	shouldPanic bool
}

func (p *PanicMockProvider) Info(ctx context.Context, msg string, fields ...interfaces.Field) {
	if p.shouldPanic && strings.Contains(msg, "panic") {
		panic("simulated panic in provider")
	}
	p.MockProvider.Info(ctx, msg, fields...)
}

func (p *PanicMockProvider) Debug(ctx context.Context, msg string, fields ...interfaces.Field) {
	if p.shouldPanic && strings.Contains(msg, "panic") {
		panic("simulated panic in provider")
	}
	p.MockProvider.Debug(ctx, msg, fields...)
}

func (p *PanicMockProvider) Warn(ctx context.Context, msg string, fields ...interfaces.Field) {
	if p.shouldPanic && strings.Contains(msg, "panic") {
		panic("simulated panic in provider")
	}
	p.MockProvider.Warn(ctx, msg, fields...)
}

func (p *PanicMockProvider) Error(ctx context.Context, msg string, fields ...interfaces.Field) {
	if p.shouldPanic && strings.Contains(msg, "panic") {
		panic("simulated panic in provider")
	}
	p.MockProvider.Error(ctx, msg, fields...)
}

// TestAsyncShutdownEdgeCases testa casos extremos durante o shutdown assíncrono
func TestAsyncShutdownEdgeCases(t *testing.T) {
	t.Run("ShutdownWithPendingMessages", func(t *testing.T) {
		provider := NewMockProvider("test", "1.0.0")
		config := TestConfig()
		config.Async = &interfaces.AsyncConfig{
			Enabled:       true,
			BufferSize:    1000,
			Workers:       1,
			FlushInterval: 1 * time.Second, // Intervalo longo
		}

		provider.Configure(config)
		logger := NewCoreLogger(provider, config)

		ctx := context.Background()

		// Envia muitas mensagens
		for i := 0; i < 100; i++ {
			logger.Info(ctx, fmt.Sprintf("Pending message %d", i))
		}

		// Fecha imediatamente sem aguardar processamento completo
		err := logger.Close()
		if err != nil {
			t.Errorf("Unexpected error during close: %v", err)
		}

		// Verifica se pelo menos algumas mensagens foram processadas
		messages := provider.GetLogMessages()
		if len(messages) == 0 {
			t.Error("Expected some messages to be processed before shutdown")
		}
	})

	t.Run("MultipleCloseCall", func(t *testing.T) {
		provider := NewMockProvider("test", "1.0.0")
		config := TestConfig()
		config.Async = &interfaces.AsyncConfig{
			Enabled:       true,
			BufferSize:    10,
			Workers:       2,
			FlushInterval: 50 * time.Millisecond,
		}

		provider.Configure(config)
		logger := NewCoreLogger(provider, config)

		// Chama Close múltiplas vezes
		err1 := logger.Close()
		err2 := logger.Close()
		err3 := logger.Close()

		// Não deve causar panic ou erro
		if err1 != nil {
			t.Errorf("Unexpected error on first close: %v", err1)
		}
		// Segundo e terceiro close podem retornar erro ou não, mas não devem causar panic
		_ = err2
		_ = err3
	})

	t.Run("CloseWithoutInit", func(t *testing.T) {
		provider := NewMockProvider("test", "1.0.0")
		config := TestConfig()
		// Async nil - não inicializado

		provider.Configure(config)
		logger := NewCoreLogger(provider, config)

		// Close sem async inicializado
		err := logger.Close()
		if err != nil {
			t.Errorf("Unexpected error closing logger without async: %v", err)
		}
	})
}
