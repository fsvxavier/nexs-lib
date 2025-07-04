package decimal

import (
	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
)

// NewDecimal é um helper para criar um decimal usando o provider padrão (ShopSpring)
func NewDecimal(value string) (interfaces.Decimal, error) {
	provider := NewProvider(ShopSpring)
	return provider.NewFromString(value)
}

// NewDecimalWithProvider é um helper para criar um decimal usando o provider especificado
func NewDecimalWithProvider(value string, providerType ProviderType) (interfaces.Decimal, error) {
	provider := NewProvider(providerType)
	return provider.NewFromString(value)
}

// ShopSpringDecimal é um helper para criar um decimal usando o provider ShopSpring
func ShopSpringDecimal(value string) (interfaces.Decimal, error) {
	provider := NewProvider(ShopSpring)
	return provider.NewFromString(value)
}

// APDDecimal é um helper para criar um decimal usando o provider APD
func APDDecimal(value string) (interfaces.Decimal, error) {
	provider := NewProvider(APD)
	return provider.NewFromString(value)
}
