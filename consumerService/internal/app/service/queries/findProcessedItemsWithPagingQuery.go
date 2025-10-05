package queries

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/samuel-poirier/go-ref/consumer/internal/repository"
)

type FindProcessedItemsWithPagingQuery struct {
	Limit  int32 `validate:"gte=0,lte=1000"`
	Offset int32 `validate:"gte=0"`
}

func (h handler) FindProcessedItemsWithPaging(ctx context.Context, q FindProcessedItemsWithPagingQuery) ([]repository.ProcessedItem, error) {

	v := validator.New()
	err := v.Struct(q)

	if err != nil {
		return nil, err
	}

	items, err := h.repo.FindProcessedItemsWithPaging(ctx, repository.FindProcessedItemsWithPagingParams{
		Offset: q.Offset,
		Limit:  q.Limit,
	})

	if err != nil {
		return nil, err
	}

	if items == nil {
		items = make([]repository.ProcessedItem, 0)
	}

	return items, nil
}
