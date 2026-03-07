package processed

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/samuel-poirier/go-ref/consumer/internal/app/service"
	"github.com/samuel-poirier/go-ref/consumer/internal/app/service/commands"
	"github.com/samuel-poirier/go-ref/events"
	"github.com/samuel-poirier/go-ref/shared/consumer"
)

type handler struct {
	service *service.Service
	ctx     context.Context
}

func New(service *service.Service, ctx context.Context) *handler {
	return &handler{
		service: service,
		ctx:     ctx,
	}
}

func (c handler) GetQueueName() string {
	return "DataGeneratedEvent"
}
func (c handler) Handle(msg consumer.Message) {
	// Use the message context which contains trace information
	ctx := msg.Context

	var message events.DataGeneratedEvent

	err := json.Unmarshal(msg.Data, &message)
	if err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal json message received from rabbitmq", slog.Any("error", err))
		err = msg.Nack(false)
		if err != nil {
			slog.ErrorContext(ctx, "failed to nack message", slog.Any("error", err))
		}
		return
	}

	slog.InfoContext(ctx, "received a message", slog.String("id", message.Id), slog.String("data", message.Data))

	cmd := commands.CreateProcessedItemCommand{
		Data: message.Data,
	}

	err = c.service.RunWithUnitOfWork(ctx, func(s service.Service) error {
		return s.Commands.CreateProcessedItem(ctx, cmd)
	})

	if err != nil {
		slog.ErrorContext(ctx, "failed to persist processed item", slog.Any("error", err))
		err = msg.Nack(true)
		time.Sleep(200 * time.Millisecond)
		if err != nil {
			slog.ErrorContext(ctx, "failed to nack message", slog.Any("error", err))
		}
	} else {
		err = msg.Ack()
		if err != nil {
			slog.ErrorContext(ctx, "failed to ack message", slog.Any("error", err))
		}
	}
}
