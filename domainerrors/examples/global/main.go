package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

// Configuração global da aplicação
type AppConfig struct {
	ServiceName   string
	Environment   string
	LogLevel      string
	EnableMetrics bool
	EnableTracing bool
}

// ErrorHandler centralizado
type ErrorHandler struct {
	config *AppConfig
	logger *log.Logger
}

func main() {
	fmt.Println("=== Exemplo Global - DomainErrors ===")

	// Configuração da aplicação
	config := &AppConfig{
		ServiceName:   "user-service",
		Environment:   "production",
		LogLevel:      "info",
		EnableMetrics: true,
		EnableTracing: true,
	}

	// Inicializar handler global
	errorHandler := NewErrorHandler(config)

	// Configurar logging global
	setupGlobalLogging(config)

	// Configurar captura de panic global
	setupGlobalPanicRecovery(errorHandler)

	// Demonstrar configuração global
	demonstrateGlobalConfiguration()

	// Demonstrar handler centralizado
	demonstrateCentralizedErrorHandling(errorHandler)

	// Demonstrar integração com contexto
	demonstrateContextIntegration(errorHandler)

	// Demonstrar customização de tipos
	demonstrateCustomErrorTypes()

	// Demonstrar configuração de stack trace
	demonstrateStackTraceConfiguration()

	fmt.Println("\n=== Exemplo global concluído! ===")
}

func NewErrorHandler(config *AppConfig) *ErrorHandler {
	return &ErrorHandler{
		config: config,
		logger: log.New(os.Stdout, fmt.Sprintf("[%s] ", config.ServiceName), log.LstdFlags|log.Lshortfile),
	}
}

func setupGlobalLogging(config *AppConfig) {
	fmt.Println("\n--- Configuração Global de Logging ---")

	// Configurar logger padrão
	log.SetPrefix(fmt.Sprintf("[%s][%s] ", config.ServiceName, config.Environment))
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Printf("✓ Logger configurado para %s em %s\n", config.ServiceName, config.Environment)
}

func setupGlobalPanicRecovery(handler *ErrorHandler) {
	fmt.Println("\n--- Configuração de Recovery Global ---")

	// Configurar captura de panic em nível de aplicação
	defer func() {
		if r := recover(); r != nil {
			// Capturar panic com stack trace
			if err := domainerrors.RecoverWithStackTrace(); err != nil {
				handler.HandleCriticalError(err)
			}
		}
	}()

	fmt.Println("✓ Recovery global configurado")
}

func demonstrateGlobalConfiguration() {
	fmt.Println("\n--- Configuração Global ---")

	// Configurar stack trace globalmente
	domainerrors.GlobalStackTraceEnabled = true
	domainerrors.GlobalMaxStackDepth = 15
	domainerrors.GlobalSkipFrames = 3

	fmt.Printf("✓ Stack trace habilitado: %v\n", domainerrors.GlobalStackTraceEnabled)
	fmt.Printf("✓ Profundidade máxima: %d\n", domainerrors.GlobalMaxStackDepth)
	fmt.Printf("✓ Frames ignorados: %d\n", domainerrors.GlobalSkipFrames)

	// Criar erro com configuração global
	err := domainerrors.New("GLOBAL_001", "Erro com configuração global")
	fmt.Printf("✓ Erro criado: %s\n", err.Error())

	// Verificar se stack trace foi capturado
	if err.StackTrace() != "" {
		fmt.Println("✓ Stack trace capturado com configuração global")
	}
}

func demonstrateCentralizedErrorHandling(handler *ErrorHandler) {
	fmt.Println("\n--- Handler Centralizado ---")

	// Simular diferentes tipos de erro
	errors := []error{
		domainerrors.NewValidationError("Dados inválidos", nil),
		domainerrors.NewBusinessError("BUSINESS_001", "Regra de negócio violada"),
		domainerrors.NewDatabaseError("Falha na consulta", nil),
		domainerrors.NewExternalServiceError("payment-api", "Serviço indisponível", nil),
	}

	for _, err := range errors {
		// Processar erro através do handler centralizado
		response := handler.HandleError(err)
		fmt.Printf("Erro processado: %+v\n", response)
	}
}

func demonstrateContextIntegration(handler *ErrorHandler) {
	fmt.Println("\n--- Integração com Contexto ---")

	// Criar contexto com informações da requisição
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", "req-12345")
	ctx = context.WithValue(ctx, "user_id", "user-67890")
	ctx = context.WithValue(ctx, "trace_id", "trace-abcdef")

	// Criar erro com contexto
	err := domainerrors.NewWithType("CTX_001", "Erro com contexto", domainerrors.ErrorTypeValidation)
	err.WithContext(ctx)

	// Enriquecer erro com informações do contexto
	enrichedErr := handler.EnrichWithContext(ctx, err)
	fmt.Printf("Erro enriquecido: %s\n", enrichedErr.Error())

	// Verificar metadados adicionados
	if domainErr, ok := enrichedErr.(*domainerrors.DomainError); ok {
		fmt.Printf("Metadados: %+v\n", domainErr.Metadata())
	}
}

func demonstrateCustomErrorTypes() {
	fmt.Println("\n--- Tipos de Erro Customizados ---")

	// Criar mapeamento customizado de tipos para HTTP status
	customMapping := map[domainerrors.ErrorType]int{
		domainerrors.ErrorTypeValidation:      400,
		domainerrors.ErrorTypeNotFound:        404,
		domainerrors.ErrorTypeBusiness:        422,
		domainerrors.ErrorTypeDatabase:        500,
		domainerrors.ErrorTypeExternalService: 502,
		domainerrors.ErrorTypeInfrastructure:  503,
	}

	// Demonstrar uso do mapeamento customizado
	for errorType, httpStatus := range customMapping {
		err := domainerrors.NewWithType("CUSTOM_001", "Erro customizado", errorType)
		fmt.Printf("Tipo: %s -> HTTP: %d (Padrão: %d)\n",
			errorType, httpStatus, err.HTTPStatus())
	}
}

