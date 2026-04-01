package sbc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Autodoc-Technology/streaming-based-config/sbcencoder"
)

type testConfig struct {
	Value int `json:"value"`
}

func testKeyBuilder() KeyBuilder[testConfig] {
	return KeyBuilderFunc[testConfig](func(testConfig) string { return "test-key" })
}

func TestSubscription_Start_Success(t *testing.T) {
	enc := sbcencoder.NewJsonEncoder()
	updCh := make(chan []byte)
	transport := &mockTransport{
		currentFn: func(ctx context.Context, key string) ([]byte, error) {
			if key != "test-key" {
				t.Errorf("expected key 'test-key', got '%s'", key)
			}
			return []byte(`{"value": 42}`), nil
		},
		updatesFn: func(ctx context.Context, key string) (<-chan []byte, error) {
			return updCh, nil
		},
	}

	sub, err := NewSubscription[testConfig](transport, enc, testKeyBuilder()).start(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sub.Unsubscribe()

	if got := sub.Get(); got.Value != 42 {
		t.Errorf("expected Value=42, got %d", got.Value)
	}
}

func TestSubscription_Start_CurrentError(t *testing.T) {
	enc := sbcencoder.NewJsonEncoder()
	transport := &mockTransport{
		currentFn: func(ctx context.Context, key string) ([]byte, error) {
			return nil, errors.New("transport error")
		},
		updatesFn: func(ctx context.Context, key string) (<-chan []byte, error) {
			return make(chan []byte), nil
		},
	}

	_, err := NewSubscription[testConfig](transport, enc, testKeyBuilder()).start(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSubscription_Start_DecodeError(t *testing.T) {
	enc := &mockEncoder{
		encodeFn: func(v any) ([]byte, error) { return nil, nil },
		decodeFn: func(data []byte, v any) error { return errors.New("decode error") },
	}
	transport := &mockTransport{
		currentFn: func(ctx context.Context, key string) ([]byte, error) {
			return []byte(`{"value": 42}`), nil
		},
		updatesFn: func(ctx context.Context, key string) (<-chan []byte, error) {
			return make(chan []byte), nil
		},
	}

	_, err := NewSubscription[testConfig](transport, enc, testKeyBuilder()).start(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSubscription_Start_UpdatesError(t *testing.T) {
	enc := sbcencoder.NewJsonEncoder()
	transport := &mockTransport{
		currentFn: func(ctx context.Context, key string) ([]byte, error) {
			return []byte(`{"value": 42}`), nil
		},
		updatesFn: func(ctx context.Context, key string) (<-chan []byte, error) {
			return nil, errors.New("updates error")
		},
	}

	_, err := NewSubscription[testConfig](transport, enc, testKeyBuilder()).start(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSubscription_GetUpdates(t *testing.T) {
	enc := sbcencoder.NewJsonEncoder()
	updCh := make(chan []byte, 1)
	transport := &mockTransport{
		currentFn: func(ctx context.Context, key string) ([]byte, error) {
			return []byte(`{"value": 1}`), nil
		},
		updatesFn: func(ctx context.Context, key string) (<-chan []byte, error) {
			return updCh, nil
		},
	}

	sub, err := NewSubscription[testConfig](transport, enc, testKeyBuilder()).start(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer sub.Unsubscribe()

	updatesCh := sub.GetUpdates()
	updCh <- []byte(`{"value": 99}`)

	select {
	case got := <-updatesCh:
		if got.Value != 99 {
			t.Errorf("expected Value=99, got %d", got.Value)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for update")
	}
}

func TestSubscription_Unsubscribe(t *testing.T) {
	enc := sbcencoder.NewJsonEncoder()
	updCh := make(chan []byte)
	transport := &mockTransport{
		currentFn: func(ctx context.Context, key string) ([]byte, error) {
			return []byte(`{"value": 1}`), nil
		},
		updatesFn: func(ctx context.Context, key string) (<-chan []byte, error) {
			return updCh, nil
		},
	}

	sub, err := NewSubscription[testConfig](transport, enc, testKeyBuilder()).start(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	updatesCh := sub.GetUpdates()
	sub.Unsubscribe()

	select {
	case _, ok := <-updatesCh:
		if ok {
			t.Error("expected channel to be closed after Unsubscribe")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for channel to close after Unsubscribe")
	}
}
