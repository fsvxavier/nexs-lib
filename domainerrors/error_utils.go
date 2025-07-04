package domainerrors

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrorParser é uma interface para parsers de erro específicos
type ErrorParser interface {
	Parse(err error) (string, bool)
}

// SQLErrorParser parseia erros SQL para extrair códigos de erro
type SQLErrorParser struct {
	// regex para extrair códigos de erro SQL
	Regex    *regexp.Regexp
	CodeFunc func(string) string
}

// NewSQLErrorParser cria um novo parser de erros SQL
func NewSQLErrorParser() *SQLErrorParser {
	return &SQLErrorParser{
		Regex: regexp.MustCompile(`^(.*)\(SQLSTATE (.*)\).*$`),
		CodeFunc: func(match string) string {
			return match
		},
	}
}

// Parse extrai o código de erro SQL de uma mensagem de erro
func (p *SQLErrorParser) Parse(err error) (string, bool) {
	if err == nil {
		return "", false
	}

	matches := p.Regex.FindStringSubmatch(err.Error())
	if len(matches) < 3 {
		return "", false
	}

	code := p.CodeFunc(matches[2])
	return code, true
}

// ErrorCodeInfo representa um código de erro específico com sua descrição e código de status
type ErrorCodeInfo struct {
	Code        string
	Description string
	StatusCode  int
}

// ErrorCodeRegistry mantém um registro de códigos de erro conhecidos
type ErrorCodeRegistry struct {
	codes map[string]ErrorCodeInfo
}

// NewErrorCodeRegistry cria um novo registro de códigos de erro
func NewErrorCodeRegistry() *ErrorCodeRegistry {
	return &ErrorCodeRegistry{
		codes: make(map[string]ErrorCodeInfo),
	}
}

// Register adiciona um novo código de erro ao registro
func (r *ErrorCodeRegistry) Register(code string, description string, statusCode int) {
	r.codes[code] = ErrorCodeInfo{
		Code:        code,
		Description: description,
		StatusCode:  statusCode,
	}
}

// Get obtém um código de erro pelo seu código
func (r *ErrorCodeRegistry) Get(code string) (ErrorCodeInfo, bool) {
	ec, ok := r.codes[code]
	return ec, ok
}

// WrapWithCode envolve um erro com um código registrado
func (r *ErrorCodeRegistry) WrapWithCode(code string, err error) error {
	if err == nil {
		return nil
	}

	errorCode, ok := r.Get(code)
	if !ok {
		return WrapError(code, "Unknown error", err)
	}

	return WrapError(code, errorCode.Description, err)
}

// ErrorStack representa um stack de erros para análise e formatação
type ErrorStack struct {
	Errors []error
}

// NewErrorStack cria um novo stack de erros
func NewErrorStack() *ErrorStack {
	return &ErrorStack{
		Errors: make([]error, 0),
	}
}

// Push adiciona um erro ao stack
func (s *ErrorStack) Push(err error) {
	if err != nil {
		s.Errors = append(s.Errors, err)
	}
}

// IsEmpty verifica se o stack está vazio
func (s *ErrorStack) IsEmpty() bool {
	return len(s.Errors) == 0
}

// Error implementa a interface error
func (s *ErrorStack) Error() string {
	if s.IsEmpty() {
		return ""
	}

	var msgs []string
	for _, err := range s.Errors {
		msgs = append(msgs, err.Error())
	}

	return strings.Join(msgs, "\n")
}

// Unwrap retorna o primeiro erro do stack
// Implementa errors.Unwrapper
func (s *ErrorStack) Unwrap() error {
	if s.IsEmpty() {
		return nil
	}
	return s.Errors[0]
}

// ToSlice retorna o stack como um slice de erros
func (s *ErrorStack) ToSlice() []error {
	return s.Errors
}

