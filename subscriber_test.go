package sbc

import (
	"context"
	"errors"
	"testing"

	"github.com/Autodoc-Technology/streaming-based-config/sbcencoder"
)

func TestSubscriber_Subscribe_Success(t *testing.T) {
	enc := sbcencoder.NewJsonEncoder()
	updCh := make(chan []byte)
	transport := &mockTransport{
		currentFn: func(ctx context.Context, key string) ([]byte, error) {
			return []byte(`{"value": 7}`), nil
		},
		updatesFn: func(ctx context.Context, key string) (<-chan []byte, error) {
			return updCh, nil
		},
	}

	subscriber := NewSubscriber[testConfig](transport, testKeyBuilder(), WithEncoder(enc))
	sub, err := subscriber.Subscribe(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if sub == nil {
		t.Fatal("expected non-nil subscription")
	}
	defer sub.Unsubscribe()

	if got := sub.Get(); got.Value != 7 {
		t.Errorf("expected Value=7, got %d", got.Value)
	}
}

func TestSubscriber_Subscribe_Error(t *testing.T) {
	transport := &mockTransport{
		currentFn: func(ctx context.Context, key string) ([]byte, error) {
			return nil, errors.New("transport error")
		},
		updatesFn: func(ctx context.Context, key string) (<-chan []byte, error) {
			return make(chan []byte), nil
		},
	}

	subscriber := NewSubscriber[testConfig](transport, testKeyBuilder())
	_, err := subscriber.Subscribe(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSubscriber_WithEncoder(t *testing.T) {
	decodeCalled := false
	enc := &mockEncoder{
		encodeFn: func(v any) ([]byte, error) { return nil, nil },
		decodeFn: func(data []byte, v any) error {
			decodeCalled = true
			cfg := v.(*testConfig)
			cfg.Value = 55
			return nil
		},
	}
	updCh := make(chan []byte)
	transport := &mockTransport{
		currentFn: func(ctx context.Context, key string) ([]byte, error) {
			return []byte(`{"value": 55}`), nil
		},
		updatesFn: func(ctx context.Context, key string) (<-chan []byte, error) {
			return updCh, nil
		},
	}

	subscriber := NewSubscriber[testConfig](transport, testKeyBuilder(), WithEncoder(enc))
	sub, err := subscriber.Subscribe(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sub.Unsubscribe()

	if !decodeCalled {
		t.Error("expected custom encoder Decode to be called")
	}
	if got := sub.Get(); got.Value != 55 {
		t.Errorf("expected Value=55, got %d", got.Value)
	}
}
