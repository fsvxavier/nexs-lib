// Package config fornece utilitários para carregamento de configurações
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fsvxavier/nexs-lib/observability/tracer/interfaces"
)

// ConfigOption define uma função para configurar opções
type ConfigOption func(*interfaces.Config)

// WithServiceName define o nome do serviço
func WithServiceName(serviceName string) ConfigOption {
	return func(c *interfaces.Config) {
		c.ServiceName = serviceName
	}
}

// WithEnvironment define o ambiente de execução
func WithEnvironment(environment string) ConfigOption {
	return func(c *interfaces.Config) {
		c.Environment = environment
	}
}

// WithExporterType define o tipo de exporter
func WithExporterType(exporterType string) ConfigOption {
	return func(c *interfaces.Config) {
		c.ExporterType = exporterType
	}
}

// WithEndpoint define o endpoint do trace collector
func WithEndpoint(endpoint string) ConfigOption {
	return func(c *interfaces.Config) {
		c.Endpoint = endpoint
	}
}

// WithAPIKey define a API key para autenticação
func WithAPIKey(apiKey string) ConfigOption {
	return func(c *interfaces.Config) {
		c.APIKey = apiKey
	}
}

// WithLicenseKey define a license key para New Relic
func WithLicenseKey(licenseKey string) ConfigOption {
	return func(c *interfaces.Config) {
		c.LicenseKey = licenseKey
	}
}

// WithVersion define a versão da aplicação
func WithVersion(version string) ConfigOption {
	return func(c *interfaces.Config) {
		c.Version = version
	}
}

// WithSamplingRatio define a proporção de traces que serão coletados
func WithSamplingRatio(ratio float64) ConfigOption {
	return func(c *interfaces.Config) {
		c.SamplingRatio = ratio
	}
}

// WithPropagators define os propagadores a serem utilizados
func WithPropagators(propagators ...string) ConfigOption {
	return func(c *interfaces.Config) {
		c.Propagators = propagators
	}
}

// WithHeaders define cabeçalhos adicionais para envio de traces
func WithHeaders(headers map[string]string) ConfigOption {
	return func(c *interfaces.Config) {
		if c.Headers == nil {
			c.Headers = make(map[string]string)
		}
		for k, v := range headers {
			c.Headers[k] = v
		}
	}
}

// WithHeader adiciona um cabeçalho específico
func WithHeader(key, value string) ConfigOption {
	return func(c *interfaces.Config) {
		if c.Headers == nil {
			c.Headers = make(map[string]string)
		}
		c.Headers[key] = value
	}
}

// WithAttributes define atributos adicionais para os traces
func WithAttributes(attributes map[string]string) ConfigOption {
	return func(c *interfaces.Config) {
		if c.Attributes == nil {
			c.Attributes = make(map[string]string)
		}
		for k, v := range attributes {
			c.Attributes[k] = v
		}
	}
}

// WithAttribute adiciona um atributo específico
func WithAttribute(key, value string) ConfigOption {
	return func(c *interfaces.Config) {
		if c.Attributes == nil {
			c.Attributes = make(map[string]string)
		}
		c.Attributes[key] = value
	}
}

// WithInsecure define se deve usar conexão insegura
func WithInsecure(insecure bool) ConfigOption {
	return func(c *interfaces.Config) {
		c.Insecure = insecure
	}
}

// NewConfig cria uma nova configuração aplicando as opções fornecidas
func NewConfig(opts ...ConfigOption) interfaces.Config {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(&config)
	}
	return config
}

// NewConfigFromEnv cria uma configuração a partir de variáveis de ambiente e aplica as opções
func NewConfigFromEnv(opts ...ConfigOption) interfaces.Config {
	config := LoadFromEnv()
	for _, opt := range opts {
		opt(&config)
	}
	return config
}

// DefaultConfig retorna uma configuração padrão
func DefaultConfig() interfaces.Config {
	return interfaces.Config{
		ServiceName:   "unknown-service",
		Environment:   "development",
		ExporterType:  "opentelemetry",
		SamplingRatio: 1.0,
		Propagators:   []string{"tracecontext", "b3"},
		Headers:       make(map[string]string),
		Attributes:    make(map[string]string),
		Insecure:      false,
		Version:       "1.0.0",
	}
}

