package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/goodaarzii/dakeh/internal/app"
	"github.com/goodaarzii/dakeh/internal/config"
	"github.com/goodaarzii/dakeh/internal/env"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	slog.Info("start", ctx)

	cfg := config.Config{
		Addr: ":8080",
		DB: config.DBConfig{
			DSN: env.GetString("GOOSE_DBSTRING", "host=localhost user=mohsen1212 password=m021021021M dbname=dakeh sslmode=disable"),
		},
	}

	pool, err := pgxpool.New(ctx, cfg.DB.DSN)
	if err != nil {
		slog.Error("error in coonect ", err)
	}
	defer pool.Close()

	slog.Info("dsn db", cfg.DB.DSN)

	api := app.Application{
		Config: cfg,
		DB:     pool,
	}

	if err := api.Run(api.Mount()); err != nil {
		slog.Error("error starting server", err)
		return
	}

}
