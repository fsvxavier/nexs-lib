package jsonparser

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/buger/jsonparser"
	jsonerrors "github.com/fsvxavier/nexs-lib/json/errors"
	jsoninterfaces "github.com/fsvxavier/nexs-lib/json/interfaces"
)

// Provider implements the json.Provider interface using jsonparser
type Provider struct{}

// New creates a new jsonparser Provider
// Note: jsonparser is a streaming parser with limited functionality compared to other providers
// It doesn't support full encoding/decoding like other libraries, so we fallback to stdlib
// for operations that are not directly supported
func New() jsoninterfaces.Provider {
	return &Provider{}
}

// Marshal returns the JSON encoding of v
// Note: jsonparser doesn't provide Marshal function, so we use the standard library
func (p *Provider) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v
// Note: jsonparser has a different API for parsing, but we use standard library for compatibility
func (p *Provider) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// NewDecoder returns a new decoder that reads from r
// Note: jsonparser doesn't have direct Decoder/Encoder concepts like the standard library
func (p *Provider) NewDecoder(r io.Reader) jsoninterfaces.Decoder {
	return &Decoder{reader: r}
}

// NewEncoder returns a new encoder that writes to w
func (p *Provider) NewEncoder(w io.Writer) jsoninterfaces.Encoder {
	return &Encoder{writer: w}
}

// Valid reports whether data is a valid JSON encoding
func (p *Provider) Valid(data []byte) bool {
	// jsonparser doesn't provide a direct Valid function
	// We can check by trying to get the type of the root element
	_, valueType, _, err := jsonparser.Get(data)
	return err == nil && valueType != jsonparser.NotExist
}

// DecodeReader decodes a reader into v
func (p *Provider) DecodeReader(r io.Reader, v interface{}) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return p.Unmarshal(data, v)
}

// Encode returns the JSON encoding of v
func (p *Provider) Encode(v interface{}) ([]byte, error) {
	return p.Marshal(v)
}

// Decoder implements a basic wrapper around jsonparser
type Decoder struct {
	reader io.Reader
	data   []byte
	err    error
}

// Decode reads the next JSON-encoded value from its input and stores it in the value pointed to by v
func (d *Decoder) Decode(v interface{}) error {
	// If we haven't read the data yet, read it all
	if d.data == nil && d.err == nil {
		d.data, d.err = io.ReadAll(d.reader)
		if d.err != nil {
			return d.err
		}
	}

	// Use standard JSON for decoding since jsonparser doesn't provide a direct equivalent
	return json.Unmarshal(d.data, v)
}

// UseNumber causes the Decoder to unmarshal a number into an interface{} as a Number instead of as a float64
func (d *Decoder) UseNumber() jsoninterfaces.Decoder {
	// This is a no-op for jsonparser since we're using standard library for decoding
	return d
}

// DisallowUnknownFields causes the Decoder to return an error when the destination is a struct and the input contains object keys which do not match any non-ignored, exported fields in the destination
func (d *Decoder) DisallowUnknownFields() jsoninterfaces.Decoder {
	// This is a no-op for jsonparser since we're using standard library for decoding
	return d
}

// Buffered returns a reader containing any bytes that were read from the underlying reader but not yet used during a decode operation
func (d *Decoder) Buffered() io.Reader {
	// Not directly supported by jsonparser
	if d.data != nil {
		return bytes.NewReader(d.data)
	}
	return nil
}

// Token returns the next JSON token in the input stream
func (d *Decoder) Token() (json.Token, error) {
	// Not directly supported by jsonparser, so we return an unsupported error
	return nil, jsonerrors.ErrUnsupportedOperation
}

// More reports whether there is another element in the current array or object being parsed
func (d *Decoder) More() bool {
	// Not directly supported by jsonparser
	return false
}

// Encoder wraps a writer for JSON encoding
type Encoder struct {
	writer io.Writer
}

// Encode writes the JSON encoding of v to the stream
func (e *Encoder) Encode(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = e.writer.Write(data)
	return err
}

// SetIndent instructs the encoder to format each subsequent encoded value as if indented by the package-level function Indent with the specified prefix and indentation
func (e *Encoder) SetIndent(prefix, indent string) jsoninterfaces.Encoder {
	// Not directly supported by jsonparser
	return e
}

// SetEscapeHTML specifies whether problematic HTML characters should be escaped inside JSON quoted strings
func (e *Encoder) SetEscapeHTML(on bool) jsoninterfaces.Encoder {
	// Not directly supported by jsonparser
	return e
}
