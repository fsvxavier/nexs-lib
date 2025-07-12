// Package parsers implementa um sistema de registry distribuído para parsers de erro
// com configuração dinâmica e arquitetura de plugins.
package parsers

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
)

// ParserPlugin define a interface para plugins de parser.
type ParserPlugin interface {
	// Name retorna o nome único do plugin
	Name() string

	// Version retorna a versão do plugin
	Version() string

	// Description retorna a descrição do plugin
	Description() string

	// CreateParser cria uma nova instância do parser
	CreateParser(config map[string]interface{}) (interfaces.ErrorParser, error)

	// ValidateConfig valida a configuração do parser
	ValidateConfig(config map[string]interface{}) error

	// DefaultConfig retorna a configuração padrão
	DefaultConfig() map[string]interface{}
}

// ParserFactory define a interface para custom error factories.
type ParserFactory interface {
	// CreateParser cria um parser baseado no tipo
	CreateParser(parserType string, config map[string]interface{}) (interfaces.ErrorParser, error)

	// SupportedTypes retorna os tipos suportados
	SupportedTypes() []string

	// RegisterCustomType registra um tipo customizado
	RegisterCustomType(typeName string, factory func(config map[string]interface{}) (interfaces.ErrorParser, error)) error
}

// DistributedParserRegistry implementa um registry distribuído para parsers.
type DistributedParserRegistry struct {
	mu             sync.RWMutex
	parsers        map[string]interfaces.ErrorParser
	plugins        map[string]ParserPlugin
	factories      map[string]ParserFactory
	configurations map[string]map[string]interface{}
	priorities     map[string]int
	enabled        map[string]bool

	// Para observabilidade
	metrics        map[string]*ParserMetrics
	healthCheckers map[string]func() error
}

// ParserMetrics contém métricas de um parser.
type ParserMetrics struct {
	TotalParsed    int64
	SuccessCount   int64
	ErrorCount     int64
	AverageLatency float64
}

// NewDistributedParserRegistry cria um novo registry distribuído.
func NewDistributedParserRegistry() *DistributedParserRegistry {
	registry := &DistributedParserRegistry{
		parsers:        make(map[string]interfaces.ErrorParser),
		plugins:        make(map[string]ParserPlugin),
		factories:      make(map[string]ParserFactory),
		configurations: make(map[string]map[string]interface{}),
		priorities:     make(map[string]int),
		enabled:        make(map[string]bool),
		metrics:        make(map[string]*ParserMetrics),
		healthCheckers: make(map[string]func() error),
	}

	// Registra parsers padrão
	registry.registerDefaultParsers()

	return registry
}

// registerDefaultParsers registra os parsers padrão do sistema.
func (r *DistributedParserRegistry) registerDefaultParsers() {
	defaultParsers := map[string]func() interfaces.ErrorParser{
		"postgresql":          func() interfaces.ErrorParser { return NewPostgreSQLErrorParser() },
		"postgresql_enhanced": func() interfaces.ErrorParser { return NewEnhancedPostgreSQLErrorParser() },
		"pgx":                 func() interfaces.ErrorParser { return NewPGXErrorParser() },
		"grpc":                func() interfaces.ErrorParser { return NewGRPCErrorParser() },
		"http":                func() interfaces.ErrorParser { return NewHTTPErrorParser() },
		"redis":               func() interfaces.ErrorParser { return NewRedisErrorParser() },
		"mongodb":             func() interfaces.ErrorParser { return NewMongoDBErrorParser() },
		"aws":                 func() interfaces.ErrorParser { return NewAWSErrorParser() },
	}

	for name, factory := range defaultParsers {
		parser := factory()
		r.parsers[name] = parser
		r.enabled[name] = true
		r.priorities[name] = 100 // Prioridade padrão
		r.metrics[name] = &ParserMetrics{}
	}
}

