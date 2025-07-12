package domainerrors

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

// BenchmarkErrorCreation testa a performance de criação de erros
func BenchmarkErrorCreation(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := New("E001", "Test error message")
		_ = err
	}
}

// BenchmarkBuilderPattern testa a performance do builder pattern
func BenchmarkBuilderPattern(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := NewBuilder().
			WithCode("E001").
			WithMessage("Test error message").
			WithType(string(types.ErrorTypeValidation)).
			WithDetail("field", "email").
			WithDetail("value", "test@example.com").
			WithTag("validation").
			WithTag("user").
			Build()
		_ = err
	}
}

// BenchmarkJSONMarshaling testa a performance de serialização JSON
func BenchmarkJSONMarshaling(b *testing.B) {
	err := NewBuilder().
		WithCode("E001").
		WithMessage("Test error message").
		WithType(string(types.ErrorTypeValidation)).
		WithDetail("field", "email").
		WithDetail("value", "test@example.com").
		WithTag("validation").
		WithTag("user").
		Build()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, _ := err.JSON()
		_ = data
	}
}

// BenchmarkJSONMarshalingStdLib compara com json.Marshal padrão
func BenchmarkJSONMarshalingStdLib(b *testing.B) {
	err := NewBuilder().
		WithCode("E001").
		WithMessage("Test error message").
		WithType(string(types.ErrorTypeValidation)).
		WithDetail("field", "email").
		WithDetail("value", "test@example.com").
		WithTag("validation").
		WithTag("user").
		Build()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, _ := json.Marshal(err)
		_ = data
	}
}

// BenchmarkErrorWrapping testa a performance de wrapping de erros
func BenchmarkErrorWrapping(b *testing.B) {
	originalErr := fmt.Errorf("original error")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := NewBuilder().
			WithCode("E001").
			WithMessage("Wrapped error").
			WithCause(originalErr).
			Build()
		_ = err
	}
}

// BenchmarkErrorChaining testa a performance de chaining de erros
func BenchmarkErrorChaining(b *testing.B) {
	baseErr := New("E001", "Base error")
	chainErr := New("E002", "Chain error")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		chained := baseErr.Chain(chainErr)
		_ = chained
	}
}

// BenchmarkValidationError testa a performance de erros de validação
func BenchmarkValidationError(b *testing.B) {
	fields := map[string][]string{
		"email":    {"invalid format", "required field"},
		"age":      {"must be positive"},
		"password": {"too short"},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := NewValidationError("Validation failed", fields)
		_ = err
	}
}

// BenchmarkValidationErrorAddField testa a performance de adição de campos
func BenchmarkValidationErrorAddField(b *testing.B) {
	err := NewValidationError("Validation failed", nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err.AddField(fmt.Sprintf("field_%d", i), "error message")
	}
}

// BenchmarkStackTraceCapture testa a performance de captura de stack trace
func BenchmarkStackTraceCapture(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := NewBuilder().
			WithCode("E001").
			WithMessage("Test error").
			Build()
		_ = err
	}
}

// BenchmarkStackTraceFormat testa a performance de formatação de stack trace
func BenchmarkStackTraceFormat(b *testing.B) {
	err := NewBuilder().
		WithCode("E001").
		WithMessage("Test error").
		Build()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		formatted := err.FormatStackTrace()
		_ = formatted
	}
}

// BenchmarkErrorTypeCheck testa a performance de verificação de tipo
func BenchmarkErrorTypeCheck(b *testing.B) {
	err := NewTimeoutError("Test timeout")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		isRetryable := IsRetryable(err)
		isTemporary := IsTemporary(err)
		errorType := GetErrorType(err)
		_ = isRetryable
		_ = isTemporary
		_ = errorType
	}
}

// BenchmarkConvenienceFunctions testa a performance das funções de conveniência
func BenchmarkConvenienceFunctions(b *testing.B) {
	b.ResetTimer()

	b.Run("NewNotFoundError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := NewNotFoundError("User", "123")
			_ = err
		}
	})

	b.Run("NewUnauthorizedError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := NewUnauthorizedError("Token expired")
			_ = err
		}
	})

	b.Run("NewTimeoutError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := NewTimeoutError("Operation timeout")
			_ = err
		}
	})
}

// BenchmarkConcurrentErrorCreation testa performance com concorrência
func BenchmarkConcurrentErrorCreation(b *testing.B) {
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := NewBuilder().
				WithCode("E001").
				WithMessage("Concurrent error").
				WithType(string(types.ErrorTypeValidation)).
				Build()
			_ = err
		}
	})
}

// BenchmarkConcurrentJSONMarshaling testa JSON marshaling com concorrência
func BenchmarkConcurrentJSONMarshaling(b *testing.B) {
	err := NewBuilder().
		WithCode("E001").
		WithMessage("Test error").
		WithType(string(types.ErrorTypeValidation)).
		Build()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			data, _ := err.JSON()
			_ = data
		}
	})
}

// BenchmarkMemoryAllocation testa alocação de memória
func BenchmarkMemoryAllocation(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Cria erro com vários detalhes para testar alocação
		err := NewBuilder().
			WithCode("E001").
			WithMessage("Memory allocation test").
			WithType(string(types.ErrorTypeValidation)).
			WithDetail("field1", "value1").
			WithDetail("field2", "value2").
			WithDetail("field3", "value3").
			WithDetail("field4", "value4").
			WithDetail("field5", "value5").
			WithTag("tag1").
			WithTag("tag2").
			WithTag("tag3").
			Build()

		// Força uso do erro para evitar otimizações
		_ = err.Error()
		_ = err.Details()
		_ = err.Tags()
	}
}

// BenchmarkComparisonWithStandardError compara com erro padrão do Go
func BenchmarkComparisonWithStandardError(b *testing.B) {
	b.Run("DomainError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := New("E001", "Test error")
			_ = err.Error()
		}
	})

	b.Run("StandardError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := fmt.Errorf("Test error")
			_ = err.Error()
		}
	})
}
