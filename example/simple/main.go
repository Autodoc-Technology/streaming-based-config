package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	sbc "github.com/Autodoc-Technology/streaming-based-config"
	"github.com/Autodoc-Technology/streaming-based-config/sbckey"
	"github.com/Autodoc-Technology/streaming-based-config/sbctransport"
)

type SimpleConfig struct {
	Value1 int    `json:"value"`
	Value2 string `json:"value2"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	sub := sbc.NewSubscriber[SimpleConfig](
		sbctransport.NewDemoTransport([]byte(`{"value": 1, "value2": "hello"}`), 3*time.Second),
		sbckey.DefaultKeyBuilder[SimpleConfig](),
	)
	confSubs, err := sub.Subscribe(ctx)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	defer confSubs.Unsubscribe()

	config := confSubs.Get()
	log.Printf("Get Config: %+v", config)

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			wg.Done()
			for val := range confSubs.GetUpdates() {
				log.Printf("%d Config: %+v", i, val)
			}
		}()
	}

	<-ctx.Done()
	wg.Wait()
}
