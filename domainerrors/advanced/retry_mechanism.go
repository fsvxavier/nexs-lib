package advanced

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// RetryConfig configuração para o mecanismo de retry
type RetryConfig struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	Jitter          bool
	JitterFactor    float64
	RetryableErrors []interfaces.ErrorType
}

// DefaultRetryConfig retorna configuração padrão de retry
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
		JitterFactor:  0.1,
		RetryableErrors: []interfaces.ErrorType{
			interfaces.TimeoutError,
			interfaces.ExternalServiceError,
			interfaces.InfrastructureError,
			interfaces.ServiceUnavailableError,
			interfaces.ResourceExhaustedError,
		},
	}
}

// RetryableFunction tipo para funções que podem ser retentadas
type RetryableFunction func(ctx context.Context) error

// RetryHandler implementa mecanismo de retry com backoff
type RetryHandler struct {
	config *RetryConfig
	mu     sync.RWMutex
}

// NewRetryHandler cria um novo handler de retry
func NewRetryHandler(config *RetryConfig) *RetryHandler {
	if config == nil {
		config = DefaultRetryConfig()
	}

	return &RetryHandler{
		config: config,
	}
}

// Execute executa uma função com retry automático
func (rh *RetryHandler) Execute(ctx context.Context, fn RetryableFunction) error {
	var lastError error

	for attempt := 1; attempt <= rh.config.MaxAttempts; attempt++ {
		// Executar função
		err := fn(ctx)
		if err == nil {
			return nil // Sucesso
		}

		lastError = err

		// Verificar se erro é retentável
		if !rh.isRetryableError(err) {
			return err // Não é retentável, retorna imediatamente
		}

		// Se é a última tentativa, não faz delay
		if attempt >= rh.config.MaxAttempts {
			break
		}

		// Calcular delay
		delay := rh.calculateDelay(attempt)

		// Aguardar ou cancelar se contexto for cancelado
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continua para próxima tentativa
		}
	}

	// Todas as tentativas falharam, retorna último erro
	return fmt.Errorf("retry failed after %d attempts: %w", rh.config.MaxAttempts, lastError)
}

// ExecuteWithResult executa função com retry e retorna resultado
func (rh *RetryHandler) ExecuteWithResult(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	var lastError error
	var result interface{}

	for attempt := 1; attempt <= rh.config.MaxAttempts; attempt++ {
		// Executar função
		res, err := fn(ctx)
		if err == nil {
			return res, nil // Sucesso
		}

		lastError = err
		result = res

		// Verificar se erro é retentável
		if !rh.isRetryableError(err) {
			return result, err
		}

		// Se é a última tentativa, não faz delay
		if attempt >= rh.config.MaxAttempts {
			break
		}

		// Calcular delay
		delay := rh.calculateDelay(attempt)

		// Aguardar ou cancelar se contexto for cancelado
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(delay):
			// Continua para próxima tentativa
		}
	}

	return result, fmt.Errorf("retry failed after %d attempts: %w", rh.config.MaxAttempts, lastError)
}

// isRetryableError verifica se um erro é retentável
func (rh *RetryHandler) isRetryableError(err error) bool {
	// Tentar converter para DomainError
	if domainErr, ok := err.(interfaces.DomainErrorInterface); ok {
		for _, retryableType := range rh.config.RetryableErrors {
			if domainErr.Type() == retryableType {
				return true
			}
		}
		return false
	}

	// Para erros que não são DomainError, aplicar lógica padrão
	// Pode ser estendido conforme necessário
	return false
}

// calculateDelay calcula o delay para próxima tentativa
func (rh *RetryHandler) calculateDelay(attempt int) time.Duration {
	// Exponential backoff
	delay := float64(rh.config.InitialDelay) * math.Pow(rh.config.BackoffFactor, float64(attempt-1))

	// Aplicar jitter se configurado
	if rh.config.Jitter {
		jitterRange := delay * rh.config.JitterFactor
		jitter := (rand.Float64() - 0.5) * 2 * jitterRange // -jitterRange a +jitterRange
		delay += jitter
	}

	// Garantir que não exceda MaxDelay
	if delay > float64(rh.config.MaxDelay) {
		delay = float64(rh.config.MaxDelay)
	}

	// Garantir que não seja negativo
	if delay < 0 {
		delay = float64(rh.config.InitialDelay)
	}

	return time.Duration(delay)
}

// UpdateConfig atualiza configuração do retry handler
func (rh *RetryHandler) UpdateConfig(config *RetryConfig) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	rh.config = config
}

