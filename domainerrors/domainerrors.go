package domainerrors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/domainerrors/internal"
)

// DomainError implementa a interface DomainErrorInterface
type DomainError struct {
	id           string
	code         string
	message      string
	errorType    interfaces.ErrorType
	metadata     map[string]interface{}
	cause        error
	stack        []interfaces.StackFrame
	timestamp    time.Time
	context      context.Context
	stackCapture interfaces.StackTraceCapture
}

// ErrorFactory implementa a interface ErrorFactory
type ErrorFactory struct {
	stackCapture interfaces.StackTraceCapture
	mu           sync.RWMutex
}

// ErrorTypeChecker implementa verificação de tipos de erro
type ErrorTypeChecker struct{}

// Manager gerencia hooks, middlewares e observadores
type Manager struct {
	hookManager       *HookManager
	middlewareManager *MiddlewareManager
	observers         []interfaces.Observer
	mu                sync.RWMutex
}

// RegisterObserver registra um observador
func (m *Manager) RegisterObserver(observer interfaces.Observer) {
	if observer == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.observers = append(m.observers, observer)
}

// UnregisterObserver remove um observador
func (m *Manager) UnregisterObserver(observer interfaces.Observer) {
	if observer == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, obs := range m.observers {
		if obs == observer {
			m.observers = append(m.observers[:i], m.observers[i+1:]...)
			break
		}
	}
}

// NotifyObservers notifica todos os observadores
func (m *Manager) NotifyObservers(ctx context.Context, err interfaces.DomainErrorInterface) error {
	if err == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, observer := range m.observers {
		if observerErr := observer.OnError(ctx, err); observerErr != nil {
			return observerErr
		}
	}
	return nil
}

// HookManager gerencia hooks do sistema
type HookManager struct {
	startHooks []interfaces.StartHookFunc
	stopHooks  []interfaces.StopHookFunc
	errorHooks []interfaces.ErrorHookFunc
	i18nHooks  []interfaces.I18nHookFunc
	mu         sync.RWMutex
}

// MiddlewareManager gerencia middlewares do sistema
type MiddlewareManager struct {
	middlewares     []interfaces.MiddlewareFunc
	i18nMiddlewares []interfaces.I18nMiddlewareFunc
	mu              sync.RWMutex
}

// RegisterStartHook registra um hook de start
func (m *HookManager) RegisterStartHook(hook interfaces.StartHookFunc) {
	if hook == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startHooks = append(m.startHooks, hook)
}

// RegisterStopHook registra um hook de stop
func (m *HookManager) RegisterStopHook(hook interfaces.StopHookFunc) {
	if hook == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopHooks = append(m.stopHooks, hook)
}

// RegisterErrorHook registra um hook de erro
func (m *HookManager) RegisterErrorHook(hook interfaces.ErrorHookFunc) {
	if hook == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorHooks = append(m.errorHooks, hook)
}

// RegisterI18nHook registra um hook de i18n
func (m *HookManager) RegisterI18nHook(hook interfaces.I18nHookFunc) {
	if hook == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.i18nHooks = append(m.i18nHooks, hook)
}

