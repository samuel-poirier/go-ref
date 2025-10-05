package processed

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/samuel-poirier/go-ref/consumer/internal/app/service"
	"github.com/samuel-poirier/go-ref/consumer/internal/app/service/commands"
	"github.com/samuel-poirier/go-ref/events"
	"github.com/samuel-poirier/go-ref/shared/consumer"
)

type handler struct {
	logger  slog.Logger
	service *service.Service
	ctx     context.Context
}

func New(service *service.Service, logger slog.Logger, ctx context.Context) *handler {
	return &handler{
		service: service,
		logger:  logger,
		ctx:     ctx,
	}
}

func (c handler) Handle(msg consumer.Message) {

	var message events.Message

	err := json.Unmarshal(msg.Data, &message)
	if err != nil {
		c.logger.Error("failed to unmarshal json message received from rabbitmq", slog.Any("error", err))
		err = msg.Nack(false)
		if err != nil {
			c.logger.Error("failed to nack message", slog.Any("error", err))
		}
		return
	}

	c.logger.Info("Received a message", slog.String("id", message.Id), slog.String("data", message.Data))
	cmd := commands.CreateProcessedItemCommand{
		Data: message.Data,
	}
	err = c.service.Commands.CreateProcessedItem(c.ctx, cmd)

	if err != nil {
		c.logger.Error("failed to persist processed item", slog.Any("error", err))
		err = msg.Nack(true)
		if err != nil {
			c.logger.Error("failed to nack message", slog.Any("error", err))
		}
	} else {
		err = msg.Ack()
		if err != nil {
			c.logger.Error("failed to ack message", slog.Any("error", err))
		}
	}
}
