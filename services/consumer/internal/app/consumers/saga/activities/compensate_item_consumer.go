package activities

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

type compensateItemConsumer struct {
	service *service.Service
	ctx     context.Context
}

func NewCompensateItemConsumer(service *service.Service, ctx context.Context) *compensateItemConsumer {
	return &compensateItemConsumer{service: service, ctx: ctx}
}

func (c compensateItemConsumer) GetQueueName() string {
	return "CompensateProcessItemCommand"
}

func (c compensateItemConsumer) Handle(msg consumer.Message) {
	ctx := msg.Context

	var message events.CompensateProcessItemCommand
	if err := json.Unmarshal(msg.Data, &message); err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal CompensateProcessItemCommand", slog.Any("error", err))
		msg.Nack(false)
		return
	}

	sagaID, err := uuid.Parse(message.SagaID)
	if err != nil {
		slog.ErrorContext(ctx, "invalid saga_id in CompensateProcessItemCommand", slog.Any("error", err))
		msg.Nack(false)
		return
	}

	err = c.service.RunWithUnitOfWork(ctx, func(s service.Service) error {
		return s.Commands.SagaCompensateProcessItem(ctx, commands.SagaCompensateProcessItemCommand{
			SagaID: sagaID,
		})
	})

	if errors.Is(err, commands.ErrDuplicateMessage) {
		msg.Ack()
		return
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to compensate process item activity", slog.Any("error", err))
		msg.Nack(true)
		time.Sleep(200 * time.Millisecond)
		return
	}
	msg.Ack()
}
