package main

import (
	"fmt"
	"os"

	"github.com/sam9291/go-pubsub-demo/consumer/internal/app"
	"github.com/sam9291/go-pubsub-demo/consumer/internal/infra"
)

func main(){

  fmt.Println("consumer service starting")

  configPath := os.Getenv("APP_CONFIG")

  fmt.Printf("loading application config from %s\n", configPath)

  config, err := app.LoadAppConfig(configPath)
  if err != nil {
    fmt.Printf("failed to load app config. error: %s\n", err)
    panic(1)
  }
  consumer := infra.NewRabbitMqConsumer(*config)
  err = consumer.StartConsuming()

  if err != nil {
    fmt.Printf("failed to start consuming. error: %s\n", err)
    panic(1)
  }
}
