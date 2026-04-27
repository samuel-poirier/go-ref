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

type succeededConsumer struct {
	service *service.Service
	ctx     context.Context
}

func NewSucceededConsumer(service *service.Service, ctx context.Context) *succeededConsumer {
	return &succeededConsumer{service: service, ctx: ctx}
}

func (c succeededConsumer) GetQueueName() string {
	return "ItemProcessedSucceededEvent"
}

func (c succeededConsumer) Handle(msg consumer.Message) {
	ctx := msg.Context

	var message events.ItemProcessedSucceededEvent
	if err := json.Unmarshal(msg.Data, &message); err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal ItemProcessedSucceededEvent", slog.Any("error", err))
		msg.Nack(false)
		return
	}

	sagaID, err := uuid.Parse(message.SagaID)
	if err != nil {
		slog.ErrorContext(ctx, "invalid saga_id in ItemProcessedSucceededEvent", slog.Any("error", err))
		msg.Nack(false)
		return
	}

	err = c.service.RunWithUnitOfWork(ctx, func(s service.Service) error {
		return s.Commands.SagaHandleSucceeded(ctx, commands.SagaHandleSucceededCommand{
			SagaID:        sagaID,
			ProcessedData: message.ProcessedData,
		})
	})

	if errors.Is(err, commands.ErrDuplicateMessage) {
		msg.Ack()
		return
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to handle item processed succeeded", slog.Any("error", err))
		msg.Nack(true)
		time.Sleep(200 * time.Millisecond)
		return
	}
	msg.Ack()
}
