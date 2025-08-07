package main

import (
	"context"
	"fmt"

	"github.com/fsvxavier/nexs-lib/domainerrors"
	"github.com/fsvxavier/nexs-lib/domainerrors/interfaces"
)

func main() {
	fmt.Println("=== Exemplo Básico de Domain Errors ===")
	fmt.Println()

	// 1. Criando erros básicos
	fmt.Println("1. Criando erros básicos:")

	// Erro de validação
	validationErr := domainerrors.NewValidationError("VAL001", "Campo 'email' é obrigatório")
	fmt.Printf("   Validation Error: %s [%s] - HTTP Status: %d\n", validationErr.Error(), validationErr.Code(), validationErr.HTTPStatus())

	// Erro de não encontrado
	notFoundErr := domainerrors.NewNotFoundError("NF001", "Usuário não encontrado")
	fmt.Printf("   NotFound Error: %s [%s] - HTTP Status: %d\n", notFoundErr.Error(), notFoundErr.Code(), notFoundErr.HTTPStatus())

	// Erro de negócio
	businessErr := domainerrors.NewBusinessError("BIZ001", "Saldo insuficiente para a operação")
	fmt.Printf("   Business Error: %s [%s] - HTTP Status: %d\n\n", businessErr.Error(), businessErr.Code(), businessErr.HTTPStatus())

	// 2. Adicionando metadados
	fmt.Println("2. Adicionando metadados:")

	enrichedErr := validationErr.WithMetadata("field", "email").
		WithMetadata("value", "invalid-email").
		WithMetadata("rule", "required")

	fmt.Printf("   Error with metadata: %s\n", enrichedErr.Error())
	metadata := enrichedErr.Metadata()
	for key, value := range metadata {
		fmt.Printf("     %s: %v\n", key, value)
	}
	fmt.Println()

	// 3. Encapsulando erros
	fmt.Println("3. Encapsulando erros:")

	originalErr := fmt.Errorf("conexão com banco de dados falhou")
	wrappedErr := domainerrors.Wrap(originalErr, interfaces.DatabaseError, "DB001", "Falha ao acessar dados do usuário")

	fmt.Printf("   Wrapped Error: %s\n", wrappedErr.Error())
	fmt.Printf("   Root Cause: %s\n\n", wrappedErr.Unwrap().Error())

	// 4. Verificação de tipos
	fmt.Println("4. Verificação de tipos:")

	isValidation := domainerrors.IsType(validationErr, interfaces.ValidationError)
	isDatabase := domainerrors.IsType(validationErr, interfaces.DatabaseError)

	fmt.Printf("   validationErr é ValidationError? %t\n", isValidation)
	fmt.Printf("   validationErr é DatabaseError? %t\n\n", isDatabase)

	// 5. Análise de cadeia de erros
	fmt.Println("5. Análise de cadeia de erros:")

	// Criando uma cadeia de erros
	layer1 := fmt.Errorf("erro na camada de persistência")
	layer2 := domainerrors.Wrap(layer1, interfaces.DatabaseError, "DB002", "erro na camada de serviço")
	layer3 := fmt.Errorf("erro na camada de apresentação: %w", layer2)

	fmt.Println("   Cadeia de erros:")
	fmt.Println(domainerrors.FormatErrorChain(layer3))

	fmt.Printf("   Causa raiz: %s\n\n", domainerrors.GetRootCause(layer3).Error())

	// 6. Serialização JSON
	fmt.Println("6. Serialização JSON:")

	jsonErr := domainerrors.NewWithMetadata(
		interfaces.ValidationError,
		"VAL002",
		"Dados inválidos fornecidos",
		map[string]interface{}{
			"field": "age",
			"value": -5,
			"min":   0,
			"max":   120,
		},
	)

	// Test serialization
	jsonData, err := jsonErr.ToJSON()
	if err != nil {
		fmt.Printf("Failed to serialize error: %v\n", err)
	} else {
		fmt.Printf("Error JSON: %s\n", string(jsonData))
	}

	// 7. Uso com context
	fmt.Println("7. Uso com context:")

	ctx := context.WithValue(context.Background(), "request_id", "req-12345")
	ctx = context.WithValue(ctx, "user_id", "user-67890")

	contextErr := businessErr.WithContext(ctx)
	fmt.Printf("   Error com context: %s [%s]\n\n", contextErr.Error(), contextErr.Code())

	// 8. Factory personalizada
	fmt.Println("8. Factory personalizada:")

	factory := domainerrors.GetFactory()
	customErr := factory.New(interfaces.TimeoutError, "TO001", "Operação expirou após 30 segundos")

	fmt.Printf("   Custom Error: %s [%s] - HTTP Status: %d\n", customErr.Error(), customErr.Code(), customErr.HTTPStatus())
	fmt.Printf("   Timestamp: %s\n", customErr.Timestamp().Format("2006-01-02 15:04:05"))

	fmt.Println("\n=== Fim do Exemplo Básico ===")
}
