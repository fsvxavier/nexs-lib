package valkey

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/cache/valkey/interfaces"
)

// ExponentialBackoffRetryPolicy implementa retry com backoff exponencial.
type ExponentialBackoffRetryPolicy struct {
	maxRetries    int
	minBackoff    time.Duration
	maxBackoff    time.Duration
	multiplier    float64
	jitterEnabled bool
}

// NewExponentialBackoffRetryPolicy cria uma nova política de retry com backoff exponencial.
func NewExponentialBackoffRetryPolicy(maxRetries int, minBackoff, maxBackoff time.Duration) interfaces.IRetryPolicy {
	return &ExponentialBackoffRetryPolicy{
		maxRetries:    maxRetries,
		minBackoff:    minBackoff,
		maxBackoff:    maxBackoff,
		multiplier:    2.0,
		jitterEnabled: true,
	}
}

// ShouldRetry determina se deve tentar novamente baseado no número de tentativas e erro.
func (r *ExponentialBackoffRetryPolicy) ShouldRetry(attempt int, err error) bool {
	if attempt >= r.maxRetries {
		return false
	}

	// Não fazer retry para certos tipos de erro
	if isNonRetryableError(err) {
		return false
	}

	return true
}

// NextDelay calcula o próximo delay baseado no número de tentativas.
func (r *ExponentialBackoffRetryPolicy) NextDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return r.minBackoff
	}

	// Calcular backoff exponencial
	delay := time.Duration(float64(r.minBackoff) * math.Pow(r.multiplier, float64(attempt-1)))

	// Aplicar limite máximo
	if delay > r.maxBackoff {
		delay = r.maxBackoff
	}

	// Aplicar jitter se habilitado
	if r.jitterEnabled {
		jitter := time.Duration(rand.Float64() * float64(delay) * 0.1) // 10% jitter
		delay += jitter
	}

	return delay
}

// isNonRetryableError verifica se um erro não deve ser reprocessado.
func isNonRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Erros que não devem ser reprocessados
	nonRetryableErrors := []string{
		"context canceled",
		"context deadline exceeded",
		"authentication failed",
		"permission denied",
		"invalid command",
		"wrong number of arguments",
		"unknown command",
	}

	for _, nonRetryable := range nonRetryableErrors {
		if contains(errStr, nonRetryable) {
			return true
		}
	}

	return false
}

// contains verifica se uma string contém uma substring.
func contains(str, substr string) bool {
	return len(str) >= len(substr) && str[len(str)-len(substr):] == substr ||
		len(str) > len(substr) && str[:len(substr)] == substr ||
		len(str) > len(substr) && str[len(str)/2-len(substr)/2:len(str)/2+len(substr)/2] == substr
}

// CircuitBreakerState representa o estado do circuit breaker.
type CircuitBreakerState string

const (
	StateClosed   CircuitBreakerState = "CLOSED"
	StateOpen     CircuitBreakerState = "OPEN"
	StateHalfOpen CircuitBreakerState = "HALF_OPEN"
)

// CircuitBreaker implementa o padrão circuit breaker.
type CircuitBreaker struct {
	mu           sync.RWMutex
	state        CircuitBreakerState
	failureCount int
	threshold    int
	timeout      time.Duration
	maxRequests  int
	requests     int
	lastFailure  time.Time
}

// NewCircuitBreaker cria um novo circuit breaker.
func NewCircuitBreaker(threshold int, timeout time.Duration, maxRequests int) interfaces.ICircuitBreaker {
	return &CircuitBreaker{
		state:       StateClosed,
		threshold:   threshold,
		timeout:     timeout,
		maxRequests: maxRequests,
	}
}

// Execute executa uma função através do circuit breaker.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	if !cb.allowRequest() {
		return nil, errors.New("circuit breaker is open")
	}

	result, err := fn()

	cb.recordResult(err == nil)

	return result, err
}

// State retorna o estado atual do circuit breaker.
func (cb *CircuitBreaker) State() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return string(cb.state)
}

// Reset reseta o circuit breaker para o estado fechado.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.requests = 0
	cb.lastFailure = time.Time{}
}

// allowRequest verifica se uma requisição pode ser feita.
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Verificar se é hora de tentar half-open
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.state = StateHalfOpen
			cb.requests = 0
			return true
		}
		return false
	case StateHalfOpen:
		return cb.requests < cb.maxRequests
	default:
		return false
	}
}

// recordResult registra o resultado de uma operação.
func (cb *CircuitBreaker) recordResult(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if success {
		cb.handleSuccess()
	} else {
		cb.handleFailure()
	}
}

// handleSuccess trata um resultado de sucesso.
func (cb *CircuitBreaker) handleSuccess() {
	switch cb.state {
	case StateClosed:
		cb.failureCount = 0
	case StateHalfOpen:
		cb.requests++
		if cb.requests >= cb.maxRequests {
			cb.state = StateClosed
			cb.failureCount = 0
			cb.requests = 0
		}
	}
}

// handleFailure trata um resultado de falha.
func (cb *CircuitBreaker) handleFailure() {
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		cb.failureCount++
		if cb.failureCount >= cb.threshold {
			cb.state = StateOpen
		}
	case StateHalfOpen:
		cb.state = StateOpen
		cb.requests = 0
	}
}

// LinearRetryPolicy implementa retry com delay linear.
type LinearRetryPolicy struct {
	maxRetries int
	delay      time.Duration
}

// NewLinearRetryPolicy cria uma nova política de retry linear.
func NewLinearRetryPolicy(maxRetries int, delay time.Duration) interfaces.IRetryPolicy {
	return &LinearRetryPolicy{
		maxRetries: maxRetries,
		delay:      delay,
	}
}

// ShouldRetry implementa interfaces.IRetryPolicy.
func (r *LinearRetryPolicy) ShouldRetry(attempt int, err error) bool {
	if attempt >= r.maxRetries {
		return false
	}

	return !isNonRetryableError(err)
}

// NextDelay implementa interfaces.IRetryPolicy.
func (r *LinearRetryPolicy) NextDelay(attempt int) time.Duration {
	return r.delay
}

// NoRetryPolicy implementa uma política sem retry.
type NoRetryPolicy struct{}

// NewNoRetryPolicy cria uma política sem retry.
func NewNoRetryPolicy() interfaces.IRetryPolicy {
	return &NoRetryPolicy{}
}

// ShouldRetry implementa interfaces.IRetryPolicy.
func (r *NoRetryPolicy) ShouldRetry(attempt int, err error) bool {
	return false
}

// NextDelay implementa interfaces.IRetryPolicy.
func (r *NoRetryPolicy) NextDelay(attempt int) time.Duration {
	return 0
}

// FixedDelayRetryPolicy implementa retry com delay fixo.
type FixedDelayRetryPolicy struct {
	maxRetries int
	delay      time.Duration
}

// NewFixedDelayRetryPolicy cria uma nova política de retry com delay fixo.
func NewFixedDelayRetryPolicy(maxRetries int, delay time.Duration) interfaces.IRetryPolicy {
	return &FixedDelayRetryPolicy{
		maxRetries: maxRetries,
		delay:      delay,
	}
}

// ShouldRetry implementa interfaces.IRetryPolicy.
func (r *FixedDelayRetryPolicy) ShouldRetry(attempt int, err error) bool {
	if attempt >= r.maxRetries {
		return false
	}

	return !isNonRetryableError(err)
}

// NextDelay implementa interfaces.IRetryPolicy.
func (r *FixedDelayRetryPolicy) NextDelay(attempt int) time.Duration {
	return r.delay
}
