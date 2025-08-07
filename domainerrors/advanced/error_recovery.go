package advanced

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// RecoveryStrategy define diferentes estratégias de recuperação
type RecoveryStrategy int

const (
	// NoRecovery sem recuperação
	NoRecovery RecoveryStrategy = iota
	// RetryStrategy tenta novamente
	RetryStrategy
	// FallbackStrategy usa valor/função fallback
	FallbackStrategy
	// CircuitBreakerStrategy para o circuito temporariamente
	CircuitBreakerStrategy
	// GracefulDegradationStrategy degrada funcionalidade graciosamente
	GracefulDegradationStrategy
)

// RecoveryConfig configuração para recuperação
type RecoveryConfig struct {
	Strategy      RecoveryStrategy
	MaxAttempts   int
	Timeout       time.Duration
	FallbackValue interface{}
	FallbackFunc  func(ctx context.Context, err error) (interface{}, error)
}

// RecoveryHandler gerenciador de recuperação de erros
type RecoveryHandler struct {
	strategies map[interfaces.ErrorType]*RecoveryConfig
	mu         sync.RWMutex
}

// NewRecoveryHandler cria um novo handler de recuperação
func NewRecoveryHandler() *RecoveryHandler {
	rh := &RecoveryHandler{
		strategies: make(map[interfaces.ErrorType]*RecoveryConfig),
	}

	// Configurações padrão para diferentes tipos de erro
	rh.setDefaultStrategies()

	return rh
}

// setDefaultStrategies define estratégias padrão
func (rh *RecoveryHandler) setDefaultStrategies() {
	// Timeout: Retry com timeout
	rh.strategies[interfaces.TimeoutError] = &RecoveryConfig{
		Strategy:    RetryStrategy,
		MaxAttempts: 3,
		Timeout:     5 * time.Second,
	}

	// External Service: Circuit Breaker
	rh.strategies[interfaces.ExternalServiceError] = &RecoveryConfig{
		Strategy:    CircuitBreakerStrategy,
		MaxAttempts: 5,
		Timeout:     30 * time.Second,
	}

	// Resource Exhausted: Graceful Degradation
	rh.strategies[interfaces.ResourceExhaustedError] = &RecoveryConfig{
		Strategy:    GracefulDegradationStrategy,
		MaxAttempts: 1,
		Timeout:     1 * time.Second,
	}

	// Service Unavailable: Fallback
	rh.strategies[interfaces.ServiceUnavailableError] = &RecoveryConfig{
		Strategy:      FallbackStrategy,
		MaxAttempts:   2,
		Timeout:       10 * time.Second,
		FallbackValue: "Service temporarily unavailable",
	}
}

// RegisterStrategy registra estratégia para tipo de erro
func (rh *RecoveryHandler) RegisterStrategy(errorType interfaces.ErrorType, config *RecoveryConfig) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	rh.strategies[errorType] = config
}

// Recover tenta recuperar de um erro usando estratégia apropriada
func (rh *RecoveryHandler) Recover(ctx context.Context, err error, operation func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	// Verificar se é DomainError
	domainErr, ok := err.(interfaces.DomainErrorInterface)
	if !ok {
		return nil, err // Não pode recuperar erros não-domain
	}

	rh.mu.RLock()
	config, exists := rh.strategies[domainErr.Type()]
	rh.mu.RUnlock()

	if !exists {
		return nil, err // Sem estratégia de recuperação
	}

	switch config.Strategy {
	case RetryStrategy:
		return rh.executeRetryRecovery(ctx, err, operation, config)
	case FallbackStrategy:
		return rh.executeFallbackRecovery(ctx, err, operation, config)
	case CircuitBreakerStrategy:
		return rh.executeCircuitBreakerRecovery(ctx, err, operation, config)
	case GracefulDegradationStrategy:
		return rh.executeGracefulDegradation(ctx, err, operation, config)
	default:
		return nil, err
	}
}

// executeRetryRecovery executa recuperação com retry
func (rh *RecoveryHandler) executeRetryRecovery(ctx context.Context, originalErr error, operation func(ctx context.Context) (interface{}, error), config *RecoveryConfig) (interface{}, error) {
	// Criar contexto com timeout se especificado
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	var lastError error = originalErr

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		result, err := operation(ctx)
		if err == nil {
			return result, nil
		}

		lastError = err

		// Se não é a última tentativa, aguarda um pouco
		if attempt < config.MaxAttempts {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * 100 * time.Millisecond):
				// Continua
			}
		}
	}

	return nil, fmt.Errorf("recovery failed after %d attempts: %w", config.MaxAttempts, lastError)
}

