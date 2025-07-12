// Package domainerrors implementa um sistema robusto de tratamento de erros
// seguindo os princípios de Clean Architecture e Design Patterns.
package domainerrors

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

// Pool de objetos para otimização de performance
var (
	domainErrorPool = sync.Pool{
		New: func() interface{} {
			return &DomainError{
				details:  make(map[string]interface{}),
				metadata: make(map[string]interface{}),
				headers:  make(map[string]string),
				tags:     make([]string, 0, 4),
				stack:    make([]StackFrame, 0, 8),
			}
		},
	}

	stackFramePool = sync.Pool{
		New: func() interface{} {
			return &StackFrame{}
		},
	}
)

// DomainError implementa interfaces.DomainErrorInterface com foco em performance.
type DomainError struct {
	// Campos principais
	code      string
	message   string
	errorType types.ErrorType
	severity  types.ErrorSeverity
	category  string

	// Hierarquia de erros
	cause   error
	wrapped []error

	// Metadados e contexto
	details  map[string]interface{}
	metadata map[string]interface{}
	tags     []string

	// HTTP específico
	statusCode int
	headers    map[string]string

	// Stack trace otimizado
	stack []StackFrame

	// Timestamps
	timestamp time.Time

	// Flags de estado
	retryable bool
	temporary bool

	// Mutex para thread safety em operações concorrentes
	mu sync.RWMutex
}

// StackFrame representa um frame no stack trace otimizado.
type StackFrame struct {
	Function string
	File     string
	Line     int
	Message  string
}

// newDomainError cria uma nova instância de DomainError usando object pool.
func newDomainError() *DomainError {
	de := domainErrorPool.Get().(*DomainError)
	de.reset()
	return de
}

// reset restaura o estado inicial do DomainError para reutilização.
func (e *DomainError) reset() {
	e.code = ""
	e.message = ""
	e.errorType = ""
	e.severity = types.SeverityMedium
	e.category = ""
	e.cause = nil
	e.wrapped = e.wrapped[:0]
	e.statusCode = 0
	e.timestamp = time.Now()
	e.retryable = false
	e.temporary = false

	// Limpa maps mas mantém capacidade
	for k := range e.details {
		delete(e.details, k)
	}
	for k := range e.metadata {
		delete(e.metadata, k)
	}
	for k := range e.headers {
		delete(e.headers, k)
	}

	// Limpa slices mas mantém capacidade
	e.tags = e.tags[:0]
	e.stack = e.stack[:0]
}

// release retorna o DomainError para o pool de objetos.
func (e *DomainError) release() {
	e.reset()
	domainErrorPool.Put(e)
}

// Error implementa a interface error padrão do Go.
func (e *DomainError) Error() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var b strings.Builder
	b.Grow(256) // Pre-aloca buffer maior para múltiplos erros

	if e.code != "" {
		b.WriteString("[")
		b.WriteString(e.code)
		b.WriteString("] ")
	}

	b.WriteString(e.message)

	if e.cause != nil {
		b.WriteString(": ")
		b.WriteString(e.cause.Error())
	}

	// Inclui erros encadeados via Chain()
	for i, wrappedErr := range e.wrapped {
		if i == 0 && e.cause == nil {
			b.WriteString(": ")
		} else {
			b.WriteString("; ")
		}
		b.WriteString(wrappedErr.Error())
	}

	return b.String()
}

// Unwrap implementa a interface errors.Wrapper para compatibilidade com errors.Is/As.
func (e *DomainError) Unwrap() error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.cause
}

