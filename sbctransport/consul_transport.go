package sbctransport

import (
	"context"
	"fmt"
	capi "github.com/hashicorp/consul/api"
	"time"
)

// ConsulTransport represents a transport mechanism for accessing and manipulating data stored in Consul Key-Value store.
type ConsulTransport struct {
	kv             *capi.KV
	updateInterval time.Duration
}

// defaultUpdateInterval is the default update interval for ConsulTransport.
const defaultUpdateInterval = 10 * time.Second

// NewConsulTransport creates a new ConsulTransport.
func NewConsulTransport(kv *capi.KV, opts ...ConsulTransportOpt) *ConsulTransport {
	c := &ConsulTransport{kv: kv, updateInterval: defaultUpdateInterval}
	// Apply options
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c ConsulTransport) Current(ctx context.Context, key string) ([]byte, error) {
	pair, _, err := c.get(ctx, key)
	if err != nil {
		return nil, err
	}
	return pair.Value, nil
}

func (c ConsulTransport) Updates(ctx context.Context, key string) (<-chan []byte, error) {
	ch := make(chan []byte)
	go func() {
		defer close(ch)
		lastIndex := uint64(0)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(c.updateInterval):
				pair, meta, err := c.get(ctx, key)
				if err != nil {
					continue
				}
				if meta.LastIndex <= lastIndex {
					continue
				}
				lastIndex = meta.LastIndex
				ch <- pair.Value
			}
		}
	}()
	return ch, nil
}

// ErrEmptyKey is an error that is returned when the key is empty.
var ErrEmptyKey = fmt.Errorf("key is empty")

// get gets the current value for the specified key from Consul KV.
func (c ConsulTransport) get(ctx context.Context, key string) (*capi.KVPair, *capi.QueryMeta, error) {
	opts := &capi.QueryOptions{
		RequireConsistent: true,
	}
	pair, meta, err := c.kv.Get(key, opts.WithContext(ctx))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current value for key '%s' from Consul KV: %w", key, err)
	}
	if pair == nil {
		return nil, meta, fmt.Errorf("failed to get current value for key '%s' from Consul KV: %w", key, ErrEmptyKey)
	}
	return pair, meta, nil
}

// ConsulTransportOpt is a function type that modifies the properties of a ConsulTransport.
// It accepts a pointer to a ConsulTransport instance and updates its properties.
// TODO: add Consul enterprise options, like Namespace, Partition, etc. usefully options.
type ConsulTransportOpt func(*ConsulTransport)

// WithUpdateInterval is a ConsulTransportOpt function that sets the update interval for ConsulTransport.
// It ensures that the update interval is at least 100 milliseconds.
// It accepts a duration parameter and updates the ConsulTransport's update interval.
func WithUpdateInterval(interval time.Duration) ConsulTransportOpt {
	return func(c *ConsulTransport) {
		// Ensure the update interval is at least 100ms
		if interval < 100*time.Millisecond {
			interval = 100 * time.Millisecond
		}
		c.updateInterval = interval
	}
}
