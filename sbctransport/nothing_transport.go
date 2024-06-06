package sbctransport

import (
	"context"
)

// NothingTransport represents a type that doesn't actually perform any transportation.
// It is a struct with no fields and implements the Current and Updates methods
// of the transport interface. Both methods return empty results or nil.
type NothingTransport struct{}

// NewNothingTransport creates a new NothingTransport.
func NewNothingTransport() *NothingTransport {
	return &NothingTransport{}
}

func (n NothingTransport) Current(_ context.Context, _ string) ([]byte, error) {
	return []byte("{}"), nil
}

func (n NothingTransport) Updates(_ context.Context, _ string) (<-chan []byte, error) {
	return nil, nil
}
