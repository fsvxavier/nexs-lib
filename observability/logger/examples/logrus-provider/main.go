package main

import (
	"context"
	"log"
	"os"

	"github.com/fsvxavier/nexs-lib/observability/logger/interfaces"
	logrusProvider "github.com/fsvxavier/nexs-lib/observability/logger/providers/logrus"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	// Exemplo 1: Provider básico com configuração padrão
	log.Println("=== Exemplo 1: Provider Básico ===")
	basicProvider := logrusProvider.NewProvider()

	// Logs básicos
	basicProvider.Info(ctx, "Aplicação iniciada",
		interfaces.Field{Key: "version", Value: "1.0.0"},
		interfaces.Field{Key: "environment", Value: "development"},
	)

	basicProvider.Debug(ctx, "Modo debug ativado")
	basicProvider.Warn(ctx, "Esta é uma mensagem de warning")
	basicProvider.Error(ctx, "Erro simulado",
		interfaces.Field{Key: "error_code", Value: "E001"},
	)

	// Exemplo 2: Provider com configuração personalizada
	log.Println("\n=== Exemplo 2: Provider Configurado ===")
	config := &interfaces.Config{
		Level:          interfaces.DebugLevel,
		Format:         interfaces.JSONFormat,
		ServiceName:    "exemplo-logrus",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		Fields: map[string]any{
			"component": "exemplo-logrus",
		},
	}

	configuredProvider, err := logrusProvider.NewWithConfig(config)
	if err != nil {
		log.Fatalf("Erro ao criar provider configurado: %v", err)
	}

	configuredProvider.Info(ctx, "Provider configurado criado com sucesso")

	// Exemplo 3: Logs formatados
	log.Println("\n=== Exemplo 3: Logs Formatados ===")
	configuredProvider.Infof(ctx, "Usuário %s logado com sucesso", "admin")
	configuredProvider.Debugf(ctx, "Processando %d itens", 42)

	// Exemplo 4: Logs com códigos
	log.Println("\n=== Exemplo 4: Logs com Códigos ===")
	configuredProvider.InfoWithCode(ctx, "USER_LOGIN", "Login realizado",
		interfaces.Field{Key: "user_id", Value: "123"},
		interfaces.Field{Key: "ip", Value: "192.168.1.1"},
	)

	configuredProvider.ErrorWithCode(ctx, "DB_CONNECTION_ERROR",
		"Falha na conexão com banco de dados",
		interfaces.Field{Key: "database", Value: "postgres"},
		interfaces.Field{Key: "host", Value: "localhost"},
	)

	// Exemplo 5: Logger com campos adicionais
	log.Println("\n=== Exemplo 5: Logger com Campos Adicionais ===")
	enrichedLogger := configuredProvider.WithFields(
		interfaces.Field{Key: "request_id", Value: "req-123"},
		interfaces.Field{Key: "correlation_id", Value: "corr-456"},
	)

	enrichedLogger.Info(ctx, "Processando requisição")
	enrichedLogger.Warn(ctx, "Cache miss detectado")

	// Exemplo 6: Uso de logger Logrus existente
	log.Println("\n=== Exemplo 6: Integrando Logger Logrus Existente ===")
	existingLogrus := logrus.New()
	existingLogrus.SetLevel(logrus.WarnLevel)
	existingLogrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	existingLogrus.SetOutput(os.Stdout)

	wrappedProvider := logrusProvider.NewProviderWithLogger(existingLogrus)
	wrappedProvider.Warn(ctx, "Logger Logrus existente integrado")
	wrappedProvider.Error(ctx, "Erro usando logger existente")

	// Exemplo 7: Hooks do Logrus
	log.Println("\n=== Exemplo 7: Hooks do Logrus ===")
	hookProvider := logrusProvider.NewProvider()

	// Adiciona hook personalizado
	hook := &CustomHook{}
	hookProvider.AddHook(hook)

	hookProvider.Info(ctx, "Mensagem com hook personalizado",
		interfaces.Field{Key: "hook_test", Value: true},
	)

	// Exemplo 8: Diferentes formatos
	log.Println("\n=== Exemplo 8: Diferentes Formatos ===")

	// JSON Provider
	jsonProvider := logrusProvider.NewJSONProvider()
	jsonProvider.Info(ctx, "Log em formato JSON")

	// Text Provider
	textProvider := logrusProvider.NewTextProvider()
	textProvider.Info(ctx, "Log em formato texto")

	// Console Provider
	consoleProvider := logrusProvider.NewConsoleProvider()
	consoleProvider.Info(ctx, "Log para console")

	// Exemplo 9: Provider com buffer
	log.Println("\n=== Exemplo 9: Provider com Buffer ===")
	bufferedProvider := logrusProvider.NewBufferedProvider(100, 5000000000) // 5 segundos

	for i := 0; i < 5; i++ {
		bufferedProvider.Info(ctx, "Log com buffer",
			interfaces.Field{Key: "iteration", Value: i},
		)
	}

	// Força flush do buffer
	bufferedProvider.FlushBuffer()

	// Exemplo 10: Diferentes níveis de log
	log.Println("\n=== Exemplo 10: Diferentes Níveis ===")
	configuredProvider.SetLevel(interfaces.DebugLevel)

	configuredProvider.Debug(ctx, "Informação de debug detalhada")
	configuredProvider.Info(ctx, "Informação geral")
	configuredProvider.Warn(ctx, "Situação que requer atenção")
	configuredProvider.Error(ctx, "Erro que precisa ser tratado")

	// Exemplo 11: Clonagem de logger
	log.Println("\n=== Exemplo 11: Clonagem de Logger ===")
	originalLogger := configuredProvider.WithFields(
		interfaces.Field{Key: "component", Value: "original"},
	)

	clonedLogger := originalLogger.Clone().(*logrusProvider.Provider)
	clonedLogger.Info(ctx, "Logger clonado funcionando")

	log.Println("\n=== Exemplo Logrus Provider concluído com sucesso! ===")
}

// CustomHook exemplo de hook personalizado para Logrus
type CustomHook struct{}

func (h *CustomHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *CustomHook) Fire(entry *logrus.Entry) error {
	// Adiciona timestamp personalizado
	entry.Data["custom_timestamp"] = entry.Time.Unix()
	entry.Data["hook_applied"] = true
	return nil
}
