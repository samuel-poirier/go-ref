package infra

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sam9291/go-pubsub-demo/consumer/internal/app"
	publisherintegrationevents "github.com/sam9291/go-pubsub-demo/publisherIntegrationEvents"
)

type RabbitMqConsumer struct {
  connectionString string
}

func (consumer *RabbitMqConsumer) StartConsuming() error {

  conn, err := amqp.Dial(consumer.connectionString)

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
  msgs, err := ch.Consume(
    q.Name, // queue
    "",     // consumer
    true,   // auto-ack
    false,  // exclusive
    false,  // no-local
    false,  // no-wait
    nil,    // args
  )
  if err != nil {
    return err
  }
  consumeStopped := make(chan bool)

  go func() {
    fmt.Println("consumer listening for messages...")
    for d := range msgs {

      var message publisherintegrationevents.Message

      err := json.Unmarshal(d.Body, &message)
      if err != nil {
        fmt.Println("failed to unmarshal json message received from rabbitmq")
      } else {
        fmt.Printf("Received a message with id %s and data %s\n", message.Id, message.Data)
      }
    }

    fmt.Println("consumer stopped")
    consumeStopped<-true
  }()

  <-consumeStopped

  return nil
}

func NewRabbitMqConsumer(config app.AppConfig) app.Consumer {
	return &RabbitMqConsumer{
    connectionString: config.ConnectionStrings.RabbitMq,
  }
}
