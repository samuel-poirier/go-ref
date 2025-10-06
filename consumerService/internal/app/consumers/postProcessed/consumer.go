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

func (c handler) GetQueueName() string {
	return "DataProcessedEvent"
}

func (c handler) Handle(msg consumer.Message) {

	var message events.DataProcessedEvent

	err := json.Unmarshal(msg.Data, &message)
	if err != nil {
		c.logger.Error("failed to unmarshal json message received from rabbitmq", slog.Any("error", err))
		err = msg.Nack(false)
		if err != nil {
			c.logger.Error("failed to nack message", slog.Any("error", err))
		}
		return
	}

	c.logger.Info("handled post processed event", slog.String("id", message.Id), slog.String("data", message.Data))

	msg.Ack()
}
