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
	logger            slog.Logger
	processingChannel <-chan struct{}
	SignalChannel     chan<- struct{}
	outboxReader      OutboxReader[T]
}

func NewReader[T OutboxMessageProvider](logger slog.Logger, outboxReader OutboxReader[T], publisher publisher.Publisher) Reader[T] {
	processingChannel := make(chan struct{})
	return Reader[T]{
		logger:            logger,
		processingChannel: processingChannel,
		SignalChannel:     processingChannel,
		outboxReader:      outboxReader,
		publisher:         publisher,
	}
}

func (r Reader[T]) StartBackgroundReader(ctx context.Context) {
	r.logger.Info("starting outbox reader")
	go r.processMessages(ctx)
	go r.scheduleReads(ctx, 500*time.Millisecond)
}

func (r Reader[T]) processMessages(ctx context.Context) {
	defer r.logger.Info("stopping outbox reader")

	for {
		select {
		case <-r.processingChannel:
			message, err := r.outboxReader.FindFirstOutboxMessageByScheduledTime(ctx)

			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			if err != nil {
				r.logger.Warn("failed to find first message, retrying...", slog.Any("error", err))
				continue
			}

			id := message.GetID()
			payload := message.GetPayload()
			metadata := message.GetMetadata()

			var metadataMap map[string]string

			err = json.Unmarshal(metadata, &metadataMap)

			if err != nil {
				r.logger.Warn("failed to deserialize metadata map. skipping corrupted message", slog.Any("error", err))
				r.outboxReader.DeleteOutboxMessageById(ctx, id)
				go func() {
					r.SignalChannel <- struct{}{}
				}()
				continue
			}

			m := publisher.MessageEnvelope{
				QueueName: metadataMap["queueName"],
				Message:   payload,
				Metadata:  metadata,
			}

			err = r.publisher.Publish(m)

			if err != nil {
				r.logger.Warn("published message, bug failed to delete, retrying...", slog.Any("error", err))
				r.outboxReader.IncrementOutboxMessageTimesAttemptedById(ctx, id)
				continue
			}

			err = r.outboxReader.DeleteOutboxMessageById(ctx, id)

			if err != nil {
				r.logger.Warn("published message, bug failed to delete, retrying...", slog.Any("error", err))
				r.outboxReader.IncrementOutboxMessageTimesAttemptedById(ctx, id)
				continue
			}
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
