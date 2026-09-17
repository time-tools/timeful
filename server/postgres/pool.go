package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"
)

const (
	applicationURIEnvironment = "POSTGRES_APPLICATION_URI"
	connectTimeoutEnvironment = "POSTGRES_CONNECT_TIMEOUT_SECONDS"
	maxConnectionsEnvironment = "POSTGRES_MAX_CONNS"
)

var Pool *pgxpool.Pool

// Init connects and verifies the PostgreSQL store before the server accepts
// requests, and installs the query tracer so database work joins its request
// trace. PostgreSQL-owned event routes may be enabled after startup.
func Init(tracerProvider trace.TracerProvider) func() {
	uri := os.Getenv(applicationURIEnvironment)
	if uri == "" {
		panic(applicationURIEnvironment + " environment variable is required")
	}

	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		panic(fmt.Sprintf("invalid %s: %v", applicationURIEnvironment, err))
	}
	config.MaxConns = environmentInt(maxConnectionsEnvironment, 10)
	config.ConnConfig.Tracer = newQueryTracer(tracerProvider)

	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout())
	defer cancel()
	Pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		panic(fmt.Sprintf("connect to PostgreSQL: %v", err))
	}
	if err := Pool.Ping(ctx); err != nil {
		Pool.Close()
		Pool = nil
		panic(fmt.Sprintf("ping PostgreSQL: %v", err))
	}

	return func() {
		if Pool != nil {
			Pool.Close()
			Pool = nil
		}
	}
}

func Ping(ctx context.Context) error {
	if Pool == nil {
		return errors.New("postgresql pool is not initialized")
	}
	return Pool.Ping(ctx)
}

func connectionTimeout() time.Duration {
	return time.Duration(environmentInt(connectTimeoutEnvironment, 10)) * time.Second
}

func environmentInt(name string, fallback int) int32 {
	value := os.Getenv(name)
	if value == "" {
		return int32(fallback)
	}
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil || parsed < 1 {
		panic(fmt.Sprintf("%s must be a positive integer", name))
	}
	return int32(parsed)
}
