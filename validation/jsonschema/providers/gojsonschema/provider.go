package gojsonschema

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema/interfaces"
	"github.com/xeipuuv/gojsonschema"
)

// Provider implementa validação usando xeipuuv/gojsonschema (compatibilidade com _old/validator)
type Provider struct {
	customFormats map[string]gojsonschema.FormatChecker
	errorMapping  map[string]string
}

// NewProvider cria um novo provider gojsonschema
func NewProvider() *Provider {
	return &Provider{
		customFormats: make(map[string]gojsonschema.FormatChecker),
		errorMapping:  getDefaultErrorMapping(),
	}
}

// Validate executa validação usando xeipuuv/gojsonschema
func (p *Provider) Validate(schema interface{}, data interface{}) ([]interfaces.ValidationError, error) {
	var schemaLoader gojsonschema.JSONLoader
	var dataLoader gojsonschema.JSONLoader

	// Configura schema loader
	switch s := schema.(type) {
	case string:
		schemaLoader = gojsonschema.NewStringLoader(s)
	case []byte:
		schemaLoader = gojsonschema.NewBytesLoader(s)
	case map[string]interface{}:
		schemaLoader = gojsonschema.NewGoLoader(s)
	default:
		return nil, fmt.Errorf("unsupported schema type: %T", schema)
	}

	// Configura data loader
	switch d := data.(type) {
	case string:
		// Tenta fazer parse do JSON se for string
		var parsed interface{}
		if err := json.Unmarshal([]byte(d), &parsed); err == nil {
			dataLoader = gojsonschema.NewGoLoader(parsed)
		} else {
			dataLoader = gojsonschema.NewStringLoader(d)
		}
	case []byte:
		dataLoader = gojsonschema.NewBytesLoader(d)
	default:
		dataLoader = gojsonschema.NewGoLoader(d)
	}

	// Executa validação
	result, err := gojsonschema.Validate(schemaLoader, dataLoader)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if result.Valid() {
		return []interfaces.ValidationError{}, nil
	}

	// Converte erros para o formato padrão
	return p.convertErrors(result), nil
}

// RegisterCustomFormat registra um formato customizado
func (p *Provider) RegisterCustomFormat(name string, validator interface{}) error {
	if checker, ok := validator.(gojsonschema.FormatChecker); ok {
		gojsonschema.FormatCheckers.Add(name, checker)
		p.customFormats[name] = checker
		return nil
	}

	// Se não for um FormatChecker, tenta criar um a partir de uma função
	if fn, ok := validator.(func(interface{}) bool); ok {
		checker := &customFormatChecker{validateFunc: fn}
		gojsonschema.FormatCheckers.Add(name, checker)
		p.customFormats[name] = checker
		return nil
	}

	return fmt.Errorf("validator must implement gojsonschema.FormatChecker or be a func(interface{}) bool")
}

// GetName retorna o nome do provider
func (p *Provider) GetName() string {
	return "xeipuuv/gojsonschema"
}

// convertErrors converte erros do gojsonschema para o formato padrão
func (p *Provider) convertErrors(result *gojsonschema.Result) []interfaces.ValidationError {
	var errors []interfaces.ValidationError

	for _, err := range result.Errors() {
		field := p.extractField(err)
		errorType := p.mapErrorType(err.Type())

		validationErr := interfaces.ValidationError{
			Field:       field,
			Message:     p.getErrorMessage(errorType),
			ErrorType:   errorType,
			Value:       err.Value(),
			Description: err.Description(),
		}

		errors = append(errors, validationErr)
	}

	return errors
}

// extractField extrai o nome do campo do erro (compatível com _old/validator)
func (p *Provider) extractField(err gojsonschema.ResultError) string {
	field := err.Field()
	if field == "(root)" {
		// Tenta extrair property dos detalhes como no código original
		if property, found := err.Details()["property"]; found {
			if propertyStr, ok := property.(string); ok {
				return propertyStr
			}
		}
	}
	return field
}

// mapErrorType mapeia tipos de erro (compatível com _old/validator)
func (p *Provider) mapErrorType(errorType string) string {
	if mappedType, exists := p.errorMapping[errorType]; exists {
		return mappedType
	}
	return "INVALID_DATA_TYPE"
}

// getErrorMessage retorna a mensagem de erro mapeada
func (p *Provider) getErrorMessage(errorType string) string {
	switch errorType {
	case "REQUIRED_ATTRIBUTE_MISSING":
		return "Required attribute is missing"
	case "INVALID_DATA_TYPE":
		return "Invalid data type"
	case "INVALID_VALUE":
		return "Invalid value"
	case "INVALID_FORMAT":
		return "Invalid format"
	case "INVALID_LENGTH":
		return "Invalid length"
	default:
		return "Validation error"
	}
}

// ValidateFromFile valida dados usando schema de arquivo
func (p *Provider) ValidateFromFile(schemaPath string, data interface{}) ([]interfaces.ValidationError, error) {
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	return p.Validate(schemaBytes, data)
}

// SetErrorMapping define mapeamento customizado de erros
func (p *Provider) SetErrorMapping(mapping map[string]string) {
	p.errorMapping = mapping
}

// customFormatChecker implementa gojsonschema.FormatChecker para funções simples
type customFormatChecker struct {
	validateFunc func(interface{}) bool
}

// IsFormat implementa gojsonschema.FormatChecker
func (c *customFormatChecker) IsFormat(input interface{}) bool {
	return c.validateFunc(input)
}

// getDefaultErrorMapping retorna o mapeamento padrão de erros (compatível com _old/validator)
func getDefaultErrorMapping() map[string]string {
	return map[string]string{
		"required":                        "REQUIRED_ATTRIBUTE_MISSING",
		"invalid_type":                    "INVALID_DATA_TYPE",
		"number_any_of":                   "INVALID_DATA_TYPE",
		"number_one_of":                   "INVALID_DATA_TYPE",
		"number_all_of":                   "INVALID_DATA_TYPE",
		"number_not":                      "INVALID_DATA_TYPE",
		"missing_dependency":              "INVALID_DATA_TYPE",
		"internal":                        "INVALID_DATA_TYPE",
		"const":                           "INVALID_DATA_TYPE",
		"enum":                            "INVALID_VALUE",
		"array_no_additional_items":       "INVALID_DATA_TYPE",
		"array_min_items":                 "INVALID_DATA_TYPE",
		"array_max_items":                 "INVALID_DATA_TYPE",
		"unique":                          "INVALID_DATA_TYPE",
		"contains":                        "INVALID_DATA_TYPE",
		"array_min_properties":            "INVALID_DATA_TYPE",
		"array_max_properties":            "INVALID_DATA_TYPE",
		"additional_property_not_allowed": "INVALID_DATA_TYPE",
		"invalid_property_pattern":        "INVALID_DATA_TYPE",
		"invalid_property_name":           "INVALID_DATA_TYPE",
		"string_gte":                      "INVALID_LENGTH",
		"string_lte":                      "INVALID_LENGTH",
		"pattern":                         "INVALID_DATA_TYPE",
		"multiple_of":                     "INVALID_DATA_TYPE",
		"number_gte":                      "INVALID_VALUE",
		"number_gt":                       "INVALID_VALUE",
		"number_lte":                      "INVALID_VALUE",
		"number_lt":                       "INVALID_VALUE",
		"condition_then":                  "INVALID_DATA_TYPE",
		"condition_else":                  "INVALID_DATA_TYPE",
		"format":                          "INVALID_FORMAT",
	}
}
