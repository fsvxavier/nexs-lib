package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/fsvxavier/nexs-lib/pagination"
	"github.com/fsvxavier/nexs-lib/pagination/config"
	"github.com/fsvxavier/nexs-lib/pagination/interfaces"
	"github.com/fsvxavier/nexs-lib/pagination/providers"
)

// CustomRequestParser implementa parsing personalizado
type CustomRequestParser struct {
	config *config.Config
}

func NewCustomRequestParser(cfg *config.Config) *CustomRequestParser {
	return &CustomRequestParser{config: cfg}
}

func (p *CustomRequestParser) ParsePaginationParams(params url.Values) (*interfaces.PaginationParams, error) {
	// Parsing personalizado: aceita 'p' para page e 'size' para limit
	page := 1
	limit := p.config.DefaultLimit

	if pageStr := params.Get("p"); pageStr != "" {
		fmt.Printf("🔍 Parsing página personalizada: %s\n", pageStr)
		// Usar providers standard para validação
		standardParser := providers.NewStandardRequestParser(p.config)
		tempParams := url.Values{"page": []string{pageStr}}
		tempResult, err := standardParser.ParsePaginationParams(tempParams)
		if err != nil {
			return nil, err
		}
		page = tempResult.Page
	}

	if limitStr := params.Get("size"); limitStr != "" {
		fmt.Printf("🔍 Parsing tamanho personalizado: %s\n", limitStr)
		standardParser := providers.NewStandardRequestParser(p.config)
		tempParams := url.Values{"limit": []string{limitStr}}
		tempResult, err := standardParser.ParsePaginationParams(tempParams)
		if err != nil {
			return nil, err
		}
		limit = tempResult.Limit
	}

	// Ordenação personalizada: aceita 'order_by' e 'direction'
	sortField := p.config.DefaultSortField
	sortOrder := p.config.DefaultSortOrder

	if orderBy := params.Get("order_by"); orderBy != "" {
		fmt.Printf("🔍 Parsing campo de ordenação personalizado: %s\n", orderBy)
		sortField = orderBy
	}

	if direction := params.Get("direction"); direction != "" {
		fmt.Printf("🔍 Parsing direção personalizada: %s\n", direction)
		sortOrder = direction
	}

	return &interfaces.PaginationParams{
		Page:             page,
		Limit:            limit,
		SortField:        sortField,
		SortOrder:        sortOrder,
		MaxLimit:         p.config.MaxLimit,
		DefaultLimit:     p.config.DefaultLimit,
		DefaultSortField: p.config.DefaultSortField,
		DefaultSortOrder: p.config.DefaultSortOrder,
	}, nil
}

// CustomValidator implementa validação personalizada
type CustomValidator struct {
	config *config.Config
}

func NewCustomValidator(cfg *config.Config) *CustomValidator {
	return &CustomValidator{config: cfg}
}

func (v *CustomValidator) ValidateParams(params *interfaces.PaginationParams, sortableFields []string) error {
	fmt.Printf("✅ Validando parâmetros personalizados\n")

	// Usar validador padrão como base
	standardValidator := providers.NewStandardValidator(v.config)
	if err := standardValidator.ValidateParams(params, sortableFields); err != nil {
		return err
	}

	// Validações personalizadas adicionais
	// Exemplo: não permitir páginas muito altas para proteger performance
	if params.Page > 1000 {
		return fmt.Errorf("[CUSTOM_PAGE_LIMIT] página não pode exceder 1000 para proteção de performance")
	}

	// Exemplo: limites específicos por campo de ordenação
	if params.SortField == "price" && params.Limit > 20 {
		return fmt.Errorf("[CUSTOM_PRICE_LIMIT] quando ordenando por preço, limite máximo é 20")
	}

	fmt.Printf("✅ Validação personalizada passou\n")
	return nil
}

