package advanced

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors/hooks"
	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// ErrorAggregator implementa um sistema para agregar múltiplos erros
type ErrorAggregator struct {
	errors    []interfaces.DomainErrorInterface
	threshold int
	window    time.Duration
	mu        sync.RWMutex
	timer     *time.Timer
	flushChan chan struct{}
	closed    bool
}

// NewErrorAggregator cria um novo agregador de erros
func NewErrorAggregator(threshold int, window time.Duration) *ErrorAggregator {
	ea := &ErrorAggregator{
		errors:    make([]interfaces.DomainErrorInterface, 0),
		threshold: threshold,
		window:    window,
		flushChan: make(chan struct{}, 1),
	}

	// Inicia goroutine para flush automático
	go ea.autoFlush()

	return ea
}

// Add adiciona um erro ao agregador
func (ea *ErrorAggregator) Add(err interfaces.DomainErrorInterface) error {
	ea.mu.Lock()
	defer ea.mu.Unlock()

	ea.errors = append(ea.errors, err)

	// Reset timer para window
	if ea.timer != nil {
		ea.timer.Reset(ea.window)
	} else {
		ea.timer = time.AfterFunc(ea.window, func() {
			ea.mu.RLock()
			closed := ea.closed
			ea.mu.RUnlock()

			if !closed {
				select {
				case ea.flushChan <- struct{}{}:
				default:
				}
			}
		})
	}

	// Flush se atingiu threshold
	if len(ea.errors) >= ea.threshold {
		return ea.flush()
	}

	return nil
}

// flush processa todos os erros agregados
func (ea *ErrorAggregator) flush() error {
	if len(ea.errors) == 0 {
		return nil
	}

	aggregatedErr := ea.createAggregatedError()
	ea.errors = ea.errors[:0] // Clear slice

	// Executar hooks para erro agregado
	return hooks.ExecuteGlobalErrorHooks(context.Background(), aggregatedErr)
}

// Flush força o flush dos erros pendentes
func (ea *ErrorAggregator) Flush() error {
	ea.mu.Lock()
	defer ea.mu.Unlock()
	return ea.flush()
}

// createAggregatedError cria um erro agregado
func (ea *ErrorAggregator) createAggregatedError() interfaces.DomainErrorInterface {
	errorCodes := make([]string, len(ea.errors))
	errorTypes := make(map[string]int)

	for i, err := range ea.errors {
		errorCodes[i] = err.Code()
		errorTypes[string(err.Type())]++
	}

	// Determina tipo mais comum
	var dominantType interfaces.ErrorType = interfaces.ServerError
	maxCount := 0
	for errType, count := range errorTypes {
		if count > maxCount {
			maxCount = count
			dominantType = interfaces.ErrorType(errType)
		}
	}

	// Criar erro agregado usando factory pattern
	// Assumindo que existe uma função para criar erros
	return &AggregatedError{
		errorType: dominantType,
		code:      "AGGREGATED_ERRORS",
		message:   fmt.Sprintf("Aggregated %d errors", len(ea.errors)),
		metadata: map[string]interface{}{
			"error_count":   len(ea.errors),
			"error_codes":   errorCodes,
			"error_types":   errorTypes,
			"dominant_type": string(dominantType),
			"aggregated_at": time.Now().Format(time.RFC3339),
		},
		timestamp: time.Now(),
		errors:    ea.errors,
	}
}

// extractCodes extrai códigos dos erros
func (ea *ErrorAggregator) extractCodes() []string {
	codes := make([]string, len(ea.errors))
	for i, err := range ea.errors {
		codes[i] = err.Code()
	}
	return codes
}

// autoFlush executa flush automático baseado no window
func (ea *ErrorAggregator) autoFlush() {
	for range ea.flushChan {
		ea.mu.Lock()
		ea.flush()
		ea.mu.Unlock()
	}
}

// Count retorna o número de erros agregados
func (ea *ErrorAggregator) Count() int {
	ea.mu.RLock()
	defer ea.mu.RUnlock()
	return len(ea.errors)
}

// HasErrors verifica se há erros pendentes
func (ea *ErrorAggregator) HasErrors() bool {
	ea.mu.RLock()
	defer ea.mu.RUnlock()
	return len(ea.errors) > 0
}

// Close para o agregador e faz flush final
func (ea *ErrorAggregator) Close() error {
	ea.mu.Lock()
	defer ea.mu.Unlock()

	if ea.closed {
		return nil
	}

	ea.closed = true

	if ea.timer != nil {
		ea.timer.Stop()
	}
	close(ea.flushChan)

	return ea.flush()
} // AggregatedError implementa DomainErrorInterface para erros agregados
type AggregatedError struct {
	errorType interfaces.ErrorType
	code      string
	message   string
	metadata  map[string]interface{}
	timestamp time.Time
	errors    []interfaces.DomainErrorInterface
}

func (ae *AggregatedError) Error() string {
	return ae.message
}

func (ae *AggregatedError) Unwrap() error {
	if len(ae.errors) > 0 {
		return ae.errors[0]
	}
	return nil
}

func (ae *AggregatedError) Type() interfaces.ErrorType {
	return ae.errorType
}

func (ae *AggregatedError) Metadata() map[string]interface{} {
	return ae.metadata
}

func (ae *AggregatedError) HTTPStatus() int {
	// Retorna o status HTTP baseado no tipo dominante
	switch ae.errorType {
	case interfaces.ValidationError, interfaces.BadRequestError:
		return 400
	case interfaces.NotFoundError:
		return 404
	case interfaces.ConflictError:
		return 409
	case interfaces.UnprocessableEntityError:
		return 422
	case interfaces.ServiceUnavailableError:
		return 503
	default:
		return 500
	}
}

func (ae *AggregatedError) StackTrace() string {
	return "Aggregated error - see individual errors for stack traces"
}

func (ae *AggregatedError) WithContext(ctx context.Context) interfaces.DomainErrorInterface {
	ae.metadata["context"] = ctx
	return ae
}

func (ae *AggregatedError) Wrap(err error) interfaces.DomainErrorInterface {
	ae.metadata["wrapped_error"] = err.Error()
	return ae
}

func (ae *AggregatedError) WithMetadata(key string, value interface{}) interfaces.DomainErrorInterface {
	ae.metadata[key] = value
	return ae
}

func (ae *AggregatedError) Code() string {
	return ae.code
}

func (ae *AggregatedError) Timestamp() time.Time {
	return ae.timestamp
}

func (ae *AggregatedError) ToJSON() ([]byte, error) {
	// Implementação básica - pode ser expandida
	return nil, fmt.Errorf("ToJSON not implemented for AggregatedError")
}

// GetAggregatedErrors retorna os erros individuais
func (ae *AggregatedError) GetAggregatedErrors() []interfaces.DomainErrorInterface {
	return ae.errors
}
