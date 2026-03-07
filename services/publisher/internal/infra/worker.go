package infra

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	events "github.com/samuel-poirier/go-ref/events"
	"github.com/samuel-poirier/go-ref/publisher/internal/domain"
	"github.com/samuel-poirier/go-ref/shared/publisher"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type PeriodicPublisherMessageBackgroundWorker struct {
	sleepDuration time.Duration
	publisher     *publisher.Publisher
}

func NewPeriodicPublisherBackgroundWorker(time time.Duration, publisher *publisher.Publisher) domain.BackgroundWorker {
	return &PeriodicPublisherMessageBackgroundWorker{
		sleepDuration: time,
		publisher:     publisher,
	}
}

func (w *PeriodicPublisherMessageBackgroundWorker) Start(ctx context.Context) error {

	if w.publisher == nil {
		return fmt.Errorf("cannot start with nil publisher")
	}
	pub := *w.publisher

	slog.InfoContext(ctx, "starting periodic publisher background worker")
	defer func() {
		slog.InfoContext(ctx, "stopping periodic publisher background worker")
	}()

	tracer := otel.Tracer("background-worker")

	for i := 0; ; i++ {
		// Create a span for each worker cycle
		cycleCtx, span := tracer.Start(ctx, "worker.publish_cycle",
			trace.WithAttributes(attribute.Int("worker.iteration", i)),
		)

		id := uuid.New()

		slog.InfoContext(cycleCtx, "publishing message", slog.Int("iteration", i), slog.String("id", id.String()))

		message := events.DataGeneratedEvent{
			Id:   id.String(),
			Data: strconv.Itoa(i),
		}

		m, err := publisher.NewMessageEnvelope(message)

		if err != nil {
			slog.ErrorContext(cycleCtx, "error while publishing message", slog.Int("iteration", i), slog.String("id", id.String()), slog.Any("error", err))
		} else {
			// Pass the cycle context to link this publish to the worker cycle span
			err = pub.Publish(cycleCtx, m)

			if err != nil {
				slog.ErrorContext(cycleCtx, "error while publishing message", slog.Int("iteration", i), slog.String("id", id.String()), slog.Any("error", err))
			}
		}

		span.End()

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		time.Sleep(w.sleepDuration)

	}
}
