// Package domainerrors fornece uma estrutura robusta para tratamento de erros
// em aplicações Go, seguindo os princípios de Domain-Driven Design (DDD).
//
// Este pacote oferece:
// - Categorização de erros por tipos específicos
// - Empilhamento de erros com informações contextuais
// - Captura automática de stack trace
// - Mapeamento para códigos HTTP
// - Suporte para metadados e serialização JSON
// - Utilitários para manipulação e análise de erros
package domainerrors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/domainerrors/internal"
)

// Variáveis globais para configuração
var (
	// Configuração global para stack trace
	GlobalStackTraceEnabled = true
	GlobalMaxStackDepth     = 10
	GlobalSkipFrames        = 2
)

// ErrorType define os tipos de erro disponíveis
type ErrorType = interfaces.ErrorType

// Constantes dos tipos de erro
const (
	ErrorTypeValidation         = interfaces.ErrorTypeValidation
	ErrorTypeNotFound           = interfaces.ErrorTypeNotFound
	ErrorTypeBusiness           = interfaces.ErrorTypeBusiness
	ErrorTypeDatabase           = interfaces.ErrorTypeDatabase
	ErrorTypeExternalService    = interfaces.ErrorTypeExternalService
	ErrorTypeInfrastructure     = interfaces.ErrorTypeInfrastructure
	ErrorTypeDependency         = interfaces.ErrorTypeDependency
	ErrorTypeAuthentication     = interfaces.ErrorTypeAuthentication
	ErrorTypeAuthorization      = interfaces.ErrorTypeAuthorization
	ErrorTypeSecurity           = interfaces.ErrorTypeSecurity
	ErrorTypeTimeout            = interfaces.ErrorTypeTimeout
	ErrorTypeRateLimit          = interfaces.ErrorTypeRateLimit
	ErrorTypeResourceExhausted  = interfaces.ErrorTypeResourceExhausted
	ErrorTypeCircuitBreaker     = interfaces.ErrorTypeCircuitBreaker
	ErrorTypeSerialization      = interfaces.ErrorTypeSerialization
	ErrorTypeCache              = interfaces.ErrorTypeCache
	ErrorTypeMigration          = interfaces.ErrorTypeMigration
	ErrorTypeConfiguration      = interfaces.ErrorTypeConfiguration
	ErrorTypeUnsupported        = interfaces.ErrorTypeUnsupported
	ErrorTypeBadRequest         = interfaces.ErrorTypeBadRequest
	ErrorTypeConflict           = interfaces.ErrorTypeConflict
	ErrorTypeInvalidSchema      = interfaces.ErrorTypeInvalidSchema
	ErrorTypeUnsupportedMedia   = interfaces.ErrorTypeUnsupportedMedia
	ErrorTypeServer             = interfaces.ErrorTypeServer
	ErrorTypeUnprocessable      = interfaces.ErrorTypeUnprocessable
	ErrorTypeServiceUnavailable = interfaces.ErrorTypeServiceUnavailable
	ErrorTypeWorkflow           = interfaces.ErrorTypeWorkflow
)

// DomainError representa um erro de domínio genérico
type DomainError struct {
	CodeField        string                 `json:"code,omitempty"`
	Message          string                 `json:"message"`
	ErrorType        ErrorType              `json:"error_type"`
	Context          context.Context        `json:"-"`
	Cause            error                  `json:"-"`
	MetadataMap      map[string]interface{} `json:"metadata,omitempty"`
	Timestamp        time.Time              `json:"timestamp"`
	ID               string                 `json:"id"`
	StackTraceString string                 `json:"stack_trace,omitempty"`
} // New cria um novo erro de domínio
func New(code, message string) *DomainError {
	return &DomainError{
		CodeField:        code,
		Message:          message,
		ErrorType:        ErrorTypeBusiness,
		MetadataMap:      make(map[string]interface{}),
		Timestamp:        time.Now(),
		ID:               generateErrorID(),
		StackTraceString: captureStackTrace(),
	}
}

// NewWithError cria um novo erro de domínio encapsulando outro erro
func NewWithError(code, message string, cause error) *DomainError {
	return &DomainError{
		CodeField:        code,
		Message:          message,
		ErrorType:        ErrorTypeBusiness,
		Cause:            cause,
		MetadataMap:      make(map[string]interface{}),
		Timestamp:        time.Now(),
		ID:               generateErrorID(),
		StackTraceString: captureStackTrace(),
	}
}

// NewWithType cria um novo erro de domínio com tipo específico
func NewWithType(code, message string, errorType ErrorType) *DomainError {
	return &DomainError{
		CodeField:        code,
		Message:          message,
		ErrorType:        errorType,
		MetadataMap:      make(map[string]interface{}),
		Timestamp:        time.Now(),
		ID:               generateErrorID(),
		StackTraceString: captureStackTrace(),
	}
}

// Error implementa a interface error
func (e *DomainError) Error() string {
	var b strings.Builder

	if e.CodeField != "" {
		b.WriteString(fmt.Sprintf("[%s] ", e.CodeField))
	}

	b.WriteString(e.Message)

	if e.Cause != nil {
		b.WriteString(": ")
		b.WriteString(e.Cause.Error())
	}

	return b.String()
}

// Unwrap implementa a interface errors.Wrapper
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// Type retorna o tipo do erro
func (e *DomainError) Type() ErrorType {
	return e.ErrorType
}

