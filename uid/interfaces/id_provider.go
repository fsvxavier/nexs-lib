package interfaces

import (
	"time"
)

// IDType representa o tipo de identificador
type IDType string

const (
	// ULIDType representa um identificador ULID
	ULIDType IDType = "ulid"

	// UUIDType representa um identificador UUID
	UUIDType IDType = "uuid"
)

// IDData contém os dados do identificador
type IDData struct {
	// Timestamp é o momento de criação do ID
	Timestamp time.Time

	// Value é o valor do ID em formato string nativo
	Value string

	// HexValue é o valor do ID em hexadecimal
	HexValue string

	// UUIDString é o valor do ID em formato UUID
	UUIDString string

	// HexBytes é a representação em bytes do ID
	HexBytes []byte

	// Type é o tipo de ID
	Type IDType
}

// IDProvider define a interface para provedores de IDs
type IDProvider interface {
	// New cria um novo ID
	New() (*IDData, error)

	// NewWithTime cria um novo ID com um timestamp específico
	NewWithTime(t time.Time) (*IDData, error)

	// Parse converte uma string em um IDData
	Parse(id string) (*IDData, error)

	// ExtractTimestamp extrai o timestamp de um ID
	ExtractTimestamp(id string) (time.Time, error)

	// IsValid verifica se um ID é válido
	IsValid(id string) bool

	// Type retorna o tipo de ID que este provider gera
	Type() IDType
}

// Converter define a interface para conversão entre tipos de ID
type Converter interface {
	// ToBytes converte o ID em uma representação de bytes
	ToBytes(id string) ([]byte, error)

	// FromBytes converte bytes para um ID
	FromBytes(b []byte) (string, error)

	// ToUUID converte o ID para o formato UUID
	ToUUID(id string) (string, error)

	// FromUUID converte de UUID para o formato nativo
	FromUUID(uuid string) (string, error)

	// ToHex converte o ID para representação hexadecimal
	ToHex(id string) (string, error)

	// FromHex converte de hexadecimal para o formato nativo
	FromHex(hex string) (string, error)
}

// Factory define a interface para fábrica de provedores de ID
type Factory interface {
	// GetProvider retorna o provedor de ID do tipo especificado
	GetProvider(idType IDType) (IDProvider, error)

	// RegisterProvider registra um novo provedor de ID
	RegisterProvider(provider IDProvider) error
}
