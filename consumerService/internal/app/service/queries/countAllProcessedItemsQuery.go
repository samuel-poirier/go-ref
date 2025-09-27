package queries

import "context"

func (q queries) CountAllProcessedItems(ctx context.Context) (int64, error) {
	return q.repo.CountAllProcessedItems(ctx)
}
