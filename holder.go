package sbc

import (
	"context"
	"sync"
)

// Holder represents a synchronized container that holds a value of any type.
type Holder[T any] struct {
	mu    sync.RWMutex
	cond  *sync.Cond
	value T
}

// NewHolder creates a new envelope with the given value
func NewHolder[T any](value T) *Holder[T] {
	h := &Holder[T]{value: value}
	h.cond = sync.NewCond(&h.mu)
	return h
}

// GetValue returns the value of the envelope
func (e *Holder[T]) GetValue() T {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.value
}

// SetValue sets the value of the envelope
func (e *Holder[T]) setValue(value T) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.value = value
	// signal all goroutines waiting on the condition
	e.cond.Broadcast()
}

// Updates returns a receive-only channel that sends updates of type T.
// The channel is closed when the context is done.
// It continuously sends the current value of the Holder to the channel.
// If the context is done, it stops sending updates and closes the channel.
func (e *Holder[T]) Updates(ctx context.Context) <-chan T {
	go func() {
		// unfreeze the goroutine when the context is done
		<-ctx.Done()
		e.cond.Signal()
	}()
	out := make(chan T)
	go func() {
		defer close(out)
		for {
			// wait for the context to be done or the value to change
			e.mu.Lock()
			e.cond.Wait()
			val := e.value
			e.mu.Unlock()
			if ctx.Err() != nil {
				return
			}
			out <- val
		}
	}()
	return out
}
