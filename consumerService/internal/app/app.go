package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samuel-poirier/go-ref/consumer/internal/app/consumers/processed"
	"github.com/samuel-poirier/go-ref/consumer/internal/app/service"
	"github.com/samuel-poirier/go-ref/consumer/internal/infra/database"
	"github.com/samuel-poirier/go-ref/consumer/internal/repository"
	"github.com/samuel-poirier/go-ref/shared/consumer"
)

type App struct {
	config     AppConfig
	logger     *slog.Logger
	consumer   *consumer.Consumer
	httpServer *http.Server
	db         *pgxpool.Pool
}

func New(config AppConfig, logger *slog.Logger, consumer *consumer.Consumer, httpServer *http.Server) *App {
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

	msgConsumer := *a.consumer
	a.logger.Info("consumer service starting")

	db, err := database.Connect(ctx, a.logger)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	a.db = db

	repo := repository.New(a.db)
	service := service.New(repo, a.db)
	consumerHandlers := make([]consumer.ConsumerHandler, 0)
	consumerHandlers = append(consumerHandlers, processed.New(service, *a.logger, ctx))

	errCh := make(chan error, 1)
	stopping := false

	for _, handler := range consumerHandlers {
		go func() {
			registerConsumer(ctx, stopping, handler, msgConsumer, a)
		}()
	}

	router, err := a.loadRoutes(service)

	if err != nil {
		return fmt.Errorf("failed when loading routes: %w", err)
	}

	a.httpServer.Addr = a.config.Addr
	a.httpServer.Handler = router

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

func registerConsumer(ctx context.Context, stopping bool, handler consumer.ConsumerHandler, msgConsumer consumer.Consumer, a *App) {
	msgChan := make(chan consumer.Message)
	defer close(msgChan)

	var subscribeMsgChan chan<- consumer.Message = msgChan

	go func(h consumer.ConsumerHandler, c <-chan consumer.Message) {
		for message := range c {
			if stopping {
				message.Nack(true)
				return
			}
			h.Handle(message)
		}
	}(handler, msgChan)

	a.logger.Info("registering consumer", slog.String("queue", a.config.QueueName), slog.String("handler", fmt.Sprintf("%T", handler)))

	for {
		err := msgConsumer.Subscribe(a.config.QueueName, &subscribeMsgChan, ctx)

		if err != nil {
			a.logger.Warn("failed to consumer, retrying...", slog.String("queue", a.config.QueueName), slog.String("handler", fmt.Sprintf("%T", handler)))
			time.Sleep(500 * time.Millisecond)
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

	}
}
