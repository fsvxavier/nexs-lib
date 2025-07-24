//go:build integration
// +build integration

package logger_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/observability/logger"
	_ "github.com/fsvxavier/nexs-lib/observability/logger/providers/slog"
)

func TestIntegrationRealWorldScenario(t *testing.T) {
	// Simula cenário real de aplicação web
	var buf bytes.Buffer
	config := &logger.Config{
		Level:          logger.InfoLevel,
		Format:         logger.JSONFormat,
		Output:         &buf,
		ServiceName:    "integration-test",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		AddSource:      true,
		AddStacktrace:  true,
		Fields: map[string]any{
			"datacenter": "test-dc",
			"instance":   "test-01",
		},
	}

	err := logger.SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider: %v", err)
	}

	// Simula middleware HTTP
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.TraceIDKey, "trace-integration-123")
	ctx = context.WithValue(ctx, logger.RequestIDKey, "req-integration-456")
	ctx = context.WithValue(ctx, logger.UserIDKey, "user-integration-789")

	// Log de início da requisição
	logger.Info(ctx, "Request started",
		logger.String("method", "POST"),
		logger.String("path", "/api/users"),
		logger.String("user_agent", "integration-test/1.0"),
	)

	// Simula processamento de negócio
	serviceLogger := logger.WithFields(
		logger.String("service", "user-service"),
		logger.String("operation", "create_user"),
	)

	serviceLogger.Info(ctx, "Processing user creation",
		logger.String("user_email", "test@integration.com"),
	)

	// Simula operação de banco de dados
	start := time.Now()
	time.Sleep(10 * time.Millisecond) // Simula operação
	duration := time.Since(start)

	serviceLogger.Info(ctx, "Database operation completed",
		logger.Duration("db_duration", duration),
		logger.String("operation", "INSERT"),
		logger.String("table", "users"),
	)

	// Simula erro recuperável
	serviceLogger.Warn(ctx, "Rate limit approaching",
		logger.Int("current_requests", 95),
		logger.Int("limit", 100),
	)

	// Simula sucesso final
	current := logger.GetCurrentProvider()
	current.InfoWithCode(ctx, "USER_CREATED", "User created successfully",
		logger.String("user_id", "user-123"),
		logger.String("email", "test@integration.com"),
	)

	// Log de fim da requisição
	logger.Info(ctx, "Request completed",
		logger.String("status", "201"),
		logger.Duration("total_duration", duration*2),
	)

	// Verifica saída
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Deve ter logs suficientes
	if len(lines) < 6 {
		t.Errorf("Expected at least 6 log lines, got %d", len(lines))
	}

	// Verifica se todos os logs contêm os campos de contexto
	expectedFields := []string{
		"trace-integration-123",
		"req-integration-456",
		"user-integration-789",
		"integration-test",
		"test-dc",
		"test-01",
	}

	for _, line := range lines {
		for _, field := range expectedFields {
			if !strings.Contains(line, field) {
				t.Errorf("Expected field '%s' in log line: %s", field, line)
			}
		}
	}

	// Verifica se o código de sucesso está presente
	if !strings.Contains(output, "USER_CREATED") {
		t.Error("Expected success code 'USER_CREATED' in output")
	}
}

func TestIntegrationProviderSwitching(t *testing.T) {
	// Testa mudança dinâmica de providers
	var buf1 bytes.Buffer
	config1 := &logger.Config{
		Level:       logger.InfoLevel,
		Format:      logger.JSONFormat,
		Output:      &buf1,
		ServiceName: "switching-test",
	}

	err := logger.SetProvider("slog", config1)
	if err != nil {
		t.Fatalf("Failed to set first provider: %v", err)
	}

	ctx := context.Background()
	logger.Info(ctx, "First provider message")

	// Muda para nova configuração
	var buf2 bytes.Buffer
	config2 := &logger.Config{
		Level:       logger.DebugLevel,
		Format:      logger.TextFormat,
		Output:      &buf2,
		ServiceName: "switching-test-2",
	}

	err = logger.SetProvider("slog", config2)
	if err != nil {
		t.Fatalf("Failed to set second provider: %v", err)
	}

	logger.Info(ctx, "Second provider message")
	logger.Debug(ctx, "Debug message") // Só deve aparecer no segundo

	// Verifica primeira saída
	output1 := buf1.String()
	if !strings.Contains(output1, "First provider message") {
		t.Error("Expected first message in first buffer")
	}

	if !strings.Contains(output1, "switching-test") {
		t.Error("Expected first service name in first buffer")
	}

	// Verifica segunda saída
	output2 := buf2.String()
	if !strings.Contains(output2, "Second provider message") {
		t.Error("Expected second message in second buffer")
	}

	if !strings.Contains(output2, "Debug message") {
		t.Error("Expected debug message in second buffer")
	}

	if !strings.Contains(output2, "switching-test-2") {
		t.Error("Expected second service name in second buffer")
	}
}