// ExecuteStartHooks executa todos os hooks de start
func (m *HookManager) ExecuteStartHooks(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, hook := range m.startHooks {
		if err := hook(ctx); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteStopHooks executa todos os hooks de stop
func (m *HookManager) ExecuteStopHooks(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, hook := range m.stopHooks {
		if err := hook(ctx); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteErrorHooks executa todos os hooks de erro
func (m *HookManager) ExecuteErrorHooks(ctx context.Context, err interfaces.DomainErrorInterface) error {
	if err == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, hook := range m.errorHooks {
		if hookErr := hook(ctx, err); hookErr != nil {
			return hookErr
		}
	}
	return nil
}

// ExecuteI18nHooks executa todos os hooks de i18n
func (m *HookManager) ExecuteI18nHooks(ctx context.Context, err interfaces.DomainErrorInterface, locale string) error {
	if err == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, hook := range m.i18nHooks {
		if hookErr := hook(ctx, err, locale); hookErr != nil {
			return hookErr
		}
	}
	return nil
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

// ExecuteMiddlewares executa todos os middlewares
func (m *MiddlewareManager) ExecuteMiddlewares(ctx context.Context, err interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
	if err == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := err
	for _, middleware := range m.middlewares {
		result = middleware(ctx, result, func(e interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			return e
		})
	}
	return result
}

// ExecuteI18nMiddlewares executa todos os middlewares de i18n
func (m *MiddlewareManager) ExecuteI18nMiddlewares(ctx context.Context, err interfaces.DomainErrorInterface, locale string) interfaces.DomainErrorInterface {
	if err == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := err
	for _, middleware := range m.i18nMiddlewares {
		result = middleware(ctx, result, locale, func(e interfaces.DomainErrorInterface) interfaces.DomainErrorInterface {
			return e
		})
	}
	return result
}

// Instâncias globais
var (
	defaultFactory *ErrorFactory
	defaultManager *Manager
	defaultChecker *ErrorTypeChecker
	initOnce       sync.Once
)

func init() {
	initOnce.Do(func() {
		defaultFactory = NewErrorFactory(internal.DefaultStackTraceCapture())
		defaultManager = NewManager()
		defaultChecker = &ErrorTypeChecker{}
	})
}

// NewErrorFactory cria uma nova fábrica de erros
func NewErrorFactory(stackCapture interfaces.StackTraceCapture) *ErrorFactory {
	return &ErrorFactory{
		stackCapture: stackCapture,
	}
}

// NewManager cria um novo gerenciador
func NewManager() *Manager {
	return &Manager{
		hookManager:       &HookManager{},
		middlewareManager: &MiddlewareManager{},
		observers:         make([]interfaces.Observer, 0),
	}
}

// Error implementa a interface error
func (e *DomainError) Error() string {
	return e.message
}

// Unwrap retorna a causa raiz do erro
func (e *DomainError) Unwrap() error {
	return e.cause
}

// Type retorna o tipo do erro
func (e *DomainError) Type() interfaces.ErrorType {
	return e.errorType
}

// Metadata retorna os metadados do erro
func (e *DomainError) Metadata() map[string]interface{} {
	if e.metadata == nil {
		return make(map[string]interface{})
	}
	// Retorna uma cópia para evitar modificações externas
	result := make(map[string]interface{})
	for k, v := range e.metadata {
		result[k] = v
	}
	return result
}

// HTTPStatus retorna o código HTTP correspondente ao tipo de erro
func (e *DomainError) HTTPStatus() int {
	return MapHTTPStatus(e.errorType)
}

// StackTrace retorna o stack trace formatado
func (e *DomainError) StackTrace() string {
	if e.stackCapture == nil || len(e.stack) == 0 {
		return ""
	}
	return e.stackCapture.FormatStackTrace(e.stack)
}

// WithContext adiciona contexto ao erro
func (e *DomainError) WithContext(ctx context.Context) interfaces.DomainErrorInterface {
	newError := e.clone()
	newError.context = ctx
	return newError
}

// Wrap encapsula outro erro mantendo o contexto
func (e *DomainError) Wrap(err error) interfaces.DomainErrorInterface {
	newError := e.clone()
	newError.cause = err
	return newError
}

// WithMetadata adiciona metadados ao erro
func (e *DomainError) WithMetadata(key string, value interface{}) interfaces.DomainErrorInterface {
	newError := e.clone()
	if newError.metadata == nil {
		newError.metadata = make(map[string]interface{})
	}
	newError.metadata[key] = value
	return newError
}

// Code retorna o código único do erro
func (e *DomainError) Code() string {
	return e.code
}

// Timestamp retorna o momento da criação do erro
func (e *DomainError) Timestamp() time.Time {
	return e.timestamp
}

// ToJSON serializa o erro para JSON
func (e *DomainError) ToJSON() ([]byte, error) {
	type errorJSON struct {
		ID        string                  `json:"id"`
		Code      string                  `json:"code"`
		Message   string                  `json:"message"`
		Type      interfaces.ErrorType    `json:"type"`
		Metadata  map[string]interface{}  `json:"metadata,omitempty"`
		Stack     []interfaces.StackFrame `json:"stack,omitempty"`
		Timestamp time.Time               `json:"timestamp"`
		Cause     string                  `json:"cause,omitempty"`
	}

	jsonErr := errorJSON{
		ID:        e.id,
		Code:      e.code,
		Message:   e.message,
		Type:      e.errorType,
		Metadata:  e.metadata,
		Stack:     e.stack,
		Timestamp: e.timestamp,
	}

	if e.cause != nil {
		jsonErr.Cause = e.cause.Error()
	}

	return json.Marshal(jsonErr)
}

// clone cria uma cópia profunda do erro
func (e *DomainError) clone() *DomainError {
	newError := &DomainError{
		id:           e.id,
		code:         e.code,
		message:      e.message,
		errorType:    e.errorType,
		cause:        e.cause,
		stack:        e.stack,
		timestamp:    e.timestamp,
		context:      e.context,
		stackCapture: e.stackCapture,
	}

	// Clone metadata
	if e.metadata != nil {
		newError.metadata = make(map[string]interface{})
		for k, v := range e.metadata {
			newError.metadata[k] = v
		}
	}

	return newError
}

// New cria um novo erro de domínio
func (f *ErrorFactory) New(errorType interfaces.ErrorType, code, message string) interfaces.DomainErrorInterface {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return &DomainError{
		id:           uuid.New().String(),
		code:         code,
		message:      message,
		errorType:    errorType,
		metadata:     make(map[string]interface{}),
		timestamp:    time.Now(),
		stack:        f.stackCapture.CaptureStackTrace(1),
		stackCapture: f.stackCapture,
	}
}

// NewWithMetadata cria um novo erro com metadados
func (f *ErrorFactory) NewWithMetadata(errorType interfaces.ErrorType, code, message string, metadata map[string]interface{}) interfaces.DomainErrorInterface {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return &DomainError{
		id:           uuid.New().String(),
		code:         code,
		message:      message,
		errorType:    errorType,
		metadata:     metadata,
		timestamp:    time.Now(),
		stack:        f.stackCapture.CaptureStackTrace(1),
		stackCapture: f.stackCapture,
	}
}

// Wrap encapsula um erro existente
func (f *ErrorFactory) Wrap(err error, errorType interfaces.ErrorType, code, message string) interfaces.DomainErrorInterface {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return &DomainError{
		id:           uuid.New().String(),
		code:         code,
		message:      message,
		errorType:    errorType,
		metadata:     make(map[string]interface{}),
		cause:        err,
		timestamp:    time.Now(),
		stack:        f.stackCapture.CaptureStackTrace(1),
		stackCapture: f.stackCapture,
	}
}

// IsType verifica se um erro é de um tipo específico
func (c *ErrorTypeChecker) IsType(err error, errorType interfaces.ErrorType) bool {
	if err == nil {
		return false
	}

	// Verifica se é um DomainError
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Type() == errorType
	}

	// Verifica se implementa a interface
	if domainErr, ok := err.(interfaces.DomainErrorInterface); ok {
		return domainErr.Type() == errorType
	}

	// Tenta fazer unwrap recursivamente
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		return c.IsType(unwrapped, errorType)
	}

	return false
}

// MapHTTPStatus mapeia tipos de erro para códigos HTTP
func MapHTTPStatus(errorType interfaces.ErrorType) int {
	statusMap := map[interfaces.ErrorType]int{
		interfaces.ValidationError:           http.StatusBadRequest,           // 400
		interfaces.BadRequestError:           http.StatusBadRequest,           // 400
		interfaces.AuthenticationError:       http.StatusUnauthorized,         // 401
		interfaces.AuthorizationError:        http.StatusForbidden,            // 403
		interfaces.NotFoundError:             http.StatusNotFound,             // 404
		interfaces.ConflictError:             http.StatusConflict,             // 409
		interfaces.UnprocessableEntityError:  http.StatusUnprocessableEntity,  // 422
		interfaces.UnsupportedMediaTypeError: http.StatusUnsupportedMediaType, // 415
		interfaces.RateLimitError:            http.StatusTooManyRequests,      // 429
		interfaces.BusinessError:             http.StatusUnprocessableEntity,  // 422
		interfaces.WorkflowError:             http.StatusUnprocessableEntity,  // 422
		interfaces.DatabaseError:             http.StatusInternalServerError,  // 500
		interfaces.ExternalServiceError:      http.StatusBadGateway,           // 502
		interfaces.ServiceUnavailableError:   http.StatusServiceUnavailable,   // 503
		interfaces.TimeoutError:              http.StatusGatewayTimeout,       // 504
		interfaces.InfrastructureError:       http.StatusInternalServerError,  // 500
		interfaces.DependencyError:           http.StatusInternalServerError,  // 500
		interfaces.SecurityError:             http.StatusInternalServerError,  // 500
		interfaces.ResourceExhaustedError:    http.StatusInternalServerError,  // 500
		interfaces.CircuitBreakerError:       http.StatusServiceUnavailable,   // 503
		interfaces.SerializationError:        http.StatusInternalServerError,  // 500
		interfaces.CacheError:                http.StatusInternalServerError,  // 500
		interfaces.MigrationError:            http.StatusInternalServerError,  // 500
		interfaces.ConfigurationError:        http.StatusInternalServerError,  // 500
		interfaces.UnsupportedOperationError: http.StatusNotImplemented,       // 501
		interfaces.InvalidSchemaError:        http.StatusBadRequest,           // 400
		interfaces.ServerError:               http.StatusInternalServerError,  // 500
	}

	if status, exists := statusMap[errorType]; exists {
		return status
	}

	return http.StatusInternalServerError // Default 500
}

// Funções de conveniência globais

// New cria um novo erro usando a fábrica padrão
func New(errorType interfaces.ErrorType, code, message string) interfaces.DomainErrorInterface {
	return defaultFactory.New(errorType, code, message)
}

// NewWithMetadata cria um novo erro com metadados usando a fábrica padrão
func NewWithMetadata(errorType interfaces.ErrorType, code, message string, metadata map[string]interface{}) interfaces.DomainErrorInterface {
	return defaultFactory.NewWithMetadata(errorType, code, message, metadata)
}

// Wrap encapsula um erro existente usando a fábrica padrão
func Wrap(err error, errorType interfaces.ErrorType, code, message string) interfaces.DomainErrorInterface {
	return defaultFactory.Wrap(err, errorType, code, message)
}

// IsType verifica se um erro é de um tipo específico usando o verificador padrão
func IsType(err error, errorType interfaces.ErrorType) bool {
	return defaultChecker.IsType(err, errorType)
}

// GetManager retorna o gerenciador padrão
func GetManager() *Manager {
	return defaultManager
}

// GetFactory retorna a fábrica padrão
func GetFactory() *ErrorFactory {
	return defaultFactory
}

// GetChecker retorna o verificador padrão
func GetChecker() *ErrorTypeChecker {
	return defaultChecker
}

// Funções de conveniência para tipos específicos

// NewValidationError cria um erro de validação
func NewValidationError(code, message string) interfaces.DomainErrorInterface {
	return New(interfaces.ValidationError, code, message)
}

// NewNotFoundError cria um erro de não encontrado
func NewNotFoundError(code, message string) interfaces.DomainErrorInterface {
	return New(interfaces.NotFoundError, code, message)
}

// NewBusinessError cria um erro de negócio
func NewBusinessError(code, message string) interfaces.DomainErrorInterface {
	return New(interfaces.BusinessError, code, message)
}

// NewDatabaseError cria um erro de banco de dados
func NewDatabaseError(code, message string) interfaces.DomainErrorInterface {
	return New(interfaces.DatabaseError, code, message)
}

// NewAuthenticationError cria um erro de autenticação
func NewAuthenticationError(code, message string) interfaces.DomainErrorInterface {
	return New(interfaces.AuthenticationError, code, message)
}

// NewAuthorizationError cria um erro de autorização
func NewAuthorizationError(code, message string) interfaces.DomainErrorInterface {
	return New(interfaces.AuthorizationError, code, message)
}

// NewTimeoutError cria um erro de timeout
func NewTimeoutError(code, message string) interfaces.DomainErrorInterface {
	return New(interfaces.TimeoutError, code, message)
}

// NewRateLimitError cria um erro de rate limit
func NewRateLimitError(code, message string) interfaces.DomainErrorInterface {
	return New(interfaces.RateLimitError, code, message)
}

// Funções de utilitário para análise de erros

// GetRootCause retorna a causa raiz de uma cadeia de erros
func GetRootCause(err error) error {
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

// GetErrorChain retorna toda a cadeia de erros
func GetErrorChain(err error) []error {
	var chain []error
	current := err

	for current != nil {
		chain = append(chain, current)
		if unwrapped := errors.Unwrap(current); unwrapped != nil {
			current = unwrapped
		} else {
			break
		}
	}

	return chain
}

// FormatErrorChain formata uma cadeia de erros para exibição
func FormatErrorChain(err error) string {
	chain := GetErrorChain(err)
	if len(chain) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("Error chain:\n")

	for i, e := range chain {
		builder.WriteString(fmt.Sprintf("  %d. %s", i+1, e.Error()))
		if domainErr, ok := e.(interfaces.DomainErrorInterface); ok {
			builder.WriteString(fmt.Sprintf(" [%s:%s]", domainErr.Type(), domainErr.Code()))
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
