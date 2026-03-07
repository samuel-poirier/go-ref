package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/samuel-poirier/go-ref/shared/publisher"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type OutboxMessageProvider interface {
	GetID() uuid.UUID
	GetMetadata() []byte
	GetPayload() []byte
}

type OutboxReader[T OutboxMessageProvider] interface {
	FindFirstOutboxMessageByScheduledTime(ctx context.Context) (T, error)
	IncrementOutboxMessageTimesAttemptedById(ctx context.Context, id uuid.UUID) error
	DeleteOutboxMessageById(ctx context.Context, id uuid.UUID) error
}

type Reader[T OutboxMessageProvider] struct {
	publisher         publisher.Publisher
	processingChannel <-chan struct{}
	SignalChannel     chan<- struct{}
	outboxReader      OutboxReader[T]
}

func NewReader[T OutboxMessageProvider](outboxReader OutboxReader[T], publisher publisher.Publisher) Reader[T] {
	processingChannel := make(chan struct{})
	return Reader[T]{
		processingChannel: processingChannel,
		SignalChannel:     processingChannel,
		outboxReader:      outboxReader,
		publisher:         publisher,
	}
}

func (r Reader[T]) StartBackgroundReader(ctx context.Context) {
	slog.InfoContext(ctx, "starting outbox reader")
	go r.processMessages(ctx)
	go r.scheduleReads(ctx, 500*time.Millisecond)
}

func (r Reader[T]) processMessages(ctx context.Context) {
	defer slog.InfoContext(ctx, "stopping outbox reader")
	tracer := otel.Tracer("outbox-reader")

	for {
		select {
		case <-r.processingChannel:
			ctx, span := tracer.Start(ctx, "outbox.process")
			span.SetAttributes(attribute.String("outbox.operation", "read_and_publish"))

			message, err := r.outboxReader.FindFirstOutboxMessageByScheduledTime(ctx)

			if errors.Is(err, sql.ErrNoRows) {
				span.SetStatus(codes.Ok, "no messages to process")
				span.End()
				continue
			}

			if err != nil {
				slog.WarnContext(ctx, "failed to find first message, retrying...", slog.Any("error", err))
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to find outbox message")
				span.End()
				continue
			}

			id := message.GetID()
			payload := message.GetPayload()
			metadata := message.GetMetadata()

			span.SetAttributes(
				attribute.String("outbox.message_id", id.String()),
			)

			var metadataMap map[string]string

			err = json.Unmarshal(metadata, &metadataMap)

			if err != nil {
				slog.WarnContext(ctx, "failed to deserialize metadata map. skipping corrupted message", slog.Any("error", err))
				r.outboxReader.DeleteOutboxMessageById(ctx, id)
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to deserialize metadata")
				span.End()
				go func() {
					r.SignalChannel <- struct{}{}
				}()
				continue
			}

			span.SetAttributes(
				attribute.String("outbox.queue_name", metadataMap["queueName"]),
			)

			m := publisher.MessageEnvelope{
				QueueName: metadataMap["queueName"],
				Message:   payload,
				Metadata:  metadata,
			}

			// Pass context to link outbox publish to the outbox processing span
			err = r.publisher.Publish(ctx, m)

			if err != nil {
				slog.WarnContext(ctx, "published message, bug failed to delete, retrying...", slog.Any("error", err))
				r.outboxReader.IncrementOutboxMessageTimesAttemptedById(ctx, id)
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to publish message")
				span.End()
				continue
			}

			err = r.outboxReader.DeleteOutboxMessageById(ctx, id)

			if err != nil {
				slog.WarnContext(ctx, "published message, bug failed to delete, retrying...", slog.Any("error", err))
				r.outboxReader.IncrementOutboxMessageTimesAttemptedById(ctx, id)
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to delete outbox message")
				span.End()
				continue
			}

			span.SetStatus(codes.Ok, "outbox message processed successfully")
			span.End()

			go func() {
				r.SignalChannel <- struct{}{}
			}()
		case <-ctx.Done():
			close(r.SignalChannel)
			return
		}
	}
}

func (r Reader[T]) scheduleReads(ctx context.Context, interval time.Duration) {
	for {
		go func() {
			r.SignalChannel <- struct{}{}
		}()
		time.Sleep(interval)

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}
