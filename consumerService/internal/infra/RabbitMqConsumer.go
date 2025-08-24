package infra

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sam9291/go-pubsub-demo/consumer/internal/app"
)

type RabbitMqConsumer struct {
}

func (r *RabbitMqConsumer) StartConsuming() {

  forever := make(chan bool)
  go func() {

    for i := 0; true; i++ {
      id := uuid.New()
      fmt.Printf("consumer iteration %d. GUID = %s\n", i, &id)
      time.Sleep(2 * time.Second)
    }

  }()
  <-forever
}

func New() app.Consumer {
	return &RabbitMqConsumer{}
}
