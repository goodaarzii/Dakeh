package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/goodaarzii/dakeh/internal/app"
	"github.com/goodaarzii/dakeh/internal/config"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run keeps main tiny and lets deferred cleanup (pool.Close) actually run.
func run() error {
	if err := godotenv.Load(); err != nil {
		slog.Warn(".env not loaded", "err", err)
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return err
	}

	return app.New(cfg, pool).Run(ctx)
}
