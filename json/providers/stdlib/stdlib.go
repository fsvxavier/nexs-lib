package stdlib

import (
	"encoding/json"
	"io"

	jsoninterfaces "github.com/fsvxavier/nexs-lib/json/interfaces"
)

// Provider implements the json.Provider interface using the standard library encoding/json
type Provider struct{}

// New creates a new stdlib Provider
func New() jsoninterfaces.Provider {
	return &Provider{}
}

// Marshal returns the JSON encoding of v
func (p *Provider) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v
func (p *Provider) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// NewDecoder returns a new decoder that reads from r
func (p *Provider) NewDecoder(r io.Reader) jsoninterfaces.Decoder {
	return &Decoder{decoder: json.NewDecoder(r)}
}

// NewEncoder returns a new encoder that writes to w
func (p *Provider) NewEncoder(w io.Writer) jsoninterfaces.Encoder {
	return &Encoder{encoder: json.NewEncoder(w)}
}

// Valid reports whether data is a valid JSON encoding
func (p *Provider) Valid(data []byte) bool {
	return json.Valid(data)
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

// Decoder wraps the standard library json.Decoder
type Decoder struct {
	decoder *json.Decoder
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
func (d *Decoder) Token() (json.Token, error) {
	return d.decoder.Token()
}

// More reports whether there is another element in the current array or object being parsed
func (d *Decoder) More() bool {
	return d.decoder.More()
}

// Encoder wraps the standard library json.Encoder
type Encoder struct {
	encoder *json.Encoder
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
