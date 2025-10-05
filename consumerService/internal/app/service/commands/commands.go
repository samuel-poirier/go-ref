package commands

import (
	"context"

	"github.com/samuel-poirier/go-ref/consumer/internal/repository"
)

type Commands interface {
	CreateProcessedItem(ctx context.Context, cmd CreateProcessedItemCommand) error
}

type commands struct {
	repo repository.Queries
}

func New(repo *repository.Queries) Commands {
	return commands{
		repo: *repo,
	}
}
