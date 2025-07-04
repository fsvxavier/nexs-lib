package ulid

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/fsvxavier/nexs-lib/uid/interfaces"
	"github.com/fsvxavier/nexs-lib/uid/model"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

const (
	// LEN26 é o comprimento padrão da string ULID
	LEN26 = 26

	// LEN16 é o comprimento dos bytes de um UUID/ULID
	LEN16 = 16

	// INVALID_UNIX_YEAR é usado para validação de timestamp
	INVALID_UNIX_YEAR = 1969
)

// Provider implementa interfaces.IDProvider para ULID
type Provider struct {
	entropy *ulid.MonotonicEntropy
	mutex   sync.Mutex
}

// NewProvider cria uma nova instância do provider ULID
func NewProvider() *Provider {
	return &Provider{
		entropy: ulid.Monotonic(rand.Reader, 0),
	}
}

// New gera um novo ULID
func (p *Provider) New() (*interfaces.IDData, error) {
	p.mutex.Lock()
	id := ulid.MustNew(ulid.Timestamp(time.Now()), p.entropy)
	p.mutex.Unlock()

	return p.dataFromULID(id), nil
}

// NewWithTime gera um novo ULID com um timestamp específico
func (p *Provider) NewWithTime(t time.Time) (*interfaces.IDData, error) {
	p.mutex.Lock()
	// Verifica se o timestamp é válido para ULID (não pode ser anterior a 1970)
	if t.Year() < 1970 {
		return nil, errors.New("timestamp deve ser após 1970")
	}

	// Verifica se o timestamp não excede o máximo suportado pelo ULID (48 bits = 281474976710655)
	ms := t.UnixMilli()
	if ms > 281474976710655 {
		return nil, errors.New("timestamp excede o máximo suportado pelo ULID")
	}

	id := ulid.MustNew(uint64(ms), p.entropy)
	p.mutex.Unlock()

	return p.dataFromULID(id), nil
}

// Parse tenta analisar uma string ULID ou UUID e converter para IDData
func (p *Provider) Parse(str string) (*interfaces.IDData, error) {
	// Verifica se é um ULID (26 caracteres)
	if len(str) == LEN26 {
		id, err := ulid.Parse(str)
		if err != nil {
			return nil, err
		}
		return p.dataFromULID(id), nil
	}

	// Tenta converter de UUID para ULID
	str = trimUUIDHyphens(str)
	hexBytes, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	if len(hexBytes) != LEN16 {
		return nil, errors.New("tamanho inválido para UUID/ULID")
	}

	var id ulid.ULID
	copy(id[:], hexBytes)

	return p.dataFromULID(id), nil
}

// ExtractTimestamp extrai o timestamp de um ID ULID ou UUID
func (p *Provider) ExtractTimestamp(id string) (time.Time, error) {
	// Se for um ULID direto
	if len(id) == LEN26 {
		parsed, err := ulid.Parse(id)
		if err != nil {
			return time.Time{}, err
		}
		return time.UnixMilli(int64(parsed.Time())), nil
	}

	// Se for um UUID, tenta extrair o timestamp dos primeiros 6 bytes
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return time.Time{}, err
	}

	timestamp := uint64(uuidID[0])<<40 |
		uint64(uuidID[1])<<32 |
		uint64(uuidID[2])<<24 |
		uint64(uuidID[3])<<16 |
		uint64(uuidID[4])<<8 |
		uint64(uuidID[5])

	return time.UnixMilli(int64(timestamp)), nil
}

// IsValid verifica se um ID é um ULID válido
func (p *Provider) IsValid(value string) bool {
	if len(value) == LEN26 {
		_, err := ulid.Parse(value)
		return err == nil
	}

	// Se for um UUID, valida e verifica o timestamp
	err := uuid.Validate(value)
	if err != nil {
		return false
	}

	// Tenta extrair o timestamp e verifica se é válido
	t, err := p.ExtractTimestamp(value)
	if err != nil {
		return false
	}

	return t.Year() > INVALID_UNIX_YEAR
}

// Type retorna o tipo de ID que este provider gera
func (p *Provider) Type() interfaces.IDType {
	return interfaces.ULIDType
}

// dataFromULID converte um ULID em IDData
func (p *Provider) dataFromULID(id ulid.ULID) *interfaces.IDData {
	data := &model.IDData{
		Timestamp: time.UnixMilli(int64(id.Time())),
		Value:     id.String(),
		HexValue:  hex.EncodeToString(id.Bytes()),
		Type:      interfaces.ULIDType,
	}

	data.HexBytes = id.Bytes()

	// Converte para UUID
	uid, err := uuid.FromBytes(data.HexBytes)
	if err == nil {
		data.UUIDString = uid.String()
	}

	return data.ToInterface()
}

// Função auxiliar para remover hífens de UUIDs
func trimUUIDHyphens(str string) string {
	if len(str) == 36 && str[8] == '-' && str[13] == '-' && str[18] == '-' && str[23] == '-' {
		return str[0:8] + str[9:13] + str[14:18] + str[19:23] + str[24:]
	}
	return str
}
