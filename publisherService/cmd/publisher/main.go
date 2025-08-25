package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/sam9291/go-pubsub-demo/publisher/internal/app"
	"github.com/sam9291/go-pubsub-demo/publisher/internal/infra"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// todo: refactore to use go env module to load from .env
	configPath := os.Getenv("APP_CONFIG")
	logger.Info("loading application config", slog.String("APP_CONFIG", configPath))
	config, err := app.LoadAppConfig(configPath)

	if err != nil {
		logger.Error("failed to load app config", slog.Any("error", err))
		panic(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	publisher := infra.NewRabbitMqPublisher(*config, logger)

	app := app.New(*config, logger, &publisher)

	if err := app.Start(ctx); err != nil {
		logger.Error("failed to start app", slog.Any("error", err))
	}

}
