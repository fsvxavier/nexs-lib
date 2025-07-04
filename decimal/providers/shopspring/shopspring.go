package shopspring

import (
	"encoding/json"

	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
	"github.com/shopspring/decimal"
)

// Provider é uma implementação de interfaces.Provider usando a biblioteca github.com/shopspring/decimal
type Provider struct{}

// NewProvider cria um novo provider
func NewProvider() *Provider {
	return &Provider{}
}

// NewFromString cria um novo decimal a partir de uma string
func (p *Provider) NewFromString(numString string) (interfaces.Decimal, error) {
	d, err := decimal.NewFromString(numString)
	if err != nil {
		return nil, err
	}
	return &Decimal{d}, nil
}

// NewFromFloat cria um novo decimal a partir de um float64
func (p *Provider) NewFromFloat(numFloat float64) (interfaces.Decimal, error) {
	d := decimal.NewFromFloat(numFloat)
	return &Decimal{d}, nil
}

// NewFromInt cria um novo decimal a partir de um int64
func (p *Provider) NewFromInt(numInt int64) (interfaces.Decimal, error) {
	d := decimal.NewFromInt(numInt)
	return &Decimal{d}, nil
}

// Decimal é um wrapper para decimal.Decimal que implementa a interface interfaces.Decimal
type Decimal struct {
	decimal decimal.Decimal
}

// String retorna a representação em string do decimal
func (d *Decimal) String() string {
	return d.decimal.String()
}

// Float64 retorna a representação em float64 do decimal
func (d *Decimal) Float64() (float64, error) {
	val, _ := d.decimal.Float64()
	return val, nil
}

// Int64 retorna a representação em int64 do decimal
func (d *Decimal) Int64() (int64, error) {
	return d.decimal.IntPart(), nil
}

// IsZero retorna true se o decimal é zero
func (d *Decimal) IsZero() bool {
	return d.decimal.IsZero()
}

// IsNegative retorna true se o decimal é negativo
func (d *Decimal) IsNegative() bool {
	return d.decimal.IsNegative()
}

// IsPositive retorna true se o decimal é positivo
func (d *Decimal) IsPositive() bool {
	return d.decimal.IsPositive()
}

// Equals retorna true se o decimal é igual a outro
func (d *Decimal) Equals(other interfaces.Decimal) bool {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return false
	}
	return d.decimal.Equal(otherDecimal.decimal)
}

// GreaterThan retorna true se o decimal é maior que outro
func (d *Decimal) GreaterThan(other interfaces.Decimal) bool {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return false
	}
	return d.decimal.GreaterThan(otherDecimal.decimal)
}

// LessThan retorna true se o decimal é menor que outro
func (d *Decimal) LessThan(other interfaces.Decimal) bool {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return false
	}
	return d.decimal.LessThan(otherDecimal.decimal)
}

// GreaterThanOrEqual retorna true se o decimal é maior ou igual a outro
func (d *Decimal) GreaterThanOrEqual(other interfaces.Decimal) bool {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return false
	}
	return d.decimal.GreaterThanOrEqual(otherDecimal.decimal)
}

// LessThanOrEqual retorna true se o decimal é menor ou igual a outro
func (d *Decimal) LessThanOrEqual(other interfaces.Decimal) bool {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return false
	}
	return d.decimal.LessThanOrEqual(otherDecimal.decimal)
}

// Add adiciona outro decimal ao decimal atual e retorna o resultado
func (d *Decimal) Add(other interfaces.Decimal) interfaces.Decimal {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return d // ou retornar erro?
	}
	result := d.decimal.Add(otherDecimal.decimal)
	return &Decimal{result}
}

// Sub subtrai outro decimal do decimal atual e retorna o resultado
func (d *Decimal) Sub(other interfaces.Decimal) interfaces.Decimal {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return d // ou retornar erro?
	}
	result := d.decimal.Sub(otherDecimal.decimal)
	return &Decimal{result}
}

// Mul multiplica outro decimal pelo decimal atual e retorna o resultado
func (d *Decimal) Mul(other interfaces.Decimal) interfaces.Decimal {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return d // ou retornar erro?
	}
	result := d.decimal.Mul(otherDecimal.decimal)
	return &Decimal{result}
}

// Div divide o decimal atual por outro decimal e retorna o resultado
func (d *Decimal) Div(other interfaces.Decimal) (interfaces.Decimal, error) {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return d, nil // ou retornar erro?
	}
	result := d.decimal.Div(otherDecimal.decimal)
	return &Decimal{result}, nil
}

// Abs retorna o valor absoluto do decimal
func (d *Decimal) Abs() interfaces.Decimal {
	result := d.decimal.Abs()
	return &Decimal{result}
}

// Round arredonda o decimal para o número especificado de casas decimais
func (d *Decimal) Round(places int32) interfaces.Decimal {
	result := d.decimal.Round(places)
	return &Decimal{result}
}

// Truncate trunca o decimal para o número especificado de casas decimais
func (d *Decimal) Truncate(places int32) interfaces.Decimal {
	result := d.decimal.Truncate(places)
	return &Decimal{result}
}

// MarshalJSON implementa a interface json.Marshaler
func (d *Decimal) MarshalJSON() ([]byte, error) {
	return d.decimal.MarshalJSON()
}

// UnmarshalJSON implementa a interface json.Unmarshaler
func (d *Decimal) UnmarshalJSON(data []byte) error {
	var dec decimal.Decimal
	err := json.Unmarshal(data, &dec)
	if err != nil {
		return err
	}
	d.decimal = dec
	return nil
}
