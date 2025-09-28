package service

import (
	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/app/service/commands"
	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/app/service/queries"
	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/repository"
)

type Commands interface {
}

type Service struct {
	Queries  queries.Queries
	Commands Commands
}

func New(repo *repository.Queries) *Service {
	return &Service{
		Queries:  queries.New(repo),
		Commands: commands.New(repo),
	}
}
