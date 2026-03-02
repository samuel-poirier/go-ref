package commands

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/samuel-poirier/go-ref/consumer/internal/repository"
	"github.com/samuel-poirier/go-ref/events"
)

type CreateProcessedItemCommand struct {
	Data string `validate:"required"`
}

func (c commands) CreateProcessedItem(ctx context.Context, cmd CreateProcessedItemCommand) (retErr error) {
	v := validator.New()
	err := v.Struct(cmd)

	if err != nil {
		return err
	}

	err = c.repo.CreateProcessedItem(ctx, cmd.Data)

	if err != nil {
		return err
	}

	nextEvent := events.DataProcessedEvent{
		Id:   uuid.New().String(),
		Data: cmd.Data,
	}

	message, err := repository.NewCreateOutboxMessageParams(nextEvent)

	if err != nil {
		return err
	}

	return c.eventOutbox.Write(ctx, message)
}
