package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/samuel-poirier/go-ref/shared/publisher"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type messageWithContext struct {
	ctx     context.Context
	message publisher.MessageEnvelope
}

type RabbitMqPublisher struct {
	connectionString string
	logger           *slog.Logger
	eventChannel     *chan messageWithContext
}

func NewRabbitMqPublisher(connectionString string, logger *slog.Logger) publisher.Publisher {
	eventChannel := make(chan messageWithContext)
	return &RabbitMqPublisher{
		connectionString: connectionString,
		logger:           logger,
		eventChannel:     &eventChannel,
	}
}

func (pub *RabbitMqPublisher) Publish(ctx context.Context, message publisher.MessageEnvelope) error {
	if pub.eventChannel == nil {
		return fmt.Errorf("failed to publish, publishing channel not initialized")
	}
	*pub.eventChannel <- messageWithContext{ctx: ctx, message: message}
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

	messageBuffer := make([]messageWithContext, 0)
	processingChannel := make(chan struct{})

	go func() {
		for msgWithCtx := range *pub.eventChannel {
			messageBuffer = append(messageBuffer, msgWithCtx)
			go func() { processingChannel <- struct{}{} }()
		}
	}()

	defer func() {
		close(processingChannel)
	}()

	go func() {
		tracer := otel.Tracer("rabbitmq-publisher")
		for range processingChannel {
			msgWithCtx := messageBuffer[0]
			messageBuffer = messageBuffer[1:]

			// Start a new span for publishing, using the provided context to link to parent span
			spanCtx, span := tracer.Start(msgWithCtx.ctx, "rabbitmq.publish",
				trace.WithSpanKind(trace.SpanKindProducer),
				trace.WithAttributes(
					attribute.String("messaging.system", "rabbitmq"),
					attribute.String("messaging.destination", msgWithCtx.message.QueueName),
					attribute.String("messaging.operation", "publish"),
				),
			)

			q, err := ch.QueueDeclare(
				msgWithCtx.message.QueueName, // name
				true,              // durable
				false,             // delete when unused
				false,             // exclusive
				false,             // no-wait
				nil,               // arguments
			)

			if err != nil {
				pub.logger.ErrorContext(spanCtx, "failed to declare queue", slog.Any("error", err))
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to declare queue")
				span.End()
				continue
			}

			ch, conn = ensureChannelIsOpen(spanCtx, ch, conn, pub)

			// Inject trace context into message headers
			headers := make(amqp091.Table)
			propagator := otel.GetTextMapPropagator()
			propagator.Inject(spanCtx, &amqpHeaderCarrier{headers: headers})

			err = ch.Publish(
				"",
				q.Name,
				true,
				false,
				amqp091.Publishing{
					ContentType: "application/json",
					Body:        msgWithCtx.message.Message,
					Headers:     headers,
				},
			)
			if err != nil {
				pub.logger.ErrorContext(spanCtx, "failed publishing", slog.Any("error", err))
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to publish message")
			} else {
				span.SetStatus(codes.Ok, "message published successfully")
			}
			span.End()
		}
	}()

	<-ctx.Done()

	return nil
}

func ensureChannelIsOpen(ctx context.Context, ch *amqp091.Channel, conn *amqp091.Connection, publisher *RabbitMqPublisher) (*amqp091.Channel, *amqp091.Connection) {
	var err error
	for ch == nil || ch.IsClosed() {
		conn, err = amqp091.Dial(publisher.connectionString)

		if err != nil {
			publisher.logger.WarnContext(ctx, "failed to re-open closed connection... retrying", slog.Any("error", err))
			time.Sleep(500 * time.Millisecond)
			continue
		}

		ch, err = conn.Channel()

		if err != nil {
			publisher.logger.WarnContext(ctx, "failed to re-open closed channel... retrying", slog.Any("error", err))
			time.Sleep(500 * time.Millisecond)
			continue
		}
	}
	return ch, conn
}

// amqpHeaderCarrier adapts AMQP headers to propagation.TextMapCarrier
type amqpHeaderCarrier struct {
	headers amqp091.Table
}

func (c *amqpHeaderCarrier) Get(key string) string {
	if val, ok := c.headers[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (c *amqpHeaderCarrier) Set(key, value string) {
	c.headers[key] = value
}

func (c *amqpHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c.headers))
	for k := range c.headers {
		keys = append(keys, k)
	}
	return keys
}
