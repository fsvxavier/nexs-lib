// Package registry implementa um sistema de registro de códigos de erro
// seguindo o padrão Registry e Singleton para gerenciamento centralizado.
package registry

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

// ErrorCodeRegistry implementa interfaces.ErrorRegistry com thread safety.
type ErrorCodeRegistry struct {
	codes   map[string]interfaces.ErrorCodeInfo
	mutex   sync.RWMutex
	factory interfaces.ErrorFactory
}

// NewErrorCodeRegistry cria um novo registro de códigos de erro.
func NewErrorCodeRegistry() interfaces.ErrorRegistry {
	registry := &ErrorCodeRegistry{
		codes:   make(map[string]interfaces.ErrorCodeInfo),
		factory: nil, // Factory será injetada quando necessário
	}

	// Registra códigos comuns automaticamente
	registry.registerCommonCodes()

	return registry
}

// NewErrorCodeRegistryWithFactory cria um registro com factory customizada.
func NewErrorCodeRegistryWithFactory(factory interfaces.ErrorFactory) interfaces.ErrorRegistry {
	registry := &ErrorCodeRegistry{
		codes:   make(map[string]interfaces.ErrorCodeInfo),
		factory: factory,
	}

	// Registra códigos comuns automaticamente
	registry.registerCommonCodes()

	return registry
}

// Register adiciona um novo código de erro ao registro.
func (r *ErrorCodeRegistry) Register(info interfaces.ErrorCodeInfo) error {
	if info.Code == "" {
		return fmt.Errorf("error code cannot be empty")
	}

	if info.Message == "" {
		return fmt.Errorf("error message cannot be empty for code: %s", info.Code)
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Verifica se o código já existe
	if _, exists := r.codes[info.Code]; exists {
		return fmt.Errorf("error code already exists: %s", info.Code)
	}

	// Validações adicionais
	if err := r.validateErrorCodeInfo(info); err != nil {
		return fmt.Errorf("invalid error code info for %s: %w", info.Code, err)
	}

	r.codes[info.Code] = info
	return nil
}

// RegisterMultiple adiciona múltiplos códigos de erro de uma vez.
func (r *ErrorCodeRegistry) RegisterMultiple(infos []interfaces.ErrorCodeInfo) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Valida todos os códigos primeiro
	for _, info := range infos {
		if err := r.validateErrorCodeInfo(info); err != nil {
			return fmt.Errorf("validation failed for code %s: %w", info.Code, err)
		}

		if _, exists := r.codes[info.Code]; exists {
			return fmt.Errorf("error code already exists: %s", info.Code)
		}
	}

	// Se todas as validações passaram, registra todos
	for _, info := range infos {
		r.codes[info.Code] = info
	}

	return nil
}

// Get obtém informações de um código de erro.
func (r *ErrorCodeRegistry) Get(code string) (interfaces.ErrorCodeInfo, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	info, exists := r.codes[code]
	return info, exists
}

// Exists verifica se um código existe.
func (r *ErrorCodeRegistry) Exists(code string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.codes[code]
	return exists
}

// List retorna todos os códigos registrados.
func (r *ErrorCodeRegistry) List() []interfaces.ErrorCodeInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	infos := make([]interfaces.ErrorCodeInfo, 0, len(r.codes))
	for _, info := range r.codes {
		infos = append(infos, info)
	}

	// Ordena por código para resultado consistente
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Code < infos[j].Code
	})

	return infos
}

// ListByType retorna códigos filtrados por tipo.
func (r *ErrorCodeRegistry) ListByType(errorType string) []interfaces.ErrorCodeInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	infos := make([]interfaces.ErrorCodeInfo, 0)
	for _, info := range r.codes {
		if info.Type == errorType {
			infos = append(infos, info)
		}
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Code < infos[j].Code
	})

	return infos
}

// ListBySeverity retorna códigos filtrados por severidade.
func (r *ErrorCodeRegistry) ListBySeverity(severity interfaces.Severity) []interfaces.ErrorCodeInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	infos := make([]interfaces.ErrorCodeInfo, 0)
	for _, info := range r.codes {
		if info.Severity == severity {
			infos = append(infos, info)
		}
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Code < infos[j].Code
	})

	return infos
}

