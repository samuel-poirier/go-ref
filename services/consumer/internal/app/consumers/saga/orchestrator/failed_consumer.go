package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/samuel-poirier/go-ref/consumer/internal/app/service"
	"github.com/samuel-poirier/go-ref/consumer/internal/app/service/commands"
	"github.com/samuel-poirier/go-ref/events"
	"github.com/samuel-poirier/go-ref/shared/consumer"
)

type failedConsumer struct {
	service *service.Service
	ctx     context.Context
}

func NewFailedConsumer(service *service.Service, ctx context.Context) *failedConsumer {
	return &failedConsumer{service: service, ctx: ctx}
}

func (c failedConsumer) GetQueueName() string {
	return "ItemProcessedFailedEvent"
}

func (c failedConsumer) Handle(msg consumer.Message) {
	ctx := msg.Context

	var message events.ItemProcessedFailedEvent
	if err := json.Unmarshal(msg.Data, &message); err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal ItemProcessedFailedEvent", slog.Any("error", err))
		msg.Nack(false)
		return
	}

	sagaID, err := uuid.Parse(message.SagaID)
	if err != nil {
		slog.ErrorContext(ctx, "invalid saga_id in ItemProcessedFailedEvent", slog.Any("error", err))
		msg.Nack(false)
		return
	}

	err = c.service.RunWithUnitOfWork(ctx, func(s service.Service) error {
		return s.Commands.SagaHandleFailed(ctx, commands.SagaHandleFailedCommand{
			SagaID: sagaID,
			Reason: message.Reason,
		})
	})

	if errors.Is(err, commands.ErrDuplicateMessage) {
		msg.Ack()
		return
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to handle item processed failed", slog.Any("error", err))
		msg.Nack(true)
		time.Sleep(200 * time.Millisecond)
		return
	}
	msg.Ack()
}