// RegisterParser registra um parser no registry.
func (r *DistributedParserRegistry) RegisterParser(name string, parser interfaces.ErrorParser, priority int) error {
	if name == "" {
		return fmt.Errorf("parser name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.parsers[name] = parser
	r.priorities[name] = priority
	r.enabled[name] = true
	r.metrics[name] = &ParserMetrics{}

	return nil
}

// RegisterPlugin registra um plugin de parser.
func (r *DistributedParserRegistry) RegisterPlugin(plugin ParserPlugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Valida configuração padrão
	if err := plugin.ValidateConfig(plugin.DefaultConfig()); err != nil {
		return fmt.Errorf("invalid default config for plugin %s: %w", name, err)
	}

	r.plugins[name] = plugin
	r.configurations[name] = plugin.DefaultConfig()

	// Cria parser com configuração padrão
	parser, err := plugin.CreateParser(plugin.DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to create parser for plugin %s: %w", name, err)
	}

	r.parsers[name] = parser
	r.enabled[name] = true
	r.priorities[name] = 50 // Prioridade mais baixa para plugins
	r.metrics[name] = &ParserMetrics{}

	return nil
}

// RegisterFactory registra uma factory customizada.
func (r *DistributedParserRegistry) RegisterFactory(name string, factory ParserFactory) error {
	if name == "" {
		return fmt.Errorf("factory name cannot be empty")
	}
	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.factories[name] = factory

	return nil
}

// ConfigureParser configura um parser específico.
func (r *DistributedParserRegistry) ConfigureParser(name string, config map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Verifica se é um plugin
	if plugin, exists := r.plugins[name]; exists {
		// Valida configuração
		if err := plugin.ValidateConfig(config); err != nil {
			return fmt.Errorf("invalid config for plugin %s: %w", name, err)
		}

		// Cria novo parser com nova configuração
		parser, err := plugin.CreateParser(config)
		if err != nil {
			return fmt.Errorf("failed to reconfigure parser %s: %w", name, err)
		}

		r.parsers[name] = parser
		r.configurations[name] = config

		return nil
	}

	// Verifica se é uma factory
	for factoryName, factory := range r.factories {
		for _, supportedType := range factory.SupportedTypes() {
			if supportedType == name {
				parser, err := factory.CreateParser(name, config)
				if err != nil {
					return fmt.Errorf("failed to create parser %s using factory %s: %w", name, factoryName, err)
				}

				r.parsers[name] = parser
				r.configurations[name] = config

				return nil
			}
		}
	}

	return fmt.Errorf("parser %s not found or not configurable", name)
}

// EnableParser habilita um parser específico.
func (r *DistributedParserRegistry) EnableParser(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.parsers[name]; !exists {
		return fmt.Errorf("parser %s not found", name)
	}

	r.enabled[name] = true
	return nil
}

// DisableParser desabilita um parser específico.
func (r *DistributedParserRegistry) DisableParser(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.parsers[name]; !exists {
		return fmt.Errorf("parser %s not found", name)
	}

	r.enabled[name] = false
	return nil
}

// SetPriority define a prioridade de um parser.
func (r *DistributedParserRegistry) SetPriority(name string, priority int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.parsers[name]; !exists {
		return fmt.Errorf("parser %s not found", name)
	}

	r.priorities[name] = priority
	return nil
}

// Parse tenta parsear um erro usando os parsers registrados em ordem de prioridade.
func (r *DistributedParserRegistry) Parse(ctx context.Context, err error) (*interfaces.ParsedError, string, error) {
	if err == nil {
		return nil, "", fmt.Errorf("error cannot be nil")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Obtém parsers ordenados por prioridade
	orderedParsers := r.getOrderedParsers()

	for _, parserInfo := range orderedParsers {
		name := parserInfo.name
		parser := parserInfo.parser

		// Verifica se está habilitado
		if !r.enabled[name] {
			continue
		}

		// Verifica contexto de cancelamento
		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		default:
		}

		// Tenta parsear
		if parser.CanParse(err) {
			parsed := parser.Parse(err)

			// Atualiza métricas
			r.updateMetrics(name, true)

			return &parsed, name, nil
		}
	}

	// Nenhum parser conseguiu processar
	return nil, "", fmt.Errorf("no parser found for error: %v", err)
}

// parserInfo contém informações de um parser para ordenação.
type parserInfo struct {
	name     string
	parser   interfaces.ErrorParser
	priority int
}

// getOrderedParsers retorna parsers ordenados por prioridade (maior prioridade primeiro).
func (r *DistributedParserRegistry) getOrderedParsers() []parserInfo {
	var parsers []parserInfo

	for name, parser := range r.parsers {
		parsers = append(parsers, parserInfo{
			name:     name,
			parser:   parser,
			priority: r.priorities[name],
		})
	}

	// Ordena por prioridade (decrescente)
	sort.Slice(parsers, func(i, j int) bool {
		return parsers[i].priority > parsers[j].priority
	})

	return parsers
}

// updateMetrics atualiza as métricas de um parser.
func (r *DistributedParserRegistry) updateMetrics(name string, success bool) {
	if metrics, exists := r.metrics[name]; exists {
		metrics.TotalParsed++
		if success {
			metrics.SuccessCount++
		} else {
			metrics.ErrorCount++
		}
	}
}

// GetMetrics retorna as métricas de um parser.
func (r *DistributedParserRegistry) GetMetrics(name string) (*ParserMetrics, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if metrics, exists := r.metrics[name]; exists {
		// Retorna uma cópia para evitar modificações concorrentes
		return &ParserMetrics{
			TotalParsed:    metrics.TotalParsed,
			SuccessCount:   metrics.SuccessCount,
			ErrorCount:     metrics.ErrorCount,
			AverageLatency: metrics.AverageLatency,
		}, nil
	}

	return nil, fmt.Errorf("metrics not found for parser %s", name)
}

// ListParsers retorna lista de parsers registrados.
func (r *DistributedParserRegistry) ListParsers() map[string]bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]bool)
	for name := range r.parsers {
		result[name] = r.enabled[name]
	}

	return result
}