// Code retorna o código do erro
func (e *DomainError) Code() string {
	return e.CodeField
}

// Metadata retorna os metadados do erro
func (e *DomainError) Metadata() map[string]interface{} {
	if e.MetadataMap == nil {
		e.MetadataMap = make(map[string]interface{})
	}
	return e.MetadataMap
}

// HTTPStatus retorna o código de status HTTP apropriado
func (e *DomainError) HTTPStatus() int {
	return mapErrorTypeToHTTPStatus(e.ErrorType)
}

// StackTrace retorna o stack trace capturado
func (e *DomainError) StackTrace() string {
	return e.StackTraceString
}

// WithContext adiciona contexto ao erro
func (e *DomainError) WithContext(ctx context.Context) interfaces.DomainError {
	e.Context = ctx
	return e
}

// Wrap encapsula outro erro com contexto opcional
func (e *DomainError) Wrap(message string, err error) interfaces.DomainError {
	wrapped := &DomainError{
		CodeField:        e.CodeField,
		Message:          message,
		ErrorType:        e.ErrorType,
		Cause:            err,
		MetadataMap:      make(map[string]interface{}),
		Timestamp:        time.Now(),
		ID:               generateErrorID(),
		StackTraceString: captureStackTrace(),
	}

	// Copia metadados do erro original
	for k, v := range e.MetadataMap {
		wrapped.MetadataMap[k] = v
	}

	return wrapped
}

// JSON serializa o erro para JSON
func (e *DomainError) JSON() ([]byte, error) {
	return json.Marshal(e)
}

// WithMetadata adiciona metadados ao erro
func (e *DomainError) WithMetadata(key string, value interface{}) *DomainError {
	if e.MetadataMap == nil {
		e.MetadataMap = make(map[string]interface{})
	}
	e.MetadataMap[key] = value
	return e
}

// WithMetadataMap adiciona múltiplos metadados ao erro
func (e *DomainError) WithMetadataMap(metadata map[string]interface{}) *DomainError {
	if e.MetadataMap == nil {
		e.MetadataMap = make(map[string]interface{})
	}
	for k, v := range metadata {
		e.MetadataMap[k] = v
	}
	return e
}

// WithType define o tipo do erro
func (e *DomainError) WithType(errorType ErrorType) *DomainError {
	e.ErrorType = errorType
	return e
}

// WithCode define o código do erro
func (e *DomainError) WithCode(code string) *DomainError {
	e.CodeField = code
	return e
}

// WithStackTrace habilita/desabilita captura de stack trace
func (e *DomainError) WithStackTrace(enabled bool) *DomainError {
	if enabled && e.StackTraceString == "" {
		e.StackTraceString = captureStackTrace()
	} else if !enabled {
		e.StackTraceString = ""
	}
	return e
}

// Implementações de tipos específicos de erro

// ValidationError representa um erro de validação
type ValidationError struct {
	*DomainError
	Fields map[string][]string `json:"fields,omitempty"`
}

// NewValidationError cria um erro de validação
func NewValidationError(message string, fields map[string][]string) *ValidationError {
	return &ValidationError{
		DomainError: NewWithType("VALIDATION_ERROR", message, ErrorTypeValidation),
		Fields:      fields,
	}
}

// WithField adiciona um campo de validação
func (e *ValidationError) WithField(field, message string) *ValidationError {
	if e.Fields == nil {
		e.Fields = make(map[string][]string)
	}
	e.Fields[field] = append(e.Fields[field], message)
	return e
}

// StatusCode implementa interfaces.HTTPStatusProvider
func (e *ValidationError) StatusCode() int {
	return http.StatusBadRequest
}

// NotFoundError representa um erro de recurso não encontrado
type NotFoundError struct {
	*DomainError
	Resource   string `json:"resource,omitempty"`
	ResourceID string `json:"resource_id,omitempty"`
}

// NewNotFoundError cria um erro de recurso não encontrado
func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{
		DomainError: NewWithType("NOT_FOUND", message, ErrorTypeNotFound),
	}
}

// WithResource adiciona informações do recurso
func (e *NotFoundError) WithResource(resource, resourceID string) *NotFoundError {
	e.Resource = resource
	e.ResourceID = resourceID
	return e
}

// StatusCode implementa interfaces.HTTPStatusProvider
func (e *NotFoundError) StatusCode() int {
	return http.StatusNotFound
}

// BusinessError representa um erro de regra de negócio
type BusinessError struct {
	*DomainError
	BusinessCode string `json:"business_code,omitempty"`
	RuleName     string `json:"rule_name,omitempty"`
}

// NewBusinessError cria um erro de regra de negócio
func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{
		DomainError:  NewWithType("BUSINESS_ERROR", message, ErrorTypeBusiness),
		BusinessCode: code,
	}
}

// WithRule adiciona informações da regra violada
func (e *BusinessError) WithRule(ruleName string) *BusinessError {
	e.RuleName = ruleName
	return e
}

// StatusCode implementa interfaces.HTTPStatusProvider
func (e *BusinessError) StatusCode() int {
	return http.StatusUnprocessableEntity
}

// Code implementa interfaces.HasCode
func (e *BusinessError) Code() string {
	return e.BusinessCode
}

// DatabaseError representa um erro de banco de dados
type DatabaseError struct {
	*DomainError
	Operation string `json:"operation,omitempty"`
	Table     string `json:"table,omitempty"`
	Query     string `json:"query,omitempty"`
}