// CreateError cria um erro baseado em um código registrado.
func (r *ErrorCodeRegistry) CreateError(code string, args ...interface{}) (interfaces.DomainErrorInterface, error) {
	info, exists := r.Get(code)
	if !exists {
		return nil, fmt.Errorf("error code not found: %s", code)
	}

	message := info.Message
	if len(args) > 0 {
		message = fmt.Sprintf(info.Message, args...)
	}

	// Se factory não estiver disponível, retorna erro simples
	if r.factory == nil {
		return &basicError{
			code:       code,
			message:    message,
			errorType:  info.Type,
			statusCode: info.StatusCode,
			severity:   types.ErrorSeverity(info.Severity),
			tags:       info.Tags,
			timestamp:  time.Now(),
		}, nil
	}

	builder := r.factory.Builder()
	err := builder.
		WithCode(code).
		WithMessage(message).
		WithType(info.Type).
		WithStatusCode(info.StatusCode).
		WithSeverity(info.Severity).
		WithTags(info.Tags).
		Build()

	return err, nil
}

// Update atualiza um código de erro existente.
func (r *ErrorCodeRegistry) Update(info interfaces.ErrorCodeInfo) error {
	if info.Code == "" {
		return fmt.Errorf("error code cannot be empty")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.codes[info.Code]; !exists {
		return fmt.Errorf("error code not found: %s", info.Code)
	}

	if err := r.validateErrorCodeInfo(info); err != nil {
		return fmt.Errorf("invalid error code info for %s: %w", info.Code, err)
	}

	r.codes[info.Code] = info
	return nil
}

// Remove remove um código de erro do registro.
func (r *ErrorCodeRegistry) Remove(code string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.codes[code]; !exists {
		return fmt.Errorf("error code not found: %s", code)
	}

	delete(r.codes, code)
	return nil
}

// Clear remove todos os códigos de erro.
func (r *ErrorCodeRegistry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.codes = make(map[string]interfaces.ErrorCodeInfo)
}

// Count retorna o número de códigos registrados.
func (r *ErrorCodeRegistry) Count() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.codes)
}

// Search busca códigos por pattern na mensagem ou descrição.
func (r *ErrorCodeRegistry) Search(pattern string) []interfaces.ErrorCodeInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	pattern = strings.ToLower(pattern)
	infos := make([]interfaces.ErrorCodeInfo, 0)

	for _, info := range r.codes {
		if strings.Contains(strings.ToLower(info.Message), pattern) ||
			strings.Contains(strings.ToLower(info.Description), pattern) ||
			strings.Contains(strings.ToLower(info.Code), pattern) {
			infos = append(infos, info)
		}
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Code < infos[j].Code
	})

	return infos
}

// Export exporta todos os códigos de erro em formato estruturado.
func (r *ErrorCodeRegistry) Export() map[string]interfaces.ErrorCodeInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	export := make(map[string]interfaces.ErrorCodeInfo)
	for code, info := range r.codes {
		export[code] = info
	}

	return export
}

// Import importa códigos de erro de um mapa.
func (r *ErrorCodeRegistry) Import(codes map[string]interfaces.ErrorCodeInfo, overwrite bool) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Valida todos os códigos primeiro
	for code, info := range codes {
		if code != info.Code {
			return fmt.Errorf("code mismatch: key '%s' != info.Code '%s'", code, info.Code)
		}

		if err := r.validateErrorCodeInfo(info); err != nil {
			return fmt.Errorf("validation failed for code %s: %w", code, err)
		}

		if !overwrite {
			if _, exists := r.codes[code]; exists {
				return fmt.Errorf("error code already exists: %s", code)
			}
		}
	}

	// Se todas as validações passaram, importa todos
	for code, info := range codes {
		r.codes[code] = info
	}

	return nil
}

// validateErrorCodeInfo valida a estrutura de um ErrorCodeInfo.
func (r *ErrorCodeRegistry) validateErrorCodeInfo(info interfaces.ErrorCodeInfo) error {
	if info.Code == "" {
		return fmt.Errorf("code cannot be empty")
	}

	if info.Message == "" {
		return fmt.Errorf("message cannot be empty")
	}

	if info.StatusCode < 100 || info.StatusCode > 599 {
		return fmt.Errorf("invalid HTTP status code: %d", info.StatusCode)
	}

	// Valida tipo de erro se especificado
	if info.Type != "" {
		errorType := types.ErrorType(info.Type)
		if !errorType.IsValid() {
			return fmt.Errorf("invalid error type: %s", info.Type)
		}
	}

	return nil
}

