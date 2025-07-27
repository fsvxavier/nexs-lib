package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/pagination"
	"github.com/fsvxavier/nexs-lib/pagination/config"
)

// ErrorScenario representa um cenário de teste de erro
type ErrorScenario struct {
	Name        string
	Description string
	Params      url.Values
	SortFields  []string
	ExpectedErr string
}

// APIErrorResponse representa uma resposta de erro formatada
type APIErrorResponse struct {
	Error     string                 `json:"error"`
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp"`
	RequestID string                 `json:"request_id,omitempty"`
}

func createErrorScenarios() []ErrorScenario {
	return []ErrorScenario{
		{
			Name:        "Página Inválida - Texto",
			Description: "Parâmetro page com valor não numérico",
			Params: url.Values{
				"page": []string{"abc"},
			},
			SortFields:  []string{"id", "name"},
			ExpectedErr: "INVALID_PAGE_PARAMETER",
		},
		{
			Name:        "Página Inválida - Negativa",
			Description: "Número de página negativo",
			Params: url.Values{
				"page": []string{"-1"},
			},
			SortFields:  []string{"id", "name"},
			ExpectedErr: "INVALID_PAGE_PARAMETER",
		},
		{
			Name:        "Página Inválida - Zero",
			Description: "Página zero (primeira página é 1)",
			Params: url.Values{
				"page": []string{"0"},
			},
			SortFields:  []string{"id", "name"},
			ExpectedErr: "INVALID_PAGE_PARAMETER",
		},
		{
			Name:        "Limite Inválido - Texto",
			Description: "Parâmetro limit com valor não numérico",
			Params: url.Values{
				"limit": []string{"xyz"},
			},
			SortFields:  []string{"id", "name"},
			ExpectedErr: "INVALID_LIMIT_PARAMETER",
		},
		{
			Name:        "Limite Inválido - Negativo",
			Description: "Limite negativo",
			Params: url.Values{
				"limit": []string{"-10"},
			},
			SortFields:  []string{"id", "name"},
			ExpectedErr: "INVALID_LIMIT_PARAMETER",
		},
		{
			Name:        "Limite Inválido - Zero",
			Description: "Limite zero",
			Params: url.Values{
				"limit": []string{"0"},
			},
			SortFields:  []string{"id", "name"},
			ExpectedErr: "INVALID_LIMIT_PARAMETER",
		},
		{
			Name:        "Limite Excede Máximo",
			Description: "Limite maior que o máximo permitido",
			Params: url.Values{
				"limit": []string{"500"},
			},
			SortFields:  []string{"id", "name"},
			ExpectedErr: "LIMIT_TOO_LARGE",
		},
		{
			Name:        "Campo de Ordenação Inválido",
			Description: "Campo de ordenação não permitido",
			Params: url.Values{
				"sort": []string{"password"},
			},
			SortFields:  []string{"id", "name", "email"},
			ExpectedErr: "INVALID_SORT_FIELD",
		},
		{
			Name:        "Ordem de Classificação Inválida",
			Description: "Ordem que não seja asc ou desc",
			Params: url.Values{
				"order": []string{"random"},
			},
			SortFields:  []string{"id", "name"},
			ExpectedErr: "INVALID_SORT_ORDER",
		},
		{
			Name:        "Múltiplos Erros",
			Description: "Múltiplos parâmetros inválidos simultâneos",
			Params: url.Values{
				"page":  []string{"-5"},
				"limit": []string{"abc"},
				"sort":  []string{"invalid_field"},
				"order": []string{"random"},
			},
			SortFields:  []string{"id", "name"},
			ExpectedErr: "INVALID_PAGE_PARAMETER", // Primeiro erro encontrado
		},
	}
}

