package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

// HTTPErrorResponse representa uma resposta HTTP de erro
type HTTPErrorResponse struct {
	Error     ErrorDetails `json:"error"`
	RequestID string       `json:"request_id,omitempty"`
	Timestamp string       `json:"timestamp"`
}

type ErrorDetails struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Type       string                 `json:"type"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
}

// ValidationResult representa o resultado de uma validação
type ValidationResult struct {
	Valid  bool                   `json:"valid"`
	Errors []ValidationErrorItem  `json:"errors,omitempty"`
	Fields map[string]interface{} `json:"fields,omitempty"`
}

type ValidationErrorItem struct {
	Field   string      `json:"field"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// User representa um usuário do sistema
type User struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Age    int    `json:"age"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// BankAccount representa uma conta bancária
type BankAccount struct {
	ID      string  `json:"id"`
	UserID  string  `json:"user_id"`
	Balance float64 `json:"balance"`
	Status  string  `json:"status"`
	Type    string  `json:"type"`
}

// Transaction representa uma transação bancária
type Transaction struct {
	ID          string    `json:"id"`
	FromAccount string    `json:"from_account"`
	ToAccount   string    `json:"to_account"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

func main() {
	fmt.Print("=== Outros Casos de Uso - Domain Errors ===\n\n")

	// 1. Validação de formulário complexo
	fmt.Println("1. Validação de Formulário Complexo:")
	demonstrateFormValidation()

	fmt.Println("\n" + strings.Repeat("-", 50))

	// 2. Processamento de transação bancária
	fmt.Println("\n2. Processamento de Transação Bancária:")
	demonstrateBankingTransaction()

	fmt.Println("\n" + strings.Repeat("-", 50))

	// 3. API REST com tratamento de erros
	fmt.Println("\n3. Simulação de API REST:")
	demonstrateRESTAPI()

	fmt.Println("\n" + strings.Repeat("-", 50))

	// 4. Sistema de autenticação
	fmt.Println("\n4. Sistema de Autenticação:")
	demonstrateAuthentication()

	fmt.Println("\n" + strings.Repeat("-", 50))

	// 5. Integração com serviços externos
	fmt.Println("\n5. Integração com Serviços Externos:")
	demonstrateExternalServiceIntegration()

	fmt.Println("\n" + strings.Repeat("-", 50))

	// 6. Sistema de cache com fallback
	fmt.Println("\n6. Sistema de Cache com Fallback:")
	demonstrateCacheSystem()

	fmt.Println("\n=== Fim dos exemplos ===")
}

func demonstrateFormValidation() {
	// Simular dados de um formulário de registro
	formData := map[string]interface{}{
		"name":     "",              // Erro: campo vazio
		"email":    "invalid-email", // Erro: formato inválido
		"age":      "15",            // Erro: idade mínima
		"password": "123",           // Erro: senha muito simples
		"role":     "super-admin",   // Erro: role não permitida
	}

	fmt.Println("Dados do formulário recebidos:")
	printJSON(formData)

	result := validateUser(formData)

	fmt.Printf("\nResultado da validação: %s\n", getBoolIcon(result.Valid))
	if !result.Valid {
		fmt.Printf("Encontrados %d erros:\n", len(result.Errors))
		for i, err := range result.Errors {
			fmt.Printf("  %d. %s: %s (valor: %v)\n",
				i+1, err.Field, err.Message, err.Value)
		}
	}
}

func validateUser(data map[string]interface{}) ValidationResult {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationErrorItem{},
		Fields: make(map[string]interface{}),
	}

	// Validar nome
	if name, ok := data["name"].(string); !ok || strings.TrimSpace(name) == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationErrorItem{
			Field:   "name",
			Code:    "REQUIRED_FIELD",
			Message: "Nome é obrigatório",
			Value:   data["name"],
		})
	}

	// Validar email
	if email, ok := data["email"].(string); !ok || !isValidEmail(email) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationErrorItem{
			Field:   "email",
			Code:    "INVALID_EMAIL",
			Message: "Formato de email inválido",
			Value:   data["email"],
		})
	}

	// Validar idade
	if ageStr, ok := data["age"].(string); ok {
		if age, err := strconv.Atoi(ageStr); err != nil || age < 18 {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationErrorItem{
				Field:   "age",
				Code:    "INVALID_AGE",
				Message: "Idade deve ser maior ou igual a 18 anos",
				Value:   data["age"],
			})
		}
	}

	// Validar password
	if password, ok := data["password"].(string); !ok || len(password) < 8 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationErrorItem{
			Field:   "password",
			Code:    "WEAK_PASSWORD",
			Message: "Senha deve ter pelo menos 8 caracteres",
			Value:   "***",
		})
	}

	// Validar role
	if role, ok := data["role"].(string); ok {
		allowedRoles := []string{"user", "admin", "moderator"}
		if !contains(allowedRoles, role) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationErrorItem{
				Field:   "role",
				Code:    "INVALID_ROLE",
				Message: "Role deve ser: user, admin ou moderator",
				Value:   data["role"],
			})
		}
	}

	return result
}

