package ulid_test

import (
	"testing"

	"github.com/dock-tech/isis-golang-lib/ulid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewUlid(t *testing.T) {
	data := ulid.NewUlid()
	assert.NotNil(t, data)
	assert.Equal(t, len(data.Value), ulid.LEN26)
}

func TestParse(t *testing.T) {
	id := ulid.NewUlid()
	parsed, err := ulid.Parse(id.Value)
	assert.NoError(t, err)
	assert.NotNil(t, parsed)
	assert.Equal(t, parsed.Value, id.Value)

	// Test with UUID string
	uuidStr := uuid.New().String()
	parsed, err = ulid.Parse(uuidStr)
	assert.NoError(t, err)
	assert.NotNil(t, parsed)
}

func TestExtractTimestampFromUlid(t *testing.T) {
	id := ulid.NewUlid()
	timestamp, err := ulid.ExtractTimestampFromUlid(id.UUIDString)
	assert.NoError(t, err)
	assert.Equal(t, id.Timestamp.UnixMilli(), timestamp.UnixMilli())
}

func TestIsValidUlid(t *testing.T) {
	id := ulid.NewUlid()
	valid := ulid.IsValidUlid(id.UUIDString)
	assert.True(t, valid)

	invalid := ulid.IsValidUlid("invalid-ulid")
	assert.False(t, invalid)
}
