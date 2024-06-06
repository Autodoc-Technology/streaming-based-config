package sbc

// Encoder encodes the given value v into a byte slice.
// Returns the encoded data as a byte slice and an error if encoding fails.
type Encoder interface {
	Encode(v any) ([]byte, error)
	Decode(data []byte, v any) error
}
