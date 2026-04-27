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

type SagaHandleFailedCommand struct {
	SagaID uuid.UUID
	Reason string
}

func (c commands) SagaHandleFailed(ctx context.Context, cmd SagaHandleFailedCommand) error {
	_, err := c.repo.TransitionSagaStepStatus(ctx, repository.TransitionSagaStepStatusParams{
		SagaID:        cmd.SagaID,
		StepName:      events.StepProcessItem,
		CurrentStatus: "processing",
		NewStatus:     "failed",
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDuplicateMessage
	}
	if err != nil {
		return err
	}

	if err = c.repo.CreateSagaStep(ctx, repository.CreateSagaStepParams{
		SagaID:   cmd.SagaID,
		StepName: events.StepCompensateProcessItem,
		Status:   "pending",
	}); err != nil {
		return err
	}

	msg, err := repository.NewCreateOutboxMessageParams(events.CompensateProcessItemCommand{
		SagaID: cmd.SagaID.String(),
	})
	if err != nil {
		return err
	}

	slog.WarnContext(ctx, "saga step failed, dispatching compensation",
		slog.String("saga_id", cmd.SagaID.String()),
		slog.String("reason", cmd.Reason),
	)
	return c.eventOutbox.Write(ctx, msg)
}