// NewDatabaseError cria um erro de banco de dados
func NewDatabaseError(message string, cause error) *DatabaseError {
	return &DatabaseError{
		DomainError: NewWithError("DATABASE_ERROR", message, cause).WithType(ErrorTypeDatabase),
	}
}

// WithOperation adiciona informações da operação
func (e *DatabaseError) WithOperation(operation, table string) *DatabaseError {
	e.Operation = operation
	e.Table = table
	return e
}

// WithQuery adiciona a query SQL
func (e *DatabaseError) WithQuery(query string) *DatabaseError {
	e.Query = query
	return e
}

// StatusCode implementa interfaces.HTTPStatusProvider
func (e *DatabaseError) StatusCode() int {
	return http.StatusInternalServerError
}

// ExternalServiceError representa um erro de serviço externo
type ExternalServiceError struct {
	*DomainError
	Service        string `json:"service,omitempty"`
	Endpoint       string `json:"endpoint,omitempty"`
	HTTPStatusCode int    `json:"http_status_code,omitempty"`
	Response       string `json:"response,omitempty"`
}

// NewExternalServiceError cria um erro de serviço externo
func NewExternalServiceError(service, message string, cause error) *ExternalServiceError {
	return &ExternalServiceError{
		DomainError: NewWithError("EXTERNAL_SERVICE_ERROR", message, cause).WithType(ErrorTypeExternalService),
		Service:     service,
	}
}

// WithEndpoint adiciona informações do endpoint
func (e *ExternalServiceError) WithEndpoint(endpoint string) *ExternalServiceError {
	e.Endpoint = endpoint
	return e
}

// WithStatusCode adiciona o código de status HTTP
func (e *ExternalServiceError) WithStatusCode(statusCode int) *ExternalServiceError {
	e.HTTPStatusCode = statusCode
	return e
}

// WithResponse adiciona a resposta do serviço
func (e *ExternalServiceError) WithResponse(response string) *ExternalServiceError {
	e.Response = response
	return e
}

// StatusCode implementa interfaces.HTTPStatusProvider
func (e *ExternalServiceError) StatusCode() int {
	if e.HTTPStatusCode > 0 {
		return e.HTTPStatusCode
	}
	return http.StatusBadGateway
}

// InfrastructureError representa um erro de infraestrutura
type InfrastructureError struct {
	*DomainError
	Component string `json:"component,omitempty"`
	Details   string `json:"details,omitempty"`
}

// NewInfrastructureError cria um erro de infraestrutura
func NewInfrastructureError(component, message string, cause error) *InfrastructureError {
	return &InfrastructureError{
		DomainError: NewWithError("INFRASTRUCTURE_ERROR", message, cause).WithType(ErrorTypeInfrastructure),
		Component:   component,
	}
}

// WithDetails adiciona detalhes do erro
func (e *InfrastructureError) WithDetails(details string) *InfrastructureError {
	e.Details = details
	return e
}

// StatusCode implementa interfaces.HTTPStatusProvider
func (e *InfrastructureError) StatusCode() int {
	return http.StatusInternalServerError
}

// AuthenticationError representa um erro de autenticação
type AuthenticationError struct {
	*DomainError
	Reason string `json:"reason,omitempty"`
}

// NewAuthenticationError cria um erro de autenticação
func NewAuthenticationError(message string) *AuthenticationError {
	return &AuthenticationError{
		DomainError: NewWithType("AUTHENTICATION_ERROR", message, ErrorTypeAuthentication),
	}
}

// WithReason adiciona o motivo da falha
func (e *AuthenticationError) WithReason(reason string) *AuthenticationError {
	e.Reason = reason
	return e
}

// StatusCode implementa interfaces.HTTPStatusProvider
func (e *AuthenticationError) StatusCode() int {
	return http.StatusUnauthorized
}

// AuthorizationError representa um erro de autorização
type AuthorizationError struct {
	*DomainError
	Resource   string `json:"resource,omitempty"`
	Action     string `json:"action,omitempty"`
	Permission string `json:"permission,omitempty"`
}

// NewAuthorizationError cria um erro de autorização
func NewAuthorizationError(message string) *AuthorizationError {
	return &AuthorizationError{
		DomainError: NewWithType("AUTHORIZATION_ERROR", message, ErrorTypeAuthorization),
	}
}

// WithResource adiciona informações do recurso
func (e *AuthorizationError) WithResource(resource, action string) *AuthorizationError {
	e.Resource = resource
	e.Action = action
	return e
}

// WithPermission adiciona a permissão necessária
func (e *AuthorizationError) WithPermission(permission string) *AuthorizationError {
	e.Permission = permission
	return e
}

// StatusCode implementa interfaces.HTTPStatusProvider
func (e *AuthorizationError) StatusCode() int {
	return http.StatusForbidden
}

// TimeoutError representa um erro de timeout
type TimeoutError struct {
	*DomainError
	Duration time.Duration `json:"duration,omitempty"`
	Timeout  time.Duration `json:"timeout,omitempty"`
}

// NewTimeoutError cria um erro de timeout
func NewTimeoutError(message string, duration, timeout time.Duration) *TimeoutError {
	return &TimeoutError{
		DomainError: NewWithType("TIMEOUT_ERROR", message, ErrorTypeTimeout),
		Duration:    duration,
		Timeout:     timeout,
	}
}

// StatusCode implementa interfaces.HTTPStatusProvider
func (e *TimeoutError) StatusCode() int {
	return http.StatusRequestTimeout
}

