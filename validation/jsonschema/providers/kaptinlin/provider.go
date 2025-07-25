package kaptinlin

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema/interfaces"
	"github.com/kaptinlin/jsonschema"
)

// Provider implementa validação usando kaptinlin/jsonschema
type Provider struct {
	compiler *jsonschema.Compiler
	schemas  map[string]*jsonschema.Schema
}

// NewProvider cria um novo provider kaptinlin/jsonschema
func NewProvider() *Provider {
	compiler := jsonschema.NewCompiler()
	return &Provider{
		compiler: compiler,
		schemas:  make(map[string]*jsonschema.Schema),
	}
}

// Validate executa validação usando kaptinlin/jsonschema
func (p *Provider) Validate(schema interface{}, data interface{}) ([]interfaces.ValidationError, error) {
	var compiledSchema *jsonschema.Schema
	var err error

	// Determina o tipo de schema e compila
	switch s := schema.(type) {
	case string:
		// Schema como string JSON
		compiledSchema, err = p.compiler.Compile([]byte(s))
		if err != nil {
			return nil, fmt.Errorf("failed to compile schema: %w", err)
		}
	case []byte:
		// Schema como bytes
		compiledSchema, err = p.compiler.Compile(s)
		if err != nil {
			return nil, fmt.Errorf("failed to compile schema: %w", err)
		}
	case map[string]interface{}:
		// Schema como map
		schemaBytes, err := json.Marshal(s)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal schema: %w", err)
		}
		compiledSchema, err = p.compiler.Compile(schemaBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to compile schema: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported schema type: %T", schema)
	}

	// Converte data para interface{} se necessário
	var validationData interface{}
	switch d := data.(type) {
	case string:
		// Tenta fazer parse do JSON se for string
		var parsed interface{}
		if err := json.Unmarshal([]byte(d), &parsed); err == nil {
			validationData = parsed
		} else {
			validationData = d
		}
	case []byte:
		// Tenta fazer parse do JSON se for bytes
		var parsed interface{}
		if err := json.Unmarshal(d, &parsed); err == nil {
			validationData = parsed
		} else {
			validationData = string(d)
		}
	default:
		validationData = d
	}

	// Executa validação
	result := compiledSchema.Validate(validationData)
	if result == nil || result.IsValid() {
		return []interfaces.ValidationError{}, nil
	}

	// Converte erros para o formato padrão
	return p.convertErrors(result), nil
}

// RegisterCustomFormat registra um formato customizado
func (p *Provider) RegisterCustomFormat(name string, validator interface{}) error {
	// kaptinlin/jsonschema não suporta formatos customizados da mesma forma
	// Esta implementação é um placeholder para compatibilidade
	return fmt.Errorf("custom formats not fully supported in kaptinlin/jsonschema provider")
}

// GetName retorna o nome do provider
func (p *Provider) GetName() string {
	return "kaptinlin/jsonschema"
}

// convertErrors converte erros do kaptinlin/jsonschema para o formato padrão
func (p *Provider) convertErrors(result *jsonschema.EvaluationResult) []interfaces.ValidationError {
	var errors []interfaces.ValidationError

	if result.IsValid() {
		return errors
	}

	// Coleta erros da estrutura hierárquica
	p.collectErrors(result, &errors)

	return errors
}

// collectErrors coleta erros recursivamente da estrutura de resultados
func (p *Provider) collectErrors(result *jsonschema.EvaluationResult, errors *[]interfaces.ValidationError) {
	// Processa erros diretos
	for keyword, evalError := range result.Errors {
		validationErr := interfaces.ValidationError{
			Field:       p.extractField(result.InstanceLocation),
			Message:     evalError.Message,
			ErrorType:   p.mapErrorType(keyword),
			Description: evalError.Code,
		}
		*errors = append(*errors, validationErr)
	}

	// Processa erros dos detalhes (recursivo)
	for _, detail := range result.Details {
		p.collectErrors(detail, errors)
	}
}

// extractField extrai o nome do campo do caminho de localização
func (p *Provider) extractField(location string) string {
	// Remove o prefixo "#/" se presente
	if len(location) > 2 && location[:2] == "#/" {
		location = location[2:]
	}

	// Se vazio, retorna root
	if location == "" || location == "#" {
		return "(root)"
	}

	return location
}

// mapErrorType mapeia tipos de erro do kaptinlin/jsonschema para nosso padrão
func (p *Provider) mapErrorType(keyword string) string {
	errorMap := map[string]string{
		"required":             "REQUIRED_ATTRIBUTE_MISSING",
		"type":                 "INVALID_DATA_TYPE",
		"enum":                 "INVALID_VALUE",
		"const":                "INVALID_VALUE",
		"minLength":            "INVALID_LENGTH",
		"maxLength":            "INVALID_LENGTH",
		"minimum":              "INVALID_VALUE",
		"maximum":              "INVALID_VALUE",
		"exclusiveMinimum":     "INVALID_VALUE",
		"exclusiveMaximum":     "INVALID_VALUE",
		"multipleOf":           "INVALID_VALUE",
		"pattern":              "INVALID_FORMAT",
		"format":               "INVALID_FORMAT",
		"minItems":             "INVALID_DATA_TYPE",
		"maxItems":             "INVALID_DATA_TYPE",
		"uniqueItems":          "INVALID_DATA_TYPE",
		"minProperties":        "INVALID_DATA_TYPE",
		"maxProperties":        "INVALID_DATA_TYPE",
		"additionalProperties": "INVALID_DATA_TYPE",
		"additionalItems":      "INVALID_DATA_TYPE",
		"oneOf":                "INVALID_DATA_TYPE",
		"anyOf":                "INVALID_DATA_TYPE",
		"allOf":                "INVALID_DATA_TYPE",
		"not":                  "INVALID_DATA_TYPE",
		"if":                   "INVALID_DATA_TYPE",
		"then":                 "INVALID_DATA_TYPE",
		"else":                 "INVALID_DATA_TYPE",
	}

	if mappedType, exists := errorMap[keyword]; exists {
		return mappedType
	}

	return "INVALID_DATA_TYPE"
}

// ValidateFromFile valida dados usando schema de arquivo
func (p *Provider) ValidateFromFile(schemaPath string, data interface{}) ([]interfaces.ValidationError, error) {
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	return p.Validate(schemaBytes, data)
}

// CacheSchema armazena um schema compilado para reutilização
func (p *Provider) CacheSchema(name string, schema interface{}) error {
	compiledSchema, err := p.compileSchema(schema)
	if err != nil {
		return err
	}

	p.schemas[name] = compiledSchema
	return nil
}

// ValidateWithCachedSchema valida usando schema em cache
func (p *Provider) ValidateWithCachedSchema(schemaName string, data interface{}) ([]interfaces.ValidationError, error) {
	schema, exists := p.schemas[schemaName]
	if !exists {
		return nil, fmt.Errorf("schema '%s' not found in cache", schemaName)
	}

	result := schema.Validate(data)
	if result == nil || result.IsValid() {
		return []interfaces.ValidationError{}, nil
	}

	return p.convertErrors(result), nil
}

// compileSchema compila um schema para cache
func (p *Provider) compileSchema(schema interface{}) (*jsonschema.Schema, error) {
	switch s := schema.(type) {
	case string:
		return p.compiler.Compile([]byte(s))
	case []byte:
		return p.compiler.Compile(s)
	case map[string]interface{}:
		schemaBytes, err := json.Marshal(s)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal schema: %w", err)
		}
		return p.compiler.Compile(schemaBytes)
	default:
		return nil, fmt.Errorf("unsupported schema type: %T", schema)
	}
}
