package infra

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sam9291/go-pubsub-demo/publisher/internal/app"
)


type RabbitMqPublisher struct {

}

func New() app.Publisher {
  return &RabbitMqPublisher{}
}

func(publisher *RabbitMqPublisher) StartPublishing() {
  forever := make(chan bool)
  go func() {

    for i := 0; true; i++ {
      id := uuid.New()
      fmt.Printf("publisher iteration %d. GUID = %s\n", i, &id)
      time.Sleep(2 * time.Second)
    }

  }()
  <-forever
}