func demonstrateBankingTransaction() {
	// Simular contas bancárias
	accounts := map[string]BankAccount{
		"acc-001": {
			ID:      "acc-001",
			UserID:  "user-123",
			Balance: 1000.00,
			Status:  "active",
			Type:    "checking",
		},
		"acc-002": {
			ID:      "acc-002",
			UserID:  "user-456",
			Balance: 500.00,
			Status:  "active",
			Type:    "savings",
		},
		"acc-003": {
			ID:      "acc-003",
			UserID:  "user-789",
			Balance: 0.00,
			Status:  "frozen",
			Type:    "checking",
		},
	}

	// Cenários de transação
	transactions := []Transaction{
		{
			ID:          "tx-001",
			FromAccount: "acc-001",
			ToAccount:   "acc-002",
			Amount:      200.00,
			Type:        "transfer",
		},
		{
			ID:          "tx-002",
			FromAccount: "acc-001",
			ToAccount:   "acc-002",
			Amount:      1500.00, // Excede saldo
			Type:        "transfer",
		},
		{
			ID:          "tx-003",
			FromAccount: "acc-003", // Conta congelada
			ToAccount:   "acc-002",
			Amount:      100.00,
			Type:        "transfer",
		},
		{
			ID:          "tx-004",
			FromAccount: "acc-001",
			ToAccount:   "acc-999", // Conta não existe
			Amount:      50.00,
			Type:        "transfer",
		},
	}

	for i, tx := range transactions {
		fmt.Printf("\n--- Transação %d ---\n", i+1)
		fmt.Printf("ID: %s | De: %s -> Para: %s | Valor: R$ %.2f\n",
			tx.ID, tx.FromAccount, tx.ToAccount, tx.Amount)

		err := processTransaction(tx, accounts)
		if err != nil {
			fmt.Printf("❌ Falha: %s\n", err.Error())

			// Converter para domain error se necessário
			if domainErr := convertToDomainError(err); domainErr != nil {
				fmt.Printf("   Código: %s | Tipo: %s | HTTP: %d\n",
					domainErr.Code(), domainErr.Type(), domainErr.HTTPStatus())

				// Mostrar metadados se existirem
				if metadata := domainErr.Metadata(); len(metadata) > 0 {
					fmt.Printf("   Metadados: ")
					printJSON(metadata)
				}
			}
		} else {
			fmt.Printf("✅ Sucesso: Transação processada\n")
		}
	}
}

func processTransaction(tx Transaction, accounts map[string]BankAccount) error {
	// Verificar conta de origem
	fromAccount, exists := accounts[tx.FromAccount]
	if !exists {
		return domainerrors.NewNotFoundError(
			"ACCOUNT_NOT_FOUND",
			fmt.Sprintf("Conta de origem não encontrada: %s", tx.FromAccount),
		).WithMetadata("account_id", tx.FromAccount).
			WithMetadata("transaction_id", tx.ID)
	}

	// Verificar conta de destino
	_, exists = accounts[tx.ToAccount]
	if !exists {
		return domainerrors.NewNotFoundError(
			"DESTINATION_ACCOUNT_NOT_FOUND",
			fmt.Sprintf("Conta de destino não encontrada: %s", tx.ToAccount),
		).WithMetadata("account_id", tx.ToAccount).
			WithMetadata("transaction_id", tx.ID)
	}

	// Verificar status da conta
	if fromAccount.Status != "active" {
		return domainerrors.NewBusinessError(
			"ACCOUNT_FROZEN",
			fmt.Sprintf("Conta de origem está %s", fromAccount.Status),
		).WithMetadata("account_id", tx.FromAccount).
			WithMetadata("account_status", fromAccount.Status).
			WithMetadata("transaction_id", tx.ID)
	}

	// Verificar saldo
	if fromAccount.Balance < tx.Amount {
		return domainerrors.NewBusinessError(
			"INSUFFICIENT_FUNDS",
			"Saldo insuficiente para realizar a transação",
		).WithMetadata("account_id", tx.FromAccount).
			WithMetadata("available_balance", fromAccount.Balance).
			WithMetadata("required_amount", tx.Amount).
			WithMetadata("transaction_id", tx.ID)
	}

	// Verificar limites (exemplo: limite diário)
	dailyLimit := 2000.00
	if tx.Amount > dailyLimit {
		return domainerrors.NewBusinessError(
			"DAILY_LIMIT_EXCEEDED",
			fmt.Sprintf("Valor excede limite diário de R$ %.2f", dailyLimit),
		).WithMetadata("amount", tx.Amount).
			WithMetadata("daily_limit", dailyLimit).
			WithMetadata("transaction_id", tx.ID)
	}

	// Simular processamento bem-sucedido
	return nil
}

