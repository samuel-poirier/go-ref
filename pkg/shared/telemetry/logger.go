package telemetry

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

// NewLoggerProvider creates a new logger provider with OTLP exporter
func NewLoggerProvider(ctx context.Context, config *Config, res *resource.Resource) (*sdklog.LoggerProvider, error) {
	// Create OTLP log exporter
	logExporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(config.OTLPEndpoint),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Create logger provider options
	opts := []sdklog.LoggerProviderOption{
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(res),
	}

	// Optionally add stdout exporter for debugging
	if config.EnableStdout {
		stdoutExporter, err := stdoutlog.New(
			stdoutlog.WithPrettyPrint(),
		)
		if err != nil {
			return nil, err
		}
		opts = append(opts, sdklog.WithProcessor(sdklog.NewBatchProcessor(stdoutExporter)))
	}

	return sdklog.NewLoggerProvider(opts...), nil
}

// NewLogger creates a new slog.Logger that bridges to OpenTelemetry
func NewLogger(loggerProvider *sdklog.LoggerProvider) *slog.Logger {
	// Set the global logger provider
	global.SetLoggerProvider(loggerProvider)

	// Create an otelslog handler that bridges slog to OpenTelemetry
	handler := otelslog.NewHandler("telemetry")

	// Create a multi-handler that writes to both stdout and OpenTelemetry
	// This ensures logs are visible in stdout AND sent to the collector
	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	// Combine both handlers
	multiHandler := &multiHandler{
		handlers: []slog.Handler{stdoutHandler, handler},
	}

	return slog.New(multiHandler)
}

// multiHandler implements slog.Handler and forwards to multiple handlers
type multiHandler struct {
	handlers []slog.Handler
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// Enable if any handler is enabled
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	// Send to all handlers
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, record); err != nil {
			return err
		}
	}
	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &multiHandler{handlers: newHandlers}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &multiHandler{handlers: newHandlers}
}