// registerCommonCodes registra os códigos de erro comuns automaticamente.
func (r *ErrorCodeRegistry) registerCommonCodes() {
	for code, commonInfo := range types.CommonErrorCodes {
		info := interfaces.ErrorCodeInfo{
			Code:        code,
			Message:     commonInfo.Message,
			Type:        string(commonInfo.Type),
			StatusCode:  commonInfo.StatusCode,
			Severity:    interfaces.Severity(commonInfo.Severity),
			Retryable:   commonInfo.Retryable,
			Temporary:   false,
			Tags:        []string{"common"},
			Description: fmt.Sprintf("Common error code: %s", commonInfo.Message),
			Examples:    []string{},
		}

		// Não falha se não conseguir registrar códigos comuns
		_ = r.Register(info)
	}
}

// basicError é uma implementação simples sem dependências circulares
type basicError struct {
	code       string
	message    string
	errorType  string
	statusCode int
	severity   types.ErrorSeverity
	tags       []string
	timestamp  time.Time
}

func (e *basicError) Error() string                                                  { return e.message }
func (e *basicError) Code() string                                                   { return e.code }
func (e *basicError) Type() string                                                   { return e.errorType }
func (e *basicError) Message() string                                                { return e.message }
func (e *basicError) StatusCode() int                                                { return e.statusCode }
func (e *basicError) Severity() interfaces.Severity                                  { return interfaces.Severity(e.severity) }
func (e *basicError) Tags() []string                                                 { return e.tags }
func (e *basicError) Timestamp() time.Time                                           { return e.timestamp }
func (e *basicError) IsRetryable() bool                                              { return false }
func (e *basicError) IsTemporary() bool                                              { return false }
func (e *basicError) Context() map[string]interface{}                                { return nil }
func (e *basicError) StackTrace() []string                                           { return nil }
func (e *basicError) Category() interfaces.Category                                  { return interfaces.CategoryTechnical }
func (e *basicError) Chain(err error) interfaces.DomainErrorInterface                { return e }
func (e *basicError) Unwrap() error                                                  { return nil }
func (e *basicError) Wrap(message string, err error) interfaces.DomainErrorInterface { return e }
func (e *basicError) RootCause() error                                               { return e }
func (e *basicError) String() string                                                 { return e.message }
func (e *basicError) JSON() ([]byte, error)                                          { return e.ToJSON() }
func (e *basicError) Metadata() map[string]interface{}                               { return e.Details() }
func (e *basicError) Headers() map[string]string                                     { return nil }
func (e *basicError) ResponseBody() interface{}                                      { return nil }
func (e *basicError) SetStatusCode(code int) interfaces.DomainErrorInterface {
	e.statusCode = code
	return e
}
func (e *basicError) ToJSON() ([]byte, error) {
	data := map[string]interface{}{
		"code":       e.code,
		"message":    e.message,
		"type":       e.errorType,
		"statusCode": e.statusCode,
		"severity":   e.severity.String(),
		"tags":       e.tags,
		"timestamp":  e.timestamp,
	}
	return json.Marshal(data)
}
func (e *basicError) FormatStackTrace() string { return "" }

func (e *basicError) Details() map[string]interface{} {
	return map[string]interface{}{
		"code":       e.code,
		"type":       e.errorType,
		"severity":   e.severity.String(),
		"statusCode": e.statusCode,
		"tags":       e.tags,
		"timestamp":  e.timestamp,
	}
}
func (e *basicError) DetailedString() string {
	return fmt.Sprintf("[%s] %s (type: %s, severity: %s)", e.code, e.message, e.errorType, e.severity.String())
}

// Singleton instance para uso global
var (
	globalRegistry interfaces.ErrorRegistry
	registryOnce   sync.Once
)

// GetGlobalRegistry retorna a instância global do registry (singleton).
func GetGlobalRegistry() interfaces.ErrorRegistry {
	registryOnce.Do(func() {
		globalRegistry = NewErrorCodeRegistry()
	})
	return globalRegistry
}

// SetGlobalRegistry define uma instância customizada como registry global.
func SetGlobalRegistry(registry interfaces.ErrorRegistry) {
	globalRegistry = registry
}

// Funções de conveniência que usam o registry global

// RegisterGlobal registra um código no registry global.
func RegisterGlobal(info interfaces.ErrorCodeInfo) error {
	return GetGlobalRegistry().Register(info)
}

// GetGlobal obtém um código do registry global.
func GetGlobal(code string) (interfaces.ErrorCodeInfo, bool) {
	return GetGlobalRegistry().Get(code)
}

// CreateErrorGlobal cria um erro usando o registry global.
func CreateErrorGlobal(code string, args ...interface{}) (interfaces.DomainErrorInterface, error) {
	return GetGlobalRegistry().CreateError(code, args...)
}

// ExistsGlobal verifica se um código existe no registry global.
func ExistsGlobal(code string) bool {
	return GetGlobalRegistry().Exists(code)
}
