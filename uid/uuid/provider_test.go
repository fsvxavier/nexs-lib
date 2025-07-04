package uuid

import (
	"sync"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/uid/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUUIDProvider_New(t *testing.T) {
	p := NewProvider()
	data, err := p.New()

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, interfaces.UUIDType, data.Type)
	assert.Equal(t, LEN36, len(data.Value))
}

func TestUUIDProvider_NewWithTime(t *testing.T) {
	p := NewProvider()
	now := time.Now()
	data, err := p.NewWithTime(now)

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, interfaces.UUIDType, data.Type)
	assert.Equal(t, LEN36, len(data.Value))

	// Verifica se o timestamp foi preservado (com tolerância de milissegundos)
	timestampDiff := now.UnixMilli() - data.Timestamp.UnixMilli()
	assert.LessOrEqual(t, timestampDiff, int64(1), "O timestamp deve ser preservado com tolerância de 1ms")
}

func TestUUIDProvider_Parse(t *testing.T) {
	p := NewProvider()

	// Cria um novo UUID para testar
	data, err := p.New()
	require.NoError(t, err)

	// Testa parse do UUID gerado
	parsed, err := p.Parse(data.Value)
	assert.NoError(t, err)
	assert.NotNil(t, parsed)
	assert.Equal(t, data.Value, parsed.Value)

	// Testa parse com string inválida
	_, err = p.Parse("invalid-uuid")
	assert.Error(t, err)

	// Testa parse com string sem hífens (formato hex)
	hexID := data.Value
	hexID = hexID[0:8] + hexID[9:13] + hexID[14:18] + hexID[19:23] + hexID[24:]
	parsed, err = p.Parse(hexID)
	assert.NoError(t, err)
	assert.NotNil(t, parsed)
}

func TestUUIDProvider_ExtractTimestamp(t *testing.T) {
	p := NewProvider()

	// Cria um UUID com timestamp conhecido
	now := time.Now()
	data, err := p.NewWithTime(now)
	require.NoError(t, err)

	// Extrai timestamp do UUID
	timestamp, err := p.ExtractTimestamp(data.Value)
	assert.NoError(t, err)
	assert.Equal(t, now.UnixMilli(), timestamp.UnixMilli())

	// Teste com ID inválido
	_, err = p.ExtractTimestamp("invalid-id")
	assert.Error(t, err)
}

func TestUUIDProvider_IsValid(t *testing.T) {
	p := NewProvider()

	// Gera um UUID válido
	data, err := p.New()
	require.NoError(t, err)

	// Verifica UUID válido
	valid := p.IsValid(data.Value)
	assert.True(t, valid)

	// Verifica string inválida
	valid = p.IsValid("invalid-uuid")
	assert.False(t, valid)
}

func TestUUIDProvider_Type(t *testing.T) {
	p := NewProvider()
	assert.Equal(t, interfaces.UUIDType, p.Type())
}

// Testes de race condition
func TestUUIDProvider_RaceCondition(t *testing.T) {
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

// Benchmark para geração de UUIDs
func BenchmarkUUIDProvider_New(b *testing.B) {
	p := NewProvider()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.New()
	}
}

// Benchmark para parse de UUIDs
func BenchmarkUUIDProvider_Parse(b *testing.B) {
	p := NewProvider()
	data, _ := p.New()
	id := data.Value

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Parse(id)
	}
}