// LoadFromEnv carrega a configuração das variáveis de ambiente
func LoadFromEnv() interfaces.Config {
	config := DefaultConfig()

	if serviceName := os.Getenv("TRACER_SERVICE_NAME"); serviceName != "" {
		config.ServiceName = serviceName
	}

	if environment := os.Getenv("TRACER_ENVIRONMENT"); environment != "" {
		config.Environment = environment
	}

	if exporterType := os.Getenv("TRACER_EXPORTER_TYPE"); exporterType != "" {
		config.ExporterType = exporterType
	}

	if endpoint := os.Getenv("TRACER_ENDPOINT"); endpoint != "" {
		config.Endpoint = endpoint
	}

	if apiKey := os.Getenv("TRACER_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}

	if licenseKey := os.Getenv("TRACER_LICENSE_KEY"); licenseKey != "" {
		config.LicenseKey = licenseKey
	}

	if version := os.Getenv("TRACER_VERSION"); version != "" {
		config.Version = version
	}

	if samplingRatio := os.Getenv("TRACER_SAMPLING_RATIO"); samplingRatio != "" {
		if ratio, err := strconv.ParseFloat(samplingRatio, 64); err == nil {
			config.SamplingRatio = ratio
		}
	}

	if insecure := os.Getenv("TRACER_INSECURE"); insecure != "" {
		if insecureBool, err := strconv.ParseBool(insecure); err == nil {
			config.Insecure = insecureBool
		}
	}

	if propagators := os.Getenv("TRACER_PROPAGATORS"); propagators != "" {
		config.Propagators = strings.Split(propagators, ",")
		for i, p := range config.Propagators {
			config.Propagators[i] = strings.TrimSpace(p)
		}
	}

	// Load headers from environment variables with TRACER_HEADER_ prefix
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "TRACER_HEADER_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				headerName := strings.ToLower(strings.TrimPrefix(parts[0], "TRACER_HEADER_"))
				headerName = strings.ReplaceAll(headerName, "_", "-")
				config.Headers[headerName] = parts[1]
			}
		}
	}

	// Load attributes from environment variables with TRACER_ATTR_ prefix
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "TRACER_ATTR_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				attrName := strings.ToLower(strings.TrimPrefix(parts[0], "TRACER_ATTR_"))
				config.Attributes[attrName] = parts[1]
			}
		}
	}

	return config
}

// Validate valida a configuração
func Validate(config interfaces.Config) error {
	if config.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}

	if config.ExporterType == "" {
		return fmt.Errorf("exporter type is required")
	}

	supportedTypes := []string{"datadog", "grafana", "newrelic", "opentelemetry"}
	found := false
	for _, t := range supportedTypes {
		if t == config.ExporterType {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("unsupported exporter type: %s, supported types: %v", config.ExporterType, supportedTypes)
	}

	if config.SamplingRatio < 0 || config.SamplingRatio > 1 {
		return fmt.Errorf("sampling ratio must be between 0 and 1, got: %f", config.SamplingRatio)
	}

	// Validate specific requirements per exporter type
	switch config.ExporterType {
	case "datadog":
		if config.APIKey == "" {
			return fmt.Errorf("API key is required for Datadog exporter")
		}
	case "newrelic":
		if config.LicenseKey == "" {
			return fmt.Errorf("License key is required for New Relic exporter")
		}
	case "grafana", "opentelemetry":
		if config.Endpoint == "" {
			return fmt.Errorf("endpoint is required for %s exporter", config.ExporterType)
		}
	}

	return nil
}

// MergeConfigs mescla duas configurações, priorizando os valores de override
func MergeConfigs(base, override interfaces.Config) interfaces.Config {
	result := base

	if override.ServiceName != "" {
		result.ServiceName = override.ServiceName
	}
	if override.Environment != "" {
		result.Environment = override.Environment
	}
	if override.ExporterType != "" {
		result.ExporterType = override.ExporterType
	}
	if override.Endpoint != "" {
		result.Endpoint = override.Endpoint
	}
	if override.APIKey != "" {
		result.APIKey = override.APIKey
	}
	if override.LicenseKey != "" {
		result.LicenseKey = override.LicenseKey
	}
	if override.Version != "" {
		result.Version = override.Version
	}
	if override.SamplingRatio != 0 {
		result.SamplingRatio = override.SamplingRatio
	}
	if len(override.Propagators) > 0 {
		result.Propagators = override.Propagators
	}

	// Merge headers
	if result.Headers == nil {
		result.Headers = make(map[string]string)
	}
	for k, v := range override.Headers {
		result.Headers[k] = v
	}

	// Merge attributes
	if result.Attributes == nil {
		result.Attributes = make(map[string]string)
	}
	for k, v := range override.Attributes {
		result.Attributes[k] = v
	}

	return result
}
