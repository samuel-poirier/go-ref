package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samuel-poirier/go-ref/consumer/internal/app/service/commands"
	"github.com/samuel-poirier/go-ref/consumer/internal/app/service/queries"
	"github.com/samuel-poirier/go-ref/consumer/internal/repository"
)

type Service struct {
	queries.Queries
	commands.Commands
	repo *repository.Queries
	db   *pgxpool.Pool
}

func New(repo *repository.Queries, db *pgxpool.Pool) *Service {
	return &Service{
		Queries:  queries.New(repo),
		Commands: commands.New(repo),
		repo:     repo,
		db:       db,
	}
}

func (s Service) RunWithUnitOfWork(ctx context.Context, uow func(Service) error) error {

	tx, err := s.db.Begin(ctx)

	if err != nil {
		return err
	}

	txRepo := s.repo.WithTx(tx)

	service := Service{
		Queries:  queries.New(txRepo),
		Commands: commands.New(txRepo),
	}

	err = uow(service)

	if err != nil {
		err2 := tx.Rollback(ctx)
		if err2 != nil {
			err = errors.Join(err, err2)
		}
		return err
	}

	return tx.Commit(ctx)

}
