package queries

import (
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
