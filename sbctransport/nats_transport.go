package sbctransport

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

// NatsTransport represents a transport mechanism for accessing and manipulating data stored in NATS Key-Value store.
type NatsTransport struct {
	kv jetstream.KeyValue
}

// NewNatsTransport creates a new NatsTransport.
func NewNatsTransport(kv jetstream.KeyValue) *NatsTransport {
	return &NatsTransport{kv: kv}
}

func (n NatsTransport) Current(ctx context.Context, key string) ([]byte, error) {
	entry, err := n.kv.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get current value for key '%s' from NATS KV: %w", key, err)
	}
	return entry.Value(), nil
}

func (n NatsTransport) Updates(ctx context.Context, key string) (<-chan []byte, error) {
	watcher, err := n.kv.Watch(ctx, key, jetstream.UpdatesOnly(), jetstream.IgnoreDeletes())
	if err != nil {
		return nil, fmt.Errorf("failed to watch updates from NATS KV: %w", err)
	}
	ch := make(chan []byte)
	go func() {
		defer close(ch)
		for entry := range watcher.Updates() {
			ch <- entry.Value()
		}
	}()
	return ch, nil
}
