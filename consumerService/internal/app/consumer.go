package app

import (
	"context"

	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/repository"
)

type Consumer interface {
	StartConsuming(ctx context.Context, repo *repository.Queries) error
}