func TestIntegrationHighVolumeLogging(t *testing.T) {
	// Testa logging de alto volume
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	err := logger.SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider: %v", err)
	}

	ctx := context.Background()
	iterations := 1000
	start := time.Now()

	// Logging intensivo
	for i := 0; i < iterations; i++ {
		logger.Info(ctx, "High volume test message",
			logger.Int("iteration", i),
			logger.String("batch", "performance-test"),
			logger.Bool("high_volume", true),
		)
	}

	duration := time.Since(start)
	avgDuration := duration / time.Duration(iterations)

	// Verifica performance (deve ser menor que 1ms por log)
	if avgDuration > time.Millisecond {
		t.Errorf("Average log duration too high: %v", avgDuration)
	}

	// Verifica se todos os logs foram escritos
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != iterations {
		t.Errorf("Expected %d log lines, got %d", iterations, len(lines))
	}

	// Verifica algumas linhas aleatórias
	for i := 0; i < 10; i++ {
		line := lines[i*100]
		if !strings.Contains(line, "High volume test message") {
			t.Errorf("Expected test message in line %d", i*100)
		}
	}
}

func TestIntegrationConcurrentLogging(t *testing.T) {
	// Testa logging concorrente
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	err := logger.SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider: %v", err)
	}

	ctx := context.Background()
	goroutines := 10
	messagesPerGoroutine := 100

	// Canal para sincronização
	done := make(chan bool, goroutines)

	// Lança goroutines concorrentes
	for g := 0; g < goroutines; g++ {
		go func(goroutineID int) {
			for i := 0; i < messagesPerGoroutine; i++ {
				logger.Info(ctx, "Concurrent log message",
					logger.Int("goroutine", goroutineID),
					logger.Int("message", i),
					logger.String("test", "concurrent"),
				)
			}
			done <- true
		}(g)
	}

	// Aguarda todas as goroutines terminarem
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Verifica se todos os logs foram escritos
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	expectedLines := goroutines * messagesPerGoroutine
	if len(lines) != expectedLines {
		t.Errorf("Expected %d log lines, got %d", expectedLines, len(lines))
	}

	// Verifica se logs de diferentes goroutines estão presentes
	goroutineCounts := make(map[int]int)
	for _, line := range lines {
		// Não fazemos parsing JSON completo, apenas verificamos se tem o padrão
		if strings.Contains(line, "Concurrent log message") {
			// Conta como válido
			goroutineCounts[0]++
		}
	}

	if goroutineCounts[0] != expectedLines {
		t.Errorf("Expected %d valid log lines, got %d", expectedLines, goroutineCounts[0])
	}
}

func TestIntegrationErrorHandling(t *testing.T) {
	// Testa tratamento de erros
	var buf bytes.Buffer
	config := &logger.Config{
		Level:         logger.ErrorLevel,
		Format:        logger.JSONFormat,
		Output:        &buf,
		AddStacktrace: true,
	}

	err := logger.SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider: %v", err)
	}

	ctx := context.Background()

	// Simula erro de aplicação
	appError := &ApplicationError{
		Code:    "APP_ERROR",
		Message: "Something went wrong",
		Details: map[string]string{
			"component": "integration-test",
			"operation": "error-handling",
		},
	}

	logger.Error(ctx, "Application error occurred",
		logger.ErrorField(appError),
		logger.String("error_type", "application"),
	)

	// Simula erro de sistema
	logger.Error(ctx, "System error occurred",
		logger.String("error_type", "system"),
		logger.String("subsystem", "database"),
	)

	// Verifica saída
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) < 2 {
		t.Errorf("Expected at least 2 error log lines, got %d", len(lines))
	}

	// Verifica se erros estão presentes
	if !strings.Contains(output, "Application error occurred") {
		t.Error("Expected application error in output")
	}

	if !strings.Contains(output, "System error occurred") {
		t.Error("Expected system error in output")
	}

	if !strings.Contains(output, "APP_ERROR") {
		t.Error("Expected error code in output")
	}
}

func TestIntegrationEnvironmentConfiguration(t *testing.T) {
	// Salva variáveis originais
	originalLevel := os.Getenv("LOG_LEVEL")
	originalFormat := os.Getenv("LOG_FORMAT")
	originalService := os.Getenv("SERVICE_NAME")

	// Define variáveis de ambiente
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "json")
	os.Setenv("SERVICE_NAME", "integration-env-test")

	// Testa configuração automática
	var buf bytes.Buffer
	config := logger.EnvironmentConfig()
	config.Output = &buf

	err := logger.SetProvider("slog", config)
	if err != nil {
		t.Fatalf("Failed to set provider: %v", err)
	}

	ctx := context.Background()

	// Testa se debug está habilitado
	logger.Debug(ctx, "Debug message from environment config")
	logger.Info(ctx, "Info message from environment config")

	// Verifica saída
	output := buf.String()

	if !strings.Contains(output, "Debug message from environment config") {
		t.Error("Expected debug message in output")
	}

	if !strings.Contains(output, "integration-env-test") {
		t.Error("Expected service name from environment in output")
	}

	// Restaura variáveis originais
	os.Setenv("LOG_LEVEL", originalLevel)
	os.Setenv("LOG_FORMAT", originalFormat)
	os.Setenv("SERVICE_NAME", originalService)
}

// ApplicationError simula um erro de aplicação
type ApplicationError struct {
	Code    string
	Message string
	Details map[string]string
}

func (e *ApplicationError) Error() string {
	return e.Message
}

func BenchmarkIntegrationLogging(b *testing.B) {
	var buf bytes.Buffer
	config := &logger.Config{
		Level:  logger.InfoLevel,
		Format: logger.JSONFormat,
		Output: &buf,
	}

	err := logger.SetProvider("slog", config)
	if err != nil {
		b.Fatalf("Failed to set provider: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info(ctx, "Benchmark integration message",
				logger.String("test", "integration"),
				logger.Int("iteration", 1),
				logger.Bool("benchmark", true),
			)
		}
	})
}
