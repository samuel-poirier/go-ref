package main

import (
	"fmt"

	"github.com/sam9291/go-pubsub-demo/consumer/internal/infra"
)

func main(){

  fmt.Println("consumer service starting")

  consumer := infra.New()
  consumer.StartConsuming()
}
