package interfaces

// Provider é uma interface que define as operações básicas de um provedor decimal
type Provider interface {
	// NewFromString cria um novo decimal a partir de uma string
	NewFromString(numString string) (Decimal, error)

	// NewFromFloat cria um novo decimal a partir de um float64
	NewFromFloat(numFloat float64) (Decimal, error)

	// NewFromInt cria um novo decimal a partir de um int64
	NewFromInt(numInt int64) (Decimal, error)
}

// Decimal é uma interface que define as operações básicas de um decimal
type Decimal interface {
	// String retorna a representação em string do decimal
	String() string

	// Float64 retorna a representação em float64 do decimal
	Float64() (float64, error)

	// Int64 retorna a representação em int64 do decimal
	Int64() (int64, error)

	// IsZero retorna true se o decimal é zero
	IsZero() bool

	// IsNegative retorna true se o decimal é negativo
	IsNegative() bool

	// IsPositive retorna true se o decimal é positivo
	IsPositive() bool

	// Equals retorna true se o decimal é igual a outro
	Equals(d Decimal) bool

	// GreaterThan retorna true se o decimal é maior que outro
	GreaterThan(d Decimal) bool

	// LessThan retorna true se o decimal é menor que outro
	LessThan(d Decimal) bool

	// GreaterThanOrEqual retorna true se o decimal é maior ou igual a outro
	GreaterThanOrEqual(d Decimal) bool

	// LessThanOrEqual retorna true se o decimal é menor ou igual a outro
	LessThanOrEqual(d Decimal) bool

	// Add adiciona outro decimal ao decimal atual e retorna o resultado
	Add(d Decimal) Decimal

	// Sub subtrai outro decimal do decimal atual e retorna o resultado
	Sub(d Decimal) Decimal

	// Mul multiplica outro decimal pelo decimal atual e retorna o resultado
	Mul(d Decimal) Decimal

	// Div divide o decimal atual por outro decimal e retorna o resultado
	Div(d Decimal) (Decimal, error)

	// Abs retorna o valor absoluto do decimal
	Abs() Decimal

	// Round arredonda o decimal para o número especificado de casas decimais
	Round(places int32) Decimal

	// Truncate trunca o decimal para o número especificado de casas decimais
	Truncate(places int32) Decimal

	// MarshalJSON implementa a interface json.Marshaler
	MarshalJSON() ([]byte, error)

	// UnmarshalJSON implementa a interface json.Unmarshaler
	UnmarshalJSON(data []byte) error
}
