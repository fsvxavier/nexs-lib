package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/domainerrors/hooks"
	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/domainerrors/middlewares"
)

// ErrorMetrics simula um sistema de métricas
type ErrorMetrics struct {
	mu     sync.RWMutex
	counts map[string]int
}

func NewErrorMetrics() *ErrorMetrics {
	return &ErrorMetrics{
		counts: make(map[string]int),
	}
}

func (m *ErrorMetrics) IncrementError(errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counts[errorType]++
}

func (m *ErrorMetrics) GetCounts() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]int)
	for k, v := range m.counts {
		result[k] = v
	}
	return result
}

// AuditLogger simula um sistema de audit
type AuditLogger struct {
	logs []AuditEntry
	mu   sync.Mutex
}

type AuditEntry struct {
	Timestamp time.Time
	Code      string
	Message   string
	Context   map[string]interface{}
}

func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		logs: make([]AuditEntry, 0),
	}
}

func (a *AuditLogger) LogError(code, message string, context map[string]interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.logs = append(a.logs, AuditEntry{
		Timestamp: time.Now(),
		Code:      code,
		Message:   message,
		Context:   context,
	})
}

func (a *AuditLogger) GetLogs() []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	result := make([]AuditEntry, len(a.logs))
	copy(result, a.logs)
	return result
}

// CircuitBreaker simula um circuit breaker
type CircuitBreaker struct {
	mu           sync.RWMutex
	failures     int
	threshold    int
	resetTimeout time.Duration
	lastFailure  time.Time
	state        string
}

func NewCircuitBreaker(threshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold:    threshold,
		resetTimeout: resetTimeout,
		state:        "closed",
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.failures >= cb.threshold {
		cb.state = "open"
		fmt.Printf("🔴 Circuit Breaker: Estado alterado para OPEN após %d falhas\n", cb.failures)
	}
}

func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.state == "closed" {
		return true
	}

	if time.Since(cb.lastFailure) > cb.resetTimeout {
		return true // Permite tentativa para half-open
	}

	return false
}

func (cb *CircuitBreaker) GetState() (string, int) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state, cb.failures
}

// Instâncias globais dos componentes
var (
	errorMetrics   = NewErrorMetrics()
	auditLogger    = NewAuditLogger()
	circuitBreaker = NewCircuitBreaker(3, 30*time.Second)
)

func init() {
	setupAdvancedHooks()
	setupAdvancedMiddlewares()
}

func setupAdvancedHooks() {
	// Hook de inicialização com verificação de dependências
	hooks.RegisterGlobalStartHook(func(ctx context.Context) error {
		fmt.Println("🔧 Sistema: Verificando dependências...")

		// Simular verificação de dependências
		dependencies := []string{"database", "redis", "external-api"}
		for _, dep := range dependencies {
			fmt.Printf("   ✅ %s: OK\n", dep)
		}

		fmt.Println("🚀 Sistema: Todas as dependências verificadas, sistema pronto!")
		return nil
	})

	// Hook de parada com cleanup
	hooks.RegisterGlobalStopHook(func(ctx context.Context) error {
		fmt.Println("🧹 Sistema: Executando cleanup...")

		// Mostrar estatísticas finais
		counts := errorMetrics.GetCounts()
		fmt.Printf("   📊 Total de erros processados: %d tipos\n", len(counts))

		logs := auditLogger.GetLogs()
		fmt.Printf("   📝 Total de entradas de audit: %d\n", len(logs))

		fmt.Println("🛑 Sistema: Cleanup finalizado, sistema parado!")
		return nil
	})

	// Hook de erro com métricas e circuit breaker
	hooks.RegisterGlobalErrorHook(func(ctx context.Context, err interfaces.DomainErrorInterface) error {
		errorType := string(err.Type())

		// Incrementar métricas
		errorMetrics.IncrementError(errorType)

		// Registrar no circuit breaker para erros críticos
		if isCriticalError(err.Type()) {
			circuitBreaker.RecordFailure()
		}

		// Log com severity baseada no tipo de erro
		severity := getSeverity(err.Type())
		fmt.Printf("%s Error Hook: [%s] %s - %s\n",
			getSeverityIcon(severity), severity, err.Code(), err.Error())

		return nil
	})

	// Hook de i18n com detecção automática de locale
	hooks.RegisterGlobalI18nHook(func(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error {
		fmt.Printf("🌍 I18n Hook: Processando %s para %s (auto-detectado: %s)\n",
			err.Code(), locale, detectPreferredLocale(ctx))
		return nil
	})
}

func setupAdvancedMiddlewares() {
	// Middleware de enriquecimento com contexto
	middlewares.RegisterGlobalMiddleware(func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
		fmt.Printf("💼 Context Middleware: Enriquecendo erro %s\n", err.Code())

		enriched := err.
			WithMetadata("request_id", getRequestID(ctx)).
			WithMetadata("user_id", getUserID(ctx)).
			WithMetadata("correlation_id", getCorrelationID(ctx)).
			WithMetadata("service", "domainerrors-example").
			WithMetadata("environment", "development").
			WithMetadata("processing_time", time.Now())

		return next(enriched)
	})

	// Middleware de rate limiting
	middlewares.RegisterGlobalMiddleware(func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
		fmt.Printf("🚦 Rate Limit Middleware: Verificando %s\n", err.Code())

		// Simular verificação de rate limit
		if shouldRateLimit(err.Type()) {
			rateLimitedErr := domainerrors.NewRateLimitError(
				"RATE_LIMITED_"+err.Code(),
				"Rate limit exceeded for error type: "+string(err.Type()),
			).WithMetadata("original_error", err.Code()).
				WithMetadata("rate_limit_policy", "5_per_minute")

			return next(rateLimitedErr)
		}

		return next(err)
	})

	// Middleware de audit avançado
	middlewares.RegisterGlobalMiddleware(func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
		fmt.Printf("📋 Audit Middleware: Registrando %s\n", err.Code())

		// Registrar no audit log
		auditContext := make(map[string]interface{})
		if metadata := err.Metadata(); metadata != nil {
			for k, v := range metadata {
				auditContext[k] = v
			}
		}
		auditContext["severity"] = getSeverity(err.Type())
		auditContext["http_status"] = err.HTTPStatus()

		auditLogger.LogError(err.Code(), err.Error(), auditContext)

		return next(err)
	})

	// Middleware de i18n avançado com fallback
	middlewares.RegisterGlobalI18nMiddleware(func(ctx context.Context, err interfaces.DomainErrorInterface, locale string, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
		fmt.Printf("🌐 Advanced I18n Middleware: %s -> %s\n", err.Code(), locale)

		translated := translateMessage(err.Error(), locale)

		// Adicionar metadados de tradução
		translatedErr := domainerrors.NewWithMetadata(
			err.Type(),
			err.Code(),
			translated,
			err.Metadata(),
		).WithMetadata("translated", true).
			WithMetadata("original_message", err.Error()).
			WithMetadata("target_locale", locale).
			WithMetadata("translation_confidence", getTranslationConfidence(locale))

		return next(translatedErr)
	})
}

