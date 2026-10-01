package config

import (
	"errors"

	"github.com/goodaarzii/dakeh/internal/env"
)

type Config struct {
	Addr        string
	DatabaseURL string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:        env.GetString("HTTP_ADDR", ":8080"),
		DatabaseURL: env.GetString("DATABASE_URL", "host=localhost user=mohsen1212 password=m021021021M dbname=dakeh sslmode=disable"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	return cfg, nil
}
