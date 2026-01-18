// Package json provides a drop-in replacement for encoding/json.
// This version uses the standard library for Go 1.24 compatibility.
package json

import (
	stdjson "encoding/json"
	"io"
)

func Marshal(v any) ([]byte, error) {
	return stdjson.Marshal(v)
}

func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return stdjson.MarshalIndent(v, prefix, indent)
}

func Unmarshal(data []byte, v any) error {
	return stdjson.Unmarshal(data, v)
}

func Valid(data []byte) bool {
	return stdjson.Valid(data)
}

type (
	RawMessage            = stdjson.RawMessage
	Number                = stdjson.Number
	Marshaler             = stdjson.Marshaler
	Unmarshaler           = stdjson.Unmarshaler
	Delim                 = stdjson.Delim
	Token                 = stdjson.Token
	InvalidUnmarshalError = stdjson.InvalidUnmarshalError
	MarshalerError        = stdjson.MarshalerError
	SyntaxError           = stdjson.SyntaxError
	UnmarshalTypeError    = stdjson.UnmarshalTypeError
	UnsupportedTypeError  = stdjson.UnsupportedTypeError
	UnsupportedValueError = stdjson.UnsupportedValueError
)

type Encoder struct {
	enc *stdjson.Encoder
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{enc: stdjson.NewEncoder(w)}
}

func (e *Encoder) Encode(v any) error {
	return e.enc.Encode(v)
}

func (e *Encoder) SetIndent(prefix, indent string) {
	e.enc.SetIndent(prefix, indent)
}

func (e *Encoder) SetEscapeHTML(on bool) {
	e.enc.SetEscapeHTML(on)
}

type Decoder struct {
	dec *stdjson.Decoder
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{dec: stdjson.NewDecoder(r)}
}

func (d *Decoder) Decode(v any) error {
	return d.dec.Decode(v)
}

func (d *Decoder) UseNumber() {
	d.dec.UseNumber()
}

func (d *Decoder) DisallowUnknownFields() {
	d.dec.DisallowUnknownFields()
}

func (d *Decoder) More() bool {
	return d.dec.More()
}

func (d *Decoder) InputOffset() int64 {
	return d.dec.InputOffset()
}

func (d *Decoder) Buffered() io.Reader {
	return d.dec.Buffered()
}
