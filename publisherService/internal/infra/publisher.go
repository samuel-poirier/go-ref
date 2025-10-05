package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	events "github.com/samuel-poirier/go-ref/events"
	"github.com/samuel-poirier/go-ref/publisher/internal/domain"
)

type RabbitMqPublisher struct {
	connectionString string
	queueName        string
	logger           *slog.Logger
	eventChannel     *chan events.Message
}

func NewRabbitMqPublisher(connectionString, queueName string, logger *slog.Logger) domain.Publisher {
	eventChannel := make(chan events.Message)
	return &RabbitMqPublisher{
		connectionString: connectionString,
		queueName:        queueName,
		logger:           logger,
		eventChannel:     &eventChannel,
	}
}

func (publisher *RabbitMqPublisher) Publish(message events.Message) error {
	if publisher.eventChannel == nil {
		return fmt.Errorf("failed to publish, publishing channel not initialized")
	}
	*publisher.eventChannel <- message
	return nil
}

func (publisher *RabbitMqPublisher) Close() {
	if publisher.eventChannel != nil {
		close(*publisher.eventChannel)
	}
}
func (publisher *RabbitMqPublisher) Initialize(ctx context.Context) error {
	if publisher.eventChannel == nil {
		return fmt.Errorf("failed to initialize publisher with nil publishing channel")
	}

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
	processingChannel := make(chan struct{})

	go func() {
		for message := range *publisher.eventChannel {
			messageBuffer = append(messageBuffer, message)
			go func() { processingChannel <- struct{}{} }()
		}
	}()

	defer func() {
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