// GetConfiguration retorna a configuração de um parser.
func (r *DistributedParserRegistry) GetConfiguration(name string) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if config, exists := r.configurations[name]; exists {
		// Retorna uma cópia da configuração
		result := make(map[string]interface{})
		for k, v := range config {
			result[k] = v
		}
		return result, nil
	}

	return nil, fmt.Errorf("configuration not found for parser %s", name)
}

// RegisterHealthChecker registra um health checker para um parser.
func (r *DistributedParserRegistry) RegisterHealthChecker(name string, checker func() error) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.parsers[name]; !exists {
		return fmt.Errorf("parser %s not found", name)
	}

	r.healthCheckers[name] = checker
	return nil
}

// HealthCheck executa health check em todos os parsers ou em um específico.
func (r *DistributedParserRegistry) HealthCheck(name string) map[string]error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]error)

	if name != "" {
		// Health check específico
		if checker, exists := r.healthCheckers[name]; exists {
			results[name] = checker()
		} else {
			results[name] = fmt.Errorf("no health checker for parser %s", name)
		}
	} else {
		// Health check de todos
		for parserName, checker := range r.healthCheckers {
			results[parserName] = checker()
		}
	}

	return results
}

// Global registry instance
var globalParserRegistry *DistributedParserRegistry
var globalRegistryOnce sync.Once

// GetGlobalParserRegistry retorna a instância global do registry.
func GetGlobalParserRegistry() *DistributedParserRegistry {
	globalRegistryOnce.Do(func() {
		globalParserRegistry = NewDistributedParserRegistry()
	})
	return globalParserRegistry
}

// ParseWithGlobalRegistry parseia um erro usando o registry global.
func ParseWithGlobalRegistry(ctx context.Context, err error) (*interfaces.ParsedError, string, error) {
	return GetGlobalParserRegistry().Parse(ctx, err)
}
