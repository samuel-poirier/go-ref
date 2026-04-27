package commands

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/samuel-poirier/go-ref/consumer/internal/repository"
	"github.com/samuel-poirier/go-ref/shared/outbox"
)

// ErrDuplicateMessage is returned when a message has already been processed.
// Consumers should ACK and skip on this error.
var ErrDuplicateMessage = errors.New("saga: duplicate message, already processed")

type Commands interface {
	CreateProcessedItem(ctx context.Context, cmd CreateProcessedItemCommand) error
	// HTTP trigger
	SagaCreate(ctx context.Context, cmd SagaCreateCommand) (uuid.UUID, error)
	// Orchestrator handlers
	SagaHandleStart(ctx context.Context, cmd SagaHandleStartCommand) error
	SagaHandleSucceeded(ctx context.Context, cmd SagaHandleSucceededCommand) error
	SagaHandleFailed(ctx context.Context, cmd SagaHandleFailedCommand) error
	SagaHandleCompensated(ctx context.Context, cmd SagaHandleCompensatedCommand) error
	// Activity handlers
	SagaProcessItem(ctx context.Context, cmd SagaProcessItemCommand) error
	SagaCompensateProcessItem(ctx context.Context, cmd SagaCompensateProcessItemCommand) error
}

type commands struct {
	repo        repository.Queries
	eventOutbox outbox.Writer[repository.CreateOutboxMessageParams]
}

func New(repo *repository.Queries, eventOutbox outbox.Writer[repository.CreateOutboxMessageParams]) Commands {
	return commands{
		repo:        *repo,
		eventOutbox: eventOutbox,
	}
}
