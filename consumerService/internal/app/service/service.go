package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/app/service/commands"
	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/app/service/queries"
	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/repository"
)

type Queries interface {
	FindProcessedItemsWithPaging(ctx context.Context, offset, limit int32) ([]repository.ProcessedItem, error)
	FindProcessedItemById(ctx context.Context, id uuid.UUID) (*repository.ProcessedItem, error)
	CountAllProcessedItems(ctx context.Context) (int64, error)
}

type Commands interface {
}

type Service struct {
	Queries  Queries
	Commands Commands
}

func New(repo *repository.Queries) *Service {
	return &Service{
		Queries:  queries.New(repo),
		Commands: commands.New(repo),
	}
}
