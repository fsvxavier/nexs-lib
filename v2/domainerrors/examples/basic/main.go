// Package main demonstra o uso básico do sistema de erros de domínio.
package main

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

func main() {
	fmt.Println("=== Exemplos Básicos de Domain Errors ===")
	fmt.Println()

	// Exemplo 1: Criação simples de erro
	example1()

	// Exemplo 2: Usando builder para erro complexo
	example2()

	// Exemplo 3: Wrapping de erros
	example3()

	// Exemplo 4: Erro de validação
	example4()

	// Exemplo 5: Diferentes tipos de erro
	example5()

	// Exemplo 6: Análise de erros
	example6()
}

func example1() {
	fmt.Println("--- Exemplo 1: Criação Simples ---")

	// Criação básica de erro
	err := domainerrors.New("E001", "User not found")

	fmt.Printf("Error: %s\n", err.Error())
	fmt.Printf("Code: %s\n", err.Code())
	fmt.Printf("Type: %s\n", err.Type())
	fmt.Printf("Status Code: %d\n", err.StatusCode())

	fmt.Println()
}

func example2() {
	fmt.Println("--- Exemplo 2: Builder Pattern ---")

	// Usando builder para construção fluente
	err := domainerrors.NewBuilder().
		WithCode("E002").
		WithMessage("Invalid user data").
		WithType(string(types.ErrorTypeValidation)).
		WithDetail("field", "email").
		WithDetail("value", "invalid-email").
		WithTag("validation").
		WithTag("user").
		Build()

	fmt.Printf("Error: %s\n", err.Error())
	fmt.Printf("Details: %+v\n", err.Details())
	fmt.Printf("Tags: %v\n", err.Tags())

	// Serialização para JSON
	jsonData, _ := err.JSON()
	fmt.Printf("JSON: %s\n", string(jsonData))

	fmt.Println()
}

func example3() {
	fmt.Println("--- Exemplo 3: Wrapping de Erros ---")

	// Simula um erro de banco de dados
	dbErr := fmt.Errorf("connection refused")

	// Envolve o erro com contexto de domínio
	wrappedErr := domainerrors.NewBuilder().
		WithCode("DB001").
		WithMessage("Failed to save user").
		WithType(string(types.ErrorTypeDatabase)).
		WithCause(dbErr).
		Build()

	fmt.Printf("Wrapped Error: %s\n", wrappedErr.Error())
	fmt.Printf("Root Cause: %s\n", wrappedErr.RootCause())
	fmt.Printf("Stack Trace:\n%s\n", wrappedErr.FormatStackTrace())

	fmt.Println()
}

func example4() {
	fmt.Println("--- Exemplo 4: Erro de Validação ---")

	// Erro de validação com múltiplos campos
	fields := map[string][]string{
		"email":    {"invalid format", "required field"},
		"age":      {"must be positive", "must be less than 120"},
		"password": {"too short", "must contain special characters"},
	}

	validationErr := domainerrors.NewValidationError("User validation failed", fields)

	fmt.Printf("Validation Error: %s\n", validationErr.Error())
	fmt.Printf("Has email errors: %t\n", validationErr.HasField("email"))
	fmt.Printf("Email errors: %v\n", validationErr.FieldErrors("email"))

	// Adiciona novo erro de campo
	validationErr.AddField("name", "required field")
	fmt.Printf("After adding name error: %v\n", validationErr.Fields())

	fmt.Println()
}

func example5() {
	fmt.Println("--- Exemplo 5: Diferentes Tipos de Erro ---")

	// Erro de não encontrado
	notFoundErr := domainerrors.NewNotFoundError("User", "12345")
	fmt.Printf("Not Found: %s (Status: %d)\n", notFoundErr.Error(), notFoundErr.StatusCode())

	// Erro de autorização
	authErr := domainerrors.NewUnauthorizedError("Token expired")
	fmt.Printf("Unauthorized: %s (Status: %d)\n", authErr.Error(), authErr.StatusCode())

	// Erro de timeout
	timeoutErr := domainerrors.NewTimeoutError("Database query timeout")
	fmt.Printf("Timeout: %s (Status: %d)\n", timeoutErr.Error(), timeoutErr.StatusCode())

	// Erro de circuit breaker
	circuitErr := domainerrors.NewCircuitBreakerError("payment-service")
	fmt.Printf("Circuit Breaker: %s (Status: %d)\n", circuitErr.Error(), circuitErr.StatusCode())

	fmt.Println()
}

