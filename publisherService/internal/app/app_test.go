package app_test

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sam9291/go-pubsub-demo/events"
	"github.com/sam9291/go-pubsub-demo/publisher/internal/app"
	"github.com/sam9291/go-pubsub-demo/publisher/internal/domain"
	"github.com/sam9291/go-pubsub-demo/publisher/internal/infra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func TestAppIntegrationTests(t *testing.T) {
	ctx := t.Context()

	rabbitmqContainer, err := rabbitmq.Run(ctx, "rabbitmq:4.1.2-management-alpine", rabbitmq.WithAdminUsername("guest"), rabbitmq.WithAdminPassword("guest"))
	defer func() {
		if err := testcontainers.TerminateContainer(rabbitmqContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	if err != nil {
		t.Errorf("failed to start container: %s", err)
		return
	}
	rabbitmqUrl, err := rabbitmqContainer.AmqpURL(t.Context())

	if err != nil {
		t.Errorf("failed to get amqp url container: %s", err)
		return
	}

	queueName := "queue-name"
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	publisher := infra.NewRabbitMqPublisher(rabbitmqUrl, queueName, logger)

	conn, err := amqp091.Dial(rabbitmqUrl)
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	t.Run("Application starts and stops gracefully", func(t *testing.T) {
		ts := httptest.NewServer(nil)
		defer ts.Close()
		u, err := url.Parse(ts.URL)
		if err == nil {
			assert.NoError(t, err)
			return
		}

		_, port, err := net.SplitHostPort(u.Host)
		if err == nil {
			assert.NoError(t, err)
			return
		}
		addr := ":" + port
		config := app.AppConfig{
			Addr:                     addr,
			QueueName:                "queue-name",
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
		ts := httptest.NewServer(nil)
		defer ts.Close()
		u, err := url.Parse(ts.URL)
		if err == nil {
			assert.NoError(t, err)
			return
		}

		_, port, err := net.SplitHostPort(u.Host)
		if err == nil {
			assert.NoError(t, err)
			return
		}
		addr := ":" + port
		config := app.AppConfig{
			Addr:                     addr,
			QueueName:                "queue-name",
			RabbitMqConnectionString: rabbitmqUrl,
		}
		app := app.New(config, logger, &publisher, new([]domain.BackgroundWorker), ts.Config)
		wg := sync.WaitGroup{}
		wg.Add(1)
		ctx, cancel := context.WithCancel(context.Background())
		errChan := make(chan error)

		defer func() {
			cancel()
			err = <-errChan
			assert.NoError(t, err)
		}()

		go func() {
			errChan <- app.Start(ctx, &wg)
		}()

		wg.Wait()

		resp, err := http.Get(ts.URL + "/")

		if err != nil {
			assert.NoError(t, err)
			return
		}

		assert.Equal(t, resp.StatusCode, 200)

		bytedata, err := io.ReadAll(resp.Body)

		if err != nil {
			assert.NoError(t, err)
			return
		}

		reqBodyString := string(bytedata)
		assert.Equal(t, reqBodyString, "hello world")

		msgs, err := ch.Consume(config.QueueName, "", true, false, false, false, nil)
		require.NoError(t, err)
		consumedMessage := <-msgs

		var message events.Message

		err = json.Unmarshal(consumedMessage.Body, &message)
		if err != nil {
			assert.NoError(t, err, "failed to unmarshal message")
		}
		assert.Equal(t, message.Data, "PUBLISHED FROM HELLO WORLD ENDPOINT")
	})
}
