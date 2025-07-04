package decimal

import (
	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
	"github.com/fsvxavier/nexs-lib/decimal/providers/apd"
	"github.com/fsvxavier/nexs-lib/decimal/providers/shopspring"
)

// ProviderType representa o tipo de provider
type ProviderType string

const (
	// APD representa o provider baseado em github.com/cockroachdb/apd
	APD ProviderType = "apd"

	// ShopSpring representa o provider baseado em github.com/shopspring/decimal
	ShopSpring ProviderType = "shopspring"
)

// NewProvider cria um novo provider decimal com base no tipo especificado
func NewProvider(providerType ProviderType) interfaces.Provider {
	switch providerType {
	case ShopSpring:
		return shopspring.NewProvider()
	case APD:
		return apd.NewProvider()
	default:
		// Retorna o provider padrão (ShopSpring)
		return shopspring.NewProvider()
	}
}
