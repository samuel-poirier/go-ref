package queries

import (
	"context"

	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/repository"
)

type Queries interface {
	FindProcessedItemsWithPaging(ctx context.Context, q FindProcessedItemsWithPagingQuery) ([]repository.ProcessedItem, error)
	FindProcessedItemById(ctx context.Context, q FindProcessedItemByIdQuery) (*repository.ProcessedItem, error)
	CountAllProcessedItems(ctx context.Context, q CountAllProcessedItemsQuery) (int64, error)
}

type handler struct {
	repo repository.Queries
}

func New(repo *repository.Queries) Queries {
	return handler{
		repo: *repo,
	}
}
