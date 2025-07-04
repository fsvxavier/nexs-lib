package uid

import (
	"time"

	"github.com/fsvxavier/nexs-lib/uid/interfaces"
)

var (
	// defaultFactory é uma instância global de factory
	defaultFactory interfaces.Factory

	// defaultConverter é uma instância global de converter
	defaultConverter *Converter
)

func init() {
	// Inicializa a factory e o converter padrões
	defaultFactory = SetupProviders()
	defaultConverter = NewConverter()
}

// NewID cria um novo ID com o tipo especificado
func NewID(idType interfaces.IDType) (*interfaces.IDData, error) {
	provider, err := defaultFactory.GetProvider(idType)
	if err != nil {
		return nil, err
	}

	return provider.New()
}

// NewIDWithTime cria um novo ID com o tipo e timestamp especificados
func NewIDWithTime(idType interfaces.IDType, t time.Time) (*interfaces.IDData, error) {
	provider, err := defaultFactory.GetProvider(idType)
	if err != nil {
		return nil, err
	}

	return provider.NewWithTime(t)
}

// Parse converte uma string para um IDData
func Parse(idType interfaces.IDType, id string) (*interfaces.IDData, error) {
	provider, err := defaultFactory.GetProvider(idType)
	if err != nil {
		return nil, err
	}

	return provider.Parse(id)
}

// ExtractTimestamp extrai o timestamp de um ID
func ExtractTimestamp(idType interfaces.IDType, id string) (time.Time, error) {
	provider, err := defaultFactory.GetProvider(idType)
	if err != nil {
		return time.Time{}, err
	}

	return provider.ExtractTimestamp(id)
}

// IsValid verifica se um ID é válido
func IsValid(idType interfaces.IDType, id string) bool {
	provider, err := defaultFactory.GetProvider(idType)
	if err != nil {
		return false
	}

	return provider.IsValid(id)
}

// ToUUID converte qualquer ID para o formato UUID
func ToUUID(id string) (string, error) {
	return defaultConverter.ToUUID(id)
}

// FromUUID converte de UUID para ULID
func FromUUID(uuidStr string) (string, error) {
	return defaultConverter.FromUUID(uuidStr)
}

// ToHex converte um ID para hexadecimal
func ToHex(id string) (string, error) {
	return defaultConverter.ToHex(id)
}

// FromHex converte de hexadecimal para ID
func FromHex(hexStr string) (string, error) {
	return defaultConverter.FromHex(hexStr)
}

// ToBytes converte um ID para bytes
func ToBytes(id string) ([]byte, error) {
	return defaultConverter.ToBytes(id)
}

// FromBytes converte bytes para ID
func FromBytes(b []byte) (string, error) {
	return defaultConverter.FromBytes(b)
}

// Funções de conveniência para ULID

// NewULID cria um novo ULID
func NewULID() (*interfaces.IDData, error) {
	return NewID(interfaces.ULIDType)
}

// NewULIDWithTime cria um novo ULID com timestamp específico
func NewULIDWithTime(t time.Time) (*interfaces.IDData, error) {
	return NewIDWithTime(interfaces.ULIDType, t)
}

// ParseULID tenta fazer o parse de um ULID
func ParseULID(id string) (*interfaces.IDData, error) {
	return Parse(interfaces.ULIDType, id)
}

// IsValidULID verifica se é um ULID válido
func IsValidULID(id string) bool {
	return IsValid(interfaces.ULIDType, id)
}

// ExtractTimestampFromULID extrai o timestamp de um ULID
func ExtractTimestampFromULID(id string) (time.Time, error) {
	return ExtractTimestamp(interfaces.ULIDType, id)
}

// Funções de conveniência para UUID

// NewUUID cria um novo UUID
func NewUUID() (*interfaces.IDData, error) {
	return NewID(interfaces.UUIDType)
}

// NewUUIDWithTime cria um novo UUID com timestamp específico
func NewUUIDWithTime(t time.Time) (*interfaces.IDData, error) {
	return NewIDWithTime(interfaces.UUIDType, t)
}

// ParseUUID tenta fazer o parse de um UUID
func ParseUUID(id string) (*interfaces.IDData, error) {
	return Parse(interfaces.UUIDType, id)
}

// IsValidUUID verifica se é um UUID válido
func IsValidUUID(id string) bool {
	return IsValid(interfaces.UUIDType, id)
}

// ExtractTimestampFromUUID extrai o timestamp de um UUID
func ExtractTimestampFromUUID(id string) (time.Time, error) {
	return ExtractTimestamp(interfaces.UUIDType, id)
}
