package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	sbc "github.com/Autodoc-Technology/streaming-based-config"
	"github.com/Autodoc-Technology/streaming-based-config/sbckey"
	"github.com/Autodoc-Technology/streaming-based-config/sbctransport"
)

// AppConfig holds the application configuration loaded from Kafka.
// Publish a JSON value to the topic "main.AppConfig" (derived automatically
// by KafkaDefaultKeyBuilder) to update the running application.
//
// Example producer command (kafka-console-producer):
//
//	echo '{"env":"prod","debug":false}' | \
//	  kafka-console-producer.sh --broker-list localhost:9092 --topic main.AppConfig
type AppConfig struct {
	Env   string `json:"env"`
	Debug bool   `json:"debug"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	brokers := []string{"localhost:9092"}

	transport, err := sbctransport.NewKafkaTransport(
		brokers,
		sbctransport.WithKafkaGroupID("my-app-config-consumer"),
	)
	if err != nil {
		log.Fatalf("Failed to create Kafka transport: %v", err)
	}

	sub := sbc.NewSubscriber[AppConfig](
		transport,
		sbckey.KafkaDefaultKeyBuilder[AppConfig](),
	)

	confSubs, err := sub.Subscribe(ctx)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	defer confSubs.Unsubscribe()

	log.Printf("Current config: %+v", confSubs.Get())

	go func() {
		for val := range confSubs.GetUpdates() {
			log.Printf("Config updated: %+v", val)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down.")
}