func demonstrateStackTraceConfiguration() {
	fmt.Println("\n--- Configuração de Stack Trace ---")

	// Testar diferentes configurações de stack trace
	configurations := []struct {
		name       string
		enabled    bool
		maxDepth   int
		skipFrames int
	}{
		{"Padrão", true, 10, 2},
		{"Detalhado", true, 20, 1},
		{"Mínimo", true, 5, 3},
		{"Desabilitado", false, 0, 0},
	}

	for _, config := range configurations {
		fmt.Printf("\n%s:\n", config.name)

		// Aplicar configuração
		domainerrors.GlobalStackTraceEnabled = config.enabled
		domainerrors.GlobalMaxStackDepth = config.maxDepth
		domainerrors.GlobalSkipFrames = config.skipFrames

		// Criar erro para testar
		err := domainerrors.New("TRACE_001", "Teste de stack trace")

		if err.StackTrace() != "" {
			fmt.Printf("  Stack trace capturado (%d caracteres)\n", len(err.StackTrace()))
		} else {
			fmt.Printf("  Stack trace não capturado\n")
		}
	}

	// Restaurar configuração padrão
	domainerrors.GlobalStackTraceEnabled = true
	domainerrors.GlobalMaxStackDepth = 10
	domainerrors.GlobalSkipFrames = 2
}

// Implementação do ErrorHandler

func (h *ErrorHandler) HandleError(err error) map[string]interface{} {
	response := map[string]interface{}{
		"error":       true,
		"service":     h.config.ServiceName,
		"environment": h.config.Environment,
		"timestamp":   time.Now().Unix(),
	}

	if domainErr, ok := err.(*domainerrors.DomainError); ok {
		response["code"] = domainErr.Code
		response["message"] = domainErr.Message
		response["type"] = domainErr.ErrorType
		response["http_status"] = domainErr.HTTPStatus()

		// Adicionar metadados se existirem
		if metadata := domainErr.Metadata(); len(metadata) > 0 {
			response["metadata"] = metadata
		}

		// Adicionar stack trace em ambiente de desenvolvimento
		if h.config.Environment == "development" && domainErr.StackTrace() != "" {
			response["stack_trace"] = domainErr.StackTrace()
		}
	} else {
		response["message"] = err.Error()
		response["type"] = "unknown"
		response["http_status"] = 500
	}

	// Log do erro
	h.logError(err, response)

	// Enviar métricas se habilitado
	if h.config.EnableMetrics {
		h.recordMetrics(err)
	}

	// Enviar tracing se habilitado
	if h.config.EnableTracing {
		h.recordTrace(err)
	}

	return response
}

func (h *ErrorHandler) HandleCriticalError(err error) {
	h.logger.Printf("CRITICAL ERROR: %s", err.Error())

	// Em produção, poderia enviar alerta
	if h.config.Environment == "production" {
		h.sendAlert(err)
	}

	// Processar erro normalmente
	h.HandleError(err)
}

func (h *ErrorHandler) EnrichWithContext(ctx context.Context, err error) error {
	if domainErr, ok := err.(*domainerrors.DomainError); ok {
		// Adicionar informações do contexto
		if requestID, ok := ctx.Value("request_id").(string); ok {
			domainErr.WithMetadata("request_id", requestID)
		}

		if userID, ok := ctx.Value("user_id").(string); ok {
			domainErr.WithMetadata("user_id", userID)
		}

		if traceID, ok := ctx.Value("trace_id").(string); ok {
			domainErr.WithMetadata("trace_id", traceID)
		}

		// Adicionar informações do serviço
		domainErr.WithMetadata("service", h.config.ServiceName)
		domainErr.WithMetadata("environment", h.config.Environment)

		return domainErr
	}

	return err
}

func (h *ErrorHandler) logError(err error, response map[string]interface{}) {
	logLevel := h.determineLogLevel(err)

	switch logLevel {
	case "error":
		h.logger.Printf("ERROR: %s | Response: %+v", err.Error(), response)
	case "warn":
		h.logger.Printf("WARN: %s | Response: %+v", err.Error(), response)
	case "info":
		h.logger.Printf("INFO: %s | Response: %+v", err.Error(), response)
	}
}

func (h *ErrorHandler) determineLogLevel(err error) string {
	if domainErr, ok := err.(*domainerrors.DomainError); ok {
		severity := domainerrors.GetSeverity(domainErr)
		switch severity {
		case "high":
			return "error"
		case "medium":
			return "warn"
		default:
			return "info"
		}
	}
	return "error"
}

func (h *ErrorHandler) recordMetrics(err error) {
	// Simulação de envio de métricas
	if domainErr, ok := err.(*domainerrors.DomainError); ok {
		fmt.Printf("📊 Métrica: error.count{type=%s,service=%s} +1\n",
			domainErr.ErrorType, h.config.ServiceName)
	}
}

func (h *ErrorHandler) recordTrace(err error) {
	// Simulação de envio de tracing
	if domainErr, ok := err.(*domainerrors.DomainError); ok {
		if traceID, exists := domainErr.Metadata()["trace_id"]; exists {
			fmt.Printf("🔍 Trace: %s error recorded\n", traceID)
		}
	}
}

func (h *ErrorHandler) sendAlert(err error) {
	// Simulação de envio de alerta
	fmt.Printf("🚨 ALERTA: Erro crítico detectado: %s\n", err.Error())
}
