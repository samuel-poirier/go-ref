package telemetry

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Provider holds the telemetry providers
type Provider struct {
	tracerProvider *sdktrace.TracerProvider
	loggerProvider *sdklog.LoggerProvider
	logger         *slog.Logger
}

// InitProvider initializes the OpenTelemetry providers and returns both the provider and a configured logger
func InitProvider(ctx context.Context, config *Config) (*Provider, *slog.Logger, error) {
	// Create resource with service information (shared by tracer and logger)
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
		),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, nil, err
	}

	// Initialize tracer provider
	tracerProvider, err := NewTracerProviderWithResource(ctx, config, res)
	if err != nil {
		return nil, nil, err
	}

	// Initialize logger provider
	loggerProvider, err := NewLoggerProvider(ctx, config, res)
	if err != nil {
		return nil, nil, err
	}

	// Create the bridged logger
	logger := NewLogger(loggerProvider)

	// Set global tracer provider
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator for context propagation (W3C Trace Context)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	logger.Info("OpenTelemetry initialized",
		slog.String("service", config.ServiceName),
		slog.String("version", config.ServiceVersion),
		slog.String("endpoint", config.OTLPEndpoint),
	)

	return &Provider{
		tracerProvider: tracerProvider,
		loggerProvider: loggerProvider,
		logger:         logger,
	}, logger, nil
}

// Shutdown gracefully shuts down the telemetry providers
func (p *Provider) Shutdown(ctx context.Context) error {
	// Create a context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Shutdown tracer provider
	if err := p.tracerProvider.Shutdown(shutdownCtx); err != nil {
		p.logger.Error("failed to shutdown tracer provider", slog.Any("error", err))
		return err
	}

	// Shutdown logger provider
	if err := p.loggerProvider.Shutdown(shutdownCtx); err != nil {
		p.logger.Error("failed to shutdown logger provider", slog.Any("error", err))
		return err
	}

	p.logger.Info("OpenTelemetry shutdown completed")
	return nil
}
