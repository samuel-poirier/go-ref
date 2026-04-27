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

type processItemConsumer struct {
	service *service.Service
	ctx     context.Context
}

func NewProcessItemConsumer(service *service.Service, ctx context.Context) *processItemConsumer {
	return &processItemConsumer{service: service, ctx: ctx}
}

func (c processItemConsumer) GetQueueName() string {
	return "ProcessItemCommand"
}

func (c processItemConsumer) Handle(msg consumer.Message) {
	ctx := msg.Context

	var message events.ProcessItemCommand
	if err := json.Unmarshal(msg.Data, &message); err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal ProcessItemCommand", slog.Any("error", err))
		msg.Nack(false)
		return
	}

	sagaID, err := uuid.Parse(message.SagaID)
	if err != nil {
		slog.ErrorContext(ctx, "invalid saga_id in ProcessItemCommand", slog.Any("error", err))
		msg.Nack(false)
		return
	}

	err = c.service.RunWithUnitOfWork(ctx, func(s service.Service) error {
		return s.Commands.SagaProcessItem(ctx, commands.SagaProcessItemCommand{
			SagaID: sagaID,
			Data:   message.Data,
		})
	})

	if errors.Is(err, commands.ErrDuplicateMessage) {
		msg.Ack()
		return
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to process item activity", slog.Any("error", err))
		msg.Nack(true)
		time.Sleep(200 * time.Millisecond)
		return
	}
	msg.Ack()
}
