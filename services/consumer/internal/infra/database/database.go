package database

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samuel-poirier/go-ref/consumer/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

//go:embed migrations/*.sql
var migrations embed.FS

var (
	ErrMissingMigrationsPath = errors.New("MIGRATIONS_PATH env missing")
	ErrMissingDatabaseURL    = errors.New("DATABASE_URL env missing")
)

func loadConfigFromURL() (*pgxpool.Config, error) {
	dbURL, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		return nil, fmt.Errorf("must set DATABASE_URL env var")
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

func loadConfig() (*pgxpool.Config, error) {
	cfg, err := config.NewDatabase()
	if err != nil {
		return loadConfigFromURL()
	}

	return pgxpool.ParseConfig(fmt.Sprintf(
		"user=%s password=%s host=%s port=%d dbname=%s sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	))
}

func dbURL() (string, error) {
	cfg, err := config.NewDatabase()
	if err != nil {
		dbURL, ok := os.LookupEnv("DATABASE_URL")
		if !ok {
			return "", fmt.Errorf("must set DATABASE_URL env var")
		}

		return dbURL, nil
	}

	return cfg.URL(), nil
}

func Connect(ctx context.Context, logger *slog.Logger) (*pgxpool.Pool, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	// Add query tracer for OpenTelemetry
	config.ConnConfig.Tracer = &pgxTracer{tracer: otel.Tracer("postgresql")}

	conn, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	logger.Debug("Running migrations")

	url, err := dbURL()
	if err != nil {
		return nil, err
	}

	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("create source: %w", err)
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", source, url)
	if err != nil {
		return nil, fmt.Errorf("migrate new: %s", err)
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return conn, nil
}

// pgxTracer implements pgx.QueryTracer for OpenTelemetry
type pgxTracer struct {
	tracer trace.Tracer
}

type pgxTraceCtxKey struct{}

func (t *pgxTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if !trace.SpanFromContext(ctx).IsRecording() {
		return ctx
	}

	ctx, span := t.tracer.Start(ctx, "db.query",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.statement", data.SQL),
		),
	)

	return context.WithValue(ctx, pgxTraceCtxKey{}, span)
}

func (t *pgxTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	span, ok := ctx.Value(pgxTraceCtxKey{}).(trace.Span)
	if !ok {
		return
	}

	if data.Err != nil {
		span.RecordError(data.Err)
		span.SetStatus(codes.Error, data.Err.Error())
	} else {
		span.SetStatus(codes.Ok, "query executed successfully")
		span.SetAttributes(
			attribute.Int64("db.rows_affected", data.CommandTag.RowsAffected()),
		)
	}

	span.End()
}
