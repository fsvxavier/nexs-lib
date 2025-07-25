package santhosh

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema/interfaces"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Provider implementa validação usando santhosh-tekuri/jsonschema v6
type Provider struct {
	compiler     *jsonschema.Compiler
	schemas      map[string]*jsonschema.Schema
	errorMapping map[string]string
}

// NewProvider cria um novo provider santhosh-tekuri/jsonschema v6
func NewProvider() *Provider {
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft7) // Usar Draft 7 por padrão

	return &Provider{
		compiler:     compiler,
		schemas:      make(map[string]*jsonschema.Schema),
		errorMapping: getDefaultErrorMapping(),
	}
}

// Validate executa validação usando santhosh-tekuri/jsonschema v6
func (p *Provider) Validate(schema interface{}, data interface{}) ([]interfaces.ValidationError, error) {
	var compiledSchema *jsonschema.Schema
	var err error

	// Determina o tipo de schema e compila
	switch s := schema.(type) {
	case string:
		// Schema como string JSON
		err = p.compiler.AddResource("schema.json", s)
		if err != nil {
			return nil, fmt.Errorf("failed to add schema resource: %w", err)
		}
		compiledSchema, err = p.compiler.Compile("schema.json")
		if err != nil {
			return nil, fmt.Errorf("failed to compile schema: %w", err)
		}
	case []byte:
		// Schema como bytes - converte para string
		err = p.compiler.AddResource("schema.json", string(s))
		if err != nil {
			return nil, fmt.Errorf("failed to add schema resource: %w", err)
		}
		compiledSchema, err = p.compiler.Compile("schema.json")
		if err != nil {
			return nil, fmt.Errorf("failed to compile schema: %w", err)
		}
	case map[string]interface{}:
		// Schema como map
		err = p.compiler.AddResource("schema.json", s)
		if err != nil {
			return nil, fmt.Errorf("failed to add schema resource: %w", err)
		}
		compiledSchema, err = p.compiler.Compile("schema.json")
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
	err = compiledSchema.Validate(validationData)
	if err == nil {
		return []interfaces.ValidationError{}, nil
	}

	// Converte erros para o formato padrão
	return p.convertErrors(err), nil
}

// RegisterCustomFormat registra um formato customizado
func (p *Provider) RegisterCustomFormat(name string, validator interface{}) error {
	if fn, ok := validator.(func(interface{}) bool); ok {
		format := &jsonschema.Format{
			Name: name,
			Validate: func(v interface{}) error {
				if fn(v) {
					return nil
				}
				return fmt.Errorf("invalid format: %s", name)
			},
		}
		p.compiler.RegisterFormat(format)
		return nil
	}

	return fmt.Errorf("validator must be a func(interface{}) bool")
}

// GetName retorna o nome do provider
func (p *Provider) GetName() string {
	return "santhosh-tekuri/jsonschema-v6"
}

// convertErrors converte erros do santhosh-tekuri/jsonschema v6 para o formato padrão
func (p *Provider) convertErrors(err error) []interfaces.ValidationError {
	var errors []interfaces.ValidationError

	if validationErr, ok := err.(*jsonschema.ValidationError); ok {
		// Processa o erro principal
		mainError := interfaces.ValidationError{
			Field:       p.extractField(validationErr.InstanceLocation),
			Message:     p.getErrorMessage(validationErr.Error()),
			ErrorType:   p.mapErrorTypeFromKind(validationErr.ErrorKind),
			Description: validationErr.Error(),
		}
		errors = append(errors, mainError)

		// Processa sub-erros recursivamente
		for _, cause := range validationErr.Causes {
			subErrors := p.convertErrors(cause)
			errors = append(errors, subErrors...)
		}
	}

	return errors
}

// extractField extrai o nome do campo do caminho de localização
func (p *Provider) extractField(location []string) string {
	if len(location) == 0 {
		return "(root)"
	}

	// Junta o caminho com pontos
	return strings.Join(location, ".")
}

// mapErrorTypeFromKind mapeia ErrorKind para nosso padrão
func (p *Provider) mapErrorTypeFromKind(kind jsonschema.ErrorKind) string {
	// Como ErrorKind é uma interface, vamos usar o string da mensagem
	kindStr := fmt.Sprintf("%v", kind)

	// Mapeamento baseado em palavras-chave comuns
	if strings.Contains(kindStr, "required") {
		return "REQUIRED_ATTRIBUTE_MISSING"
	}
	if strings.Contains(kindStr, "type") {
		return "INVALID_DATA_TYPE"
	}
	if strings.Contains(kindStr, "enum") || strings.Contains(kindStr, "const") {
		return "INVALID_VALUE"
	}
	if strings.Contains(kindStr, "minLength") || strings.Contains(kindStr, "maxLength") {
		return "INVALID_LENGTH"
	}
	if strings.Contains(kindStr, "minimum") || strings.Contains(kindStr, "maximum") ||
		strings.Contains(kindStr, "exclusiveMinimum") || strings.Contains(kindStr, "exclusiveMaximum") ||
		strings.Contains(kindStr, "multipleOf") {
		return "INVALID_VALUE"
	}
	if strings.Contains(kindStr, "pattern") || strings.Contains(kindStr, "format") {
		return "INVALID_FORMAT"
	}
	if strings.Contains(kindStr, "minItems") || strings.Contains(kindStr, "maxItems") ||
		strings.Contains(kindStr, "uniqueItems") {
		return "INVALID_DATA_TYPE"
	}
	if strings.Contains(kindStr, "minProperties") || strings.Contains(kindStr, "maxProperties") ||
		strings.Contains(kindStr, "additionalProperties") {
		return "INVALID_DATA_TYPE"
	}
	if strings.Contains(kindStr, "oneOf") || strings.Contains(kindStr, "anyOf") ||
		strings.Contains(kindStr, "allOf") || strings.Contains(kindStr, "not") {
		return "INVALID_DATA_TYPE"
	}

	return "INVALID_DATA_TYPE"
}

// getErrorMessage retorna mensagem de erro baseada no tipo
func (p *Provider) getErrorMessage(originalMessage string) string {
	// Retorna a mensagem original por enquanto
	// Pode ser customizada conforme necessário
	return originalMessage
}

// ValidateFromFile valida dados usando schema de arquivo
func (p *Provider) ValidateFromFile(schemaPath string, data interface{}) ([]interfaces.ValidationError, error) {
	// Na v6, usamos file:// URLs para carregar arquivos
	schemaURL := "file://" + schemaPath

	schema, err := p.compiler.Compile(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema from file: %w", err)
	}

	// Converte data se necessário
	var validationData interface{}
	switch d := data.(type) {
	case string:
		var parsed interface{}
		if err := json.Unmarshal([]byte(d), &parsed); err == nil {
			validationData = parsed
		} else {
			validationData = d
		}
	case []byte:
		var parsed interface{}
		if err := json.Unmarshal(d, &parsed); err == nil {
			validationData = parsed
		} else {
			validationData = string(d)
		}
	default:
		validationData = d
	}

	err = schema.Validate(validationData)
	if err == nil {
		return []interfaces.ValidationError{}, nil
	}

	return p.convertErrors(err), nil
}

// CacheSchema armazena um schema compilado para reutilização
func (p *Provider) CacheSchema(name string, schema interface{}) error {
	switch s := schema.(type) {
	case string:
		err := p.compiler.AddResource(name, s)
		if err != nil {
			return err
		}
	case []byte:
		err := p.compiler.AddResource(name, string(s))
		if err != nil {
			return err
		}
	case map[string]interface{}:
		err := p.compiler.AddResource(name, s)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported schema type: %T", schema)
	}

	compiledSchema, err := p.compiler.Compile(name)
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

	// Converte data se necessário
	var validationData interface{}
	switch d := data.(type) {
	case string:
		var parsed interface{}
		if err := json.Unmarshal([]byte(d), &parsed); err == nil {
			validationData = parsed
		} else {
			validationData = d
		}
	case []byte:
		var parsed interface{}
		if err := json.Unmarshal(d, &parsed); err == nil {
			validationData = parsed
		} else {
			validationData = string(d)
		}
	default:
		validationData = d
	}

	err := schema.Validate(validationData)
	if err == nil {
		return []interfaces.ValidationError{}, nil
	}

	return p.convertErrors(err), nil
}

// SetErrorMapping define mapeamento customizado de erros
func (p *Provider) SetErrorMapping(mapping map[string]string) {
	p.errorMapping = mapping
}

// getDefaultErrorMapping retorna o mapeamento padrão de erros
func getDefaultErrorMapping() map[string]string {
	return map[string]string{
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
	}
}
