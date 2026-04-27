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

type startConsumer struct {
	service *service.Service
	ctx     context.Context
}

func NewStartConsumer(service *service.Service, ctx context.Context) *startConsumer {
	return &startConsumer{service: service, ctx: ctx}
}

func (c startConsumer) GetQueueName() string {
	return "StartItemSagaCommand"
}

func (c startConsumer) Handle(msg consumer.Message) {
	ctx := msg.Context

	var message events.StartItemSagaCommand
	if err := json.Unmarshal(msg.Data, &message); err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal StartItemSagaCommand", slog.Any("error", err))
		msg.Nack(false)
		return
	}

	sagaID, err := uuid.Parse(message.SagaID)
	if err != nil {
		slog.ErrorContext(ctx, "invalid saga_id in StartItemSagaCommand", slog.Any("error", err))
		msg.Nack(false)
		return
	}

	err = c.service.RunWithUnitOfWork(ctx, func(s service.Service) error {
		return s.Commands.SagaHandleStart(ctx, commands.SagaHandleStartCommand{
			SagaID: sagaID,
			Data:   message.Data,
		})
	})

	if errors.Is(err, commands.ErrDuplicateMessage) {
		msg.Ack()
		return
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to handle saga start", slog.Any("error", err))
		msg.Nack(true)
		time.Sleep(200 * time.Millisecond)
		return
	}
	msg.Ack()
}
