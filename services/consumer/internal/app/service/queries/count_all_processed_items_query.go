package queries

import "context"

type CountAllProcessedItemsQuery struct{}

func (h handler) CountAllProcessedItems(ctx context.Context, q CountAllProcessedItemsQuery) (int64, error) {
	return h.repo.CountAllProcessedItems(ctx)
}
