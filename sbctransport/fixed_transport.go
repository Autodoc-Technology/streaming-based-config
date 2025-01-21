package sbctransport

import (
	"context"
)

// FixedTransport is transport that returns a predefined payload for data retrieval operations.
type FixedTransport struct {
	payload []byte
}

// NewFixedTransport creates a FixedTransport instance with a predefined payload for retrieval operations.
func NewFixedTransport(payload []byte) *FixedTransport {
	return &FixedTransport{payload: payload}
}

// Current retrieves the predefined payload associated with the FixedTransport instance.
func (d FixedTransport) Current(_ context.Context, _ string) ([]byte, error) {
	return d.payload, nil
}

// Updates creates a channel for receiving updates, closing it when the provided context is done.
func (d FixedTransport) Updates(ctx context.Context, _ string) (<-chan []byte, error) {
	ch := make(chan []byte)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}
