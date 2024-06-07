package main

import (
	"context"
	sbc "github.com/Autodoc-Technology/streaming-based-config"
	"github.com/Autodoc-Technology/streaming-based-config/sbckey"
	"github.com/Autodoc-Technology/streaming-based-config/sbctransport"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"log"
	"os"
	"os/signal"
	"sync"
)

// Config is a struct that contains two string fields and a generic field.
// type-parameter (generic) is used to show that the library can work with any type,
// but sbckey.NatsDefaultKeyBuilder should be used to generate a valid NATS key.
type Config[T any] struct {
	Key1 string `json:"key1"`
	Key2 string `json:"key2"`
	Key  T      `json:"key"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	natsConn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsConn.Close()

	js, err := jetstream.New(natsConn)
	if err != nil {
		log.Fatalf("Failed to create JetStream: %v", err)
	}

	// Run following console command to initialize NATS KV value
	// `nats kv put deb_sbc_example main.Config_string_ '{"key1":"val1.1","key2":"val2.1", "key":"type-parameter-value.1"}'`
	kv, err := js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: "deb_sbc_example",
	})
	if err != nil {
		log.Fatalf("Failed to create KeyValue: %v", err)
	}

	sub := sbc.NewSubscriber[Config[string]](
		sbctransport.NewNatsTransport(kv),
		sbckey.NatsDefaultKeyBuilder[Config[string]](),
	)
	confSubs, err := sub.Subscribe(ctx)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	defer confSubs.Unsubscribe()

	config := confSubs.Get()
	log.Printf("Get Config: %+v", config)

	var wg sync.WaitGroup
	for i := 0; i < 1; i++ {
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
