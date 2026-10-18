package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goodaarzii/dakeh/internal/env"
)

type Config struct {
	Addr        string
	DatabaseURL string

	JWTSecret  string
	JWTTTL     time.Duration // access token lifetime (short: 15 minutes)
	RefreshTTL time.Duration // refresh token lifetime (long: 30 days)

	// Browser origins allowed to call the API (see middleware.CORS).
	CORSOrigins []string

	// Object storage for product images (MinIO or any S3-compatible service).
	S3Endpoint      string // host:port WITHOUT http://, e.g. localhost:9000
	S3AccessKey     string
	S3SecretKey     string
	S3Bucket        string
	S3PublicBaseURL string // what a BROWSER uses, e.g. http://localhost:9000/dakeh-images
	S3UseSSL        bool

	// Redis (the shopping cart).
	RedisAddr     string // host:port, e.g. localhost:6379
	RedisPassword string
	RedisDB       int
	CartTTL       time.Duration // a cart that nobody touches disappears after this long
}

// Load reads every setting in one place and fails at startup if something is missing or unsafe.
func Load() (Config, error) {
	cfg := Config{
		Addr:        env.GetString("HTTP_ADDR", ":8080"),
		DatabaseURL: env.GetString("DATABASE_URL", ""),

		JWTSecret:  env.GetString("JWTSecret", ""),
		JWTTTL:     time.Duration(env.GetInt("JWTExpirationInSeconds", 900)) * time.Second,
		RefreshTTL: time.Duration(env.GetInt("RefreshExpirationInSeconds", 2592000)) * time.Second,

		CORSOrigins: splitList(env.GetString("CORS_ORIGINS", "http://localhost:5173,http://localhost:3000")),

		S3Endpoint:      env.GetString("S3_ENDPOINT", ""),
		S3AccessKey:     env.GetString("S3_ACCESS_KEY", ""),
		S3SecretKey:     env.GetString("S3_SECRET_KEY", ""),
		S3Bucket:        env.GetString("S3_BUCKET", ""),
		S3PublicBaseURL: env.GetString("S3_PUBLIC_BASE_URL", ""),
		S3UseSSL:        env.GetString("S3_USE_SSL", "false") == "true",

		RedisAddr:     env.GetString("REDIS_ADDR", ""),
		RedisPassword: env.GetString("REDIS_PASSWORD", ""),
		RedisDB:       int(env.GetInt("REDIS_DB", 0)),
		CartTTL:       time.Duration(env.GetInt("CartExpirationInSeconds", 2592000)) * time.Second,
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
	if cfg.RefreshTTL <= cfg.JWTTTL {
		return Config{}, errors.New("RefreshExpirationInSeconds must be greater than JWTExpirationInSeconds")
	}
	if cfg.CartTTL <= 0 {
		return Config{}, errors.New("CartExpirationInSeconds must be greater than 0")
	}

	required := []struct{ name, value string }{
		{"S3_ENDPOINT", cfg.S3Endpoint},
		{"S3_ACCESS_KEY", cfg.S3AccessKey},
		{"S3_SECRET_KEY", cfg.S3SecretKey},
		{"S3_BUCKET", cfg.S3Bucket},
		{"S3_PUBLIC_BASE_URL", cfg.S3PublicBaseURL},
		{"REDIS_ADDR", cfg.RedisAddr},
		{"REDIS_PASSWORD", cfg.RedisPassword},
	}
	for _, r := range required {
		if r.value == "" {
			return Config{}, fmt.Errorf("%s is required", r.name)
		}
	}
	if strings.Contains(cfg.S3Endpoint, "://") {
		return Config{}, errors.New("S3_ENDPOINT must be host:port without http:// (for example localhost:9000)")
	}

	return cfg, nil
}

// splitList turns "a, b ,c" into ["a" "b" "c"] and skips empty items.
func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
