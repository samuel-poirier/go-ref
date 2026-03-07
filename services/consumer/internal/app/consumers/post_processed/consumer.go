package postprocessed

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/samuel-poirier/go-ref/consumer/internal/app/service"
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
	return "DataProcessedEvent"
}

func (c handler) Handle(msg consumer.Message) {
	// Use the message context which contains trace information
	ctx := msg.Context

	var message events.DataProcessedEvent

	err := json.Unmarshal(msg.Data, &message)
	if err != nil {
		slog.ErrorContext(ctx, "failed to unmarshal json message received from rabbitmq", slog.Any("error", err))
		err = msg.Nack(false)
		if err != nil {
			slog.ErrorContext(ctx, "failed to nack message", slog.Any("error", err))
		}
		return
	}

	slog.InfoContext(ctx, "handled post processed event", slog.String("id", message.Id), slog.String("data", message.Data))

	msg.Ack()
}
