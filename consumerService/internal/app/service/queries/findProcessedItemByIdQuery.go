package queries

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/repository"
)

type FindProcessedItemByIdQuery struct {
	Id uuid.UUID `validate:"required"`
}

func (h handler) FindProcessedItemById(ctx context.Context, q FindProcessedItemByIdQuery) (*repository.ProcessedItem, error) {

	v := validator.New()
	err := v.Struct(q)

	if err != nil {
		return nil, err
	}

	item, err := h.repo.FindProcessedItemById(ctx, q.Id)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &item, nil
}