func example6() {
	fmt.Println("--- Exemplo 6: Análise de Erros ---")

	// Cria erros de diferentes tipos
	errors := []error{
		domainerrors.NewTimeoutError("Operation timeout"),
		domainerrors.NewBuilder().WithType(string(types.ErrorTypeValidation)).WithMessage("Invalid input").Build(),
		domainerrors.NewInternalError("Database connection failed", nil),
	}

	for i, err := range errors {
		fmt.Printf("Error %d:\n", i+1)
		fmt.Printf("  Type: %s\n", domainerrors.GetErrorType(err))
		fmt.Printf("  Code: %s\n", domainerrors.GetErrorCode(err))
		fmt.Printf("  Status: %d\n", domainerrors.GetStatusCode(err))
		fmt.Printf("  Retryable: %t\n", domainerrors.IsRetryable(err))
		fmt.Printf("  Temporary: %t\n", domainerrors.IsTemporary(err))
		fmt.Println()
	}
}

// Exemplos de uso em funções de negócio

// UserService simula um serviço de usuário.
type UserService struct{}

// GetUser demonstra como retornar erros de domínio em serviços.
func (s *UserService) GetUser(id string) (*User, error) {
	if id == "" {
		return nil, domainerrors.NewBadRequestError("User ID is required")
	}

	// Simula busca no banco
	if id == "not-found" {
		return nil, domainerrors.NewNotFoundError("User", id)
	}

	if id == "db-error" {
		dbErr := fmt.Errorf("connection timeout")
		return nil, domainerrors.NewInternalError("Failed to query user", dbErr)
	}

	return &User{ID: id, Name: "John Doe"}, nil
}

// CreateUser demonstra validação com erros de domínio.
func (s *UserService) CreateUser(req CreateUserRequest) (*User, error) {
	// Validação com erro estruturado
	validationErr := domainerrors.NewValidationError("User validation failed", nil)

	if req.Name == "" {
		validationErr.AddField("name", "required field")
	}

	if req.Email == "" {
		validationErr.AddField("email", "required field")
	} else if !isValidEmail(req.Email) {
		validationErr.AddField("email", "invalid format")
	}

	if req.Age < 0 {
		validationErr.AddField("age", "must be positive")
	}

	// Retorna erro de validação se houver problemas
	if len(validationErr.Fields()) > 0 {
		return nil, validationErr
	}

	// Simula conflito de email
	if req.Email == "conflict@example.com" {
		return nil, domainerrors.NewConflictError("Email already exists")
	}

	return &User{
		ID:    "new-id",
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}, nil
}

// User representa um usuário.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	Age   int    `json:"age,omitempty"`
}

// CreateUserRequest representa uma requisição de criação de usuário.
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func isValidEmail(email string) bool {
	// Validação simples para o exemplo
	return len(email) > 0 &&
		len(email) < 255 &&
		containsChar(email, '@') &&
		containsChar(email, '.')
}

func containsChar(s string, c rune) bool {
	for _, r := range s {
		if r == c {
			return true
		}
	}
	return false
}

// Demonstração de uso do serviço
func init() {
	fmt.Println("\n=== Demonstração de Serviço ===")

	service := &UserService{}

	// Teste casos de erro
	testCases := []string{"", "not-found", "db-error", "valid-id"}

	for _, tc := range testCases {
		user, err := service.GetUser(tc)
		if err != nil {
			fmt.Printf("GetUser('%s') failed: %s\n", tc, err.Error())

			// Análise do erro
			errorType := domainerrors.GetErrorType(err)
			fmt.Printf("  Error type: %s\n", errorType)
		} else {
			fmt.Printf("GetUser('%s') success: %+v\n", tc, user)
		}
	}

	// Teste criação com validação
	invalidReq := CreateUserRequest{
		Name:  "",
		Email: "invalid-email",
		Age:   -5,
	}

	_, err := service.CreateUser(invalidReq)
	if err != nil {
		fmt.Printf("CreateUser validation error: %s\n", err.Error())

		// Verifica se é um erro de validação baseado no tipo
		if domainerrors.GetErrorType(err) == string(types.ErrorTypeValidation) {
			fmt.Printf("This is a validation error\n")
		}
	}

	fmt.Println()
}
