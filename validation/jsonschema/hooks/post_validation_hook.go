package hooks

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema/interfaces"
)

// ErrorEnrichmentHook enriquece erros com informações adicionais
type ErrorEnrichmentHook struct {
	AddContext     bool
	AddSuggestions bool
}

// Execute enriquece erros de validação com contexto adicional
func (h *ErrorEnrichmentHook) Execute(data interface{}, errors []interfaces.ValidationError) ([]interfaces.ValidationError, error) {
	if len(errors) == 0 {
		return errors, nil
	}

	enrichedErrors := make([]interfaces.ValidationError, len(errors))
	copy(enrichedErrors, errors)

	for i, err := range enrichedErrors {
		if h.AddContext {
			enrichedErrors[i].Description = h.addContextToError(err)
		}

		if h.AddSuggestions {
			enrichedErrors[i].Description += h.addSuggestionToError(err)
		}
	}

	return enrichedErrors, nil
}

func (h *ErrorEnrichmentHook) addContextToError(err interfaces.ValidationError) string {
	switch err.ErrorType {
	case "REQUIRED_ATTRIBUTE_MISSING":
		return fmt.Sprintf("Campo obrigatório '%s' está ausente", err.Field)
	case "INVALID_DATA_TYPE":
		return fmt.Sprintf("Tipo de dados inválido no campo '%s'", err.Field)
	case "INVALID_FORMAT":
		return fmt.Sprintf("Formato inválido no campo '%s'", err.Field)
	case "INVALID_VALUE":
		return fmt.Sprintf("Valor inválido no campo '%s'", err.Field)
	case "INVALID_LENGTH":
		return fmt.Sprintf("Comprimento inválido no campo '%s'", err.Field)
	default:
		return fmt.Sprintf("Erro de validação no campo '%s'", err.Field)
	}
}

func (h *ErrorEnrichmentHook) addSuggestionToError(err interfaces.ValidationError) string {
	switch err.ErrorType {
	case "REQUIRED_ATTRIBUTE_MISSING":
		return ". Certifique-se de incluir este campo na requisição."
	case "INVALID_DATA_TYPE":
		return ". Verifique se o tipo de dados está correto (string, number, boolean, etc.)."
	case "INVALID_FORMAT":
		return ". Verifique se o formato está de acordo com o esperado."
	case "INVALID_VALUE":
		return ". Verifique se o valor está dentro do intervalo permitido."
	case "INVALID_LENGTH":
		return ". Verifique se o comprimento está dentro dos limites esperados."
	default:
		return ". Verifique os dados fornecidos."
	}
}

// ValidationSummaryHook cria um resumo dos resultados de validação
type ValidationSummaryHook struct {
	LogSummary bool
}

// Execute cria um resumo dos erros de validação
func (h *ValidationSummaryHook) Execute(data interface{}, errors []interfaces.ValidationError) ([]interfaces.ValidationError, error) {
	if h.LogSummary && len(errors) > 0 {
		errorsByType := make(map[string]int)
		errorsByField := make(map[string]int)

		for _, err := range errors {
			errorsByType[err.ErrorType]++
			errorsByField[err.Field]++
		}

		fmt.Printf("[VALIDATION] Summary: %d errors found\n", len(errors))
		fmt.Printf("[VALIDATION] Errors by type: %+v\n", errorsByType)
		fmt.Printf("[VALIDATION] Errors by field: %+v\n", errorsByField)
	}

	return errors, nil
}