// Format formata o stack de erros para exibição
func (s *ErrorStack) Format() string {
	if s.IsEmpty() {
		return "No errors"
	}

	var b strings.Builder
	b.WriteString("Error stack:\n")

	for i, err := range s.Errors {
		b.WriteString(fmt.Sprintf("[%d] %s\n", i+1, err.Error()))

		// Se o erro contém um stack trace, inclua
		if de, ok := err.(*DomainError); ok {
			trace := de.FormatStackTrace()
			if trace != "" {
				b.WriteString("  └─ ")
				b.WriteString(strings.ReplaceAll(trace, "\n", "\n     "))
				b.WriteString("\n")
			}
		}
	}

	return b.String()
}

// RecoverHandler é uma função para lidar com pânicos e convertê-los em erros de domínio
func RecoverHandler(r interface{}) error {
	switch v := r.(type) {
	case string:
		return New("PANIC", "Recovered from panic: "+v).WithType(ErrorTypeInternal)
	case error:
		return NewWithError("PANIC", "Recovered from panic", v).WithType(ErrorTypeInternal)
	default:
		return New("PANIC", fmt.Sprintf("Recovered from panic: %v", v)).WithType(ErrorTypeInternal)
	}
}

// RecoverMiddleware é uma função middleware para recuperar de pânicos
func RecoverMiddleware(next func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = RecoverHandler(r)
		}
	}()
	return next()
}

// FormatErrorChain formata uma cadeia de erros aninhados
func FormatErrorChain(err error) string {
	if err == nil {
		return "no error"
	}

	var b strings.Builder
	b.WriteString(err.Error())

	// Unwrap erros aninhados
	current := err
	for {
		unwrapped := errors.Unwrap(current)
		if unwrapped == nil {
			break
		}
		b.WriteString("\n  └─ ")
		b.WriteString(unwrapped.Error())
		current = unwrapped
	}

	return b.String()
}

// IsErrorType verifica se um erro é de um determinado tipo
func IsErrorType(err error, errType ErrorType) bool {
	// Primeiro, tenta verificar se é um DomainError direto
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Type == errType
	}

	// Se não é um DomainError direto, verifica os tipos específicos
	switch errType {
	case ErrorTypeValidation:
		var validationErr *ValidationError
		if errors.As(err, &validationErr) {
			return validationErr.DomainError.Type == errType
		}
	case ErrorTypeNotFound:
		var notFoundErr *NotFoundError
		if errors.As(err, &notFoundErr) {
			return notFoundErr.DomainError.Type == errType
		}
	case ErrorTypeBusinessRule:
		var businessErr *BusinessError
		if errors.As(err, &businessErr) {
			return businessErr.DomainError.Type == errType
		}
	case ErrorTypeInfrastructure:
		var infraErr *InfrastructureError
		if errors.As(err, &infraErr) {
			return infraErr.DomainError.Type == errType
		}
	case ErrorTypeExternalService:
		var extErr *ExternalServiceError
		if errors.As(err, &extErr) {
			return extErr.DomainError.Type == errType
		}
	case ErrorTypeAuthentication:
		var authErr *AuthenticationError
		if errors.As(err, &authErr) {
			return authErr.DomainError.Type == errType
		}
	case ErrorTypeAuthorization:
		var authzErr *AuthorizationError
		if errors.As(err, &authzErr) {
			return authzErr.DomainError.Type == errType
		}
	case ErrorTypeTimeout:
		var timeoutErr *TimeoutError
		if errors.As(err, &timeoutErr) {
			return timeoutErr.DomainError.Type == errType
		}
	case ErrorTypeUnsupported:
		var unsupportedErr *UnsupportedOperationError
		if errors.As(err, &unsupportedErr) {
			return unsupportedErr.DomainError.Type == errType
		}
	case ErrorTypeBadRequest:
		var badReqErr *BadRequestError
		if errors.As(err, &badReqErr) {
			return badReqErr.DomainError.Type == errType
		}
	case ErrorTypeConflict:
		var conflictErr *ConflictError
		if errors.As(err, &conflictErr) {
			return conflictErr.DomainError.Type == errType
		}
	case ErrorTypeRateLimit:
		var rateLimitErr *RateLimitError
		if errors.As(err, &rateLimitErr) {
			return rateLimitErr.DomainError.Type == errType
		}
	case ErrorTypeCircuitBreaker:
		var circuitErr *CircuitBreakerError
		if errors.As(err, &circuitErr) {
			return circuitErr.DomainError.Type == errType
		}
	case ErrorTypeConfiguration:
		var configErr *ConfigurationError
		if errors.As(err, &configErr) {
			return configErr.DomainError.Type == errType
		}
	case ErrorTypeSecurity:
		var securityErr *SecurityError
		if errors.As(err, &securityErr) {
			return securityErr.DomainError.Type == errType
		}
	case ErrorTypeResourceExhausted:
		var resourceErr *ResourceExhaustedError
		if errors.As(err, &resourceErr) {
			return resourceErr.DomainError.Type == errType
		}
	case ErrorTypeDependency:
		var depErr *DependencyError
		if errors.As(err, &depErr) {
			return depErr.DomainError.Type == errType
		}
	case ErrorTypeSerialization:
		var serErr *SerializationError
		if errors.As(err, &serErr) {
			return serErr.DomainError.Type == errType
		}
	case ErrorTypeCache:
		var cacheErr *CacheError
		if errors.As(err, &cacheErr) {
			return cacheErr.DomainError.Type == errType
		}
	case ErrorTypeWorkflow:
		var workflowErr *WorkflowError
		if errors.As(err, &workflowErr) {
			return workflowErr.DomainError.Type == errType
		}
	case ErrorTypeMigration:
		var migrationErr *MigrationError
		if errors.As(err, &migrationErr) {
			return migrationErr.DomainError.Type == errType
		}
	}

	return false
}