// Wrap adiciona um erro ao stack de erros.
func (e *DomainError) Wrap(message string, err error) interfaces.DomainErrorInterface {
	if err == nil {
		return e
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Captura stack trace apenas uma vez por nível
	e.captureStackTrace(message, 2)

	// Se o erro anterior era o cause, move para wrapped
	if e.cause != nil {
		e.wrapped = append(e.wrapped, e.cause)
	}

	e.cause = err

	// Herda metadados de DomainError interno
	if de, ok := err.(*DomainError); ok {
		e.inheritFromDomainError(de)
	}

	return e
}

// Chain adiciona um erro ao final da cadeia.
func (e *DomainError) Chain(err error) interfaces.DomainErrorInterface {
	if err == nil {
		return e
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Captura stack trace para erros encadeados
	e.captureStackTrace(fmt.Sprintf("Chained error: %s", err.Error()), 2)

	e.wrapped = append(e.wrapped, err)
	return e
}

// RootCause retorna o erro original da cadeia.
func (e *DomainError) RootCause() error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Proteção contra referências circulares
	visited := make(map[*DomainError]bool)

	current := e.cause
	for current != nil {
		if de, ok := current.(*DomainError); ok {
			// Verifica se já visitamos este erro (referência circular)
			if visited[de] {
				break
			}
			visited[de] = true

			if de.cause != nil {
				current = de.cause
				continue
			}
		}
		break
	}

	return current
}

// String retorna uma representação string detalhada.
func (e *DomainError) String() string {
	return e.Error()
}

// JSON retorna uma representação JSON do erro.
func (e *DomainError) JSON() ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Estrutura otimizada para JSON
	data := struct {
		Code      string                 `json:"code"`
		Message   string                 `json:"message"`
		Type      string                 `json:"type"`
		Severity  string                 `json:"severity"`
		Category  string                 `json:"category,omitempty"`
		Details   map[string]interface{} `json:"details,omitempty"`
		Tags      []string               `json:"tags,omitempty"`
		Timestamp string                 `json:"timestamp"`
		Retryable bool                   `json:"retryable,omitempty"`
		Temporary bool                   `json:"temporary,omitempty"`
	}{
		Code:      e.code,
		Message:   e.message,
		Type:      e.errorType.String(),
		Severity:  e.severity.String(),
		Category:  e.category,
		Details:   e.details,
		Tags:      e.tags,
		Timestamp: e.timestamp.Format(time.RFC3339),
		Retryable: e.retryable,
		Temporary: e.temporary,
	}

	return json.Marshal(data)
}

// FormatStackTrace retorna o stack trace formatado.
func (e *DomainError) FormatStackTrace() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.stack) == 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(len(e.stack) * 100) // Estimativa de tamanho

	b.WriteString("Stack Trace:\n")
	for i, frame := range e.stack {
		b.WriteString(fmt.Sprintf("%d. %s\n   at %s (%s:%d)\n",
			i+1, frame.Message, frame.Function, frame.File, frame.Line))
	}

	return b.String()
}

// DetailedString retorna uma string detalhada incluindo metadados.
func (e *DomainError) DetailedString() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var b strings.Builder
	b.Grow(256)

	b.WriteString(e.Error())
	b.WriteString("\n")

	if e.errorType != "" {
		b.WriteString(fmt.Sprintf("Type: %s\n", e.errorType))
	}

	if e.severity >= 0 {
		b.WriteString(fmt.Sprintf("Severity: %s\n", e.severity))
	}

	if len(e.details) > 0 {
		b.WriteString("Details:\n")
		for k, v := range e.details {
			b.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
		}
	}

	if len(e.tags) > 0 {
		b.WriteString(fmt.Sprintf("Tags: %v\n", e.tags))
	}

	return b.String()
}

// Métodos de acesso (getters) thread-safe
func (e *DomainError) Code() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.code
}

func (e *DomainError) Type() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.errorType.String()
}

func (e *DomainError) Message() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.message
}

func (e *DomainError) Details() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Retorna cópia para evitar modificações concorrentes
	result := make(map[string]interface{}, len(e.details))
	for k, v := range e.details {
		result[k] = v
	}
	return result
}

func (e *DomainError) Metadata() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Retorna cópia para evitar modificações concorrentes
	result := make(map[string]interface{}, len(e.metadata))
	for k, v := range e.metadata {
		result[k] = v
	}
	return result
}

func (e *DomainError) Severity() interfaces.Severity {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return interfaces.Severity(e.severity)
}

func (e *DomainError) Category() interfaces.Category {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return interfaces.Category(e.category)
}

func (e *DomainError) Tags() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Retorna cópia para evitar modificações concorrentes
	result := make([]string, len(e.tags))
	copy(result, e.tags)
	return result
}

func (e *DomainError) StatusCode() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.statusCode > 0 {
		return e.statusCode
	}

	return e.errorType.DefaultStatusCode()
}

