package ulid

import (
	"sync"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/uid/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestULIDProvider_New(t *testing.T) {
	p := NewProvider()
	data, err := p.New()

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, interfaces.ULIDType, data.Type)
	assert.Equal(t, LEN26, len(data.Value))
}

func TestULIDProvider_NewWithTime(t *testing.T) {
	p := NewProvider()
	now := time.Now()
	data, err := p.NewWithTime(now)

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, interfaces.ULIDType, data.Type)
	assert.Equal(t, LEN26, len(data.Value))

	// Verifica se o timestamp foi preservado (com tolerância de milissegundos)
	timestampDiff := now.UnixMilli() - data.Timestamp.UnixMilli()
	assert.LessOrEqual(t, timestampDiff, int64(1), "O timestamp deve ser preservado com tolerância de 1ms")

	// Testa com timestamp inválido
	_, err = p.NewWithTime(time.Date(1969, 12, 31, 23, 59, 59, 0, time.UTC))
	assert.Error(t, err, "Deve rejeitar timestamp anterior a 1970")

	// Testa com timestamp muito grande
	maxTime := time.Unix(1<<48/1000, 0)
	_, err = p.NewWithTime(maxTime.Add(24 * time.Hour))
	assert.Error(t, err, "Deve rejeitar timestamp que excede o máximo suportado")
}

func TestULIDProvider_Parse(t *testing.T) {
	p := NewProvider()

	// Cria um novo ULID para testar
	data, err := p.New()
	require.NoError(t, err)

	// Testa parse do ULID gerado
	parsed, err := p.Parse(data.Value)
	assert.NoError(t, err)
	assert.NotNil(t, parsed)
	assert.Equal(t, data.Value, parsed.Value)

	// Testa parse com string inválida
	_, err = p.Parse("invalid-ulid")
	assert.Error(t, err)

	// Testa parse com string UUID
	uuidStr := data.UUIDString
	parsed, err = p.Parse(uuidStr)
	assert.NoError(t, err)
	assert.NotNil(t, parsed)
}

func TestULIDProvider_ExtractTimestamp(t *testing.T) {
	p := NewProvider()

	// Cria um ULID com timestamp conhecido
	now := time.Now()
	data, err := p.NewWithTime(now)
	require.NoError(t, err)

	// Extrai timestamp do ULID
	timestamp, err := p.ExtractTimestamp(data.Value)
	assert.NoError(t, err)
	assert.Equal(t, now.UnixMilli(), timestamp.UnixMilli())

	// Extrai timestamp do formato UUID
	timestamp, err = p.ExtractTimestamp(data.UUIDString)
	assert.NoError(t, err)
	assert.Equal(t, now.UnixMilli(), timestamp.UnixMilli())

	// Teste com ID inválido
	_, err = p.ExtractTimestamp("invalid-id")
	assert.Error(t, err)
}

func TestULIDProvider_IsValid(t *testing.T) {
	p := NewProvider()

	// Gera um ULID válido
	data, err := p.New()
	require.NoError(t, err)

	// Verifica ULID válido
	valid := p.IsValid(data.Value)
	assert.True(t, valid)

	// Verifica UUID válido
	valid = p.IsValid(data.UUIDString)
	assert.True(t, valid)

	// Verifica string inválida
	valid = p.IsValid("invalid-ulid")
	assert.False(t, valid)
}

func TestULIDProvider_Type(t *testing.T) {
	p := NewProvider()
	assert.Equal(t, interfaces.ULIDType, p.Type())
}

// Testes de race condition
func TestULIDProvider_RaceCondition(t *testing.T) {
	p := NewProvider()
	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			data, err := p.New()
			assert.NoError(t, err)
			assert.NotNil(t, data)
		}()
	}

	wg.Wait()
}

// Benchmark para geração de ULIDs
func BenchmarkULIDProvider_New(b *testing.B) {
	p := NewProvider()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.New()
	}
}

// Benchmark para parse de ULIDs
func BenchmarkULIDProvider_Parse(b *testing.B) {
	p := NewProvider()
	data, _ := p.New()
	id := data.Value

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Parse(id)
	}
}
