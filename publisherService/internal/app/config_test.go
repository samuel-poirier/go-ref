package app_test

import (
	"testing"

	"github.com/samuel-poirier/go-pubsub-demo/publisher/internal/app"
	"github.com/stretchr/testify/assert"
)

func TestAppConfigValidate(t *testing.T) {
	validConfig := app.AppConfig{
		Hostname:                 "a",
		Addr:                     "a",
		QueueName:                "a",
		RabbitMqConnectionString: "a",
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
		config.RabbitMqConnectionString = ""
		assert.EqualError(t, config.Validate(), "rabbitmq connection string not configured")
	})
	t.Run("Test missing hostname", func(t *testing.T) {
		config := validConfig
		config.Hostname = ""
		assert.EqualError(t, config.Validate(), "app hostname not configured")
	})
	t.Run("Test missing addr", func(t *testing.T) {
		config := validConfig
		config.Addr = ""
		assert.EqualError(t, config.Validate(), "app port not configured")
	})
}
