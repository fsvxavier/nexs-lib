package domainerrors

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fsvxavier/nexs-lib/v2/domainerrors/interfaces"
	"github.com/fsvxavier/nexs-lib/v2/domainerrors/types"
)

// ValidationError implementa interfaces.ValidationErrorInterface para erros de validação específicos.
type ValidationError struct {
	*DomainError
	ValidatedFields map[string][]string `json:"validated_fields,omitempty"`
}

// NewValidationError cria um novo erro de validação.
func NewValidationError(message string, fields map[string][]string) interfaces.ValidationErrorInterface {
	domainErr := newDomainError()
	domainErr.code = "VALIDATION_ERROR"
	domainErr.message = message
	domainErr.errorType = types.ErrorTypeValidation
	domainErr.severity = types.SeverityLow
	domainErr.category = types.ErrorTypeValidation.Category()
	domainErr.statusCode = types.ErrorTypeValidation.DefaultStatusCode()

	if fields == nil {
		fields = make(map[string][]string)
	}

	return &ValidationError{
		DomainError:     domainErr,
		ValidatedFields: fields,
	}
}

// Fields retorna os campos com erro.
func (e *ValidationError) Fields() map[string][]string {
	if e.ValidatedFields == nil {
		return make(map[string][]string)
	}

	// Retorna cópia para evitar modificações externas
	result := make(map[string][]string)
	for k, v := range e.ValidatedFields {
		messages := make([]string, len(v))
		copy(messages, v)
		result[k] = messages
	}

	return result
}

// AddField adiciona um erro de campo.
func (e *ValidationError) AddField(field, message string) interfaces.ValidationErrorInterface {
	if e.ValidatedFields == nil {
		e.ValidatedFields = make(map[string][]string)
	}

	e.ValidatedFields[field] = append(e.ValidatedFields[field], message)
	return e
}

// AddFields adiciona múltiplos erros de campo.
func (e *ValidationError) AddFields(fields map[string][]string) interfaces.ValidationErrorInterface {
	if e.ValidatedFields == nil {
		e.ValidatedFields = make(map[string][]string)
	}

	for field, messages := range fields {
		e.ValidatedFields[field] = append(e.ValidatedFields[field], messages...)
	}

	return e
}

// HasField verifica se um campo tem erro.
func (e *ValidationError) HasField(field string) bool {
	if e.ValidatedFields == nil {
		return false
	}

	errors, exists := e.ValidatedFields[field]
	return exists && len(errors) > 0
}

// FieldErrors retorna os erros de um campo específico.
func (e *ValidationError) FieldErrors(field string) []string {
	if e.ValidatedFields == nil {
		return nil
	}

	errors, exists := e.ValidatedFields[field]
	if !exists {
		return nil
	}

	// Retorna cópia para evitar modificações externas
	result := make([]string, len(errors))
	copy(result, errors)
	return result
}

// Error implementa a interface error com informações de validação.
func (e *ValidationError) Error() string {
	var b strings.Builder
	b.Grow(256)

	// Adiciona erro base
	if e.DomainError != nil {
		b.WriteString(e.DomainError.Error())
	} else {
		b.WriteString("Validation failed")
	}

	// Adiciona informações de campos se disponíveis
	if len(e.ValidatedFields) > 0 {
		b.WriteString(" - Fields: ")
		fieldCount := 0
		for field, messages := range e.ValidatedFields {
			if fieldCount > 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%s (%d errors)", field, len(messages)))
			fieldCount++
		}
	}

	return b.String()
}

// JSON retorna uma representação JSON específica para erros de validação.
func (e *ValidationError) JSON() ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Estrutura específica para ValidationError
	data := struct {
		Code            string                 `json:"code"`
		Message         string                 `json:"message"`
		Type            string                 `json:"type"`
		Severity        string                 `json:"severity"`
		ValidatedFields map[string][]string    `json:"validated_fields,omitempty"`
		Details         map[string]interface{} `json:"details,omitempty"`
		Tags            []string               `json:"tags,omitempty"`
		Timestamp       string                 `json:"timestamp"`
	}{
		Code:            e.code,
		Message:         e.message,
		Type:            e.errorType.String(),
		Severity:        e.severity.String(),
		ValidatedFields: e.ValidatedFields,
		Details:         e.details,
		Tags:            e.tags,
		Timestamp:       e.timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}

	return json.Marshal(data)
}

// DetailedString retorna uma string detalhada incluindo campos de validação.
func (e *ValidationError) DetailedString() string {
	var b strings.Builder
	b.Grow(512)

	// Adiciona detalhes base
	if e.DomainError != nil {
		b.WriteString(e.DomainError.DetailedString())
	}

	// Adiciona detalhes de validação
	if len(e.ValidatedFields) > 0 {
		b.WriteString("\nValidation Errors:\n")
		for field, messages := range e.ValidatedFields {
			b.WriteString(fmt.Sprintf("  %s:\n", field))
			for _, msg := range messages {
				b.WriteString(fmt.Sprintf("    - %s\n", msg))
			}
		}
	}

	return b.String()
}

// Clone cria uma cópia independente do ValidationError.
func (e *ValidationError) Clone() *ValidationError {
	clone := &ValidationError{
		DomainError:     e.DomainError.Clone(),
		ValidatedFields: make(map[string][]string),
	}

	// Copia campos de validação
	for field, messages := range e.ValidatedFields {
		fieldMessages := make([]string, len(messages))
		copy(fieldMessages, messages)
		clone.ValidatedFields[field] = fieldMessages
	}

	return clone
}

// TotalErrors retorna o número total de erros de validação.
func (e *ValidationError) TotalErrors() int {
	total := 0
	for _, messages := range e.ValidatedFields {
		total += len(messages)
	}
	return total
}

// FirstError retorna o primeiro erro encontrado (útil para logs simples).
func (e *ValidationError) FirstError() string {
	for _, messages := range e.ValidatedFields {
		if len(messages) > 0 {
			return messages[0]
		}
	}
	return e.Message()
}

// FieldNames retorna uma lista de todos os campos com erro.
func (e *ValidationError) FieldNames() []string {
	names := make([]string, 0, len(e.ValidatedFields))
	for field := range e.ValidatedFields {
		names = append(names, field)
	}
	return names
}

// IsEmpty verifica se o erro de validação está vazio (sem campos).
func (e *ValidationError) IsEmpty() bool {
	return len(e.ValidatedFields) == 0
}

// Merge combina este ValidationError com outro.
func (e *ValidationError) Merge(other interfaces.ValidationErrorInterface) interfaces.ValidationErrorInterface {
	if other == nil {
		return e
	}

	otherFields := other.Fields()
	for field, messages := range otherFields {
		for _, message := range messages {
			e.AddField(field, message)
		}
	}

	return e
}

// WithFieldPrefix adiciona um prefixo a todos os nomes de campo.
func (e *ValidationError) WithFieldPrefix(prefix string) interfaces.ValidationErrorInterface {
	if prefix == "" {
		return e
	}

	newFields := make(map[string][]string)
	for field, messages := range e.ValidatedFields {
		newField := prefix + "." + field
		newFields[newField] = messages
	}

	e.ValidatedFields = newFields
	return e
}

// ResponseBody retorna o corpo da resposta específico para ValidationError.
func (e *ValidationError) ResponseBody() interface{} {
	jsonData, _ := e.JSON()
	return json.RawMessage(jsonData)
}
