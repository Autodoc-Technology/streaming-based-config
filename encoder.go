package sbc

// Encoder encodes the given value v into a byte slice.
// Returns the encoded data as a byte slice and an error if encoding fails.
type Encoder interface {

	// Encode encodes the input value v into a byte slice and returns it along with an error if the encoding fails.
	Encode(v any) ([]byte, error)

	// Decode decodes the provided byte slice into the target variable v of any type.
	// Returns an error if the decoding process fails.
	Decode(data []byte, v any) error
}
