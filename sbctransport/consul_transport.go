package sbctransport

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	capi "github.com/hashicorp/consul/api"
)

// ConsulTransport represents a transport mechanism for accessing and manipulating data stored in Consul Key-Value store.
type ConsulTransport struct {
	kv             *capi.KV
	updateInterval time.Duration
	updateJitter   time.Duration
}

// defaultUpdateInterval is the default update interval for ConsulTransport.
const defaultUpdateInterval = 10 * time.Second

// defaultUpdateJitter is the default update jitter for ConsulTransport.
const defaultUpdateJitter = 1 * time.Second

// NewConsulTransport creates a new ConsulTransport.
func NewConsulTransport(kv *capi.KV, opts ...ConsulTransportOpt) *ConsulTransport {
	c := &ConsulTransport{kv: kv, updateInterval: defaultUpdateInterval, updateJitter: defaultUpdateJitter}
	// Apply options
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Current retrieves the current value for the specified key from the Consul Key-Value store.
func (c ConsulTransport) Current(ctx context.Context, key string) ([]byte, error) {
	pair, _, err := c.get(ctx, key)
	if err != nil {
		return nil, err
	}
	return pair.Value, nil
}

// Updates creates a channel that streams updates for a given key in the Consul KV store, emitting updated values.
func (c ConsulTransport) Updates(ctx context.Context, key string) (<-chan []byte, error) {
	ch := make(chan []byte)
	go func() {
		defer close(ch)
		lastIndex := uint64(0)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(c.getIntervalWithJitter()):
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

// getIntervalWithJitter returns the update interval with +- updateJitter.
func (c ConsulTransport) getIntervalWithJitter() time.Duration {
	// Generate a random float64 number between -1 and 1
	randomSign := func() int {
		if rand.Intn(2) == 0 {
			return -1
		}
		return 1
	}
	// Calculate the jitter
	jitter := c.updateJitter * time.Duration(randomSign())
	// Return the update interval with jitter
	return c.updateInterval + jitter
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

// WithUpdateJitter is a ConsulTransportOpt function that sets the update jitter for ConsulTransport.
// It accepts a duration parameter and updates the ConsulTransport's update jitter.
func WithUpdateJitter(jitter time.Duration) ConsulTransportOpt {
	return func(c *ConsulTransport) {
		// Ensure the update jitter is at least 10 milliseconds
		if jitter < 10*time.Millisecond {
			jitter = 10 * time.Millisecond
		}
		c.updateJitter = jitter
	}
}