func formatDomainError(err error) *APIErrorResponse {
	// Verificar se é um domain error
	if domainErr, ok := err.(*domainerrors.DomainError); ok {
		return &APIErrorResponse{
			Error:   "Validation Failed",
			Code:    domainErr.Code,
			Message: domainErr.Message,
			Details: map[string]interface{}{
				"type":       domainErr.Type,
				"field":      extractFieldFromError(domainErr.Message),
				"suggestion": generateSuggestion(domainErr.Code),
			},
			Timestamp: "2024-01-01T12:00:00Z",
			RequestID: "req-" + generateRequestID(),
		}
	}

	// Erro genérico
	return &APIErrorResponse{
		Error:     "Internal Error",
		Code:      "UNKNOWN_ERROR",
		Message:   err.Error(),
		Timestamp: "2024-01-01T12:00:00Z",
		RequestID: "req-" + generateRequestID(),
	}
}

func extractFieldFromError(message string) string {
	// Extrair campo do erro baseado na mensagem
	if strings.Contains(message, "page") {
		return "page"
	}
	if strings.Contains(message, "limit") {
		return "limit"
	}
	if strings.Contains(message, "sort") {
		return "sort"
	}
	if strings.Contains(message, "order") {
		return "order"
	}
	return "unknown"
}

func generateSuggestion(code string) string {
	suggestions := map[string]string{
		"INVALID_PAGE_PARAMETER":  "Use um número inteiro positivo para a página (ex: page=1)",
		"INVALID_LIMIT_PARAMETER": "Use um número inteiro positivo para o limite (ex: limit=10)",
		"LIMIT_TOO_LARGE":         "Reduza o limite para um valor menor ou igual ao máximo permitido",
		"INVALID_SORT_FIELD":      "Use um campo de ordenação válido da lista permitida",
		"INVALID_SORT_ORDER":      "Use 'asc' para crescente ou 'desc' para decrescente",
	}

	if suggestion, exists := suggestions[code]; exists {
		return suggestion
	}
	return "Verifique a documentação da API para parâmetros válidos"
}

func generateRequestID() string {
	// Simular geração de request ID
	return fmt.Sprintf("%d", 12345)
}

func demonstrateErrorScenarios() {
	fmt.Println("=== 1. Demonstração de Cenários de Erro ===")

	// Configurar serviço com limites restritivos para demonstração
	cfg := config.NewDefaultConfig()
	cfg.MaxLimit = 100
	service := pagination.NewPaginationService(cfg)

	scenarios := createErrorScenarios()

	for i, scenario := range scenarios {
		fmt.Printf("\n🧪 Teste %d: %s\n", i+1, scenario.Name)
		fmt.Printf("📝 Descrição: %s\n", scenario.Description)
		fmt.Printf("🔗 Parâmetros: %s\n", formatParams(scenario.Params))

		_, err := service.ParseRequest(scenario.Params, scenario.SortFields...)

		if err != nil {
			fmt.Printf("❌ Erro capturado: %v\n", err)

			// Formatar erro para API
			apiError := formatDomainError(err)
			jsonError, _ := json.MarshalIndent(apiError, "   ", "  ")
			fmt.Printf("📋 Resposta da API:\n   %s\n", string(jsonError))

			// Verificar se é o erro esperado
			if strings.Contains(err.Error(), scenario.ExpectedErr) {
				fmt.Printf("✅ Erro corresponde ao esperado: %s\n", scenario.ExpectedErr)
			} else {
				fmt.Printf("⚠️  Erro diferente do esperado. Esperado: %s\n", scenario.ExpectedErr)
			}
		} else {
			fmt.Printf("⚠️  Nenhum erro capturado (deveria ter falhado)\n")
		}
	}
}

func formatParams(params url.Values) string {
	var parts []string
	for key, values := range params {
		for _, value := range values {
			parts = append(parts, fmt.Sprintf("%s=%s", key, value))
		}
	}
	return strings.Join(parts, "&")
}

