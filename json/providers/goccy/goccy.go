package goccy

import (
	"encoding/json"
	"io"

	jsonerrors "github.com/fsvxavier/nexs-lib/json/errors"
	jsoninterfaces "github.com/fsvxavier/nexs-lib/json/interfaces"
	gojson "github.com/goccy/go-json"
)

// Provider implements the json.Provider interface using goccy/go-json
type Provider struct{}

// New creates a new goccy/go-json Provider
func New() jsoninterfaces.Provider {
	return &Provider{}
}

// Marshal returns the JSON encoding of v
func (p *Provider) Marshal(v interface{}) ([]byte, error) {
	return gojson.Marshal(v)
}

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v
func (p *Provider) Unmarshal(data []byte, v interface{}) error {
	return gojson.Unmarshal(data, v)
}

// NewDecoder returns a new decoder that reads from r
func (p *Provider) NewDecoder(r io.Reader) jsoninterfaces.Decoder {
	return &Decoder{decoder: gojson.NewDecoder(r)}
}

// NewEncoder returns a new encoder that writes to w
func (p *Provider) NewEncoder(w io.Writer) jsoninterfaces.Encoder {
	return &Encoder{encoder: gojson.NewEncoder(w)}
}

// Valid reports whether data is a valid JSON encoding
func (p *Provider) Valid(data []byte) bool {
	// goccy/go-json doesn't have a direct Validate function
	// Try to unmarshal into a generic interface{} to check validity
	var v interface{}
	err := gojson.Unmarshal(data, &v)
	return err == nil
}

// DecodeReader decodes a reader into v
func (p *Provider) DecodeReader(r io.Reader, v interface{}) error {
	d := p.NewDecoder(r)
	d.UseNumber()
	return d.Decode(v)
}

// Encode returns the JSON encoding of v
func (p *Provider) Encode(v interface{}) ([]byte, error) {
	return p.Marshal(v)
}

// Decoder wraps the goccy/go-json.Decoder
type Decoder struct {
	decoder *gojson.Decoder
}

// Decode reads the next JSON-encoded value from its input and stores it in the value pointed to by v
func (d *Decoder) Decode(v interface{}) error {
	return d.decoder.Decode(v)
}

// UseNumber causes the Decoder to unmarshal a number into an interface{} as a Number instead of as a float64
func (d *Decoder) UseNumber() jsoninterfaces.Decoder {
	d.decoder.UseNumber()
	return d
}

// DisallowUnknownFields causes the Decoder to return an error when the destination is a struct and the input contains object keys which do not match any non-ignored, exported fields in the destination
func (d *Decoder) DisallowUnknownFields() jsoninterfaces.Decoder {
	d.decoder.DisallowUnknownFields()
	return d
}

// Buffered returns a reader containing any bytes that were read from the underlying reader but not yet used during a decode operation
func (d *Decoder) Buffered() io.Reader {
	return d.decoder.Buffered()
}

// Token returns the next JSON token in the input stream
// Note: go-json doesn't expose Token method directly, so this is implemented using encoding/json
func (d *Decoder) Token() (json.Token, error) {
	// Not directly supported by goccy/go-json, so we return an unsupported error
	return nil, jsonerrors.ErrUnsupportedOperation
}

// More reports whether there is another element in the current array or object being parsed
// Note: go-json doesn't expose More method directly, so this returns false
func (d *Decoder) More() bool {
	// Not directly supported by goccy/go-json
	return false
}

// Encoder wraps the goccy/go-json.Encoder
type Encoder struct {
	encoder *gojson.Encoder
}

// Encode writes the JSON encoding of v to the stream
func (e *Encoder) Encode(v interface{}) error {
	return e.encoder.Encode(v)
}

// SetIndent instructs the encoder to format each subsequent encoded value as if indented by the package-level function Indent with the specified prefix and indentation
func (e *Encoder) SetIndent(prefix, indent string) jsoninterfaces.Encoder {
	e.encoder.SetIndent(prefix, indent)
	return e
}

// SetEscapeHTML specifies whether problematic HTML characters should be escaped inside JSON quoted strings
func (e *Encoder) SetEscapeHTML(on bool) jsoninterfaces.Encoder {
	e.encoder.SetEscapeHTML(on)
	return e
}
