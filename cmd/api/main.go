package main

import (
	"context"
	"log/slog"
	"os"
)

func main() {
	ctx := context.Background()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	slog.Info("start", ctx)

}
