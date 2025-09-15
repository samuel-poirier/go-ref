package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/sam9291/go-pubsub-demo/publisher/internal/domain"
)

type App struct {
	config            AppConfig
	logger            *slog.Logger
	publisher         *domain.Publisher
	backgroundWorkers *[]domain.BackgroundWorker
	httpServer        *http.Server
}

func New(config AppConfig, logger *slog.Logger, publisher *domain.Publisher, workers *[]domain.BackgroundWorker, httpServer *http.Server) *App {
	return &App{
		config:            config,
		logger:            logger,
		publisher:         publisher,
		backgroundWorkers: workers,
		httpServer:        httpServer,
	}
}

func (a *App) Start(ctx context.Context, wg *sync.WaitGroup) error {

	if a.publisher == nil {
		return fmt.Errorf("failed to start app with nil publisher")
	}

	if wg == nil {
		return fmt.Errorf("failed to start app with nil wait group")
	}

	if a.httpServer == nil {
		return fmt.Errorf("failed to start app with nil http server")
	}

	publisher := *a.publisher
	a.logger.Info("publisher service starting")

	stopping := false
	go func() {
		defer publisher.Close()
		for !stopping { // loop until cancel signal
			err := publisher.Initialize(ctx)
			if err != nil {
				a.logger.Warn("failed to start publishing, retrying to reconnect in 1 sec", slog.Any("error", err))
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

	for i, worker := range *a.backgroundWorkers {
		go func(index int, w domain.BackgroundWorker) {
			a.logger.Info("starting background worker", slog.Int("index", index))
			err := w.Start(ctx)
			if err != nil {
				a.logger.Error("error returned from background worker", slog.Int("index", index), slog.Any("error", err))
				errCh <- fmt.Errorf("background worker failed with error: %w", err)
			}
		}(i, worker)
	}

	go func() {
		err := a.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("failed to listen and serve: %w", err)
		}

		close(errCh)
	}()

	a.logger.Info("server running", slog.String("port", a.config.Addr))

	wg.Done()

	select {
	// Wait until we receive SIGINT (ctrl+c on cli)
	case <-ctx.Done():
		break
	case err := <-errCh:
		return err
	}
	stopping = true

	a.logger.Info("publisher service stopping")

	return nil
}
