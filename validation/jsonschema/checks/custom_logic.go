package checks

import (
	"fmt"
	"regexp"
	"time"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema/interfaces"
)

// BusinessLogicCheck implementa validações de regras de negócio customizadas
type BusinessLogicCheck struct {
	Rules []BusinessRule
}

// BusinessRule define uma regra de negócio customizada
type BusinessRule struct {
	Name        string
	Description string
	Validator   func(data interface{}) []interfaces.ValidationError
}

// Validate executa todas as regras de negócio configuradas
func (c *BusinessLogicCheck) Validate(data interface{}) []interfaces.ValidationError {
	var errors []interfaces.ValidationError

	for _, rule := range c.Rules {
		if ruleErrors := rule.Validator(data); len(ruleErrors) > 0 {
			errors = append(errors, ruleErrors...)
		}
	}

	return errors
}

// GetName retorna o nome do check
func (c *BusinessLogicCheck) GetName() string {
	return "business_logic"
}

// DateValidationCheck implementa validações específicas para datas
type DateValidationCheck struct {
	DateFields  []string
	MinDate     *time.Time
	MaxDate     *time.Time
	AllowFuture bool
	AllowPast   bool
	DateFormats []string
}

// Validate executa validações específicas para campos de data
func (c *DateValidationCheck) Validate(data interface{}) []interfaces.ValidationError {
	var errors []interfaces.ValidationError

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errors
	}

	if len(c.DateFormats) == 0 {
		c.DateFormats = []string{
			time.RFC3339,
			"2006-01-02",
			"2006-01-02T15:04:05",
		}
	}

	for _, field := range c.DateFields {
		if value, exists := dataMap[field]; exists {
			if dateStr, ok := value.(string); ok {
				if date, err := c.parseDate(dateStr); err == nil {
					if !c.isDateValid(date) {
						errors = append(errors, interfaces.ValidationError{
							Field:       field,
							Message:     "Date is outside allowed range or constraints",
							ErrorType:   "INVALID_VALUE",
							Value:       value,
							Description: "Date validation failed",
						})
					}
				} else {
					errors = append(errors, interfaces.ValidationError{
						Field:       field,
						Message:     "Invalid date format",
						ErrorType:   "INVALID_FORMAT",
						Value:       value,
						Description: "Date could not be parsed",
					})
				}
			}
		}
	}

	return errors
}

// GetName retorna o nome do check
func (c *DateValidationCheck) GetName() string {
	return "date_validation"
}

func (c *DateValidationCheck) parseDate(dateStr string) (time.Time, error) {
	for _, format := range c.DateFormats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func (c *DateValidationCheck) isDateValid(date time.Time) bool {
	now := time.Now()

	if !c.AllowFuture && date.After(now) {
		return false
	}

	if !c.AllowPast && date.Before(now) {
		return false
	}

	if c.MinDate != nil && date.Before(*c.MinDate) {
		return false
	}

	if c.MaxDate != nil && date.After(*c.MaxDate) {
		return false
	}

	return true
}

// RegexValidationCheck implementa validações baseadas em expressões regulares
type RegexValidationCheck struct {
	Patterns map[string]*regexp.Regexp
}

// Validate executa validações baseadas em regex
func (c *RegexValidationCheck) Validate(data interface{}) []interfaces.ValidationError {
	var errors []interfaces.ValidationError

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errors
	}

	for field, pattern := range c.Patterns {
		if value, exists := dataMap[field]; exists {
			if str, ok := value.(string); ok {
				if !pattern.MatchString(str) {
					errors = append(errors, interfaces.ValidationError{
						Field:       field,
						Message:     "Value does not match required pattern",
						ErrorType:   "INVALID_FORMAT",
						Value:       value,
						Description: fmt.Sprintf("Value must match pattern: %s", pattern.String()),
					})
				}
			}
		}
	}

	return errors
}

// GetName retorna o nome do check
func (c *RegexValidationCheck) GetName() string {
	return "regex_validation"
}

// CrossFieldValidationCheck implementa validações entre campos
type CrossFieldValidationCheck struct {
	Rules []CrossFieldRule
}

// CrossFieldRule define uma regra de validação cruzada entre campos
type CrossFieldRule struct {
	Name        string
	Description string
	Validator   func(data map[string]interface{}) []interfaces.ValidationError
}

// Validate executa validações cruzadas entre campos
func (c *CrossFieldValidationCheck) Validate(data interface{}) []interfaces.ValidationError {
	var errors []interfaces.ValidationError

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errors
	}

	for _, rule := range c.Rules {
		if ruleErrors := rule.Validator(dataMap); len(ruleErrors) > 0 {
			errors = append(errors, ruleErrors...)
		}
	}

	return errors
}

// GetName retorna o nome do check
func (c *CrossFieldValidationCheck) GetName() string {
	return "cross_field_validation"
}
