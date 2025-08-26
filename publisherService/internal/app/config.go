package app

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	QueueName         string `yaml:"queueName"`
	ConnectionStrings struct {
		RabbitMq string `yaml:"rabbitMq"`
	} `yaml:"connectionStrings"`
}

func LoadAppConfig(path string) (*AppConfig, error) {
	var config AppConfig
	filename, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	yamlFile, err := os.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(yamlFile, &config)

	if err != nil {
		return nil, err
	}

	if config.QueueName == "" {
		return nil, fmt.Errorf("queue name not configured")
	}

	if config.ConnectionStrings.RabbitMq == "" {
		return nil, fmt.Errorf("rabbitmq connection string not configured")
	}

	return &config, nil
}
