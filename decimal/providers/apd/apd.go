package apd

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/cockroachdb/apd/v3"
	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
)

const (
	// Constantes para configuração do context APD
	MaxPrecision = 21 // total de dígitos, antes e depois do ponto decimal
	MaxExponent  = 13 // total de dígitos, após o ponto decimal
	MinExponent  = -8 // total de dígitos, antes do ponto decimal
)

// Provider é uma implementação de dec.Provider usando a biblioteca github.com/cockroachdb/apd/v3
type Provider struct {
	ctx *apd.Context
}

// NewProvider cria um novo provider
func NewProvider() *Provider {
	ctx := apd.BaseContext.WithPrecision(MaxPrecision)
	ctx.MaxExponent = MaxExponent
	ctx.MinExponent = MinExponent
	ctx.Rounding = apd.RoundDown

	return &Provider{
		ctx: ctx,
	}
}

// NewProviderWithContext cria um novo provider com contexto personalizado
func NewProviderWithContext(precision uint32, maxExp, minExp int32, rounding apd.Rounder) *Provider {
	ctx := apd.BaseContext.WithPrecision(precision)
	ctx.MaxExponent = maxExp
	ctx.MinExponent = minExp
	ctx.Rounding = rounding

	return &Provider{
		ctx: ctx,
	}
}

// NewFromString cria um novo decimal a partir de uma string
func (p *Provider) NewFromString(numString string) (interfaces.Decimal, error) {
	d, _, err := apd.NewFromString(numString)
	if err != nil {
		return nil, err
	}
	return &Decimal{decimal: d, ctx: p.ctx}, nil
}

// NewFromFloat cria um novo decimal a partir de um float64
func (p *Provider) NewFromFloat(numFloat float64) (interfaces.Decimal, error) {
	// Converte float64 para string e depois para decimal para evitar problemas de precisão
	str := strconv.FormatFloat(numFloat, 'f', -1, 64)
	d, _, err := apd.NewFromString(str)
	if err != nil {
		return nil, err
	}
	return &Decimal{decimal: d, ctx: p.ctx}, nil
}

// NewFromInt cria um novo decimal a partir de um int64
func (p *Provider) NewFromInt(numInt int64) (interfaces.Decimal, error) {
	d := apd.New(numInt, 0)
	return &Decimal{decimal: d, ctx: p.ctx}, nil
}

// Decimal é um wrapper para apd.Decimal que implementa a interface interfaces.Decimal
type Decimal struct {
	decimal *apd.Decimal
	ctx     *apd.Context
}

// String retorna a representação em string do decimal
func (d *Decimal) String() string {
	return d.decimal.Text('f')
}

// Float64 retorna a representação em float64 do decimal
func (d *Decimal) Float64() (float64, error) {
	f, err := d.decimal.Float64()
	if err != nil {
		return 0, err
	}
	return f, nil
}

// Int64 retorna a representação em int64 do decimal
func (d *Decimal) Int64() (int64, error) {
	i, err := strconv.ParseInt(d.decimal.String(), 10, 64)
	if err != nil {
		// Para decimais que não podem ser representados como int64,
		// fazemos truncamento para obter a parte inteira
		truncated := &apd.Decimal{}
		_, err := d.ctx.Quantize(truncated, d.decimal, 0)
		if err != nil {
			return 0, err
		}

		i, err := truncated.Int64()
		if err != nil {
			return 0, err
		}

		return i, nil
	}
	return i, nil
}

// IsZero retorna true se o decimal é zero
func (d *Decimal) IsZero() bool {
	return d.decimal.IsZero()
}

// IsNegative retorna true se o decimal é negativo
func (d *Decimal) IsNegative() bool {
	return d.decimal.Negative
}

// IsPositive retorna true se o decimal é positivo
func (d *Decimal) IsPositive() bool {
	return !d.decimal.IsZero() && !d.decimal.Negative
}

// Equals retorna true se o decimal é igual a outro
func (d *Decimal) Equals(other interfaces.Decimal) bool {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return false
	}
	return d.decimal.Cmp(otherDecimal.decimal) == 0
}

// GreaterThan retorna true se o decimal é maior que outro
func (d *Decimal) GreaterThan(other interfaces.Decimal) bool {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return false
	}
	return d.decimal.Cmp(otherDecimal.decimal) > 0
}

// LessThan retorna true se o decimal é menor que outro
func (d *Decimal) LessThan(other interfaces.Decimal) bool {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return false
	}
	return d.decimal.Cmp(otherDecimal.decimal) < 0
}

