package app

import (
	"github.com/goodaarzii/dakeh/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Application struct {
	Config config.Config
	DB     *pgxpool.Pool
}
