package checks

import (
	"github.com/fsvxavier/nexs-lib/validation/jsonschema/interfaces"
)

// EnumConstraintsCheck valida se valores estão dentro de enumerações permitidas
type EnumConstraintsCheck struct {
	Constraints map[string][]interface{}
}

// Validate verifica se valores estão dentro dos enums permitidos
func (c *EnumConstraintsCheck) Validate(data interface{}) []interfaces.ValidationError {
	var errors []interfaces.ValidationError

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errors
	}

	for field, allowedValues := range c.Constraints {
		if value, exists := dataMap[field]; exists {
			if !c.isValueInEnum(value, allowedValues) {
				errors = append(errors, interfaces.ValidationError{
					Field:       field,
					Message:     "Value is not in allowed enumeration",
					ErrorType:   "INVALID_VALUE",
					Value:       value,
					Description: "Value must be one of the allowed values",
				})
			}
		}
	}

	return errors
}

// GetName retorna o nome do check
func (c *EnumConstraintsCheck) GetName() string {
	return "enum_constraints"
}

func (c *EnumConstraintsCheck) isValueInEnum(value interface{}, allowedValues []interface{}) bool {
	for _, allowed := range allowedValues {
		if c.valuesEqual(value, allowed) {
			return true
		}
	}
	return false
}

func (c *EnumConstraintsCheck) valuesEqual(a, b interface{}) bool {
	return a == b
}

// RangeConstraintsCheck valida se valores numéricos estão dentro de intervalos
type RangeConstraintsCheck struct {
	Constraints map[string]RangeConstraint
}

// RangeConstraint define um intervalo numérico
type RangeConstraint struct {
	Min          *float64
	Max          *float64
	ExclusiveMin bool
	ExclusiveMax bool
}

// Validate verifica se valores estão dentro dos intervalos permitidos
func (c *RangeConstraintsCheck) Validate(data interface{}) []interfaces.ValidationError {
	var errors []interfaces.ValidationError

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errors
	}

	for field, constraint := range c.Constraints {
		if value, exists := dataMap[field]; exists {
			if numValue, isNum := c.toFloat64(value); isNum {
				if !c.isValueInRange(numValue, constraint) {
					errors = append(errors, interfaces.ValidationError{
						Field:       field,
						Message:     "Value is outside allowed range",
						ErrorType:   "INVALID_VALUE",
						Value:       value,
						Description: "Value must be within the specified range",
					})
				}
			}
		}
	}

	return errors
}

// GetName retorna o nome do check
func (c *RangeConstraintsCheck) GetName() string {
	return "range_constraints"
}

func (c *RangeConstraintsCheck) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func (c *RangeConstraintsCheck) isValueInRange(value float64, constraint RangeConstraint) bool {
	if constraint.Min != nil {
		if constraint.ExclusiveMin {
			if value <= *constraint.Min {
				return false
			}
		} else {
			if value < *constraint.Min {
				return false
			}
		}
	}

	if constraint.Max != nil {
		if constraint.ExclusiveMax {
			if value >= *constraint.Max {
				return false
			}
		} else {
			if value > *constraint.Max {
				return false
			}
		}
	}

	return true
}

// LengthConstraintsCheck valida comprimentos de strings e arrays
type LengthConstraintsCheck struct {
	Constraints map[string]LengthConstraint
}

// LengthConstraint define restrições de comprimento
type LengthConstraint struct {
	MinLength *int
	MaxLength *int
}

// Validate verifica se comprimentos estão dentro dos limites
func (c *LengthConstraintsCheck) Validate(data interface{}) []interfaces.ValidationError {
	var errors []interfaces.ValidationError

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errors
	}

	for field, constraint := range c.Constraints {
		if value, exists := dataMap[field]; exists {
			length := c.getLength(value)
			if length >= 0 && !c.isLengthValid(length, constraint) {
				errors = append(errors, interfaces.ValidationError{
					Field:       field,
					Message:     "Value length is outside allowed range",
					ErrorType:   "INVALID_LENGTH",
					Value:       value,
					Description: "Value length must be within the specified range",
				})
			}
		}
	}

	return errors
}

// GetName retorna o nome do check
func (c *LengthConstraintsCheck) GetName() string {
	return "length_constraints"
}

func (c *LengthConstraintsCheck) getLength(value interface{}) int {
	switch v := value.(type) {
	case string:
		return len(v)
	case []interface{}:
		return len(v)
	default:
		return -1 // Indica que não é um tipo com comprimento
	}
}

func (c *LengthConstraintsCheck) isLengthValid(length int, constraint LengthConstraint) bool {
	if constraint.MinLength != nil && length < *constraint.MinLength {
		return false
	}

	if constraint.MaxLength != nil && length > *constraint.MaxLength {
		return false
	}

	return true
}
