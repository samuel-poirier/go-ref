package infra

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sam9291/go-pubsub-demo/publisher/internal/app"
	publisherintegrationevents "github.com/sam9291/go-pubsub-demo/publisherIntegrationEvents"
)


type RabbitMqPublisher struct {
  connectionString string

}

func NewRabbitMqPublisher(config app.AppConfig) app.Publisher {
  return &RabbitMqPublisher{
    connectionString: config.ConnectionStrings.RabbitMq,
  }
}

func(publisher *RabbitMqPublisher) StartPublishing() error {

  conn, err := amqp.Dial(publisher.connectionString)

  if err != nil {
    return err
  }

  defer conn.Close()

  ch, err := conn.Channel()

  if err != nil {
    return err
  }

  defer ch.Close()

  q, err := ch.QueueDeclare(
    "demo-queue", // name
    true,   // durable
    false,   // delete when unused
    false,   // exclusive
    false,   // no-wait
    nil,     // arguments
  )

  if err != nil {
    return err
  }

  forever := make(chan bool)

  go func() {

    for i := 0; true; i++ {
      id := uuid.New()
      fmt.Printf("publisher iteration %d. GUID = %s\n", i, &id)

      message := publisherintegrationevents.Message{
        Id: id.String(),
        Data: strconv.Itoa(i),
      }
      json, err := json.Marshal(message)
      if err != nil {
        fmt.Printf("failed publishing for iteration %d. error: %s\n", i, err)
      } else {
        err := ch.Publish(
          "",
          q.Name,
          true,
          false,
          amqp.Publishing{
              ContentType: "application/json",
              Body:        json,
          },
        )
        if err != nil {
          fmt.Printf("failed publishing for iteration %d. error: %s\n", i, err)
        }
      }
      time.Sleep(2 * time.Second)
    }

  }()
  <-forever

  return nil
}
