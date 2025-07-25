package config

import (
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/interfaces"
)

// ProviderType define os tipos de providers disponíveis
type ProviderType string

const (
	GoJSONSchemaProvider ProviderType = "gojsonschema"
	JSONSchemaProvider   ProviderType = "jsonschema"
	SchemaJSONProvider   ProviderType = "schemajson"
)

// Config define a configuração para o validador JSON Schema
type Config struct {
	// Provider especifica qual engine de validação usar
	Provider ProviderType `json:"provider"`

	// PreValidationHooks são executados antes da validação
	PreValidationHooks []interfaces.PreValidationHook `json:"-"`

	// PostValidationHooks são executados após a validação
	PostValidationHooks []interfaces.PostValidationHook `json:"-"`

	// ErrorHooks são executados quando há erros de validação
	ErrorHooks []interfaces.ErrorHook `json:"-"`

	// AdditionalChecks são validações customizadas executadas além do JSON Schema
	AdditionalChecks []interfaces.Check `json:"-"`

	// CustomFormats mapeia nomes de formatos para validadores customizados
	CustomFormats map[string]interfaces.FormatChecker `json:"-"`

	// StrictMode define se deve falhar em propriedades adicionais não permitidas
	StrictMode bool `json:"strict_mode"`

	// SchemaRegistry armazena schemas pré-carregados por nome
	SchemaRegistry map[string]interface{} `json:"-"`

	// ErrorMapping mapeia tipos de erro para mensagens customizadas
	ErrorMapping map[string]string `json:"error_mapping"`
}

// NewConfig cria uma nova configuração com valores padrão
func NewConfig() *Config {
	return &Config{
		Provider:            JSONSchemaProvider,
		PreValidationHooks:  make([]interfaces.PreValidationHook, 0),
		PostValidationHooks: make([]interfaces.PostValidationHook, 0),
		ErrorHooks:          make([]interfaces.ErrorHook, 0),
		AdditionalChecks:    make([]interfaces.Check, 0),
		CustomFormats:       make(map[string]interfaces.FormatChecker),
		StrictMode:          false,
		SchemaRegistry:      make(map[string]interface{}),
		ErrorMapping:        getDefaultErrorMapping(),
	}
}

// WithProvider define o provider a ser usado
func (c *Config) WithProvider(provider ProviderType) *Config {
	c.Provider = provider
	return c
}

// WithStrictMode habilita/desabilita o modo estrito
func (c *Config) WithStrictMode(strict bool) *Config {
	c.StrictMode = strict
	return c
}

// AddPreValidationHook adiciona um hook de pré-validação
func (c *Config) AddPreValidationHook(hook interfaces.PreValidationHook) *Config {
	c.PreValidationHooks = append(c.PreValidationHooks, hook)
	return c
}

// AddPostValidationHook adiciona um hook de pós-validação
func (c *Config) AddPostValidationHook(hook interfaces.PostValidationHook) *Config {
	c.PostValidationHooks = append(c.PostValidationHooks, hook)
	return c
}

// AddErrorHook adiciona um hook de erro
func (c *Config) AddErrorHook(hook interfaces.ErrorHook) *Config {
	c.ErrorHooks = append(c.ErrorHooks, hook)
	return c
}

// AddCheck adiciona uma validação customizada
func (c *Config) AddCheck(check interfaces.Check) *Config {
	c.AdditionalChecks = append(c.AdditionalChecks, check)
	return c
}

// AddCustomFormat adiciona um formato customizado
func (c *Config) AddCustomFormat(name string, checker interfaces.FormatChecker) *Config {
	c.CustomFormats[name] = checker
	return c
}

// RegisterSchema registra um schema para uso posterior
func (c *Config) RegisterSchema(name string, schema interface{}) *Config {
	c.SchemaRegistry[name] = schema
	return c
}

// SetErrorMapping define mapeamento customizado de erros
func (c *Config) SetErrorMapping(mapping map[string]string) *Config {
	c.ErrorMapping = mapping
	return c
}

// getDefaultErrorMapping retorna o mapeamento padrão de erros
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