// IsNotFoundError verifica se um erro é do tipo not found
func IsNotFoundError(err error) bool {
	return IsErrorType(err, ErrorTypeNotFound)
}

// IsValidationError verifica se um erro é do tipo validation
func IsValidationError(err error) bool {
	return IsErrorType(err, ErrorTypeValidation)
}

// IsBusinessError verifica se um erro é do tipo business
func IsBusinessError(err error) bool {
	return IsErrorType(err, ErrorTypeBusinessRule)
}

// IsAuthenticationError verifica se um erro é do tipo authentication
func IsAuthenticationError(err error) bool {
	return IsErrorType(err, ErrorTypeAuthentication)
}

// IsAuthorizationError verifica se um erro é do tipo authorization
func IsAuthorizationError(err error) bool {
	return IsErrorType(err, ErrorTypeAuthorization)
}

// IsDatabaseError verifica se um erro é um erro de banco de dados
func IsDatabaseError(err error) bool {
	var dbErr *DatabaseError
	return errors.As(err, &dbErr)
}

// IsExternalServiceError verifica se um erro é de serviço externo
func IsExternalServiceError(err error) bool {
	return IsErrorType(err, ErrorTypeExternalService)
}

// IsConflictError verifica se um erro é do tipo conflict
func IsConflictError(err error) bool {
	return IsErrorType(err, ErrorTypeConflict)
}

// IsRateLimitError verifica se um erro é do tipo rate limit
func IsRateLimitError(err error) bool {
	return IsErrorType(err, ErrorTypeRateLimit)
}

// IsCircuitBreakerError verifica se um erro é do tipo circuit breaker
func IsCircuitBreakerError(err error) bool {
	return IsErrorType(err, ErrorTypeCircuitBreaker)
}

// IsConfigurationError verifica se um erro é do tipo configuration
func IsConfigurationError(err error) bool {
	return IsErrorType(err, ErrorTypeConfiguration)
}

// IsSecurityError verifica se um erro é do tipo security
func IsSecurityError(err error) bool {
	return IsErrorType(err, ErrorTypeSecurity)
}

// IsResourceExhaustedError verifica se um erro é do tipo resource exhausted
func IsResourceExhaustedError(err error) bool {
	return IsErrorType(err, ErrorTypeResourceExhausted)
}

// IsDependencyError verifica se um erro é do tipo dependency
func IsDependencyError(err error) bool {
	return IsErrorType(err, ErrorTypeDependency)
}

