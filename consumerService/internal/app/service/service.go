package service

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samuel-poirier/go-ref/consumer/internal/app/service/commands"
	"github.com/samuel-poirier/go-ref/consumer/internal/app/service/queries"
	"github.com/samuel-poirier/go-ref/consumer/internal/repository"
)

type Service struct {
	queries.Queries
	commands.Commands
}

func New(repo *repository.Queries, db *pgxpool.Pool) *Service {
	return &Service{
		Queries:  queries.New(repo),
		Commands: commands.New(repo, db),
	}
}