func demonstrateRESTAPI() {
	// Simular endpoints de uma API REST
	endpoints := []struct {
		method   string
		path     string
		simulate func() (interface{}, error)
	}{
		{"GET", "/users/123", func() (interface{}, error) {
			// Usuário encontrado
			return User{
				ID:     "123",
				Name:   "João Silva",
				Email:  "joao@example.com",
				Age:    30,
				Role:   "user",
				Status: "active",
			}, nil
		}},
		{"GET", "/users/999", func() (interface{}, error) {
			// Usuário não encontrado
			return nil, domainerrors.NewNotFoundError(
				"USER_NOT_FOUND",
				"Usuário não encontrado",
			).WithMetadata("user_id", "999")
		}},
		{"POST", "/users", func() (interface{}, error) {
			// Erro de validação
			return nil, domainerrors.NewValidationError(
				"VALIDATION_ERROR",
				"Dados de entrada inválidos",
			).WithMetadata("field", "email").
				WithMetadata("value", "invalid-email")
		}},
		{"PUT", "/users/123", func() (interface{}, error) {
			// Erro de autorização
			return nil, domainerrors.NewAuthorizationError(
				"INSUFFICIENT_PERMISSIONS",
				"Permissões insuficientes para esta operação",
			).WithMetadata("user_id", "123").
				WithMetadata("required_role", "admin").
				WithMetadata("current_role", "user")
		}},
	}

	for i, endpoint := range endpoints {
		fmt.Printf("\n--- Request %d ---\n", i+1)
		fmt.Printf("%s %s\n", endpoint.method, endpoint.path)

		data, err := endpoint.simulate()

		if err != nil {
			response := convertErrorToHTTPResponse(err)
			fmt.Printf("Status: %d\n", getHTTPStatusFromError(err))
			fmt.Printf("Response: ")
			printJSON(response)
		} else {
			fmt.Printf("Status: 200 OK\n")
			fmt.Printf("Response: ")
			printJSON(data)
		}
	}
}

func demonstrateAuthentication() {
	// Cenários de autenticação
	authAttempts := []struct {
		username string
		password string
		token    string
	}{
		{"admin", "admin123", ""},       // Login válido
		{"user", "wrongpass", ""},       // Senha errada
		{"", "", "invalid-jwt-token"},   // Token inválido
		{"blocked_user", "pass123", ""}, // Usuário bloqueado
		{"", "", "expired-jwt-token"},   // Token expirado
	}

	for i, attempt := range authAttempts {
		fmt.Printf("\n--- Tentativa de Autenticação %d ---\n", i+1)

		var authType string
		if attempt.token != "" {
			authType = "Token"
			fmt.Printf("Tipo: %s | Token: %s\n", authType, attempt.token)
		} else {
			authType = "Credenciais"
			fmt.Printf("Tipo: %s | Usuário: %s | Senha: %s\n",
				authType, attempt.username, maskPassword(attempt.password))
		}

		user, err := authenticate(attempt.username, attempt.password, attempt.token)

		if err != nil {
			fmt.Printf("❌ Falha na autenticação: %s\n", err.Error())

			if domainErr := convertToDomainError(err); domainErr != nil {
				fmt.Printf("   Código: %s | HTTP: %d\n",
					domainErr.Code(), domainErr.HTTPStatus())
			}
		} else {
			fmt.Printf("✅ Autenticação bem-sucedida: %s (%s)\n",
				user.Name, user.Role)
		}
	}
}

