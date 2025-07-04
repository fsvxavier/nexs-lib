package uid

import (
	"encoding/hex"
	"errors"

	"github.com/fsvxavier/nexs-lib/uid/factory"
	"github.com/fsvxavier/nexs-lib/uid/interfaces"
	"github.com/fsvxavier/nexs-lib/uid/ulid"
	uuidprovider "github.com/fsvxavier/nexs-lib/uid/uuid"
	"github.com/google/uuid"
	okulidlib "github.com/oklog/ulid/v2"
)

// Converter é uma implementação de interfaces.Converter
type Converter struct{}

// NewConverter cria uma nova instância de Converter
func NewConverter() *Converter {
	return &Converter{}
}

// ToBytes converte um ID (ULID ou UUID) em bytes
func (c *Converter) ToBytes(id string) ([]byte, error) {
	// Tenta ULID primeiro
	if len(id) == 26 {
		ulid, err := okulidlib.Parse(id)
		if err == nil {
			return ulid.Bytes(), nil
		}
	}

	// Tenta UUID
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return uid[:], nil
}

// FromBytes converte bytes em um ID (preferência para formato ULID)
func (c *Converter) FromBytes(b []byte) (string, error) {
	if len(b) != 16 {
		return "", errors.New("bytes inválidos: devem ter 16 bytes de comprimento")
	}

	var ulid okulidlib.ULID
	copy(ulid[:], b)
	return ulid.String(), nil
}

// ToUUID converte um ID para o formato UUID
func (c *Converter) ToUUID(id string) (string, error) {
	// Se já for UUID
	if len(id) == 36 {
		_, err := uuid.Parse(id)
		if err == nil {
			return id, nil
		}
	}

	// Se for ULID
	if len(id) == 26 {
		ulid, err := okulidlib.Parse(id)
		if err != nil {
			return "", err
		}

		uid, err := uuid.FromBytes(ulid.Bytes())
		if err != nil {
			return "", err
		}

		return uid.String(), nil
	}

	return "", errors.New("formato de ID inválido")
}

// FromUUID converte de UUID para ULID
func (c *Converter) FromUUID(uuidStr string) (string, error) {
	uid, err := uuid.Parse(uuidStr)
	if err != nil {
		return "", err
	}

	var ulid okulidlib.ULID
	copy(ulid[:], uid[:])

	return ulid.String(), nil
}

// ToHex converte um ID para representação hexadecimal
func (c *Converter) ToHex(id string) (string, error) {
	bytes, err := c.ToBytes(id)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// FromHex converte de hexadecimal para o formato ULID
func (c *Converter) FromHex(hexStr string) (string, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}

	return c.FromBytes(bytes)
}

// SetupProviders configura todos os provedores no factory padrão
func SetupProviders() interfaces.Factory {
	f := factory.NewFactory()

	// Registra o provedor ULID
	f.RegisterProvider(ulid.NewProvider())

	// Registra o provedor UUID
	f.RegisterProvider(uuidprovider.NewProvider())

	return f
}
