package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/infra/database"
	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/repository"
)

type App struct {
	config     AppConfig
	logger     *slog.Logger
	consumer   *Consumer
	httpServer *http.Server
	db         *pgxpool.Pool
}

func New(config AppConfig, logger *slog.Logger, consumer *Consumer, httpServer *http.Server) *App {
	return &App{
		config:     config,
		logger:     logger,
		consumer:   consumer,
		httpServer: httpServer,
	}
}

func (a *App) Start(ctx context.Context) error {

	if a.consumer == nil {
		return fmt.Errorf("failed to start app with nil consumer")
	}

	if a.httpServer == nil {
		return fmt.Errorf("failed to start app with nil http server")
	}

	consumer := *a.consumer
	a.logger.Info("consumer service starting")

	db, err := database.Connect(ctx, a.logger)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	a.db = db
	stopping := false
	go func() {
		for !stopping { // loop until cancel signal
			repo := repository.New(a.db)
			err := consumer.StartConsuming(ctx, repo)

			if err != nil {
				a.logger.Error("failed to start consuming, retrying in 1 sec.", slog.Any("error", err))
			} else {
				a.logger.Info("consumer disconnected. trying connection")
			}
			time.Sleep(1 * time.Second)
		}
	}()

	router, err := a.loadRoutes()

	if err != nil {
		return fmt.Errorf("failed when loading routes: %w", err)
	}

	a.httpServer.Addr = a.config.Addr
	a.httpServer.Handler = router

	errCh := make(chan error, 1)

	go func() {
		err := a.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("failed to listen and serve: %w", err)
		}

		close(errCh)
	}()

	a.logger.Info("server running", slog.String("port", a.config.Addr))
	select {
	// Wait until we receive SIGINT (ctrl+c on cli)
	case <-ctx.Done():
		break
	case err := <-errCh:
		return err
	}

	stopping = true

	a.logger.Info("consumer service stopping")

	return nil
}
