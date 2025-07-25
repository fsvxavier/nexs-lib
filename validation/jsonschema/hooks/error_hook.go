package hooks

import (
	"fmt"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema/interfaces"
)

// ErrorFilterHook filtra erros baseado em critérios específicos
type ErrorFilterHook struct {
	IgnoreFields []string
	MaxErrors    int
}

// Execute filtra erros de validação baseado na configuração
func (h *ErrorFilterHook) Execute(errors []interfaces.ValidationError) []interfaces.ValidationError {
	if len(errors) == 0 {
		return errors
	}

	filteredErrors := make([]interfaces.ValidationError, 0, len(errors))

	for _, err := range errors {
		// Ignora campos específicos se configurado
		if h.shouldIgnoreField(err.Field) {
			continue
		}

		filteredErrors = append(filteredErrors, err)

		// Limita número máximo de erros se configurado
		if h.MaxErrors > 0 && len(filteredErrors) >= h.MaxErrors {
			break
		}
	}

	return filteredErrors
}

func (h *ErrorFilterHook) shouldIgnoreField(field string) bool {
	for _, ignoredField := range h.IgnoreFields {
		if field == ignoredField {
			return true
		}
	}
	return false
}

// ErrorNotificationHook notifica sistemas externos sobre erros
type ErrorNotificationHook struct {
	NotifyOnError bool
	WebhookURL    string
}

// Execute notifica sistemas externos quando há erros
func (h *ErrorNotificationHook) Execute(errors []interfaces.ValidationError) []interfaces.ValidationError {
	if h.NotifyOnError && len(errors) > 0 {
		h.sendNotification(errors)
	}
	return errors
}

func (h *ErrorNotificationHook) sendNotification(errors []interfaces.ValidationError) {
	// Em uma implementação real, enviaria para webhook/sistema externo
	fmt.Printf("[NOTIFICATION] %d validation errors occurred\n", len(errors))
	if h.WebhookURL != "" {
		fmt.Printf("[NOTIFICATION] Would send to webhook: %s\n", h.WebhookURL)
	}
}

// ErrorAggregationHook agrega erros similares
type ErrorAggregationHook struct {
	AggregateByType  bool
	AggregateByField bool
}

// Execute agrega erros similares para reduzir redundância
func (h *ErrorAggregationHook) Execute(errors []interfaces.ValidationError) []interfaces.ValidationError {
	if len(errors) == 0 {
		return errors
	}

	if h.AggregateByType {
		return h.aggregateByType(errors)
	}

	if h.AggregateByField {
		return h.aggregateByField(errors)
	}

	return errors
}

func (h *ErrorAggregationHook) aggregateByType(errors []interfaces.ValidationError) []interfaces.ValidationError {
	typeMap := make(map[string][]interfaces.ValidationError)

	for _, err := range errors {
		typeMap[err.ErrorType] = append(typeMap[err.ErrorType], err)
	}

	aggregated := make([]interfaces.ValidationError, 0, len(typeMap))
	for errorType, errs := range typeMap {
		if len(errs) == 1 {
			aggregated = append(aggregated, errs[0])
		} else {
			fields := make([]string, len(errs))
			for i, err := range errs {
				fields[i] = err.Field
			}

			aggregated = append(aggregated, interfaces.ValidationError{
				Field:     fmt.Sprintf("multiple_fields_%s", errorType),
				Message:   fmt.Sprintf("%d fields with %s errors: %v", len(errs), errorType, fields),
				ErrorType: errorType,
			})
		}
	}

	return aggregated
}

func (h *ErrorAggregationHook) aggregateByField(errors []interfaces.ValidationError) []interfaces.ValidationError {
	fieldMap := make(map[string][]interfaces.ValidationError)

	for _, err := range errors {
		fieldMap[err.Field] = append(fieldMap[err.Field], err)
	}

	aggregated := make([]interfaces.ValidationError, 0, len(fieldMap))
	for field, errs := range fieldMap {
		if len(errs) == 1 {
			aggregated = append(aggregated, errs[0])
		} else {
			types := make([]string, len(errs))
			for i, err := range errs {
				types[i] = err.ErrorType
			}

			aggregated = append(aggregated, interfaces.ValidationError{
				Field:     field,
				Message:   fmt.Sprintf("Multiple validation errors: %v", types),
				ErrorType: "MULTIPLE_ERRORS",
			})
		}
	}

	return aggregated
}
