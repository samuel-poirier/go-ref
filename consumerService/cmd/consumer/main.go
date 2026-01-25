package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	_ "github.com/samuel-poirier/go-ref/consumer/docs"
	"github.com/samuel-poirier/go-ref/consumer/internal/app"
	"github.com/samuel-poirier/go-ref/consumer/internal/infra/rabbitmq/consumer"
	"github.com/samuel-poirier/go-ref/shared/env"
	"github.com/samuel-poirier/go-ref/shared/publisher/rabbitmq"
	"github.com/samuel-poirier/go-ref/shared/telemetry"
)

//	@title			Go PubSub Demo Consumer API
//	@version		1.0
//	@description	Example consumer API

//	@contact.url	https://github.com/samuel-poirier/go-ref

//	@license.name	MIT
//	@license.url	https://github.com/samuel-poirier/go-ref/blob/main/LICENSE

// @host		localhost:8081
// @BasePath	/
func main() {
	configPath := env.GetEnvOrDefault("APP_CONFIG", "../../configs/.env")

	config, err := app.LoadAppConfig(configPath)
	if err != nil {
		// Use temporary logger before telemetry is initialized
		tempLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		tempLogger.Error("failed to load app config", slog.Any("error", err))
		panic(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Initialize OpenTelemetry (this also creates the logger with log bridge)
	otelConfig := telemetry.NewConfig()
	otelProvider, logger, err := telemetry.InitProvider(ctx, otelConfig)
	if err != nil {
		// Use temporary logger if telemetry init fails
		tempLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		tempLogger.Error("failed to initialize telemetry", slog.Any("error", err))
		panic(1)
	}
	defer func() {
		if err := otelProvider.Shutdown(context.Background()); err != nil {
			logger.Error("failed to shutdown telemetry", slog.Any("error", err))
		}
	}()

	consumer := consumer.New(*config, logger)

	publisher := rabbitmq.NewRabbitMqPublisher(config.RabbitMqConnectionString, logger)

	app := app.New(*config, logger, &consumer, &publisher, &http.Server{})

	if err := app.Start(ctx); err != nil {
		logger.Error("failed to start app", slog.Any("error", err))
	}

}
