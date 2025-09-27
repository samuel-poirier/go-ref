package queries

import (
	"context"

	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/repository"
)

type queries struct {
	repo repository.Queries
}

func New(repo *repository.Queries) queries {
	return queries{
		repo: *repo,
	}
}

func (q queries) CountAllProcessedItems(ctx context.Context) (int64, error) {
	return q.repo.CountAllProcessedItems(ctx)
}

func (q queries) FindProcessedItemsWithPaging(ctx context.Context, offset, limit int32) ([]repository.ProcessedItem, error) {
	return q.repo.FindProcessedItemsWithPaging(ctx, repository.FindProcessedItemsWithPagingParams{
		Offset: offset,
		Limit:  limit,
	})
}
