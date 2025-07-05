package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/message-queue/interfaces"
)

// Retryer define a interface para o sistema de retry
type Retryer interface {
	// Execute executa uma função com retry baseado na política configurada
	Execute(ctx context.Context, fn func() error) error

	// ExecuteWithCallback executa uma função com retry e callback para cada tentativa
	ExecuteWithCallback(ctx context.Context, fn func() error, callback func(attempt int, err error)) error

	// GetPolicy retorna a política de retry configurada
	GetPolicy() *interfaces.RetryPolicy

	// SetPolicy configura uma nova política de retry
	SetPolicy(policy *interfaces.RetryPolicy)
}

// DefaultRetryer implementa o sistema de retry padrão
type DefaultRetryer struct {
	policy *interfaces.RetryPolicy
	rand   *rand.Rand
}

// NewRetryer cria uma nova instância do retryer
func NewRetryer(policy *interfaces.RetryPolicy) Retryer {
	if policy == nil {
		policy = DefaultRetryPolicy()
	}

	return &DefaultRetryer{
		policy: policy,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Execute executa uma função com retry baseado na política configurada
func (r *DefaultRetryer) Execute(ctx context.Context, fn func() error) error {
	return r.ExecuteWithCallback(ctx, fn, nil)
}

// ExecuteWithCallback executa uma função com retry e callback para cada tentativa
func (r *DefaultRetryer) ExecuteWithCallback(ctx context.Context, fn func() error, callback func(attempt int, err error)) error {
	var lastErr error

	for attempt := 1; attempt <= r.policy.MaxAttempts; attempt++ {
		// Verifica se o contexto foi cancelado
		if ctx.Err() != nil {
			return domainerrors.New(
				"RETRY_CANCELLED",
				"retry cancelled due to context cancellation",
			).WithType(domainerrors.ErrorTypeRepository).Wrap("context cancelled", ctx.Err())
		}

		// Executa a função
		err := fn()
		if err == nil {
			return nil // Sucesso!
		}

		lastErr = err

		// Chama callback se fornecido
		if callback != nil {
			callback(attempt, err)
		}

		// Se é a última tentativa, não precisa fazer delay
		if attempt == r.policy.MaxAttempts {
			break
		}

		// Calcula o delay para a próxima tentativa
		delay := r.calculateDelay(attempt)

		// Espera o delay ou cancela se o contexto for cancelado
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return domainerrors.New(
				"RETRY_CANCELLED",
				"retry cancelled due to context cancellation",
			).WithType(domainerrors.ErrorTypeRepository).Wrap("context cancelled", ctx.Err())
		case <-timer.C:
			// Continua para a próxima tentativa
		}
	}

	// Todas as tentativas falharam
	return domainerrors.New(
		"RETRY_EXHAUSTED",
		fmt.Sprintf("retry exhausted after %d attempts", r.policy.MaxAttempts),
	).WithType(domainerrors.ErrorTypeRepository).Wrap("retry failed", lastErr)
}

// GetPolicy retorna a política de retry configurada
func (r *DefaultRetryer) GetPolicy() *interfaces.RetryPolicy {
	return r.policy
}

// SetPolicy configura uma nova política de retry
func (r *DefaultRetryer) SetPolicy(policy *interfaces.RetryPolicy) {
	if policy != nil {
		r.policy = policy
	}
}

// calculateDelay calcula o delay para uma tentativa específica
func (r *DefaultRetryer) calculateDelay(attempt int) time.Duration {
	// Calcula o delay base usando backoff exponencial
	delay := float64(r.policy.InitialDelay) * math.Pow(r.policy.BackoffMultiplier, float64(attempt-1))

	// Aplica o limite máximo
	if delay > float64(r.policy.MaxDelay) {
		delay = float64(r.policy.MaxDelay)
	}

	// Aplica jitter se habilitado
	if r.policy.Jitter {
		// Aplica jitter de ±25%
		jitter := 0.25
		minDelay := delay * (1.0 - jitter)
		maxDelay := delay * (1.0 + jitter)
		delay = minDelay + r.rand.Float64()*(maxDelay-minDelay)
	}

	return time.Duration(delay)
}

// DefaultRetryPolicy retorna uma política de retry padrão
func DefaultRetryPolicy() *interfaces.RetryPolicy {
	return &interfaces.RetryPolicy{
		MaxAttempts:       3,
		InitialDelay:      1 * time.Second,
		BackoffMultiplier: 2.0,
		MaxDelay:          30 * time.Second,
		Jitter:            true,
	}
}

// ExponentialBackoffPolicy retorna uma política de backoff exponencial
func ExponentialBackoffPolicy(maxAttempts int, initialDelay, maxDelay time.Duration) *interfaces.RetryPolicy {
	return &interfaces.RetryPolicy{
		MaxAttempts:       maxAttempts,
		InitialDelay:      initialDelay,
		BackoffMultiplier: 2.0,
		MaxDelay:          maxDelay,
		Jitter:            true,
	}
}

// LinearBackoffPolicy retorna uma política de backoff linear
func LinearBackoffPolicy(maxAttempts int, delay time.Duration) *interfaces.RetryPolicy {
	return &interfaces.RetryPolicy{
		MaxAttempts:       maxAttempts,
		InitialDelay:      delay,
		BackoffMultiplier: 1.0, // Linear
		MaxDelay:          delay * time.Duration(maxAttempts),
		Jitter:            false,
	}
}

// NoRetryPolicy retorna uma política sem retry
func NoRetryPolicy() *interfaces.RetryPolicy {
	return &interfaces.RetryPolicy{
		MaxAttempts:       1,
		InitialDelay:      0,
		BackoffMultiplier: 1.0,
		MaxDelay:          0,
		Jitter:            false,
	}
}

// IsRetryableError verifica se um erro deve ser retriado
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Verifica se é um TimeoutError específico
	if _, ok := err.(*domainerrors.TimeoutError); ok {
		return true // TimeoutErrors são sempre retriáveis
	}

	// Verifica se é um InfrastructureError específico
	if _, ok := err.(*domainerrors.InfrastructureError); ok {
		return true // InfrastructureErrors são sempre retriáveis
	}

	// Verifica se é um erro de domínio específico
	if domainErr, ok := err.(*domainerrors.DomainError); ok {
		switch domainErr.Type {
		case domainerrors.ErrorTypeRepository:
			// Erros de repositório geralmente são retriáveis
			return true
		case domainerrors.ErrorTypeValidation:
			// Erros de validação não são retriáveis
			return false
		case domainerrors.ErrorTypeBusinessRule:
			// Erros de regra de negócio não são retriáveis
			return false
		case domainerrors.ErrorTypeTimeout:
			// Erros de timeout são retriáveis
			return true
		case domainerrors.ErrorTypeInfrastructure:
			// Erros de infraestrutura são retriáveis
			return true
		default:
			// Por padrão, considera como retriável
			return true
		}
	}

	// Para outros tipos de erro, verifica palavras-chave
	errMsg := err.Error()
	retriableKeywords := []string{
		"connection",
		"timeout",
		"network",
		"temporary",
		"unavailable",
		"circuit breaker",
		"rate limit",
		"throttle",
	}

	for _, keyword := range retriableKeywords {
		if contains(errMsg, keyword) {
			return true
		}
	}

	return false
}

// contains verifica se uma string contém uma substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && anyIndex(s, substr) >= 0)
}

// anyIndex é uma implementação simples para encontrar substring
func anyIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