// executeFallbackRecovery executa recuperação com fallback
func (rh *RecoveryHandler) executeFallbackRecovery(ctx context.Context, originalErr error, operation func(ctx context.Context) (interface{}, error), config *RecoveryConfig) (interface{}, error) {
	// Primeiro tenta a operação original
	if result, err := operation(ctx); err == nil {
		return result, nil
	}

	// Se falhou, usa fallback
	if config.FallbackFunc != nil {
		return config.FallbackFunc(ctx, originalErr)
	}

	if config.FallbackValue != nil {
		return config.FallbackValue, nil
	}

	return nil, fmt.Errorf("no fallback available for error: %w", originalErr)
}

// CircuitBreaker simples para recuperação
type CircuitBreaker struct {
	state       CircuitState
	failures    int
	maxFailures int
	timeout     time.Duration
	lastFailure time.Time
	mu          sync.RWMutex
}

type CircuitState int

const (
	Closed CircuitState = iota
	Open
	HalfOpen
)

// executeCircuitBreakerRecovery executa recuperação com circuit breaker
func (rh *RecoveryHandler) executeCircuitBreakerRecovery(ctx context.Context, originalErr error, operation func(ctx context.Context) (interface{}, error), config *RecoveryConfig) (interface{}, error) {
	// Implementação simplificada de circuit breaker
	// Em produção, seria mais sofisticado

	cb := &CircuitBreaker{
		state:       Closed,
		maxFailures: config.MaxAttempts,
		timeout:     config.Timeout,
	}

	return cb.execute(ctx, operation)
}

// execute executa operação com circuit breaker
func (cb *CircuitBreaker) execute(ctx context.Context, operation func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Verificar estado atual
	if cb.state == Open {
		if time.Since(cb.lastFailure) < cb.timeout {
			return nil, fmt.Errorf("circuit breaker is open")
		}
		// Transição para Half-Open
		cb.state = HalfOpen
	}

	// Executar operação
	result, err := operation(ctx)

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()

		if cb.failures >= cb.maxFailures {
			cb.state = Open
		}

		return nil, err
	}

	// Sucesso - reset circuit breaker
	cb.failures = 0
	cb.state = Closed

	return result, nil
}

// executeGracefulDegradation executa degradação graciosa
func (rh *RecoveryHandler) executeGracefulDegradation(ctx context.Context, originalErr error, operation func(ctx context.Context) (interface{}, error), config *RecoveryConfig) (interface{}, error) {
	// Tenta operação com timeout reduzido
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	result, err := operation(ctx)
	if err == nil {
		return result, nil
	}

	// Retorna resposta degradada
	degradedResponse := map[string]interface{}{
		"status":    "degraded",
		"message":   "Service operating in degraded mode",
		"error":     originalErr.Error(),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	return degradedResponse, nil
}

// GetStrategy retorna estratégia para tipo de erro
func (rh *RecoveryHandler) GetStrategy(errorType interfaces.ErrorType) *RecoveryConfig {
	rh.mu.RLock()
	defer rh.mu.RUnlock()

	if config, exists := rh.strategies[errorType]; exists {
		// Retorna cópia para evitar modificações
		configCopy := *config
		return &configCopy
	}

	return nil
}

// ListStrategies lista todas as estratégias registradas
func (rh *RecoveryHandler) ListStrategies() map[interfaces.ErrorType]*RecoveryConfig {
	rh.mu.RLock()
	defer rh.mu.RUnlock()

	strategies := make(map[interfaces.ErrorType]*RecoveryConfig)
	for k, v := range rh.strategies {
		configCopy := *v
		strategies[k] = &configCopy
	}

	return strategies
}

// RemoveStrategy remove estratégia para tipo de erro
func (rh *RecoveryHandler) RemoveStrategy(errorType interfaces.ErrorType) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	delete(rh.strategies, errorType)
}

// RecoveryMiddleware middleware para recuperação automática
func (rh *RecoveryHandler) RecoveryMiddleware() interfaces.MiddlewareFunc {
	return func(ctx context.Context, err interfaces.DomainErrorInterface, next func(interfaces.DomainErrorInterface) interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
		// Se há estratégia de recuperação, tenta aplicar
		if config := rh.GetStrategy(err.Type()); config != nil {
			// Para middleware, apenas adiciona metadados sobre recuperação
			return err.WithMetadata("recovery_available", true).
				WithMetadata("recovery_strategy", config.Strategy)
		}

		return next(err)
	}
}

// Instância global
var globalRecoveryHandler = NewRecoveryHandler()

// Recover usa handler global para recuperação
func Recover(ctx context.Context, err error, operation func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	return globalRecoveryHandler.Recover(ctx, err, operation)
}

// RegisterGlobalStrategy registra estratégia globalmente
func RegisterGlobalStrategy(errorType interfaces.ErrorType, config *RecoveryConfig) {
	globalRecoveryHandler.RegisterStrategy(errorType, config)
}

// GetGlobalStrategy retorna estratégia global
func GetGlobalStrategy(errorType interfaces.ErrorType) *RecoveryConfig {
	return globalRecoveryHandler.GetStrategy(errorType)
}
