package commands

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samuel-poirier/go-ref/consumer/internal/repository"
	"github.com/samuel-poirier/go-ref/events"
)

type SagaCompensateProcessItemCommand struct {
	SagaID uuid.UUID
}

func (c commands) SagaCompensateProcessItem(ctx context.Context, cmd SagaCompensateProcessItemCommand) error {
	// Simulate compensation work before the transaction.
	time.Sleep(500 * time.Millisecond)

	_, err := c.repo.TransitionSagaStepStatus(ctx, repository.TransitionSagaStepStatusParams{
		SagaID:        cmd.SagaID,
		StepName:      events.StepCompensateProcessItem,
		CurrentStatus: "pending",
		NewStatus:     "compensating",
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDuplicateMessage
	}
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "compensate_process_item activity executed",
		slog.String("saga_id", cmd.SagaID.String()),
	)

	msg, err := repository.NewCreateOutboxMessageParams(events.ProcessItemCompensatedEvent{
		SagaID: cmd.SagaID.String(),
	})
	if err != nil {
		return err
	}

	return c.eventOutbox.Write(ctx, msg)
}
