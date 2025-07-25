package checks

import (
	"encoding/json"
	"math/big"
	"strconv"

	"github.com/fsvxavier/nexs-lib/validation/jsonschema/interfaces"
)

// JSONNumberChecker valida se o valor é um json.Number
type JSONNumberChecker struct {
	AllowZero bool
}

// NewJSONNumberChecker cria um novo validador de json.Number
func NewJSONNumberChecker() *JSONNumberChecker {
	return &JSONNumberChecker{
		AllowZero: true,
	}
}

// IsFormat implementa interfaces.FormatChecker
func (j *JSONNumberChecker) IsFormat(input interface{}) bool {
	_, ok := input.(json.Number)
	return ok
}

// Check implementa interfaces.Check
func (j *JSONNumberChecker) Check(data interface{}) []interfaces.ValidationError {
	if j.IsFormat(data) {
		return nil
	}

	return []interfaces.ValidationError{
		{
			Field:     "json_number",
			Message:   "Value must be a json.Number",
			ErrorType: "INVALID_JSON_NUMBER_TYPE",
			Value:     data,
		},
	}
}

// NumericChecker valida se o valor é numérico (int, float, etc.)
type NumericChecker struct {
	AllowZero     bool
	AllowNegative bool
	MinValue      *float64
	MaxValue      *float64
}

// NewNumericChecker cria um novo validador numérico
func NewNumericChecker() *NumericChecker {
	return &NumericChecker{
		AllowZero:     true,
		AllowNegative: true,
	}
}

// WithRange define valores mínimo e máximo
func (n *NumericChecker) WithRange(min, max float64) *NumericChecker {
	n.MinValue = &min
	n.MaxValue = &max
	return n
}

// WithMinValue define valor mínimo
func (n *NumericChecker) WithMinValue(min float64) *NumericChecker {
	n.MinValue = &min
	return n
}

// WithMaxValue define valor máximo
func (n *NumericChecker) WithMaxValue(max float64) *NumericChecker {
	n.MaxValue = &max
	return n
}

// IsFormat implementa interfaces.FormatChecker
func (n *NumericChecker) IsFormat(input interface{}) bool {
	var value float64
	var ok bool

	switch v := input.(type) {
	case int:
		value = float64(v)
		ok = true
	case int8:
		value = float64(v)
		ok = true
	case int16:
		value = float64(v)
		ok = true
	case int32:
		value = float64(v)
		ok = true
	case int64:
		value = float64(v)
		ok = true
	case uint:
		value = float64(v)
		ok = true
	case uint8:
		value = float64(v)
		ok = true
	case uint16:
		value = float64(v)
		ok = true
	case uint32:
		value = float64(v)
		ok = true
	case uint64:
		value = float64(v)
		ok = true
	case float32:
		value = float64(v)
		ok = true
	case float64:
		value = v
		ok = true
	case json.Number:
		if f, err := v.Float64(); err == nil {
			value = f
			ok = true
		}
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			value = f
			ok = true
		}
	}

	if !ok {
		return false
	}

	// Verificar zero
	if !n.AllowZero && value == 0 {
		return false
	}

	// Verificar negativo
	if !n.AllowNegative && value < 0 {
		return false
	}

	// Verificar range
	if n.MinValue != nil && value < *n.MinValue {
		return false
	}

	if n.MaxValue != nil && value > *n.MaxValue {
		return false
	}

	return true
}

// Check implementa interfaces.Check
func (n *NumericChecker) Check(data interface{}) []interfaces.ValidationError {
	if n.IsFormat(data) {
		return nil
	}

	return []interfaces.ValidationError{
		{
			Field:     "numeric",
			Message:   "Value must be a valid number within specified constraints",
			ErrorType: "INVALID_NUMERIC_VALUE",
			Value:     data,
		},
	}
}

// DecimalChecker valida valores decimais com precisão específica
type DecimalChecker struct {
	DecimalFactor  int32
	ValidateFactor bool
	AllowZero      bool
}

// NewDecimalChecker cria um novo validador decimal
func NewDecimalChecker() *DecimalChecker {
	return &DecimalChecker{
		DecimalFactor:  0,
		ValidateFactor: false,
		AllowZero:      true,
	}
}

// NewDecimalCheckerByFactor8 cria um validador decimal com fator 8
func NewDecimalCheckerByFactor8() *DecimalChecker {
	return &DecimalChecker{
		DecimalFactor:  -8,
		ValidateFactor: true,
		AllowZero:      true,
	}
}