// Funções auxiliares
func isCriticalError(errorType interfaces.ErrorType) bool {
	criticalTypes := []interfaces.ErrorType{
		interfaces.DatabaseError,
		interfaces.ExternalServiceError,
		interfaces.InfrastructureError,
		interfaces.SecurityError,
	}

	for _, ct := range criticalTypes {
		if errorType == ct {
			return true
		}
	}
	return false
}

func getSeverity(errorType interfaces.ErrorType) string {
	switch errorType {
	case interfaces.ValidationError, interfaces.BadRequestError:
		return "LOW"
	case interfaces.NotFoundError, interfaces.AuthenticationError:
		return "MEDIUM"
	case interfaces.BusinessError, interfaces.AuthorizationError:
		return "HIGH"
	case interfaces.DatabaseError, interfaces.ExternalServiceError, interfaces.SecurityError:
		return "CRITICAL"
	default:
		return "MEDIUM"
	}
}

func getSeverityIcon(severity string) string {
	switch severity {
	case "LOW":
		return "🟢"
	case "MEDIUM":
		return "🟡"
	case "HIGH":
		return "🟠"
	case "CRITICAL":
		return "🔴"
	default:
		return "⚪"
	}
}

func shouldRateLimit(errorType interfaces.ErrorType) bool {
	// Simular rate limiting para validation errors
	return errorType == interfaces.ValidationError
}

func detectPreferredLocale(ctx context.Context) string {
	// Simular detecção de locale do contexto
	return "pt-BR"
}

func getRequestID(ctx context.Context) string {
	// Simular extração de request ID do contexto
	return "req-12345"
}

func getUserID(ctx context.Context) string {
	// Simular extração de user ID do contexto
	return "user-67890"
}

func getCorrelationID(ctx context.Context) string {
	// Simular geração de correlation ID
	return fmt.Sprintf("corr-%d", time.Now().UnixNano())
}

func translateMessage(message, locale string) string {
	translations := map[string]map[string]string{
		"pt-BR": {
			"Campo obrigatório não informado": "Campo obrigatório não foi informado",
			"Invalid user credentials":        "Credenciais de usuário inválidas",
			"Resource not found":              "Recurso não encontrado",
		},
		"es-ES": {
			"Campo obrigatório não informado": "Campo obligatorio no informado",
			"Invalid user credentials":        "Credenciales de usuario inválidas",
			"Resource not found":              "Recurso no encontrado",
		},
		"en-US": {
			"Campo obrigatório não informado": "Required field not provided",
			"Invalid user credentials":        "Invalid user credentials",
			"Resource not found":              "Resource not found",
		},
	}

	if localeMap, exists := translations[locale]; exists {
		if translated, exists := localeMap[message]; exists {
			return translated
		}
	}

	return message // Fallback para mensagem original
}

