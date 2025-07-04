package uid

import (
	"encoding/hex"
	"sync"
	"testing"

	"github.com/fsvxavier/nexs-lib/uid/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_ToBytes(t *testing.T) {
	c := NewConverter()

	// Teste com ULID
	ulid, err := NewULID()
	require.NoError(t, err)

	bytes, err := c.ToBytes(ulid.Value)
	assert.NoError(t, err)
	assert.Equal(t, 16, len(bytes))

	// Teste com UUID
	uuid, err := NewUUID()
	require.NoError(t, err)

	bytes, err = c.ToBytes(uuid.Value)
	assert.NoError(t, err)
	assert.Equal(t, 16, len(bytes))

	// Teste com valor inválido
	_, err = c.ToBytes("invalid-id")
	assert.Error(t, err)
}

func TestConverter_FromBytes(t *testing.T) {
	c := NewConverter()

	// Gera ID e converte para bytes
	ulid, err := NewULID()
	require.NoError(t, err)

	bytes, err := c.ToBytes(ulid.Value)
	require.NoError(t, err)

	// Converte de volta para string
	id, err := c.FromBytes(bytes)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	// Teste com bytes inválidos
	_, err = c.FromBytes([]byte{1, 2, 3})
	assert.Error(t, err)
}

func TestConverter_ToUUID(t *testing.T) {
	c := NewConverter()

	// Teste com ULID
	ulid, err := NewULID()
	require.NoError(t, err)

	uuidStr, err := c.ToUUID(ulid.Value)
	assert.NoError(t, err)
	assert.Equal(t, 36, len(uuidStr))

	// Teste com UUID
	uuid, err := NewUUID()
	require.NoError(t, err)

	uuidStr2, err := c.ToUUID(uuid.Value)
	assert.NoError(t, err)
	assert.Equal(t, uuid.Value, uuidStr2)

	// Teste com valor inválido
	_, err = c.ToUUID("invalid-id")
	assert.Error(t, err)
}

func TestConverter_FromUUID(t *testing.T) {
	c := NewConverter()

	// Teste com UUID válido
	uuid, err := NewUUID()
	require.NoError(t, err)

	ulid, err := c.FromUUID(uuid.Value)
	assert.NoError(t, err)
	assert.Equal(t, 26, len(ulid))

	// Teste com valor inválido
	_, err = c.FromUUID("invalid-uuid")
	assert.Error(t, err)
}

func TestConverter_ToHex(t *testing.T) {
	c := NewConverter()

	// Teste com ULID
	ulid, err := NewULID()
	require.NoError(t, err)

	hexStr, err := c.ToHex(ulid.Value)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(hexStr))

	// Teste com UUID
	uuid, err := NewUUID()
	require.NoError(t, err)

	hexStr, err = c.ToHex(uuid.Value)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(hexStr))

	// Verifica se é hexadecimal válido
	_, err = hex.DecodeString(hexStr)
	assert.NoError(t, err)

	// Teste com valor inválido
	_, err = c.ToHex("invalid-id")
	assert.Error(t, err)
}

func TestConverter_FromHex(t *testing.T) {
	c := NewConverter()

	// Gera ID e converte para hex
	ulid, err := NewULID()
	require.NoError(t, err)

	hexStr, err := c.ToHex(ulid.Value)
	require.NoError(t, err)

	// Converte de volta para string
	id, err := c.FromHex(hexStr)
	assert.NoError(t, err)
	assert.Equal(t, 26, len(id))

	// Teste com hex inválido
	_, err = c.FromHex("zzzz")
	assert.Error(t, err)
}

func TestSetupProviders(t *testing.T) {
	factory := SetupProviders()

	// Verifica se o provedor ULID foi registrado
	provider, err := factory.GetProvider(interfaces.ULIDType)
	assert.NoError(t, err)
	assert.Equal(t, interfaces.ULIDType, provider.Type())

	// Verifica se o provedor UUID foi registrado
	provider, err = factory.GetProvider(interfaces.UUIDType)
	assert.NoError(t, err)
	assert.Equal(t, interfaces.UUIDType, provider.Type())

	// Verifica erro para tipo desconhecido
	_, err = factory.GetProvider("unknown")
	assert.Error(t, err)
}

// Testes de race condition
func TestConverter_RaceCondition(t *testing.T) {
	c := NewConverter()
	ulid, err := NewULID()
	require.NoError(t, err)

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			_, err := c.ToUUID(ulid.Value)
			assert.NoError(t, err)

			_, err = c.ToHex(ulid.Value)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
}

// Benchmarks
func BenchmarkConverter_ToUUID(b *testing.B) {
	c := NewConverter()
	ulid, _ := NewULID()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.ToUUID(ulid.Value)
	}
}

func BenchmarkConverter_ToHex(b *testing.B) {
	c := NewConverter()
	ulid, _ := NewULID()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.ToHex(ulid.Value)
	}
}
