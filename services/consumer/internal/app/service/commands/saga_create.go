package commands

import (
	"context"
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/samuel-poirier/go-ref/consumer/internal/repository"
	"github.com/samuel-poirier/go-ref/events"
)

type SagaCreateCommand struct {
	Data string `validate:"required"`
}

func (c commands) SagaCreate(ctx context.Context, cmd SagaCreateCommand) (uuid.UUID, error) {
	v := validator.New()
	if err := v.Struct(cmd); err != nil {
		return uuid.Nil, err
	}

	sagaID := uuid.New()

	payload, err := json.Marshal(map[string]string{"data": cmd.Data})
	if err != nil {
		return uuid.Nil, err
	}

	_, err = c.repo.CreateSagaInstance(ctx, repository.CreateSagaInstanceParams{
		ID:      sagaID,
		Status:  "pending",
		Payload: payload,
	})
	if err != nil {
		return uuid.Nil, err
	}

	msg, err := repository.NewCreateOutboxMessageParams(events.StartItemSagaCommand{
		SagaID: sagaID.String(),
		Data:   cmd.Data,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return sagaID, c.eventOutbox.Write(ctx, msg)
}