func getTranslationConfidence(locale string) float64 {
	// Simular confiança na tradução baseada no locale
	switch locale {
	case "pt-BR":
		return 1.0
	case "en-US":
		return 0.95
	case "es-ES":
		return 0.90
	default:
		return 0.5
	}
}

func main() {
	ctx := context.Background()

	fmt.Println("=== Exemplo Avançado de Domain Errors ===\n")

	// 1. Inicializar sistema
	fmt.Println("1. Inicializando sistema avançado:")
	if err := hooks.ExecuteGlobalStartHooks(ctx); err != nil {
		log.Printf("Erro ao inicializar: %v", err)
	}
	fmt.Print("\n")

	// 2. Simular diferentes tipos de erro
	fmt.Println("2. Simulando diferentes cenários de erro:")

	errors := []interfaces.DomainErrorInterface{
		domainerrors.NewValidationError("FIELD_REQUIRED", "Campo obrigatório não informado").
			WithMetadata("field", "email"),

		domainerrors.NewAuthenticationError("INVALID_CREDENTIALS", "Invalid user credentials").
			WithMetadata("attempt", 3).
			WithMetadata("ip", "192.168.1.100"),

		domainerrors.NewNotFoundError("USER_NOT_FOUND", "Resource not found").
			WithMetadata("resource_id", "user-123").
			WithMetadata("resource_type", "user"),

		domainerrors.NewDatabaseError("CONNECTION_TIMEOUT", "Database connection timeout").
			WithMetadata("timeout_ms", 5000).
			WithMetadata("retry_count", 3),

		domainerrors.NewBusinessError("INSUFFICIENT_FUNDS", "Insufficient account balance").
			WithMetadata("account_id", "acc-456").
			WithMetadata("balance", 100.50).
			WithMetadata("required", 250.00),
	}

	// 3. Processar cada erro
	for i, err := range errors {
		fmt.Printf("\n--- Processando Erro %d ---\n", i+1)
		fmt.Printf("Tipo: %s | Código: %s\n", err.Type(), err.Code())

		// Processar através dos middlewares
		processed := middlewares.ExecuteGlobalMiddlewares(ctx, err)

		// Executar hooks de erro
		if hookErr := hooks.ExecuteGlobalErrorHooks(ctx, processed); hookErr != nil {
			log.Printf("Erro no hook: %v", hookErr)
		}

		// Demonstrar i18n para o primeiro erro
		if i == 0 {
			fmt.Println("\nDemonstrando I18n:")
			locales := []string{"pt-BR", "en-US", "es-ES"}
			for _, locale := range locales {
				translated := middlewares.ExecuteGlobalI18nMiddlewares(ctx, err, locale)
				if hookErr := hooks.ExecuteGlobalI18nHooks(ctx, translated, locale); hookErr != nil {
					log.Printf("Erro no hook i18n: %v", hookErr)
				}
				fmt.Printf("  %s: %s\n", locale, translated.Error())
			}
		}
	}

	// 4. Mostrar estatísticas avançadas
	fmt.Println("\n4. Estatísticas do Sistema:")

	// Métricas de erro
	fmt.Println("\n📊 Métricas de Erro:")
	for errorType, count := range errorMetrics.GetCounts() {
		fmt.Printf("  %s: %d ocorrências\n", errorType, count)
	}

	// Estado do Circuit Breaker
	state, failures := circuitBreaker.GetState()
	fmt.Printf("\n🔄 Circuit Breaker: Estado=%s, Falhas=%d\n", state, failures)

	// Logs de Audit
	auditLogs := auditLogger.GetLogs()
	fmt.Printf("\n📝 Audit Log (%d entradas):\n", len(auditLogs))
	for i, entry := range auditLogs {
		if i < 3 { // Mostrar apenas as primeiras 3
			fmt.Printf("  [%s] %s: %s\n",
				entry.Timestamp.Format("15:04:05"),
				entry.Code,
				entry.Message)
		}
	}
	if len(auditLogs) > 3 {
		fmt.Printf("  ... e mais %d entradas\n", len(auditLogs)-3)
	}

	// 5. Teste de Circuit Breaker
	fmt.Println("\n5. Testando Circuit Breaker:")
	fmt.Printf("Pode executar? %v\n", circuitBreaker.CanExecute())

	// 6. Finalizar sistema
	fmt.Println("\n6. Finalizando sistema:")
	if err := hooks.ExecuteGlobalStopHooks(ctx); err != nil {
		log.Printf("Erro ao finalizar: %v", err)
	}

	fmt.Println("\n=== Fim do exemplo avançado ===")
}
