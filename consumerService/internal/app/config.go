package app

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/samuel-poirier/go-ref/shared/env"
)

type AppConfig struct {
	Hostname                 string
	Addr                     string
	QueueName                string
	RabbitMqConnectionString string
}

func LoadAppConfig(path string) (*AppConfig, error) {

	_, err := os.Stat(path)

	if err == nil {
		err := godotenv.Load(path)
		if err != nil {
			return nil, err
		}
	}

	config := AppConfig{
		Hostname:                 env.GetEnvOrDefault("APP_HOSTNAME", "localhost"),
		Addr:                     fmt.Sprintf(":%s", env.GetEnvOrDefault("APP_PORT", "8081")),
		QueueName:                env.GetEnvOrDefault("QUEUE_NAME", "demo-queue"),
		RabbitMqConnectionString: env.GetEnvOrDefault("RABBIT_MQ_CONNECTION_STRING", ""),
	}

	err = config.Validate()

	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *AppConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("nil app config")
	}

	if c.Hostname == "" {
		return fmt.Errorf("app hostname not configured")
	}

	if c.Addr == "" {
		return fmt.Errorf("app port not configured")
	}

	if c.QueueName == "" {
		return fmt.Errorf("queue name not configured")
	}

	if c.RabbitMqConnectionString == "" {
		return fmt.Errorf("rabbitmq connection string not configured")
	}

	return nil
}
