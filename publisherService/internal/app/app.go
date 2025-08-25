package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type App struct {
	config    AppConfig
	logger    *slog.Logger
	publisher *Publisher
}

func New(config AppConfig, logger *slog.Logger, publisher *Publisher) *App {
	return &App{
		config:    config,
		logger:    logger,
		publisher: publisher,
	}
}

func (a *App) Start(ctx context.Context) error {

	if a.publisher == nil {
		return fmt.Errorf("failed to start app with nil publisher")
	}

	publisher := *a.publisher
	a.logger.Info("publisher service starting")

	stopping := false
	go func() {
		for !stopping { // loop until cancel signal
			err := publisher.StartPublishing(ctx)
			if err != nil {
				a.logger.Warn("failed to start publishing, retrying to reconnect in 1 sec", slog.Any("error", err))
			}

			time.Sleep(1 * time.Second)
		}
	}()

	// Wait until we receive SIGINT (ctrl+c on cli)
	<-ctx.Done()
	stopping = true

	a.logger.Info("publisher service stopping")

	return nil
}
