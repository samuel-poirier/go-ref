package app_test

import (
	"testing"

	"github.com/sam9291/go-pubsub-demo/publisher/internal/app"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalAndValidateConfig(t *testing.T) {
	validConfig := app.AppConfig{
		QueueName: "a",
		ConnectionStrings: struct {
			RabbitMq string "yaml:\"rabbitMq\""
		}{
			RabbitMq: "b",
		},
	}
	t.Run("Test valid", func(t *testing.T) {
		config := validConfig
		assert.NoError(t, config.Validate())
	})
	t.Run("Test nil config", func(t *testing.T) {
		var config *app.AppConfig = nil
		assert.EqualError(t, config.Validate(), "nil app config")
	})
	t.Run("Test missing queue name", func(t *testing.T) {
		config := validConfig
		config.QueueName = ""
		assert.EqualError(t, config.Validate(), "queue name not configured")
	})
	t.Run("Test missing connection string", func(t *testing.T) {
		config := validConfig
		config.ConnectionStrings.RabbitMq = ""
		assert.EqualError(t, config.Validate(), "rabbitmq connection string not configured")
	})
}
