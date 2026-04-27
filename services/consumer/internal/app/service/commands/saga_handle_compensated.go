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

type SagaHandleCompensatedCommand struct {
	SagaID uuid.UUID
}

func (c commands) SagaHandleCompensated(ctx context.Context, cmd SagaHandleCompensatedCommand) error {
	_, err := c.repo.TransitionSagaStepStatus(ctx, repository.TransitionSagaStepStatusParams{
		SagaID:        cmd.SagaID,
		StepName:      events.StepCompensateProcessItem,
		CurrentStatus: "compensating",
		NewStatus:     "compensated",
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
		NewStatus:     "processed-failed",
	})
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "saga completed with failure, compensation applied",
		slog.String("saga_id", cmd.SagaID.String()),
	)
	return nil
}
