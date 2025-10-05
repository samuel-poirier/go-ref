package consumer

import (
	"context"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/samuel-poirier/go-ref/consumer/internal/app"
	"github.com/samuel-poirier/go-ref/shared/consumer"
)

type RabbitMqConsumer struct {
	connectionString string
	logger           *slog.Logger
}

func (c *RabbitMqConsumer) Subscribe(queueName string, msgChan *chan<- consumer.Message, ctx context.Context) error {

	if msgChan == nil {
		return fmt.Errorf("unexpected nil message channel")
	}

	conn, err := amqp.Dial(c.connectionString)

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
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		return err
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	c.logger.Info("consumer listening for messages...")
	processingChannel := *msgChan
	for d := range msgs {

		message := &consumer.Message{
			Data:        d.Body,
			Redelivered: d.Redelivered,
			Ack:         func() error { return d.Ack(false) },
			Nack:        func(requeue bool) error { return d.Nack(false, requeue) },
		}

		processingChannel <- *message
	}

	c.logger.Info("consumer stopped")

	return nil
}

func New(config app.AppConfig, logger *slog.Logger) consumer.Consumer {
	return &RabbitMqConsumer{
		connectionString: config.RabbitMqConnectionString,
		logger:           logger,
	}
}
