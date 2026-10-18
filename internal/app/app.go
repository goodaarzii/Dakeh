package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/goodaarzii/dakeh/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/goodaarzii/dakeh/internal/config"
)

// Application holds shared dependencies. Add Redis, MinIO, etc. here later.
type Application struct {
	Config config.Config
	DB     *pgxpool.Pool
	Store  *storage.MinioStorage
	Rdb    *redis.Client
}

func New(cfg config.Config, db *pgxpool.Pool, store *storage.MinioStorage, rdb *redis.Client) *Application {
	return &Application{Config: cfg, DB: db, Store: store, Rdb: rdb}
}

// Run starts the server and shuts down gracefully when ctx is cancelled (Ctrl+C / SIGTERM).
func (a *Application) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:         a.Config.Addr,
		Handler:      a.Mount(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", a.Config.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
