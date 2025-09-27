package queries

import (
	"context"

	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/repository"
)

func (q queries) FindProcessedItemsWithPaging(ctx context.Context, offset, limit int32) ([]repository.ProcessedItem, error) {
	return q.repo.FindProcessedItemsWithPaging(ctx, repository.FindProcessedItemsWithPagingParams{
		Offset: offset,
		Limit:  limit,
	})
}
