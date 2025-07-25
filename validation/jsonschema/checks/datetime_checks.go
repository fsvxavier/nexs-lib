package checks

import (
	"time"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema/interfaces"
)

// DateTimeFormatChecker valida formatos de data e hora
type DateTimeFormatChecker struct {
	// Formats especifica os formatos aceitos. Se vazio, usa os formatos padrão
	Formats []string

	// AllowEmpty permite valores vazios
	AllowEmpty bool
}

// NewDateTimeFormatChecker cria um novo validador de formato de data/hora
func NewDateTimeFormatChecker() *DateTimeFormatChecker {
	return &DateTimeFormatChecker{
		Formats:    CommonDateTimeFormats,
		AllowEmpty: true,
	}
}

// NewDateTimeFormatCheckerWithFormats cria um validador com formatos customizados
func NewDateTimeFormatCheckerWithFormats(formats []string) *DateTimeFormatChecker {
	return &DateTimeFormatChecker{
		Formats:    formats,
		AllowEmpty: false,
	}
}

// WithFormats define os formatos aceitos (builder pattern)
func (d *DateTimeFormatChecker) WithFormats(formats []string) *DateTimeFormatChecker {
	d.Formats = formats
	return d
}

// IsFormat implementa interfaces.FormatChecker
func (d *DateTimeFormatChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	// Se permitir vazio e a string for vazia
	if d.AllowEmpty && asString == "" {
		return true
	}

	// Se não permitir vazio e a string for vazia
	if !d.AllowEmpty && asString == "" {
		return false
	}

	formats := d.Formats
	if len(formats) == 0 {
		formats = CommonDateTimeFormats
	}

	for _, format := range formats {
		if _, err := time.Parse(format, asString); err == nil {
			return true
		}
	}

	return false
}

// Implementação compatível com interface Check
func (d *DateTimeFormatChecker) Check(data interface{}) []interfaces.ValidationError {
	if d.IsFormat(data) {
		return nil
	}

	return []interfaces.ValidationError{
		{
			Field:     "datetime",
			Message:   "Invalid date/time format",
			ErrorType: "INVALID_DATETIME_FORMAT",
			Value:     data,
		},
	}
}

// ISO8601DateChecker valida especificamente formato ISO8601 de data
type ISO8601DateChecker struct {
	AllowEmpty bool
}

// NewISO8601DateChecker cria um novo validador ISO8601
func NewISO8601DateChecker() *ISO8601DateChecker {
	return &ISO8601DateChecker{
		AllowEmpty: true,
	}
}

// IsFormat implementa interfaces.FormatChecker
func (i *ISO8601DateChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	if i.AllowEmpty && asString == "" {
		return true
	}

	if !i.AllowEmpty && asString == "" {
		return false
	}

	_, err := time.Parse(ISO8601DateTimeFormat, asString)
	return err == nil
}

// Check implementa interfaces.Check
func (i *ISO8601DateChecker) Check(data interface{}) []interfaces.ValidationError {
	if i.IsFormat(data) {
		return nil
	}

	return []interfaces.ValidationError{
		{
			Field:     "iso8601_date",
			Message:   "Invalid ISO8601 date format",
			ErrorType: "INVALID_ISO8601_DATE",
			Value:     data,
		},
	}
}

// TimeOnlyChecker valida apenas formatos de hora
type TimeOnlyChecker struct {
	// AllowEmpty permite valores vazios
	AllowEmpty bool
}

// NewTimeOnlyChecker cria um novo validador de hora
func NewTimeOnlyChecker() *TimeOnlyChecker {
	return &TimeOnlyChecker{
		AllowEmpty: true,
	}
}

// IsFormat implementa interfaces.FormatChecker
func (t *TimeOnlyChecker) IsFormat(input interface{}) bool {
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

	timeFormats := []string{
		time.TimeOnly,         // 15:04:05
		RFC3339TimeOnlyFormat, // 15:04:05-07:00
		ISO8601TimeFormat,     // 15:04:05
	}

	for _, format := range timeFormats {
		if _, err := time.Parse(format, asString); err == nil {
			return true
		}
	}

	return false
}

// Check implementa interfaces.Check
func (t *TimeOnlyChecker) Check(data interface{}) []interfaces.ValidationError {
	if t.IsFormat(data) {
		return nil
	}

	return []interfaces.ValidationError{
		{
			Field:     "time",
			ErrorType: "INVALID_TIME_FORMAT",
			Message:   "Value must be a valid time format",
			Value:     data,
		},
	}
}

// DateOnlyChecker valida apenas formatos de data
type DateOnlyChecker struct {
	// AllowEmpty permite valores vazios
	AllowEmpty bool
}

// NewDateOnlyChecker cria um novo validador de data
func NewDateOnlyChecker() *DateOnlyChecker {
	return &DateOnlyChecker{
		AllowEmpty: true,
	}
}

// IsFormat implementa interfaces.FormatChecker
func (d *DateOnlyChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	if d.AllowEmpty && asString == "" {
		return true
	}

	if !d.AllowEmpty && asString == "" {
		return false
	}

	// Apenas formato de data, sem horário
	if _, err := time.Parse(time.DateOnly, asString); err == nil {
		return true
	}

	return false
}

// Check implementa interfaces.Check
func (d *DateOnlyChecker) Check(data interface{}) []interfaces.ValidationError {
	if d.IsFormat(data) {
		return nil
	}

	return []interfaces.ValidationError{
		{
			Field:     "date",
			ErrorType: "INVALID_DATE_FORMAT",
			Message:   "Value must be a valid date format (YYYY-MM-DD)",
			Value:     data,
		},
	}
}
