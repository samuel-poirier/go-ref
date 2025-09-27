package queries

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/repository"
)

func (q queries) FindProcessedItemById(ctx context.Context, id uuid.UUID) (*repository.ProcessedItem, error) {
	item, err := q.repo.FindProcessedItemById(ctx, id)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &item, nil
}
