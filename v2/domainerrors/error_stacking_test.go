package domainerrors

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestErrorStackingComprehensive testa empilhamento abrangente de erros
func TestErrorStackingComprehensive(t *testing.T) {
	t.Run("MultiLevelWrapping", func(t *testing.T) {
		// Erro original
		rootErr := errors.New("connection refused")

		// Primeiro nível de wrap
		dbErr := New("DB001", "Database connection failed").Wrap("connecting to database", rootErr)

		// Segundo nível de wrap
		serviceErr := New("SVC001", "Service unavailable").Wrap("calling database service", dbErr)

		// Terceiro nível de wrap
		apiErr := New("API001", "API request failed").Wrap("processing user request", serviceErr)

		// Verifica unwrap de múltiplos níveis
		if apiErr.Unwrap() != serviceErr {
			t.Error("First unwrap should return service error")
		}

		if serviceErr.Unwrap() != dbErr {
			t.Error("Second unwrap should return database error")
		}

		if dbErr.Unwrap() != rootErr {
			t.Error("Third unwrap should return root error")
		}

		// Verifica RootCause atravessa toda a cadeia
		if apiErr.RootCause() != rootErr {
			t.Error("RootCause should return the original error")
		}
	})

	t.Run("ChainWithMultipleErrors", func(t *testing.T) {
		err1 := errors.New("validation failed for field 'email'")
		err2 := errors.New("validation failed for field 'password'")
		err3 := errors.New("validation failed for field 'age'")

		validationErr := New("VAL001", "Multiple validation errors").
			Chain(err1).
			Chain(err2).
			Chain(err3)

		// Verifica se todos os erros estão na representação string
		errorStr := validationErr.Error()
		if !strings.Contains(errorStr, "email") {
			t.Error("Error string should contain 'email' validation")
		}
		if !strings.Contains(errorStr, "password") {
			t.Error("Error string should contain 'password' validation")
		}
		if !strings.Contains(errorStr, "age") {
			t.Error("Error string should contain 'age' validation")
		}
	})

	t.Run("WrapWithChainCombination", func(t *testing.T) {
		// Combina Wrap e Chain
		baseErr := errors.New("disk full")
		validationErr := errors.New("invalid input")
		timeoutErr := errors.New("operation timeout")

		complexErr := New("COMPLEX001", "Complex operation failed").
			Wrap("attempting to save data", baseErr).
			Chain(validationErr).
			Chain(timeoutErr)

		// Verifica que Wrap define o cause
		if complexErr.Unwrap() != baseErr {
			t.Error("Unwrap should return the wrapped error (cause)")
		}

		// Verifica que RootCause encontra o erro original
		if complexErr.RootCause() != baseErr {
			t.Error("RootCause should return the base error")
		}
	})

	t.Run("DeepNestingPerformance", func(t *testing.T) {
		// Testa performance com muitos níveis de aninhamento
		var currentErr error = errors.New("bottom error")

		// Cria 50 níveis de wrapping
		for i := 0; i < 50; i++ {
			currentErr = New(fmt.Sprintf("ERR%03d", i), fmt.Sprintf("Level %d error", i)).
				Wrap(fmt.Sprintf("wrapping level %d", i), currentErr)
		}

		// Verifica que RootCause ainda funciona eficientemente
		domainErr := currentErr.(*DomainError)
		root := domainErr.RootCause()

		if root.Error() != "bottom error" {
			t.Error("RootCause should find bottom error even with deep nesting")
		}
	})

	t.Run("ErrorsIsCompatibility", func(t *testing.T) {
		// Testa compatibilidade com errors.Is
		targetErr := errors.New("target error")

		wrappedErr := New("WRAP001", "Wrapped error").Wrap("wrapping", targetErr)

		// Deve funcionar com errors.Is padrão do Go
		if !errors.Is(wrappedErr, targetErr) {
			t.Error("errors.Is should work with wrapped domain errors")
		}

		// Deve funcionar com a função Is do pacote
		if !Is(wrappedErr, targetErr) {
			t.Error("domainerrors.Is should work with wrapped errors")
		}
	})

	t.Run("ErrorsAsCompatibility", func(t *testing.T) {
		// Testa compatibilidade com errors.As
		var targetDomainErr *DomainError

		originalDomainErr := New("ORIG001", "Original domain error")
		wrappedErr := New("WRAP001", "Wrapper").Wrap("wrapping domain error", originalDomainErr)

		// Deve funcionar com errors.As padrão do Go (encontra o wrapper)
		if !errors.As(wrappedErr, &targetDomainErr) {
			t.Error("errors.As should extract DomainError from wrapped error")
		}

		// O errors.As encontra o primeiro DomainError (wrapper)
		if targetDomainErr.Code() != "WRAP001" {
			t.Errorf("Expected extracted error to have code WRAP001, got %s", targetDomainErr.Code())
		}

		// Para encontrar o erro original, deve usar Unwrap() ou RootCause()
		if originalErr, ok := targetDomainErr.Unwrap().(*DomainError); ok {
			if originalErr.Code() != "ORIG001" {
				t.Errorf("Expected unwrapped error to have code ORIG001, got %s", originalErr.Code())
			}
		} else {
			t.Error("Unwrap should return the original DomainError")
		}

		// Deve funcionar com a função As do pacote
		if !As(wrappedErr, &targetDomainErr) {
			t.Error("domainerrors.As should extract DomainError from wrapped error")
		}
	})

	t.Run("StackTracePreservation", func(t *testing.T) {
		// Verifica se stack traces são preservados através do wrapping
		baseErr := New("BASE001", "Base error")
		wrappedErr := New("WRAP001", "Wrapped error").Wrap("wrapping", baseErr)

		stackTrace := wrappedErr.FormatStackTrace()
		if stackTrace == "" {
			t.Error("Stack trace should be preserved in wrapped errors")
		}

		// Verifica se contém informações de múltiplos níveis
		if !strings.Contains(stackTrace, "wrapping") {
			t.Error("Stack trace should contain wrapping context")
		}
	})

	t.Run("MetadataInheritance", func(t *testing.T) {
		// Testa herança de metadados através do wrapping
		originalErr := NewBuilder().
			WithCode("ORIG001").
			WithMessage("Original error").
			WithDetail("user_id", "12345").
			WithTag("critical").
			WithMetadata(map[string]interface{}{
				"component": "database",
				"operation": "query",
			}).
			Build()

		wrappedErr := NewBuilder().
			WithCode("WRAP001").
			WithMessage("Wrapped error").
			WithDetail("request_id", "req-789").
			Build().
			Wrap("operation failed", originalErr)

		// Verifica herança de detalhes
		details := wrappedErr.Details()
		if details["user_id"] != "12345" {
			t.Error("Should inherit user_id from original error")
		}
		if details["request_id"] != "req-789" {
			t.Error("Should preserve own request_id")
		}

		// Verifica herança de tags
		tags := wrappedErr.Tags()
		hasCritical := false
		for _, tag := range tags {
			if tag == "critical" {
				hasCritical = true
				break
			}
		}
		if !hasCritical {
			t.Error("Should inherit 'critical' tag from original error")
		}

		// Verifica herança de metadados
		metadata := wrappedErr.Metadata()
		if metadata["component"] != "database" {
			t.Error("Should inherit component metadata from original error")
		}
	})

	t.Run("CircularReferenceProtection", func(t *testing.T) {
		// Testa proteção contra referências circulares
		err1 := New("ERR001", "First error")
		err2 := New("ERR002", "Second error").Wrap("wrapping first", err1)

		// Tenta criar uma referência circular (deve ser evitada internamente)
		// Isso não deveria causar loop infinito
		err1.Wrap("wrapping second", err2)

		// Verifica que RootCause não entra em loop infinito
		root := err1.RootCause()
		if root == nil {
			t.Error("RootCause should handle circular references gracefully")
		}
	})
}
