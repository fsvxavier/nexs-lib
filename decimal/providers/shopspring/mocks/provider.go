package mocks

import (
	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
	"github.com/stretchr/testify/mock"
)

// MockProvider is a mock implementation of DecimalProvider for shopspring
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) NewFromString(value string) (interfaces.Decimal, error) {
	args := m.Called(value)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(interfaces.Decimal), args.Error(1)
}

func (m *MockProvider) NewFromFloat(value float64) (interfaces.Decimal, error) {
	args := m.Called(value)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(interfaces.Decimal), args.Error(1)
}

func (m *MockProvider) NewFromInt(value int64) (interfaces.Decimal, error) {
	args := m.Called(value)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(interfaces.Decimal), args.Error(1)
}

func (m *MockProvider) Zero() interfaces.Decimal {
	args := m.Called()
	return args.Get(0).(interfaces.Decimal)
}

func (m *MockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) Version() string {
	args := m.Called()
	return args.String(0)
}

// MockDecimal is a mock implementation of Decimal for shopspring
type MockDecimal struct {
	mock.Mock
}

func (m *MockDecimal) String() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDecimal) Text(format byte) string {
	args := m.Called(format)
	return args.String(0)
}

func (m *MockDecimal) IsEqual(other interfaces.Decimal) bool {
	args := m.Called(other)
	return args.Bool(0)
}

func (m *MockDecimal) IsGreaterThan(other interfaces.Decimal) bool {
	args := m.Called(other)
	return args.Bool(0)
}

func (m *MockDecimal) IsLessThan(other interfaces.Decimal) bool {
	args := m.Called(other)
	return args.Bool(0)
}

func (m *MockDecimal) IsGreaterThanOrEqual(other interfaces.Decimal) bool {
	args := m.Called(other)
	return args.Bool(0)
}

func (m *MockDecimal) IsLessThanOrEqual(other interfaces.Decimal) bool {
	args := m.Called(other)
	return args.Bool(0)
}

func (m *MockDecimal) IsZero() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockDecimal) IsPositive() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockDecimal) IsNegative() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockDecimal) Add(other interfaces.Decimal) (interfaces.Decimal, error) {
	args := m.Called(other)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(interfaces.Decimal), args.Error(1)
}

func (m *MockDecimal) Sub(other interfaces.Decimal) (interfaces.Decimal, error) {
	args := m.Called(other)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(interfaces.Decimal), args.Error(1)
}

func (m *MockDecimal) Mul(other interfaces.Decimal) (interfaces.Decimal, error) {
	args := m.Called(other)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(interfaces.Decimal), args.Error(1)
}

func (m *MockDecimal) Div(other interfaces.Decimal) (interfaces.Decimal, error) {
	args := m.Called(other)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(interfaces.Decimal), args.Error(1)
}

func (m *MockDecimal) Mod(other interfaces.Decimal) (interfaces.Decimal, error) {
	args := m.Called(other)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(interfaces.Decimal), args.Error(1)
}

func (m *MockDecimal) Abs() interfaces.Decimal {
	args := m.Called()
	return args.Get(0).(interfaces.Decimal)
}

func (m *MockDecimal) Neg() interfaces.Decimal {
	args := m.Called()
	return args.Get(0).(interfaces.Decimal)
}

func (m *MockDecimal) Truncate(precision uint32, minExponent int32) interfaces.Decimal {
	args := m.Called(precision, minExponent)
	return args.Get(0).(interfaces.Decimal)
}

func (m *MockDecimal) TrimZerosRight() interfaces.Decimal {
	args := m.Called()
	return args.Get(0).(interfaces.Decimal)
}

func (m *MockDecimal) Round(places int32) interfaces.Decimal {
	args := m.Called(places)
	return args.Get(0).(interfaces.Decimal)
}

func (m *MockDecimal) Float64() (float64, error) {
	args := m.Called()
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockDecimal) Int64() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDecimal) MarshalJSON() ([]byte, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockDecimal) UnmarshalJSON(data []byte) error {
	args := m.Called(data)
	return args.Error(0)
}

func (m *MockDecimal) InternalValue() interface{} {
	args := m.Called()
	return args.Get(0)
}
