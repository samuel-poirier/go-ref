package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	_ "github.com/samuel-poirier/go-pubsub-demo/publisher/docs"
	"github.com/samuel-poirier/go-pubsub-demo/publisher/internal/app"
	"github.com/samuel-poirier/go-pubsub-demo/publisher/internal/domain"
	"github.com/samuel-poirier/go-pubsub-demo/publisher/internal/infra"
	"github.com/samuel-poirier/go-pubsub-demo/shared/env"
)

//	@title			Go PubSub Demo Publisher API
//	@version		1.0
//	@description	Example publisher API

//	@contact.url	https://github.com/samuel-poirier/go-pubsub-demo

//	@license.name	MIT
//	@license.url	https://github.com/samuel-poirier/go-pubsub-demo/blob/main/LICENSE

// @host		localhost:8080
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

	publisher := infra.NewRabbitMqPublisher(config.RabbitMqConnectionString, config.QueueName, logger)

	workers := []domain.BackgroundWorker{
		infra.NewPeriodicPublisherBackgroundWorker(2*time.Second, &publisher, logger),
	}

	app := app.New(*config, logger, &publisher, &workers, &http.Server{})

	wg := sync.WaitGroup{}
	wg.Add(1)
	if err := app.Start(ctx, &wg); err != nil {
		logger.Error("failed to start app", slog.Any("error", err))
	}

}
