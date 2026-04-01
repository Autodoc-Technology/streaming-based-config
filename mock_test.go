package sbc

import "context"

type mockTransport struct {
	currentFn func(ctx context.Context, key string) ([]byte, error)
	updatesFn func(ctx context.Context, key string) (<-chan []byte, error)
}

func (m *mockTransport) Current(ctx context.Context, key string) ([]byte, error) {
	return m.currentFn(ctx, key)
}

func (m *mockTransport) Updates(ctx context.Context, key string) (<-chan []byte, error) {
	return m.updatesFn(ctx, key)
}

type mockEncoder struct {
	encodeFn func(v any) ([]byte, error)
	decodeFn func(data []byte, v any) error
}

func (m *mockEncoder) Encode(v any) ([]byte, error) {
	return m.encodeFn(v)
}

func (m *mockEncoder) Decode(data []byte, v any) error {
	return m.decodeFn(data, v)
}
