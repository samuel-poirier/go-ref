package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	events "github.com/sam9291/go-pubsub-demo/events"
	"github.com/sam9291/go-pubsub-demo/publisher/internal/domain"
)

type RabbitMqPublisher struct {
	connectionString string
	queueName        string
	logger           *slog.Logger
	publishingChan   *chan<- events.Message
}

func NewRabbitMqPublisher(connectionString, queueName string, logger *slog.Logger) domain.Publisher {
	return &RabbitMqPublisher{
		connectionString: connectionString,
		queueName:        queueName,
		logger:           logger,
	}
}

func (publisher *RabbitMqPublisher) Publish(message events.Message) error {
	if publisher.publishingChan == nil {
		return fmt.Errorf("failed to publish, publishing channel not initialized")
	}
	*publisher.publishingChan <- message
	return nil
}

func (publisher *RabbitMqPublisher) Initialize(ctx context.Context) error {

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

	eventChannel := make(chan events.Message)
	var publishingChannel chan<- events.Message = eventChannel
	var receivingChannel <-chan events.Message = eventChannel
	publisher.publishingChan = &publishingChannel

	defer func() {
		close(eventChannel)
	}()

	go func() {
		for event := range receivingChannel {

			json, err := json.Marshal(event)

			if err != nil {
				publisher.logger.Error("failed publishing", slog.Any("error", err))
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
					publisher.logger.Error("failed publishing", slog.Any("error", err))
					stopping = true
				}
			}
		}
	}()

	for i := 0; !stopping; i++ {

		id := uuid.New()

		publisher.logger.Info("publishing message", slog.Int("iteration", i), slog.String("id", id.String()))

		message := events.Message{
			Id:   id.String(),
			Data: strconv.Itoa(i),
		}

		publisher.Publish(message)

		time.Sleep(2 * time.Second)
	}

	return nil
}
