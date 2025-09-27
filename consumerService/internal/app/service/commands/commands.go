package commands

import "github.com/samuel-poirier/go-pubsub-demo/consumer/internal/repository"

type commands struct {
	repo repository.Queries
}

func New(repo *repository.Queries) commands {
	return commands{
		repo: *repo,
	}
}
