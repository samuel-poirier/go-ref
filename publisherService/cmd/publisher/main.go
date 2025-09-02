package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/sam9291/go-pubsub-demo/publisher/internal/app"
	"github.com/sam9291/go-pubsub-demo/publisher/internal/domain"
	"github.com/sam9291/go-pubsub-demo/publisher/internal/infra"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	configPath := os.Getenv("APP_CONFIG")
	config, err := app.LoadAppConfig(configPath)

	if err != nil {
		logger.Error("failed to load app config", slog.Any("error", err))
		panic(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	publisher := infra.NewRabbitMqPublisher(config.ConnectionStrings.RabbitMq, config.QueueName, logger)

	workers := []domain.BackgroundWorker{
		infra.NewPeriodicPublisherBackgroundWorker(2*time.Second, &publisher, logger),
	}

	app := app.New(*config, logger, &publisher, &workers)

	if err := app.Start(ctx); err != nil {
		logger.Error("failed to start app", slog.Any("error", err))
	}

}
