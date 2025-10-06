package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/samuel-poirier/go-ref/shared/publisher"
)

type RabbitMqPublisher struct {
	connectionString string
	logger           *slog.Logger
	eventChannel     *chan publisher.MessageEnvelope
}

func NewRabbitMqPublisher(connectionString string, logger *slog.Logger) publisher.Publisher {
	eventChannel := make(chan publisher.MessageEnvelope)
	return &RabbitMqPublisher{
		connectionString: connectionString,
		logger:           logger,
		eventChannel:     &eventChannel,
	}
}

func (publisher *RabbitMqPublisher) Publish(message publisher.MessageEnvelope) error {
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
func (pub *RabbitMqPublisher) Initialize(ctx context.Context) error {
	if pub.eventChannel == nil {
		return fmt.Errorf("failed to initialize publisher with nil publishing channel")
	}

	conn, err := amqp091.Dial(pub.connectionString)

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

	messageBuffer := make([]publisher.MessageEnvelope, 0)
	processingChannel := make(chan struct{})

	go func() {
		for message := range *pub.eventChannel {
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
			q, err := ch.QueueDeclare(
				message.QueueName, // name
				true,              // durable
				false,             // delete when unused
				false,             // exclusive
				false,             // no-wait
				nil,               // arguments
			)

			if err != nil {
				pub.logger.Error("failed to declare queue", slog.Any("error", err))
				continue
			}

			ch, conn = ensureChannelIsOpen(ch, conn, pub)
			err = ch.Publish(
				"",
				q.Name,
				true,
				false,
				amqp091.Publishing{
					ContentType: "application/json",
					Body:        message.Message,
				},
			)
			if err != nil {
				pub.logger.Error("failed publishing", slog.Any("error", err))
			}
		}
	}()

	<-ctx.Done()

	return nil
}

func ensureChannelIsOpen(ch *amqp091.Channel, conn *amqp091.Connection, publisher *RabbitMqPublisher) (*amqp091.Channel, *amqp091.Connection) {
	var err error
	for ch == nil || ch.IsClosed() {
		conn, err = amqp091.Dial(publisher.connectionString)

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