func demonstrateErrorRecovery() {
	fmt.Println("\n=== 2. Demonstração de Recuperação de Erros ===")

	service := pagination.NewPaginationService(nil)

	// Cenários de recuperação graceful
	recoveryScenarios := []struct {
		name        string
		params      url.Values
		description string
	}{
		{
			name:        "Parâmetros Vazios",
			params:      url.Values{},
			description: "Deve usar valores padrão quando nenhum parâmetro é fornecido",
		},
		{
			name: "Página Ausente",
			params: url.Values{
				"limit": []string{"20"},
				"sort":  []string{"name"},
			},
			description: "Deve usar página 1 como padrão",
		},
		{
			name: "Limite Ausente",
			params: url.Values{
				"page": []string{"2"},
				"sort": []string{"name"},
			},
			description: "Deve usar limite padrão",
		},
		{
			name: "Ordenação Ausente",
			params: url.Values{
				"page":  []string{"1"},
				"limit": []string{"10"},
			},
			description: "Deve usar ordenação padrão",
		},
	}

	for i, scenario := range recoveryScenarios {
		fmt.Printf("\n🔄 Teste de Recuperação %d: %s\n", i+1, scenario.name)
		fmt.Printf("📝 %s\n", scenario.description)
		fmt.Printf("🔗 Parâmetros: %s\n", formatParams(scenario.params))

		result, err := service.ParseRequest(scenario.params, "id", "name", "email")

		if err != nil {
			fmt.Printf("❌ Erro inesperado: %v\n", err)
		} else {
			fmt.Printf("✅ Recuperação bem-sucedida:\n")
			fmt.Printf("   Página: %d\n", result.Page)
			fmt.Printf("   Limite: %d\n", result.Limit)
			fmt.Printf("   Campo de ordenação: %s\n", result.SortField)
			fmt.Printf("   Ordem: %s\n", result.SortOrder)
		}
	}
}

func demonstrateErrorChaining() {
	fmt.Println("\n=== 3. Demonstração de Encadeamento de Erros ===")

	service := pagination.NewPaginationService(nil)

	// Simular uma cadeia de operações que pode falhar
	operations := []struct {
		name   string
		params url.Values
		fields []string
	}{
		{
			name: "Parse de Parâmetros",
			params: url.Values{
				"page":  []string{"abc"},
				"limit": []string{"10"},
			},
			fields: []string{"id", "name"},
		},
		{
			name: "Validação de Campos",
			params: url.Values{
				"page":  []string{"1"},
				"limit": []string{"10"},
				"sort":  []string{"password"},
			},
			fields: []string{"id", "name"},
		},
		{
			name: "Validação de Limites",
			params: url.Values{
				"page":  []string{"1"},
				"limit": []string{"1000"},
			},
			fields: []string{"id", "name"},
		},
	}

	for i, op := range operations {
		fmt.Printf("\n🔗 Operação %d: %s\n", i+1, op.name)

		// Parse
		params, parseErr := service.ParseRequest(op.params, op.fields...)
		if parseErr != nil {
			fmt.Printf("❌ Falha no parse: %v\n", parseErr)

			// Demonstrar que mesmo com erro, podemos tentar operações subsequentes
			fmt.Printf("🔄 Tentando construir query mesmo assim...\n")

			// Usar parâmetros padrão para continuar
			defaultParams := url.Values{
				"page":  []string{"1"},
				"limit": []string{"10"},
			}
			fallbackParams, _ := service.ParseRequest(defaultParams, op.fields...)
			query := service.BuildQuery("SELECT * FROM users", fallbackParams)
			fmt.Printf("✅ Query de fallback construída: %s\n", query)
			continue
		}

		// Construir query
		query := service.BuildQuery("SELECT * FROM users", params)
		fmt.Printf("✅ Query construída: %s\n", query)

		// Validar página vs total de registros
		totalRecords := 100
		pageValidationErr := service.ValidatePageNumber(params, totalRecords)
		if pageValidationErr != nil {
			fmt.Printf("❌ Validação de página falhou: %v\n", pageValidationErr)
		} else {
			fmt.Printf("✅ Página válida para %d registros\n", totalRecords)
		}
	}
}

