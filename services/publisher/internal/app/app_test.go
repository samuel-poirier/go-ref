package app_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"github.com/samuel-poirier/go-ref/events"
	"github.com/samuel-poirier/go-ref/publisher/internal/app"
	"github.com/samuel-poirier/go-ref/publisher/internal/domain"
	rabbitPublisher "github.com/samuel-poirier/go-ref/shared/publisher/rabbitmq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func TestAppIntegrationTests(t *testing.T) {
	ctx := t.Context()

	rabbitmqContainer, err := rabbitmq.Run(ctx, "rabbitmq:4.1.2-management-alpine", rabbitmq.WithAdminUsername("guest"), rabbitmq.WithAdminPassword("guest"))

	defer testcontainers.TerminateContainer(rabbitmqContainer)

	if err != nil {
		t.Errorf("failed to start container: %s", err)
		return
	}
	rabbitmqUrl, err := rabbitmqContainer.AmqpURL(t.Context())

	if err != nil {
		t.Errorf("failed to get amqp url container: %s", err)
		return
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	conn, err := amqp091.Dial(rabbitmqUrl)
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	t.Run("Application starts and stops gracefully", func(t *testing.T) {
		publisher := rabbitPublisher.NewRabbitMqPublisher(rabbitmqUrl, logger)
		ts := httptest.NewServer(nil)
		defer ts.Close()
		addr := ":0" // Assigns a random free port
		config := app.AppConfig{
			Addr:                     addr,
			RabbitMqConnectionString: rabbitmqUrl,
		}
		app := app.New(config, logger, &publisher, new([]domain.BackgroundWorker), ts.Config)
		wg := sync.WaitGroup{}
		wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())

		errChan := make(chan error)
		go func() {
			errChan <- app.Start(ctx, &wg)
		}()
		wg.Wait()
		cancel()
		err = <-errChan
		assert.NoError(t, err)
	})

	t.Run("GET / publishes message to rabbit", func(t *testing.T) {
		publisher := rabbitPublisher.NewRabbitMqPublisher(rabbitmqUrl, logger)
		ts := httptest.NewServer(nil)
		defer ts.Close()

		addr := ":0" // Assigns a random free port
		config := app.AppConfig{
			Addr:                     addr,
			RabbitMqConnectionString: rabbitmqUrl,
			Hostname:                 "localhost",
		}
		app := app.New(config, logger, &publisher, new([]domain.BackgroundWorker), ts.Config)
		wg := sync.WaitGroup{}
		wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		errChan := make(chan error)

		queueName := "DataGeneratedEvent"
		q, err := ch.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)

		defer func() {
			cancel()
			err = <-errChan
			assert.NoError(t, err)
		}()

		assert.NoError(t, err)
		if err != nil {
			return
		}

		go func() {
			errChan <- app.Start(ctx, &wg)
		}()

		wg.Wait()

		resp, err := http.Get(ts.URL + "/api/v1/hello")

		assert.NoError(t, err)
		if err != nil {
			return
		}

		assert.Equal(t, resp.StatusCode, 200)

		bytedata, err := io.ReadAll(resp.Body)

		assert.NoError(t, err)
		if err != nil {
			return
		}

		var unmarshaledResp events.DataGeneratedEvent

		err = json.Unmarshal(bytedata, &unmarshaledResp)
		assert.NoError(t, err)

		assert.Equal(t, "PUBLISHED FROM HELLO WORLD ENDPOINT", unmarshaledResp.Data)

		assert.NoError(t, err)
		msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)

		require.NoError(t, err)
		consumedMessage := <-msgs

		var message events.DataGeneratedEvent

		err = json.Unmarshal(consumedMessage.Body, &message)
		if err != nil {
			assert.NoError(t, err, "failed to unmarshal message")
		}
		assert.Equal(t, message.Data, "PUBLISHED FROM HELLO WORLD ENDPOINT")
	})
}
