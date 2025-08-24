package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sam9291/go-pubsub-demo/publisher/internal/app"
	"github.com/sam9291/go-pubsub-demo/publisher/internal/infra"
)

func main(){

  fmt.Println("publisher service starting")

  configPath := os.Getenv("APP_CONFIG")

  fmt.Printf("loading application config from %s\n", configPath)

  config, err := app.LoadAppConfig(configPath)
  if err != nil {
    fmt.Printf("failed to load app config. error: %s\n", err)
    panic(1)
  }

  publisher := infra.NewRabbitMqPublisher(*config)

  for { // Infinite loop
    err = publisher.StartPublishing()
    if err != nil {
      fmt.Printf("failed to start publishing, retrying to reconnect in 1 sec. error: %s\n", err)
    }

    time.Sleep(1*time.Second)
  }
}