func demonstrateBasicCustomConfig() {
	fmt.Println("=== 1. Demonstração de Configuração Personalizada ===")

	// Configuração personalizada
	customConfig := &config.Config{
		DefaultLimit:      25,                              // Mais registros por página
		MaxLimit:          200,                             // Limite maior
		DefaultSortField:  "created_at",                    // Ordenação por data
		DefaultSortOrder:  "desc",                          // Mais recentes primeiro
		AllowedSortOrders: []string{"asc", "desc", "rand"}, // Incluir ordenação aleatória
		ValidationEnabled: true,
		StrictMode:        true, // Modo rigoroso ativado
	}

	// Validar configuração
	if err := customConfig.Validate(); err != nil {
		log.Fatalf("Configuração inválida: %v", err)
	}

	service := pagination.NewPaginationService(customConfig)

	// Testar com parâmetros padrão
	emptyParams := url.Values{}
	result, err := service.ParseRequest(emptyParams)
	if err != nil {
		log.Fatalf("Erro: %v", err)
	}

	fmt.Printf("📋 Configuração personalizada aplicada:\n")
	fmt.Printf("  Limit padrão: %d\n", result.Limit)
	fmt.Printf("  Campo de ordenação padrão: %s\n", result.SortField)
	fmt.Printf("  Ordem padrão: %s\n", result.SortOrder)
	fmt.Printf("  Limit máximo: %d\n", result.MaxLimit)
	fmt.Println()
}

func demonstrateCustomProviders() {
	fmt.Println("=== 2. Demonstração de Providers Personalizados ===")

	// Configuração base
	cfg := config.NewDefaultConfig()
	cfg.MaxLimit = 100

	// Criar providers personalizados
	customParser := NewCustomRequestParser(cfg)
	customValidator := NewCustomValidator(cfg)
	standardQueryBuilder := providers.NewStandardQueryBuilder()
	standardCalculator := providers.NewStandardPaginationCalculator()

	// Criar serviço com providers personalizados
	service := pagination.NewPaginationServiceWithProviders(
		cfg,
		customParser,
		customValidator,
		standardQueryBuilder,
		standardCalculator,
	)

	// Testar parsing personalizado
	// Usar 'p' para página, 'size' para limit, 'order_by' e 'direction'
	customParams := url.Values{
		"p":         []string{"3"},
		"size":      []string{"15"},
		"order_by":  []string{"name"},
		"direction": []string{"desc"},
	}

	fmt.Printf("🔧 Parâmetros de entrada personalizados: p=3&size=15&order_by=name&direction=desc\n")

	result, err := service.ParseRequest(customParams, "name", "price", "created_at")
	if err != nil {
		log.Fatalf("Erro no parsing personalizado: %v", err)
	}

	fmt.Printf("📊 Resultado do parsing personalizado:\n")
	fmt.Printf("  Página: %d\n", result.Page)
	fmt.Printf("  Limite: %d\n", result.Limit)
	fmt.Printf("  Campo de ordenação: %s\n", result.SortField)
	fmt.Printf("  Ordem: %s\n", result.SortOrder)
	fmt.Println()
}

func demonstrateValidationRules() {
	fmt.Println("=== 3. Demonstração de Regras de Validação Personalizadas ===")

	cfg := config.NewDefaultConfig()
	customValidator := NewCustomValidator(cfg)
	service := pagination.NewPaginationService(cfg)
	service.SetValidator(customValidator)

	// Teste 1: Página muito alta (deve falhar)
	fmt.Printf("🧪 Teste 1: Página muito alta (1001)\n")
	highPageParams := url.Values{
		"page": []string{"1001"},
	}
	_, err := service.ParseRequest(highPageParams)
	if err != nil {
		fmt.Printf("❌ Erro esperado: %v\n", err)
	} else {
		fmt.Printf("✅ Não deveria ter passado!\n")
	}
	fmt.Println()

	// Teste 2: Limite alto com ordenação por preço (deve falhar)
	fmt.Printf("🧪 Teste 2: Limite alto (25) com ordenação por preço\n")
	priceLimitParams := url.Values{
		"limit": []string{"25"},
		"sort":  []string{"price"},
	}
	_, err = service.ParseRequest(priceLimitParams, "price", "name")
	if err != nil {
		fmt.Printf("❌ Erro esperado: %v\n", err)
	} else {
		fmt.Printf("✅ Não deveria ter passado!\n")
	}
	fmt.Println()

	// Teste 3: Parâmetros válidos
	fmt.Printf("🧪 Teste 3: Parâmetros válidos\n")
	validParams := url.Values{
		"page":  []string{"5"},
		"limit": []string{"15"},
		"sort":  []string{"name"},
	}
	result, err := service.ParseRequest(validParams, "price", "name")
	if err != nil {
		fmt.Printf("❌ Erro inesperado: %v\n", err)
	} else {
		fmt.Printf("✅ Validação passou: página %d, limite %d, ordenação %s\n",
			result.Page, result.Limit, result.SortField)
	}
	fmt.Println()
}

