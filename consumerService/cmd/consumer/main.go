package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	_ "github.com/sam9291/go-pubsub-demo/consumer/docs"
	"github.com/sam9291/go-pubsub-demo/consumer/internal/app"
	"github.com/sam9291/go-pubsub-demo/consumer/internal/infra"
	"github.com/sam9291/go-pubsub-demo/shared/env"
)

//	@title			Go PubSub Demo Consumer API
//	@version		1.0
//	@description	Example consumer API

//	@contact.url	https://github.com/sam9291/go-pubsub-demo

//	@license.name	MIT
//	@license.url	https://github.com/sam9291/go-pubsub-demo/blob/main/LICENSE

// @host		localhost:8081
// @BasePath	/
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	configPath := env.GetEnvOrDefault("APP_CONFIG", "../../configs/.env")

	config, err := app.LoadAppConfig(configPath)
	if err != nil {
		logger.Error("failed to load app config", slog.Any("error", err))
		panic(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	consumer := infra.NewRabbitMqConsumer(*config, logger)

	app := app.New(*config, logger, &consumer, &http.Server{})

	if err := app.Start(ctx); err != nil {
		logger.Error("failed to start app", slog.Any("error", err))
	}

}
