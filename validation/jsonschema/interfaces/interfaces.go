package interfaces

// ValidationError representa um erro de validação com detalhes específicos
type ValidationError struct {
	Field       string      `json:"field"`
	Message     string      `json:"message"`
	ErrorType   string      `json:"error_type"`
	Value       interface{} `json:"value,omitempty"`
	Description string      `json:"description,omitempty"`
}

// Error implementa a interface error
func (ve ValidationError) Error() string {
	return ve.Message
}

// Validator define o contrato principal para validação JSON Schema
type Validator interface {
	ValidateFromFile(schemaPath string, data interface{}) ([]ValidationError, error)
	ValidateFromBytes(schema []byte, data interface{}) ([]ValidationError, error)
	ValidateFromStruct(schemaName string, data interface{}) ([]ValidationError, error)
}

// Provider define o contrato para implementações de validação específicas
type Provider interface {
	Validate(schema interface{}, data interface{}) ([]ValidationError, error)
	ValidateFromFile(schemaPath string, data interface{}) ([]ValidationError, error)
	RegisterCustomFormat(name string, validator interface{}) error
	GetName() string
}

// PreValidationHook executa lógica antes da validação
type PreValidationHook interface {
	Execute(data interface{}) (interface{}, error)
}

// PostValidationHook executa lógica após a validação
type PostValidationHook interface {
	Execute(data interface{}, errors []ValidationError) ([]ValidationError, error)
}

// ErrorHook executa lógica quando há erros de validação
type ErrorHook interface {
	Execute(errors []ValidationError) []ValidationError
}

// Check define contratos para validações customizadas adicionais
type Check interface {
	Validate(data interface{}) []ValidationError
	GetName() string
}

// FormatChecker define contratos para validadores de formato customizados
type FormatChecker interface {
	IsFormat(input interface{}) bool
}
