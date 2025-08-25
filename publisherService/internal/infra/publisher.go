package infra

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	events "github.com/sam9291/go-pubsub-demo/events"
	"github.com/sam9291/go-pubsub-demo/publisher/internal/app"
)

type RabbitMqPublisher struct {
	connectionString string
	queueName        string
	logger           *slog.Logger
}

func NewRabbitMqPublisher(config app.AppConfig, logger *slog.Logger) app.Publisher {
	return &RabbitMqPublisher{
		connectionString: config.ConnectionStrings.RabbitMq,
		queueName:        config.QueueName,
		logger:           logger,
	}
}

func (publisher *RabbitMqPublisher) StartPublishing(ctx context.Context) error {

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
		publisher.queueName, // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)

	if err != nil {
		return err
	}

	stopping := false

	go func() {
		<-ctx.Done()
		stopping = true
	}()

	for i := 0; !stopping; i++ {

		id := uuid.New()

		publisher.logger.Info("publishing message", slog.Int("iteration", i), slog.String("id", id.String()))

		message := events.Message{
			Id:   id.String(),
			Data: strconv.Itoa(i),
		}

		json, err := json.Marshal(message)

		if err != nil {
			publisher.logger.Error("failed publishing", slog.Int("iteration", i), slog.Any("error", err))
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
				publisher.logger.Error("failed publishing", slog.Int("iteration", i), slog.Any("error", err))
				stopping = true
			}
		}
		time.Sleep(2 * time.Second)
	}

	return nil
}
