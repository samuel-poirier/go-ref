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

type SagaHandleSucceededCommand struct {
	SagaID        uuid.UUID
	ProcessedData string
}

func (c commands) SagaHandleSucceeded(ctx context.Context, cmd SagaHandleSucceededCommand) error {
	_, err := c.repo.TransitionSagaStepStatus(ctx, repository.TransitionSagaStepStatusParams{
		SagaID:        cmd.SagaID,
		StepName:      events.StepProcessItem,
		CurrentStatus: "processing",
		NewStatus:     "succeeded",
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDuplicateMessage
	}
	if err != nil {
		return err
	}

	_, err = c.repo.TransitionSagaInstanceStatus(ctx, repository.TransitionSagaInstanceStatusParams{
		ID:            cmd.SagaID,
		CurrentStatus: "processing",
		NewStatus:     "processed-succeeded",
	})
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "saga completed successfully", slog.String("saga_id", cmd.SagaID.String()))
	return nil
}
