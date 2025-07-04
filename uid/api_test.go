package uid

import (
	"sync"
	"testing"
	"time"

	"github.com/fsvxavier/nexs-lib/uid/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewID(t *testing.T) {
	// Testa com ULID
	ulid, err := NewID(interfaces.ULIDType)
	assert.NoError(t, err)
	assert.NotNil(t, ulid)
	assert.Equal(t, interfaces.ULIDType, ulid.Type)

	// Testa com UUID
	uuid, err := NewID(interfaces.UUIDType)
	assert.NoError(t, err)
	assert.NotNil(t, uuid)
	assert.Equal(t, interfaces.UUIDType, uuid.Type)

	// Testa com tipo inválido
	_, err = NewID("invalid-type")
	assert.Error(t, err)
}

func TestNewIDWithTime(t *testing.T) {
	now := time.Now()

	// Testa com ULID
	ulid, err := NewIDWithTime(interfaces.ULIDType, now)
	assert.NoError(t, err)
	assert.NotNil(t, ulid)
	assert.Equal(t, interfaces.ULIDType, ulid.Type)

	// Verifica se o timestamp foi preservado
	assert.Equal(t, now.UnixMilli(), ulid.Timestamp.UnixMilli())

	// Testa com UUID
	uuid, err := NewIDWithTime(interfaces.UUIDType, now)
	assert.NoError(t, err)
	assert.NotNil(t, uuid)
	assert.Equal(t, interfaces.UUIDType, uuid.Type)

	// Verifica se o timestamp foi preservado
	assert.Equal(t, now.UnixMilli(), uuid.Timestamp.UnixMilli())
}

func TestParse(t *testing.T) {
	// Cria IDs para teste
	ulid, err := NewULID()
	require.NoError(t, err)

	uuid, err := NewUUID()
	require.NoError(t, err)

	// Testa parse de ULID
	parsedUlid, err := Parse(interfaces.ULIDType, ulid.Value)
	assert.NoError(t, err)
	assert.NotNil(t, parsedUlid)
	assert.Equal(t, ulid.Value, parsedUlid.Value)

	// Testa parse de UUID
	parsedUuid, err := Parse(interfaces.UUIDType, uuid.Value)
	assert.NoError(t, err)
	assert.NotNil(t, parsedUuid)
	assert.Equal(t, uuid.Value, parsedUuid.Value)

	// Testa conversão cruzada (parse de UUID como ULID)
	parsedAsUlid, err := Parse(interfaces.ULIDType, uuid.Value)
	assert.NoError(t, err)
	assert.NotNil(t, parsedAsUlid)
}

func TestExtractTimestamp(t *testing.T) {
	now := time.Now()

	// Cria IDs com timestamp conhecido
	ulid, err := NewULIDWithTime(now)
	require.NoError(t, err)

	uuid, err := NewUUIDWithTime(now)
	require.NoError(t, err)

	// Extrai timestamp de ULID
	tsUlid, err := ExtractTimestamp(interfaces.ULIDType, ulid.Value)
	assert.NoError(t, err)
	assert.Equal(t, now.UnixMilli(), tsUlid.UnixMilli())

	// Extrai timestamp de UUID
	tsUuid, err := ExtractTimestamp(interfaces.UUIDType, uuid.Value)
	assert.NoError(t, err)
	assert.Equal(t, now.UnixMilli(), tsUuid.UnixMilli())
}

func TestIsValid(t *testing.T) {
	// Cria IDs válidos para teste
	ulid, err := NewULID()
	require.NoError(t, err)

	uuid, err := NewUUID()
	require.NoError(t, err)

	// Testa validação de ULID
	assert.True(t, IsValid(interfaces.ULIDType, ulid.Value))
	assert.False(t, IsValid(interfaces.ULIDType, "invalid-ulid"))

	// Testa validação de UUID
	assert.True(t, IsValid(interfaces.UUIDType, uuid.Value))
	assert.False(t, IsValid(interfaces.UUIDType, "invalid-uuid"))
}

