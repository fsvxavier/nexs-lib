package checks

import (
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/interfaces"
)

// RequiredFieldsCheck valida se campos obrigatórios estão presentes
type RequiredFieldsCheck struct {
	RequiredFields []string
}

// Validate verifica se todos os campos obrigatórios estão presentes
func (c *RequiredFieldsCheck) Validate(data interface{}) []interfaces.ValidationError {
	var errors []interfaces.ValidationError

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errors
	}

	for _, field := range c.RequiredFields {
		if value, exists := dataMap[field]; !exists || value == nil {
			errors = append(errors, interfaces.ValidationError{
				Field:     field,
				Message:   "Required field is missing",
				ErrorType: "REQUIRED_ATTRIBUTE_MISSING",
			})
		} else {
			// Verifica se é string vazia
			if str, isString := value.(string); isString && str == "" {
				errors = append(errors, interfaces.ValidationError{
					Field:     field,
					Message:   "Required field cannot be empty",
					ErrorType: "REQUIRED_ATTRIBUTE_MISSING",
					Value:     value,
				})
			}
		}
	}

	return errors
}

// GetName retorna o nome do check
func (c *RequiredFieldsCheck) GetName() string {
	return "required_fields"
}

// NotEmptyFieldsCheck valida se campos específicos não estão vazios
type NotEmptyFieldsCheck struct {
	Fields []string
}

// Validate verifica se campos especificados não estão vazios
func (c *NotEmptyFieldsCheck) Validate(data interface{}) []interfaces.ValidationError {
	var errors []interfaces.ValidationError

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errors
	}

	for _, field := range c.Fields {
		if value, exists := dataMap[field]; exists {
			if c.isEmpty(value) {
				errors = append(errors, interfaces.ValidationError{
					Field:     field,
					Message:   "Field cannot be empty",
					ErrorType: "INVALID_VALUE",
					Value:     value,
				})
			}
		}
	}

	return errors
}

// GetName retorna o nome do check
func (c *NotEmptyFieldsCheck) GetName() string {
	return "not_empty_fields"
}

func (c *NotEmptyFieldsCheck) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

// ConditionalRequiredCheck valida campos obrigatórios baseado em condições
type ConditionalRequiredCheck struct {
	Rules []ConditionalRule
}

// ConditionalRule define uma regra condicional
type ConditionalRule struct {
	ConditionField string
	ConditionValue interface{}
	RequiredFields []string
}

// Validate aplica regras condicionais
func (c *ConditionalRequiredCheck) Validate(data interface{}) []interfaces.ValidationError {
	var errors []interfaces.ValidationError

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errors
	}

	for _, rule := range c.Rules {
		if conditionValue, exists := dataMap[rule.ConditionField]; exists {
			if c.valuesEqual(conditionValue, rule.ConditionValue) {
				// Condição atendida, verifica campos obrigatórios
				for _, requiredField := range rule.RequiredFields {
					if value, fieldExists := dataMap[requiredField]; !fieldExists || value == nil {
						errors = append(errors, interfaces.ValidationError{
							Field:       requiredField,
							Message:     "Field is required when " + rule.ConditionField + " is present",
							ErrorType:   "REQUIRED_ATTRIBUTE_MISSING",
							Description: "Conditional requirement based on " + rule.ConditionField,
						})
					}
				}
			}
		}
	}

	return errors
}

// GetName retorna o nome do check
func (c *ConditionalRequiredCheck) GetName() string {
	return "conditional_required"
}

func (c *ConditionalRequiredCheck) valuesEqual(a, b interface{}) bool {
	// Comparação simples - em uma implementação real seria mais robusta
	return a == b
}
