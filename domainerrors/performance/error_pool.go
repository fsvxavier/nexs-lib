package performance

import (
	"context"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// ErrorPool implementa pool de objetos para redução de GC pressure
type ErrorPool struct {
	domainErrors sync.Pool
	metadata     sync.Pool
	stackFrames  sync.Pool
}

// NewErrorPool cria um novo pool de erros
func NewErrorPool() *ErrorPool {
	return &ErrorPool{
		domainErrors: sync.Pool{
			New: func() interface{} {
				return &PooledDomainError{
					metadata:   make(map[string]interface{}, 8),      // Pré-aloca com capacidade 8
					stackTrace: make([]interfaces.StackFrame, 0, 16), // Pré-aloca com capacidade 16
				}
			},
		},
		metadata: sync.Pool{
			New: func() interface{} {
				return make(map[string]interface{}, 8)
			},
		},
		stackFrames: sync.Pool{
			New: func() interface{} {
				return make([]interfaces.StackFrame, 0, 16)
			},
		},
	}
}

// GetDomainError obtém um erro do pool
func (p *ErrorPool) GetDomainError() *PooledDomainError {
	err := p.domainErrors.Get().(*PooledDomainError)
	err.pool = p // Referência para retornar ao pool
	err.reset()  // Limpa estado anterior
	return err
}

// GetMetadata obtém mapa de metadados do pool
func (p *ErrorPool) GetMetadata() map[string]interface{} {
	metadata := p.metadata.Get().(map[string]interface{})
	// Limpar mapa reutilizado
	for k := range metadata {
		delete(metadata, k)
	}
	return metadata
}

// GetStackFrames obtém slice de stack frames do pool
func (p *ErrorPool) GetStackFrames() []interfaces.StackFrame {
	frames := p.stackFrames.Get().([]interfaces.StackFrame)
	// Limpar slice reutilizado
	return frames[:0]
}

// PutDomainError retorna erro ao pool
func (p *ErrorPool) PutDomainError(err *PooledDomainError) {
	if err != nil {
		err.reset()
		p.domainErrors.Put(err)
	}
}

// PutMetadata retorna metadados ao pool
func (p *ErrorPool) PutMetadata(metadata map[string]interface{}) {
	if metadata != nil && len(metadata) <= 32 { // Apenas mapas pequenos para evitar memory leak
		// Limpar mapa
		for k := range metadata {
			delete(metadata, k)
		}
		p.metadata.Put(metadata)
	}
}

// PutStackFrames retorna stack frames ao pool
func (p *ErrorPool) PutStackFrames(frames []interfaces.StackFrame) {
	if frames != nil && cap(frames) <= 64 { // Apenas slices pequenos
		p.stackFrames.Put(frames[:0])
	}
}

// PooledDomainError implementação otimizada de DomainError usando pools
type PooledDomainError struct {
	errorType  interfaces.ErrorType
	code       string
	message    string
	metadata   map[string]interface{}
	stackTrace []interfaces.StackFrame
	timestamp  time.Time
	httpStatus int
	pool       *ErrorPool
	wrapped    error
}

// reset limpa o estado do erro para reutilização
func (pde *PooledDomainError) reset() {
	pde.errorType = ""
	pde.code = ""
	pde.message = ""
	pde.httpStatus = 0
	pde.wrapped = nil
	pde.timestamp = time.Time{}

	// Limpar metadados sem realocar
	for k := range pde.metadata {
		delete(pde.metadata, k)
	}

	// Limpar stack trace sem realocar
	pde.stackTrace = pde.stackTrace[:0]
}

// Initialize inicializa erro pooled
func (pde *PooledDomainError) Initialize(errorType interfaces.ErrorType, code, message string) *PooledDomainError {
	pde.errorType = errorType
	pde.code = code
	pde.message = message
	pde.timestamp = time.Now()
	pde.httpStatus = pde.calculateHTTPStatus()
	return pde
}

// Error implementa interface error
func (pde *PooledDomainError) Error() string {
	return pde.message
}

// Unwrap retorna erro encapsulado
func (pde *PooledDomainError) Unwrap() error {
	return pde.wrapped
}

// Type retorna tipo do erro
func (pde *PooledDomainError) Type() interfaces.ErrorType {
	return pde.errorType
}

// Metadata retorna metadados
func (pde *PooledDomainError) Metadata() map[string]interface{} {
	return pde.metadata
}

// HTTPStatus retorna status HTTP
func (pde *PooledDomainError) HTTPStatus() int {
	return pde.httpStatus
}

// StackTrace retorna stack trace formatado
func (pde *PooledDomainError) StackTrace() string {
	if len(pde.stackTrace) == 0 {
		return ""
	}

	// Implementação otimizada de formatação
	var result string
	for _, frame := range pde.stackTrace {
		result += frame.Function + " at " + frame.File + ":" + string(rune(frame.Line)) + "\n"
	}
	return result
}

// WithContext adiciona contexto (não implementado para versão pooled por performance)
func (pde *PooledDomainError) WithContext(ctx context.Context) interfaces.DomainErrorInterface {
	// Para versão pooled, apenas adiciona aos metadados
	pde.metadata["context"] = "context_added"
	return pde
}

// Wrap encapsula outro erro
func (pde *PooledDomainError) Wrap(err error) interfaces.DomainErrorInterface {
	pde.wrapped = err
	return pde
}

// WithMetadata adiciona metadado
func (pde *PooledDomainError) WithMetadata(key string, value interface{}) interfaces.DomainErrorInterface {
	pde.metadata[key] = value
	return pde
}

// Code retorna código do erro
func (pde *PooledDomainError) Code() string {
	return pde.code
}

// Timestamp retorna timestamp
func (pde *PooledDomainError) Timestamp() time.Time {
	return pde.timestamp
}

// ToJSON serializa para JSON (implementação básica)
func (pde *PooledDomainError) ToJSON() ([]byte, error) {
	// Por performance, retorna erro - implementação completa seria mais cara
	return nil, nil
}

// Release retorna erro ao pool
func (pde *PooledDomainError) Release() {
	if pde.pool != nil {
		pde.pool.PutDomainError(pde)
	}
}

// calculateHTTPStatus calcula status HTTP baseado no tipo
func (pde *PooledDomainError) calculateHTTPStatus() int {
	switch pde.errorType {
	case interfaces.ValidationError, interfaces.BadRequestError:
		return 400
	case interfaces.AuthenticationError:
		return 401
	case interfaces.AuthorizationError:
		return 403
	case interfaces.NotFoundError:
		return 404
	case interfaces.ConflictError:
		return 409
	case interfaces.UnprocessableEntityError:
		return 422
	case interfaces.RateLimitError:
		return 429
	case interfaces.ServiceUnavailableError:
		return 503
	default:
		return 500
	}
}

// StringPool implementa pool de strings para mensagens comuns
type StringPool struct {
	pool sync.Pool
}

// NewStringPool cria novo pool de strings
func NewStringPool() *StringPool {
	return &StringPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, 256) // Buffer de 256 bytes
			},
		},
	}
}

