package outbox

import "context"

type OutboxPersister[T any] interface {
	CreateOutboxMessage(ctx context.Context, arg T) error
}

type Writer[T any] struct {
	readerSignalChannel chan<- struct{}
	persister           OutboxPersister[T]
}

func NewWriter[T any](readerSignalChannel chan<- struct{}, persister OutboxPersister[T]) Writer[T] {
	return Writer[T]{
		readerSignalChannel: readerSignalChannel,
		persister:           persister,
	}
}

func (w Writer[T]) Write(ctx context.Context, message T) error {
	err := w.persister.CreateOutboxMessage(ctx, message)

	if err != nil {
		return err
	}

	go func() {
		w.readerSignalChannel <- struct{}{}
	}()

	return nil
}