func TestConversionFunctions(t *testing.T) {
	// Cria IDs para teste
	ulid, err := NewULID()
	require.NoError(t, err)

	// Testa ToUUID
	uuidStr, err := ToUUID(ulid.Value)
	assert.NoError(t, err)
	assert.Equal(t, 36, len(uuidStr))

	// Testa FromUUID
	ulidStr, err := FromUUID(uuidStr)
	assert.NoError(t, err)
	assert.Equal(t, 26, len(ulidStr))

	// Testa ToHex
	hexStr, err := ToHex(ulid.Value)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(hexStr))

	// Testa FromHex
	idStr, err := FromHex(hexStr)
	assert.NoError(t, err)
	assert.NotEmpty(t, idStr)

	// Testa ToBytes
	bytes, err := ToBytes(ulid.Value)
	assert.NoError(t, err)
	assert.Equal(t, 16, len(bytes))

	// Testa FromBytes
	idStr, err = FromBytes(bytes)
	assert.NoError(t, err)
	assert.NotEmpty(t, idStr)
}

func TestULIDConvenienceFunctions(t *testing.T) {
	// Testa NewULID
	ulid, err := NewULID()
	assert.NoError(t, err)
	assert.NotNil(t, ulid)
	assert.Equal(t, interfaces.ULIDType, ulid.Type)

	// Testa NewULIDWithTime
	now := time.Now()
	ulid, err = NewULIDWithTime(now)
	assert.NoError(t, err)
	assert.Equal(t, now.UnixMilli(), ulid.Timestamp.UnixMilli())

	// Testa ParseULID
	parsed, err := ParseULID(ulid.Value)
	assert.NoError(t, err)
	assert.Equal(t, ulid.Value, parsed.Value)

	// Testa IsValidULID
	assert.True(t, IsValidULID(ulid.Value))
	assert.False(t, IsValidULID("invalid-ulid"))

	// Testa ExtractTimestampFromULID
	ts, err := ExtractTimestampFromULID(ulid.Value)
	assert.NoError(t, err)
	assert.Equal(t, now.UnixMilli(), ts.UnixMilli())
}

func TestUUIDConvenienceFunctions(t *testing.T) {
	// Testa NewUUID
	uuid, err := NewUUID()
	assert.NoError(t, err)
	assert.NotNil(t, uuid)
	assert.Equal(t, interfaces.UUIDType, uuid.Type)

	// Testa NewUUIDWithTime
	now := time.Now()
	uuid, err = NewUUIDWithTime(now)
	assert.NoError(t, err)
	assert.Equal(t, now.UnixMilli(), uuid.Timestamp.UnixMilli())

	// Testa ParseUUID
	parsed, err := ParseUUID(uuid.Value)
	assert.NoError(t, err)
	assert.Equal(t, uuid.Value, parsed.Value)

	// Testa IsValidUUID
	assert.True(t, IsValidUUID(uuid.Value))
	assert.False(t, IsValidUUID("invalid-uuid"))

	// Testa ExtractTimestampFromUUID
	ts, err := ExtractTimestampFromUUID(uuid.Value)
	assert.NoError(t, err)
	assert.Equal(t, now.UnixMilli(), ts.UnixMilli())
}

// Testes de race condition
func TestAPI_RaceCondition(t *testing.T) {
	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			// Alterna entre ULIDs e UUIDs
			idType := interfaces.ULIDType
			if i%2 == 0 {
				idType = interfaces.UUIDType
			}

			id, err := NewID(idType)
			assert.NoError(t, err)
			assert.NotNil(t, id)

			_, err = Parse(idType, id.Value)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()
}

// Benchmarks
func BenchmarkNewULID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewULID()
	}
}

func BenchmarkNewUUID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewUUID()
	}
}

func BenchmarkParseULID(b *testing.B) {
	ulid, _ := NewULID()
	id := ulid.Value

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseULID(id)
	}
}

func BenchmarkParseUUID(b *testing.B) {
	uuid, _ := NewUUID()
	id := uuid.Value

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseUUID(id)
	}
}