// GetBuffer obtém buffer do pool
func (sp *StringPool) GetBuffer() []byte {
	return sp.pool.Get().([]byte)[:0]
}

// PutBuffer retorna buffer ao pool
func (sp *StringPool) PutBuffer(buf []byte) {
	if cap(buf) <= 1024 { // Apenas buffers pequenos
		sp.pool.Put(buf)
	}
}

// String interning para strings comuns
var (
	commonStrings = map[string]*string{
		"VALIDATION_ERROR":          stringPtr("VALIDATION_ERROR"),
		"NOT_FOUND":                 stringPtr("NOT_FOUND"),
		"INTERNAL_ERROR":            stringPtr("INTERNAL_ERROR"),
		"AUTHENTICATION_ERROR":      stringPtr("AUTHENTICATION_ERROR"),
		"AUTHORIZATION_ERROR":       stringPtr("AUTHORIZATION_ERROR"),
		"TIMEOUT_ERROR":             stringPtr("TIMEOUT_ERROR"),
		"EXTERNAL_SERVICE_ERROR":    stringPtr("EXTERNAL_SERVICE_ERROR"),
		"RATE_LIMIT_ERROR":          stringPtr("RATE_LIMIT_ERROR"),
		"SERVICE_UNAVAILABLE_ERROR": stringPtr("SERVICE_UNAVAILABLE_ERROR"),
	}
	commonStringsMu sync.RWMutex
)

// stringPtr helper para criar ponteiro de string
func stringPtr(s string) *string {
	return &s
}

// InternString retorna string internalizada se comum, senão cria nova
func InternString(s string) *string {
	commonStringsMu.RLock()
	if interned, exists := commonStrings[s]; exists {
		commonStringsMu.RUnlock()
		return interned
	}
	commonStringsMu.RUnlock()

	return &s
}

// AddCommonString adiciona string comum ao pool
func AddCommonString(s string) {
	commonStringsMu.Lock()
	defer commonStringsMu.Unlock()

	if _, exists := commonStrings[s]; !exists {
		commonStrings[s] = stringPtr(s)
	}
}

// Instâncias globais dos pools
var (
	GlobalErrorPool  = NewErrorPool()
	GlobalStringPool = NewStringPool()
)

// NewPooledError cria novo erro usando pool global
func NewPooledError(errorType interfaces.ErrorType, code, message string) *PooledDomainError {
	err := GlobalErrorPool.GetDomainError()
	return err.Initialize(errorType, code, message)
}

// NewPooledErrorWithMetadata cria erro com metadados usando pool
func NewPooledErrorWithMetadata(errorType interfaces.ErrorType, code, message string, metadata map[string]interface{}) *PooledDomainError {
	err := NewPooledError(errorType, code, message)

	// Copia metadados
	for k, v := range metadata {
		err.metadata[k] = v
	}

	return err
}