// IsSerializationError verifica se um erro é do tipo serialization
func IsSerializationError(err error) bool {
	return IsErrorType(err, ErrorTypeSerialization)
}

// IsCacheError verifica se um erro é do tipo cache
func IsCacheError(err error) bool {
	return IsErrorType(err, ErrorTypeCache)
}

// IsWorkflowError verifica se um erro é do tipo workflow
func IsWorkflowError(err error) bool {
	return IsErrorType(err, ErrorTypeWorkflow)
}

// IsMigrationError verifica se um erro é do tipo migration
func IsMigrationError(err error) bool {
	return IsErrorType(err, ErrorTypeMigration)
}

// Helper functions para os novos tipos de erro

// IsInvalidSchemaError verifica se o erro é um InvalidSchemaError
func IsInvalidSchemaError(err error) bool {
	var invalidSchemaErr *InvalidSchemaError
	return errors.As(err, &invalidSchemaErr)
}

// IsUnsupportedMediaTypeError verifica se o erro é um UnsupportedMediaTypeError
func IsUnsupportedMediaTypeError(err error) bool {
	var unsupportedMediaErr *UnsupportedMediaTypeError
	return errors.As(err, &unsupportedMediaErr)
}

// IsServerError verifica se o erro é um ServerError
func IsServerError(err error) bool {
	var serverErr *ServerError
	return errors.As(err, &serverErr)
}

// IsUnprocessableEntityError verifica se o erro é um UnprocessableEntityError
func IsUnprocessableEntityError(err error) bool {
	var unprocessableErr *UnprocessableEntityError
	return errors.As(err, &unprocessableErr)
}

// IsServiceUnavailableError verifica se o erro é um ServiceUnavailableError
func IsServiceUnavailableError(err error) bool {
	var serviceUnavailableErr *ServiceUnavailableError
	return errors.As(err, &serviceUnavailableErr)
}

// NewInvalidSchemaErrorFromDetails é um helper para criar InvalidSchemaError rapidamente
func NewInvalidSchemaErrorFromDetails(schemaName string, details map[string][]string) *InvalidSchemaError {
	return NewInvalidSchemaError(fmt.Sprintf("Schema validation failed for '%s'", schemaName)).
		WithSchemaInfo(schemaName, "").
		WithSchemaDetails(details)
}

// NewUnsupportedMediaTypeErrorFromTypes é um helper para criar UnsupportedMediaTypeError rapidamente
func NewUnsupportedMediaTypeErrorFromTypes(providedType string, supportedTypes []string) *UnsupportedMediaTypeError {
	return NewUnsupportedMediaTypeError(
		fmt.Sprintf("Media type '%s' is not supported", providedType),
	).WithMediaTypeInfo(providedType, supportedTypes)
}

// NewServerErrorWithCode é um helper para criar ServerError com código específico
func NewServerErrorWithCode(code, message string, err error) *ServerError {
	return NewServerError(message, err).WithErrorCode(code)
}

// NewUnprocessableEntityErrorFromValidation é um helper para criar UnprocessableEntityError com validações
func NewUnprocessableEntityErrorFromValidation(entityType, entityID string, validationErrors map[string][]string) *UnprocessableEntityError {
	return NewUnprocessableEntityError(
		fmt.Sprintf("Entity '%s' cannot be processed", entityType),
	).WithEntityInfo(entityType, entityID).
		WithValidationErrors(validationErrors)
}

// NewServiceUnavailableErrorWithRetry é um helper para criar ServiceUnavailableError com retry info
func NewServiceUnavailableErrorWithRetry(serviceName, retryAfter string, err error) *ServiceUnavailableError {
	return NewServiceUnavailableError(
		serviceName,
		fmt.Sprintf("Service '%s' is currently unavailable", serviceName),
		err,
	).WithRetryInfo(retryAfter, "")
}

// GetErrorCode extrai o código de erro de um DomainError, se disponível
func GetErrorCode(err error) string {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code
	}
	return ""
}
