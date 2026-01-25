package outbox

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

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
	tracer := otel.Tracer("outbox-writer")
	ctx, span := tracer.Start(ctx, "outbox.write")
	span.SetAttributes(attribute.String("outbox.operation", "write"))
	defer span.End()

	err := w.persister.CreateOutboxMessage(ctx, message)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to write outbox message")
		return err
	}

	span.SetStatus(codes.Ok, "outbox message written successfully")

	go func() {
		w.readerSignalChannel <- struct{}{}
	}()

	return nil
}