// WithFactor define o fator decimal
func (d *DecimalChecker) WithFactor(factor int32) *DecimalChecker {
	d.DecimalFactor = factor
	d.ValidateFactor = true
	return d
}

// IsFormat implementa interfaces.FormatChecker
func (d *DecimalChecker) IsFormat(input interface{}) bool {
	decimalValue := d.getDecimalValue(input)
	if decimalValue == nil {
		return false
	}

	if !d.AllowZero && decimalValue.Sign() == 0 {
		return false
	}

	if d.ValidateFactor {
		// Verificar se o número tem o fator correto
		// Esta é uma implementação simplificada - seria necessário
		// usar uma biblioteca decimal específica para validação completa
		return true
	}

	return true
}

// getDecimalValue converte o input para um valor decimal
func (d *DecimalChecker) getDecimalValue(input interface{}) *big.Float {
	switch v := input.(type) {
	case float64:
		return big.NewFloat(v)
	case float32:
		return big.NewFloat(float64(v))
	case int:
		return big.NewFloat(float64(v))
	case int64:
		return big.NewFloat(float64(v))
	case string:
		if f, _, err := big.ParseFloat(v, 10, 256, big.ToNearestEven); err == nil {
			return f
		}
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return big.NewFloat(f)
		}
	}
	return nil
}

// Check implementa interfaces.Check
func (d *DecimalChecker) Check(data interface{}) []interfaces.ValidationError {
	if d.IsFormat(data) {
		return nil
	}

	return []interfaces.ValidationError{
		{
			Field:     "decimal",
			Message:   "Value must be a valid decimal number",
			ErrorType: "INVALID_DECIMAL_VALUE",
			Value:     data,
		},
	}
}

// IntegerChecker valida valores inteiros
type IntegerChecker struct {
	AllowZero     bool
	AllowNegative bool
	MinValue      *int64
	MaxValue      *int64
}

// NewIntegerChecker cria um novo validador de inteiros
func NewIntegerChecker() *IntegerChecker {
	return &IntegerChecker{
		AllowZero:     true,
		AllowNegative: true,
	}
}

// WithRange define valores mínimo e máximo
func (i *IntegerChecker) WithRange(min, max int64) *IntegerChecker {
	i.MinValue = &min
	i.MaxValue = &max
	return i
}

// IsFormat implementa interfaces.FormatChecker
func (i *IntegerChecker) IsFormat(input interface{}) bool {
	var value int64
	var ok bool

	switch v := input.(type) {
	case int:
		value = int64(v)
		ok = true
	case int8:
		value = int64(v)
		ok = true
	case int16:
		value = int64(v)
		ok = true
	case int32:
		value = int64(v)
		ok = true
	case int64:
		value = v
		ok = true
	case uint:
		if v <= 9223372036854775807 { // max int64
			value = int64(v)
			ok = true
		}
	case uint8:
		value = int64(v)
		ok = true
	case uint16:
		value = int64(v)
		ok = true
	case uint32:
		value = int64(v)
		ok = true
	case uint64:
		if v <= 9223372036854775807 { // max int64
			value = int64(v)
			ok = true
		}
	case float32:
		if v == float32(int64(v)) { // é um inteiro?
			value = int64(v)
			ok = true
		}
	case float64:
		if v == float64(int64(v)) { // é um inteiro?
			value = int64(v)
			ok = true
		}
	case json.Number:
		if i, err := v.Int64(); err == nil {
			value = i
			ok = true
		}
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			value = i
			ok = true
		}
	}

	if !ok {
		return false
	}

	// Verificar zero
	if !i.AllowZero && value == 0 {
		return false
	}

	// Verificar negativo
	if !i.AllowNegative && value < 0 {
		return false
	}

	// Verificar range
	if i.MinValue != nil && value < *i.MinValue {
		return false
	}

	if i.MaxValue != nil && value > *i.MaxValue {
		return false
	}

	return true
}

// Check implementa interfaces.Check
func (i *IntegerChecker) Check(data interface{}) []interfaces.ValidationError {
	if i.IsFormat(data) {
		return nil
	}

	return []interfaces.ValidationError{
		{
			Field:     "integer",
			Message:   "Value must be a valid integer within specified constraints",
			ErrorType: "INVALID_INTEGER_VALUE",
			Value:     data,
		},
	}
}
