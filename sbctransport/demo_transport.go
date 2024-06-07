package sbctransport

import (
	"context"
	"log"
	"time"
)

type DemoTransport struct {
	payload  []byte
	interval time.Duration
}

func NewDemoTransport(payload []byte, interval time.Duration) *DemoTransport {
	return &DemoTransport{payload: payload, interval: interval}
}

func (d DemoTransport) Current(ctx context.Context, _ string) ([]byte, error) {
	return d.payload, nil
}

func (d DemoTransport) Updates(ctx context.Context, key string) (<-chan []byte, error) {
	log.Println("Updates key:", key)
	ch := make(chan []byte)
	go func() {
		ticker := time.NewTicker(d.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			case <-ticker.C:
				ch <- d.payload
			}
		}
	}()
	return ch, nil
}