// GreaterThanOrEqual retorna true se o decimal é maior ou igual a outro
func (d *Decimal) GreaterThanOrEqual(other interfaces.Decimal) bool {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return false
	}
	return d.decimal.Cmp(otherDecimal.decimal) >= 0
}

// LessThanOrEqual retorna true se o decimal é menor ou igual a outro
func (d *Decimal) LessThanOrEqual(other interfaces.Decimal) bool {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return false
	}
	return d.decimal.Cmp(otherDecimal.decimal) <= 0
}

// Add adiciona outro decimal ao decimal atual e retorna o resultado
func (d *Decimal) Add(other interfaces.Decimal) interfaces.Decimal {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return d
	}

	result := &apd.Decimal{}
	_, err := d.ctx.Add(result, d.decimal, otherDecimal.decimal)
	if err != nil {
		// Em caso de erro, retorna o valor original
		return d
	}

	return &Decimal{decimal: result, ctx: d.ctx}
}

// Sub subtrai outro decimal do decimal atual e retorna o resultado
func (d *Decimal) Sub(other interfaces.Decimal) interfaces.Decimal {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return d
	}

	result := &apd.Decimal{}
	_, err := d.ctx.Sub(result, d.decimal, otherDecimal.decimal)
	if err != nil {
		// Em caso de erro, retorna o valor original
		return d
	}

	return &Decimal{decimal: result, ctx: d.ctx}
}

// Mul multiplica outro decimal pelo decimal atual e retorna o resultado
func (d *Decimal) Mul(other interfaces.Decimal) interfaces.Decimal {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return d
	}

	result := &apd.Decimal{}
	_, err := d.ctx.Mul(result, d.decimal, otherDecimal.decimal)
	if err != nil {
		// Em caso de erro, retorna o valor original
		return d
	}

	return &Decimal{decimal: result, ctx: d.ctx}
}

// Div divide o decimal atual por outro decimal e retorna o resultado
func (d *Decimal) Div(other interfaces.Decimal) (interfaces.Decimal, error) {
	otherDecimal, ok := other.(*Decimal)
	if !ok {
		return d, nil
	}

	if otherDecimal.IsZero() {
		return nil, errors.New("division by zero")
	}

	result := &apd.Decimal{}
	_, err := d.ctx.Quo(result, d.decimal, otherDecimal.decimal)
	if err != nil {
		return nil, err
	}

	return &Decimal{decimal: result, ctx: d.ctx}, nil
}

// Abs retorna o valor absoluto do decimal
func (d *Decimal) Abs() interfaces.Decimal {
	result := &apd.Decimal{}
	_, err := d.ctx.Abs(result, d.decimal)
	if err != nil {
		// Em caso de erro, retorna o valor original
		return d
	}

	return &Decimal{decimal: result, ctx: d.ctx}
}

// Round arredonda o decimal para o número especificado de casas decimais
func (d *Decimal) Round(places int32) interfaces.Decimal {
	result := &apd.Decimal{}

	// Copia o decimal atual
	result.Set(d.decimal)

	// Quantize o resultado para o número especificado de casas decimais
	_, err := d.ctx.Quantize(result, result, -places)
	if err != nil {
		// Em caso de erro, retorna o valor original
		return d
	}

	return &Decimal{decimal: result, ctx: d.ctx}
}

// Truncate trunca o decimal para o número especificado de casas decimais
func (d *Decimal) Truncate(places int32) interfaces.Decimal {
	// Salva a configuração de arredondamento atual
	originalRounding := d.ctx.Rounding

	// Define o arredondamento para truncamento
	d.ctx.Rounding = apd.RoundDown

	result := &apd.Decimal{}

	// Copia o decimal atual
	result.Set(d.decimal)

	// Quantize o resultado para o número especificado de casas decimais
	_, err := d.ctx.Quantize(result, result, -places)

	// Restaura a configuração de arredondamento original
	d.ctx.Rounding = originalRounding

	if err != nil {
		// Em caso de erro, retorna o valor original
		return d
	}

	return &Decimal{decimal: result, ctx: d.ctx}
}

// MarshalJSON implementa a interface json.Marshaler
func (d *Decimal) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON implementa a interface json.Unmarshaler
func (d *Decimal) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		// Tenta converter diretamente de número
		var f float64
		err := json.Unmarshal(data, &f)
		if err != nil {
			return err
		}

		dec, _, err := apd.NewFromString(strconv.FormatFloat(f, 'f', -1, 64))
		if err != nil {
			return err
		}

		d.decimal = dec
		return nil
	}

	dec, _, err := apd.NewFromString(str)
	if err != nil {
		return err
	}

	d.decimal = dec
	return nil
}