func (e *DomainError) Headers() map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Retorna cópia para evitar modificações concorrentes
	result := make(map[string]string, len(e.headers))
	for k, v := range e.headers {
		result[k] = v
	}
	return result
}

func (e *DomainError) ResponseBody() interface{} {
	jsonData, _ := e.JSON()
	return json.RawMessage(jsonData)
}

func (e *DomainError) SetStatusCode(code int) interfaces.DomainErrorInterface {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.statusCode = code
	return e
}

// captureStackTrace captura informações do stack trace de forma otimizada.
func (e *DomainError) captureStackTrace(message string, skip int) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return
	}

	fn := runtime.FuncForPC(pc)
	var funcName string
	if fn != nil {
		funcName = fn.Name()
		// Remove o path do package para economizar espaço
		if idx := strings.LastIndex(funcName, "/"); idx >= 0 {
			funcName = funcName[idx+1:]
		}
	} else {
		funcName = "unknown"
	}

	// Remove path absoluto do arquivo para economizar espaço
	if idx := strings.LastIndex(file, "/"); idx >= 0 {
		file = file[idx+1:]
	}

	frame := StackFrame{
		Function: funcName,
		File:     file,
		Line:     line,
		Message:  message,
	}

	e.stack = append(e.stack, frame)
}

// inheritFromDomainError herda metadados de outro DomainError.
func (e *DomainError) inheritFromDomainError(other *DomainError) {
	other.mu.RLock()
	defer other.mu.RUnlock()

	// Herda detalhes que não existem no erro atual
	for k, v := range other.details {
		if _, exists := e.details[k]; !exists {
			e.details[k] = v
		}
	}

	// Herda metadados que não existem no erro atual
	for k, v := range other.metadata {
		if _, exists := e.metadata[k]; !exists {
			e.metadata[k] = v
		}
	}

	// Herda tags únicas
	for _, tag := range other.tags {
		if !e.hasTag(tag) {
			e.tags = append(e.tags, tag)
		}
	}

	// Herda headers que não existem
	for k, v := range other.headers {
		if _, exists := e.headers[k]; !exists {
			e.headers[k] = v
		}
	}
}

// hasTag verifica se uma tag já existe (método auxiliar interno).
func (e *DomainError) hasTag(tag string) bool {
	for _, t := range e.tags {
		if t == tag {
			return true
		}
	}
	return false
}

// IsRetryable indica se o erro permite retry.
func (e *DomainError) IsRetryable() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.retryable || e.errorType.IsRetryable()
}

// IsTemporary indica se o erro é temporário.
func (e *DomainError) IsTemporary() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.temporary || e.errorType.IsTemporary()
}

// WithContext adiciona informações do contexto ao erro.
func (e *DomainError) WithContext(ctx context.Context) *DomainError {
	if ctx == nil {
		return e
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Extrai informações úteis do contexto
	if requestID := ctx.Value("request_id"); requestID != nil {
		e.metadata["request_id"] = requestID
	}

	if traceID := ctx.Value("trace_id"); traceID != nil {
		e.metadata["trace_id"] = traceID
	}

	if userID := ctx.Value("user_id"); userID != nil {
		e.metadata["user_id"] = userID
	}

	return e
}

// Clone cria uma cópia independente do erro.
func (e *DomainError) Clone() *DomainError {
	e.mu.RLock()
	defer e.mu.RUnlock()

	clone := newDomainError()
	clone.code = e.code
	clone.message = e.message
	clone.errorType = e.errorType
	clone.severity = e.severity
	clone.category = e.category
	clone.cause = e.cause
	clone.statusCode = e.statusCode
	clone.timestamp = e.timestamp
	clone.retryable = e.retryable
	clone.temporary = e.temporary

	// Copia maps
	for k, v := range e.details {
		clone.details[k] = v
	}
	for k, v := range e.metadata {
		clone.metadata[k] = v
	}
	for k, v := range e.headers {
		clone.headers[k] = v
	}

	// Copia slices
	clone.tags = append(clone.tags, e.tags...)
	clone.wrapped = append(clone.wrapped, e.wrapped...)
	clone.stack = append(clone.stack, e.stack...)

	return clone
}
