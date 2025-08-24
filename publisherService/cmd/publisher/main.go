package main

import (
	"fmt"

	"github.com/sam9291/go-pubsub-demo/publisher/internal/infra"
)

func main(){

  fmt.Println("publisher service starting")

  publisher := infra.New()

  publisher.StartPublishing()
}
