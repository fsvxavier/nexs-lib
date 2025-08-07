package middlewares

import (
	"context"
	"sync"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// MiddlewareManager implementa a interface interfaces.MiddlewareManager
type MiddlewareManager struct {
	middlewares     []interfaces.MiddlewareFunc
	i18nMiddlewares []interfaces.I18nMiddlewareFunc
	mu              sync.RWMutex
}

// NewMiddlewareManager cria um novo gerenciador de middlewares completo
func NewMiddlewareManager() *MiddlewareManager {
	return &MiddlewareManager{
		middlewares:     make([]interfaces.MiddlewareFunc, 0),
		i18nMiddlewares: make([]interfaces.I18nMiddlewareFunc, 0),
	}
}

// RegisterMiddleware registra um middleware
func (m *MiddlewareManager) RegisterMiddleware(middleware interfaces.MiddlewareFunc) {
	if middleware == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.middlewares = append(m.middlewares, middleware)
}

// RegisterI18nMiddleware registra um middleware de i18n
func (m *MiddlewareManager) RegisterI18nMiddleware(middleware interfaces.I18nMiddlewareFunc) {
	if middleware == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.i18nMiddlewares = append(m.i18nMiddlewares, middleware)
}

// ExecuteMiddlewares executa todos os middlewares registrados
func (m *MiddlewareManager) ExecuteMiddlewares(ctx context.Context, err interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
	if err == nil {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	result := err

	// Executa os middlewares em sequência
	for _, middleware := range m.middlewares {
		result = middleware(ctx, result, func(e interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			return e
		})
	}

	return result
}

// ExecuteI18nMiddlewares executa todos os middlewares de i18n registrados
func (m *MiddlewareManager) ExecuteI18nMiddlewares(ctx context.Context, err interfaces.DomainErrorInterface, locale string) interfaces.DomainErrorInterface {
	if err == nil {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	result := err

	// Executa os middlewares de i18n em sequência
	for _, middleware := range m.i18nMiddlewares {
		result = middleware(ctx, result, locale, func(e interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			return e
		})
	}

	return result
}

// Clear limpa todos os middlewares registrados
func (m *MiddlewareManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.middlewares = make([]interfaces.MiddlewareFunc, 0)
	m.i18nMiddlewares = make([]interfaces.I18nMiddlewareFunc, 0)
}

// GetCounts retorna o número de middlewares registrados por tipo
func (m *MiddlewareManager) GetCounts() (general, i18n int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.middlewares), len(m.i18nMiddlewares)
}

// Instância global do MiddlewareManager
var GlobalMiddlewareManager = NewMiddlewareManager()

// Funções globais de conveniência
func RegisterGlobalMiddleware(middleware interfaces.MiddlewareFunc) {
	GlobalMiddlewareManager.RegisterMiddleware(middleware)
}

func RegisterGlobalI18nMiddleware(middleware interfaces.I18nMiddlewareFunc) {
	GlobalMiddlewareManager.RegisterI18nMiddleware(middleware)
}

func ExecuteGlobalMiddlewares(ctx context.Context, err interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
	return GlobalMiddlewareManager.ExecuteMiddlewares(ctx, err)
}

func ExecuteGlobalI18nMiddlewares(ctx context.Context, err interfaces.DomainErrorInterface, locale string) interfaces.DomainErrorInterface {
	return GlobalMiddlewareManager.ExecuteI18nMiddlewares(ctx, err, locale)
}

func GetGlobalMiddlewareCounts() (general, i18n int) {
	return GlobalMiddlewareManager.GetCounts()
}

func ClearGlobalMiddlewares() {
	GlobalMiddlewareManager.Clear()
}

// Middlewares padrão

// LoggingMiddleware é um middleware de exemplo para logging
func LoggingMiddleware(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
	if err == nil {
		return next(nil)
	}

	// Adiciona informações de logging
	enrichedErr := err.WithMetadata("middleware_logged", true)
	enrichedErr = enrichedErr.WithMetadata("log_level", "error")

	return next(enrichedErr)
}

// MetricsMiddleware é um middleware de exemplo para métricas
func MetricsMiddleware(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
	if err == nil {
		return next(nil)
	}

	// Adiciona informações de métricas
	enrichedErr := err.WithMetadata("metrics_collected", true)
	enrichedErr = enrichedErr.WithMetadata("metric_type", string(err.Type()))

	return next(enrichedErr)
}

// EnrichmentMiddleware adiciona informações contextuais ao erro
func EnrichmentMiddleware(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
	if err == nil {
		return next(nil)
	}

	// Adiciona informações do contexto
	enrichedErr := err.WithMetadata("context_enriched", true)

	// Pode adicionar informações do contexto como user_id, request_id, etc.
	if ctx != nil {
		if userID := ctx.Value("user_id"); userID != nil {
			enrichedErr = enrichedErr.WithMetadata("user_id", userID)
		}
		if requestID := ctx.Value("request_id"); requestID != nil {
			enrichedErr = enrichedErr.WithMetadata("request_id", requestID)
		}
	}

	return next(enrichedErr)
}
