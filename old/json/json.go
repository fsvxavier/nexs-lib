package json

import (
	"bytes"
	"encoding/json"
	"io"
)

func DecodeReader(r io.Reader, v interface{}) error {
	d := json.NewDecoder(r)
	d.UseNumber()
	err := d.Decode(v)
	return err
}

func Decode(b []byte, v interface{}) error {
	return DecodeReader(bytes.NewBuffer(b), v)
}

func Encode(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	return b, err
}
