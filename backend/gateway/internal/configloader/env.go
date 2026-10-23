package configloader

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"aziz.dev/gateway/internal/config"
	"github.com/joho/godotenv"
)

// like that so it can work locally, docker and k8s
func LoadFromEnv() (*config.AppConfig, error) {
	if env := os.Getenv("APP_ENV"); env == "" || env == "local" {
		loadDotEnv()
	}

	accessExpiry, err := durationFromEnv("JWT_ACCESS_EXPIRY")
	if err != nil {
		return nil, err
	}

	refreshExpiry, err := durationFromEnv("JWT_REFRESH_EXPIRY")
	if err != nil {
		return nil, err
	}

	return &config.AppConfig{
		Service: config.ServiceConfig{
			Name: os.Getenv("SERVICE_NAME"),
			Port: os.Getenv("SERVICE_PORT"),
			ShortenerServiceURL: os.Getenv("SHORTENER_SERVICE_URL"),
		},
		Postgres: config.PostgresConfig{
			Host: os.Getenv("POSTGRES_HOST"),
			User: os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			DBName: os.Getenv("POSTGRES_DB"),
		},
		Redis: config.RedisConfig{
			Addr: os.Getenv("REDIS_ADDR"),
			Username: os.Getenv("REDIS_USER"),
			Password: os.Getenv("REDIS_PASSWORD"),
		},
		JWT: config.JWTConfig{
			Secret: []byte(os.Getenv("JWT_SECRET")),
			AccessExpiry: accessExpiry,
			RefreshExpiry: refreshExpiry,
		},
	}, nil
}

func loadDotEnv() {
	if _, file, _, ok := runtime.Caller(0); ok {
		_ = godotenv.Load(filepath.Join(filepath.Dir(file), "..", "..", ".env"))
	}
}

func durationFromEnv(key string) (time.Duration, error) {
	value := os.Getenv(key)
	
	if value == "" {
		return 0, fmt.Errorf("%s is required", key)
	}

	d, err := time.ParseDuration(value)

	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", key, value, err)
	}

	return d, nil
}