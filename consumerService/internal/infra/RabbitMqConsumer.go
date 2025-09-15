package infra

import (
	"context"
	"encoding/json"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sam9291/go-pubsub-demo/consumer/internal/app"
	events "github.com/sam9291/go-pubsub-demo/events"
)

type RabbitMqConsumer struct {
	connectionString string
	queueName        string
	logger           *slog.Logger
}

func (consumer *RabbitMqConsumer) StartConsuming(ctx context.Context) error {

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
		consumer.queueName, // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
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

	consumer.logger.Info("consumer listening for messages...")
	for d := range msgs {

		var message events.Message

		err := json.Unmarshal(d.Body, &message)
		if err != nil {
			consumer.logger.Error("failed to unmarshal json message received from rabbitmq", slog.Any("error", err))
		} else {
			consumer.logger.Info("Received a message", slog.String("id", message.Id), slog.String("data", message.Data))
		}
	}

	consumer.logger.Info("consumer stopped")

	return nil
}

func NewRabbitMqConsumer(config app.AppConfig, logger *slog.Logger) app.Consumer {
	return &RabbitMqConsumer{
		connectionString: config.RabbitMqConnectionString,
		queueName:        config.QueueName,
		logger:           logger,
	}
}
