package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type App struct {
	config   AppConfig
	logger   *slog.Logger
	consumer *Consumer
}

func New(config AppConfig, logger *slog.Logger, consumer *Consumer) *App {
	return &App{
		config:   config,
		logger:   logger,
		consumer: consumer,
	}
}

func (a *App) Start(ctx context.Context) error {

	if a.consumer == nil {
		return fmt.Errorf("failed to start app with nil consumer")
	}

	consumer := *a.consumer
	a.logger.Info("publisher service starting")

	stopping := false
	go func() {
		for !stopping { // loop until cancel signal
			err := consumer.StartConsuming(ctx)

			if err != nil {
				a.logger.Error("failed to start consuming, retrying in 1 sec.", slog.Any("error", err))
			} else {
				a.logger.Info("consumer disconnected. trying connection")
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Wait until we receive SIGINT (ctrl+c on cli)
	<-ctx.Done()
	stopping = true

	a.logger.Info("consumer service stopping")

	return nil
}
