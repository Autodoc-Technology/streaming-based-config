package main

import (
	"context"
	"fmt"
	"log"
	"time"

	sbc "github.com/Autodoc-Technology/streaming-based-config"
	"github.com/Autodoc-Technology/streaming-based-config/sbcencoder"
	"github.com/Autodoc-Technology/streaming-based-config/sbckey"
	"github.com/Autodoc-Technology/streaming-based-config/sbctransport"
)

// AppConfig represents our application configuration.
type AppConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Debug   bool   `json:"debug"`
	Port    int    `json:"port"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Define the base (default) configuration.
	defaultConfig := `{"name": "MyApp", "version": "1.0.0", "debug": false, "port": 8080}`
	baseTransport := sbctransport.NewFixedTransport([]byte(defaultConfig))

	// 2. Define an override configuration (e.g., from environment or another source).
	// Here, we override 'debug' to true and 'port' to 9090. 'name' and 'version' remain unchanged.
	overrideConfig := `{"debug": true, "port": 9090}`
	overrideTransport := sbctransport.NewFixedTransport([]byte(overrideConfig))

	// 3. Create the ChainTransport.
	// The order matters: transports later in the list override earlier ones.
	chainTransport := sbctransport.NewChainTransport(
		sbcencoder.NewJsonEncoder(),
		baseTransport,     // Level 1: Defaults
		overrideTransport, // Level 2: Overrides
	)

	// 4. Create the Subscriber using the ChainTransport.
	sub := sbc.NewSubscriber[AppConfig](
		chainTransport,
		sbckey.DefaultKeyBuilder[AppConfig](),
	)

	// 5. Subscribe to configuration updates.
	confSubs, err := sub.Subscribe(ctx)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	defer confSubs.Unsubscribe()

	// 6. Retrieve and print the merged configuration.
	currentConfig := confSubs.Get()
	fmt.Println("--- Merged Configuration ---")
	fmt.Printf("Name:    %s (from default)\n", currentConfig.Name)
	fmt.Printf("Version: %s (from default)\n", currentConfig.Version)
	fmt.Printf("Debug:   %v (from override)\n", currentConfig.Debug)
	fmt.Printf("Port:    %d (from override)\n", currentConfig.Port)

	// Keep the example running briefly to show it doesn't crash
	time.Sleep(1 * time.Second)
}
