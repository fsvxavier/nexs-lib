package checks

import (
	"regexp"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema/interfaces"
)

// StringFormatChecker valida se o valor é uma string
type StringFormatChecker struct {
	AllowEmpty bool
}

// NewStringFormatChecker cria um novo validador de string
func NewStringFormatChecker() *StringFormatChecker {
	return &StringFormatChecker{
		AllowEmpty: true,
	}
}

// IsFormat implementa interfaces.FormatChecker
func (s *StringFormatChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	if !s.AllowEmpty && asString == "" {
		return false
	}

	return true
}

// Check implementa interfaces.Check
func (s *StringFormatChecker) Check(data interface{}) []interfaces.ValidationError {
	if s.IsFormat(data) {
		return nil
	}

	return []interfaces.ValidationError{
		{
			Field:     "string",
			Message:   "Value must be a string",
			ErrorType: "INVALID_STRING_TYPE",
			Value:     data,
		},
	}
}

// NonEmptyStringChecker valida se o valor é uma string não vazia
type NonEmptyStringChecker struct{}

// NewNonEmptyStringChecker cria um novo validador de string não vazia
func NewNonEmptyStringChecker() *NonEmptyStringChecker {
	return &NonEmptyStringChecker{}
}

// IsFormat implementa interfaces.FormatChecker
func (n *NonEmptyStringChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	return asString != ""
}

// Check implementa interfaces.Check
func (n *NonEmptyStringChecker) Check(data interface{}) []interfaces.ValidationError {
	if n.IsFormat(data) {
		return nil
	}

	return []interfaces.ValidationError{
		{
			Field:     "non_empty_string",
			Message:   "String cannot be empty",
			ErrorType: "EMPTY_STRING",
			Value:     data,
		},
	}
}

// TextMatchChecker valida texto com letras, espaços e underscores
type TextMatchChecker struct {
	AllowEmpty bool
}

// NewTextMatchChecker cria um novo validador de texto
func NewTextMatchChecker() *TextMatchChecker {
	return &TextMatchChecker{
		AllowEmpty: false,
	}
}

// IsFormat implementa interfaces.FormatChecker
func (t *TextMatchChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	if t.AllowEmpty && asString == "" {
		return true
	}

	if !t.AllowEmpty && asString == "" {
		return false
	}

	r := regexp.MustCompile("^[a-zA-Z_ ]*$")
	return r.MatchString(asString)
}

// Check implementa interfaces.Check
func (t *TextMatchChecker) Check(data interface{}) []interfaces.ValidationError {
	if t.IsFormat(data) {
		return nil
	}

	return []interfaces.ValidationError{
		{
			Field:     "text_match",
			Message:   "Text must contain only letters, spaces and underscores",
			ErrorType: "INVALID_TEXT_FORMAT",
			Value:     data,
		},
	}
}

// TextMatchWithNumberChecker valida texto com letras, números, espaços e underscores
type TextMatchWithNumberChecker struct {
	AllowEmpty bool
}

// NewTextMatchWithNumberChecker cria um novo validador de texto com números
func NewTextMatchWithNumberChecker() *TextMatchWithNumberChecker {
	return &TextMatchWithNumberChecker{
		AllowEmpty: false,
	}
}

// IsFormat implementa interfaces.FormatChecker
func (t *TextMatchWithNumberChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	if t.AllowEmpty && asString == "" {
		return true
	}

	if !t.AllowEmpty && asString == "" {
		return false
	}

	r := regexp.MustCompile("^[a-zA-Z1-9_ ]*$")
	return r.MatchString(asString)
}

// Check implementa interfaces.Check
func (t *TextMatchWithNumberChecker) Check(data interface{}) []interfaces.ValidationError {
	if t.IsFormat(data) {
		return nil
	}

	return []interfaces.ValidationError{
		{
			Field:     "text_match_with_number",
			Message:   "Text must contain only letters, numbers, spaces and underscores",
			ErrorType: "INVALID_TEXT_WITH_NUMBER_FORMAT",
			Value:     data,
		},
	}
}

// CustomRegexChecker valida texto com regex customizada
type CustomRegexChecker struct {
	Pattern    string
	AllowEmpty bool
	regex      *regexp.Regexp
}

// NewCustomRegexChecker cria um novo validador com regex customizada
func NewCustomRegexChecker(pattern string) (*CustomRegexChecker, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return &CustomRegexChecker{
		Pattern:    pattern,
		AllowEmpty: false,
		regex:      regex,
	}, nil
}

// IsFormat implementa interfaces.FormatChecker
func (c *CustomRegexChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	if c.AllowEmpty && asString == "" {
		return true
	}

	if !c.AllowEmpty && asString == "" {
		return false
	}

	if c.regex == nil {
		// Re-compile if needed
		regex, err := regexp.Compile(c.Pattern)
		if err != nil {
			return false
		}
		c.regex = regex
	}

	return c.regex.MatchString(asString)
}

// Check implementa interfaces.Check
func (c *CustomRegexChecker) Check(data interface{}) []interfaces.ValidationError {
	if c.IsFormat(data) {
		return nil
	}

	return []interfaces.ValidationError{
		{
			Field:     "custom_regex",
			Message:   "Text does not match required pattern: " + c.Pattern,
			ErrorType: "INVALID_REGEX_PATTERN",
			Value:     data,
		},
	}
}

// StrongNameFormatChecker valida nomes em formato "forte" (maiúsculas e underscores)
type StrongNameFormatChecker struct {
	AllowEmpty bool
}

// NewStrongNameFormatChecker cria um novo validador de nome forte
func NewStrongNameFormatChecker() *StrongNameFormatChecker {
	return &StrongNameFormatChecker{
		AllowEmpty: false,
	}
}

// IsFormat implementa interfaces.FormatChecker
func (s *StrongNameFormatChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	if s.AllowEmpty && asString == "" {
		return true
	}

	if !s.AllowEmpty && asString == "" {
		return false
	}

	strongNameRegex := regexp.MustCompile("^[A-Z_]*$")
	return strongNameRegex.MatchString(asString)
}

// Check implementa interfaces.Check
func (s *StrongNameFormatChecker) Check(data interface{}) []interfaces.ValidationError {
	if s.IsFormat(data) {
		return nil
	}

	return []interfaces.ValidationError{
		{
			Field:     "strong_name",
			Message:   "Name must contain only uppercase letters and underscores",
			ErrorType: "INVALID_STRONG_NAME_FORMAT",
			Value:     data,
		},
	}
}
