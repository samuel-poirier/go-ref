package commands

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samuel-poirier/go-ref/consumer/internal/repository"
)

type Commands interface {
	CreateProcessedItem(ctx context.Context, cmd CreateProcessedItemCommand) error
}

type commands struct {
	repo repository.Queries
	db   *pgxpool.Pool
}

func New(repo *repository.Queries, db *pgxpool.Pool) Commands {
	return commands{
		repo: *repo,
		db:   db,
	}
}
