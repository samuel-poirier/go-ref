package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samuel-poirier/go-ref/consumer/internal/app/service/commands"
	"github.com/samuel-poirier/go-ref/consumer/internal/app/service/queries"
	"github.com/samuel-poirier/go-ref/consumer/internal/repository"
	"github.com/samuel-poirier/go-ref/shared/outbox"
)

type Service struct {
	queries.Queries
	commands.Commands
	EventOutbox         outbox.Writer[repository.CreateOutboxMessageParams]
	repo                *repository.Queries
	db                  *pgxpool.Pool
	readerSignalChannel chan<- struct{}
}

func New(repo *repository.Queries, db *pgxpool.Pool, readerSignalChannel chan<- struct{}) *Service {
	eventOutbox := outbox.NewWriter(readerSignalChannel, repo)
	return &Service{
		Queries:             queries.New(repo),
		Commands:            commands.New(repo, eventOutbox),
		EventOutbox:         eventOutbox,
		repo:                repo,
		db:                  db,
		readerSignalChannel: readerSignalChannel,
	}
}

func (s Service) RunWithUnitOfWork(ctx context.Context, uow func(Service) error) error {

	tx, err := s.db.Begin(ctx)

	if err != nil {
		return err
	}

	txRepo := s.repo.WithTx(tx)

	service := New(txRepo, s.db, s.readerSignalChannel)

	err = uow(*service)

	if err != nil {
		err2 := tx.Rollback(ctx)
		if err2 != nil {
			err = errors.Join(err, err2)
		}
		return err
	}

	err = tx.Commit(ctx)

	if err == nil {
		go func() {
			s.readerSignalChannel <- struct{}{}
		}()
	}

	return err

}
