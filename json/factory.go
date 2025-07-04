package json

import (
	"io"

	"github.com/fsvxavier/nexs-lib/json/interfaces"
	"github.com/fsvxavier/nexs-lib/json/providers/goccy"
	"github.com/fsvxavier/nexs-lib/json/providers/jsoniter"
	"github.com/fsvxavier/nexs-lib/json/providers/jsonparser"
	"github.com/fsvxavier/nexs-lib/json/providers/stdlib"
)

// ProviderType represents the available JSON provider implementations
type ProviderType string

const (
	// Stdlib represents the standard library encoding/json
	Stdlib ProviderType = "stdlib"

	// JSONIter represents the github.com/json-iterator/go library
	JSONIter ProviderType = "jsoniter"

	// JSONParser represents the github.com/buger/jsonparser library
	JSONParser ProviderType = "jsonparser"

	// GoccyJSON represents the github.com/goccy/go-json library
	GoccyJSON ProviderType = "goccy"

	// Default provider used when none is specified
	DefaultProvider = Stdlib
)

// New returns a new instance of a JSON provider based on the specified type
func New(providerType ProviderType) interfaces.Provider {
	switch providerType {
	case JSONIter:
		return jsoniter.New()
	case JSONParser:
		return jsonparser.New()
	case GoccyJSON:
		return goccy.New()
	case Stdlib:
		return stdlib.New()
	default:
		return stdlib.New()
	}
}

// Exposed convenience functions using the default provider

// Marshal returns the JSON encoding of v using the default provider
func Marshal(v interface{}) ([]byte, error) {
	return New(DefaultProvider).Marshal(v)
}

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v using the default provider
func Unmarshal(data []byte, v interface{}) error {
	return New(DefaultProvider).Unmarshal(data, v)
}

// NewDecoder returns a new decoder that reads from r using the default provider
func NewDecoder(r io.Reader) interfaces.Decoder {
	return New(DefaultProvider).NewDecoder(r)
}

// NewEncoder returns a new encoder that writes to w using the default provider
func NewEncoder(w io.Writer) interfaces.Encoder {
	return New(DefaultProvider).NewEncoder(w)
}

// Valid reports whether data is a valid JSON encoding using the default provider
func Valid(data []byte) bool {
	return New(DefaultProvider).Valid(data)
}

// DecodeReader decodes a reader into v using the default provider
func DecodeReader(r io.Reader, v interface{}) error {
	return New(DefaultProvider).DecodeReader(r, v)
}

// Encode returns the JSON encoding of v using the default provider
func Encode(v interface{}) ([]byte, error) {
	return New(DefaultProvider).Encode(v)
}
