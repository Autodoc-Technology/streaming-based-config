package sbc

import (
	"context"
	"fmt"

	"github.com/Autodoc-Technology/streaming-based-config/sbcencoder"
)

// Subscriber is a generic type that represents a subscriber for receiving updates from transport.
// It holds a reference to the transport, subscription, and subscriber options.
//
// Use the Subscribe() method to start a new subscription by creating a new subscription instance and starting it.
// The subscription is started by calling the start() method of the subscription instance.
//
// Use the Unsubscribe() method to stop the subscription by calling the stop() method of the subscription instance.
type Subscriber[T any] struct {
	transport    Transport
	subscription *Subscription[T]
	keyBuilder   KeyBuilder[T]
	opts         subscriberOpts
}

// NewSubscriber creates a new subscriber
func NewSubscriber[T any](transport Transport, kb KeyBuilder[T], opts ...SubscriberOpt) *Subscriber[T] {
	return &Subscriber[T]{
		transport:  transport,
		keyBuilder: kb,
		opts:       newDefaultSubscriberOpts().apply(opts),
	}
}

// Subscribe starts a new subscription by creating a new subscription instance and starting it.
//
// The subscription is started by calling the start() method of the subscription instance.
// This method initializes internal variables, gets and decodes the initial value from the transport,
// and starts receiving updates from the transport in a separate goroutine.
//
// Note that the context parameter is used to control the lifecycle of the subscription.
// When the context is canceled or done, the subscription's stop() method is called to stop the subscription
// and cancel the context.
//
// If the subscription is not successfully started or the context is already done,
// the returned subscription instance will be nil.
func (s *Subscriber[T]) Subscribe(ctx context.Context) (*Subscription[T], error) {
	sub, err := NewSubscription[T](s.transport, s.opts.encoder, s.keyBuilder).start(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start subscription: %w", err)
	}
	s.subscription = sub
	return sub, nil
}

// SubscriberOpt is a function type used to configure a Subscriber instance.
// It modifies the options of the subscriberOpts struct.
type SubscriberOpt func(*subscriberOpts)

// subscriberOpts is a struct that holds the options of a Subscriber instance.
type subscriberOpts struct {
	encoder Encoder
}

// NewDefaultSubscriberOpts creates a new subscriber options with the default values
func newDefaultSubscriberOpts() subscriberOpts {
	return subscriberOpts{
		encoder: sbcencoder.NewJsonEncoder(),
	}
}

// apply applies a list of SubscriberOpts to the subscriberOpts receiver.
//
// It iterates over each opt in opts and calls opt with the receiver o.
// It returns a copy of the modified receiver.
func (o subscriberOpts) apply(opts []SubscriberOpt) subscriberOpts {
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// WithEncoder sets the encoder option of a Subscriber instance.
func WithEncoder(encoder Encoder) SubscriberOpt {
	return func(o *subscriberOpts) {
		o.encoder = encoder
	}
}
