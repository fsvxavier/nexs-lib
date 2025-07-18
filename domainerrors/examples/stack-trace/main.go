package main

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/domainerrors"
)

func main() {
	fmt.Println("=== Demonstração FormatStackTrace ===\n")

	// Cria um erro e captura stack trace
	err := domainerrors.New("TEST001", "Test error")

	// Simula adição de frames manualmente para demonstração
	err.Stack = append(err.Stack, domainerrors.StackFrame{
		Function: "main.processUser",
		File:     "/app/user.go",
		Line:     42,
		Message:  "user processing failed",
		Time:     "2025-07-18T14:30:00Z",
	})

	err.Stack = append(err.Stack, domainerrors.StackFrame{
		Function: "main.validateUser",
		File:     "/app/validation.go",
		Line:     15,
		Message:  "validation error",
		Time:     "2025-07-18T14:30:01Z",
	})

	// Demonstra a formatação do stack trace
	fmt.Printf("Erro: %s\n", err.Error())
	fmt.Printf("Tipo: %s\n", err.Type)
	fmt.Printf("Código: %s\n\n", err.Code)

	fmt.Println("Stack Trace Formatado:")
	fmt.Print(err.FormatStackTrace())

	// Demonstra com erro vazio
	fmt.Println("\n=== Erro sem stack trace ===")
	err2 := domainerrors.New("TEST002", "Empty stack error")
	fmt.Printf("Stack Trace (vazio): '%s'\n", err2.FormatStackTrace())

	fmt.Println("=== Demonstração da Função FormatStackTrace ===\n")

	// Exemplo 1: Erro simples com stack trace
	fmt.Println("1. Erro simples com stack trace:")
	err3 := businessLogic()
	if err3 != nil {
		fmt.Printf("Erro: %s\n", err3.Error())
		fmt.Printf("Stack Trace Formatado:\n%s\n", err3.FormatStackTrace())
	}

	// Exemplo 2: Erro com multiple wrapper layers
	fmt.Println("2. Erro com múltiplas camadas de wrapper:")
	err4 := serviceLayer()
	if err4 != nil {
		fmt.Printf("Erro: %s\n", err4.Error())
		fmt.Printf("Stack Trace Formatado:\n%s\n", err4.FormatStackTrace())
	}

	// Exemplo 3: Erro de validação com stack trace
	fmt.Println("3. Erro de validação com stack trace:")
	err5 := validationLayer()
	if err5 != nil {
		fmt.Printf("Erro: %s\n", err5.Error())
		fmt.Printf("Stack Trace Formatado:\n%s\n", err5.FormatStackTrace())
	}
}

// Simula uma função de lógica de negócio que gera um erro
func businessLogic() *domainerrors.DomainError {
	err := domainerrors.New("BUS001", "Business rule violation")
	err.WithMetadata("rule", "max_users_exceeded")
	err.WithMetadata("current_count", 150)
	err.WithMetadata("max_allowed", 100)

	return err
}

// Simula uma camada de serviço que wraps o erro
func serviceLayer() *domainerrors.DomainError {
	businessErr := businessLogic()
	if businessErr != nil {
		wrapped := domainerrors.Wrap("SVC001", "Service layer error", businessErr)
		return wrapped
	}
	return nil
}

// Simula uma camada de validação
func validationLayer() *domainerrors.DomainError {
	valErr := domainerrors.NewValidationError("VAL001", "Invalid user data", nil)
	valErr.WithField("email", "invalid format")
	valErr.WithField("age", "must be positive")

	// Retorna o DomainError embedded
	return valErr.DomainError
}
