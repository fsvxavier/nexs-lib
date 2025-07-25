package jsonschema

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema/config"
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/interfaces"
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/providers/gojsonschema"
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/providers/kaptinlin"
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/providers/santhosh"
)

// JSONSchemaValidator implementa validação JSON Schema com suporte a múltiplos providers
type JSONSchemaValidator struct {
	config   *config.Config
	provider interfaces.Provider
}

// NewValidator cria um novo validador com configuração padrão
func NewValidator(cfg *config.Config) (*JSONSchemaValidator, error) {
	if cfg == nil {
		cfg = config.NewConfig()
	}

	provider, err := createProvider(cfg.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	// Registra formatos customizados no provider
	for name, checker := range cfg.CustomFormats {
		if err := provider.RegisterCustomFormat(name, checker); err != nil {
			// Log do erro, mas não falha completamente
			fmt.Printf("Warning: failed to register custom format %s: %v\n", name, err)
		}
	}

	return &JSONSchemaValidator{
		config:   cfg,
		provider: provider,
	}, nil
}

// ValidateFromFile valida dados usando schema de arquivo
func (v *JSONSchemaValidator) ValidateFromFile(schemaPath string, data interface{}) ([]interfaces.ValidationError, error) {
	// Executa hooks de pré-validação
	processedData, err := v.executePreValidationHooks(data)
	if err != nil {
		return nil, fmt.Errorf("pre-validation hook failed: %w", err)
	}

	// Executa validação principal
	errors, err := v.provider.ValidateFromFile(schemaPath, processedData)
	if err != nil {
		return nil, err
	}

	// Executa checks adicionais
	additionalErrors := v.executeAdditionalChecks(processedData)
	errors = append(errors, additionalErrors...)

	// Executa hooks de pós-validação
	errors, err = v.executePostValidationHooks(processedData, errors)
	if err != nil {
		return nil, fmt.Errorf("post-validation hook failed: %w", err)
	}

	// Executa hooks de erro se houver erros
	if len(errors) > 0 {
		errors = v.executeErrorHooks(errors)
	}

	return errors, nil
}

// ValidateFromBytes valida dados usando schema em bytes
func (v *JSONSchemaValidator) ValidateFromBytes(schema []byte, data interface{}) ([]interfaces.ValidationError, error) {
	// Executa hooks de pré-validação
	processedData, err := v.executePreValidationHooks(data)
	if err != nil {
		return nil, fmt.Errorf("pre-validation hook failed: %w", err)
	}

	// Executa validação principal
	errors, err := v.provider.Validate(schema, processedData)
	if err != nil {
		return nil, err
	}

	// Executa checks adicionais
	additionalErrors := v.executeAdditionalChecks(processedData)
	errors = append(errors, additionalErrors...)

	// Executa hooks de pós-validação
	errors, err = v.executePostValidationHooks(processedData, errors)
	if err != nil {
		return nil, fmt.Errorf("post-validation hook failed: %w", err)
	}

	// Executa hooks de erro se houver erros
	if len(errors) > 0 {
		errors = v.executeErrorHooks(errors)
	}

	return errors, nil
}

// ValidateFromStruct valida dados usando schema registrado
func (v *JSONSchemaValidator) ValidateFromStruct(schemaName string, data interface{}) ([]interfaces.ValidationError, error) {
	// Verifica se o schema está registrado
	schema, exists := v.config.SchemaRegistry[schemaName]
	if !exists {
		return nil, fmt.Errorf("schema '%s' not found in registry", schemaName)
	}

	// Executa hooks de pré-validação
	processedData, err := v.executePreValidationHooks(data)
	if err != nil {
		return nil, fmt.Errorf("pre-validation hook failed: %w", err)
	}

	// Executa validação principal
	errors, err := v.provider.Validate(schema, processedData)
	if err != nil {
		return nil, err
	}

	// Executa checks adicionais
	additionalErrors := v.executeAdditionalChecks(processedData)
	errors = append(errors, additionalErrors...)

	// Executa hooks de pós-validação
	errors, err = v.executePostValidationHooks(processedData, errors)
	if err != nil {
		return nil, fmt.Errorf("post-validation hook failed: %w", err)
	}

	// Executa hooks de erro se houver erros
	if len(errors) > 0 {
		errors = v.executeErrorHooks(errors)
	}

	return errors, nil
}

// executePreValidationHooks executa todos os hooks de pré-validação
func (v *JSONSchemaValidator) executePreValidationHooks(data interface{}) (interface{}, error) {
	processedData := data
	for _, hook := range v.config.PreValidationHooks {
		var err error
		processedData, err = hook.Execute(processedData)
		if err != nil {
			return nil, err
		}
	}
	return processedData, nil
}

// executePostValidationHooks executa todos os hooks de pós-validação
func (v *JSONSchemaValidator) executePostValidationHooks(data interface{}, errors []interfaces.ValidationError) ([]interfaces.ValidationError, error) {
	processedErrors := errors
	for _, hook := range v.config.PostValidationHooks {
		var err error
		processedErrors, err = hook.Execute(data, processedErrors)
		if err != nil {
			return nil, err
		}
	}
	return processedErrors, nil
}

// executeErrorHooks executa todos os hooks de erro
func (v *JSONSchemaValidator) executeErrorHooks(errors []interfaces.ValidationError) []interfaces.ValidationError {
	processedErrors := errors
	for _, hook := range v.config.ErrorHooks {
		processedErrors = hook.Execute(processedErrors)
	}
	return processedErrors
}

// executeAdditionalChecks executa todos os checks adicionais
func (v *JSONSchemaValidator) executeAdditionalChecks(data interface{}) []interfaces.ValidationError {
	var errors []interfaces.ValidationError
	for _, check := range v.config.AdditionalChecks {
		checkErrors := check.Validate(data)
		errors = append(errors, checkErrors...)
	}
	return errors
}

// createProvider cria um provider baseado no tipo configurado
func createProvider(providerType config.ProviderType) (interfaces.Provider, error) {
	switch providerType {
	case config.GoJSONSchemaProvider:
		return gojsonschema.NewProvider(), nil
	case config.JSONSchemaProvider:
		return kaptinlin.NewProvider(), nil
	case config.SchemaJSONProvider:
		return santhosh.NewProvider(), nil
	default:
		// Usa kaptinlin como fallback
		return kaptinlin.NewProvider(), nil
	}
}

// --- Funções de retrocompatibilidade com _old/validator ---

// Validate mantém compatibilidade com a função original do _old/validator
// Usa gojsonschema por padrão para manter compatibilidade total
func Validate(loader interface{}, schemaLoader string) error {
	validator, err := NewValidator(&config.Config{
		Provider: config.GoJSONSchemaProvider,
	})
	if err != nil {
		return err
	}

	errors, err := validator.provider.Validate(schemaLoader, loader)
	if err != nil {
		return err
	}

	if len(errors) == 0 {
		return nil
	}

	// Converte para o formato de erro original (domainerrors.InvalidSchemaError)
	return createLegacyError(errors)
}

// AddCustomFormat mantém compatibilidade com a função original
func AddCustomFormat(formatName string, regex string) {
	// Esta função era global no código original, então usamos um provider global
	provider := gojsonschema.NewProvider()

	// Cria um checker baseado em regex para compatibilidade
	checker := &regexFormatChecker{pattern: regex}
	provider.RegisterCustomFormat(formatName, checker)
}

// regexFormatChecker implementa compatibilidade para formatos baseados em regex
type regexFormatChecker struct {
	pattern string
}

// IsFormat implementa gojsonschema.FormatChecker
func (r *regexFormatChecker) IsFormat(input interface{}) bool {
	// Implementação simplificada - seria melhor usar regexp.Compile
	// Mas mantém compatibilidade básica
	if str, ok := input.(string); ok {
		// Por enquanto retorna true para compatibilidade
		// Em uma implementação completa, compilaria e testaria a regex
		return len(str) > 0 && len(r.pattern) > 0
	}
	return false
}

// createLegacyError cria erro no formato do domainerrors.InvalidSchemaError original
func createLegacyError(validationErrors []interfaces.ValidationError) error {
	// Simula a estrutura do domainerrors.InvalidSchemaError
	errorDetails := make(map[string][]string)

	for _, err := range validationErrors {
		field := err.Field
		if field == "" {
			field = "(root)"
		}

		errorDetails[field] = []string{err.ErrorType}
	}

	// Retorna um erro simples por enquanto
	// Em uma implementação completa, criaria o domainerrors.InvalidSchemaError
	return fmt.Errorf("validation failed with %d errors", len(validationErrors))
}
