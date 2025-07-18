package main

import (
	"fmt"
	"log"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

func main() {
	fmt.Println("=== Exemplo Básico do DomainErrors ===")

	// Criar um erro básico
	err := domainerrors.New("USER_001", "Erro básico de usuário")
	fmt.Printf("Erro básico: %s\n", err.Error())

	// Criar um erro com causa
	baseErr := fmt.Errorf("conexão com banco de dados falhou")
	dbErr := domainerrors.NewWithError("DB_001", "Erro de banco de dados", baseErr)
	fmt.Printf("Erro com causa: %s\n", dbErr.Error())

	// Criar um erro com tipo específico
	validationErr := domainerrors.NewWithType("VAL_001", "Dados inválidos", domainerrors.ErrorTypeValidation)
	fmt.Printf("Erro de validação: %s\n", validationErr.Error())
	fmt.Printf("Tipo: %s\n", validationErr.Type())
	fmt.Printf("HTTP Status: %d\n", validationErr.HTTPStatus())

	// Adicionar metadados
	err.WithMetadata("user_id", "12345")
	err.WithMetadata("action", "create_user")
	fmt.Printf("Metadados: %+v\n", err.Metadata())

	// Verificar tipo de erro
	if domainerrors.IsType(validationErr, domainerrors.ErrorTypeValidation) {
		fmt.Println("✓ Erro é do tipo validação")
	}

	// Serializar para JSON
	if jsonData, err := validationErr.JSON(); err == nil {
		fmt.Printf("JSON: %s\n", string(jsonData))
	}

	// Exemplo com diferentes tipos de erro
	fmt.Println("\n=== Tipos de Erro Específicos ===")

	// Erro de validação
	validationError := domainerrors.NewValidationError("Dados inválidos", nil)
	validationError.WithField("email", "Email é obrigatório")
	validationError.WithField("password", "Senha deve ter pelo menos 8 caracteres")
	fmt.Printf("Erro de validação: %s\n", validationError.Error())
	fmt.Printf("Campos: %+v\n", validationError.Fields)

	// Erro de recurso não encontrado
	notFoundError := domainerrors.NewNotFoundError("Usuário não encontrado")
	notFoundError.WithResource("user", "12345")
	fmt.Printf("Erro não encontrado: %s\n", notFoundError.Error())
	fmt.Printf("Recurso: %s ID: %s\n", notFoundError.Resource, notFoundError.ResourceID)

	// Erro de regra de negócio
	businessError := domainerrors.NewBusinessError("INSUFFICIENT_FUNDS", "Saldo insuficiente")
	businessError.WithRule("minimum_balance")
	fmt.Printf("Erro de negócio: %s\n", businessError.Error())
	fmt.Printf("Regra: %s\n", businessError.RuleName)

	// Erro de banco de dados
	databaseError := domainerrors.NewDatabaseError("Falha na consulta", baseErr)
	databaseError.WithOperation("SELECT", "users")
	databaseError.WithQuery("SELECT * FROM users WHERE id = ?")
	fmt.Printf("Erro de banco: %s\n", databaseError.Error())
	fmt.Printf("Operação: %s Tabela: %s\n", databaseError.Operation, databaseError.Table)

	// Erro de serviço externo
	externalError := domainerrors.NewExternalServiceError("payment-service", "Falha no pagamento", nil)
	externalError.WithEndpoint("/api/v1/payments")
	externalError.WithStatusCode(503)
	externalError.WithResponse("Service temporarily unavailable")
	fmt.Printf("Erro externo: %s\n", externalError.Error())
	fmt.Printf("Serviço: %s Endpoint: %s Status: %d\n",
		externalError.Service, externalError.Endpoint, externalError.HTTPStatusCode)

	// Exemplo de empilhamento de erros
	fmt.Println("\n=== Empilhamento de Erros ===")

	rootErr := fmt.Errorf("network connection failed")
	infraErr := domainerrors.NewInfrastructureError("network", "Falha de infraestrutura", rootErr)
	serviceErr := domainerrors.Wrap("Falha no serviço", infraErr)

	fmt.Printf("Erro final: %s\n", serviceErr.Error())
	fmt.Printf("Cadeia de erros: %s\n", domainerrors.FormatErrorChain(serviceErr))
	fmt.Printf("Causa raiz: %s\n", domainerrors.GetRootCause(serviceErr))

	// Exemplo de grupo de erros
	fmt.Println("\n=== Grupo de Erros ===")

	errorGroup := domainerrors.NewErrorGroup()
	errorGroup.Add(validationError)
	errorGroup.Add(businessError)
	errorGroup.Add(databaseError)

	fmt.Printf("Grupo tem erros: %v\n", errorGroup.HasErrors())
	fmt.Printf("Quantidade de erros: %d\n", errorGroup.Count())
	fmt.Printf("Primeiro erro: %s\n", errorGroup.First().Error())
	fmt.Printf("Último erro: %s\n", errorGroup.Last().Error())

	// Filtrar erros por tipo
	validationErrors := errorGroup.FilterByType(domainerrors.ErrorTypeValidation)
	fmt.Printf("Erros de validação: %d\n", len(validationErrors))

	// Utilitários
	fmt.Println("\n=== Utilitários ===")

	fmt.Printf("Severidade do erro de validação: %s\n", domainerrors.GetSeverity(validationError))
	fmt.Printf("Severidade do erro de banco: %s\n", domainerrors.GetSeverity(databaseError))
	fmt.Printf("Deve repetir erro de timeout: %v\n", domainerrors.ShouldRetry(databaseError))
	fmt.Printf("É recuperável: %v\n", domainerrors.IsRecoverable(databaseError))

	// Mapeamento HTTP
	fmt.Println("\n=== Mapeamento HTTP ===")

	errors := []error{validationError, notFoundError, businessError, databaseError, externalError}
	for _, err := range errors {
		status := domainerrors.GetHTTPStatus(err)
		fmt.Printf("Erro: %s -> HTTP %d\n", domainerrors.GetTypeName(err), status)
	}

	fmt.Println("\n=== Exemplo concluído com sucesso! ===")
}

func demonstrateRecovery() {
	defer func() {
		if r := recover(); r != nil {
			recoveredErr := domainerrors.RecoverWithStackTrace()
			if recoveredErr != nil {
				log.Printf("Erro recuperado: %s", recoveredErr.Error())
			}
		}
	}()

	// Código que pode causar panic
	panic("exemplo de panic")
}