// ServerError representa um erro interno do servidor
type ServerError struct {
	*DomainError
	RequestID     string `json:"request_id,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
	Component     string `json:"component,omitempty"`
}

// NewServerError cria um erro interno do servidor
func NewServerError(message string, cause error) *ServerError {
	return &ServerError{
		DomainError: NewWithError("SERVER_ERROR", message, cause).WithType(ErrorTypeServer),
	}
}

// WithRequestInfo adiciona informações da requisição
func (e *ServerError) WithRequestInfo(requestID, correlationID string) *ServerError {
	e.RequestID = requestID
	e.CorrelationID = correlationID
	return e
}

// WithComponent adiciona informações do componente
func (e *ServerError) WithComponent(component string) *ServerError {
	e.Component = component
	return e
}

// StatusCode implementa interfaces.HTTPStatusProvider
func (e *ServerError) StatusCode() int {
	return http.StatusInternalServerError
}

// UnprocessableEntityError representa um erro de entidade não processável
type UnprocessableEntityError struct {
	*DomainError
	EntityType string              `json:"entity_type,omitempty"`
	EntityID   string              `json:"entity_id,omitempty"`
	Violations []string            `json:"violations,omitempty"`
	Fields     map[string][]string `json:"fields,omitempty"`
}

// NewUnprocessableEntityError cria um erro de entidade não processável
func NewUnprocessableEntityError(message string) *UnprocessableEntityError {
	return &UnprocessableEntityError{
		DomainError: NewWithType("UNPROCESSABLE_ENTITY", message, ErrorTypeUnprocessable),
		Fields:      make(map[string][]string),
	}
}

// WithEntity adiciona informações da entidade
func (e *UnprocessableEntityError) WithEntity(entityType, entityID string) *UnprocessableEntityError {
	e.EntityType = entityType
	e.EntityID = entityID
	return e
}

// WithViolation adiciona uma violação
func (e *UnprocessableEntityError) WithViolation(violation string) *UnprocessableEntityError {
	e.Violations = append(e.Violations, violation)
	return e
}

// WithFieldError adiciona um erro de campo
func (e *UnprocessableEntityError) WithFieldError(field, message string) *UnprocessableEntityError {
	if e.Fields == nil {
		e.Fields = make(map[string][]string)
	}
	e.Fields[field] = append(e.Fields[field], message)
	return e
}

// StatusCode implementa interfaces.HTTPStatusProvider
func (e *UnprocessableEntityError) StatusCode() int {
	return http.StatusUnprocessableEntity
}

// Funções utilitárias

// IsType verifica se um erro é do tipo especificado
func IsType(err error, errorType ErrorType) bool {
	if err == nil {
		return false
	}

	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.ErrorType == errorType
	}

	// Verifica tipos específicos
	switch errorType {
	case ErrorTypeValidation:
		_, ok := err.(*ValidationError)
		return ok
	case ErrorTypeNotFound:
		_, ok := err.(*NotFoundError)
		return ok
	case ErrorTypeBusiness:
		_, ok := err.(*BusinessError)
		return ok
	case ErrorTypeDatabase:
		_, ok := err.(*DatabaseError)
		return ok
	case ErrorTypeExternalService:
		_, ok := err.(*ExternalServiceError)
		return ok
	case ErrorTypeInfrastructure:
		_, ok := err.(*InfrastructureError)
		return ok
	case ErrorTypeAuthentication:
		_, ok := err.(*AuthenticationError)
		return ok
	case ErrorTypeAuthorization:
		_, ok := err.(*AuthorizationError)
		return ok
	case ErrorTypeTimeout:
		_, ok := err.(*TimeoutError)
		return ok
	case ErrorTypeServer:
		_, ok := err.(*ServerError)
		return ok
	case ErrorTypeUnprocessable:
		_, ok := err.(*UnprocessableEntityError)
		return ok
	}

	return false
}

// MapHTTPStatus mapeia um tipo de erro para código HTTP
func MapHTTPStatus(errorType ErrorType) int {
	return mapErrorTypeToHTTPStatus(errorType)
}

// GetHTTPStatus retorna o código HTTP de um erro
func GetHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if provider, ok := err.(interfaces.HTTPStatusProvider); ok {
		return provider.StatusCode()
	}

	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.HTTPStatus()
	}

	return http.StatusInternalServerError
}

// Wrap encapsula um erro existente
func Wrap(message string, err error) *DomainError {
	return NewWithError("WRAPPED_ERROR", message, err)
}

// WrapWithType encapsula um erro com tipo específico
func WrapWithType(message string, err error, errorType ErrorType) *DomainError {
	return NewWithError("WRAPPED_ERROR", message, err).WithType(errorType)
}

// WrapWithCode encapsula um erro com código específico
func WrapWithCode(code, message string, err error) *DomainError {
	return NewWithError(code, message, err)
}

// WrapWithTypeAndCode encapsula um erro com tipo e código específicos
func WrapWithTypeAndCode(code, message string, err error, errorType ErrorType) *DomainError {
	return NewWithError(code, message, err).WithType(errorType)
}

// WrapWithContext encapsula um erro com contexto adicional
func WrapWithContext(ctx context.Context, message string, err error) *DomainError {
	wrapped := NewWithError("WRAPPED_ERROR", message, err)
	wrapped.WithContext(ctx)
	return wrapped
}

// WrapWithMetadata encapsula um erro com metadados adicionais
func WrapWithMetadata(message string, err error, metadata map[string]interface{}) *DomainError {
	wrapped := NewWithError("WRAPPED_ERROR", message, err)
	wrapped.WithMetadataMap(metadata)
	return wrapped
}

// ErrorStack representa uma pilha de erros com funcionalidades avançadas
type ErrorStack struct {
	errors []error
	root   error
}

// NewErrorStack cria uma nova pilha de erros
func NewErrorStack(root error) *ErrorStack {
	return &ErrorStack{
		errors: []error{root},
		root:   root,
	}
}

// Push adiciona um erro ao topo da pilha
func (es *ErrorStack) Push(err error) *ErrorStack {
	es.errors = append(es.errors, err)
	return es
}

// Pop remove e retorna o erro do topo da pilha
func (es *ErrorStack) Pop() error {
	if len(es.errors) == 0 {
		return nil
	}

	top := es.errors[len(es.errors)-1]
	es.errors = es.errors[:len(es.errors)-1]
	return top
}

// Peek retorna o erro do topo da pilha sem removê-lo
func (es *ErrorStack) Peek() error {
	if len(es.errors) == 0 {
		return nil
	}
	return es.errors[len(es.errors)-1]
}

// Root retorna a causa raiz da pilha
func (es *ErrorStack) Root() error {
	return es.root
}

// Size retorna o número de erros na pilha
func (es *ErrorStack) Size() int {
	return len(es.errors)
}

// IsEmpty verifica se a pilha está vazia
func (es *ErrorStack) IsEmpty() bool {
	return len(es.errors) == 0
}

// ToSlice retorna uma cópia dos erros como slice
func (es *ErrorStack) ToSlice() []error {
	result := make([]error, len(es.errors))
	copy(result, es.errors)
	return result
}

// Error implementa a interface error
func (es *ErrorStack) Error() string {
	if len(es.errors) == 0 {
		return ""
	}

	if len(es.errors) == 1 {
		return es.errors[0].Error()
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("error stack (%d errors):", len(es.errors)))
	for i, err := range es.errors {
		b.WriteString(fmt.Sprintf("\n  %d. %s", i+1, err.Error()))
	}
	return b.String()
}

// Unwrap retorna o erro anterior na pilha
func (es *ErrorStack) Unwrap() error {
	if len(es.errors) <= 1 {
		return nil
	}
	return es.errors[len(es.errors)-2]
}

// ErrorWrapper é uma estrutura avançada para wrapping de erros
type ErrorWrapper struct {
	originalError error
	wrappedError  error
	chain         []error
	metadata      map[string]interface{}
	timestamp     time.Time
}

// NewErrorWrapper cria um novo wrapper de erro
func NewErrorWrapper(err error) *ErrorWrapper {
	return &ErrorWrapper{
		originalError: err,
		wrappedError:  err,
		chain:         []error{err},
		metadata:      make(map[string]interface{}),
		timestamp:     time.Now(),
	}
}

// Wrap adiciona uma camada de wrapping ao erro
func (ew *ErrorWrapper) Wrap(message string) *ErrorWrapper {
	wrapped := fmt.Errorf("%s: %w", message, ew.wrappedError)
	ew.wrappedError = wrapped
	ew.chain = append(ew.chain, wrapped)
	return ew
}

// WrapWithCode adiciona uma camada de wrapping com código
func (ew *ErrorWrapper) WrapWithCode(code, message string) *ErrorWrapper {
	wrapped := NewWithError(code, message, ew.wrappedError)
	ew.wrappedError = wrapped
	ew.chain = append(ew.chain, wrapped)
	return ew
}

// WrapWithType adiciona uma camada de wrapping com tipo
func (ew *ErrorWrapper) WrapWithType(message string, errorType ErrorType) *ErrorWrapper {
	wrapped := NewWithError("WRAPPED_ERROR", message, ew.wrappedError).WithType(errorType)
	ew.wrappedError = wrapped
	ew.chain = append(ew.chain, wrapped)
	return ew
}

// WithMetadata adiciona metadados ao wrapper
func (ew *ErrorWrapper) WithMetadata(key string, value interface{}) *ErrorWrapper {
	ew.metadata[key] = value
	return ew
}

// WithContext adiciona contexto ao wrapper
func (ew *ErrorWrapper) WithContext(ctx context.Context) *ErrorWrapper {
	ew.metadata["context"] = ctx
	return ew
}

// Root retorna o erro original (raiz)
func (ew *ErrorWrapper) Root() error {
	return ew.originalError
}

// Current retorna o erro atual (wrapped)
func (ew *ErrorWrapper) Current() error {
	return ew.wrappedError
}

// Chain retorna toda a cadeia de erros
func (ew *ErrorWrapper) Chain() []error {
	result := make([]error, len(ew.chain))
	copy(result, ew.chain)
	return result
}

// Depth retorna a profundidade da cadeia de erros
func (ew *ErrorWrapper) Depth() int {
	return len(ew.chain)
}

// Error implementa a interface error
func (ew *ErrorWrapper) Error() string {
	return ew.wrappedError.Error()
}

// Unwrap implementa a interface errors.Wrapper
func (ew *ErrorWrapper) Unwrap() error {
	return ew.originalError
}

// ErrorChainNavigator permite navegar na cadeia de erros
type ErrorChainNavigator struct {
	chain   []error
	current int
}

// NewErrorChainNavigator cria um novo navegador de cadeia de erros
func NewErrorChainNavigator(err error) *ErrorChainNavigator {
	chain := GetErrorChain(err)
	return &ErrorChainNavigator{
		chain:   chain,
		current: 0,
	}
}

// Next move para o próximo erro na cadeia
func (ecn *ErrorChainNavigator) Next() error {
	if ecn.current < len(ecn.chain)-1 {
		ecn.current++
		return ecn.chain[ecn.current]
	}
	return nil
}

// Previous move para o erro anterior na cadeia
func (ecn *ErrorChainNavigator) Previous() error {
	if ecn.current > 0 {
		ecn.current--
		return ecn.chain[ecn.current]
	}
	return nil
}

// Current retorna o erro atual
func (ecn *ErrorChainNavigator) Current() error {
	if ecn.current < len(ecn.chain) {
		return ecn.chain[ecn.current]
	}
	return nil
}

// Root retorna a causa raiz
func (ecn *ErrorChainNavigator) Root() error {
	if len(ecn.chain) > 0 {
		return ecn.chain[len(ecn.chain)-1]
	}
	return nil
}

// Top retorna o erro do topo da cadeia
func (ecn *ErrorChainNavigator) Top() error {
	if len(ecn.chain) > 0 {
		return ecn.chain[0]
	}
	return nil
}

// HasNext verifica se há próximo erro
func (ecn *ErrorChainNavigator) HasNext() bool {
	return ecn.current < len(ecn.chain)-1
}

// HasPrevious verifica se há erro anterior
func (ecn *ErrorChainNavigator) HasPrevious() bool {
	return ecn.current > 0
}

// Position retorna a posição atual na cadeia
func (ecn *ErrorChainNavigator) Position() int {
	return ecn.current
}

// Size retorna o tamanho da cadeia
func (ecn *ErrorChainNavigator) Size() int {
	return len(ecn.chain)
}

// Reset volta para o início da cadeia
func (ecn *ErrorChainNavigator) Reset() {
	ecn.current = 0
}

// GoToRoot vai para a causa raiz
func (ecn *ErrorChainNavigator) GoToRoot() error {
	if len(ecn.chain) > 0 {
		ecn.current = len(ecn.chain) - 1
		return ecn.chain[ecn.current]
	}
	return nil
}

// GoToTop vai para o topo da cadeia
func (ecn *ErrorChainNavigator) GoToTop() error {
	if len(ecn.chain) > 0 {
		ecn.current = 0
		return ecn.chain[ecn.current]
	}
	return nil
}

// GetAll retorna todos os erros da cadeia
func (ecn *ErrorChainNavigator) GetAll() []error {
	result := make([]error, len(ecn.chain))
	copy(result, ecn.chain)
	return result
}

// FindByType encontra o primeiro erro do tipo especificado
func (ecn *ErrorChainNavigator) FindByType(errorType ErrorType) error {
	for _, err := range ecn.chain {
		if IsType(err, errorType) {
			return err
		}
	}
	return nil
}

// FindByCode encontra o primeiro erro com o código especificado
func (ecn *ErrorChainNavigator) FindByCode(code string) error {
	for _, err := range ecn.chain {
		if hasCode, ok := err.(interfaces.HasCode); ok && hasCode.Code() == code {
			return err
		}
	}
	return nil
}

// FilterByType filtra erros por tipo
func (ecn *ErrorChainNavigator) FilterByType(errorType ErrorType) []error {
	var filtered []error
	for _, err := range ecn.chain {
		if IsType(err, errorType) {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// Funções utilitárias avançadas para wrapping

// Wrapf encapsula um erro com formatação
func Wrapf(format string, err error, args ...interface{}) *DomainError {
	message := fmt.Sprintf(format, args...)
	return NewWithError("WRAPPED_ERROR", message, err)
}

// WrapWithCodef encapsula um erro com código e formatação
func WrapWithCodef(code, format string, err error, args ...interface{}) *DomainError {
	message := fmt.Sprintf(format, args...)
	return NewWithError(code, message, err)
}

// WrapWithTypef encapsula um erro com tipo e formatação
func WrapWithTypef(format string, err error, errorType ErrorType, args ...interface{}) *DomainError {
	message := fmt.Sprintf(format, args...)
	return NewWithError("WRAPPED_ERROR", message, err).WithType(errorType)
}

// WrapMultiple encapsula múltiplos erros em uma cadeia
func WrapMultiple(message string, errors ...error) *DomainError {
	if len(errors) == 0 {
		return nil
	}

	if len(errors) == 1 {
		return Wrap(message, errors[0])
	}

	// Criar cadeia de erros
	wrapped := Wrap(message, errors[0])
	for i := 1; i < len(errors); i++ {
		wrapped = Wrap(fmt.Sprintf("%s (error %d)", message, i+1), errors[i])
	}

	return wrapped
}

// WrapWithCause encapsula um erro preservando a causa original
func WrapWithCause(message string, err error, cause error) *DomainError {
	wrapped := NewWithError("WRAPPED_ERROR", message, err)
	wrapped.WithMetadata("original_cause", cause)
	return wrapped
}

// UnwrapAll desempacota todos os erros até a raiz
func UnwrapAll(err error) error {
	for {
		unwrapped := unwrapError(err)
		if unwrapped == nil {
			break
		}
		err = unwrapped
	}
	return err
}

// UnwrapToType desempacota até encontrar um erro do tipo especificado
func UnwrapToType(err error, errorType ErrorType) error {
	for err != nil {
		if IsType(err, errorType) {
			return err
		}
		err = unwrapError(err)
	}
	return nil
}

// UnwrapToCode desempacota até encontrar um erro com o código especificado
func UnwrapToCode(err error, code string) error {
	for err != nil {
		if hasCode, ok := err.(interfaces.HasCode); ok && hasCode.Code() == code {
			return err
		}
		err = unwrapError(err)
	}
	return nil
}

// GetErrorAtDepth retorna o erro em uma profundidade específica
func GetErrorAtDepth(err error, depth int) error {
	chain := GetErrorChain(err)
	if depth >= 0 && depth < len(chain) {
		return chain[depth]
	}
	return nil
}

// GetErrorDepth retorna a profundidade de um erro específico na cadeia
func GetErrorDepth(err error, target error) int {
	chain := GetErrorChain(err)
	for i, e := range chain {
		if e == target {
			return i
		}
	}
	return -1
}

// HasErrorInChain verifica se um erro específico está na cadeia
func HasErrorInChain(err error, target error) bool {
	return GetErrorDepth(err, target) != -1
}

// HasErrorTypeInChain verifica se um tipo de erro está na cadeia
func HasErrorTypeInChain(err error, errorType ErrorType) bool {
	return UnwrapToType(err, errorType) != nil
}

// HasErrorCodeInChain verifica se um código de erro está na cadeia
func HasErrorCodeInChain(err error, code string) bool {
	return UnwrapToCode(err, code) != nil
}

// Funções internas

// captureStackTrace captura o stack trace atual
func captureStackTrace() string {
	if !GlobalStackTraceEnabled {
		return ""
	}

	config := &internal.StackTraceConfig{
		Enabled:    true,
		MaxDepth:   GlobalMaxStackDepth,
		SkipFrames: GlobalSkipFrames + 1, // +1 para pular esta função
	}

	st := internal.NewStackTrace(config)
	return st.String()
}

// generateErrorID gera um ID único para o erro
func generateErrorID() string {
	return fmt.Sprintf("err_%d", time.Now().UnixNano())
}

// mapErrorTypeToHTTPStatus mapeia tipos de erro para códigos HTTP
func mapErrorTypeToHTTPStatus(errorType ErrorType) int {
	switch errorType {
	case ErrorTypeValidation, ErrorTypeBadRequest, ErrorTypeInvalidSchema:
		return http.StatusBadRequest
	case ErrorTypeAuthentication:
		return http.StatusUnauthorized
	case ErrorTypeAuthorization, ErrorTypeSecurity:
		return http.StatusForbidden
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeTimeout:
		return http.StatusRequestTimeout
	case ErrorTypeBusiness, ErrorTypeUnprocessable:
		return http.StatusUnprocessableEntity
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	case ErrorTypeServiceUnavailable, ErrorTypeCircuitBreaker:
		return http.StatusServiceUnavailable
	case ErrorTypeUnsupported:
		return http.StatusNotImplemented
	case ErrorTypeUnsupportedMedia:
		return http.StatusUnsupportedMediaType
	case ErrorTypeExternalService:
		return http.StatusBadGateway
	case ErrorTypeResourceExhausted:
		return http.StatusInsufficientStorage
	case ErrorTypeDependency:
		return http.StatusFailedDependency
	default:
		return http.StatusInternalServerError
	}
}

// Configuração global

// SetGlobalStackTraceEnabled habilita/desabilita stack trace globalmente
func SetGlobalStackTraceEnabled(enabled bool) {
	GlobalStackTraceEnabled = enabled
	internal.SetGlobalStackTraceEnabled(enabled)
}

// SetGlobalMaxStackDepth define a profundidade máxima do stack trace
func SetGlobalMaxStackDepth(depth int) {
	GlobalMaxStackDepth = depth
}

// SetGlobalSkipFrames define quantos frames pular no stack trace
func SetGlobalSkipFrames(skip int) {
	GlobalSkipFrames = skip
}

// GetRootCause retorna a causa raiz de um erro
func GetRootCause(err error) error {
	for err != nil {
		if unwrapped := unwrapError(err); unwrapped != nil {
			err = unwrapped
		} else {
			break
		}
	}
	return err
}

// unwrapError tenta fazer unwrap de um erro
func unwrapError(err error) error {
	if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
		return unwrapper.Unwrap()
	}
	return nil
}

// GetErrorChain retorna toda a cadeia de erros
func GetErrorChain(err error) []error {
	var chain []error
	for err != nil {
		chain = append(chain, err)
		if unwrapped := unwrapError(err); unwrapped != nil {
			err = unwrapped
		} else {
			break
		}
	}
	return chain
}

// FormatErrorChain formata a cadeia de erros
func FormatErrorChain(err error) string {
	chain := GetErrorChain(err)
	if len(chain) == 0 {
		return ""
	}

	var b strings.Builder
	for i, e := range chain {
		if i > 0 {
			b.WriteString(" -> ")
		}
		b.WriteString(e.Error())
	}
	return b.String()
}

// GetCallerInfo retorna informações sobre quem chamou a função
func GetCallerInfo(skip int) (string, string, int) {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "", "", 0
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "", file, line
	}

	return fn.Name(), file, line
}

// WithCallerInfo adiciona informações do caller aos metadados
func WithCallerInfo(err *DomainError) *DomainError {
	function, file, line := GetCallerInfo(1)
	return err.WithMetadata("caller", map[string]interface{}{
		"function": function,
		"file":     file,
		"line":     line,
	})
}

// ErrorGroup agrupa múltiplos erros
type ErrorGroup struct {
	Errors []error
}

// NewErrorGroup cria um novo grupo de erros
func NewErrorGroup() *ErrorGroup {
	return &ErrorGroup{
		Errors: make([]error, 0),
	}
}

// Add adiciona um erro ao grupo
func (eg *ErrorGroup) Add(err error) {
	if err != nil {
		eg.Errors = append(eg.Errors, err)
	}
}

// HasErrors retorna se há erros no grupo
func (eg *ErrorGroup) HasErrors() bool {
	return len(eg.Errors) > 0
}

// Error implementa a interface error
func (eg *ErrorGroup) Error() string {
	if len(eg.Errors) == 0 {
		return ""
	}

	if len(eg.Errors) == 1 {
		return eg.Errors[0].Error()
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("multiple errors (%d):", len(eg.Errors)))
	for i, err := range eg.Errors {
		b.WriteString(fmt.Sprintf("\n  %d. %s", i+1, err.Error()))
	}
	return b.String()
}

// First retorna o primeiro erro
func (eg *ErrorGroup) First() error {
	if len(eg.Errors) == 0 {
		return nil
	}
	return eg.Errors[0]
}

// Last retorna o último erro
func (eg *ErrorGroup) Last() error {
	if len(eg.Errors) == 0 {
		return nil
	}
	return eg.Errors[len(eg.Errors)-1]
}

// Count retorna o número de erros
func (eg *ErrorGroup) Count() int {
	return len(eg.Errors)
}

// Clear limpa todos os erros
func (eg *ErrorGroup) Clear() {
	eg.Errors = eg.Errors[:0]
}

// ToSlice retorna uma cópia dos erros como slice
func (eg *ErrorGroup) ToSlice() []error {
	result := make([]error, len(eg.Errors))
	copy(result, eg.Errors)
	return result
}

// FilterByType filtra erros por tipo
func (eg *ErrorGroup) FilterByType(errorType ErrorType) []error {
	var filtered []error
	for _, err := range eg.Errors {
		if IsType(err, errorType) {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// GetTypeName retorna o nome do tipo do erro
func GetTypeName(err error) string {
	if err == nil {
		return ""
	}
	return reflect.TypeOf(err).String()
}

// IsRecoverable verifica se um erro é recuperável
func IsRecoverable(err error) bool {
	if err == nil {
		return false
	}

	// Erros não recuperáveis
	nonRecoverableTypes := []ErrorType{
		ErrorTypeAuthentication,
		ErrorTypeAuthorization,
		ErrorTypeNotFound,
		ErrorTypeValidation,
		ErrorTypeBusiness,
		ErrorTypeConfiguration,
	}

	for _, errorType := range nonRecoverableTypes {
		if IsType(err, errorType) {
			return false
		}
	}

	return true
}

// ShouldRetry verifica se uma operação deve ser repetida
func ShouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Tipos que podem ser repetidos
	retryableTypes := []ErrorType{
		ErrorTypeTimeout,
		ErrorTypeExternalService,
		ErrorTypeInfrastructure,
		ErrorTypeDatabase,
		ErrorTypeCircuitBreaker,
		ErrorTypeResourceExhausted,
		ErrorTypeServiceUnavailable,
		ErrorTypeDependency,
	}

	for _, errorType := range retryableTypes {
		if IsType(err, errorType) {
			return true
		}
	}

	return false
}

// GetSeverity retorna a severidade do erro
func GetSeverity(err error) string {
	if err == nil {
		return "none"
	}

	highSeverityTypes := []ErrorType{
		ErrorTypeSecurity,
		ErrorTypeAuthentication,
		ErrorTypeAuthorization,
		ErrorTypeDatabase,
		ErrorTypeServer,
	}

	mediumSeverityTypes := []ErrorType{
		ErrorTypeExternalService,
		ErrorTypeInfrastructure,
		ErrorTypeTimeout,
		ErrorTypeCircuitBreaker,
		ErrorTypeResourceExhausted,
	}

	for _, errorType := range highSeverityTypes {
		if IsType(err, errorType) {
			return "high"
		}
	}

	for _, errorType := range mediumSeverityTypes {
		if IsType(err, errorType) {
			return "medium"
		}
	}

	return "low"
}

// Must converte um erro em panic se não for nil
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustReturn retorna um valor ou causa panic se houver erro
func MustReturn[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// Recover recupera de um panic e retorna como erro
func Recover() error {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			return err
		}
		return fmt.Errorf("panic: %v", r)
	}
	return nil
}

// RecoverWithStackTrace recupera de um panic com stack trace
func RecoverWithStackTrace() error {
	if r := recover(); r != nil {
		stackTrace := internal.CaptureStackTrace()
		if err, ok := r.(error); ok {
			return NewWithError("PANIC_RECOVERED", "panic recovered", err).
				WithStackTrace(true).
				WithMetadata("panic_value", r).
				WithMetadata("stack_trace", stackTrace)
		}
		return NewWithType("PANIC_RECOVERED", fmt.Sprintf("panic recovered: %v", r), ErrorTypeServer).
			WithStackTrace(true).
			WithMetadata("panic_value", r).
			WithMetadata("stack_trace", stackTrace)
	}
	return nil
}
