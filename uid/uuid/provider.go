package uuid

import (
	"encoding/hex"
	"errors"
	"time"

	"github.com/fsvxavier/nexs-lib/uid/interfaces"
	"github.com/fsvxavier/nexs-lib/uid/model"
	"github.com/google/uuid"
)

const (
	// LEN36 é o comprimento padrão da string UUID
	LEN36 = 36

	// LEN32 é o comprimento da string UUID sem hífens
	LEN32 = 32

	// LEN16 é o comprimento dos bytes de um UUID
	LEN16 = 16
)

// Provider implementa interfaces.IDProvider para UUID
type Provider struct{}

// NewProvider cria uma nova instância do provider UUID
func NewProvider() *Provider {
	return &Provider{}
}

// New gera um novo UUID
func (p *Provider) New() (*interfaces.IDData, error) {
	id := uuid.New()
	return p.dataFromUUID(id), nil
}

// NewWithTime gera um novo UUID v1 com timestamp
// Nota: UUIDs v1 incorporam o timestamp de forma diferente dos ULIDs
// Este método simula o comportamento dos ULIDs o máximo possível
func (p *Provider) NewWithTime(t time.Time) (*interfaces.IDData, error) {
	// UUID v1 usa timestamp, mas a implementação da google não tem suporte para
	// definir o timestamp, então simulamos colocando o timestamp nos bytes iniciais

	// Cria um UUID v4 (aleatório)
	id := uuid.New()

	// Converte o timestamp para bytes
	timestampMs := uint64(t.UnixNano() / int64(time.Millisecond))

	// Coloca o timestamp nos primeiros 6 bytes do UUID
	// Similar ao formato do ULID
	bytes := id[:]
	bytes[0] = byte(timestampMs >> 40)
	bytes[1] = byte(timestampMs >> 32)
	bytes[2] = byte(timestampMs >> 24)
	bytes[3] = byte(timestampMs >> 16)
	bytes[4] = byte(timestampMs >> 8)
	bytes[5] = byte(timestampMs)

	// Cria novo UUID com os bytes modificados
	newID, err := uuid.FromBytes(bytes)
	if err != nil {
		return nil, err
	}

	return p.dataFromUUID(newID), nil
}

// Parse analisa uma string UUID e converte para IDData
func (p *Provider) Parse(id string) (*interfaces.IDData, error) {
	// Tenta fazer o parse diretamente
	uid, err := uuid.Parse(id)
	if err != nil {
		// Se falhar, tenta como hex string sem hífens
		if len(id) == LEN32 {
			hexBytes, err := hex.DecodeString(id)
			if err != nil {
				return nil, err
			}

			if len(hexBytes) != LEN16 {
				return nil, errors.New("tamanho inválido para UUID")
			}

			uid, err = uuid.FromBytes(hexBytes)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return p.dataFromUUID(uid), nil
}

// ExtractTimestamp extrai o timestamp de um UUID
// Assumindo que os primeiros 6 bytes contêm o timestamp
func (p *Provider) ExtractTimestamp(id string) (time.Time, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return time.Time{}, err
	}

	// Extrai timestamp dos primeiros 6 bytes (similar ao formato ULID)
	timestamp := uint64(uid[0])<<40 |
		uint64(uid[1])<<32 |
		uint64(uid[2])<<24 |
		uint64(uid[3])<<16 |
		uint64(uid[4])<<8 |
		uint64(uid[5])

	return time.UnixMilli(int64(timestamp)), nil
}

// IsValid verifica se um ID é um UUID válido
func (p *Provider) IsValid(value string) bool {
	err := uuid.Validate(value)
	return err == nil
}

// Type retorna o tipo de ID que este provider gera
func (p *Provider) Type() interfaces.IDType {
	return interfaces.UUIDType
}

// dataFromUUID converte um UUID em IDData
func (p *Provider) dataFromUUID(id uuid.UUID) *interfaces.IDData {
	// Extrai o timestamp simulado a partir dos primeiros 6 bytes
	timestamp := uint64(id[0])<<40 |
		uint64(id[1])<<32 |
		uint64(id[2])<<24 |
		uint64(id[3])<<16 |
		uint64(id[4])<<8 |
		uint64(id[5])

	data := &model.IDData{
		Timestamp:  time.UnixMilli(int64(timestamp)),
		Value:      id.String(),
		UUIDString: id.String(),
		HexValue:   hex.EncodeToString(id[:]),
		HexBytes:   id[:],
		Type:       interfaces.UUIDType,
	}

	return data.ToInterface()
}
