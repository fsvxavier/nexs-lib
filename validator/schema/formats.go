package schema

import (
	"regexp"

	"github.com/fsvxavier/nexs-lib/validator/schema/checks"
)

// formatValidatorAdapter adapts checkers to FormatValidator interface
type formatValidatorAdapter struct {
	checker checks.FormatChecker
}

func (fva *formatValidatorAdapter) IsFormat(input interface{}) bool {
	return fva.checker.IsFormat(input)
}

func (fva *formatValidatorAdapter) FormatName() string {
	return fva.checker.FormatName()
}

// customRegexFormatValidator allows custom regex validation
type customRegexFormatValidator struct {
	name    string
	pattern *regexp.Regexp
}

func NewCustomRegexFormatValidator(name, pattern string) FormatValidator {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		// Return a validator that always fails if pattern is invalid
		return &funcFormatValidator{
			name: name,
			validateFunc: func(interface{}) bool {
				return false
			},
		}
	}

	return &customRegexFormatValidator{
		name:    name,
		pattern: regex,
	}
}

func (crfv *customRegexFormatValidator) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	return crfv.pattern.MatchString(asString)
}

func (crfv *customRegexFormatValidator) FormatName() string {
	return crfv.name
}

// NewFormatValidatorAdapter creates an adapter for legacy format checkers
func NewFormatValidatorAdapter(name string, isFormatFunc func(interface{}) bool) FormatValidator {
	return &funcFormatValidator{
		name:         name,
		validateFunc: isFormatFunc,
	}
}

// Factory functions using checks

// NewDateTimeFormatValidator creates a datetime format validator
func NewDateTimeFormatValidator() FormatValidator {
	return &formatValidatorAdapter{
		checker: &checks.DateTimeChecker{},
	}
}

// NewISO8601DateFormatValidator creates an ISO 8601 date format validator
func NewISO8601DateFormatValidator() FormatValidator {
	return &formatValidatorAdapter{
		checker: &checks.ISO8601DateChecker{},
	}
}

// NewJSONNumberFormatValidator creates a JSON number format validator
func NewJSONNumberFormatValidator() FormatValidator {
	return &formatValidatorAdapter{
		checker: &checks.JSONNumberChecker{},
	}
}

// NewEmptyStringFormatValidator creates an empty string format validator
func NewEmptyStringFormatValidator() FormatValidator {
	return &formatValidatorAdapter{
		checker: &checks.EmptyStringChecker{},
	}
}

// NewStringFormatValidator creates a string format validator
func NewStringFormatValidator() FormatValidator {
	return &formatValidatorAdapter{
		checker: &checks.StringChecker{},
	}
}

// NewStrongNameFormatValidator creates a strong name format validator
func NewStrongNameFormatValidator() FormatValidator {
	return &formatValidatorAdapter{
		checker: &checks.StrongNameChecker{},
	}
}

// NewTextMatchFormatValidator creates a text match format validator
func NewTextMatchFormatValidator() FormatValidator {
	return &formatValidatorAdapter{
		checker: &checks.TextMatchChecker{},
	}
}

// NewTextMatchWithNumberFormatValidator creates a text match with number format validator
func NewTextMatchWithNumberFormatValidator() FormatValidator {
	return &formatValidatorAdapter{
		checker: &checks.TextMatchWithNumberChecker{},
	}
}

// NewTextMatchCustomFormatValidator creates a custom text match format validator
func NewTextMatchCustomFormatValidator(regex string) FormatValidator {
	return &formatValidatorAdapter{
		checker: checks.NewTextMatchCustom(regex),
	}
}

// NewDecimalFormatValidator creates a decimal format validator
func NewDecimalFormatValidator() FormatValidator {
	return &formatValidatorAdapter{
		checker: checks.NewDecimal(),
	}
}

// NewDecimalByFactor8FormatValidator creates a decimal by factor 8 format validator
func NewDecimalByFactor8FormatValidator() FormatValidator {
	return &formatValidatorAdapter{
		checker: checks.NewDecimalByFactor8(),
	}
}
