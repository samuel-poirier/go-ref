package consumer

import (
	"context"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/samuel-poirier/go-ref/consumer/internal/app"
	"github.com/samuel-poirier/go-ref/shared/consumer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type RabbitMqConsumer struct {
	connectionString string
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

	slog.InfoContext(ctx, "consumer listening for messages...")
	processingChannel := *msgChan
	tracer := otel.Tracer("rabbitmq-consumer")

	for d := range msgs {
		// Extract trace context from message headers
		propagator := otel.GetTextMapPropagator()
		carrier := &amqpHeaderCarrier{headers: d.Headers}
		extractedCtx := propagator.Extract(ctx, carrier)

		// Start a new span for message consumption
		spanCtx, span := tracer.Start(extractedCtx, fmt.Sprintf("rabbitmq.consume-%s", queueName),
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(
				attribute.String("messaging.system", "rabbitmq"),
				attribute.String("messaging.destination", queueName),
				attribute.String("messaging.operation", "receive"),
				attribute.Bool("messaging.message.redelivered", d.Redelivered),
			),
		)

		// Convert AMQP headers to map
		headers := make(map[string]interface{})
		for k, v := range d.Headers {
			headers[k] = v
		}

		message := &consumer.Message{
			Data:        d.Body,
			Redelivered: d.Redelivered,
			Headers:     headers,
			Context:     spanCtx, // Pass trace context to handlers
			Ack: func() error {
				err := d.Ack(false)
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, "failed to ack message")
				} else {
					span.SetStatus(codes.Ok, "message processed successfully")
				}
				span.End()
				return err
			},
			Nack: func(requeue bool) error {
				err := d.Nack(false, requeue)
				span.RecordError(fmt.Errorf("message nacked"))
				span.SetStatus(codes.Error, "message processing failed")
				span.End()
				return err
			},
		}

		processingChannel <- *message
	}

	slog.InfoContext(ctx, "consumer stopped")

	return nil
}

func New(config app.AppConfig) consumer.Consumer {
	return &RabbitMqConsumer{
		connectionString: config.RabbitMqConnectionString,
	}
}

// amqpHeaderCarrier adapts AMQP headers to propagation.TextMapCarrier
type amqpHeaderCarrier struct {
	headers amqp.Table
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
