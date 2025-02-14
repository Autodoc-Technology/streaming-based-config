package sbc

import (
	"context"
	"fmt"
)

// Subscription represents a Subscription to receive updates from a transport.
// It holds a reference to the transport, encoder, holder, and cancel function.
//
// Use the Get() method to get the current value of the Subscription.
//
// Use the GetUpdates() method to get a receive-only channel that sends updates of type T.
//
// Use the start() method to start the Subscription and receive updates from the transport.
//
// Use the stop() method to stop the Subscription and cancel the context.
//
// Use the getAndDecode() method to get and decode the initial value from the transport.
//
// Use the decode() method to decode the byte slice into type T.
type Subscription[T any] struct {
	transport  Transport
	encoder    Encoder
	keyBuilder KeyBuilder[T]

	holder *Holder[T]
	ctx    context.Context
	cancel context.CancelFunc
}

// NewSubscription creates a new subscription
func NewSubscription[T any](transport Transport, encoder Encoder, kb KeyBuilder[T]) *Subscription[T] {
	return &Subscription[T]{transport: transport, encoder: encoder, keyBuilder: kb}
}

// Unsubscribe permanently stops the Subscription and cancel the context.
func (sub *Subscription[T]) Unsubscribe() {
	sub.cancel()
}

// Get returns the current value of the subscription.
func (sub *Subscription[T]) Get() T {
	return sub.holder.GetValue()
}

// GetUpdates returns a receive-only channel that sends updates of type T.
// The channel is closed when the specified context is done.
// It continuously sends the current value of the subscription holder to the channel.
// If the context is done, it stops sending updates and closes the channel.
func (sub *Subscription[T]) GetUpdates() <-chan T {
	return sub.holder.Updates(sub.ctx)
}

// start starts the Subscription and receives updates from the transport.
func (sub *Subscription[T]) start(ctx context.Context) (*Subscription[T], error) {
	sub.ctx, sub.cancel = context.WithCancel(ctx)
	var defT T
	key := sub.keyBuilder.BuildKey(defT)
	// get the initial value from the transport
	val, err := sub.getAndDecode(sub.ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decode initial value: %w", err)
	}
	sub.holder = NewHolder(*val)
	// iterate the transport updates and update the holder
	updates, err := sub.transport.Updates(sub.ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get updates from transport: %w", err)
	}
	go func() {
		for upd := range updates {
			val, err := sub.decode(upd)
			if err != nil {
				continue
			}
			sub.holder.setValue(val)
		}
	}()
	return sub, nil
}

// getAndDecode gets and decodes the initial value from the transport.
func (sub *Subscription[T]) getAndDecode(ctx context.Context, key string) (*T, error) {
	b, err := sub.transport.Current(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get current value: %w", err)
	}
	val, err := sub.decode(b)
	if err != nil {
		return nil, fmt.Errorf("failed to decode value: %w", err)
	}
	return &val, nil
}

// decode decodes the byte slice into type T.
func (sub *Subscription[T]) decode(b []byte) (T, error) {
	var t T
	if err := sub.encoder.Decode(b, &t); err != nil {
		return t, err
	}
	return t, nil
}
