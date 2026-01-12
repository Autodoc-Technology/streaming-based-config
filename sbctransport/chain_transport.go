package sbctransport

import (
	"context"
	"fmt"
	"sync"

	sbc "github.com/Autodoc-Technology/streaming-based-config"
)

// ChainTransport aggregates multiple transports and merges their results.
// Transports later in the chain override values from earlier transports.
type ChainTransport struct {
	transports []sbc.Transport
	encoder    sbc.Encoder
}

// NewChainTransport creates a new ChainTransport.
func NewChainTransport(encoder sbc.Encoder, transports ...sbc.Transport) *ChainTransport {
	return &ChainTransport{
		transports: transports,
		encoder:    encoder,
	}
}

// Current retrieves the merged value from all transports.
func (c *ChainTransport) Current(ctx context.Context, key string) ([]byte, error) {
	merged := make(map[string]any)

	for _, t := range c.transports {
		data, err := t.Current(ctx, key)
		if err != nil {
			// If a transport fails, we propagate the error.
			// This ensures that we don't silently return partial config.
			return nil, fmt.Errorf("chain transport: failed to get current value: %w", err)
		}

		var partial map[string]any
		if err := c.encoder.Decode(data, &partial); err != nil {
			return nil, fmt.Errorf("chain transport: failed to decode value: %w", err)
		}
		merge(merged, partial)
	}

	return c.encoder.Encode(merged)
}

// Updates returns a channel that streams merged updates.
func (c *ChainTransport) Updates(ctx context.Context, key string) (<-chan []byte, error) {
	outCh := make(chan []byte)
	
	currentStates := make([][]byte, len(c.transports))
	var mu sync.Mutex

	// Helper to merge and broadcast
	broadcast := func() {
		mu.Lock()
		defer mu.Unlock()

		merged := make(map[string]any)
		for _, state := range currentStates {
			if state == nil {
				continue
			}
			var partial map[string]any
			// Best effort decoding. If decode fails during update, we skip that layer's contribution?
			// Or should we log? For now, skipping/ignoring is safer than crashing updates.
			if err := c.encoder.Decode(state, &partial); err == nil {
				merge(merged, partial)
			}
		}

		encoded, err := c.encoder.Encode(merged)
		if err != nil {
			return
		}

		select {
		case outCh <- encoded:
		case <-ctx.Done():
		}
	}

	// Initialize subscriptions
	for i, t := range c.transports {
		// Subscribe first to capture updates happening during init
		updCh, err := t.Updates(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("chain transport: failed to subscribe: %w", err)
		}

		// Fetch current state
		val, err := t.Current(ctx, key)
		if err == nil {
			mu.Lock()
			currentStates[i] = val
			mu.Unlock()
		}

		// Consume updates
		go func(index int, ch <-chan []byte) {
			for val := range ch {
				mu.Lock()
				currentStates[index] = val
				mu.Unlock()
				broadcast()
			}
		}(i, updCh)
	}

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		close(outCh)
	}()

	return outCh, nil
}

// merge recursively merges src into dst.
// Maps are merged recursively. Other values in dst are overwritten by src.
func merge(dst, src map[string]any) {
	for k, v := range src {
		if srcMap, ok := v.(map[string]any); ok {
			if dstVal, ok := dst[k]; ok {
				if dstMap, ok := dstVal.(map[string]any); ok {
					merge(dstMap, srcMap)
					continue
				}
			}
		}
		dst[k] = v
	}
}