func authenticate(username, password, token string) (*User, error) {
	// Base de usuários simulada
	users := map[string]User{
		"admin": {
			ID:     "user-admin",
			Name:   "Administrador",
			Email:  "admin@example.com",
			Role:   "admin",
			Status: "active",
		},
		"blocked_user": {
			ID:     "user-blocked",
			Name:   "Usuário Bloqueado",
			Email:  "blocked@example.com",
			Role:   "user",
			Status: "blocked",
		},
	}

	if token != "" {
		// Validação de token
		switch token {
		case "valid-jwt-token":
			user := users["admin"]
			return &user, nil
		case "expired-jwt-token":
			return nil, domainerrors.NewAuthenticationError(
				"TOKEN_EXPIRED",
				"Token de acesso expirado",
			).WithMetadata("token_type", "JWT").
				WithMetadata("expired_at", "2024-12-14T20:00:00Z")
		default:
			return nil, domainerrors.NewAuthenticationError(
				"INVALID_TOKEN",
				"Token de acesso inválido",
			).WithMetadata("token_type", "JWT")
		}
	}

	// Validação de credenciais
	if username == "" || password == "" {
		return nil, domainerrors.NewValidationError(
			"MISSING_CREDENTIALS",
			"Usuário e senha são obrigatórios",
		)
	}

	user, exists := users[username]
	if !exists {
		return nil, domainerrors.NewAuthenticationError(
			"INVALID_CREDENTIALS",
			"Credenciais inválidas",
		).WithMetadata("username", username)
	}

	// Verificar status do usuário
	if user.Status == "blocked" {
		return nil, domainerrors.NewAuthorizationError(
			"USER_BLOCKED",
			"Usuário está bloqueado",
		).WithMetadata("user_id", user.ID).
			WithMetadata("blocked_reason", "Múltiplas tentativas de login falharam")
	}

	// Simular verificação de senha
	expectedPassword := map[string]string{
		"admin": "admin123",
	}

	if expectedPass, exists := expectedPassword[username]; !exists || password != expectedPass {
		return nil, domainerrors.NewAuthenticationError(
			"INVALID_CREDENTIALS",
			"Credenciais inválidas",
		).WithMetadata("username", username)
	}

	return &user, nil
}

func demonstrateExternalServiceIntegration() {
	// Simular chamadas para serviços externos
	services := []struct {
		name     string
		endpoint string
		simulate func() error
	}{
		{"Payment Gateway", "https://api.payment.com/charge", func() error {
			// Simular timeout
			return domainerrors.NewTimeoutError(
				"PAYMENT_TIMEOUT",
				"Timeout ao conectar com gateway de pagamento",
			).WithMetadata("service", "payment-gateway").
				WithMetadata("timeout_ms", 5000).
				WithMetadata("retry_count", 3)
		}},
		{"Email Service", "https://api.email.com/send", func() error {
			// Simular rate limit
			return domainerrors.NewRateLimitError(
				"EMAIL_RATE_LIMIT",
				"Limite de emails por minuto excedido",
			).WithMetadata("service", "email-service").
				WithMetadata("limit", 100).
				WithMetadata("window", "1m")
		}},
		{"User Service", "https://api.users.com/profile", func() error {
			// Simular erro de serviço
			return domainerrors.NewWithMetadata(
				interfaces.ExternalServiceError,
				"USER_SERVICE_ERROR",
				"Serviço de usuários indisponível",
				map[string]interface{}{
					"service":     "user-service",
					"status_code": 503,
					"retry_after": "30s",
				},
			)
		}},
		{"Cache Service", "redis://cache:6379", func() error {
			// Sucesso
			return nil
		}},
	}

	for i, service := range services {
		fmt.Printf("\n--- Chamada para Serviço %d ---\n", i+1)
		fmt.Printf("Serviço: %s\n", service.name)
		fmt.Printf("Endpoint: %s\n", service.endpoint)

		err := service.simulate()
		if err != nil {
			fmt.Printf("❌ Falha: %s\n", err.Error())

			if domainErr := convertToDomainError(err); domainErr != nil {
				fmt.Printf("   Código: %s | Tipo: %s\n",
					domainErr.Code(), domainErr.Type())

				// Sugerir ação baseada no tipo de erro
				switch domainErr.Type() {
				case interfaces.TimeoutError:
					fmt.Printf("   💡 Sugestão: Implementar retry com backoff\n")
				case interfaces.RateLimitError:
					fmt.Printf("   💡 Sugestão: Aguardar antes da próxima tentativa\n")
				case interfaces.ExternalServiceError:
					fmt.Printf("   💡 Sugestão: Usar fallback ou circuit breaker\n")
				}
			}
		} else {
			fmt.Printf("✅ Sucesso: Serviço respondeu corretamente\n")
		}
	}
}

