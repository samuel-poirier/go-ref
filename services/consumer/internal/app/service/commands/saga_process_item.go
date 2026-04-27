package commands

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samuel-poirier/go-ref/consumer/internal/repository"
	"github.com/samuel-poirier/go-ref/events"
)

type SagaProcessItemCommand struct {
	SagaID uuid.UUID
	Data   string
}

func (c commands) SagaProcessItem(ctx context.Context, cmd SagaProcessItemCommand) error {
	// Simulate work before the transaction to keep the DB connection short.
	succeeded, processedData, reason := simulateProcessing(cmd.Data)

	_, err := c.repo.TransitionSagaStepStatus(ctx, repository.TransitionSagaStepStatusParams{
		SagaID:        cmd.SagaID,
		StepName:      events.StepProcessItem,
		CurrentStatus: "pending",
		NewStatus:     "processing",
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDuplicateMessage
	}
	if err != nil {
		return err
	}

	var msg repository.CreateOutboxMessageParams
	if succeeded {
		slog.InfoContext(ctx, "process_item activity succeeded", slog.String("saga_id", cmd.SagaID.String()))
		msg, err = repository.NewCreateOutboxMessageParams(events.ItemProcessedSucceededEvent{
			SagaID:        cmd.SagaID.String(),
			ProcessedData: processedData,
		})
	} else {
		slog.WarnContext(ctx, "process_item activity failed",
			slog.String("saga_id", cmd.SagaID.String()),
			slog.String("reason", reason),
		)
		msg, err = repository.NewCreateOutboxMessageParams(events.ItemProcessedFailedEvent{
			SagaID: cmd.SagaID.String(),
			Reason: reason,
		})
	}
	if err != nil {
		return err
	}

	return c.eventOutbox.Write(ctx, msg)
}

func simulateProcessing(data string) (succeeded bool, processedData string, reason string) {
	time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)
	if rand.Float32() > 0.4 {
		return true, "processed:" + data, ""
	}
	return false, "", "simulated failure"
}