func demonstrateAdvancedConfig() {
	fmt.Println("=== 4. Demonstração de Configuração Avançada ===")

	// Configuração para diferentes contextos
	configs := map[string]*config.Config{
		"mobile": {
			DefaultLimit:      10,
			MaxLimit:          25,
			DefaultSortField:  "updated_at",
			DefaultSortOrder:  "desc",
			AllowedSortOrders: []string{"desc"}, // Apenas decrescente para mobile
			ValidationEnabled: true,
			StrictMode:        true,
		},
		"web": {
			DefaultLimit:      50,
			MaxLimit:          200,
			DefaultSortField:  "name",
			DefaultSortOrder:  "asc",
			AllowedSortOrders: []string{"asc", "desc"},
			ValidationEnabled: true,
			StrictMode:        false,
		},
		"api": {
			DefaultLimit:      100,
			MaxLimit:          1000,
			DefaultSortField:  "id",
			DefaultSortOrder:  "asc",
			AllowedSortOrders: []string{"asc", "desc", "random"},
			ValidationEnabled: false, // Sem validação para APIs internas
			StrictMode:        false,
		},
	}

	params := url.Values{
		"page":  []string{"2"},
		"limit": []string{"30"},
		"sort":  []string{"name"},
		"order": []string{"asc"},
	}

	for context, cfg := range configs {
		fmt.Printf("📱 Contexto: %s\n", context)

		service := pagination.NewPaginationService(cfg)
		result, err := service.ParseRequest(params, "id", "name", "updated_at")

		if err != nil {
			fmt.Printf("  ❌ Erro: %v\n", err)
		} else {
			fmt.Printf("  ✅ Limite aplicado: %d (máx: %d)\n", result.Limit, cfg.MaxLimit)
			fmt.Printf("  📊 Configuração: sort=%s, order=%s, strict=%v\n",
				result.SortField, result.SortOrder, cfg.StrictMode)
		}
		fmt.Println()
	}
}

func demonstrateConfigValidation() {
	fmt.Println("=== 5. Demonstração de Validação de Configuração ===")

	// Configurações problemáticas
	badConfigs := []*config.Config{
		{
			DefaultLimit: 0, // Inválido
			MaxLimit:     100,
		},
		{
			DefaultLimit: 50,
			MaxLimit:     0, // Inválido
		},
		{
			DefaultLimit: 100,
			MaxLimit:     50, // Default maior que máximo
		},
	}

	for i, cfg := range badConfigs {
		fmt.Printf("🧪 Teste de configuração %d:\n", i+1)

		err := cfg.Validate()
		if err != nil {
			fmt.Printf("  ❌ Configuração corrigida automaticamente\n")
		}

		fmt.Printf("  📊 Resultado: DefaultLimit=%d, MaxLimit=%d\n", cfg.DefaultLimit, cfg.MaxLimit)

		// Testar se funciona após correção
		service := pagination.NewPaginationService(cfg)
		result, err := service.ParseRequest(url.Values{})
		if err != nil {
			fmt.Printf("  ❌ Ainda há problemas: %v\n", err)
		} else {
			fmt.Printf("  ✅ Funcionando: limit=%d\n", result.Limit)
		}
		fmt.Println()
	}
}

func main() {
	fmt.Println("🎛️  Exemplos de Configuração Personalizada - Módulo de Paginação")
	fmt.Println("================================================================")
	fmt.Println()

	// Executar todas as demonstrações
	demonstrateBasicCustomConfig()
	demonstrateCustomProviders()
	demonstrateValidationRules()
	demonstrateAdvancedConfig()
	demonstrateConfigValidation()

	fmt.Println("🎉 Todos os exemplos de configuração personalizada foram executados!")
	fmt.Println()
	fmt.Println("💡 Principais aprendizados:")
	fmt.Println("   • Configurações podem ser completamente personalizadas")
	fmt.Println("   • Providers podem ser substituídos individualmente")
	fmt.Println("   • Validações personalizadas podem adicionar regras de negócio")
	fmt.Println("   • Diferentes contextos podem ter configurações otimizadas")
	fmt.Println("   • O sistema valida e corrige configurações automaticamente")
}