func demonstrateCacheSystem() {
	// Simular sistema de cache com fallback
	cacheKeys := []string{
		"user:123",     // Hit
		"user:456",     // Miss - dados existem na fonte
		"user:999",     // Miss - dados não existem
		"config:app",   // Hit
		"temp:session", // Expired
	}

	for i, key := range cacheKeys {
		fmt.Printf("\n--- Cache Lookup %d ---\n", i+1)
		fmt.Printf("Chave: %s\n", key)

		data, err := getCachedData(key)

		if err != nil {
			fmt.Printf("❌ Erro no cache: %s\n", err.Error())

			if domainErr := convertToDomainError(err); domainErr != nil {
				// Tentar fallback baseado no tipo de erro
				if domainErr.Type() == interfaces.CacheError {
					fmt.Printf("   🔄 Tentando fallback para fonte de dados...\n")

					fallbackData, fallbackErr := getFallbackData(key)
					if fallbackErr != nil {
						fmt.Printf("   ❌ Fallback falhou: %s\n", fallbackErr.Error())
					} else {
						fmt.Printf("   ✅ Fallback bem-sucedido: ")
						printJSON(fallbackData)
					}
				}
			}
		} else {
			fmt.Printf("✅ Cache hit: ")
			printJSON(data)
		}
	}
}

func getCachedData(key string) (interface{}, error) {
	// Simular diferentes cenários de cache
	switch key {
	case "user:123":
		return map[string]interface{}{
			"id":        "123",
			"name":      "João Silva",
			"cached_at": time.Now().Add(-5 * time.Minute),
		}, nil

	case "config:app":
		return map[string]interface{}{
			"app_name":  "Domain Errors Example",
			"version":   "1.0.0",
			"cached_at": time.Now().Add(-1 * time.Hour),
		}, nil

	case "temp:session":
		return nil, domainerrors.NewWithMetadata(
			interfaces.CacheError,
			"CACHE_KEY_EXPIRED",
			"Chave do cache expirou",
			map[string]interface{}{
				"key":         key,
				"expired_at":  time.Now().Add(-10 * time.Minute),
				"ttl_seconds": 300,
			},
		)

	default:
		return nil, domainerrors.NewWithMetadata(
			interfaces.CacheError,
			"CACHE_MISS",
			"Dados não encontrados no cache",
			map[string]interface{}{
				"key":         key,
				"cache_layer": "redis",
			},
		)
	}
}

func getFallbackData(key string) (interface{}, error) {
	// Simular busca na fonte de dados original
	switch {
	case strings.HasPrefix(key, "user:"):
		userID := strings.TrimPrefix(key, "user:")
		if userID == "456" {
			return map[string]interface{}{
				"id":     "456",
				"name":   "Maria Santos",
				"source": "database",
			}, nil
		} else if userID == "999" {
			return nil, domainerrors.NewNotFoundError(
				"USER_NOT_FOUND",
				"Usuário não encontrado na fonte de dados",
			).WithMetadata("user_id", userID)
		}

	case strings.HasPrefix(key, "config:"):
		return map[string]interface{}{
			"default_config": true,
			"source":         "fallback",
		}, nil
	}

	return nil, domainerrors.NewNotFoundError(
		"DATA_NOT_FOUND",
		"Dados não encontrados na fonte",
	).WithMetadata("key", key)
}

// Funções utilitárias
func convertToDomainError(err error) interfaces.DomainErrorInterface {
	if domainErr, ok := err.(interfaces.DomainErrorInterface); ok {
		return domainErr
	}
	return nil
}

func convertErrorToHTTPResponse(err error) HTTPErrorResponse {
	if domainErr := convertToDomainError(err); domainErr != nil {
		return HTTPErrorResponse{
			Error: ErrorDetails{
				Code:    domainErr.Code(),
				Message: domainErr.Error(),
				Type:    string(domainErr.Type()),
				Details: domainErr.Metadata(),
			},
			RequestID: "req-" + strconv.FormatInt(time.Now().UnixNano(), 36),
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	return HTTPErrorResponse{
		Error: ErrorDetails{
			Code:    "UNKNOWN_ERROR",
			Message: err.Error(),
			Type:    "unknown",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func getHTTPStatusFromError(err error) int {
	if domainErr := convertToDomainError(err); domainErr != nil {
		return domainErr.HTTPStatus()
	}
	return http.StatusInternalServerError
}

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func maskPassword(password string) string {
	if password == "" {
		return ""
	}
	return strings.Repeat("*", len(password))
}

func getBoolIcon(b bool) string {
	if b {
		return "✅"
	}
	return "❌"
}

func printJSON(data interface{}) {
	if jsonData, err := json.MarshalIndent(data, "", "  "); err == nil {
		fmt.Println(string(jsonData))
	} else {
		fmt.Printf("%+v\n", data)
	}
}
