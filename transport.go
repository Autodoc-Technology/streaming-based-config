package sbc

import (
	"context"
)

// Transport defines an interface for accessing current data and subscribing to updates using a key and context.
type Transport interface {

	// Current retrieves the current value associated with the specified key using the given context.
	// Returns the value as a byte slice and an error if retrieval fails.
	Current(ctx context.Context, key string) ([]byte, error)

	// Updates returns a channel that streams updates for the specified key using the given context.
	// It stops and returns an error if the context is canceled.
	Updates(ctx context.Context, key string) (<-chan []byte, error)
}
