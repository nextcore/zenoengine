package fastjson

import (
	gojson "github.com/goccy/go-json"
	"io"
)

// Marshal serializes v to JSON using the fast encoder.
// This is a drop-in replacement for encoding/json.Marshal
// but is 2-3x faster.
func Marshal(v interface{}) ([]byte, error) {
	return gojson.Marshal(v)
}

// Unmarshal deserializes JSON data into v using the fast decoder.
// This is a drop-in replacement for encoding/json.Unmarshal.
func Unmarshal(data []byte, v interface{}) error {
	return gojson.Unmarshal(data, v)
}

// NewEncoder creates a new JSON encoder that writes to w.
// This is a drop-in replacement for encoding/json.NewEncoder
// but uses the faster goccy/go-json implementation.
func NewEncoder(w io.Writer) *gojson.Encoder {
	return gojson.NewEncoder(w)
}

// NewDecoder creates a new JSON decoder that reads from r.
// This is a drop-in replacement for encoding/json.NewDecoder.
func NewDecoder(r io.Reader) *gojson.Decoder {
	return gojson.NewDecoder(r)
}

// MarshalIndent is like Marshal but applies indentation for pretty-printing.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return gojson.MarshalIndent(v, prefix, indent)
}
