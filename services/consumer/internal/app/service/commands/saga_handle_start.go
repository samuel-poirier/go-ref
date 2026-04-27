package commands

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samuel-poirier/go-ref/consumer/internal/repository"
	"github.com/samuel-poirier/go-ref/events"
)

type SagaHandleStartCommand struct {
	SagaID uuid.UUID
	Data   string
}

func (c commands) SagaHandleStart(ctx context.Context, cmd SagaHandleStartCommand) error {
	_, err := c.repo.TransitionSagaInstanceStatus(ctx, repository.TransitionSagaInstanceStatusParams{
		ID:            cmd.SagaID,
		CurrentStatus: "pending",
		NewStatus:     "processing",
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDuplicateMessage
	}
	if err != nil {
		return err
	}

	if err = c.repo.CreateSagaStep(ctx, repository.CreateSagaStepParams{
		SagaID:   cmd.SagaID,
		StepName: events.StepProcessItem,
		Status:   "pending",
	}); err != nil {
		return err
	}

	msg, err := repository.NewCreateOutboxMessageParams(events.ProcessItemCommand{
		SagaID: cmd.SagaID.String(),
		Data:   cmd.Data,
	})
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "saga started processing", slog.String("saga_id", cmd.SagaID.String()))
	return c.eventOutbox.Write(ctx, msg)
}
