package main

import (
	"context"
	sbc "github.com/Autodoc-Technology/streaming-based-config"
	"github.com/Autodoc-Technology/streaming-based-config/sbckey"
	"github.com/Autodoc-Technology/streaming-based-config/sbctransport"
	capi "github.com/hashicorp/consul/api"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

// Config is a struct that contains two string fields and a generic field.
// type-parameter (generic) is used to show that the library can work with any type,
// but sbckey.ConsulDefaultKeyBuilder should be used to generate a valid Consul key.
type Config[T any] struct {
	Key1 string `json:"key1"`
	Key2 string `json:"key2"`
	Key  T      `json:"key"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Get a new client
	client, err := capi.NewClient(capi.DefaultConfig())
	if err != nil {
		log.Fatalf("Failed to create Consul client: %v", err)
	}

	// Get a handle to the KV API
	kv := client.KV()

	sub := sbc.NewSubscriber[Config[string]](
		sbctransport.NewConsulTransport(kv, sbctransport.WithUpdateInterval(3*time.Second)),
		sbckey.ConsulDefaultKeyBuilder[Config[string]](),
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