// GetConfig retorna configuração atual
func (rh *RetryHandler) GetConfig() *RetryConfig {
	rh.mu.RLock()
	defer rh.mu.RUnlock()

	// Retorna cópia para evitar modificações externas
	configCopy := *rh.config
	configCopy.RetryableErrors = make([]interfaces.ErrorType, len(rh.config.RetryableErrors))
	copy(configCopy.RetryableErrors, rh.config.RetryableErrors)

	return &configCopy
}

// RetryStats estatísticas de retry
type RetryStats struct {
	TotalAttempts     int
	SuccessfulRetries int
	FailedRetries     int
	AverageDelay      time.Duration
}

// RetryHandlerWithStats handler com estatísticas
type RetryHandlerWithStats struct {
	*RetryHandler
	stats RetryStats
	mu    sync.RWMutex
}

// NewRetryHandlerWithStats cria handler com estatísticas
func NewRetryHandlerWithStats(config *RetryConfig) *RetryHandlerWithStats {
	return &RetryHandlerWithStats{
		RetryHandler: NewRetryHandler(config),
	}
}

// Execute executa com coleta de estatísticas
func (rhs *RetryHandlerWithStats) Execute(ctx context.Context, fn RetryableFunction) error {
	start := time.Now()
	attempts := 0

	var lastError error

	for attempt := 1; attempt <= rhs.config.MaxAttempts; attempt++ {
		attempts++
		rhs.updateStats(func(s *RetryStats) { s.TotalAttempts++ })

		err := fn(ctx)
		if err == nil {
			// Sucesso
			if attempt > 1 {
				rhs.updateStats(func(s *RetryStats) { s.SuccessfulRetries++ })
			}
			rhs.updateAverageDelay(time.Since(start))
			return nil
		}

		lastError = err

		if !rhs.isRetryableError(err) {
			rhs.updateStats(func(s *RetryStats) { s.FailedRetries++ })
			return err
		}

		if attempt >= rhs.config.MaxAttempts {
			break
		}

		delay := rhs.calculateDelay(attempt)

		select {
		case <-ctx.Done():
			rhs.updateStats(func(s *RetryStats) { s.FailedRetries++ })
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	rhs.updateStats(func(s *RetryStats) { s.FailedRetries++ })
	rhs.updateAverageDelay(time.Since(start))

	return fmt.Errorf("retry failed after %d attempts: %w", attempts, lastError)
}

// updateStats atualiza estatísticas thread-safe
func (rhs *RetryHandlerWithStats) updateStats(fn func(*RetryStats)) {
	rhs.mu.Lock()
	defer rhs.mu.Unlock()
	fn(&rhs.stats)
}

// updateAverageDelay atualiza delay médio
func (rhs *RetryHandlerWithStats) updateAverageDelay(duration time.Duration) {
	rhs.mu.Lock()
	defer rhs.mu.Unlock()

	totalOps := rhs.stats.SuccessfulRetries + rhs.stats.FailedRetries
	if totalOps == 0 {
		rhs.stats.AverageDelay = duration
		return
	}

	// Calcular média móvel simples
	currentAvg := int64(rhs.stats.AverageDelay)
	newAvg := (currentAvg*int64(totalOps-1) + int64(duration)) / int64(totalOps)
	rhs.stats.AverageDelay = time.Duration(newAvg)
}

// GetStats retorna estatísticas atuais
func (rhs *RetryHandlerWithStats) GetStats() RetryStats {
	rhs.mu.RLock()
	defer rhs.mu.RUnlock()
	return rhs.stats
}

// ResetStats reseta as estatísticas
func (rhs *RetryHandlerWithStats) ResetStats() {
	rhs.mu.Lock()
	defer rhs.mu.Unlock()
	rhs.stats = RetryStats{}
}

// Instância global de retry handler
var globalRetryHandler = NewRetryHandlerWithStats(DefaultRetryConfig())

// ExecuteWithRetry executa função com retry global
func ExecuteWithRetry(ctx context.Context, fn RetryableFunction) error {
	return globalRetryHandler.Execute(ctx, fn)
}

// ExecuteWithRetryAndResult executa função com retry global e retorna resultado
func ExecuteWithRetryAndResult(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	return globalRetryHandler.ExecuteWithResult(ctx, fn)
}

// SetGlobalRetryConfig define configuração global de retry
func SetGlobalRetryConfig(config *RetryConfig) {
	globalRetryHandler.UpdateConfig(config)
}

// GetGlobalRetryStats retorna estatísticas globais
func GetGlobalRetryStats() RetryStats {
	return globalRetryHandler.GetStats()
}
