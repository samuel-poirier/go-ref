package commands

import (
	"context"

	"github.com/samuel-poirier/go-ref/consumer/internal/repository"
	"github.com/samuel-poirier/go-ref/shared/outbox"
)

type Commands interface {
	CreateProcessedItem(ctx context.Context, cmd CreateProcessedItemCommand) error
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