func demonstrateUserFriendlyErrors() {
	fmt.Println("\n=== 4. Demonstração de Erros Amigáveis ao Usuário ===")

	service := pagination.NewPaginationService(nil)

	// Erros comuns com mensagens amigáveis
	userScenarios := []struct {
		params      url.Values
		userMessage string
		devMessage  string
	}{
		{
			params:      url.Values{"page": []string{"0"}},
			userMessage: "A primeira página é a número 1. Tente novamente com page=1.",
			devMessage:  "Page parameter must be >= 1",
		},
		{
			params:      url.Values{"limit": []string{"1000"}},
			userMessage: "Muitos resultados solicitados. O máximo permitido é 100 registros por página.",
			devMessage:  "Limit exceeds maximum allowed value",
		},
		{
			params:      url.Values{"sort": []string{"password"}},
			userMessage: "Não é possível ordenar por este campo. Campos disponíveis: id, name, email.",
			devMessage:  "Sort field not in allowed list",
		},
		{
			params:      url.Values{"order": []string{"random"}},
			userMessage: "Ordenação deve ser 'crescente' (asc) ou 'decrescente' (desc).",
			devMessage:  "Invalid sort order value",
		},
	}

	for i, scenario := range userScenarios {
		fmt.Printf("\n👤 Cenário de Usuário %d:\n", i+1)
		fmt.Printf("🔗 Parâmetros: %s\n", formatParams(scenario.params))

		_, err := service.ParseRequest(scenario.params, "id", "name", "email")

		if err != nil {
			fmt.Printf("🔧 Erro técnico: %v\n", err)
			fmt.Printf("📱 Mensagem para usuário: %s\n", scenario.userMessage)
			fmt.Printf("🛠️  Mensagem para desenvolvedor: %s\n", scenario.devMessage)

			// Sugerir correção automática
			correction := suggestCorrection(scenario.params)
			if correction != "" {
				fmt.Printf("💡 Sugestão de correção: %s\n", correction)
			}
		}
	}
}

func suggestCorrection(params url.Values) string {
	corrections := []string{}

	if page := params.Get("page"); page != "" {
		if pageNum, err := strconv.Atoi(page); err != nil || pageNum <= 0 {
			corrections = append(corrections, "page=1")
		}
	}

	if limit := params.Get("limit"); limit != "" {
		if limitNum, err := strconv.Atoi(limit); err != nil || limitNum <= 0 {
			corrections = append(corrections, "limit=10")
		} else if limitNum > 100 {
			corrections = append(corrections, "limit=100")
		}
	}

	if sort := params.Get("sort"); sort != "" {
		validFields := []string{"id", "name", "email"}
		valid := false
		for _, field := range validFields {
			if sort == field {
				valid = true
				break
			}
		}
		if !valid {
			corrections = append(corrections, "sort=id")
		}
	}

	if order := params.Get("order"); order != "" {
		if order != "asc" && order != "desc" {
			corrections = append(corrections, "order=asc")
		}
	}

	if len(corrections) > 0 {
		return strings.Join(corrections, "&")
	}

	return ""
}

func main() {
	fmt.Println("❌ Exemplos de Tratamento de Erros - Módulo de Paginação")
	fmt.Println("========================================================")
	fmt.Println()

	// Executar todas as demonstrações
	demonstrateErrorScenarios()
	demonstrateErrorRecovery()
	demonstrateErrorChaining()
	demonstrateUserFriendlyErrors()

	fmt.Println("\n🎉 Todos os exemplos de tratamento de erros foram executados!")
	fmt.Println()
	fmt.Println("💡 Principais aprendizados:")
	fmt.Println("   • Erros são detectados e classificados automaticamente")
	fmt.Println("   • Domain errors fornecem códigos estruturados")
	fmt.Println("   • Recuperação graceful com valores padrão")
	fmt.Println("   • Mensagens amigáveis para usuários finais")
	fmt.Println("   • Sugestões automáticas de correção")
	fmt.Println("   • Encadeamento de operações com fallbacks")
}
