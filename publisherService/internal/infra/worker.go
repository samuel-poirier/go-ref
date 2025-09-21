package infra

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	events "github.com/samuel-poirier/go-pubsub-demo/events"
	"github.com/samuel-poirier/go-pubsub-demo/publisher/internal/domain"
)

type PeriodicPublisherMessageBackgroundWorker struct {
	sleepDuration time.Duration
	publisher     *domain.Publisher
	logger        *slog.Logger
}

func NewPeriodicPublisherBackgroundWorker(time time.Duration, publisher *domain.Publisher, logger *slog.Logger) domain.BackgroundWorker {
	return &PeriodicPublisherMessageBackgroundWorker{
		sleepDuration: time,
		publisher:     publisher,
		logger:        logger,
	}
}

func (w *PeriodicPublisherMessageBackgroundWorker) Start(context.Context) error {

	if w.publisher == nil {
		return fmt.Errorf("cannot start with nil publisher")
	}
	publisher := *w.publisher

	if w.logger == nil {
		return fmt.Errorf("cannot start with nil logger")
	}
	logger := *w.logger

	logger.Info("starting periodic publisher background worker")
	defer func() {
		logger.Info("stopping periodic publisher background worker")
	}()

	for i := 0; ; i++ {
		id := uuid.New()

		logger.Info("publishing message", slog.Int("iteration", i), slog.String("id", id.String()))

		message := events.Message{
			Id:   id.String(),
			Data: strconv.Itoa(i),
		}

		err := publisher.Publish(message)

		if err != nil {
			logger.Error("error while publishing message", slog.Int("iteration", i), slog.String("id", id.String()), slog.Any("error", err))
		}

		select {
		case <-context.Background().Done():
			return nil
		default:
		}

		time.Sleep(w.sleepDuration)

	}
}
