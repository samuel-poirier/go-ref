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

type compensatedConsumer struct {
	service *service.Service
	ctx     context.Context
}

func NewCompensatedConsumer(service *service.Service, ctx context.Context) *compensatedConsumer {
	return &compensatedConsumer{service: service, ctx: ctx}
}

func (c compensatedConsumer) GetQueueName() string {
	return "ProcessItemCompensatedEvent"
}

func (c compensatedConsumer) Handle(msg consumer.Message) {
	ctx := msg.Context

	var message events.ProcessItemCompensatedEvent
	if err := json.Unmarshal(msg.Data, &message); err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal ProcessItemCompensatedEvent", slog.Any("error", err))
		msg.Nack(false)
		return
	}

	sagaID, err := uuid.Parse(message.SagaID)
	if err != nil {
		slog.ErrorContext(ctx, "invalid saga_id in ProcessItemCompensatedEvent", slog.Any("error", err))
		msg.Nack(false)
		return
	}

	err = c.service.RunWithUnitOfWork(ctx, func(s service.Service) error {
		return s.Commands.SagaHandleCompensated(ctx, commands.SagaHandleCompensatedCommand{
			SagaID: sagaID,
		})
	})

	if errors.Is(err, commands.ErrDuplicateMessage) {
		msg.Ack()
		return
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to handle process item compensated", slog.Any("error", err))
		msg.Nack(true)
		time.Sleep(200 * time.Millisecond)
		return
	}
	msg.Ack()
}
