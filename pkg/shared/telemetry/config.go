package telemetry

import (
	"os"
)

// Config holds the OpenTelemetry configuration
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	EnableStdout   bool
}

// NewConfig creates a new telemetry configuration from environment variables
func NewConfig() *Config {
	return &Config{
		ServiceName:    getEnvOrDefault("OTEL_SERVICE_NAME", "unknown-service"),
		ServiceVersion: getEnvOrDefault("OTEL_SERVICE_VERSION", "1.0.0"),
		Environment:    getEnvOrDefault("OTEL_RESOURCE_ATTRIBUTES", "deployment.environment=dev"),
		OTLPEndpoint:   getEnvOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		EnableStdout:   getEnvOrDefault("OTEL_ENABLE_STDOUT", "false") == "true",
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
