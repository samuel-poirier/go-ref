package app

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Addr              *string
	QueueName         string `yaml:"queueName"`
	ConnectionStrings struct {
		RabbitMq string `yaml:"rabbitMq"`
	} `yaml:"connectionStrings"`
}

func LoadAppConfig(path string) (*AppConfig, error) {
	var config *AppConfig
	filename, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	yamlFile, err := os.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, &config)

	if err != nil {
		return nil, err
	}

	err = config.Validate()

	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *AppConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("nil app config")
	}

	if c.QueueName == "" {
		return fmt.Errorf("queue name not configured")
	}

	if c.ConnectionStrings.RabbitMq == "" {
		return fmt.Errorf("rabbitmq connection string not configured")
	}

	return nil
}
