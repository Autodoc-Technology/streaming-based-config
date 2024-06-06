package sbcencoder

import (
	"encoding/json"
	"fmt"
)

// JsonEncoder is a type that provides encoding and decoding functionality for JSON data.
type JsonEncoder struct{}

// NewJsonEncoder creates a new JsonEncoder.
func NewJsonEncoder() *JsonEncoder {
	return &JsonEncoder{}
}

// Encode encodes the given value v into a byte slice.
func (j JsonEncoder) Encode(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to encode value: %w", err)
	}
	return b, nil
}

// Decode decodes the byte slice into type T.
func (j JsonEncoder) Decode(data []byte, v any) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to decode value: %w", err)
	}
	return nil
}
