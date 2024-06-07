package sbc

import (
	"context"
)

type Transport interface {
	Current(ctx context.Context, key string) ([]byte, error)
	Updates(ctx context.Context, key string) (<-chan []byte, error)
}
