package interfaces

import (
	"encoding/json"
	"io"
)

// Provider interface defines the common methods that all JSON providers must implement
type Provider interface {
	// Marshal returns the JSON encoding of v
	Marshal(v interface{}) ([]byte, error)

	// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v
	Unmarshal(data []byte, v interface{}) error

	// NewDecoder returns a new decoder that reads from r
	NewDecoder(r io.Reader) Decoder

	// NewEncoder returns a new encoder that writes to w
	NewEncoder(w io.Writer) Encoder

	// Valid reports whether data is a valid JSON encoding
	Valid(data []byte) bool

	// DecodeReader decodes a reader into v
	DecodeReader(r io.Reader, v interface{}) error

	// Encode returns the JSON encoding of v
	Encode(v interface{}) ([]byte, error)
}

// Decoder is an interface for reading and decoding JSON values from a stream
type Decoder interface {
	// Decode reads the next JSON-encoded value from its input and stores it in the value pointed to by v
	Decode(v interface{}) error

	// UseNumber causes the Decoder to unmarshal a number into an interface{} as a Number instead of as a float64
	UseNumber() Decoder

	// DisallowUnknownFields causes the Decoder to return an error when the destination is a struct and the input contains object keys which do not match any non-ignored, exported fields in the destination
	DisallowUnknownFields() Decoder

	// Buffered returns a reader containing any bytes that were read from the underlying reader but not yet used during a decode operation
	Buffered() io.Reader

	// Token returns the next JSON token in the input stream
	Token() (json.Token, error)

	// More reports whether there is another element in the current array or object being parsed
	More() bool
}

// Encoder is an interface for writing JSON values to a stream
type Encoder interface {
	// Encode writes the JSON encoding of v to the stream
	Encode(v interface{}) error

	// SetIndent instructs the encoder to format each subsequent encoded value as if indented by the package-level function Indent with the specified prefix and indentation
	SetIndent(prefix, indent string) Encoder

	// SetEscapeHTML specifies whether problematic HTML characters should be escaped inside JSON quoted strings
	SetEscapeHTML(on bool) Encoder
}
