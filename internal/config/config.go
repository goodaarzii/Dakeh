package config

import (
	"errors"
	"time"

	"github.com/goodaarzii/dakeh/internal/env"
)

type Config struct {
	Addr        string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
}

// Load reads every setting in one place and fails at startup if something is unsafe.
func Load() (Config, error) {
	cfg := Config{
		Addr:        env.GetString("HTTP_ADDR", ":8080"),
		DatabaseURL: env.GetString("DATABASE_URL", ""),
		JWTSecret:   env.GetString("JWTSecret", ""),
		JWTTTL:      time.Duration(env.GetInt("JWTExpirationInSeconds", 900)) * time.Second,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWTSecret is required and must be at least 32 characters")
	}
	if cfg.JWTTTL <= 0 {
		return Config{}, errors.New("JWTExpirationInSeconds must be greater than 0")
	}
	return cfg, nil
}
