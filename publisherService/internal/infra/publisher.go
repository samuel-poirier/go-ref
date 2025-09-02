package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

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

	defer func() {
		if conn != nil && !conn.IsClosed() {
			conn.Close()
		}
	}()

	ch, err := conn.Channel()

	if err != nil {
		return err
	}

	defer func() {
		if ch != nil && !ch.IsClosed() {
			ch.Close()
		}
	}()

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

	messageBuffer := make([]events.Message, 0)
	eventChannel := make(chan events.Message)
	processingChannel := make(chan struct{})
	var publishingChannel chan<- events.Message = eventChannel
	publisher.publishingChan = &publishingChannel

	go func() {
		for message := range eventChannel {
			messageBuffer = append(messageBuffer, message)
			go func() { processingChannel <- struct{}{} }()
		}
	}()

	defer func() {
		close(eventChannel)
		close(processingChannel)
	}()

	go func() {
		for range processingChannel {
			message := messageBuffer[0]
			messageBuffer = messageBuffer[1:]
			json, err := json.Marshal(message)

			if err != nil {
				publisher.logger.Error("failed publishing", slog.Any("error", err))
			} else {
				ch, conn = ensureChannelIsOpen(ch, conn, publisher)
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
				}
			}
		}
	}()

	<-ctx.Done()

	return nil
}

func ensureChannelIsOpen(ch *amqp.Channel, conn *amqp.Connection, publisher *RabbitMqPublisher) (*amqp.Channel, *amqp.Connection) {
	var err error
	for ch == nil || ch.IsClosed() {
		conn, err = amqp.Dial(publisher.connectionString)

		if err != nil {
			publisher.logger.Warn("failed to re-open closed connection... retrying", slog.Any("error", err))
			time.Sleep(500 * time.Millisecond)
			continue
		}

		ch, err = conn.Channel()

		if err != nil {
			publisher.logger.Warn("failed to re-open closed channel... retrying", slog.Any("error", err))
			time.Sleep(500 * time.Millisecond)
			continue
		}
	}
	return ch, conn
}
